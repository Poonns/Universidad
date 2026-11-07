"""Fixtures y pérdidas de datagramas exclusivas de tests."""
import asyncio
from pathlib import Path
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from client import Operation, SimulationClient
from protocol import *
from run_clients import ClientGenerator
from runtime import RetryPolicy
from server import SimulationServer

SESSION = "sesion-test"
INSTANCE = "instancia-test"


class MemoryTransport:
    def __init__(self, callback=None):
        self.messages = []
        self.callback = callback

    def sendto(self, data, addr=None):
        msg = decode_message(data)
        self.messages.append((msg, addr))
        if self.callback:
            self.callback(msg, addr)


class FaultTransport:
    def __init__(self, underlying, decide):
        self.underlying, self.decide = underlying, decide

    def sendto(self, data, addr=None):
        action = self.decide(decode_message(data), addr)
        if action == "drop":
            return
        self.underlying.sendto(data, addr) if addr is not None else self.underlying.sendto(data)
        if action == "duplicate":
            self.underlying.sendto(data, addr) if addr is not None else self.underlying.sendto(data)

    def get_extra_info(self, name, default=None):
        return self.underlying.get_extra_info(name, default)

    def close(self):
        self.underlying.close()


def drop_first(*kinds):
    pending = set(kinds)
    def decide(msg, addr):
        if msg["tipo"] in pending:
            pending.remove(msg["tipo"])
            return "drop"
    return decide


def faulty_clients(decide):
    class FaultClient(SimulationClient):
        def _send(self, msg):
            if not isinstance(self.transport, FaultTransport):
                self.transport = FaultTransport(self.transport, decide)
            super()._send(msg)
    return FaultClient


def unit_client(n=3, v=3, iterations=3):
    client = SimulationClient("127.0.0.1", 9999, iterations,
        num_clients=n, neighbors_per_group=v, session_id=SESSION, token="token-1")
    client.client_id, client.group_id, client.instance_id = 1, 1, INSTANCE
    client.neighbors = members(1, v) - {1}
    client.transport = MemoryTransport()
    client.registration_sent = True
    client.registered_at = 1.0
    return client


def ack(origin=2, target=1, iteration=1, group=1, session=SESSION):
    return message(MSG_ACK, session, origin, group, destino=target, iteracion=iteration)


async def simulate(n=4, v=2, iterations=3, server_filter=None,
                   client_factory=SimulationClient, server_deadline=3, phase_timeout=2, policy=None):
    with tempfile.TemporaryDirectory(prefix="arc-udp-test-") as temporary:
        result = Path(temporary) / "servidor.json"
        policy = policy or RetryPolicy(0.04, 0.12, 6, 0)
        server = SimulationServer("127.0.0.1", 0, n, v, iterations, session_id=SESSION,
            policy=policy, deadline=server_deadline, control_rate=10000, result_path=result)
        server_task = asyncio.create_task(server.start())
        generator = None
        try:
            await asyncio.wait_for(server.ready_event.wait(), 1)
            if server_filter:
                server.transport = FaultTransport(server.transport, server_filter)
            generator = ClientGenerator(*server.address, n, v, iterations,
                session_id=SESSION, server_result=result, policy=policy, timeout=phase_timeout,
                deadline=server_deadline + 1, request_rate=5000, retry_rate=1000,
                client_factory=client_factory)
            await generator.run()
            await server_task
            return server, generator
        finally:
            if server_task is not None and not server_task.done():
                server_task.cancel()
            await asyncio.gather(server_task, return_exceptions=True)


def certify_snapshot(server, generator, codes, n, workers):
    errors = []
    if not server or not generator:
        return False, False, ['faltan resultados S/G']
    if codes != [0, 0] or not server.get('success') or not generator.get('success'):
        errors.append('algún proceso no completó correctamente')
    for key in ('registered_clients','started_clients','completed_clients','group_done_acks'):
        if server.get(key) != n:
            errors.append('S: ' + key)
    for key in ('registered_clients','started_clients','completed_clients'):
        if generator.get(key) != n:
            errors.append('G: ' + key)
    for side in (server, generator):
        if (side.get('n'), side.get('v'), side.get('i')) != (n,25,5):
            errors.append('N/V/I distintos del experimento')
    if not generator.get('server_certificate_validated') or generator.get('completed_iterations') != n*5:
        errors.append('G: certificado o iteraciones incompletos')
    if generator.get('counters',{}).get('ack_unique') != n*5*24:
        errors.append('G: ACK únicos incompletos')
    if not server.get('real_parallelism_eligible'):
        errors.append('S: runtime no elegible como experimento sin GIL')
    for point in (server.get('runtime_initial',{}),server.get('runtime_final',{})):
        if not point.get('free_threaded_build') or point.get('gil_enabled') is not False:
            errors.append('GIL activo o build convencional')
    rows = server.get('workers', [])
    if (len(rows) != workers or len({r['pid'] for r in rows}) != 1
            or len({r['native_thread_id'] for r in rows}) != workers
            or len({r['loop_id'] for r in rows}) != workers
            or len({r['socket_inode'] for r in rows}) != workers):
        errors.append('no se verifican W threads/loops/sockets dentro de un proceso')
    if server.get('telemetry',{}).get('namespace_processes_peak') != 1:
        errors.append('más de un proceso en el contenedor')
    publication = server.get('routing_publication',{})
    if not publication.get('published_before_workers') or not publication.get('all_workers_ready_before_start'):
        errors.append('transición registro/workers/INICIO inválida')
    sc, gc = server.get('counters',{}), generator.get('counters',{})
    for side in (sc,gc):
        if any(side.get(k,0) for k in ('protocol_errors','socket_errors','failures')):
            errors.append('errores de aplicación/protocolo/socket')
    if server.get('session') != generator.get('session'):
        errors.append('sesiones diferentes')
    stable = not errors and not any(side.get(k,0) for side in (sc,gc)
        for k in ('retransmissions','timeouts')) and not any(
            r.get('socket_observed',{}).get('drops',0) or r.get('startup_deferred_drops',0) for r in rows)
    if stable:
        if (sc.get('datagrams_sent') != gc.get('datagrams_received')
                or sc.get('datagrams_received') != gc.get('datagrams_sent')
                or sc.get('datagrams_sent',0) + gc.get('datagrams_sent',0) != 373*n):
            errors.append('balance de datagramas distinto al mínimo esperado sin pérdidas')
            stable = False
    return not errors, stable, errors

