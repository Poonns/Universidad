<<<<<<< HEAD
"""Cliente virtual UDP: máquina de estados asyncio y ACK por vecino/iteración."""
from __future__ import annotations

import asyncio
from dataclasses import dataclass, field
import math
import random
import time
from typing import Optional
import uuid

from protocol import (MSG_ACK, MSG_ERROR, MSG_GROUP_DONE, MSG_GROUP_DONE_ACK,
    MSG_METRICS, MSG_METRICS_ACK, MSG_POSITION, MSG_REGISTER, MSG_REGISTER_ACK,
    MSG_START, MSG_START_ACK, ProtocolError, decode_message, encode_message,
    group_for, identifier, members, message, validate_configuration)
from runtime import Counters, RateLimiter, RetryExhausted, RetryPolicy, configure_socket, positive


class UDPClientProtocol(asyncio.DatagramProtocol):
    def __init__(self, client):
        self.client = client

    def connection_made(self, transport):
        self.client.transport = transport

    def datagram_received(self, data, addr):
        self.client.handle_datagram(data, addr)

    def error_received(self, exc):
        self.client.stats.socket_errors += 1
        self.client.fail(f"socket UDP: {exc}")

    def connection_lost(self, exc):
        self.client.closed_event.set()


@dataclass
class Operation:
    iteration: int
    coords: list
    ack_sources: set = field(default_factory=set)
    complete: asyncio.Event = field(default_factory=asyncio.Event)
    sent_at: Optional[float] = None
    last_ack_at: Optional[float] = None


