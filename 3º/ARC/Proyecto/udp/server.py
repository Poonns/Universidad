<<<<<<< HEAD
"""Servidor UDP central: registro, routing y control fiable en un único event loop."""
from __future__ import annotations

import argparse
import asyncio
from dataclasses import dataclass
import heapq
import json
import math
import random
import time
from typing import Optional
import uuid

from protocol import (MSG_ACK, MSG_ERROR, MSG_GROUP_DONE, MSG_GROUP_DONE_ACK,
    MSG_METRICS, MSG_METRICS_ACK, MSG_POSITION, MSG_REGISTER, MSG_REGISTER_ACK,
    MSG_START, MSG_START_ACK, ProtocolError, decode_message, encode_message,
    group_for, identifier, members, message, validate_configuration)
from runtime import (Counters, RateLimiter, RetryPolicy, add_common_arguments,
    configure_socket, default_server_result, policy_from, positive, write_json)


class UDPServerProtocol(asyncio.DatagramProtocol):
    def __init__(self, server):
        self.server = server

    def connection_made(self, transport):
        self.server.transport = transport

    def datagram_received(self, data, addr):
        self.server.handle_datagram(data, addr)

    def error_received(self, exc):
        self.server.stats.socket_errors += 1
        self.server.fail(f"socket UDP: {exc}")

    def connection_lost(self, exc):
        self.server.closed_event.set()


@dataclass
class Control:
    kind: str
    client_id: int
    due: float
    attempts: int = 0


