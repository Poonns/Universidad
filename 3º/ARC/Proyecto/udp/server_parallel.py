"""Experimento Linux/free-threaded. W threads/loops/sockets, control propietario en W0."""
from __future__ import annotations

import argparse
import asyncio
from collections import Counter
import json
import os
from pathlib import Path
import signal
import socket
import threading
import time
import uuid

from protocol import (MSG_ACK, MSG_ERROR, MSG_GROUP_DONE, MSG_GROUP_DONE_ACK,
    MSG_METRICS, MSG_METRICS_ACK, MSG_POSITION, MSG_REGISTER, MSG_START, MSG_START_ACK,
    ProtocolError, decode_message, encode_message, identifier, message)
from runtime import Counters, add_common_arguments, configure_socket, default_server_result, policy_from, write_json
from server import SimulationServer
from parallel_runtime import ControlInbox, LinuxSampler, Notice, Routes, require_runtime, runtime_probe, udp_sockets
from protocol import GENERATOR_REQUESTS
from distributed_config import RunManifest
from remote_control import GeneratorInbox, RemoteNotice, RemoteRun


class RegistrationServer(SimulationServer):
    """Estado mutable de protocolo con un único propietario: loop del worker 0."""
    def __init__(self, worker, **config):
        super().__init__(**config)
        self.worker = worker
        self.stats = worker.stats
        self.full_registry = asyncio.Event()
        self.parallel_ready = False
        self.start_sent_at = {}
        self.done_sent_at = {}
        self.notice_event = asyncio.Event()
        self.notice_task = None
        self.notice_delay_max = 0.0
        self.data_done_at = None
        self.remote = RemoteRun(self, worker.runtime.manifest) if worker.runtime.manifest else None
        self.grant_publish_seconds = self.grant_publish_cpu_seconds = self.grant_publish_max = 0.0
        self.grant_publish_count = 0

    def _register(self, msg, addr):
        if self.remote:
            try:
                self.remote.before_register(msg)
            except ValueError as exc:
                raise ProtocolError(str(exc)) from exc
        before = len(self.clients)
        super()._register(msg, addr)
        if self.remote and len(self.clients) > before:
            self.remote.registered[msg['datos']['generador']] += 1

    def _schedule_control(self, kind, cid):
        if kind == MSG_START and not self.parallel_ready:
            self.full_registry.set()
            return  # No INICIO antes de publicar rutas y abrir TODOS los sockets.
        super()._schedule_control(kind, cid)

    def fail(self, reason):
        super().fail(reason)
        self.worker.runtime.abort(reason)

    def _send_raw(self, data, addr, kind):
        self.worker.send_raw(data, addr, kind)

    def _send(self, msg, addr):
        kind, body = msg['tipo'], msg['datos']
        first = False
        if kind == MSG_START and body['destino'] not in self.start_sent_at:
            self.start_sent_at[body['destino']] = time.perf_counter()
            first = True
        if kind == MSG_GROUP_DONE:
            self.done_sent_at.setdefault(body['destino'], time.perf_counter())
        super()._send(msg, addr)
        if first:
            stamp = self.start_sent_at[body['destino']]
            begin, cpu = time.perf_counter(), time.thread_time()
            for receiver in self.worker.runtime.workers:
                receiver.publish_start(body['destino'], stamp)
            elapsed = time.perf_counter() - begin
            self.grant_publish_count += 1
            self.grant_publish_seconds += elapsed
            self.grant_publish_cpu_seconds += time.thread_time() - cpu
            self.grant_publish_max = max(self.grant_publish_max, elapsed)

    def accept_notice(self, notice):
        cid, kind = notice.client_id, notice.kind
        info = self.clients[cid]  # validado por rutas congeladas en el receptor
        if (notice.group_id != info['group_id'] or cid not in self.start_sent_at
                or notice.received_at < self.start_sent_at[cid]):
            raise ProtocolError('control recibido antes de INICIO')
        self.notice_delay_max = max(self.notice_delay_max, time.perf_counter() - notice.received_at)
        if kind == MSG_START_ACK:
            if notice.instance != self.instance_id:
                raise ProtocolError('INICIO_ACK de otra instancia')
            self.started_clients.add(cid)
            self.pending_controls.pop((MSG_START, cid), None)
            self._maybe_finish_group(info['group_id'])
            self.control_changed.set()
        elif kind == MSG_METRICS:
            if (notice.iterations != self.total_iterations
                    or notice.unique_acks != self.total_iterations * (self.neighbors_per_group - 1)):
                raise ProtocolError('métricas de una simulación incompleta')
            if cid in self.metrics:
                if self.metrics[cid] != notice.mean:
                    raise ProtocolError('métricas duplicadas contradictorias')
                self.stats.metrics_duplicates += 1
            else:
                self.metrics[cid] = notice.mean
                if len(self.metrics) == self.num_clients:
                    self.last_metrics_at = time.perf_counter()
                    self.all_metrics_event.set()
            self._send(self._control_message(MSG_METRICS_ACK, cid), info['addr'])
            self._maybe_finish_group(info['group_id'])
        elif kind == MSG_GROUP_DONE_ACK:
            if (notice.instance != self.instance_id or cid not in self.done_sent_at
                    or notice.received_at < self.done_sent_at[cid]
                    or info['group_id'] not in self.finished_groups):
                raise ProtocolError('GRUPO_FIN_ACK prematuro o de otra instancia')
            self.done_acks.add(cid)
            self.pending_controls.pop((MSG_GROUP_DONE, cid), None)
            self.control_changed.set()
            if (len(self.done_acks) == self.num_clients and len(self.metrics) == self.num_clients
                    and len(self.started_clients) == self.num_clients):
                self.complete_event.set()
                self.data_done_at = time.perf_counter()
                self.worker.runtime.protocol_done.set()
        else:
            raise ProtocolError('aviso de control desconocido')

    async def notices(self):
        while True:
            self.notice_event.clear()
            for worker in self.worker.runtime.workers:
                if self.remote:
                    for notice in worker.generator_inbox.drain():
                        self.remote.handle(notice)
                for notice in worker.inbox.drain():
                    try:
                        self.accept_notice(notice)
                    except ProtocolError:
                        self.stats.protocol_errors += 1
            await self.notice_event.wait()


