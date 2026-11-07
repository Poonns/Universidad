"""Dos contenedores G: cierre temprano y pérdidas deliberadas, sin benchmark."""
import argparse
import hashlib
import json
from pathlib import Path
import sys

ROOT=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(ROOT))
from lab import command,prepare,run
from runtime import write_json


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out',required=True)
    args=parser.parse_args()
    images={name:json.loads(command(['docker','image','inspect',tag]).stdout)[0]['Id']
        for name,tag in (('server_image','arc-udp-ft:phase3'),('generator_image','arc-udp-generator:phase3'))}
    folder=prepare(args.out,50,generators=2,deadline=60,**images)
    path=folder/'compose.yaml';doc=json.loads(path.read_text())
    adapter=Path(__file__).with_name('process_generator.py')
    write_json(folder/'fault-adapter.json',dict(path=str(adapter),sha256=hashlib.sha256(adapter.read_bytes()).hexdigest()))
    for name in ('g1','g2'):
        service=doc['services'][name]
        service['entrypoint']=['python3.14t','-u','/app/process_generator.py']
        service['volumes'].append(dict(type='bind',source=str(adapter),
                                      target='/app/process_generator.py',read_only=True))
        service['command']=['--manifest','/config/manifest.json','--generator-id',name,
            '--result','/results/generator.json','--usage','/results/usage.json','--fault-result','/results/fault.json']
        if name=='g1':service['command']+=['--drop-group-ack','--drop-generator-fin']
        else:service['command']+=['--start-delay','.5','--step-delay','.25']
    write_json(path,doc)
    result=run(folder)
    f1=json.loads((folder/'g1/fault.json').read_text())
    f2=json.loads((folder/'g2/fault.json').read_text())
    proof=dict(valid=result['valid'],fast_local_before_slow=f1['local_done_clock']<f2['local_done_clock'],
        fast_kept_endpoints_until_global=f1['endpoints_closed_clock']>f2['local_done_clock'],
        injected=f1['injected'],server_retries=result['server']['counters']['retransmissions'],
        generator_retries=result['generators']['g1']['control_counters']['retransmissions'],
        fast_wait_seconds=result['generators']['g1']['global_close_wait_seconds'])
    proof['passed']=(proof['valid'] and proof['fast_local_before_slow']
        and proof['fast_kept_endpoints_until_global'] and proof['server_retries']>0
        and proof['generator_retries']>0 and set(proof['injected'])=={'client_ack','generator_fin'})
    write_json(folder/'proof.json',proof)
    print(json.dumps(proof),flush=True)
    return 0 if proof['passed'] else 1


if __name__=='__main__':
    raise SystemExit(main())