class SimulationServer:
    def __init__(self, host: str, port: int, num_clients: int,
                 neighbors_per_group: int, total_iterations: int = 5, *,
                 session_id: Optional[str] = None, policy: Optional[RetryPolicy] = None,
                 deadline: float = 120, control_rate: float = 2000,
                 receive_buffer: int = 256 * 1024, send_buffer: int = 256 * 1024,
                 result_path=None, ready_path=None, benchmark=False, verbose=False):
        validate_configuration(num_clients, neighbors_per_group, total_iterations)
        positive(deadline, "deadline")
        if not 0 <= port <= 65535:
            raise ValueError("puerto fuera de rango")
        if min(receive_buffer, send_buffer) < 4096:
            raise ValueError("buffers UDP: mínimo 4096 bytes")
        self.host, self.port = host, port
        self.num_clients, self.neighbors_per_group = num_clients, neighbors_per_group
        self.total_iterations = total_iterations
        self.num_groups = num_clients // neighbors_per_group
        self.session_id = identifier(session_id or uuid.uuid4().hex, "sesion")
        self.instance_id = uuid.uuid4().hex
        self.policy = policy or RetryPolicy()
        self.deadline = deadline
        self.control_limiter = RateLimiter(control_rate)
        self.receive_buffer, self.send_buffer = receive_buffer, send_buffer
        self.result_path, self.ready_path = result_path, ready_path
        self.benchmark, self.verbose = benchmark, verbose and not benchmark
        self.clients = {}
        self.addr_to_id = {}
        self.registration_keys = {}
        self.groups = {group: [] for group in range(1, self.num_groups + 1)}
        self.metrics = {}
        self.started_clients = set()
        self.finished_groups = set()
        self.done_acks = set()
        self.pending_controls = {}
        self.control_heap = []
        self.control_serial = 0
        self.control_changed = asyncio.Event()
        self.all_metrics_event = asyncio.Event()
        self.complete_event = asyncio.Event()
        self.closed_event = asyncio.Event()
        self.ready_event = asyncio.Event()
        self.transport = None
        self.control_task = None
        self.stats = Counters()
        self.buffers = {}
        self.error = None
        self.success = False
        self.created_at = time.perf_counter()
        self.first_registration_at = None
        self.registered_at = None
        self.simulation_at = None
        self.last_metrics_at = None
        self.finished_at = None
        self.result = None

    @property
    def address(self):
        return self.transport.get_extra_info("sockname") if self.transport else (self.host, self.port)

    def fail(self, reason):
        if self.error is None:
            self.error = reason
            self.stats.failures += 1
        self.complete_event.set()
        self.control_changed.set()

    def _send_raw(self, data, addr, kind):
        try:
            self.transport.sendto(data, addr)
            self.stats.sent(kind)
        except OSError as exc:
            self.stats.socket_errors += 1
            self.fail(f"sendto: {exc}")

    def _send(self, msg, addr):
        self._send_raw(encode_message(msg), addr, msg["tipo"])

    def _control_message(self, kind, cid):
        info = self.clients[cid]
        return message(kind, self.session_id, group=info["group_id"],
                       destino=cid, token=info["token"], instancia=self.instance_id)

    def _schedule_control(self, kind, cid):
        key = (kind, cid)
        if key in self.pending_controls:
            return
        control = Control(kind, cid, asyncio.get_running_loop().time())
        self.pending_controls[key] = control
        self._push_control(control)

    def _push_control(self, control):
        self.control_serial += 1
        heapq.heappush(self.control_heap, (control.due, self.control_serial, control))
        self.control_changed.set()

    async def _control_loop(self):
        rng = random.Random(42)
        loop = asyncio.get_running_loop()
        while not self.complete_event.is_set():
            while self.control_heap and self.control_heap[0][0] <= loop.time():
                due, _, control = heapq.heappop(self.control_heap)
                key = (control.kind, control.client_id)
                if self.pending_controls.get(key) is not control or control.due != due:
                    continue
                if control.attempts:
                    self.stats.timeouts += 1
                if control.attempts >= self.policy.max_attempts:
                    self.fail(f"{control.kind} sin confirmar para cliente {control.client_id}")
                    return
                await self.control_limiter.acquire()
                if self.pending_controls.get(key) is not control:
                    continue  # ACK recibido durante pacing: no enviar trabajo obsoleto.
                if control.attempts:
                    self.stats.retransmissions += 1
                if control.kind == MSG_START:
                    self.clients[control.client_id]["start_sent"] = True
                    if self.simulation_at is None:
                        self.simulation_at = time.perf_counter()
                self._send(self._control_message(control.kind, control.client_id),
                           self.clients[control.client_id]["addr"])
                control.attempts += 1
                control.due = loop.time() + self.policy.delay(control.attempts, rng)
                self._push_control(control)
            self.control_changed.clear()
            if self.complete_event.is_set():
                return
            if self.control_heap:
                delay = max(0, self.control_heap[0][0] - loop.time())
                try:
                    await asyncio.wait_for(self.control_changed.wait(), delay)
                except asyncio.TimeoutError:
                    pass
            else:
                await self.control_changed.wait()

    def _register(self, msg, addr):
        body = msg["datos"]
        if msg["origen"] != 0 or msg["grupo"] != 0:
            raise ProtocolError("REGISTRO debe solicitar identidad con origen/grupo cero")
        if (body["num_clientes"], body["num_vecinos"], body["total_iteraciones"]) != (
                self.num_clients, self.neighbors_per_group, self.total_iterations):
            self._send(message(MSG_ERROR, self.session_id, token=body["token"],
                               mensaje="N/V/I incompatibles con el servidor"), addr)
            raise ProtocolError("N/V/I incompatibles")
        key = (body["generador"], body["token"])
        if key in self.registration_keys:
            cid = self.registration_keys[key]
            if self.clients[cid]["addr"] != addr:
                raise ProtocolError("registro intenta cambiar el endpoint")
            self.stats.registration_duplicates += 1
        else:
            if addr in self.addr_to_id or len(self.clients) >= self.num_clients:
                raise ProtocolError("endpoint ocupado o cupo de registro completo")
            cid = len(self.clients) + 1
            group = group_for(cid, self.neighbors_per_group)
            self.clients[cid] = {"addr": addr, "group_id": group,
                                 "token": body["token"], "generator": body["generador"],
                                 "start_sent": False}
            self.addr_to_id[addr] = cid
            self.registration_keys[key] = cid
            self.groups[group].append(cid)
            if self.first_registration_at is None:
                self.first_registration_at = time.perf_counter()
        group = self.clients[cid]["group_id"]
        self._send(message(MSG_REGISTER_ACK, self.session_id, group=group,
                   cliente_id=cid, grupo_id=group, num_vecinos=self.neighbors_per_group,
                   num_clientes=self.num_clients, total_iteraciones=self.total_iterations,
                   token=body["token"], instancia=self.instance_id), addr)
        if len(self.clients) == self.num_clients and self.registered_at is None:
            self.registered_at = time.perf_counter()
            for target in self.clients:
                self._schedule_control(MSG_START, target)

    def _sender(self, msg, addr):
        cid = msg["origen"]
        info = self.clients.get(cid)
        if info is None or info["addr"] != addr or info["group_id"] != msg["grupo"]:
            raise ProtocolError("origen, grupo o endpoint no registrados")
        return cid, info

    def _maybe_finish_group(self, group):
        if group in self.finished_groups:
            return
        ids = self.groups[group]
        if len(ids) == self.neighbors_per_group and all(
                cid in self.metrics and cid in self.started_clients for cid in ids):
            self.finished_groups.add(group)
            for cid in ids:
                self._schedule_control(MSG_GROUP_DONE, cid)

    def handle_datagram(self, data, addr):
        self.stats.datagrams_received += 1
        try:
            msg = decode_message(data)
            self.stats.received(msg["tipo"])
            if msg["sesion"] != self.session_id:
                raise ProtocolError("sesión ajena")
            kind, body = msg["tipo"], msg["datos"]
            if kind == MSG_REGISTER:
                self._register(msg, addr)
                return
            cid, info = self._sender(msg, addr)
            group = info["group_id"]
            if self.registered_at is None or (kind != MSG_ACK and not info["start_sent"]):
                raise ProtocolError("mensaje anterior a INICIO")
            if kind in (MSG_START_ACK, MSG_GROUP_DONE_ACK):
                if body["instancia"] != self.instance_id:
                    raise ProtocolError("confirmación de otra instancia del servidor")
                if kind == MSG_START_ACK:
                    self.started_clients.add(cid)
                    self.pending_controls.pop((MSG_START, cid), None)
                    self._maybe_finish_group(group)
                else:
                    if group not in self.finished_groups:
                        raise ProtocolError("GRUPO_FIN_ACK prematuro")
                    self.done_acks.add(cid)
                    self.pending_controls.pop((MSG_GROUP_DONE, cid), None)
                    if (len(self.done_acks) == self.num_clients
                            and len(self.metrics) == self.num_clients
                            and len(self.started_clients) == self.num_clients):
                        self.complete_event.set()
                self.control_changed.set()
            elif kind == MSG_POSITION:
                if not 1 <= body["iteracion"] <= self.total_iterations:
                    raise ProtocolError("iteración fuera de rango")
                neighbors = members(group, self.neighbors_per_group) - {cid}
                targets = set(body.get("pendientes", neighbors))
                if not targets <= neighbors:
                    raise ProtocolError("destinos fuera del grupo o incluyen al propio cliente")
                # Routing sin guardar POSICION ni ejecutar iteraciones del cliente.
                for target in sorted(targets):
                    self._send_raw(data, self.clients[target]["addr"], MSG_POSITION)
            elif kind == MSG_ACK:
                target = body["destino"]
                if (target == cid or target not in self.clients
                        or self.clients[target]["group_id"] != group
                        or not 1 <= body["iteracion"] <= self.total_iterations):
                    raise ProtocolError("ACK con destino, grupo o iteración inválidos")
                self._send_raw(data, self.clients[target]["addr"], MSG_ACK)
            elif kind == MSG_METRICS:
                if (body["total_iteraciones"] != self.total_iterations
                        or body["acks_unicos"] != self.total_iterations * (self.neighbors_per_group - 1)):
                    raise ProtocolError("métricas de una simulación incompleta")
                metric = body["latencia_media_ms"]
                if cid in self.metrics:
                    if self.metrics[cid] != metric:
                        raise ProtocolError("métricas duplicadas contradictorias")
                    self.stats.metrics_duplicates += 1
                else:
                    self.metrics[cid] = metric
                    if len(self.metrics) == self.num_clients:
                        self.last_metrics_at = time.perf_counter()
                        self.all_metrics_event.set()
                self._send(self._control_message(MSG_METRICS_ACK, cid), info["addr"])
                self._maybe_finish_group(group)
            else:
                raise ProtocolError("tipo no admitido hacia el servidor")
        except ProtocolError as exc:
            self.stats.protocol_errors += 1
            if self.verbose:
                print(f"[SERVIDOR] Datagrama rechazado: {exc}", flush=True)

    async def listen(self):
        loop = asyncio.get_running_loop()
        transport, _ = await loop.create_datagram_endpoint(
            lambda: UDPServerProtocol(self), local_addr=(self.host, self.port))
        self.transport = transport
        try:
            self.buffers = configure_socket(transport, self.receive_buffer, self.send_buffer)
        except Exception:
            transport.close()
            await self.closed_event.wait()
            raise
        self.control_task = asyncio.create_task(self._control_loop())
        self.ready_event.set()
        ready = {"event": "ready", "host": self.address[0], "port": self.address[1],
                 "session": self.session_id, "instance": self.instance_id}
        if self.ready_path:
            write_json(self.ready_path, ready)
        print(json.dumps(ready), flush=True)

    async def close(self):
        if self.control_task is not None:
            self.control_task.cancel()
            await asyncio.gather(self.control_task, return_exceptions=True)
        if self.transport is not None:
            self.transport.close()
            await self.closed_event.wait()

    def report(self):
        end = self.finished_at or time.perf_counter()
        group_means = {str(group): math.fsum(self.metrics[cid] for cid in ids) / len(ids)
                       for group, ids in self.groups.items()
                       if ids and all(cid in self.metrics for cid in ids)}
        return {"role": "server", "session": self.session_id, "instance": self.instance_id,
                "n": self.num_clients, "v": self.neighbors_per_group, "i": self.total_iterations,
                "success": self.success, "simulation_complete": self.success,
                "error": self.error, "registered_clients": len(self.clients),
                "started_clients": len(self.started_clients), "completed_clients": len(self.metrics),
                "completed_iterations": len(self.metrics) * self.total_iterations,
                "group_done_acks": len(self.done_acks), "groups_finished": len(self.finished_groups),
                "registration_seconds": (self.registered_at - self.first_registration_at)
                    if self.registered_at is not None else None,
                "simulation_seconds": (self.last_metrics_at - self.simulation_at)
                    if self.last_metrics_at is not None and self.simulation_at is not None else None,
                "total_seconds": end - self.created_at,
                "latency_mean_ms": math.fsum(self.metrics.values()) / len(self.metrics) if self.metrics else None,
                "group_means_ms": group_means, "counters": self.stats.report(), "buffers": self.buffers,
                "diagnostics": {
                    "unregistered_clients": self.num_clients-len(self.clients),
                    "missing_metrics_client_ids": sorted(set(self.clients)-self.metrics.keys()),
                    "missing_start_ack_client_ids": sorted(set(self.clients)-self.started_clients),
                    "missing_group_done_ack_client_ids": sorted(set(self.clients)-self.done_acks),
                    "metrics_expected": self.num_clients, "metrics_received": len(self.metrics)}}

    async def start(self):
        try:
            await self.listen()
            await asyncio.wait_for(self.complete_event.wait(), self.deadline)
            self.success = (self.error is None and len(self.done_acks) == self.num_clients
                            and len(self.metrics) == self.num_clients)
        except asyncio.TimeoutError:
            self.stats.timeouts += 1
            self.fail("deadline administrativo agotado")
        except asyncio.CancelledError:
            self.fail("ejecución cancelada")
            raise
        finally:
            await self.close()
            self.finished_at = time.perf_counter()
            self.result = self.report()
            if self.result_path:
                write_json(self.result_path, self.result)
        return self.success

    def display_summary(self):
        report = self.result or self.report()
        if self.benchmark:
            # Las medias de TODOS los grupos siguen en el JSON persistido.
            report = {key: value for key, value in report.items() if key != "group_means_ms"}
        print(json.dumps(report, ensure_ascii=False), flush=True)