class ReceiverProtocol(asyncio.DatagramProtocol):
    def __init__(self, worker):
        self.worker = worker

    def connection_made(self, transport):
        self.worker.transport = transport
        if self.worker.server:
            self.worker.server.transport = transport

    def datagram_received(self, data, addr):
        self.worker.receive(data, addr)

    def error_received(self, exc):
        self.worker.stats.socket_errors += 1
        self.worker.runtime.abort(f'worker {self.worker.index}: UDP {exc}')

    def connection_lost(self, exc):
        self.worker.closed.set()
        if self.worker.server:
            self.worker.server.closed_event.set()


class Worker:
    def __init__(self, runtime, index):
        self.runtime, self.index = runtime, index
        self.routes = None  # asignado ANTES de Thread.start en los peers
        self.ready = threading.Event()
        self.finished = threading.Event()
        self.loop_ready = threading.Event()
        self.inbox = ControlInbox(3 * runtime.config['num_clients'])
        self.generator_inbox = GeneratorInbox(len(runtime.manifest.generators)) if runtime.manifest else None
        self.grant_lock = threading.Lock()  # solo publicaciones de control, nunca datos
        self.grant_pending = {}
        self.grant_armed = False
        self.local_starts = {}  # exclusivamente del loop receptor
        self.deferred = {}
        self.deferred_count = self.deferred_max = self.deferred_drops = 0
        self.grants_applied = 0
        self.grant_lag_sum = self.grant_lag_max = 0.0
        self.grant_lag_over_rto = 0
        self.stats = None
        self.data_received = self.data_forwarded = 0
        self.clients_seen = set()
        self.loop = self.stop = self.closed = self.transport = self.server = None
        self.inode = None
        self.result = None
        self.error = None
        self.native_id = None
        self.started_at = self.finished_at = self.cpu_seconds = None
        self.ready_at = None
        self.observation = None
        self.loop_lag_count = 0
        self.loop_lag_sum = self.loop_lag_max = 0.
        self.transport_queue_max = 0
        self.socket_last = {}
        self.buffers = {}
        self.probe = None
        self.thread = threading.Thread(target=self.run, name=f'arc-udp-{index}', daemon=False)

    def run(self):
        self.native_id = threading.get_native_id()
        self.started_at = time.perf_counter()
        cpu_start = time.thread_time()
        self.stats = Counters()
        try:
            asyncio.run(self.serve())
        except BaseException as exc:
            self.error = f'{type(exc).__name__}: {exc}'
            self.runtime.abort(f'worker {self.index}: {self.error}')
        finally:
            self.finished_at = time.perf_counter()
            self.cpu_seconds = time.thread_time() - cpu_start
            self.result = {'worker': self.index, 'pid': os.getpid(), 'native_thread_id': self.native_id,
                'loop_id': id(self.loop), 'socket_inode': self.inode, 'error': self.error,
                'cpu_seconds': self.cpu_seconds,
                'cpu_percent_lifetime': 100 * self.cpu_seconds / (self.finished_at - self.started_at),
                'data_received': self.data_received, 'data_forwarded': self.data_forwarded,
                'unique_client_endpoints': len(self.clients_seen), 'counters': self.stats.report(),
                'startup_deferred_maximum': self.deferred_max,
                'startup_deferred_drops': self.deferred_drops, 'buffers': self.buffers,
                'control_inbox_maximum': self.inbox.maximum,
                'event_loop_lag': dict(samples=self.loop_lag_count,maximum_seconds=self.loop_lag_max,
                    mean_seconds=self.loop_lag_sum/self.loop_lag_count if self.loop_lag_count else None),
                'transport_queue_max_bytes': self.transport_queue_max,
                'grant_application': dict(count=self.grants_applied,
                    mean_seconds=self.grant_lag_sum/self.grants_applied if self.grants_applied else None,
                    maximum_seconds=self.grant_lag_max, over_initial_rto=self.grant_lag_over_rto),
                'runtime': self.probe}
            self.finished.set()

    async def serve(self):
        self.loop = asyncio.get_running_loop()
        self.stop, self.closed = asyncio.Event(), asyncio.Event()
        self.loop_ready.set()
        if self.runtime.stop_flag.is_set():
            return
        if self.index == 0:
            self.server = RegistrationServer(self, **self.runtime.config)
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        try:
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
            sock.setblocking(False)
            address = (self.runtime.config['host'], self.runtime.config['port']) if self.index == 0 else self.runtime.address
            sock.bind(address)
            if self.index == 0:
                self.runtime.address = sock.getsockname()
            if sys_platform_linux():
                target = os.readlink(f'/proc/self/fd/{sock.fileno()}')
                self.inode = int(target.partition('[')[2].rstrip(']'))
            transport, _ = await self.loop.create_datagram_endpoint(lambda: ReceiverProtocol(self), sock=sock)
            sock = None  # ownership transferido al transport
            self.buffers = configure_socket(transport, self.runtime.config['receive_buffer'],
                                            self.runtime.config['send_buffer'])
            if self.server:
                self.server.buffers = self.buffers
            if self.runtime.transport_wrapper:
                self.transport = self.runtime.transport_wrapper(self.index, self.transport)
                if self.server:
                    self.server.transport = self.transport
            self.probe = runtime_probe()
            if self.runtime.enforce_runtime:
                require_runtime(self.probe)
            self.ready_at = time.perf_counter()
            self.ready.set()
            monitor = asyncio.create_task(self.monitor())
            if self.index == 0:
                self.runtime.bound.set()
                self.server.control_task = asyncio.create_task(self.server._control_loop())
                lifecycle = asyncio.create_task(self.activate())
                notices = asyncio.create_task(self.server.notices())
                self.server.notice_task = notices
                for task in (lifecycle, notices, self.server.control_task):
                    task.add_done_callback(self.task_finished)
            else:
                lifecycle = notices = None
            if self.runtime.stop_flag.is_set():
                self.stop.set()
            await self.stop.wait()
        finally:
            if self.inode is not None:
                self.socket_last = udp_sockets().get(self.inode, {})
            if sock is not None:
                sock.close()
            for task in (locals().get('lifecycle'), locals().get('notices'), locals().get('monitor')):
                if task:
                    task.cancel()
            await asyncio.gather(*(task for task in (locals().get('lifecycle'), locals().get('notices'), locals().get('monitor')) if task),
                                 return_exceptions=True)
            if self.server:
                await self.server.close()
            elif self.transport:
                self.transport.close()
                await self.closed.wait()

    async def monitor(self):
        """Snapshot publicado por su loop: el supervisor nunca lee dicts DATA mutando."""
        while True:
            due=self.loop.time()+.1
            await asyncio.sleep(.1)
            lag=max(0,self.loop.time()-due)
            self.loop_lag_count+=1
            self.loop_lag_sum+=lag
            self.loop_lag_max=max(self.loop_lag_max,lag)
            transport=self.transport
            queue=transport.get_write_buffer_size() if hasattr(transport,'get_write_buffer_size') else 0
            self.transport_queue_max=max(self.transport_queue_max,queue)
            owner=self.server
            self.observation=dict(at=time.perf_counter(),worker=self.index,lag_seconds=lag,
                data_received=self.data_received,data_forwarded=self.data_forwarded,
                rx_by_type=dict(self.stats.rx_by_type),tx_by_type=dict(self.stats.tx_by_type),
                transport_queue_bytes=queue,timeouts=self.stats.timeouts,retries=self.stats.retransmissions,
                registered=len(owner.clients) if owner else None,
                started=len(owner.started_clients) if owner else None,
                metrics=len(owner.metrics) if owner else None,done_acks=len(owner.done_acks) if owner else None)

    def task_finished(self, task):
        if not task.cancelled() and task.exception() is not None:
            self.runtime.abort(f'worker 0 tarea de control: {task.exception()}')

    async def activate(self):
        await self.server.full_registry.wait()
        # Solo W0 ha podido recibir REGISTRO nuevos hasta este punto.
        snapshot = Routes.publish(self.server)
        self.routes = snapshot
        self.runtime.snapshot = snapshot
        self.runtime.published_at = time.perf_counter()
        self.runtime.published.set()  # sincronización explícita de publicación
        for peer in self.runtime.workers[1:]:
            peer.routes = snapshot
            self.runtime.start_worker(peer)
        while not all(worker.ready.is_set() for worker in self.runtime.workers):
            if self.runtime.stop_flag.is_set():
                return
            await asyncio.sleep(0.001)
        self.runtime.activated_at = time.perf_counter()
        self.runtime.activated.set()
        self.server.parallel_ready = True
        for cid in range(1, snapshot.n + 1):
            self.server._schedule_control(MSG_START, cid)

    def publish_start(self, cid, stamp):
        if self.index == 0:
            self.apply_grants({cid: stamp})
            return
        with self.grant_lock:
            self.grant_pending.setdefault(cid, stamp)
            wake = not self.grant_armed
            self.grant_armed = True
        if wake:
            try:
                self.loop.call_soon_threadsafe(self.flush_grants)
            except RuntimeError:
                if not self.runtime.stop_flag.is_set():
                    self.runtime.abort('loop receptor cerrado durante publicación de INICIO')

    def flush_grants(self):
        with self.grant_lock:
            batch = self.grant_pending
            self.grant_pending = {}
            self.grant_armed = False
        self.apply_grants(batch)

    def apply_grants(self, batch):
        for cid, stamp in batch.items():
            lag = time.perf_counter() - stamp
            self.grants_applied += 1
            self.grant_lag_sum += lag
            self.grant_lag_max = max(self.grant_lag_max, lag)
            self.grant_lag_over_rto += lag >= self.runtime.config['policy'].timeout
            self.local_starts.setdefault(cid, stamp)
            pending = self.deferred.pop(cid, ())
            self.deferred_count -= len(pending)
            for msg, data, addr, received in pending:
                try:
                    self.dispatch(msg, data, addr, received)
                except ProtocolError:
                    self.stats.protocol_errors += 1

    def send_raw(self, data, addr, kind):
        try:
            self.transport.sendto(data, addr)
            self.stats.sent(kind)
            if kind in (MSG_POSITION, MSG_ACK):
                self.data_forwarded += 1
        except OSError as exc:
            self.stats.socket_errors += 1
            self.runtime.abort(f'worker {self.index} sendto: {exc}')

    def send(self, msg, addr):
        self.send_raw(encode_message(msg), addr, msg['tipo'])

    def receive(self, data, addr):
        if self.routes is None:
            if self.runtime.manifest:
                try:
                    msg = decode_message(data)
                    if msg['tipo'] in GENERATOR_REQUESTS:
                        self.stats.datagrams_received += 1
                        self.stats.received(msg['tipo'])
                        self.remote_request(msg, addr)
                        return
                except ProtocolError:
                    self.stats.protocol_errors += 1
                    return
            self.server.handle_datagram(data, addr)
            return
        self.stats.datagrams_received += 1
        try:
            msg = decode_message(data)
            self.stats.received(msg['tipo'])
            if msg['tipo'] in (MSG_POSITION, MSG_ACK):
                self.data_received += 1
            self.dispatch(msg, data, addr)
        except ProtocolError:
            self.stats.protocol_errors += 1

    def dispatch(self, msg, data, addr, received=None):
        routes = self.routes
        if msg['sesion'] != routes.session:
            raise ProtocolError('sesión ajena')
        kind, body = msg['tipo'], msg['datos']
        if kind in GENERATOR_REQUESTS:
            self.remote_request(msg, addr)
            return
        if kind == MSG_REGISTER:
            try:
                reply = routes.registration_reply(msg, addr)
            except ProtocolError:
                if (body['num_clientes'], body['num_vecinos'], body['total_iteraciones']) != (
                        routes.n, routes.v, routes.iterations):
                    self.send(message(MSG_ERROR, routes.session, token=body['token'],
                                      mensaje='N/V/I incompatibles con el servidor'), addr)
                raise
            self.stats.registration_duplicates += 1
            self.send(reply, addr)
            return
        route = routes.sender(msg, addr)
        cid = route.client_id
        self.clients_seen.add(cid)
        if kind != MSG_ACK:
            stamp = self.local_starts.get(cid)
            if stamp is None:
                if self.deferred_count >= self.runtime.defer_capacity:
                    self.deferred_drops += 1
                    return
                self.deferred.setdefault(cid, []).append((msg, data, addr,
                    received if received is not None else time.perf_counter()))
                self.deferred_count += 1
                self.deferred_max = max(self.deferred_max, self.deferred_count)
                return
            if received is not None and received < stamp:
                raise ProtocolError('mensaje recibido antes del envío de INICIO')
        if kind == MSG_POSITION:
            if not 1 <= body['iteracion'] <= routes.iterations:
                raise ProtocolError('iteración fuera de rango')
            targets = body.get('pendientes', routes.groups[route.group_id])
            if 'pendientes' in body and any(target == cid or target not in routes.groups[route.group_id]
                                           for target in targets):
                raise ProtocolError('vecinos pendientes inválidos')
            for target in targets:
                if target != cid:
                    self.send_raw(data, routes.clients[target].addr, MSG_POSITION)
        elif kind == MSG_ACK:
            target = body['destino']
            if (not 1 <= target <= routes.n or target == cid
                    or routes.clients[target].group_id != route.group_id
                    or not 1 <= body['iteracion'] <= routes.iterations):
                raise ProtocolError('ACK con origen/destino/grupo/iteración inválidos')
            self.send_raw(data, routes.clients[target].addr, MSG_ACK)
        elif kind in (MSG_START_ACK, MSG_GROUP_DONE_ACK, MSG_METRICS):
            notice = Notice(kind, cid, route.group_id,
                received if received is not None else time.perf_counter(), body.get('instancia', ''),
                body.get('latencia_media_ms'), body.get('total_iteraciones', 0), body.get('acks_unicos', 0))
            status, wake = self.inbox.put(notice)
            if status == 'conflict':
                raise ProtocolError('control duplicado contradictorio')
            if status == 'full':
                self.runtime.abort('buzón de control fuera del límite de identidades')
            if status == 'duplicate' and kind == MSG_METRICS:
                self.stats.metrics_duplicates += 1
            if wake:
                owner = self.runtime.workers[0]
                try:
                    owner.loop.call_soon_threadsafe(owner.server.notice_event.set)
                except RuntimeError:
                    if not self.runtime.stop_flag.is_set():
                        self.runtime.abort('coordinador cerrado con control pendiente')
        else:
            raise ProtocolError('tipo no admitido hacia el servidor')

    def remote_request(self, msg, addr):
        manifest = self.runtime.manifest
        if not manifest or msg['sesion'] != manifest.session:
            raise ProtocolError('control remoto sin manifiesto o sesión distinta')
        body = msg['datos']
        if body['generador'] not in {s.id for s in manifest.generators}:
            raise ProtocolError('generador no autorizado')
        notice = RemoteNotice(msg['tipo'], body['generador'], tuple(body.items()), tuple(addr), time.perf_counter())
        if self.index == 0:
            self.server.remote.handle(notice)
        else:
            self.generator_inbox.put(notice)
            owner = self.runtime.workers[0]
            owner.loop.call_soon_threadsafe(owner.server.notice_event.set)


