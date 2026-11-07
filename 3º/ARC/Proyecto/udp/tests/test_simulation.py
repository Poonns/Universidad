import asyncio
import json
from pathlib import Path
import tempfile
import unittest
from types import SimpleNamespace
from support import *
from runtime import write_json


class SimulationTests(unittest.IsolatedAsyncioTestCase):
    def assert_complete(self, server, generator, n, v, iterations):
        self.assertTrue(server.success, server.result)
        self.assertTrue(generator.success, generator.result)
        self.assertEqual(len(server.clients), n)
        self.assertEqual(len(server.started_clients), n)
        self.assertEqual(len(server.metrics), n)
        self.assertEqual(len(server.done_acks), n)
        self.assertEqual(generator.result["completed_clients"], n)
        self.assertEqual(generator.result["completed_iterations"], n * iterations)
        self.assertEqual(generator.result["counters"]["ack_unique"], n * iterations * (v - 1))
        self.assertTrue(generator.result["server_certificate_validated"])
        self.assertEqual(server.stats.protocol_errors, 0)
        self.assertEqual(generator.result["counters"]["protocol_errors"], 0)
        self.assertEqual(server.stats.socket_errors, 0)
        self.assertEqual(generator.result["counters"]["socket_errors"], 0)
        self.assertTrue(all(c.closed_event.is_set() for c in generator.clients))
        self.assertTrue(server.closed_event.is_set())

    async def test_full_small_simulation_and_datagram_balance(self):
        server, generator = await simulate()
        self.assert_complete(server, generator, 4, 2, 3)
        expected = 4 * 3 * (1 + 3 * (2 - 1)) + 8 * 4
        self.assertEqual(server.stats.datagrams_sent + generator.result["counters"]["datagrams_sent"], expected)
        self.assertEqual(server.stats.datagrams_received, generator.result["counters"]["datagrams_sent"])
        self.assertEqual(server.stats.datagrams_sent, generator.result["counters"]["datagrams_received"])

    async def test_v1_no_neighbors(self):
        server, generator = await simulate(n=1, v=1)
        self.assert_complete(server, generator, 1, 1, 3)
        self.assertEqual(generator.result["counters"]["ack_unique"], 0)

    async def test_v25(self):
        # Correctness con fanout 24 y debug asyncio: no es una referencia temporal.
        # El timeout interior anterior de 2 s no crecía con server_deadline=6.
        server, generator = await simulate(n=25, v=25, iterations=2, server_deadline=15,
            phase_timeout=10, policy=RetryPolicy(.25, 2, 8, .2))
        self.assert_complete(server, generator, 25, 25, 2)

    async def test_lost_register_ack_and_start_before_registration(self):
        server, generator = await simulate(server_filter=drop_first(MSG_REGISTER_ACK))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(server.stats.registration_duplicates, 1)
        self.assertGreaterEqual(generator.result["counters"]["retransmissions"], 1)

    async def test_lost_start_is_retransmitted(self):
        server, generator = await simulate(server_filter=drop_first(MSG_START))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(server.stats.retransmissions, 1)

    async def test_lost_start_ack_does_not_stop_data_or_metrics(self):
        factory = faulty_clients(drop_first(MSG_START_ACK))
        server, generator = await simulate(client_factory=factory)
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(server.stats.retransmissions, 1)

    async def test_lost_position_is_retransmitted(self):
        server, generator = await simulate(client_factory=faulty_clients(drop_first(MSG_POSITION)))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(generator.result["counters"]["retransmissions"], 1)

    async def test_lost_ack_uses_selective_retry(self):
        lost, selective = [], []
        def server_filter(msg, addr):
            if msg["tipo"] == MSG_POSITION and "pendientes" in msg["datos"]:
                selective.append(msg["datos"]["pendientes"])
            if msg["tipo"] == MSG_ACK and msg["origen"] == 2 and msg["datos"]["destino"] == 1 and not lost:
                lost.append(True)
                return "drop"
        server, generator = await simulate(n=3, v=3, iterations=2, server_filter=server_filter)
        self.assert_complete(server, generator, 3, 3, 2)
        self.assertIn([2], selective)
        self.assertGreaterEqual(generator.result["counters"]["retransmissions"], 1)

    async def test_duplicate_acks_do_not_increase_progress(self):
        def duplicate(msg, addr):
            return "duplicate" if msg["tipo"] == MSG_ACK else None
        server, generator = await simulate(server_filter=duplicate)
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreater(generator.result["counters"]["ack_duplicates"], 0)
        self.assertEqual(len([v for c in generator.clients for v in c.latencies]), 12)

    async def test_lost_metrics_is_retried(self):
        server, generator = await simulate(client_factory=faulty_clients(drop_first(MSG_METRICS)))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(generator.result["counters"]["retransmissions"], 1)

    async def test_lost_metrics_ack_recovers_before_group_ack(self):
        server, generator = await simulate(server_filter=drop_first(MSG_METRICS_ACK))
        self.assert_complete(server, generator, 4, 2, 3)

    async def test_lost_group_done_is_retried(self):
        server, generator = await simulate(server_filter=drop_first(MSG_GROUP_DONE))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(server.stats.retransmissions, 1)

    async def test_lost_final_group_ack_recovered_without_fixed_grace(self):
        server, generator = await simulate(client_factory=faulty_clients(drop_first(MSG_GROUP_DONE_ACK)))
        self.assert_complete(server, generator, 4, 2, 3)
        self.assertGreaterEqual(server.stats.retransmissions, 1)

    async def test_fast_client_keeps_serving_slow_neighbor_without_iteration_barrier(self):
        observations = []
        class SlowClient(SimulationClient):
            async def before_position(self, iteration):
                if self.client_id == 2 and iteration == 1:
                    await asyncio.sleep(.08)
                if self.client_id == 2 and iteration == 2:
                    # El cliente 1 ya completó I, sigue activo y debe ACKear a 2.
                    observations.append((self.stats.positions_unique,
                                         self.mean_latency_ms == self.latencies[0]))
        server, generator = await simulate(n=2, v=2, iterations=3, client_factory=SlowClient)
        self.assert_complete(server, generator, 2, 2, 3)
        fast, slow = sorted(generator.clients, key=lambda c: c.client_id)
        self.assertLess(fast.iterations_finished_at, slow.iterations_finished_at)
        self.assertEqual(observations, [(3, True)])  # Media actualizada antes de la iteración 2.
        self.assertEqual(fast.stats.positions_unique, 3)

    async def test_persistent_loss_fails_with_bounded_attempts(self):
        def lose_acks(msg, addr):
            if msg["tipo"] == MSG_ACK and msg["datos"]["destino"] == 1:
                return "drop"
        server, generator = await simulate(n=2, v=2, iterations=1,
                                           server_filter=lose_acks, server_deadline=1)
        self.assertFalse(generator.success)
        self.assertFalse(server.success)
        self.assertLessEqual(generator.result["counters"]["retransmissions"], 5)
        self.assertGreater(generator.result["counters"]["failures"], 0)
        self.assertEqual(len(server.done_acks), 0)

    async def test_stale_certificate_of_other_instance_is_not_accepted(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "server.json"
            generator = ClientGenerator("127.0.0.1", 9999, 1, 1, 1,
                                        session_id=SESSION, server_result=path)
            generator.clients = [SimpleNamespace(instance_id=INSTANCE)]
            proof = {"role": "server", "session": SESSION, "instance": "old-instance",
                "n": 1, "v": 1, "i": 1, "success": True, "simulation_complete": True,
                "registered_clients": 1, "started_clients": 1, "completed_clients": 1,
                "completed_iterations": 1, "group_done_acks": 1, "groups_finished": 1}
            write_json(path, proof)
            with self.assertRaises(asyncio.TimeoutError):
                await asyncio.wait_for(generator._wait_certificate(), .03)
            self.assertIsNone(generator.certificate)
            proof["instance"] = INSTANCE
            write_json(path, proof)
            await asyncio.wait_for(generator._wait_certificate(), .1)
            self.assertEqual(generator.certificate["instance"], INSTANCE)


class ProcessTests(unittest.IsolatedAsyncioTestCase):
    async def test_server_and_generator_run_in_separate_processes(self):
        import os
        import sys
        with tempfile.TemporaryDirectory() as tmp:
            out = Path(tmp) / "case"
            process = await asyncio.create_subprocess_exec(sys.executable, str(ROOT / "run_test.py"),
                "-N", "4", "-V", "2", "-I", "3", "--rto", ".05", "--max-rto", ".2",
                "--deadline", "10", "--out", str(out), cwd=str(ROOT),
                stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)
            try:
                stdout, stderr = await asyncio.wait_for(process.communicate(), 15)
            finally:
                if process.returncode is None:
                    process.kill()
                    await process.wait()
            self.assertEqual(process.returncode, 0, (stdout, stderr,
                [(p.name, p.read_text()) for p in out.glob("*.log")]))
            report = json.loads((out / "integracion.json").read_text())
            self.assertTrue(report["success"], report)
            self.assertNotEqual(report["pids"]["server"], report["pids"]["generator"])
            self.assertNotIn(os.getpid(), report["pids"].values())
            self.assertEqual(report["returncodes"], [0, 0])
            self.assertEqual(report["generator"]["completed_iterations"], 12)
