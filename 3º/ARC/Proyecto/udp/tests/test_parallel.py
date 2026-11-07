"""Correctness de threads/registro/transición. No son medidas de rendimiento."""
import asyncio
from dataclasses import FrozenInstanceError
from pathlib import Path
import tempfile
import threading
import time
import unittest

from support import (SESSION, INSTANCE, SimulationServer, ClientGenerator, FaultTransport,
    RetryPolicy, faulty_clients, drop_first, MemoryTransport, ack)
from protocol import *
from runtime import Counters
from parallel_runtime import ControlInbox, Notice, Routes, require_runtime
from server_parallel import ParallelServer, RegistrationServer


def snapshot():
    server = SimulationServer('127.0.0.1', 0, 2, 2, 3, session_id=SESSION)
    server.instance_id = INSTANCE
    server.clients = {cid: {'addr': ('127.0.0.1', 10000 + cid), 'group_id': 1,
        'token': f'token-{cid}', 'generator': 'g1', 'start_sent': False} for cid in (1, 2)}
    server.groups = {1: [1, 2]}
    server.registered_at = 1
    return server, Routes.publish(server)


async def parallel_case(workers=2, n=4, v=2, iterations=3, wrapper=None,
                        factory=None, deadline=6):
    with tempfile.TemporaryDirectory(prefix='arc-parallel-') as tmp:
        result = Path(tmp) / 'server.json'
        server = ParallelServer(workers=workers, host='127.0.0.1', port=0,
            num_clients=n, neighbors_per_group=v, total_iterations=iterations,
            session_id=SESSION, policy=RetryPolicy(.08, .2, 8, 0),
            deadline=deadline, control_rate=10000, result_path=result,
            enforce_runtime=False, transport_wrapper=wrapper)
        supervisor = threading.Thread(target=server.run, name='test-supervisor')
        supervisor.start()
        generator = None
        try:
            until = time.monotonic() + 2
            while not server.bound.is_set():
                if server.stop_flag.is_set() or time.monotonic() > until:
                    raise AssertionError(f'no READY: {server.error}')
                await asyncio.sleep(.005)
            args = {} if factory is None else {'client_factory': factory}
            generator = ClientGenerator(*server.address, n, v, iterations,
                session_id=SESSION, server_result=result, deadline=deadline+1,
                timeout=deadline, policy=RetryPolicy(.08, .2, 8, 0),
                request_rate=5000, retry_rate=1000, **args)
            await generator.run()
            until = time.monotonic() + 2
            while supervisor.is_alive() and time.monotonic() < until:
                await asyncio.sleep(.005)
            assert not supervisor.is_alive(), 'supervisor no terminó'
            supervisor.join()
            return server, generator
        finally:
            server.request_stop()
            until = time.monotonic() + 2
            while supervisor.is_alive() and time.monotonic() < until:
                await asyncio.sleep(.005)
            assert not supervisor.is_alive(), 'hilos sin cerrar'
            supervisor.join()


