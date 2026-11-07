"""Protecciones e instrumentación; no son referencias de rendimiento."""
import asyncio
import tempfile
import unittest
from pathlib import Path

from support import *
from benchmark import cpu_window,cpu_generator,resource_forecast,SustainedLossGuard
from distributed_config import make_manifest
from lab import compose_document
from runtime import RateLimiter,write_json


class ObservationTests(unittest.IsolatedAsyncioTestCase):
    async def test_metered_rate_limiter_keeps_pacing_and_counts_only_grants(self):
        limiter=RateLimiter(100,observe=True)
        begin=asyncio.get_running_loop().time()
        for _ in range(8):await limiter.acquire()
        self.assertGreaterEqual(asyncio.get_running_loop().time()-begin,.065)
        self.assertEqual(limiter.observation['grants'],8)
        self.assertGreater(limiter.observation['actual_sleep_seconds'],0)
        pending=asyncio.create_task(limiter.acquire());await asyncio.sleep(.001)
        pending.cancel();await asyncio.gather(pending,return_exceptions=True)
        self.assertEqual(limiter.observation['grants'],8)

    async def test_unobserved_limiter_retains_default_interface(self):
        limiter=RateLimiter(1000)
        await limiter.acquire()
        self.assertIsNone(limiter.observation)


class BenchmarkTests(unittest.TestCase):
    def test_cpu_uses_only_simulation_window_and_requested_threads(self):
        samples=[dict(at=0,thread_cpu_percent={}),dict(at=1,thread_cpu_percent={'1':20,'2':60}),
                 dict(at=2,thread_cpu_percent={'1':10,'2':90}),dict(at=20,thread_cpu_percent={'1':0,'2':0})]
        self.assertEqual(cpu_window(samples,.5,1.5),90)
        self.assertEqual(cpu_window(samples,.5,1.5,{'2'}),75)
        self.assertIsNone(cpu_window(samples,None,1))

    def test_generator_cpu_interpolates_cumulative_cpu(self):
        samples=[dict(at=0,cpu_seconds=0),dict(at=1,cpu_seconds=.5),dict(at=2,cpu_seconds=1)]
        self.assertEqual(cpu_generator(samples,.2,1.8),50)

    def test_compose_varies_workers_but_retains_resources_and_protocol(self):
        m=make_manifest(10000,5)
        for w in (1,2,4,6,8):
            doc=compose_document(Path('/tmp/case'),m,'s','g',workers=w,generator_memory='2g',observe=True)
            s=doc['services']['server'];argv=s['command']
            self.assertEqual(argv[argv.index('--workers')+1],str(w))
            self.assertEqual(s['cpus'],6)
            self.assertEqual(len(doc['services']),6)
            self.assertNotIn('ports',s)
            self.assertEqual(doc['services']['g1']['cpus'],1)
            self.assertIn('--observe',doc['services']['g1']['command'])

    def test_resource_gate_separates_memory_from_ephemeral_ports(self):
        history=[dict(valid=True,n=30000,generator_peak_rss_kib={'g':2814*1024}),
                 dict(valid=True,n=40000,generator_peak_rss_kib={'g':3677*1024})]
        self.assertTrue(resource_forecast(60000,7.2*1024**3,7.75*1024**3,history)['allowed'])
        self.assertFalse(resource_forecast(80000,7.2*1024**3,7.75*1024**3,history)['allowed'])
        ports=resource_forecast(160000,128*1024**3,128*1024**3,history)
        self.assertFalse(ports['allowed'])
        self.assertIn('efímero',ports['reason'])

    def test_loss_guard_does_not_abort_for_a_single_recovered_burst(self):
        guard=SustainedLossGuard()
        with tempfile.TemporaryDirectory() as tmp:
            folder=Path(tmp);(folder/'server').mkdir()
            for stamp in range(10):
                write_json(folder/'server/live.json',dict(at=stamp,sockets={'1':dict(drops=2000)}))
                self.assertIsNone(guard(folder,stamp,[]))

    def test_loss_guard_stops_sustained_kernel_discards(self):
        guard=SustainedLossGuard()
        with tempfile.TemporaryDirectory() as tmp:
            folder=Path(tmp);(folder/'server').mkdir()
            reason=None
            for stamp in range(8):
                write_json(folder/'server/live.json',dict(at=stamp,sockets={'1':dict(drops=500*stamp)}))
                reason=guard(folder,stamp,[])
            self.assertIn('sostenidos',reason)

    def test_aborted_run_keeps_real_counters_without_inventing_complete_throughput(self):
        from benchmark import enrich_rows
        with tempfile.TemporaryDirectory() as tmp:
            folder=Path(tmp);(folder/'failed').mkdir()
            g=dict(counters=dict(retransmissions=1,timeouts=1,tx_by_type={'POSICION':4}),
                   completed_iterations=2,phase_clock={'finished':4},observations=[])
            s=dict(counters=dict(retransmissions=2,timeouts=3,rx_by_type={'POSICION':15}),
                   phase_clock=dict(first_start=1,finished=4),
                   workers=[dict(native_thread_id=2,socket_observed={'drops':7},data_received=20)],telemetry={'samples':[]})
            write_json(folder/'failed/result.json',dict(server=s,generators={f'g{j}':g for j in range(5)}))
            original=dict(case='failed',valid=False,n=25,position_per_second=None)
            row=enrich_rows(folder,[original])[0]
            self.assertEqual((row['retries'],row['timeouts'],row['drops_S']),(7,8,7))
            self.assertEqual(row['raw']['position_emitted'],20)
            self.assertEqual(row['raw']['position_received'],15)
            self.assertEqual(row['raw']['data_received'],20)
            self.assertIsNone(row['position_per_second'])
            self.assertNotIn('raw',original)

    def test_missing_generator_report_is_marked_as_partial_counter_coverage(self):
        from benchmark import enrich_rows
        with tempfile.TemporaryDirectory() as tmp:
            folder=Path(tmp);(folder/'failed').mkdir()
            write_json(folder/'failed/result.json',dict(server={},generators={'g1':dict(counters={})}))
            row=enrich_rows(folder,[dict(case='failed',valid=False,n=25,position_per_second=None)])[0]
            self.assertEqual(row['raw']['generator_reports_available'],1)
            self.assertIsNone(row['raw']['position_emitted'])


