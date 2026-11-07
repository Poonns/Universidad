import asyncio
import time
import unittest
from support import *
from runtime import RateLimiter, RetryExhausted


class ClientStateTests(unittest.IsolatedAsyncioTestCase):
    async def test_late_ack_cancels_retry_waiting_for_rate_budget(self):
        client = unit_client()
        client.policy = RetryPolicy(.01, .02, 3, 0)
        entered, blocked = asyncio.Event(), asyncio.Event()
        class BlockedLimiter:
            async def acquire(self):
                entered.set()
                await blocked.wait()
        client.retry_limiter = BlockedLimiter()
        operation = Operation(1, [0, 0, 0], sent_at=time.perf_counter())
        client.operations[1] = operation
        task = asyncio.create_task(client._exchange(lambda: message(MSG_POSITION, SESSION,
            1, 1, iteracion=1, coords=[0, 0, 0]), operation.complete, "posición"))
        try:
            await asyncio.wait_for(entered.wait(), .5)
            self.assertEqual(len(client.transport.messages), 1)
            client.handle_datagram(encode_message(ack()))
            client.handle_datagram(encode_message(ack(origin=3)))
            await asyncio.wait_for(task, .5)
            self.assertFalse(blocked.is_set())
            self.assertEqual(len(client.transport.messages), 1)
            self.assertEqual(client.stats.retransmissions, 0)
            self.assertIsNotNone(operation.last_ack_at)
        finally:
            if not task.done():
                task.cancel()
            await asyncio.gather(task, return_exceptions=True)

    async def test_ack_duplicates_do_not_complete(self):
        client = unit_client()
        operation = Operation(1, [0, 0, 0], sent_at=time.perf_counter())
        client.operations[1] = operation
        for _ in range(4):
            client.handle_datagram(encode_message(ack()))
        self.assertEqual(operation.ack_sources, {2})
        self.assertFalse(operation.complete.is_set())
        client.handle_datagram(encode_message(ack(origin=3)))
        self.assertTrue(operation.complete.is_set())
        self.assertEqual(client.stats.ack_unique, 2)
        self.assertEqual(client.stats.ack_duplicates, 3)

    async def test_invalid_ack_identity_destination_session_and_unsent_iteration(self):
        client = unit_client(n=6)
        client.operations[1] = Operation(1, [0, 0, 0], sent_at=1)
        packets = [ack(origin=1), ack(origin=4), ack(target=2), ack(group=2),
                   ack(session="otra-sesion"), ack(iteration=2)]
        for packet in packets:
            client.handle_datagram(encode_message(packet))
        client.handle_datagram(encode_message(ack()), ("192.0.2.1", 9999))
        self.assertEqual(client.stats.ack_unique, 0)
        self.assertEqual(client.stats.protocol_errors, len(packets) + 1)

    async def test_ack_state_preserved_for_different_iterations(self):
        client = unit_client()
        for iteration in (1, 2):
            client.operations[iteration] = Operation(iteration, [0, 0, 0], sent_at=1)
        client.handle_datagram(encode_message(ack(iteration=2)))
        client.handle_datagram(encode_message(ack(iteration=1)))
        self.assertEqual(client.operations[1].ack_sources, {2})
        self.assertEqual(client.operations[2].ack_sources, {2})
        client.handle_datagram(encode_message(ack(origin=3, iteration=1)))
        client.completed_iterations = 1
        client.handle_datagram(encode_message(ack(iteration=1)))
        self.assertEqual(client.operations[2].ack_sources, {2})
        self.assertEqual(client.stats.late_acks, 1)
        self.assertEqual(client.stats.ack_duplicates, 1)

    async def test_duplicate_positions_reack_without_changing_metrics(self):
        client = unit_client()
        client.completed_iterations = 3
        client.latencies = [1, 2, 3]
        packet = message(MSG_POSITION, SESSION, 2, 1, iteracion=1, coords=[10, 20, 0])
        for _ in range(2):
            client.handle_datagram(encode_message(packet))
        self.assertEqual(len(client.transport.messages), 2)
        self.assertTrue(all(m[0]["tipo"] == MSG_ACK for m in client.transport.messages))
        self.assertEqual(client.stats.position_duplicates, 1)
        self.assertEqual(client.stats.positions_unique, 1)
        self.assertEqual(client.latencies, [1, 2, 3])
        packet["datos"]["coords"][0] = 99
        client.handle_datagram(encode_message(packet))
        self.assertEqual(len(client.transport.messages), 2)
        self.assertEqual(client.stats.protocol_errors, 1)

    async def test_start_before_register_ack_is_deferred(self):
        client = SimulationClient("127.0.0.1", 9999, 2, num_clients=2,
            neighbors_per_group=2, session_id=SESSION, token="token-1")
        client.transport = MemoryTransport()
        client.registration_sent = True
        start = message(MSG_START, SESSION, group=1, destino=1,
                        token=client.token, instancia=INSTANCE)
        client.handle_datagram(encode_message(start))
        self.assertEqual(client.client_id, 0)
        self.assertFalse(client.start_event.is_set())
        self.assertEqual(len(client.transport.messages), 0)
        register_ack = message(MSG_REGISTER_ACK, SESSION, group=1, cliente_id=1,
            grupo_id=1, num_vecinos=2, num_clientes=2, total_iteraciones=2,
            token=client.token, instancia=INSTANCE)
        client.handle_datagram(encode_message(register_ack))
        self.assertEqual(client.client_id, 1)
        self.assertTrue(client.start_event.is_set())
        self.assertEqual(client.transport.messages[0][0]["tipo"], MSG_START_ACK)

    async def test_partial_ack_does_not_reset_rto_or_bypass_retry_limit(self):
        client = unit_client()
        client.policy = RetryPolicy(.01, .02, 3, 0)
        client.request_limiter, client.retry_limiter = RateLimiter(10000), RateLimiter(10000)
        client.operations[1] = Operation(1, [0, 0, 0], sent_at=1)
        client.transport = MemoryTransport(lambda msg, addr:
            client.handle_datagram(encode_message(ack())))
        start = time.perf_counter()
        with self.assertRaises(RetryExhausted):
            await client._exchange(lambda: message(MSG_POSITION, SESSION, 1, 1,
                iteracion=1, coords=[0, 0, 0]), client.operations[1].complete, "test")
        self.assertLess(time.perf_counter() - start, 1)
        self.assertEqual(len(client.transport.messages), 3)
        self.assertEqual(client.stats.retransmissions, 2)
        self.assertEqual(client.stats.timeouts, 3)
        self.assertFalse(client.operations[1].complete.is_set())


