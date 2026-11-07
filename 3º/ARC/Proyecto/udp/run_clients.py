"""Un generador asyncio independiente. No importa ni arranca SimulationServer."""
from __future__ import annotations

import argparse
import asyncio
from collections import Counter
import json
from pathlib import Path
import socket
import threading
import time
import resource
import signal

from client import SimulationClient
from protocol import identifier, validate_configuration
from runtime import (Counters, RateLimiter, RetryPolicy, add_common_arguments,
    default_server_result, percentiles, policy_from, positive, write_json)
from distributed_config import RunManifest
from remote_control import GeneratorControl, validate_certificate
from parallel_runtime import udp_snmp, runtime_probe, key_values
from observability import memory_accounting


class ClientGenerator:
    def __init__(self, host, port, num_clients, neighbors, iterations, *, session_id,
                 server_result=None, result_path=None, generator_id="generador-1", local_clients=None,
                 local_host="0.0.0.0", timeout=60, deadline=120,
                 policy=None, request_rate=2000, retry_rate=100,
                 position_window=64, registration_window=64,
                 receive_buffer=64 * 1024, send_buffer=64 * 1024,
                 benchmark=False, verbose=False, client_factory=SimulationClient,
                 manifest=None, completion_mode=None, registration_rate=None, control_wrapper=None,
                 observe=False, export_samples=False):
        validate_configuration(num_clients, neighbors, iterations)
        positive(timeout, "timeout")
        positive(deadline, "deadline")
        if (type(position_window) is not int or position_window <= 0
                or type(registration_window) is not int or registration_window <= 0):
            raise ValueError("las ventanas deben ser enteros positivos")
        self.host, self.port = host, port
        self.n, self.v, self.i = num_clients, neighbors, iterations
        self.local_n = num_clients if local_clients is None else local_clients
        if type(self.local_n) is not int or not 1 <= self.local_n <= self.n:
            raise ValueError('N local debe ser entero entre 1 y N global')
        self.session_id = identifier(session_id, "sesion")
        self.generator_id = identifier(generator_id, "generador")
        self.server_result, self.result_path = Path(server_result) if server_result is not None else None, result_path
        self.completion_mode = completion_mode or ('local' if self.server_result is not None else 'remote')
        self.manifest, self.controller, self.control_wrapper = manifest, None, control_wrapper
        if manifest:
            spec = manifest.generator(self.generator_id)
            if (manifest.session, manifest.n, manifest.v, manifest.i, spec.clients) != (
                    self.session_id, self.n, self.v, self.i, self.local_n):
                raise ValueError('N global/local, sesión o V/I no coinciden con el manifiesto')
            actual = dict(request_rate=request_rate, retry_rate=retry_rate, registration_rate=registration_rate,
                          position_window=position_window, registration_window=registration_window)
            if actual != spec.load():
                raise ValueError('presupuestos del generador distintos del manifiesto')
        if self.completion_mode not in ('local', 'remote'):
            raise ValueError('completion-mode debe ser local o remote')
        if self.completion_mode == 'remote' and manifest is None:
            raise ValueError('el cierre remoto exige un manifiesto global')
        if self.completion_mode == 'local' and self.server_result is None:
            raise ValueError('cierre local exige --server-result explícito')
        if self.local_n != self.n and manifest is None:
            raise ValueError('varios generadores exigen un manifiesto que valide la suma de N locales')
        self.local_host, self.timeout, self.deadline = local_host, timeout, deadline
        self.policy = policy or RetryPolicy()
        self.request_limiter, self.retry_limiter = RateLimiter(request_rate,observe), RateLimiter(retry_rate,observe)
        self.registration_limiter = RateLimiter(registration_rate,observe) if registration_rate is not None else None
        self.observe, self.export_samples = observe, export_samples
        self.observations = []
        self.cpu_initial = key_values('/sys/fs/cgroup/cpu.stat') if observe else {}
        self.load = dict(request_rate=request_rate, registration_rate=registration_rate, retry_rate=retry_rate,
                         position_window=position_window, registration_window=registration_window)
        if retry_rate > request_rate:
            raise ValueError("retry-rate debe ser <= request-rate")
        self.position_window = asyncio.Semaphore(position_window)
        self.registration_window = asyncio.Semaphore(registration_window)
        self.receive_buffer, self.send_buffer = receive_buffer, send_buffer
        self.benchmark, self.verbose = benchmark, verbose and not benchmark
        self.client_factory = client_factory
        self.clients = []
        self.tasks = []
        self.created_at = time.perf_counter()
        self.finished_at = None
        self.error = None
        self.success = False
        self.certificate = None
        self.result = None
        self.local_done_at = self.certificate_at = self.endpoints_closed_at = None
        self.loop_lag_max = self.loop_lag_sum = 0.0
        self.loop_lag_samples = 0
        self.monitor_task = None
        self.udp_initial = udp_snmp() if manifest else {}

    async def _monitor(self):
        loop = asyncio.get_running_loop()
        while True:
            expected = loop.time() + .1
            await asyncio.sleep(.1)
            lag = max(0, loop.time()-expected)
            self.loop_lag_max = max(self.loop_lag_max, lag)
            self.loop_lag_sum += lag
            self.loop_lag_samples += 1
            if self.observe:
                usage=resource.getrusage(resource.RUSAGE_SELF)
                self.observations.append(dict(at=loop.time(),lag_seconds=lag,
                    cpu_seconds=usage.ru_utime+usage.ru_stime,memory=memory_accounting(),
                    request_grants=self.request_limiter.observation['grants'],
                    retry_grants=self.retry_limiter.observation['grants'],
                    registration_grants=self.registration_limiter.observation['grants']
                        if self.registration_limiter else None))

    async def _wait_certificate(self):
        instances = {client.instance_id for client in self.clients}
        if len(instances) != 1 or None in instances:
            raise RuntimeError("los clientes no pertenecen a una misma instancia S")
        instance = instances.pop()
        if self.completion_mode == 'remote':
            if instance != self.controller.instance:
                raise RuntimeError('los clientes y el control pertenecen a instancias distintas')
            self.certificate = await self.controller.certificate()
            return
        while True:
            try:
                report = json.loads(self.server_result.read_text(encoding="utf-8"))
            except (FileNotFoundError, json.JSONDecodeError):
                report = None
            if isinstance(report, dict) and (report.get("session"), report.get("instance")) == (
                    self.session_id, instance):
                if report.get("success") is not True:
                    raise RuntimeError(f"servidor incompleto: {report.get('error')}")
                self.certificate = validate_certificate(report, self.session_id, instance, self.n, self.v, self.i)
                return
            # Sondeo de coordinación local, fuera del dataplane y del RTT.
            await asyncio.sleep(0.01)

    async def _generate(self):
        loop = asyncio.get_running_loop()
        # Resolución única de arranque, sin executor/thread interno del generador.
        addresses = socket.getaddrinfo(self.host, self.port, family=socket.AF_INET, type=socket.SOCK_DGRAM)
        remote_host, remote_port = addresses[0][4]
        if self.completion_mode == 'remote':
            self.controller = GeneratorControl(self.manifest, self.manifest.generator(self.generator_id), self.policy)
            await loop.create_datagram_endpoint(lambda: self.controller,
                local_addr=(self.local_host, 0), remote_addr=(remote_host, remote_port))
            if self.control_wrapper:
                self.controller.transport = self.control_wrapper(self.controller.transport)
            await self.controller.hello()
            self.monitor_task = asyncio.create_task(self._monitor())
        self.clients = [self.client_factory(remote_host, remote_port, self.i, self.timeout,
            num_clients=self.n, neighbors_per_group=self.v, session_id=self.session_id,
            generator_id=self.generator_id, policy=self.policy,
            request_limiter=self.request_limiter, retry_limiter=self.retry_limiter,
            registration_limiter=self.registration_limiter,
            position_window=self.position_window, registration_window=self.registration_window,
            local_host=self.local_host, receive_buffer=self.receive_buffer,
            send_buffer=self.send_buffer, verbose=self.verbose) for _ in range(self.local_n)]
        self.tasks = [asyncio.create_task(client.run(keep_open=True)) for client in self.clients]
        for completed in asyncio.as_completed(self.tasks):
            if not await completed:
                failures = [c.error for c in self.clients if c.error]
                raise RuntimeError(failures[0] if failures else "cliente fallido")
        self.local_done_at = time.perf_counter()
        # Mantener callbacks para recuperar GRUPO_FIN_ACK hasta confirmación global S.
        await self._wait_certificate()
        self.certificate_at = time.perf_counter()
        if any(client.error is not None for client in self.clients):
            raise RuntimeError("error UDP durante la espera del certificado")
        if self.controller:
            await asyncio.gather(*(client.close() for client in self.clients))
            self.endpoints_closed_at = time.perf_counter()
            await self.controller.finish()
        self.success = True

    async def run(self):
        try:
            await asyncio.wait_for(self._generate(), self.deadline)
        except asyncio.TimeoutError:
            self.error = "deadline del generador agotado"
        except (ValueError, RuntimeError, OSError) as exc:
            self.error = str(exc)
        except asyncio.CancelledError:
            self.error = "generador cancelado"
            raise
        finally:
            if self.monitor_task:
                self.monitor_task.cancel()
                await asyncio.gather(self.monitor_task, return_exceptions=True)
            for task in self.tasks:
                if not task.done():
                    task.cancel()
            await asyncio.gather(*self.tasks, return_exceptions=True)
            await asyncio.gather(*(client.close() for client in self.clients))
            if self.controller:
                await self.controller.close()
            self.finished_at = time.perf_counter()
            self.result = self.report()
            if self.result_path:
                write_json(self.result_path, self.result)
        return self.success

    def report(self):
        udp_final = udp_snmp() if self.manifest else {}
        counters = Counters().report()
        stats = [c.stats for c in self.clients] + ([self.controller.stats] if self.controller else [])
        for stat in stats:
            for name, value in stat.report().items():
                if isinstance(value, dict):
                    aggregate = Counter(counters[name])
                    aggregate.update(value)
                    counters[name] = dict(aggregate)
                else:
                    counters[name] += value
        registered = [c.registered_at for c in self.clients if c.registered_at is not None]
        starts = [c.simulation_at for c in self.clients if c.simulation_at is not None]
        ends = [c.iterations_finished_at for c in self.clients if c.iterations_finished_at is not None]
        samples=[value for c in self.clients for value in c.latencies]
        cpu_final=key_values('/sys/fs/cgroup/cpu.stat') if self.observe else {}
        return {"role": "generator", "session": self.session_id,
                "generator": self.generator_id, "n": self.n, "v": self.v, "i": self.i,
                "n_global": self.n, "n_local": self.local_n, "completion_mode": self.completion_mode,
                "manifest_sha256": self.manifest.digest if self.manifest else None,
                "certificate_instance": self.certificate.get('instance') if self.certificate else None,
                "runtime": runtime_probe() if self.manifest else None,
                "simulation_clock": dict(start=min(starts) if starts else None,end=max(ends) if ends else None),
                "phase_clock": dict(created=self.created_at,registered=max(registered) if registered else None,
                    first_start=min(starts) if starts else None,last_start=max(starts) if starts else None,
                    iterations_finished=max(ends) if ends else None,local_done=self.local_done_at,
                    certificate=self.certificate_at,endpoints_closed=self.endpoints_closed_at,finished=self.finished_at),
                "limiter_observations": {name:limiter.observation for name,limiter in (
                    ('request',self.request_limiter),('retry',self.retry_limiter),('registration',self.registration_limiter)) if limiter},
                "observations": self.observations,
                "cgroup_memory": memory_accounting() if self.observe else {},
                "cgroup_cpu_delta": {key:cpu_final[key]-value for key,value in self.cpu_initial.items() if key in cpu_final},
                "udp_namespace_delta": {k:udp_final[k]-v for k,v in self.udp_initial.items() if k in udp_final},
                "ephemeral_port_range": Path('/proc/sys/net/ipv4/ip_local_port_range').read_text().strip()
                    if self.manifest and Path('/proc/sys/net/ipv4/ip_local_port_range').exists() else None,
                "load": self.load, "python_threads": threading.active_count(),
                "os_threads": len(list(Path('/proc/self/task').iterdir())) if Path('/proc/self/task').exists() else None,
                "local_done_seconds": self.local_done_at-self.created_at if self.local_done_at else None,
                "certificate_seconds": self.certificate_at-self.created_at if self.certificate_at else None,
                "endpoints_closed_seconds": self.endpoints_closed_at-self.created_at if self.endpoints_closed_at else None,
                "global_close_wait_seconds": self.certificate_at-self.local_done_at
                    if self.certificate_at and self.local_done_at else None,
                "event_loop_lag": dict(maximum_seconds=self.loop_lag_max,
                    mean_seconds=self.loop_lag_sum/self.loop_lag_samples if self.loop_lag_samples else None),
                "control_counters": self.controller.stats.report() if self.controller else {},
                "control_endpoint": list(self.controller.local_endpoint) if self.controller else None,
                "server_endpoint": list(self.clients[0].peer) if self.clients else None,
                "control_observed_endpoint": self.controller.observed_endpoint if self.controller else None,
                "endpoints": [dict(id=c.client_id, group=c.group_id, endpoint=list(c.local_endpoint))
                              for c in self.clients if c.local_endpoint and c.client_id],
                "success": self.success, "server_certificate_validated": self.certificate is not None,
                "error": self.error, "registered_clients": sum(c.client_id > 0 for c in self.clients),
                "started_clients": len(starts), "completed_clients": sum(c.success for c in self.clients),
                "completed_iterations": sum(c.completed_iterations for c in self.clients),
                "group_done_clients": sum(c.success for c in self.clients),
                "metrics_confirmed_clients": sum(c.metrics_ack_received for c in self.clients),
                "missing_metrics_ack_client_ids": [c.client_id for c in self.clients
                    if c.client_id and not c.metrics_ack_received],
                "registration_seconds": max(registered) - self.created_at if registered else None,
                "simulation_seconds": max(ends) - min(starts) if ends and starts else None,
                "total_seconds": (self.finished_at or time.perf_counter()) - self.created_at,
                "latency": percentiles(samples),
                "latency_samples_ms": samples if self.export_samples else None,
                "counters": counters,
                "buffers_per_socket": self.clients[0].buffers if self.clients else {},
                "failures": [{"client_id": c.client_id, "token": c.token, "error": c.error}
                             for c in self.clients if c.error]}


