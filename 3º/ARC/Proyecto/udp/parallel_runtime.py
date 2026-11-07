"""Rutas inmutables, avisos de control acotados y observación fuera del hot path."""
from __future__ import annotations

from dataclasses import dataclass, field
import os
from pathlib import Path
import sys
import sysconfig
import threading
from types import MappingProxyType
from typing import Optional

from protocol import ProtocolError, MSG_REGISTER_ACK, message, group_for
from observability import vm_memory, memory_accounting


def runtime_probe():
    getter = getattr(sys, '_is_gil_enabled', None)
    try:
        affinity = sorted(os.sched_getaffinity(0))
    except AttributeError:
        affinity = list(range(os.cpu_count() or 1))
    quota = None
    try:
        limit, period = Path('/sys/fs/cgroup/cpu.max').read_text().split()
        if limit != 'max':
            quota = int(limit) / int(period)
    except (OSError, ValueError):
        pass
    return {'python': sys.version, 'platform': sys.platform,
        'free_threaded_build': sysconfig.get_config_var('Py_GIL_DISABLED') == 1,
        'gil_enabled': bool(getter()) if getter else True,
        'affinity_cpus': affinity, 'cpu_quota_cores': quota,
        'effective_cpus': min(len(affinity), quota) if quota is not None else len(affinity),
        'pid': os.getpid(), 'python_threads': threading.active_count()}


def require_runtime(probe=None):
    probe = probe or runtime_probe()
    if probe['platform'] != 'linux':
        raise RuntimeError('el experimento requiere Linux y SO_REUSEPORT, no macOS')
    if not probe['free_threaded_build']:
        raise RuntimeError('CPython no es free-threaded (Py_GIL_DISABLED != 1)')
    if probe['gil_enabled']:
        raise RuntimeError('el GIL está activado: no es una prueba de paralelismo sin GIL')
    return probe


@dataclass(frozen=True)
class Route:
    client_id: int
    group_id: int
    addr: tuple
    token: str
    generator: str


@dataclass(frozen=True)
class Routes:
    session: str
    instance: str
    n: int
    v: int
    iterations: int
    clients: tuple
    groups: tuple
    registrations: object

    @classmethod
    def publish(cls, server):
        if len(server.clients) != server.num_clients or server.registered_at is None:
            raise ProtocolError('publicación anterior al registro completo')
        rows = [None]
        keys = {}
        for cid in range(1, server.num_clients + 1):
            info = server.clients[cid]
            group = group_for(cid, server.neighbors_per_group)
            if info['group_id'] != group:
                raise ProtocolError('grupo de registro inconsistente')
            route = Route(cid, group, tuple(info['addr']), info['token'], info['generator'])
            if (route.generator, route.token) in keys:
                raise ProtocolError('identidad de registro duplicada')
            rows.append(route)
            keys[(route.generator, route.token)] = route
        groups = ((),) + tuple(tuple(server.groups[g]) for g in range(1, server.num_groups + 1))
        if any(len(ids) != server.neighbors_per_group for ids in groups[1:]):
            raise ProtocolError('grupo incompleto al publicar rutas')
        return cls(server.session_id, server.instance_id, server.num_clients,
                   server.neighbors_per_group, server.total_iterations, tuple(rows), groups,
                   MappingProxyType(keys))

    def sender(self, msg, addr):
        cid = msg['origen']
        if not 1 <= cid <= self.n:
            raise ProtocolError('origen no registrado')
        route = self.clients[cid]
        if route.addr != addr or route.group_id != msg['grupo']:
            raise ProtocolError('endpoint o grupo no corresponden al origen')
        return route

    def registration_reply(self, msg, addr):
        body = msg['datos']
        if msg['origen'] or msg['grupo']:
            raise ProtocolError('REGISTRO con identidad ya asignada')
        if (body['num_clientes'], body['num_vecinos'], body['total_iteraciones']) != (
                self.n, self.v, self.iterations):
            raise ProtocolError('configuración de registro distinta')
        route = self.registrations.get((body['generador'], body['token']))
        if route is None or route.addr != addr:
            raise ProtocolError('registro desconocido, cupo cerrado o cambio de endpoint')
        return message(MSG_REGISTER_ACK, self.session, group=route.group_id,
            cliente_id=route.client_id, grupo_id=route.group_id, num_clientes=self.n,
            num_vecinos=self.v, total_iteraciones=self.iterations,
            token=route.token, instancia=self.instance)


@dataclass(frozen=True)
class Notice:
    kind: str
    client_id: int
    group_id: int
    received_at: float = field(compare=False)
    instance: str = ''
    mean: Optional[float] = None
    iterations: int = 0
    unique_acks: int = 0


class ControlInbox:
    """Por receptor, solo hechos de control; nunca POSICION/ACK ni bytes UDP."""
    def __init__(self, limit):
        self.limit = limit
        self.lock = threading.Lock()
        self.values = {}
        self.armed = False
        self.maximum = 0

    def put(self, notice):
        key = (notice.kind, notice.client_id)
        with self.lock:
            old = self.values.get(key)
            if old is not None and old != notice:
                return 'conflict', False
            if old is None and len(self.values) >= self.limit:
                return 'full', False
            self.values[key] = notice  # conservar el timestamp del último duplicado válido
            self.maximum = max(self.maximum, len(self.values))
            wake = not self.armed
            self.armed = True
            return ('duplicate' if old is not None else 'new'), wake

    def drain(self):
        with self.lock:
            batch = tuple(self.values.values())
            self.values.clear()
            self.armed = False
            return batch


