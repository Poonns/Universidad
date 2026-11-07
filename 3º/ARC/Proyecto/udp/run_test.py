<<<<<<< HEAD
"""Integración: supervisa dos procesos; nunca embebe S/G en el mismo event loop."""
from __future__ import annotations

import argparse
import asyncio
import hashlib
import json
from pathlib import Path
import platform
import sys
import time
import uuid

from protocol import validate_configuration
from runtime import RetryPolicy, add_common_arguments, positive, write_json

ROOT = Path(__file__).resolve().parent


async def run_case(args):
    validate_configuration(args.clients, args.neighbors, args.iterations)
    RetryPolicy(args.rto, args.max_rto, args.max_attempts, args.jitter)
    positive(args.deadline, "deadline")
    out = Path(args.out or ROOT / "resultados" / ("integracion-" + uuid.uuid4().hex)).resolve()
    out.mkdir(parents=True, exist_ok=False)
    session = uuid.uuid4().hex
    server_path, client_path, ready_path = out / "servidor.json", out / "clientes.json", out / "ready.json"
    common = ["-N", str(args.clients), "-V", str(args.neighbors), "-I", str(args.iterations),
        "--session", session, "--rto", str(args.rto), "--max-rto", str(args.max_rto),
        "--max-attempts", str(args.max_attempts), "--jitter", str(args.jitter),
        "--deadline", str(args.deadline), "--benchmark"]
    server_command = [sys.executable, "-u", str(ROOT / "server.py"), "--host", args.host,
        "--port", "0", "--ready-file", str(ready_path), "--result", str(server_path),
        "--control-rate", str(args.control_rate)] + common
    processes = []
    started = time.perf_counter()
    report = {"success": False, "session": session, "n": args.clients, "v": args.neighbors,
              "i": args.iterations, "commands": [server_command], "out": str(out),
              "python": sys.version, "platform": platform.platform(),
              "source_sha256": {name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest()
                  for name in ("protocol.py", "runtime.py", "server.py", "client.py",
                               "run_clients.py", "run_test.py")}}
    try:
        with (out / "servidor.log").open("wb") as server_log, (out / "clientes.log").open("wb") as client_log:
            server = await asyncio.create_subprocess_exec(*server_command, cwd=str(ROOT),
                stdout=server_log, stderr=asyncio.subprocess.STDOUT)
            processes.append(server)
            ready_deadline = asyncio.get_running_loop().time() + min(10, args.deadline)
            while not ready_path.exists():
                if server.returncode is not None:
                    raise RuntimeError("servidor terminó antes de READY; ver servidor.log")
                if asyncio.get_running_loop().time() >= ready_deadline:
                    raise RuntimeError("servidor no publicó READY")
                await asyncio.sleep(0.01)
            ready = json.loads(ready_path.read_text())
            command = [sys.executable, "-u", str(ROOT / "run_clients.py"), "--host", args.host,
                "--port", str(ready["port"]), "--server-result", str(server_path),
                "--result", str(client_path), "--request-rate", str(args.request_rate),
                "--retry-rate", str(args.retry_rate), "--position-window", str(args.position_window)] + common
            report["commands"].append(command)
            clients = await asyncio.create_subprocess_exec(*command, cwd=str(ROOT),
                stdout=client_log, stderr=asyncio.subprocess.STDOUT)
            processes.append(clients)
            report["pids"] = {"server": server.pid, "generator": clients.pid}
            remaining = max(0.001, args.deadline - (time.perf_counter() - started))
            codes = await asyncio.wait_for(asyncio.gather(server.wait(), clients.wait()), remaining)
            report["returncodes"] = codes
            if not server_path.exists() or not client_path.exists():
                raise RuntimeError("falta resultado de algún proceso")
            sr, gr = json.loads(server_path.read_text()), json.loads(client_path.read_text())
            report["server"], report["generator"] = sr, gr
            expected = args.clients * args.iterations
            checks = (codes == [0, 0], sr.get("success") is True, gr.get("success") is True,
                gr.get("server_certificate_validated") is True,
                sr.get("session") == gr.get("session") == session,
                sr.get("group_done_acks") == args.clients,
                gr.get("registered_clients") == args.clients,
                gr.get("completed_clients") == args.clients,
                gr.get("completed_iterations") == expected,
                gr.get("counters", {}).get("ack_unique") == expected * (args.neighbors - 1))
            report["success"] = all(checks)
            if not report["success"]:
                report["error"] = "certificación de integración incompleta"
    except (RuntimeError, OSError, ValueError, asyncio.TimeoutError) as exc:
        report["error"] = str(exc) or type(exc).__name__
    finally:
        for process in processes:
            if process.returncode is None:
                process.terminate()
        for process in processes:
            try:
                await asyncio.wait_for(process.wait(), 3)
            except asyncio.TimeoutError:
                process.kill()
                await process.wait()
        report["total_seconds"] = time.perf_counter() - started
        write_json(out / "integracion.json", report)
    print(json.dumps({key: value for key, value in report.items()
                      if key not in ("server", "generator", "commands")}), flush=True)
    return report


def parse_args(argv=None):
    parser = argparse.ArgumentParser(description="Integración UDP con servidor y generador separados")
    add_common_arguments(parser)
    parser.add_argument("--out", help="directorio nuevo para logs y resultados")
    parser.add_argument("--request-rate", type=float, default=2000)
    parser.add_argument("--retry-rate", type=float, default=100)
    parser.add_argument("--control-rate", type=float, default=2000)
    parser.add_argument("--position-window", type=int, default=64)
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)
    try:
        report = asyncio.run(run_case(args))
        return 0 if report["success"] else 1
    except (ValueError, OSError) as exc:
        print(json.dumps({"success": False, "error": str(exc)}), flush=True)
        return 1
    except KeyboardInterrupt:
        return 130


if __name__ == "__main__":
    raise SystemExit(main())
=======
import asyncio
import sys
from client import SimulationClient
from server import SimulationServer


async def main():
    host = "127.0.0.1"
    port = 9999
    num_clients = 100  # N
    neighbors = 25  # V
    iterations = 5  # I

    # 1. Iniciar servidor UDP en background
    server = SimulationServer(host, port, num_clients, neighbors)
    server_task = asyncio.create_task(server.start())

    # Esperar brevemente a que el socket del servidor esté a la escucha
    await asyncio.sleep(0.2)

    # 2. Iniciar clientes UDP concurrentemente
    clients = [
        SimulationClient(host=host, port=port, total_iterations=iterations)
        for _ in range(num_clients)
    ]
    client_tasks = [asyncio.create_task(c.run()) for c in clients]

    # 3. Esperar que todos los clientes y el servidor finalicen
    await asyncio.gather(*client_tasks)
    await server_task


if __name__ == "__main__":
    asyncio.run(main())
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