def parse_args(argv=None):
    parser = argparse.ArgumentParser(description="Servidor central UDP Python/asyncio")
    add_common_arguments(parser)
    parser.add_argument("--session", help="identificador único del run; UUID si se omite")
    parser.add_argument("--control-rate", type=float, default=2000)
    parser.add_argument("--receive-buffer", type=int, default=256 * 1024)
    parser.add_argument("--send-buffer", type=int, default=256 * 1024)
    parser.add_argument("--ready-file")
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)
    try:
        session = identifier(args.session or uuid.uuid4().hex, "sesion")
        async def execute():
            # Python 3.9 vincula Event/Lock al loop en su construcción.
            server = SimulationServer(args.host, args.port, args.clients, args.neighbors, args.iterations,
                session_id=session, policy=policy_from(args), deadline=args.deadline,
                control_rate=args.control_rate, receive_buffer=args.receive_buffer,
                send_buffer=args.send_buffer, result_path=args.result or default_server_result(session),
                ready_path=args.ready_file, benchmark=args.benchmark, verbose=args.verbose)
            ok = await server.start()
            server.display_summary()
            return 0 if ok else 1
        return asyncio.run(execute())
    except (ValueError, OSError) as exc:
        print(json.dumps({"success": False, "error": str(exc)}), flush=True)
        return 1
    except KeyboardInterrupt:
        return 130