class ParallelStateTests(unittest.IsolatedAsyncioTestCase):
    async def test_publish_requires_complete_registry_and_routes_are_immutable(self):
        original, routes = snapshot()
        original.clients[1]['addr'] = ('192.0.2.1', 9)
        self.assertEqual(routes.clients[1].addr, ('127.0.0.1', 10001))
        with self.assertRaises(TypeError):
            routes.registrations[('g1', 'token-3')] = routes.clients[1]
        with self.assertRaises(FrozenInstanceError):
            routes.clients[1].client_id = 2
        original.clients.pop(2)
        with self.assertRaises(ProtocolError):
            Routes.publish(original)

    async def test_duplicate_registration_any_receiver_uses_same_identity(self):
        _, routes = snapshot()
        packet = message(MSG_REGISTER, SESSION, token='token-1', generador='g1',
                         num_clientes=2, num_vecinos=2, total_iteraciones=3)
        for _ in range(4):
            response = routes.registration_reply(packet, ('127.0.0.1', 10001))
            self.assertEqual(response['datos']['cliente_id'], 1)
            self.assertEqual(response['datos']['grupo_id'], 1)
        with self.assertRaises(ProtocolError):
            routes.registration_reply(packet, ('127.0.0.1', 10002))
        packet['datos']['token'] = 'new-token'
        with self.assertRaises(ProtocolError):
            routes.registration_reply(packet, ('127.0.0.1', 10003))

    async def test_local_start_race_preserves_valid_packet_and_rejects_premature_packet(self):
        _, routes = snapshot()
        server = ParallelServer(workers=2, host='127.0.0.1', num_clients=2,
                                neighbors_per_group=2, total_iterations=3, enforce_runtime=False)
        worker = server.workers[1]
        worker.routes, worker.stats, worker.transport = routes, Counters(), MemoryTransport()
        position = message(MSG_POSITION, SESSION, 1, 1, iteracion=1, coords=[1, 2, 3])
        worker.dispatch(position, encode_message(position), routes.clients[1].addr, received=1)
        worker.apply_grants({1: 2})
        self.assertEqual(worker.stats.protocol_errors, 1)
        self.assertEqual(worker.transport.messages, [])
        position['origen'] = 2
        worker.dispatch(position, encode_message(position), routes.clients[2].addr, received=3)
        self.assertEqual(worker.deferred_count, 1)
        worker.apply_grants({2: 2})
        self.assertEqual(worker.deferred_count, 0)
        self.assertEqual(worker.transport.messages[-1][1], routes.clients[1].addr)

    async def test_neighbor_ack_allowed_before_own_start_but_spoofing_rejected(self):
        _, routes = snapshot()
        server = ParallelServer(workers=1, host='127.0.0.1', num_clients=2,
                                neighbors_per_group=2, total_iterations=3, enforce_runtime=False)
        worker = server.workers[0]
        worker.routes, worker.stats, worker.transport = routes, Counters(), MemoryTransport()
        packet = ack()
        worker.dispatch(packet, encode_message(packet), routes.clients[2].addr)
        self.assertEqual(len(worker.transport.messages), 1)
        with self.assertRaises(ProtocolError):
            worker.dispatch(packet, encode_message(packet), ('127.0.0.1', 9999))
        packet['datos']['destino'] = 2
        with self.assertRaises(ProtocolError):
            worker.dispatch(packet, encode_message(packet), routes.clients[2].addr)

    async def test_group_done_ack_not_accepted_before_actual_control_send(self):
        old, routes = snapshot()
        runtime = ParallelServer(workers=1, host='127.0.0.1', num_clients=2,
                                 neighbors_per_group=2, total_iterations=3, enforce_runtime=False)
        worker = runtime.workers[0]
        worker.stats = Counters()
        server = RegistrationServer(worker, **runtime.config)
        server.instance_id, server.clients, server.groups = INSTANCE, old.clients, old.groups
        server.start_sent_at = {1: 2}
        server.finished_groups = {1}
        server.done_sent_at = {1: 5}
        with self.assertRaises(ProtocolError):
            server.accept_notice(Notice(MSG_GROUP_DONE_ACK, 1, 1, 4, INSTANCE))
        self.assertFalse(server.done_acks)
        server.accept_notice(Notice(MSG_GROUP_DONE_ACK, 1, 1, 6, INSTANCE))
        self.assertEqual(server.done_acks, {1})

    async def test_control_inbox_coalesces_and_rejects_conflicts(self):
        inbox = ControlInbox(1)
        first = Notice(MSG_METRICS, 1, 1, 1, mean=2, iterations=3, unique_acks=3)
        duplicate = Notice(MSG_METRICS, 1, 1, 2, mean=2, iterations=3, unique_acks=3)
        self.assertEqual(inbox.put(first), ('new', True))
        self.assertEqual(inbox.put(duplicate), ('duplicate', False))
        conflict = Notice(MSG_METRICS, 1, 1, 3, mean=99, iterations=3, unique_acks=3)
        self.assertEqual(inbox.put(conflict), ('conflict', False))
        self.assertEqual(inbox.put(Notice(MSG_START_ACK, 1, 1, 4, INSTANCE)), ('full', False))
        self.assertEqual(inbox.drain(), (duplicate,))
        self.assertEqual(inbox.put(first), ('new', True))

    async def test_runtime_guard_rejects_gil_or_non_free_threaded_build(self):
        for patch in ({'platform':'darwin','free_threaded_build':True,'gil_enabled':False},
                      {'platform':'linux','free_threaded_build':False,'gil_enabled':False},
                      {'platform':'linux','free_threaded_build':True,'gil_enabled':True}):
            with self.assertRaises(RuntimeError):
                require_runtime(patch)
        require_runtime({'platform':'linux','free_threaded_build':True,'gil_enabled':False})