def sys_platform_linux():
    import sys
    return sys.platform == 'linux'


class ParallelServer:
    def __init__(self, *, workers=1, host='0.0.0.0', port=8888, num_clients=100,
                 neighbors_per_group=25, total_iterations=5, session_id=None, policy=None,
                 deadline=120, control_rate=2000, receive_buffer=256*1024, send_buffer=256*1024,
                 result_path=None, ready_path=None, defer_capacity=1024,
                 enforce_runtime=True, transport_wrapper=None, manifest=None, terminal_linger=None,
                 minimum_vm_available_mib=0, live_path=None):
        from protocol import validate_configuration
        from runtime import RetryPolicy, positive
        validate_configuration(num_clients, neighbors_per_group, total_iterations)
        if manifest and (num_clients, neighbors_per_group, total_iterations, session_id) != (
                manifest.n, manifest.v, manifest.i, manifest.session):
            raise ValueError('N/V/I/sesión no coinciden con el manifiesto')
        self.manifest = manifest
        self.minimum_vm_available_bytes=minimum_vm_available_mib*1024*1024
        if minimum_vm_available_mib<0:
            raise ValueError('mínimo de RAM disponible no puede ser negativo')
        self.live_path=live_path
        if workers not in (1, 2, 4, 6, 8):
            raise ValueError('workers debe ser 1, 2, 4, 6 u 8')
        if not hasattr(socket, 'SO_REUSEPORT'):
            raise RuntimeError('SO_REUSEPORT no disponible')
        socket.inet_pton(socket.AF_INET, host)
        if not 0 <= port <= 65535 or min(receive_buffer, send_buffer) < 4096 or defer_capacity <= 0:
            raise ValueError('puerto, buffers o capacidad de transición inválidos')
        positive(deadline, 'deadline')
        self.config = dict(host=host, port=port, num_clients=num_clients,
            neighbors_per_group=neighbors_per_group, total_iterations=total_iterations,
            session_id=identifier(session_id or uuid.uuid4().hex, 'sesion'), policy=policy or RetryPolicy(),
            deadline=deadline, control_rate=control_rate, receive_buffer=receive_buffer, send_buffer=send_buffer)
        retry_policy = self.config['policy']
        recovery = sum(min(retry_policy.max_timeout, retry_policy.timeout*2**j)
                       for j in range(retry_policy.max_attempts)) + retry_policy.max_timeout
        self.terminal_linger = recovery if terminal_linger is None else terminal_linger
        positive(self.terminal_linger, 'terminal_linger')
        self.enforce_runtime, self.transport_wrapper = enforce_runtime, transport_wrapper
        self.result_path, self.ready_path, self.defer_capacity = result_path, ready_path, defer_capacity
        self.stop_flag, self.bound, self.published, self.activated = (threading.Event() for _ in range(4))
        self.protocol_done = threading.Event()
        self.error_lock = threading.Lock()  # solo errores, fuera de la ruta de datos
        self.error = None
        self.started_lock = threading.Lock()  # solo ciclo de vida de los hilos
        self.started = []
        self.address = self.snapshot = self.activated_at = self.published_at = None
        self.workers = tuple(Worker(self, index) for index in range(workers))
        self.result = None
        self.initial_probe = None

    def start_worker(self, worker):
        with self.started_lock:
            if self.stop_flag.is_set():
                return
            self.started.append(worker)
            worker.thread.start()

    def abort(self, reason):
        with self.error_lock:
            if self.error is None:
                self.error = reason
        self.request_stop()

    def request_stop(self):
        self.stop_flag.set()
        for worker in self.workers:
            if worker.loop_ready.is_set():
                try:
                    worker.loop.call_soon_threadsafe(worker.stop.set)
                except RuntimeError:
                    pass

    def run(self):
        self.initial_probe = runtime_probe()
        if self.enforce_runtime:
            require_runtime(self.initial_probe)
        sampler = LinuxSampler()
        start = time.perf_counter()
        self.start_worker(self.workers[0])
        ready_written = False
        last_live = 0.
        memory_low_since = None
        try:
            while not self.stop_flag.is_set():
                now = time.perf_counter()
                if self.enforce_runtime:
                    require_runtime()
                if self.bound.is_set() and not ready_written:
                    ready = {'event': 'ready', 'phase': 'registration', 'host': self.address[0],
                        'port': self.address[1], 'session': self.config['session_id'],
                        'instance': self.workers[0].server.instance_id, 'workers': len(self.workers),
                        'active_receivers_at_ready': 1, 'runtime': self.initial_probe}
                    if self.ready_path:
                        write_json(self.ready_path, ready)
                    print(json.dumps(ready), flush=True)
                    ready_written = True
                sampler.sample(now, self.workers)
                available=sampler.latest_vm.get('memavailable_bytes')
                if self.minimum_vm_available_bytes and available is not None and available<self.minimum_vm_available_bytes:
                    if memory_low_since is None:
                        memory_low_since=now
                    if now-memory_low_since>=.5:
                        self.abort('RAM disponible de la VM bajo el margen preventivo configurado')
                        break
                else:
                    memory_low_since=None
                if self.live_path and now-last_live>=1:
                    write_json(self.live_path,dict(at=now,vm=sampler.latest_vm,
                        workers=[w.observation for w in self.workers if w.observation],
                        sockets=sampler.socket_max))
                    last_live=now
                if self.protocol_done.is_set():
                    remote = self.workers[0].server.remote
                    if remote is None or (remote.all_released.is_set()
                            and now - remote.all_released_at >= self.terminal_linger):
                        self.request_stop()
                        break
                if now - start >= self.config['deadline']:
                    self.abort('deadline administrativo agotado')
                    break
                self.stop_flag.wait(0.1)
        except BaseException as exc:
            self.abort(f'supervisor: {type(exc).__name__}: {exc}')
            raise
        finally:
            self.request_stop()
            with self.started_lock:
                started = tuple(self.started)
            for worker in started:
                worker.thread.join()
            end = time.perf_counter()
            server = self.workers[0].server
            base = server.report() if server else {}
            counters = Counters().report()
            total_data_rx = sum(w.data_received for w in started)
            duration = base.get('simulation_seconds') or (end - start)
            rows = []
            for worker in started:
                row = dict(worker.result)
                for key, value in row['counters'].items():
                    if isinstance(value, dict):
                        merged = Counter(counters[key]); merged.update(value); counters[key] = dict(merged)
                    else:
                        counters[key] += value
                row['data_rx_percent'] = 100 * row['data_received'] / total_data_rx if total_data_rx else 0
                row['data_rx_per_second'] = row['data_received'] / duration
                row['data_tx_per_second'] = row['data_forwarded'] / duration
                observed = dict(sampler.socket_max.get(worker.inode, {}))
                for key, value in worker.socket_last.items():
                    observed[key] = max(observed.get(key, 0), value)
                row['socket_observed'] = observed
                rows.append(row)
            remote = server.remote if server else None
            success = (self.error is None and self.protocol_done.is_set() and len(started) == len(self.workers)
                       and (remote is None or len(remote.released) == len(self.manifest.generators)))
            tids = {str(w.native_id) for w in started}
            concurrent_cpu = max((sum(value for tid,value in sample['thread_cpu_percent'].items() if tid in tids)
                                  for sample in sampler.samples), default=0)
            self.result = dict(base, role='server', success=success, simulation_complete=success,
                error=self.error, counters=counters, total_seconds=end-start, workers=rows,
                simulation_clock=dict(start=server.simulation_at if server else None,
                    end=server.last_metrics_at if server else None),
                phase_clock=dict(created=start,first_registration=server.first_registration_at if server else None,
                    registered=server.registered_at if server else None,published=self.published_at,
                    first_start=min(server.start_sent_at.values()) if server and server.start_sent_at else None,
                    last_start=max(server.start_sent_at.values()) if server and server.start_sent_at else None,
                    last_metrics=server.last_metrics_at if server else None,data_done=server.data_done_at if server else None,
                    generators_done=remote.all_released_at if remote else None,finished=end),
                configured_workers=len(self.workers), server_process_pid=os.getpid(),
                runtime_initial=self.initial_probe, runtime_final=runtime_probe(),
                real_parallelism_eligible=self.enforce_runtime and success,
                routing_publication={'registered': len(server.clients) if server else 0,
                    'published_before_workers': self.published_at is not None and all(
                        self.published_at <= w.started_at for w in started if w.index != 0),
                    'all_workers_ready_before_start': bool(server and server.start_sent_at) and all(
                        w.ready_at is not None and w.ready_at <= min(server.start_sent_at.values()) for w in started),
                    'activation_seconds': self.activated_at-start if self.activated_at else None},
                control_owner_worker=0, control_notice_max_delay_seconds=server.notice_delay_max if server else None,
                start_authorization_publication=dict(count=server.grant_publish_count if server else 0,
                    total_seconds=server.grant_publish_seconds if server else 0,
                    cpu_seconds=server.grant_publish_cpu_seconds if server else 0,
                    maximum_seconds=server.grant_publish_max if server else 0,
                    start_spread_seconds=(max(server.start_sent_at.values())-min(server.start_sent_at.values()))
                        if server and server.start_sent_at else None),
                distributed=remote.report() if remote else None,
                terminal_linger_seconds=self.terminal_linger if remote else None,
                simultaneous_worker_cpu_peak_percent=concurrent_cpu,
                data_rx_per_second=total_data_rx/duration,
                data_tx_per_second=sum(w.data_forwarded for w in started)/duration,
                telemetry=sampler.report(end-start))
            if self.result_path:
                write_json(self.result_path, self.result)
        return self.result