def proc_stat(path):
    text = Path(path).read_text()
    fields = text[text.rfind(')') + 2:].split()
    return int(fields[11]) + int(fields[12])  # utime + stime, campos 14/15


def udp_sockets():
    result = {}
    try:
        for row in Path('/proc/net/udp').read_text().splitlines()[1:]:
            fields = row.split()
            result[int(fields[9])] = {'drops': int(fields[-1]),
                'rx_queue_bytes': int(fields[4].split(':')[1], 16),
                'tx_queue_bytes': int(fields[4].split(':')[0], 16)}
    except (OSError, ValueError, IndexError):
        pass
    return result


def udp_snmp():
    try:
        lines = Path('/proc/net/snmp').read_text().splitlines()
        for index, row in enumerate(lines[:-1]):
            if row.startswith('Udp:'):
                return dict(zip(row.split()[1:], map(int, lines[index + 1].split()[1:])))
    except (OSError, ValueError):
        pass
    return {}


def key_values(path):
    try:
        return {row.split()[0]: int(row.split()[1]) for row in Path(path).read_text().splitlines()}
    except (OSError, ValueError, IndexError):
        return {}


class LinuxSampler:
    def __init__(self):
        self.ticks = os.sysconf('SC_CLK_TCK')
        self.process_ticks = proc_stat('/proc/self/stat') if sys.platform == 'linux' else None
        self.peak_rss_kib = self.peak_threads = self.processes_peak = 0
        self.thread_ticks_max = {}
        self.thread_cpu_peak = {}
        self.last_thread_ticks = {}
        self.last_at = None
        self.socket_max = {}
        self.snmp_start = udp_snmp()
        self.cgroup_cpu_start = key_values('/sys/fs/cgroup/cpu.stat')
        self.samples = []
        self.latest_vm = {}

    def sample(self, now, workers):
        if sys.platform != 'linux':
            return
        status = {}
        for line in Path('/proc/self/status').read_text().splitlines():
            key, _, value = line.partition(':')
            status[key] = value.strip()
        self.peak_rss_kib = max(self.peak_rss_kib, int(status.get('VmHWM', '0').split()[0]))
        self.peak_threads = max(self.peak_threads, int(status['Threads']))
        count = sum(p.name.isdigit() for p in Path('/proc').iterdir())
        self.processes_peak = max(self.processes_peak, count)
        thread_row = {}
        for path in Path('/proc/self/task').iterdir():
            try:
                ticks = proc_stat(path / 'stat')
            except (FileNotFoundError, ProcessLookupError):
                continue
            tid = int(path.name)
            self.thread_ticks_max[tid] = max(ticks, self.thread_ticks_max.get(tid, 0))
            if self.last_at is not None and tid in self.last_thread_ticks:
                cpu = 100 * (ticks - self.last_thread_ticks[tid]) / self.ticks / (now - self.last_at)
                self.thread_cpu_peak[tid] = max(cpu, self.thread_cpu_peak.get(tid, 0))
                thread_row[str(tid)] = cpu
            self.last_thread_ticks[tid] = ticks
        for inode, values in udp_sockets().items():
            previous = self.socket_max.setdefault(inode, {'drops': 0, 'rx_queue_bytes': 0, 'tx_queue_bytes': 0})
            for name in previous:
                previous[name] = max(previous[name], values[name])
        self.latest_vm = vm_memory()
        self.samples.append({'at': now, 'rss_kib': int(status.get('VmRSS', '0').split()[0]),
            'threads': int(status['Threads']), 'thread_cpu_percent': thread_row,
            'vm': self.latest_vm, 'memory': memory_accounting(),
            'workers': [worker.observation for worker in workers if getattr(worker, 'observation', None)]})
        self.last_at = now

    def report(self, duration):
        snmp_end = udp_snmp()
        cgroup_cpu_end = key_values('/sys/fs/cgroup/cpu.stat')
        return {'peak_rss_kib': self.peak_rss_kib, 'peak_os_threads': self.peak_threads,
            'namespace_processes_peak': self.processes_peak,
            'process_cpu_percent': (100 * (proc_stat('/proc/self/stat') - self.process_ticks)
                / self.ticks / duration) if self.process_ticks is not None else None,
            'udp_namespace_delta': {k: snmp_end[k] - v for k, v in self.snmp_start.items() if k in snmp_end},
            'cgroup_cpu_delta': {k: cgroup_cpu_end[k] - v for k,v in self.cgroup_cpu_start.items() if k in cgroup_cpu_end},
            'cgroup_memory_events': key_values('/sys/fs/cgroup/memory.events'),
            'cgroup_memory': memory_accounting(),
            'thread_cpu_peak_percent': {str(k): v for k, v in self.thread_cpu_peak.items()},
            'samples': self.samples}