class ParallelNetworkTests(unittest.IsolatedAsyncioTestCase):
    def check(self, server, generator, n=4, v=2, iterations=3):
        self.assertTrue(generator.success, generator.result)
        self.assertTrue(server.result['success'], server.result)
        self.assertEqual(server.result['registered_clients'], n)
        self.assertEqual(server.result['started_clients'], n)
        self.assertEqual(server.result['completed_clients'], n)
        self.assertEqual(server.result['group_done_acks'], n)
        self.assertEqual(generator.result['completed_iterations'], n*iterations)
        self.assertEqual(generator.result['counters']['ack_unique'], n*iterations*(v-1))
        self.assertEqual(server.result['counters']['protocol_errors'], 0)
        self.assertTrue(server.result['routing_publication']['published_before_workers'])
        self.assertTrue(server.result['routing_publication']['all_workers_ready_before_start'])
        rows = server.result['workers']
        self.assertEqual(len({row['native_thread_id'] for row in rows}), len(rows))
        self.assertEqual(len({row['loop_id'] for row in rows}), len(rows))
        self.assertEqual(len({row['pid'] for row in rows}), 1)
        self.assertTrue(all(worker.finished.is_set() for worker in server.workers))
        self.assertFalse(server.result['real_parallelism_eligible'])

    async def test_workers_1_2_4_6_8_small_simulations(self):
        for workers in (1,2,4,6,8):
            with self.subTest(workers=workers):
                server, generator = await parallel_case(workers=workers)
                self.check(server, generator)
                self.assertEqual(sum(row['data_received'] for row in server.result['workers']), 4*3*2)

    async def test_v25_with_multiple_workers(self):
        server, generator = await parallel_case(workers=4, n=25, v=25, iterations=2)
        self.check(server, generator, 25, 25, 2)

    async def test_registration_loss_during_socket_transition_and_lost_start(self):
        pending = {MSG_REGISTER_ACK, MSG_START}
        registrations = {}
        lock = threading.Lock()
        def wrapper(index, transport):
            def decide(msg, addr):
                with lock:
                    if msg['tipo'] == MSG_REGISTER_ACK:
                        body = msg['datos']
                        old = registrations.setdefault(body['token'], body['cliente_id'])
                        assert old == body['cliente_id']
                    if msg['tipo'] in pending:
                        pending.remove(msg['tipo'])
                        return 'drop'
            return FaultTransport(transport, decide)
        server, generator = await parallel_case(workers=4, wrapper=wrapper)
        self.check(server, generator)
        self.assertEqual(len(registrations), 4)
        self.assertGreaterEqual(server.result['counters']['registration_duplicates'], 1)
        self.assertGreaterEqual(server.result['counters']['retransmissions'], 1)

    async def test_lost_start_ack_and_group_done_ack_are_recovered(self):
        factory = faulty_clients(drop_first(MSG_START_ACK, MSG_GROUP_DONE_ACK))
        server, generator = await parallel_case(workers=4, factory=factory)
        self.check(server, generator)
        self.assertGreaterEqual(server.result['counters']['retransmissions'], 2)

    async def test_lost_metrics_and_final_group_control_are_recovered(self):
        pending = {MSG_GROUP_DONE, MSG_METRICS_ACK}
        lock = threading.Lock()
        def wrapper(index, transport):
            def decide(msg, addr):
                with lock:
                    if msg['tipo'] in pending:
                        pending.remove(msg['tipo'])
                        return 'drop'
            return FaultTransport(transport, decide)
        factory = faulty_clients(drop_first(MSG_METRICS))
        server, generator = await parallel_case(workers=2, factory=factory, wrapper=wrapper)
        self.check(server, generator)


class BenchmarkCertificationTests(unittest.TestCase):
    def complete(self):
        n = 25
        counters = dict(datagrams_sent=6100, datagrams_received=3225,
                        retransmissions=0, timeouts=0, protocol_errors=0, socket_errors=0, failures=0)
        server = dict(n=n,v=25,i=5,session='run',success=True,
            registered_clients=n,started_clients=n,completed_clients=n,group_done_acks=n,
            real_parallelism_eligible=True,
            runtime_initial={'free_threaded_build':True,'gil_enabled':False},
            runtime_final={'free_threaded_build':True,'gil_enabled':False},
            telemetry={'namespace_processes_peak':1},
            routing_publication={'published_before_workers':True,'all_workers_ready_before_start':True},
            workers=[{'pid':1,'native_thread_id':10,'loop_id':20,'socket_inode':30,
                      'startup_deferred_drops':0,'socket_observed':{'drops':0}}],counters=counters)
        generator = dict(n=n,v=25,i=5,session='run',success=True,
            registered_clients=n,started_clients=n,completed_clients=n,completed_iterations=125,
            server_certificate_validated=True,counters=dict(counters,datagrams_sent=3225,
                datagrams_received=6100,ack_unique=3000))
        return server,generator

    def test_certification_requires_iterations_and_unique_acks_not_only_sockets(self):
        from support import certify_snapshot as certify
        server,generator=self.complete()
        generator['completed_iterations']=0
        generator['counters']['ack_unique']=0
        valid,stable,errors=certify(server,generator,[0,0],25,1)
        self.assertFalse(valid)
        self.assertFalse(stable)
        self.assertTrue(errors)

    def test_complete_recovered_run_is_valid_but_not_loss_free_stable(self):
        from support import certify_snapshot as certify
        server,generator=self.complete()
        self.assertEqual(certify(server,generator,[0,0],25,1),(True,True,[]))
        server['counters']['retransmissions']=1
        server['counters']['timeouts']=1
        self.assertEqual(certify(server,generator,[0,0],25,1),(True,False,[]))
        server['counters']['protocol_errors']=1
        self.assertFalse(certify(server,generator,[0,0],25,1)[0])
