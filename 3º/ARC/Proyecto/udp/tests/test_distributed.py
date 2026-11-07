"""Corrección distribuida; las pruebas API no son benchmarks de PCs independientes."""
import asyncio
import copy
import tempfile
import threading
import time
import unittest
from pathlib import Path

from support import *
from distributed_config import RunManifest, make_manifest
from remote_control import validate_certificate
from server_parallel import ParallelServer


async def distributed_case(counts, v=25, factory=SimulationClient, lost=(), sequential=False):
    n = sum(counts)
    manifest = RunManifest.parse(dict(session=SESSION, n=n, v=v, i=2,
        load=dict(request_rate=5000, registration_rate=5000, retry_rate=1000,
                  position_window=64, registration_window=64),
        generators=[dict(id=f'g{j+1}',clients=c) for j,c in enumerate(counts)]))
    pending, lock = set(lost), threading.Lock()
    def wrapper(index, transport):
        def decide(msg, addr):
            with lock:
                if msg['tipo'] in pending:
                    pending.remove(msg['tipo'])
                    return 'drop'
        return FaultTransport(transport, decide)
    policy = RetryPolicy(.15, .5, 8, .2)
    with tempfile.TemporaryDirectory() as tmp:
        server = ParallelServer(workers=4, host='127.0.0.1', port=0, num_clients=n,
            neighbors_per_group=v, total_iterations=2, session_id=SESSION, manifest=manifest,
            policy=policy, deadline=12, control_rate=5000, enforce_runtime=False,
            result_path=Path(tmp)/'server.json', transport_wrapper=wrapper, terminal_linger=.5)
        thread = threading.Thread(target=server.run)
        thread.start()
        generators, tasks = [], []
        try:
            while not server.bound.is_set():
                if server.stop_flag.is_set():
                    raise AssertionError(server.error)
                await asyncio.sleep(.005)
            for spec in manifest.generators:
                generator = ClientGenerator(*server.address, n, v, 2, local_clients=spec.clients,
                    session_id=SESSION, generator_id=spec.id, manifest=manifest, policy=policy,
                    timeout=8, deadline=12, client_factory=factory, **spec.load())
                generators.append(generator)
                tasks.append(asyncio.create_task(generator.run()))
                if sequential:
                    while len(server.workers[0].server.clients) < sum(counts[:len(generators)]):
                        await asyncio.sleep(.005)
            await asyncio.wait_for(asyncio.gather(*tasks), 13)
            until=time.monotonic()+3
            while thread.is_alive() and time.monotonic()<until:
                await asyncio.sleep(.005)
            assert not thread.is_alive(), 'supervisor no finaliza'
            return server, generators
        finally:
            server.request_stop()
            for task in tasks:
                if not task.done():
                    task.cancel()
            await asyncio.gather(*tasks, return_exceptions=True)
            while thread.is_alive():
                await asyncio.sleep(.005)
            thread.join()