class SimulationClient:
    def __init__(self, host: str, port: int, total_iterations: int = 5,
                 timeout: float = 60, *, num_clients: int = 100,
                 neighbors_per_group: int = 25, session_id: str,
                 generator_id: str = "generador-1", token: Optional[str] = None,
                 policy: Optional[RetryPolicy] = None, request_limiter=None,
                 retry_limiter=None, position_window=None, registration_window=None, registration_limiter=None,
                 local_host: str = "0.0.0.0", receive_buffer: int = 64 * 1024,
                 send_buffer: int = 64 * 1024, verbose: bool = False):
        validate_configuration(num_clients, neighbors_per_group, total_iterations)
        positive(timeout, "timeout")
        if not 1 <= port <= 65535 or min(receive_buffer, send_buffer) < 4096:
            raise ValueError("puerto o buffers UDP inválidos")
        self.host, self.port = host, port
        self.total_iterations, self.timeout = total_iterations, timeout
        self.num_clients, self.num_neighbors = num_clients, neighbors_per_group
        self.session_id = identifier(session_id, "sesion")
        self.generator_id = identifier(generator_id, "generador")
        self.token = identifier(token or uuid.uuid4().hex, "token")
        self.policy = policy or RetryPolicy()
        self.request_limiter = request_limiter or RateLimiter(2000)
        self.retry_limiter = retry_limiter or RateLimiter(100)
        self.registration_limiter = registration_limiter
        self.position_window = position_window or asyncio.Semaphore(64)
        self.registration_window = registration_window or asyncio.Semaphore(64)
        self.local_host = local_host
        self.receive_buffer, self.send_buffer = receive_buffer, send_buffer
        self.verbose = verbose
        self.client_id = self.group_id = 0
        self.instance_id = None
        self.neighbors = set()
        self.transport = None
        self.peer = (host, port)
        self.registered_event = asyncio.Event()
        self.start_event = asyncio.Event()
        self.metrics_ack_event = asyncio.Event()
        self.metrics_ack_received = False
        self.group_done_event = asyncio.Event()
        self.closed_event = asyncio.Event()
        self.operations = {}  # Iteraciones propias, también las ya completadas.
        self.seen_positions = {}  # (origen, iteración) -> coordenadas inmutables.
        self.pending_start = None
        self.pending_positions = {}
        self.latencies = []
        self.latency_sum_ms = 0.0
        self.mean_latency_ms = 0.0
        self.completed_iterations = 0
        self.stats = Counters()
        self.error = None
        self.success = False
        self.buffers = {}
        self.local_endpoint = None
        self.registration_sent = False
        self.metrics_sent = False
        self.registered_at = self.simulation_at = self.iterations_finished_at = None
        self.rng = random.Random(self.token)

    def fail(self, reason):
        if self.error is None:
            self.error = reason
        for event in (self.registered_event, self.start_event,
                      self.metrics_ack_event, self.group_done_event):
            event.set()
        for operation in self.operations.values():
            operation.complete.set()

    def _check_failure(self):
        if self.error is not None:
            raise RuntimeError(self.error)

    def _send(self, msg):
        try:
            self.transport.sendto(encode_message(msg))
            self.stats.sent(msg["tipo"])
        except OSError as exc:
            self.stats.socket_errors += 1
            self.fail(f"sendto: {exc}")

    def _server_control(self, msg):
        body = msg["datos"]
        if (msg["origen"] != 0 or msg["grupo"] != self.group_id
                or body["destino"] != self.client_id or body["token"] != self.token
                or body["instancia"] != self.instance_id):
            raise ProtocolError("control de otro cliente, grupo o instancia")

    def _register_ack(self, msg):
        body = msg["datos"]
        cid, group = body["cliente_id"], body["grupo_id"]
        if (not self.registration_sent or body["token"] != self.token or msg["origen"] != 0
                or not 1 <= cid <= self.num_clients or group != msg["grupo"]
                or group_for(cid, self.num_neighbors) != group
                or (body["num_clientes"], body["num_vecinos"], body["total_iteraciones"]) != (
                    self.num_clients, self.num_neighbors, self.total_iterations)):
            raise ProtocolError("REGISTRO_ACK no corresponde a la solicitud")
        if self.client_id and (cid != self.client_id or group != self.group_id
                               or body["instancia"] != self.instance_id):
            raise ProtocolError("REGISTRO_ACK cambia una identidad ya asignada")
        self.client_id, self.group_id, self.instance_id = cid, group, body["instancia"]
        self.neighbors = members(group, self.num_neighbors) - {cid}
        if self.registered_at is None:
            self.registered_at = time.perf_counter()
            if self.verbose:
                print(f"[CLIENTE {cid}] Registrado en grupo {group}", flush=True)
        self.registered_event.set()
        if self.pending_start is not None:
            pending, self.pending_start = self.pending_start, None
            self._start(pending)
        pending, self.pending_positions = self.pending_positions, {}
        for position in pending.values():
            self._position(position)

    def _start(self, msg):
        body = msg["datos"]
        if not self.client_id:
            if (msg["origen"] != 0 or body["token"] != self.token
                    or not 1 <= body["destino"] <= self.num_clients
                    or group_for(body["destino"], self.num_neighbors) != msg["grupo"]):
                raise ProtocolError("INICIO anticipado inválido")
            self.pending_start = msg
            return  # No registrar ni confirmar hasta recibir REGISTRO_ACK.
        self._server_control(msg)
        self._send(message(MSG_START_ACK, self.session_id, self.client_id, self.group_id,
                           instancia=self.instance_id))
        if not self.start_event.is_set():
            self.simulation_at = time.perf_counter()
            if self.verbose:
                print(f"[CLIENTE {self.client_id}] INICIO recibido", flush=True)
        self.start_event.set()

    def _position(self, msg):
        origin, body = msg["origen"], msg["datos"]
        iteration = body["iteracion"]
        if not (1 <= origin <= self.num_clients and 1 <= iteration <= self.total_iterations
                and group_for(origin, self.num_neighbors) == msg["grupo"]):
            raise ProtocolError("POSICION con identidad o iteración inválida")
        if not self.client_id:
            key = (origin, iteration)
            if len(self.pending_positions) < (self.num_neighbors - 1) * self.total_iterations:
                self.pending_positions[key] = msg
            return
        if origin not in self.neighbors or msg["grupo"] != self.group_id:
            raise ProtocolError("POSICION de otro grupo o del propio cliente")
        if "pendientes" in body and self.client_id not in body["pendientes"]:
            raise ProtocolError("POSICION selectiva no dirigida a este cliente")
        key, coords = (origin, iteration), tuple(body["coords"])
        if key in self.seen_positions:
            if self.seen_positions[key] != coords:
                raise ProtocolError("POSICION duplicada con coordenadas contradictorias")
            self.stats.position_duplicates += 1
        else:
            self.seen_positions[key] = coords
            self.stats.positions_unique += 1
        # Siempre ACKear duplicados, independientemente de la iteración propia.
        self._send(message(MSG_ACK, self.session_id, self.client_id, self.group_id,
                           destino=origin, iteracion=iteration))
        if self.verbose:
            print(f"[CLIENTE {self.client_id}] POSICION de {origin}, iteración {iteration}: ACK", flush=True)

    def _ack(self, msg):
        body, source = msg["datos"], msg["origen"]
        operation = self.operations.get(body["iteracion"])
        if (not self.client_id or source not in self.neighbors or msg["grupo"] != self.group_id
                or body["destino"] != self.client_id or operation is None
                or operation.sent_at is None):
            raise ProtocolError("ACK sin operación enviada o con origen/destino/grupo inválidos")
        if operation.iteration <= self.completed_iterations:
            self.stats.late_acks += 1
        if source in operation.ack_sources:
            self.stats.ack_duplicates += 1
            return
        operation.ack_sources.add(source)
        self.stats.ack_unique += 1
        if operation.ack_sources == self.neighbors:
            operation.last_ack_at = time.perf_counter()
            operation.complete.set()

    def handle_datagram(self, data, addr=None):
        self.stats.datagrams_received += 1
        try:
            if addr is not None and addr != self.peer:
                raise ProtocolError("datagrama de un endpoint diferente del servidor")
            msg = decode_message(data)
            self.stats.received(msg["tipo"])
            if msg["sesion"] != self.session_id:
                raise ProtocolError("sesión ajena")
            kind = msg["tipo"]
            if kind == MSG_REGISTER_ACK:
                self._register_ack(msg)
            elif kind == MSG_START:
                self._start(msg)
            elif kind == MSG_POSITION:
                self._position(msg)
            elif kind == MSG_ACK:
                self._ack(msg)
            elif kind in (MSG_METRICS_ACK, MSG_GROUP_DONE):
                self._server_control(msg)
                if not self.metrics_sent or self.completed_iterations != self.total_iterations:
                    raise ProtocolError("finalización anterior a completar y enviar métricas")
                if kind == MSG_METRICS_ACK:
                    self.metrics_ack_received = True
                    self.metrics_ack_event.set()
                if kind == MSG_GROUP_DONE:
                    self.group_done_event.set()
                # Si GRUPO_FIN adelanta a METRICAS_ACK, retener la confirmación
                # de cierre: S debe seguir disponible para recuperar ese ACK.
                if self.metrics_ack_received and self.group_done_event.is_set():
                    self._send(message(MSG_GROUP_DONE_ACK, self.session_id,
                                       self.client_id, self.group_id, instancia=self.instance_id))
            elif kind == MSG_ERROR:
                if msg["origen"] != 0 or msg["datos"]["token"] != self.token:
                    raise ProtocolError("ERROR no dirigido al cliente")
                self.fail(msg["datos"]["mensaje"])
            else:
                raise ProtocolError("tipo no admitido hacia el cliente")
        except ProtocolError as exc:
            self.stats.protocol_errors += 1
            if self.verbose:
                print(f"[CLIENTE {self.client_id}] Datagrama rechazado: {exc}", flush=True)

    async def _retry_permit(self, limiter, event, deadline, phase):
        """Un ACK tardío cancela también la espera de tokens, no solo el send."""
        permit = asyncio.create_task(limiter.acquire())
        response = asyncio.create_task(event.wait())
        try:
            remaining = max(0, deadline - asyncio.get_running_loop().time())
            done, _ = await asyncio.wait((permit, response), timeout=remaining,
                                         return_when=asyncio.FIRST_COMPLETED)
            if not done:
                self.stats.timeouts += 1
                raise RetryExhausted(f"{phase}: timeout en admisión de retry")
            self._check_failure()
            if event.is_set():
                return False
            await permit
            return True
        finally:
            for task in (permit, response):
                if not task.done():
                    task.cancel()
            await asyncio.gather(permit, response, return_exceptions=True)

    async def _exchange(self, make_message, event, phase, first_send=None):
        loop = asyncio.get_running_loop()
        deadline = loop.time() + self.timeout
        for attempt in range(1, self.policy.max_attempts + 1):
            self._check_failure()
            if event.is_set():
                return
            limiters = ([self.retry_limiter] if attempt > 1 else [])
            if phase == 'registro' and self.registration_limiter is not None:
                limiters.append(self.registration_limiter)
            limiters.append(self.request_limiter)
            for limiter in limiters:
                remaining = deadline - loop.time()
                if remaining <= 0:
                    self.stats.timeouts += 1
                    raise RetryExhausted(f"{phase}: timeout de fase en admisión")
                if attempt > 1:
                    if not await self._retry_permit(limiter, event, deadline, phase):
                        return
                else:
                    try:
                        await asyncio.wait_for(limiter.acquire(), remaining)
                    except asyncio.TimeoutError as exc:
                        self.stats.timeouts += 1
                        raise RetryExhausted(f"{phase}: timeout de fase en admisión") from exc
            self._check_failure()
            if event.is_set():
                return  # Un ACK tardío puede cancelar una retransmisión esperando tokens.
            if attempt == 1 and first_send is not None:
                first_send()
            if attempt > 1:
                self.stats.retransmissions += 1
            self._send(make_message())
            self._check_failure()
            remaining = deadline - loop.time()
            if remaining <= 0:
                self.stats.timeouts += 1
                raise RetryExhausted(f"{phase}: timeout de fase")
            try:
                await asyncio.wait_for(event.wait(), min(self.policy.delay(attempt, self.rng), remaining))
                self._check_failure()
                return
            except asyncio.TimeoutError:
                self.stats.timeouts += 1
        raise RetryExhausted(f"{phase}: agotados {self.policy.max_attempts} envíos")

    async def _wait(self, event, phase):
        try:
            await asyncio.wait_for(event.wait(), self.timeout)
            self._check_failure()
        except asyncio.TimeoutError as exc:
            self.stats.timeouts += 1
            raise RetryExhausted(f"{phase}: timeout de fase") from exc

    async def before_position(self, iteration):
        """Hook sin espera en producción; permite probar clientes deliberadamente lentos."""

    async def run(self, keep_open=False):
        try:
            loop = asyncio.get_running_loop()
            transport, _ = await loop.create_datagram_endpoint(lambda: UDPClientProtocol(self),
                local_addr=(self.local_host, 0), remote_addr=(self.host, self.port))
            self.transport = transport
            self.peer = transport.get_extra_info("peername")
            self.local_endpoint = transport.get_extra_info('sockname')
            self.buffers = configure_socket(transport, self.receive_buffer, self.send_buffer)
            def registered_send():
                self.registration_sent = True
            async with self.registration_window:
                await self._exchange(lambda: message(MSG_REGISTER, self.session_id,
                    token=self.token, generador=self.generator_id, num_clientes=self.num_clients,
                    num_vecinos=self.num_neighbors, total_iteraciones=self.total_iterations),
                    self.registered_event, "registro", registered_send)
            await self._wait(self.start_event, "inicio")
            for iteration in range(1, self.total_iterations + 1):
                await self.before_position(iteration)
                async with self.position_window:
                    operation = Operation(iteration, [])
                    self.operations[iteration] = operation
                    def initial_send():
                        operation.sent_at = time.perf_counter()
                        operation.coords = [10.0 * iteration, 20.0 * iteration, 0.0]
                    def position():
                        body = {"iteracion": iteration, "coords": operation.coords}
                        if operation.sent_at is not None and operation.ack_sources:
                            missing = self.neighbors - operation.ack_sources
                            if missing:
                                body["pendientes"] = sorted(missing)
                        return message(MSG_POSITION, self.session_id, self.client_id, self.group_id, **body)
                    # V=1: no vecinos; se completa después del envío inicial, nunca antes.
                    if not self.neighbors:
                        await asyncio.wait_for(self.request_limiter.acquire(), self.timeout)
                        initial_send()
                        self._send(position())
                        self._check_failure()
                    else:
                        await self._exchange(position, operation.complete, f"posición {iteration}", initial_send)
                    end = operation.last_ack_at if self.neighbors else time.perf_counter()
                    self.latencies.append((end - operation.sent_at) * 1000)
                    self.completed_iterations += 1
                    self.latency_sum_ms += self.latencies[-1]
                    self.mean_latency_ms = self.latency_sum_ms / self.completed_iterations
                    if self.verbose:
                        print(f"[CLIENTE {self.client_id}] Iteración {iteration}/{self.total_iterations}: "
                              f"RTT {self.latencies[-1]:.3f} ms", flush=True)
            self.iterations_finished_at = time.perf_counter()
            mean_latency = math.fsum(self.latencies) / len(self.latencies)
            self.mean_latency_ms = mean_latency
            metrics = message(MSG_METRICS, self.session_id, self.client_id, self.group_id,
                latencia_media_ms=mean_latency, total_iteraciones=self.completed_iterations,
                acks_unicos=self.stats.ack_unique)
            def metrics_send():
                self.metrics_sent = True
            await self._exchange(lambda: metrics, self.metrics_ack_event, "métricas", metrics_send)
            await self._wait(self.group_done_event, "fin de grupo")
            self.success = True
            return True
        except asyncio.CancelledError:
            self.fail("cliente cancelado")
            self.stats.failures += 1
            raise
        except (OSError, RuntimeError, ValueError, asyncio.TimeoutError) as exc:
            self.fail(str(exc))
            self.stats.failures += 1
            return False
        finally:
            # El generador conserva todos los endpoints hasta el certificado S.
            if not keep_open:
                await self.close()

    async def close(self):
        if self.transport is not None:
            self.transport.close()
            await self.closed_event.wait()


