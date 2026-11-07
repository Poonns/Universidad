"""Confirmación remota del run sobre el mismo servicio UDP, fuera de DATA."""
from __future__ import annotations

import asyncio
from dataclasses import dataclass
import threading
import time
import uuid

from protocol import (GENERATOR_REQUESTS, MSG_ERROR, MSG_GENERATOR_HELLO,
    MSG_GENERATOR_HELLO_ACK, MSG_RUN_STATUS, MSG_RUN_STATUS_ACK,
    MSG_GENERATOR_DONE, MSG_GENERATOR_DONE_ACK, ProtocolError, decode_message,
    encode_message, message)
from runtime import Counters


def validate_certificate(proof, session, instance, n, v, iterations, digest=None):
    expected = dict(role='server', session=session, instance=instance, n=n, v=v, i=iterations,
        success=True, simulation_complete=True, registered_clients=n, started_clients=n,
        completed_clients=n, completed_iterations=n*iterations, group_done_acks=n,
        groups_finished=n//v)
    if digest is not None:
        expected['manifest_sha256'] = digest
    if not isinstance(proof, dict) or any(type(proof.get(k)) is not type(value)
            or proof[k] != value for k,value in expected.items()):
        raise ProtocolError('certificado global incoherente o de otra sesión/instancia')
    return proof


@dataclass(frozen=True)
class RemoteNotice:
    kind: str
    generator: str
    body: tuple
    addr: tuple
    received_at: float


class GeneratorInbox:
    """Hasta tres solicitudes por generador; jamás transporta POSITION/ACK."""
    def __init__(self, count):
        self.limit = 3*count
        self.lock = threading.Lock()
        self.values = {}
        self.maximum = 0

    def put(self, notice):
        key = notice.kind, notice.generator
        with self.lock:
            if key not in self.values and len(self.values) >= self.limit:
                raise ProtocolError('buzón de generadores lleno')
            self.values[key] = notice
            self.maximum = max(self.maximum, len(self.values))

    def drain(self):
        with self.lock:
            rows = tuple(self.values.values())
            self.values.clear()
            return rows


class GeneratorControl(asyncio.DatagramProtocol):
    def __init__(self, manifest, spec, policy):
        self.manifest, self.spec, self.policy = manifest, spec, policy
        self.nonce = uuid.uuid4().hex
        self.instance = None
        self.sequence = 0
        self.transport = None
        self.closed, self.response_event = asyncio.Event(), asyncio.Event()
        self.response = self.expected = self.error = None
        self.stats = Counters()
        self.observed_endpoint = self.local_endpoint = None

    def connection_made(self, transport):
        self.transport = transport
        self.local_endpoint = transport.get_extra_info('sockname')

    def connection_lost(self, exc):
        self.closed.set()

    def error_received(self, exc):
        self.stats.socket_errors += 1
        self.error = 'socket de control: ' + str(exc)
        self.response_event.set()

    def datagram_received(self, data, addr):
        self.stats.datagrams_received += 1
        try:
            msg = decode_message(data)
            self.stats.received(msg['tipo'])
            body = msg['datos']
            if msg['sesion'] != self.manifest.session:
                raise ProtocolError('control remoto de otra sesión')
            if msg['tipo'] == MSG_ERROR and body['token'] == self.nonce and msg['origen'] == 0:
                self.error = body['mensaje']
                self.response_event.set()
                return
            if (body['generador'] != self.spec.id or body['nonce'] != self.nonce
                    or body['manifiesto'] != self.manifest.digest):
                raise ProtocolError('respuesta de control no corresponde al generador')
            if body['secuencia'] < self.sequence:
                return  # respuesta tardía de una consulta ya completada
            if body['secuencia'] != self.sequence or msg['tipo'] != self.expected:
                raise ProtocolError('tipo/secuencia no corresponde a la consulta')
            if self.instance is not None and body['instancia'] != self.instance:
                raise ProtocolError('la respuesta cambia de instancia')
            self.response = body
            self.response_event.set()
        except ProtocolError:
            self.stats.protocol_errors += 1

    async def exchange(self, kind):
        expected = {MSG_GENERATOR_HELLO: MSG_GENERATOR_HELLO_ACK,
                    MSG_RUN_STATUS: MSG_RUN_STATUS_ACK, MSG_GENERATOR_DONE: MSG_GENERATOR_DONE_ACK}
        self.expected = expected[kind]
        self.sequence += 1
        self.response = None
        self.response_event.clear()
        body = dict(generador=self.spec.id, nonce=self.nonce, manifiesto=self.manifest.digest,
                    secuencia=self.sequence)
        if kind == MSG_GENERATOR_HELLO:
            body.update(n=self.manifest.n, v=self.manifest.v, i=self.manifest.i, local=self.spec.clients)
        else:
            body['instancia'] = self.instance
        import random
        rng = random.Random(self.nonce + str(self.sequence))
        for attempt in range(1, self.policy.max_attempts+1):
            if attempt > 1:
                self.stats.retransmissions += 1
            self.transport.sendto(encode_message(message(kind, self.manifest.session, **body)))
            self.stats.sent(kind)
            try:
                await asyncio.wait_for(self.response_event.wait(), self.policy.delay(attempt, rng))
            except asyncio.TimeoutError:
                self.stats.timeouts += 1
                continue
            if self.error:
                raise RuntimeError(self.error)
            return self.response
        raise RuntimeError(kind + ': agotados los reintentos de control')

    async def hello(self):
        body = await self.exchange(MSG_GENERATOR_HELLO)
        self.instance = body['instancia']
        self.observed_endpoint = body.get('endpoint_observado')

    async def certificate(self):
        while True:
            body = await self.exchange(MSG_RUN_STATUS)
            if body['completo']:
                return validate_certificate(body['certificado'], self.manifest.session,
                    self.instance, self.manifest.n, self.manifest.v, self.manifest.i, self.manifest.digest)
            await asyncio.sleep(min(self.policy.max_timeout, max(.1, self.policy.timeout)))

    async def finish(self):
        await self.exchange(MSG_GENERATOR_DONE)

    async def close(self):
        if self.transport:
            self.transport.close()
            await self.closed.wait()


class RemoteRun:
    """Solo lo usa loop 0; registro/control de generadores y certificado terminal."""
    def __init__(self, owner, manifest):
        self.owner, self.manifest = owner, manifest
        self.generators = {}
        self.registered = {spec.id: 0 for spec in manifest.generators}
        self.released = set()
        self.all_released = threading.Event()
        self.completed_at = self.all_released_at = None
        self.status_pending = self.status_complete = 0

    def before_register(self, msg):
        body = msg['datos']
        spec = self.manifest.generator(body['generador'])
        if spec.id not in self.generators:
            raise ProtocolError('generador debe confirmar HOLA antes de registrar clientes')
        if (spec.id, body['token']) not in self.owner.registration_keys and self.registered[spec.id] >= spec.clients:
            raise ProtocolError('generador excede su N local autorizado')

    def handle(self, notice):
        s, m = self.owner, self.manifest
        body, kind = dict(notice.body), notice.kind
        try:
            spec = m.generator(notice.generator)
            if body['manifiesto'] != m.digest:
                raise ProtocolError('manifiestos distintos')
            entry = self.generators.get(spec.id)
            identity = (body['nonce'], notice.addr)
            if kind == MSG_GENERATOR_HELLO:
                if (body['n'], body['v'], body['i'], body['local']) != (m.n, m.v, m.i, spec.clients):
                    raise ProtocolError('N global/local o V/I distintos del manifiesto')
                if entry is not None and entry != identity:
                    raise ProtocolError('generator_id ocupado por otro nonce/endpoint')
                self.generators[spec.id] = identity
                extra = dict(endpoint_observado=list(notice.addr))
                response = MSG_GENERATOR_HELLO_ACK
            else:
                if entry != identity or body['instancia'] != s.instance_id:
                    raise ProtocolError('control de otra identidad/instancia')
                complete = s.worker.runtime.protocol_done.is_set() and s.worker.runtime.error is None
                if kind == MSG_RUN_STATUS:
                    response = MSG_RUN_STATUS_ACK
                    extra = dict(completo=complete)
                    if complete:
                        self.status_complete += 1
                        keys = ('role', 'session', 'instance', 'n', 'v', 'i', 'registered_clients',
                            'started_clients', 'completed_clients', 'completed_iterations',
                            'group_done_acks', 'groups_finished')
                        report = s.report()
                        extra['certificado'] = {k: report[k] for k in keys}
                        extra['certificado'].update(success=True, simulation_complete=True, manifest_sha256=m.digest)
                    else:
                        self.status_pending += 1
                elif kind == MSG_GENERATOR_DONE:
                    if not complete:
                        raise ProtocolError('GENERADOR_FIN anterior al certificado global')
                    self.released.add(spec.id)
                    if len(self.released) == len(m.generators) and self.all_released_at is None:
                        self.all_released_at = time.perf_counter()
                        self.all_released.set()
                    response, extra = MSG_GENERATOR_DONE_ACK, {}
                else:
                    raise ProtocolError('control de generador desconocido')
            common = dict(generador=spec.id, nonce=body['nonce'], manifiesto=m.digest,
                          secuencia=body['secuencia'], instancia=s.instance_id)
            s._send(message(response, m.session, **common, **extra), notice.addr)
        except (ProtocolError, ValueError) as exc:
            s.stats.protocol_errors += 1
            s._send(message(MSG_ERROR, m.session, token=body['nonce'], mensaje=str(exc)), notice.addr)

    def report(self):
        s = self.owner
        mixed = []
        distribution = {}
        for group, ids in s.groups.items():
            names = sorted({s.clients[cid]['generator'] for cid in ids})
            distribution[str(group)] = names
            if len(names) > 1:
                mixed.append(group)
        return dict(manifest_sha256=self.manifest.digest, registered_per_generator=self.registered,
            expected_per_generator={r.id:r.clients for r in self.manifest.generators},
            hello_endpoints={key:list(value[1]) for key,value in self.generators.items()},
            released_generators=sorted(self.released), mixed_groups=len(mixed),
            group_generators=distribution, status_pending=self.status_pending,
            status_complete=self.status_complete,
            endpoints=[dict(id=cid, generator=info['generator'], group=info['group_id'],
                            endpoint=list(info['addr'])) for cid,info in s.clients.items()])