class CancellationTests(unittest.IsolatedAsyncioTestCase):
    async def test_sigterm_cancels_generator_cooperatively_and_keeps_partial_json(self):
        import json,signal,sys
        from support import ROOT
        received=asyncio.Event()
        class DropServer(asyncio.DatagramProtocol):
            def datagram_received(self,data,addr):received.set()
        transport,_=await asyncio.get_running_loop().create_datagram_endpoint(DropServer,
            local_addr=('127.0.0.1',0))
        try:
            with tempfile.TemporaryDirectory() as tmp:
                path=Path(tmp);result=path/'generator.json';usage=path/'usage.json'
                process=await asyncio.create_subprocess_exec(sys.executable,str(ROOT/'generator_profile.py'),
                    '--usage',str(usage),'--','--host','127.0.0.1','--port',str(transport.get_extra_info('sockname')[1]),
                    '-N','100','-V','25','-I','5','--session',SESSION,'--server-result',str(path/'not-present.json'),
                    '--result',str(result),'--deadline','20',cwd=ROOT,
                    stdout=asyncio.subprocess.PIPE,stderr=asyncio.subprocess.PIPE)
                try:
                    await asyncio.wait_for(received.wait(),3)
                    process.send_signal(signal.SIGTERM)
                    stdout,stderr=await asyncio.wait_for(process.communicate(),5)
                finally:
                    if process.returncode is None:
                        process.kill();await process.wait()
                self.assertEqual(process.returncode,130,(stdout,stderr))
                partial=json.loads(result.read_text())
                self.assertFalse(partial['success'])
                self.assertEqual(partial['error'],'generador cancelado')
                self.assertEqual(json.loads(usage.read_text())['exit_code'],130)
                self.assertNotIn(b'event loop is already running',stderr)
        finally:
            transport.close()