def parse_args(argv=None):
    parser = argparse.ArgumentParser(description='Servidor experimental UDP Linux CPython free-threaded')
    add_common_arguments(parser)
    parser.set_defaults(host='0.0.0.0')
    parser.add_argument('--workers', type=int, default=1, choices=(1,2,4,6,8))
    parser.add_argument('--session')
    parser.add_argument('--ready-file')
    parser.add_argument('--control-rate', type=float, default=2000)
    parser.add_argument('--receive-buffer', type=int, default=256*1024)
    parser.add_argument('--send-buffer', type=int, default=256*1024)
    parser.add_argument('--defer-capacity', type=int, default=1024)
    parser.add_argument('--check-runtime', action='store_true')
    parser.add_argument('--manifest', help='manifiesto global estático; habilita cierre remoto')
    parser.add_argument('--minimum-vm-available-mib',type=int,default=0)
    parser.add_argument('--live-file')
    parser.set_defaults(clients=None, neighbors=None, iterations=None)
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)
    instance = None
    try:
        if args.check_runtime:
            print(json.dumps(require_runtime()), flush=True)
            return 0
        manifest = RunManifest.read(args.manifest) if args.manifest else None
        session = identifier(args.session or (manifest.session if manifest else uuid.uuid4().hex), 'sesion')
        n, v, iterations = (args.clients if args.clients is not None else (manifest.n if manifest else 100),
            args.neighbors if args.neighbors is not None else (manifest.v if manifest else 25),
            args.iterations if args.iterations is not None else (manifest.i if manifest else 5))
        instance = ParallelServer(workers=args.workers, host=args.host, port=args.port,
            num_clients=n, neighbors_per_group=v,
            total_iterations=iterations, session_id=session, policy=policy_from(args), manifest=manifest,
            deadline=args.deadline, control_rate=args.control_rate,
            receive_buffer=args.receive_buffer, send_buffer=args.send_buffer,
            result_path=args.result or default_server_result(session), ready_path=args.ready_file,
            defer_capacity=args.defer_capacity,minimum_vm_available_mib=args.minimum_vm_available_mib,
            live_path=args.live_file)
        previous = {}
        for signum in (signal.SIGINT, signal.SIGTERM):
            previous[signum] = signal.signal(signum, lambda *_: instance.abort('señal de parada'))
        try:
            report = instance.run()
        finally:
            for signum, handler in previous.items():
                signal.signal(signum, handler)
        print(json.dumps({k:v for k,v in report.items() if k not in ('group_means_ms','telemetry')}, ensure_ascii=False), flush=True)
        return 0 if report['success'] else 1
    except (ValueError, RuntimeError, OSError) as exc:
        error = {'success': False, 'error': str(exc), 'runtime': runtime_probe(), 'role': 'server'}
        if args.result and (instance is None or instance.result is None):
            write_json(args.result, error)
        print(json.dumps(error), flush=True)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