if __name__ == "__main__":
    raise SystemExit(main())
=======
import argparse
import asyncio
import socket
import sys
from typing import Any, Dict, List, Optional, Tuple

from protocol import (
    MSG_ACK,
    MSG_METRICS,
    MSG_METRICS_ACK,
    MSG_POSITION,
    MSG_REGISTER,
    MSG_REGISTER_ACK,
    MSG_START,
    decode_message,
    encode_message,
    send_message,
)


class UDPServerProtocol(asyncio.DatagramProtocol):
    """Asyncio DatagramProtocol implementation for the UDP Central Simulation Server."""

    def __init__(self, server: "SimulationServer"):
        self.server = server
        self.transport: Optional[asyncio.DatagramTransport] = None

    def connection_made(self, transport: asyncio.DatagramTransport) -> None:
        self.transport = transport
        self.server.transport = transport

        # Optimize OS UDP socket buffers to prevent packet drop on high-throughput bursts
        sock = transport.get_extra_info("socket")
        if sock is not None:
            try: #tamaño en MB
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 8 * 1024 * 1024)
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 8 * 1024 * 1024)
            except OSError:
                pass

    def datagram_received(self, data: bytes, addr: Tuple[str, int]) -> None:
        self.server.handle_datagram(data, addr)

    def error_received(self, exc: Exception) -> None:
        print(f"[SERVIDOR] Error recibido en transporte UDP: {exc}", file=sys.stderr)