def parse_args(argv=None):
    parser = argparse.ArgumentParser(description="Un proceso generador de N clientes UDP asyncio")
    add_common_arguments(parser)
    parser.add_argument("--session", help="la misma sesión que el servidor; implícita con manifiesto")
    parser.add_argument("--server-result", help="certificado JSON local del servidor")
    parser.add_argument('--manifest')
    parser.add_argument('--local-clients', type=int)
    parser.add_argument('--completion-mode', choices=('remote', 'local'))
    parser.add_argument('--observe',action='store_true')
    parser.add_argument('--export-samples',action='store_true')
    parser.add_argument("--generator-id", default="generador-1")
    parser.add_argument("--local-host", default="0.0.0.0")
    parser.add_argument("--timeout", type=float, default=60, help="límite por fase/operación")
    parser.add_argument("--request-rate", type=float, default=2000,
                        help="REGISTRO/POSICION/METRICAS iniciales y retries por segundo")
    parser.add_argument("--retry-rate", type=float, default=100)
    parser.add_argument('--registration-rate', type=float)
    parser.add_argument("--position-window", type=int, default=64)
    parser.add_argument("--registration-window", type=int, default=64)
    parser.add_argument("--receive-buffer", type=int, default=64 * 1024)
    parser.add_argument("--send-buffer", type=int, default=64 * 1024)
    parser.set_defaults(session=None, clients=None, neighbors=None, iterations=None,
        request_rate=None, retry_rate=None, position_window=None, registration_window=None)
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)
    active_task = None
    stopping = False
    def terminate(*_):
        # No lanzar excepciones asíncronas dentro de asyncio.Runner/cleanup.
        nonlocal stopping
        stopping = True
        if active_task is not None and not active_task.done():
            active_task.get_loop().call_soon_threadsafe(active_task.cancel)
    previous = signal.signal(signal.SIGTERM,terminate)
    try:
        manifest = RunManifest.read(args.manifest) if args.manifest else None
        if manifest:
            spec = manifest.generator(args.generator_id)
            for attr, expected in dict(clients=manifest.n, neighbors=manifest.v, iterations=manifest.i,
                    session=manifest.session, local_clients=spec.clients, **spec.load()).items():
                value = getattr(args, attr)
                if value is not None and value != expected:
                    raise ValueError(attr + ' contradice el manifiesto')
                setattr(args, attr, expected)
        else:
            for attr, default in dict(clients=100, neighbors=25, iterations=5, request_rate=2000,
                    retry_rate=100, position_window=64, registration_window=64).items():
                if getattr(args, attr) is None:
                    setattr(args, attr, default)
            if args.session is None:
                raise ValueError('--session obligatorio sin manifiesto')
        async def execute():
            nonlocal active_task
            active_task = asyncio.current_task()
            if stopping:
                asyncio.get_running_loop().call_soon(active_task.cancel)
            generator = ClientGenerator(args.host, args.port, args.clients, args.neighbors, args.iterations,
                session_id=args.session, server_result=args.server_result, result_path=args.result,
                local_clients=args.local_clients, manifest=manifest, completion_mode=args.completion_mode,
                registration_rate=args.registration_rate,
                observe=args.observe,export_samples=args.export_samples,
                generator_id=args.generator_id, local_host=args.local_host, timeout=args.timeout,
                deadline=args.deadline, policy=policy_from(args), request_rate=args.request_rate,
                retry_rate=args.retry_rate, position_window=args.position_window,
                registration_window=args.registration_window, receive_buffer=args.receive_buffer,
                send_buffer=args.send_buffer, benchmark=args.benchmark, verbose=args.verbose)
            ok = await generator.run()
            print(json.dumps(generator.result, ensure_ascii=False), flush=True)
            return 0 if ok else 1
        return asyncio.run(execute())
    except (ValueError, OSError) as exc:
        print(json.dumps({"success": False, "error": str(exc)}), flush=True)
        return 1
    except KeyboardInterrupt:
        return 130
    except asyncio.CancelledError:
        return 130
    finally:
        signal.signal(signal.SIGTERM,previous)


if __name__ == "__main__":
    raise SystemExit(main())