class ManifestTests(unittest.TestCase):
    def test_exact_counts_distinct_ids_and_global_budget(self):
        m=make_manifest(10000,5)
        self.assertEqual(sum(s.clients for s in m.generators),10000)
        self.assertEqual(sum(s.load()['request_rate'] for s in m.generators),2000)
        self.assertEqual(sum(s.load()['position_window'] for s in m.generators),64)
        self.assertEqual(m.digest,RunManifest.parse(m.document).digest)
        # N local no tiene por qué ser múltiplo de V: no asigna grupos por G.
        self.assertEqual([s.clients for s in make_manifest(25,2).generators],[13,12])

    def test_invalid_sum_duplicate_generator_and_over_budget_rejected(self):
        base=make_manifest(25,5).document
        for change in ('sum','duplicate','rate','window','unknown'):
            d=copy.deepcopy(base)
            if change=='sum': d['generators'][0]['clients']+=1
            if change=='duplicate': d['generators'][1]['id']='g1'
            if change=='rate': d['generators'][0]['limits']['request_rate']+=1
            if change=='window': d['generators'][0]['limits']['position_window']=0
            if change=='unknown': d['generators'][0]['limits']['extra']=1
            with self.subTest(change=change),self.assertRaises(ValueError):
                RunManifest.parse(d)

    def test_generator_cannot_override_manifest_or_use_wrong_local_n(self):
        async def check():
            m=make_manifest(25,5)
            for local in (4,6,25):
                with self.assertRaises(ValueError):
                    ClientGenerator('127.0.0.1',8888,25,25,5,local_clients=local,
                        session_id=m.session,generator_id='g1',manifest=m,**m.generator('g1').load())
        asyncio.run(check())

    def test_certificate_rejects_stale_instance_false_counts_and_boolean_integer(self):
        p=dict(role='server',session=SESSION,instance=INSTANCE,n=25,v=25,i=2,success=True,
               simulation_complete=True,registered_clients=25,started_clients=25,
               completed_clients=25,completed_iterations=50,group_done_acks=25,groups_finished=1,
               manifest_sha256='digest')
        validate_certificate(p,SESSION,INSTANCE,25,25,2,'digest')
        for k,bad in (('instance','old'),('session','old'),('group_done_acks',24),
                      ('completed_iterations',49),('manifest_sha256','other'),('groups_finished',True)):
            with self.subTest(key=k),self.assertRaises(ProtocolError):
                validate_certificate(dict(p,**{k:bad}),SESSION,INSTANCE,25,25,2,'digest')

    def test_late_control_response_is_ignored_after_next_query(self):
        from remote_control import GeneratorControl
        async def check():
            m=make_manifest(25,5)
            c=GeneratorControl(m,m.generator('g1'),RetryPolicy())
            c.sequence,c.expected,c.instance=2,MSG_RUN_STATUS_ACK,INSTANCE
            packet=message(MSG_GENERATOR_HELLO_ACK,m.session,generador='g1',nonce=c.nonce,
                manifiesto=m.digest,secuencia=1,instancia=INSTANCE)
            c.datagram_received(encode_message(packet),None)
            self.assertEqual(c.stats.protocol_errors,0)
            self.assertFalse(c.response_event.is_set())
        asyncio.run(check())

    def remote_owner(self):
        from types import SimpleNamespace
        from remote_control import RemoteRun
        from runtime import Counters
        m=make_manifest(25,5)
        sent=[]
        owner=SimpleNamespace(instance_id=INSTANCE,stats=Counters(),registration_keys={},
            worker=SimpleNamespace(runtime=SimpleNamespace(protocol_done=threading.Event(),error=None)))
        owner._send=lambda msg,addr:sent.append(msg)
        return m,RemoteRun(owner,m),sent

    def test_live_generator_id_is_exclusive_and_fin_cannot_precede_global_completion(self):
        from remote_control import RemoteNotice
        m,state,sent=self.remote_owner()
        body=dict(generador='g1',nonce='nonce1',manifiesto=m.digest,secuencia=1,n=25,v=25,i=5,local=5)
        def send(kind,b):state.handle(RemoteNotice(kind,'g1',tuple(b.items()),('127.0.0.1',12345),1))
        send(MSG_GENERATOR_HELLO,body)
        self.assertEqual(sent[-1]['tipo'],MSG_GENERATOR_HELLO_ACK)
        send(MSG_GENERATOR_HELLO,body)
        self.assertEqual(sent[-1]['tipo'],MSG_GENERATOR_HELLO_ACK)
        send(MSG_GENERATOR_HELLO,dict(body,nonce='competitor'))
        self.assertEqual(sent[-1]['tipo'],MSG_ERROR)
        send(MSG_GENERATOR_DONE,dict(body,instancia=INSTANCE))
        self.assertEqual(sent[-1]['tipo'],MSG_ERROR)
        self.assertFalse(state.released)

    def test_registration_requires_generator_hello_and_enforces_local_quota_idempotently(self):
        from remote_control import RemoteNotice
        m,state,sent=self.remote_owner()
        reg=message(MSG_REGISTER,m.session,generador='g1',token='existing',num_clientes=25,
                    num_vecinos=25,total_iteraciones=5)
        with self.assertRaises(ProtocolError):state.before_register(reg)
        hello=dict(generador='g1',nonce='nonce1',manifiesto=m.digest,secuencia=1,n=25,v=25,i=5,local=5)
        state.handle(RemoteNotice(MSG_GENERATOR_HELLO,'g1',tuple(hello.items()),('127.0.0.1',12345),1))
        state.registered['g1']=5
        with self.assertRaises(ProtocolError):state.before_register(reg)
        state.owner.registration_keys[('g1','existing')]=1
        state.before_register(reg)


class DistributedNetworkTests(unittest.IsolatedAsyncioTestCase):
    def check(self,s,gs,counts,v=25):
        self.assertTrue(s.result['success'],s.result)
        n=sum(counts)
        self.assertEqual(s.result['group_done_acks'],n)
        self.assertEqual(s.result['completed_iterations'],n*2)
        self.assertEqual([g.result['n_local'] for g in gs],list(counts))
        self.assertEqual(sum(g.result['counters']['ack_unique'] for g in gs),n*2*(v-1))
        ids=[c.client_id for g in gs for c in g.clients]
        self.assertEqual(sorted(ids),list(range(1,n+1)))
        for g in gs:
            self.assertTrue(g.success,g.result)
            self.assertTrue(g.result['server_certificate_validated'])
            self.assertEqual(g.result['counters']['protocol_errors'],0)
            self.assertGreaterEqual(g.endpoints_closed_at,g.certificate_at)
            self.assertTrue(all(c.closed_event.is_set() for c in g.clients))
        self.assertEqual(s.result['counters']['protocol_errors'],0)
        self.assertEqual(s.result['distributed']['registered_per_generator'],
                         {f'g{j+1}':c for j,c in enumerate(counts)})

    async def test_two_generators_global_n_with_mixed_group(self):
        s,gs=await distributed_case([12,13])
        self.check(s,gs,[12,13])
        self.assertEqual(s.result['distributed']['mixed_groups'],1)

    async def test_five_generators_and_remote_control_losses(self):
        s,gs=await distributed_case([5]*5,lost=(MSG_GENERATOR_HELLO_ACK,
            MSG_RUN_STATUS_ACK,MSG_GENERATOR_DONE_ACK,MSG_REGISTER_ACK,MSG_START))
        self.check(s,gs,[5]*5)
        self.assertEqual(s.result['distributed']['mixed_groups'],1)
        self.assertGreater(sum(g.result['control_counters']['retransmissions'] for g in gs),0)

    async def test_fast_generator_waits_for_slow_and_recovers_final_client_ack(self):
        class SlowClient(SimulationClient):
            async def before_position(self,iteration):
                if self.generator_id=='g2':
                    await asyncio.sleep(.25)
        pending={MSG_GROUP_DONE_ACK}
        class FaultSlow(SlowClient):
            def _send(self,msg):
                if msg['tipo'] in pending:
                    pending.remove(msg['tipo'])
                    return
                super()._send(msg)
        s,gs=await distributed_case([25,25],factory=FaultSlow,sequential=True)
        self.check(s,gs,[25,25])
        self.assertLess(gs[0].local_done_at,gs[1].local_done_at)
        self.assertGreater(gs[0].certificate_at,gs[1].local_done_at)
        self.assertGreater(gs[0].result['global_close_wait_seconds'],.1)
        self.assertGreater(s.result['counters']['retransmissions'],0)