class SimulationServer:
    def __init__(
        self,
        host: str,
        port: int,
        num_clients: int,
        neighbors_per_group: int,
        metrics_timeout: Optional[float] = None,
    ):
        self.host = host
        self.port = port
        self.num_clients = num_clients
        self.neighbors_per_group = neighbors_per_group

        if self.num_clients % self.neighbors_per_group != 0:
            raise ValueError(
                f"El número de clientes N ({num_clients}) debe ser divisible por V ({neighbors_per_group})"
            )

        self.num_groups = self.num_clients // self.neighbors_per_group

        # Registration state: mappings between client_id, group_id, and UDP address (ip, port)
        self.clients: Dict[int, Dict[str, Any]] = {}
        self.addr_to_id: Dict[Tuple[str, int], int] = {}
        self.groups: Dict[int, List[int]] = {g: [] for g in range(1, self.num_groups + 1)}

        # Metrics collection
        self.metrics: Dict[int, float] = {}
        self.first_metric_event = asyncio.Event()
        self.all_metrics_event = asyncio.Event()

        # Timeout dinámico: base de 5s (cubre retries del cliente) + 0.05s por cliente
        if metrics_timeout is not None:
            self.metrics_timeout = metrics_timeout
        else:
            self.metrics_timeout = max(5.0, 5.0 + (self.num_clients * 0.05))

        self.transport: Optional[asyncio.DatagramTransport] = None

    def handle_datagram(self, data: bytes, addr: Tuple[str, int]) -> None:
        """Central non-blocking message router for incoming UDP datagrams."""
        try:
            msg = decode_message(data)
        except Exception as e:
            print(f"[SERVIDOR] Mensaje inválido recibido de {addr}: {e}")
            return

        msg_type = msg.get("tipo")

        # 1. Registration phase
        if msg_type == MSG_REGISTER:
            if addr in self.addr_to_id:
                # Handle possible duplicate registration datagram
                client_id = self.addr_to_id[addr]
                group_id = self.clients[client_id]["group_id"]
            else:
                client_id = len(self.clients) + 1
                group_id = (client_id - 1) // self.neighbors_per_group + 1
                self.clients[client_id] = {
                    "addr": addr,
                    "group_id": group_id,
                }
                self.addr_to_id[addr] = client_id
                self.groups[group_id].append(client_id)
                print(f"[SERVIDOR] Registrado Cliente {client_id} en Grupo {group_id} ({addr})")

            ack_msg = {
                "tipo": MSG_REGISTER_ACK,
                "origen": 0,
                "grupo": group_id,
                "datos": {
                    "cliente_id": client_id,
                    "grupo_id": group_id,
                    "num_vecinos": self.neighbors_per_group,
                },
            }
            send_message(self.transport, ack_msg, addr)

            # If all required clients are registered, broadcast START
            if len(self.clients) == self.num_clients:
                print(
                    f"[SERVIDOR] Se conectaron los {self.num_clients} clientes requeridos. "
                    "Iniciando simulación..."
                )
                for cid, info in self.clients.items():
                    start_msg = {
                        "tipo": MSG_START,
                        "origen": 0,
                        "grupo": info["group_id"],
                        "datos": {"mensaje": "Inicio de simulacion"},
                    }
                    send_message(self.transport, start_msg, info["addr"])

        # 2. Position forwarding loop
        elif msg_type == MSG_POSITION:
            src_id = msg.get("origen")
            if src_id in self.clients:
                grp = self.clients[src_id]["group_id"]
                neighbors = [nid for nid in self.groups[grp] if nid != src_id]
                # Forward raw datagram directly to all group neighbors
                for nid in neighbors:
                    target_addr = self.clients[nid]["addr"]
                    self.transport.sendto(data, target_addr)

        # 3. ACK forwarding loop
        elif msg_type == MSG_ACK:
            dest_id = msg.get("datos", {}).get("destino")
            if dest_id in self.clients:
                target_addr = self.clients[dest_id]["addr"]
                self.transport.sendto(data, target_addr)

        # 4. Metrics collection
        elif msg_type == MSG_METRICS:
            src_id = msg.get("origen")
            mean_lat = msg.get("datos", {}).get("latencia_media_ms", 0.0)
            is_first_reception = src_id not in self.metrics
            self.metrics[src_id] = mean_lat
            if is_first_reception:
                print(f"[SERVIDOR] Métricas recibidas de Cliente {src_id}: {mean_lat:.3f} ms")
                self.first_metric_event.set()

            # Enviar confirmación METRICAS_ACK al cliente (Stop-and-Wait idempotente)
            ack_msg = {
                "tipo": MSG_METRICS_ACK,
                "origen": 0,
                "grupo": self.clients.get(src_id, {}).get("group_id", 0),
                "datos": {"cliente_id": src_id},
            }
            send_message(self.transport, ack_msg, addr)

            if len(self.metrics) == self.num_clients:
                self.all_metrics_event.set()

    async def start(self) -> None: #metodo start del server
        loop = asyncio.get_running_loop()
        transport, _ = await loop.create_datagram_endpoint( #ejecucion de connection_made
            lambda: UDPServerProtocol(self),
            local_addr=(self.host, self.port),
        )
        self.transport = transport
        try:
            # Esperar a que comience la fase de métricas (arribo del primer cliente)
            await self.first_metric_event.wait()

            # Esperar al resto de los clientes con timeout de guarda dinámico
            try:
                await asyncio.wait_for(self.all_metrics_event.wait(), timeout=self.metrics_timeout)
            except asyncio.TimeoutError:
                missing = set(self.clients.keys()) - set(self.metrics.keys())
                print(
                    f"\n[SERVIDOR] ADVERTENCIA: Timeout de {self.metrics_timeout:.1f}s alcanzado esperando métricas. "
                    f"Recibidas: {len(self.metrics)}/{self.num_clients}. Clientes faltantes: {sorted(missing)}"
                )

            self.display_summary()
        finally:
            self.transport.close()

    def display_summary(self) -> None:
        print("\n" + "=" * 55)
        print("        RESUMEN FINAL DE LA SIMULACIÓN (UDP)")
        print("=" * 55)

        global_sum = 0.0
        for g_id, member_ids in self.groups.items():
            group_latencies = [self.metrics[m] for m in member_ids if m in self.metrics]
            group_avg = sum(group_latencies) / len(group_latencies) if group_latencies else 0.0
            global_sum += sum(group_latencies)
            print(f" Grupo {g_id} (Clientes {member_ids}): Latencia Media = {group_avg:.3f} ms")

        overall_avg = global_sum / len(self.metrics) if self.metrics else 0.0
        print("-" * 55)
        print(f" LATENCIA MEDIA GLOBAL ({len(self.metrics)}/{self.num_clients} clientes reportados): {overall_avg:.3f} ms")
        print("=" * 55 + "\n")


def parse_args():
    parser = argparse.ArgumentParser(
        description="Servidor UDP de Realidad Aumentada Colaborativa (RAC)"
    )
    parser.add_argument("--host", default="127.0.0.1", help="Dirección IP de escucha (default: 127.0.0.1)")
    parser.add_argument("--port", type=int, default=8888, help="Puerto UDP de escucha (default: 8888)")
    parser.add_argument("-N", "--clients", type=int, default=4, help="Número total de clientes N (default: 4)")
    parser.add_argument("-V", "--neighbors", type=int, default=4, help="Vecinos por grupo V (default: 4)")
    parser.add_argument(
        "--metrics-timeout",
        type=float,
        default=None,
        help="Timeout en segundos para esperar métricas tras la primera recibida (default: auto)",
    )
    return parser.parse_args()


if __name__ == "__main__":
    args = parse_args()
    server = SimulationServer(args.host, args.port, args.clients, args.neighbors, args.metrics_timeout)
    try:
        asyncio.run(server.start())
    except KeyboardInterrupt:
        print("\n[SERVIDOR] Detenido por el usuario.")
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
