"""Adaptador exclusivo del laboratorio de fallos; no se usa en referencias de carga."""
import argparse
import asyncio
import os
import resource
import time

from client import SimulationClient
from distributed_config import RunManifest
from protocol import MSG_GENERATOR_DONE, MSG_GROUP_DONE_ACK, decode_message
from run_clients import ClientGenerator
from runtime import write_json


async def execute(args):
    manifest=RunManifest.read(args.manifest)
    spec=manifest.generator(args.generator_id)
    await asyncio.sleep(args.start_delay)
    injected=[]

    class FaultClient(SimulationClient):
        async def before_position(self,iteration):
            if args.step_delay:
                await asyncio.sleep(args.step_delay)

        def _send(self,msg):
            if args.drop_group_ack and msg['tipo']==MSG_GROUP_DONE_ACK and 'client_ack' not in injected:
                injected.append('client_ack')
                return
            super()._send(msg)

    class ControlTransport:
        def __init__(self,underlying):self.underlying=underlying
        def sendto(self,data,addr=None):
            if (args.drop_generator_fin and decode_message(data)['tipo']==MSG_GENERATOR_DONE
                    and 'generator_fin' not in injected):
                injected.append('generator_fin')
                return
            self.underlying.sendto(data)
        def close(self):self.underlying.close()

    g=ClientGenerator(args.host,8888,manifest.n,manifest.v,manifest.i,
        local_clients=spec.clients,session_id=manifest.session,generator_id=spec.id,
        manifest=manifest,timeout=60,deadline=60,client_factory=FaultClient,
        result_path=args.result,control_wrapper=ControlTransport,**spec.load())
    success=await g.run()
    write_json(args.fault_result,dict(injected=injected,generator=spec.id,
        local_done_clock=g.local_done_at,certificate_clock=g.certificate_at,
        endpoints_closed_clock=g.endpoints_closed_at,success=success,
        note='clocks comparable only within this shared Linux VM'))
    return success


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--manifest',required=True)
    p.add_argument('--generator-id',required=True)
    p.add_argument('--host',default='server')
    p.add_argument('--result',required=True)
    p.add_argument('--usage',required=True)
    p.add_argument('--fault-result',required=True)
    p.add_argument('--start-delay',type=float,default=0)
    p.add_argument('--step-delay',type=float,default=0)
    p.add_argument('--drop-group-ack',action='store_true')
    p.add_argument('--drop-generator-fin',action='store_true')
    args=p.parse_args()
    start=time.perf_counter();before=resource.getrusage(resource.RUSAGE_SELF)
    ok=asyncio.run(execute(args))
    after=resource.getrusage(resource.RUSAGE_SELF);elapsed=time.perf_counter()-start
    write_json(args.usage,dict(pid=os.getpid(),cpu_percent=100*(after.ru_utime+after.ru_stime-
        before.ru_utime-before.ru_stime)/elapsed,peak_rss_kib=after.ru_maxrss,elapsed_seconds=elapsed))
    raise SystemExit(0 if ok else 1)
