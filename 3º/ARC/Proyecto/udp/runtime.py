"""Política ARQ, pacing, instrumentación y resultados; solo biblioteca estándar."""
from __future__ import annotations

import asyncio
from dataclasses import asdict, dataclass, field
import json
import math
import os
from pathlib import Path
import random
import socket
import tempfile
from typing import Optional


@dataclass(frozen=True)
class RetryPolicy:
    timeout: float = 0.25
    max_timeout: float = 2.0
    max_attempts: int = 8  # incluye el envío inicial
    jitter: float = 0.2

    def __post_init__(self):
        if not (math.isfinite(self.timeout) and self.timeout > 0
                and math.isfinite(self.max_timeout) and self.max_timeout >= self.timeout):
            raise ValueError("RTO debe ser positivo y max-rto >= rto")
        if type(self.max_attempts) is not int or not 1 <= self.max_attempts <= 64:
            raise ValueError("max-attempts debe estar entre 1 y 64")
        if not math.isfinite(self.jitter) or not 0 <= self.jitter < 1:
            raise ValueError("jitter debe estar en [0,1)")

    def delay(self, attempt: int, rng: random.Random) -> float:
        return min(self.max_timeout, self.timeout * 2 ** (attempt - 1)) * rng.uniform(
            1 - self.jitter, 1)


class RetryExhausted(RuntimeError):
    pass


class RateLimiter:
    """Pacing compartido sin ráfaga acumulada; un consumidor despierto cada vez."""
    def __init__(self, rate: float, observe: bool = False):
        if not math.isfinite(rate) or rate <= 0:
            raise ValueError("la tasa debe ser positiva y finita")
        self.interval = 1 / rate
        self.next_at = 0.0
        self.lock = asyncio.Lock()
        self.observation = (dict(grants=0, queued_seconds=0., requested_sleep_seconds=0.,
            actual_sleep_seconds=0., oversleep_max_seconds=0., first_grant=None, last_grant=None)
            if observe else None)

    async def acquire(self):
        observation = self.observation
        queued = asyncio.get_running_loop().time() if observation is not None else None
        async with self.lock:
            loop = asyncio.get_running_loop()
            entered = loop.time() if observation is not None else None
            delay = self.next_at - loop.time()
            if delay > 0:
                await asyncio.sleep(delay)
            self.next_at = loop.time() + self.interval
            if observation is not None:
                granted = self.next_at-self.interval
                observation['grants'] += 1
                observation['queued_seconds'] += entered-queued
                observation['requested_sleep_seconds'] += max(0, delay)
                observation['actual_sleep_seconds'] += granted-entered
                observation['oversleep_max_seconds'] = max(observation['oversleep_max_seconds'], granted-entered-max(0,delay))
                if observation['first_grant'] is None:
                    observation['first_grant'] = granted
                observation['last_grant'] = granted


@dataclass
class Counters:
    datagrams_sent: int = 0
    datagrams_received: int = 0
    ack_duplicates: int = 0
    ack_unique: int = 0
    position_duplicates: int = 0
    positions_unique: int = 0
    registration_duplicates: int = 0
    metrics_duplicates: int = 0
    retransmissions: int = 0
    timeouts: int = 0
    protocol_errors: int = 0
    socket_errors: int = 0
    failures: int = 0
    late_acks: int = 0
    tx_by_type: dict = field(default_factory=dict)
    rx_by_type: dict = field(default_factory=dict)

    def sent(self, kind: str):
        self.datagrams_sent += 1
        self.tx_by_type[kind] = self.tx_by_type.get(kind, 0) + 1

    def received(self, kind: str):
        self.rx_by_type[kind] = self.rx_by_type.get(kind, 0) + 1

    def report(self) -> dict:
        return asdict(self)


def configure_socket(transport, receive: int, send: int) -> dict:
    sock = transport.get_extra_info("socket")
    if sock is None:
        return {}
    for value in (receive, send):
        if type(value) is not int or value < 4096:
            raise ValueError("buffers UDP: mínimo 4096 bytes")
    # Un error se propaga: no anunciar un buffer que no se ha podido configurar.
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, receive)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, send)
    return {"receive_requested": receive, "send_requested": send,
            "receive_effective": sock.getsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF),
            "send_effective": sock.getsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF)}


def percentiles(samples: list[float]) -> dict:
    ordered = sorted(samples)
    def rank(p: float, minimum: int) -> Optional[float]:
        return ordered[math.ceil(p * len(ordered)) - 1] if len(ordered) >= minimum else None
    return {"samples": len(ordered),
            "mean_ms": math.fsum(ordered) / len(ordered) if ordered else None,
            "p95_ms": rank(0.95, 20), "p99_ms": rank(0.99, 100)}


def write_json(path, data: dict) -> None:
    """Publicación atómica: el lector nunca ve un JSON a medio escribir."""
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temporary = tempfile.mkstemp(prefix=path.name + ".", dir=str(path.parent))
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as stream:
            json.dump(data, stream, ensure_ascii=False, allow_nan=False, indent=2)
            stream.write("\n")
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def default_server_result(session: str) -> Path:
    return Path("resultados") / session / "servidor.json"


def positive(value: float, name: str) -> None:
    if not math.isfinite(value) or value <= 0:
        raise ValueError(f"{name} debe ser positivo y finito")


def add_common_arguments(parser):
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8888)
    parser.add_argument("-N", "--clients", type=int, default=100)
    parser.add_argument("-V", "--neighbors", type=int, default=25)
    parser.add_argument("-I", "--iterations", type=int, default=5)
    parser.add_argument("--rto", type=float, default=0.25)
    parser.add_argument("--max-rto", type=float, default=2.0)
    parser.add_argument("--max-attempts", type=int, default=8,
                        help="máximo de envíos por operación, incluido el inicial")
    parser.add_argument("--jitter", type=float, default=0.2)
    parser.add_argument("--deadline", type=float, default=120.0,
                        help="límite administrativo total en segundos")
    parser.add_argument("--benchmark", action="store_true")
    parser.add_argument("--verbose", action="store_true")
    parser.add_argument("--result", help="resultado JSON del proceso")


def policy_from(args) -> RetryPolicy:
    return RetryPolicy(args.rto, args.max_rto, args.max_attempts, args.jitter)