class ServerStateTests(unittest.IsolatedAsyncioTestCase):
    async def test_metrics_are_idempotent_and_conflicting_duplicates_rejected(self):
        server = self.make_server(n=2, v=2)
        for cid in (1, 2):
            self.register(server, f"token-{cid}", ("127.0.0.1", 10000 + cid), n=2)
            server.clients[cid]["start_sent"] = True
        packet = message(MSG_METRICS, SESSION, 1, 1, latencia_media_ms=2.0,
                         total_iteraciones=3, acks_unicos=3)
        addr = ("127.0.0.1", 10001)
        for _ in range(2):
            server.handle_datagram(encode_message(packet), addr)
        self.assertEqual(server.metrics, {1: 2.0})
        self.assertEqual(server.stats.metrics_duplicates, 1)
        self.assertFalse(server.finished_groups)
        packet["datos"]["latencia_media_ms"] = 3.0
        server.handle_datagram(encode_message(packet), addr)
        self.assertEqual(server.metrics, {1: 2.0})
        self.assertEqual(server.stats.protocol_errors, 1)

    def make_server(self, n=4, v=2):
        server = SimulationServer("127.0.0.1", 0, n, v, 3, session_id=SESSION)
        server.transport = MemoryTransport()
        return server

    def register(self, server, token, addr, n=4, v=2):
        packet = message(MSG_REGISTER, SESSION, token=token, generador="g1",
                         num_clientes=n, num_vecinos=v, total_iteraciones=3)
        server.handle_datagram(encode_message(packet), addr)
        return packet

    async def test_registration_of_one_client(self):
        server = self.make_server()
        self.register(server, "token-1", ("127.0.0.1", 10001))
        self.assertEqual(len(server.clients), 1)
        response = server.transport.messages[-1][0]
        self.assertEqual(response["tipo"], MSG_REGISTER_ACK)
        self.assertEqual(response["datos"]["cliente_id"], 1)
        self.assertEqual(response["datos"]["grupo_id"], 1)

    async def test_duplicate_registration_preserves_identity(self):
        server = self.make_server()
        addr = ("127.0.0.1", 10001)
        packet = self.register(server, "token-1", addr)
        server.handle_datagram(encode_message(packet), addr)
        self.assertEqual(len(server.clients), 1)
        self.assertEqual(server.groups[1], [1])
        responses = server.transport.messages
        self.assertEqual(responses[0][0], responses[1][0])
        self.assertEqual(server.stats.registration_duplicates, 1)
        server.handle_datagram(encode_message(packet), ("127.0.0.1", 10002))
        self.assertEqual(len(server.clients), 1)
        self.assertEqual(server.stats.protocol_errors, 1)

    async def test_unregistered_and_spoofed_ack_rejected(self):
        server = self.make_server()
        for cid in range(1, 5):
            self.register(server, f"token-{cid}", ("127.0.0.1", 10000 + cid))
            server.clients[cid]["start_sent"] = True
        baseline = len(server.transport.messages)
        for packet, addr in [(ack(), ("127.0.0.1", 12345)),
            (ack(target=3), ("127.0.0.1", 10002)),
            (ack(target=2), ("127.0.0.1", 10002)),
            (ack(iteration=4), ("127.0.0.1", 10002)),
            (ack(session="otro-run"), ("127.0.0.1", 10002))]:
            server.handle_datagram(encode_message(packet), addr)
        self.assertEqual(len(server.transport.messages), baseline)
        self.assertEqual(server.stats.protocol_errors, 5)
        server.handle_datagram(encode_message(ack()), ("127.0.0.1", 10002))
        self.assertEqual(len(server.transport.messages), baseline + 1)
        self.assertEqual(server.transport.messages[-1][1], ("127.0.0.1", 10001))

    async def test_registration_configuration_mismatch(self):
        server = self.make_server()
        self.register(server, "token-1", ("127.0.0.1", 10001), n=6)
        self.assertFalse(server.clients)
        self.assertEqual(server.transport.messages[-1][0]["tipo"], MSG_ERROR)
