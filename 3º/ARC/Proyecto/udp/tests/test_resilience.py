"""Pérdidas de métricas y persistencia del controlador; no son benchmarks."""
import asyncio
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from support import *
from lab import prepare, run
from runtime import write_json


class MetricsTests(unittest.IsolatedAsyncioTestCase):
    async def test_parallel_server_deadline_also_persists_missing_metrics(self):
        from test_parallel import parallel_case
        factory=faulty_clients(lambda msg,addr:'drop' if msg['tipo']==MSG_METRICS else None)
        server,generator=await parallel_case(factory=factory,deadline=.5)
        self.assertFalse(server.result['success'])
        self.assertFalse(generator.success)
        self.assertEqual(server.result['diagnostics']['missing_metrics_client_ids'],[1,2,3,4])
        self.assertEqual(generator.result['group_done_clients'],0)

    async def test_no_metrics_ever_arrive_deadline_lists_every_missing_client(self):
        factory=faulty_clients(lambda msg,addr:'drop' if msg['tipo']==MSG_METRICS else None)
        server,generator=await simulate(client_factory=factory,server_deadline=.5)
        self.assertFalse(server.success)
        self.assertFalse(generator.success)
        self.assertEqual(server.result['diagnostics']['missing_metrics_client_ids'],[1,2,3,4])
        self.assertEqual(generator.result['group_done_clients'],0)
        self.assertEqual(server.result['diagnostics']['metrics_received'],0)
        self.assertIn('deadline',server.result['error'])

    async def test_lost_metrics_ack_requires_explicit_confirmation_and_retry(self):
        server,generator=await simulate(server_filter=drop_first(MSG_METRICS_ACK))
        self.assertTrue(generator.success,generator.error)
        self.assertEqual(generator.result['metrics_confirmed_clients'],4)
        self.assertGreater(server.stats.metrics_duplicates,0)
        self.assertGreater(generator.result['counters']['retransmissions'],0)
        self.assertEqual(len(server.metrics),4)

    async def test_group_done_cannot_mask_permanently_lost_metrics_ack(self):
        server,generator=await simulate(server_filter=lambda msg,addr:
            'drop' if msg['tipo']==MSG_METRICS_ACK else None)
        self.assertFalse(generator.success)
        self.assertEqual(generator.result['metrics_confirmed_clients'],0)
        self.assertGreater(generator.result['counters']['timeouts'],0)

    async def test_duplicate_metrics_do_not_duplicate_statistics(self):
        factory=faulty_clients(lambda msg,addr:'duplicate' if msg['tipo']==MSG_METRICS else None)
        server,generator=await simulate(client_factory=factory)
        self.assertTrue(generator.success)
        self.assertEqual(len(server.metrics),4)
        self.assertEqual(server.result['completed_iterations'],12)
        self.assertGreater(server.stats.metrics_duplicates,0)


class PersistenceTests(unittest.TestCase):
    def test_worker_summary_uses_complete_comparisons_when_campaign_is_stopped(self):
        from benchmark import summary_lines
        rows=[]
        for n,counts in [(100000,{6:3,8:3}),(120000,{6:3,8:2})]:
            for w,count in counts.items():
                for rep in range(count):
                    rows.append(dict(valid=True,n=n,workers=w,position_per_second=
                        (5000 if w==6 else 4500) if n==100000 else (4200 if w==6 else 4600),
                        case=f'n{n}-w{w}-r{rep}',latency=dict(mean_ms=1,p95_ms=2,p99_ms=3),
                        server_cpu_simulation_percent=500,generator_cpu_simulation_percent={'g1':90},
                        observed_combined_cgroup_peak_bytes=1024**3,retries=0,timeouts=0,
                        simulation_seconds=100,wall_seconds=200))
        lines=summary_lines(rows)
        self.assertIn('referencia: 6',lines[1])
        self.assertIn('N=100000',lines[1])
        self.assertIn('120000',lines[2])
        self.assertLessEqual(len(lines),15)

    def test_checkpoint_cannot_advertise_success_before_cleanup_finishes(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder=prepare(Path(tmp)/'case',25)
            write_json(folder/'server/ready.json',{'ready':True})
            containers=[{'Name':f'/{name}','State':dict(Running=False,ExitCode=0,OOMKilled=False)}
                        for name in ('server','g1','g2','g3','g4','g5')]
            observed=[]
            def command(argv,**kw):
                if 'stop' in argv:observed.append(json.loads((folder/'result.json').read_text())['valid'])
                output=json.dumps(containers) if 'inspect' in argv else 'id\n' if '-aq' in argv else ''
                return subprocess.CompletedProcess(argv,0,output,'')
            with patch('lab.command',side_effect=command),patch('lab.certify',return_value=dict(valid=True,stable_for_growth=True,errors=[])):
                report=run(folder)
            self.assertEqual(observed,[False])
            self.assertTrue(report['valid'])

    def test_atomic_write_error_preserves_previous_json_and_cleans_temporary(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'result.json';write_json(p,{'valid':False})
            with self.assertRaises(ValueError):write_json(p,{'bad':float('nan')})
            self.assertEqual(json.loads(p.read_text()),{'valid':False})
            self.assertEqual(list(Path(tmp).iterdir()),[p])

    def test_cancel_before_first_result_always_persists_incomplete_report(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder=prepare(Path(tmp)/'case',25)
            with patch('lab.command',side_effect=KeyboardInterrupt):report=run(folder)
            saved=json.loads((folder/'result.json').read_text())
            self.assertFalse(report['valid'])
            self.assertEqual(saved['status'],'cancelled')
            self.assertEqual(len(saved['diagnostics']['missing_generator_reports']),5)

    def test_unexpected_generator_exit_and_cleanup_timeout_keep_diagnostics(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder=prepare(Path(tmp)/'case',25)
            write_json(folder/'server/ready.json',{'ready':True})
            containers=[{'Name':f'/{name}','State':dict(Running=name!='g3',ExitCode=2 if name=='g3' else 0,
                OOMKilled=False)} for name in ('server','g1','g2','g3','g4','g5')]
            def command(argv,**kw):
                if 'stop' in argv:raise subprocess.TimeoutExpired(argv,1)
                output=json.dumps(containers) if 'inspect' in argv else 'id\n' if '-aq' in argv else ''
                return subprocess.CompletedProcess(argv,0,output,'')
            with patch('lab.command',side_effect=command):report=run(folder)
            saved=json.loads((folder/'result.json').read_text())
            self.assertFalse(saved['valid'])
            self.assertIn('contenedor falló',saved['errors'][0])
            self.assertTrue(saved['cleanup_errors'])
            self.assertTrue(any(r['name']=='/g3' and r['exit_code']==2 for r in saved['diagnostics']['containers']))

    def test_controller_timeout_before_ready_keeps_partial(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder=prepare(Path(tmp)/'case',25)
            with patch('lab.command',side_effect=subprocess.TimeoutExpired('docker',1)):
                report=run(folder)
            self.assertFalse(report['valid'])
            self.assertIn('TimeoutExpired',json.loads((folder/'result.json').read_text())['errors'][0])

    def test_port_extension_is_scoped_only_to_generator_namespaces(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder=prepare(Path(tmp)/'case',25,port_range='10240 65535',neighbors=5,iterations=3)
            doc=json.loads((folder/'compose.yaml').read_text())
            self.assertNotIn('sysctls',doc['services']['server'])
            self.assertEqual(doc['services']['g1']['sysctls'],{'net.ipv4.ip_local_port_range':'10240 65535'})
            m=json.loads((folder/'manifest.json').read_text())
            self.assertEqual((m['v'],m['i']),(5,3))