if __name__ == "__main__":
    # Entrada compatible: ahora lanza el generador, nunca un servidor embebido.
    from run_clients import main
    raise SystemExit(main())
=======
import argparse
import asyncio
import socket
import sys
import time
from typing import List, Optional

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


class UDPClientProtocol(asyncio.DatagramProtocol):
    """Asyncio DatagramProtocol implementation for the UDP Simulation Client."""

    def __init__(self, client: "SimulationClient"):
        self.client = client
        self.transport: Optional[asyncio.DatagramTransport] = None

    def connection_made(self, transport: asyncio.DatagramTransport) -> None:
        self.transport = transport
        self.client.transport = transport

        # Increase UDP socket buffer to handle burst traffic
        sock = transport.get_extra_info("socket")
        if sock is not None:
            try:
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 4 * 1024 * 1024)
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 4 * 1024 * 1024)
            except OSError:
                pass

    def datagram_received(self, data: bytes, addr: tuple) -> None:
        self.client.handle_datagram(data)
    
    def error_received(self, exc: Exception) -> None:
        print(f"[CLIENTE] Error en socket UDP: {exc}", file=sys.stderr)


class SimulationClient:
    def __init__(
        self,
        host: str,
        port: int,
        total_iterations: int = 10,
        timeout: float = 5.0,
    ):
        self.host = host
        self.port = port
        self.total_iterations = total_iterations
        self.timeout = timeout

        self.client_id: int = 0
        self.group_id: int = 0
        self.num_neighbors: int = 0  # V

        self.transport: Optional[asyncio.DatagramTransport] = None

        # Internal queue to decouple incoming ACKs from the datagram receiver
        self.ack_queue: asyncio.Queue = asyncio.Queue()
        self.latencies: List[float] = []

        self.registered_event: asyncio.Event = asyncio.Event()
        self.start_event: asyncio.Event = asyncio.Event()
        self.metrics_ack_event: asyncio.Event = asyncio.Event()

    def handle_datagram(self, data: bytes) -> None:
        """Event-driven datagram processing routine."""
        try:
            msg = decode_message(data)
        except Exception:
            return

        msg_type = msg.get("tipo")

        if msg_type == MSG_REGISTER_ACK:
            data_body = msg.get("datos", {})
            self.client_id = data_body["cliente_id"]
            self.group_id = data_body["grupo_id"]
            self.num_neighbors = data_body["num_vecinos"]
            print(
                f"[CLIENTE] Conectado y registrado como ID {self.client_id} "
                f"(Grupo {self.group_id}, {self.num_neighbors} vecinos)."
            )
            self.registered_event.set()

        elif msg_type == MSG_START:
            print(f"[CLIENTE {self.client_id}] ¡Señal de INICIO recibida! Comenzando simulación...")
            self.start_event.set()

        elif msg_type == MSG_POSITION:
            # Rol Reactivo: Responder ACK inmediato al vecino emisor a través del servidor
            origin = msg.get("origen")
            it = msg.get("datos", {}).get("iteracion", 0)
            ack = {
                "tipo": MSG_ACK,
                "origen": self.client_id,
                "grupo": self.group_id,
                "datos": {
                    "destino": origin,
                    "iteracion": it,
                },
            }
            send_message(self.transport, ack)

        elif msg_type == MSG_ACK:
            # Rol Activo: Despachar a la cola interna para el bucle de simulación
            self.ack_queue.put_nowait(msg)

        elif msg_type == MSG_METRICS_ACK:
            self.metrics_ack_event.set()

    async def run(self) -> None:
        """Main execution flow: registration handshake, simulation iterations, and metrics."""
        loop = asyncio.get_running_loop()
        # Connect UDP socket to server remote_addr to enable direct send() / sendto()
        transport, _ = await loop.create_datagram_endpoint(
            lambda: UDPClientProtocol(self),
            remote_addr=(self.host, self.port),
        )

        try:
            # 1. Enviar solicitud de registro
            reg_msg = {
                "tipo": MSG_REGISTER,
                "origen": 0,
                "grupo": 0,
                "datos": {},
            }
            send_message(self.transport, reg_msg)

            # 2. Recibir confirmación de registro
            await asyncio.wait_for(self.registered_event.wait(), timeout=self.timeout)

            # 3. Esperar señal global de INICIO
            print(f"[CLIENTE {self.client_id}] Esperando señal de INICIO del servidor...")
            await asyncio.wait_for(self.start_event.wait(), timeout=self.timeout)

            expected_acks = self.num_neighbors - 1

            for it in range(1, self.total_iterations + 1):
                t0 = time.perf_counter()

                # Publicar nueva posición
                pos_msg = {
                    "tipo": MSG_POSITION,
                    "origen": self.client_id,
                    "grupo": self.group_id,
                    "datos": {
                        "iteracion": it,
                        "coords": [10.0 * it, 20.0 * it, 0.0],
                    },
                }
                send_message(self.transport, pos_msg)

                # Esperar V - 1 ACKs correspondientes a ESTA iteración
                acks_received = 0
                while acks_received < expected_acks:
                    ack = await asyncio.wait_for(self.ack_queue.get(), timeout=self.timeout)
                    if ack.get("datos", {}).get("iteracion") == it:
                        acks_received += 1

                rtt_ms = (time.perf_counter() - t0) * 1000
                self.latencies.append(rtt_ms)
                print(
                    f"[CLIENTE {self.client_id}] Iteración {it}/{self.total_iterations} "
                    f"completada (RTT: {rtt_ms:.3f} ms)"
                )

            # Cálculo de latencia media
            mean_latency = sum(self.latencies) / len(self.latencies) if self.latencies else 0.0
            print(
                f"[CLIENTE {self.client_id}] Simulación finalizada. "
                f"Latencia media acumulada: {mean_latency:.3f} ms"
            )

            # Enviar reporte de métricas al servidor con protocolo Stop-and-Wait (ARQ)
            metrics_msg = {
                "tipo": MSG_METRICS,
                "origen": self.client_id,
                "grupo": self.group_id,
                "datos": {
                    "latencia_media_ms": mean_latency,
                    "total_iteraciones": self.total_iterations,
                },
            }

            max_retries = 5
            ack_timeout = 1.0
            metrics_acked = False
            self.metrics_ack_event.clear() #necesario por si se vuelve a correr el cliente

            for attempt in range(1, max_retries + 1):
                send_message(self.transport, metrics_msg)
                try:
                    await asyncio.wait_for(self.metrics_ack_event.wait(), timeout=ack_timeout)
                    metrics_acked = True
                    break
                except asyncio.TimeoutError:
                    print(
                        f"[CLIENTE {self.client_id}] Timeout esperando METRICAS_ACK (intento {attempt}/{max_retries}). Reintentando..."
                    )

            if not metrics_acked:
                print(
                    f"[CLIENTE {self.client_id}] ADVERTENCIA: No se recibió METRICAS_ACK tras {max_retries} intentos."
                )

            # Período de gracia para responder ACKs a vecinos rezagados
            await asyncio.sleep(1.5)

        except asyncio.TimeoutError:
            print(f"[CLIENTE {self.client_id}] ERROR: Timeout esperando mensajes o ACKs.")
        finally:
            transport.close()


def parse_args():
    parser = argparse.ArgumentParser(
        description="Cliente UDP de Realidad Aumentada Colaborativa (RAC)"
    )
    parser.add_argument("--host", default="127.0.0.1", help="Dirección IP del servidor (default: 127.0.0.1)")
    parser.add_argument("--port", type=int, default=8888, help="Puerto UDP del servidor (default: 8888)")
    parser.add_argument("-I", "--iterations", type=int, default=10, help="Número de iteraciones I (default: 10)")
    parser.add_argument("--timeout", type=float, default=20.0, help="Timeout en segundos (default: 20.0)")
    return parser.parse_args()


if __name__ == "__main__":
    args = parse_args()
    client = SimulationClient(
        host=args.host,
        port=args.port,
        total_iterations=args.iterations,
        timeout=args.timeout,
    )
    try:
        asyncio.run(client.run())
    except KeyboardInterrupt:
        print("\n[CLIENTE] Detenido por el usuario.")
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
