"""Laboratorio Compose y certificación global; supervisión, no parte de S/G."""
from __future__ import annotations

import argparse
from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import subprocess
import signal
import time
import uuid

from distributed_config import DEFAULT_LOAD, RunManifest, make_manifest
from runtime import write_json

ROOT=Path(__file__).resolve().parent


def command(argv, timeout=60, check=True):
    p=subprocess.run(argv,text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=timeout)
    if check and p.returncode:
        raise RuntimeError(f'{argv[0]} exit {p.returncode}: {p.stderr}')
    return p


def compose_document(folder, manifest, server_image, generator_image, deadline=600, *, workers=4,
                     server_memory='2g',generator_memory='1g',observe=False,minimum_vm_available_mib=0,
                     port_range=None):
    services={}
    base=dict(user=f'{os.getuid()}:{os.getgid()}',pids_limit=128,
        ulimits=dict(nofile=dict(soft=65536,hard=65536)),networks=['simulation'])
    common=['--manifest','/config/manifest.json','--deadline',str(deadline),'--benchmark']
    manifest_mount=dict(type='bind',source=str(folder/'manifest.json'),target='/config/manifest.json',read_only=True)
    server=dict(base,image=server_image,cpus=6,mem_limit=server_memory,volumes=[manifest_mount,
        dict(type='bind',source=str(folder/'server'),target='/results')],
        command=common+['--workers',str(workers),'--host','0.0.0.0','--control-rate','2000',
                       '--ready-file','/results/ready.json','--result','/results/server.json'])
    if observe:
        server['command']+=['--live-file','/results/live.json','--minimum-vm-available-mib',str(minimum_vm_available_mib)]
    services['server']=server
    for spec in manifest.generators:
        services[spec.id]=dict(base,image=generator_image,cpus=1,mem_limit=generator_memory,
            volumes=[manifest_mount,dict(type='bind',source=str(folder/spec.id),target='/results')],
            depends_on=dict(server=dict(condition='service_started')),
            command=['--usage','/results/usage.json','--']+common+[
                '--generator-id',spec.id,'--host','server','--timeout',str(deadline),
                '--completion-mode','remote','--result','/results/generator.json'])
        if observe:
            services[spec.id]['command']+=['--observe','--export-samples']
        if port_range:
            low,high=map(int,port_range.split())
            if not 1024<=low<high<=65535:
                raise ValueError('rango de puertos inválido')
            services[spec.id]['sysctls']={'net.ipv4.ip_local_port_range':f'{low} {high}'}
    # No se publican puertos host: G->S usa directamente la red bridge de Compose.
    return dict(services=services,networks=dict(simulation=dict(driver='bridge')))


def prepare(folder,n,generators=5,deadline=600,server_image='arc-udp-server:final',
            generator_image='arc-udp-generator:final',workers=6,server_memory='2g',generator_memory='4g',
            observe=False,minimum_vm_available_mib=0,port_range=None,neighbors=25,iterations=5,**load):
    folder=Path(folder).resolve();folder.mkdir(parents=True,exist_ok=False)
    manifest=make_manifest(n,generators,uuid.uuid4().hex,neighbors=neighbors,iterations=iterations,**load)
    write_json(folder/'manifest.json',manifest.document)
    for name in ['server']+[s.id for s in manifest.generators]:
        (folder/name).mkdir()
    doc=compose_document(folder,manifest,server_image,generator_image,deadline,workers=workers,
        server_memory=server_memory,generator_memory=generator_memory,observe=observe,
        minimum_vm_available_mib=minimum_vm_available_mib,port_range=port_range)
    write_json(folder/'compose.yaml',doc)  # JSON es YAML válido para Compose; cero dependencias externas.
    write_json(folder/'config.json',dict(n=n,generators=generators,deadline=deadline,
        server_image=server_image,generator_image=generator_image,host_uid=os.getuid(),
        workers=workers,server_memory=server_memory,generator_memory=generator_memory,observe=observe,
        minimum_vm_available_mib=minimum_vm_available_mib,
        port_range=port_range,
        manifest_sha256=manifest.digest,load=manifest.document['load'],
        source_sha256={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in ROOT.glob('*.py')}))
    return folder


def read(path):
    try:return json.loads(Path(path).read_text())
    except (OSError,ValueError):return None


def certify(folder,inspect):
    folder=Path(folder);m=RunManifest.read(folder/'manifest.json')
    expected_workers=read(folder/'config.json').get('workers',4)
    s=read(folder/'server/server.json')
    gs={spec.id:read(folder/spec.id/'generator.json') for spec in m.generators}
    errors=[]
    if not s or not all(gs.values()):
        return dict(valid=False,stable_for_growth=False,errors=['faltan resultados S/G'],server=s,generators=gs)
    for key in ('registered_clients','started_clients','completed_clients','group_done_acks'):
        if s.get(key)!=m.n:errors.append('S '+key)
    if not s.get('success') or not s.get('simulation_complete'):errors.append('S no completo')
    if s.get('completed_iterations')!=m.n*m.i:errors.append('iteraciones globales S')
    if not all(s['routing_publication'].get(k) for k in ('published_before_workers','all_workers_ready_before_start')):
        errors.append('transición inválida')
    if s.get('configured_workers')!=expected_workers or s['telemetry']['namespace_processes_peak']!=1:
        errors.append('arquitectura del servidor')
    if any(p.get('gil_enabled') is not False or not p.get('free_threaded_build')
           for p in (s['runtime_initial'],s['runtime_final'])):errors.append('GIL/build S')
    rows=s.get('workers',[])
    if len(rows)!=expected_workers or len({r['pid'] for r in rows})!=1 or any(
            len({r[key] for r in rows})!=expected_workers for key in ('native_thread_id','loop_id','socket_inode')):
        errors.append('W threads/loops/sockets no distintos')
    endpoints={r['id']:r for r in s['distributed']['endpoints']}
    ids=[];nat=[];ipset=set();allcounts=[]
    for spec in m.generators:
        g=gs[spec.id]
        if (not g.get('success') or not g.get('server_certificate_validated')
                or g.get('certificate_instance')!=s.get('instance')):errors.append(spec.id+' sin certificado/instancia')
        if (g.get('n_global'),g.get('n_local'),g.get('v'),g.get('i'),g.get('session'),g.get('manifest_sha256'))!=(
                m.n,spec.clients,m.v,m.i,m.session,m.digest):errors.append(spec.id+' configuración')
        for key in ('registered_clients','started_clients','completed_clients','group_done_clients','metrics_confirmed_clients'):
            if g.get(key)!=spec.clients:errors.append(spec.id+' '+key)
        if g.get('completed_iterations')!=spec.clients*m.i or g['counters']['ack_unique']!=spec.clients*m.i*(m.v-1):
            errors.append(spec.id+' iteraciones/ACK')
        if g.get('completion_mode')!='remote' or g.get('os_threads')!=1:errors.append(spec.id+' modelo de generador')
        if g.get('endpoints_closed_seconds',-1)<g.get('certificate_seconds',0):errors.append(spec.id+' cierre prematuro')
        for r in g['endpoints']:
            ids.append(r['id']);ipset.add(r['endpoint'][0])
            if r['id'] not in endpoints or endpoints[r['id']]['generator']!=spec.id:
                errors.append('ID/origen no coincide en S/G');continue
            if r['endpoint']!=endpoints[r['id']]['endpoint']:
                nat.append(dict(id=r['id'],local=r['endpoint'],observed=endpoints[r['id']]['endpoint']))
        allcounts.append(g['counters'])
    if sorted(ids)!=list(range(1,m.n+1)):errors.append('cobertura IDs globales')
    if len(ipset)!=len(m.generators):errors.append('IPs de generador no independientes')
    if s['distributed']['registered_per_generator']!={r.id:r.clients for r in m.generators}:
        errors.append('registro por generador')
    if set(s['distributed']['released_generators'])!={r.id for r in m.generators}:errors.append('cierre remoto incompleto')
    allcounts.append(s['counters'])
    if any(c.get(k,0) for c in allcounts for k in ('protocol_errors','socket_errors','failures')):
        errors.append('errores protocolo/socket/fallos')
    if any(row['State']['ExitCode']!=0 or row['State'].get('OOMKilled') for row in inspect):
        errors.append('exit no cero u OOM')
    drops=sum(r['socket_observed'].get('drops',0) for r in rows)
    generator_drops={name:g.get('udp_namespace_delta',{}).get('RcvbufErrors',0) for name,g in gs.items()}
    retries=sum(c.get('retransmissions',0) for c in allcounts)
    timeouts=sum(c.get('timeouts',0) for c in allcounts)
    aggregate=Counter()
    for c in allcounts:aggregate.update({k:v for k,v in c.items() if type(v) is int})
    valid=not errors
    # Recovery acotado es admisible; no se exige una red UDP sin un solo retry.
    stable=valid and drops==0 and not any(generator_drops.values()) and retries/max(1,m.n*m.i)<=.001 and not any(
        r['startup_deferred_drops'] for r in rows)
    return dict(valid=valid,stable_for_growth=stable,errors=errors,n=m.n,v=m.v,i=m.i,
        generator_count=len(m.generators),workers=expected_workers,simulation_seconds=s['simulation_seconds'],
        position_per_second=m.n*m.i/s['simulation_seconds'],
        generator_ips=sorted(ipset),endpoint_pairs_checked=len(ids),nat_endpoint_mismatches=len(nat),
        nat_examples=nat[:10],mixed_groups=s['distributed']['mixed_groups'],socket_drops_S=drops,
        socket_rcvbuf_errors_G=generator_drops,
        retries=retries,timeouts=timeouts,counters=dict(aggregate),server=s,generators=gs,
        generator_usage={spec.id:read(folder/spec.id/'usage.json') for spec in m.generators})


def run(folder,stop_check=None):
    """Siempre persiste estado incompleto/final; cleanup no puede ocultar el error original."""
    folder=Path(folder).resolve()
    if (folder/'result.json').exists():raise ValueError('caso ya ejecutado; usar directorio nuevo')
    report=dict(valid=False,stable_for_growth=False,status='running',errors=[],cleanup_errors=[])
    started=time.monotonic();created=False;base=None;services={};names=[];rows=[]
    write_json(folder/'result.json',report)
    def safe_command(argv,**kw):
        try:
            result=command(argv,check=False,**kw)
            if result.returncode:
                report['cleanup_errors'].append(f'{argv[0]} exit {result.returncode}: {result.stderr[:1000]}')
            return result
        except (Exception,KeyboardInterrupt) as exc:
            report['cleanup_errors'].append(f'{type(exc).__name__}: {exc}')
            return None
    try:
        cfg=read(folder/'config.json')
        manifest=RunManifest.read(folder/'manifest.json')
        names=[r.id for r in manifest.generators]
        services=read(folder/'compose.yaml')['services']
        project='arc'+uuid.uuid4().hex[:12]
        base=['docker','compose','-p',project,'-f',str(folder/'compose.yaml')]
        write_json(folder/'execution.json',dict(project=project,command=base+['up','-d']))
        command(base+['config','--quiet'])
        created=True;command(base+['up','-d','server'])
        until=time.monotonic()+30
        while not (folder/'server/ready.json').exists():
            ids=command(base+['ps','-aq']).stdout.split()
            current=json.loads(command(['docker','inspect']+ids).stdout) if ids else []
            if any(not c['State']['Running'] for c in current):
                raise RuntimeError('servidor terminó antes de READY')
            if time.monotonic()>until:raise RuntimeError('servidor sin READY')
            time.sleep(.1)
        command(base+['up','-d','--no-deps']+names)
        ids=command(base+['ps','-aq']).stdout.split()
        if ids:write_json(folder/'container-inspect-running.json',json.loads(command(['docker','inspect']+ids).stdout))
        while True:
            ids=command(base+['ps','-aq']).stdout.split()
            rows=json.loads(command(['docker','inspect']+ids).stdout) if ids else []
            if len(rows)==len(services) and all(not r['State']['Running'] for r in rows):break
            if any(not r['State']['Running'] and r['State']['ExitCode']!=0 for r in rows):
                raise RuntimeError('contenedor falló; ver inspect y diagnóstico parcial')
            if stop_check:
                reason=stop_check(folder,time.monotonic()-started,rows)
                if reason:raise RuntimeError(reason)
            if time.monotonic()-started>cfg['deadline']+45:raise RuntimeError('deadline supervisor Compose')
            time.sleep(.5)
        write_json(folder/'container-inspect.json',rows)
        verified=certify(folder,rows)
        report.update(verified)
    except KeyboardInterrupt:
        report.update(cancelled=True)
        report['errors'].append('benchmark cancelado de forma controlada')
    except Exception as exc:
        report['errors'].append(f'{type(exc).__name__}: {exc}')
    finally:
        # Guardar el error ANTES de consultar otra vez Docker. SIGKILL no ejecuta finally.
        report['wall_seconds']=time.monotonic()-started
        report['status']='incomplete'
        write_json(folder/'result.json',dict(report,valid=False,stable_for_growth=False,status='finalizing'))
        try:
            if created:
                safe_command(base+['stop','--timeout','30'],timeout=90)
                ids_result=safe_command(base+['ps','-aq'])
                ids=ids_result.stdout.split() if ids_result else []
                if ids:
                    data=safe_command(['docker','inspect']+ids)
                    if data:
                        try:
                            rows=json.loads(data.stdout);write_json(folder/'container-inspect.json',rows)
                        except (ValueError,OSError) as exc:report['cleanup_errors'].append(str(exc))
                for service in services:
                    log=safe_command(base+['logs','--no-color',service])
                    if log:
                        try:(folder/service/'container.log').write_text(log.stdout+log.stderr)
                        except OSError as exc:report['cleanup_errors'].append(str(exc))
                safe_command(base+['down','--timeout','30'],timeout=90)
        finally:
            report['server']=report.get('server') or read(folder/'server/server.json')
            report['generators']=report.get('generators') or {name:read(folder/name/'generator.json') for name in names}
            report['generator_usage']={name:read(folder/name/'usage.json') for name in names}
            report['diagnostics']=dict(
                missing_generator_reports=[name for name,g in report['generators'].items() if not g],
                server=report['server'].get('diagnostics') if report['server'] else None,
                containers=[dict(name=r.get('Name'),exit_code=r['State']['ExitCode'],
                    oom_killed=r['State'].get('OOMKilled'),running=r['State']['Running']) for r in rows])
            if report['errors'] or report['cleanup_errors']:
                report.update(valid=False,stable_for_growth=False)
            report['status']='valid_recovered' if report['valid'] and report.get('retries',0) else (
                'valid' if report['valid'] else 'cancelled' if report.get('cancelled') else 'incomplete')
            report['wall_seconds']=time.monotonic()-started
            write_json(folder/'result.json',report)
    print(json.dumps(dict(case=folder.name,status=report['status'],valid=report['valid'],
        errors=report['errors'],wall_seconds=report['wall_seconds']),ensure_ascii=False),flush=True)
    return report


def main(argv=None):
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('action',choices=('prepare','run','bench'))
    p.add_argument('--out',required=True)
    p.add_argument('--clients',type=int,nargs='+',default=[10000])
    p.add_argument('--generators',type=int,default=5)
    p.add_argument('--deadline',type=float,default=600)
    p.add_argument('--server-image',default='arc-udp-server:final')
    p.add_argument('--generator-image',default='arc-udp-generator:final')
    p.add_argument('--workers',type=int,choices=(1,2,4,6,8),default=6)
    p.add_argument('--generator-memory',default='4g')
    p.add_argument('--port-range',help='rango efímero del namespace G; ejemplo: "10240 65535"')
    p.add_argument('-V','--neighbors',type=int,default=25)
    p.add_argument('-I','--iterations',type=int,default=5)
    p.add_argument('--observe',action='store_true')
    p.add_argument('--minimum-vm-available-mib',type=int,default=0)
    for key,value in DEFAULT_LOAD.items():
        p.add_argument('--'+key.replace('_','-'),type=int if key.endswith('_window') else float,default=value)
    a=p.parse_args(argv)
    def interrupt(*_):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM,interrupt)
    if a.action=='run':return 0 if run(a.out)['valid'] else 1
    images={}
    for name,tag in (('server_image',a.server_image),('generator_image',a.generator_image)):
        images[name]=json.loads(command(['docker','image','inspect',tag]).stdout)[0]['Id']
    options=dict(generators=a.generators,deadline=a.deadline,**images,
                 workers=a.workers,generator_memory=a.generator_memory,observe=a.observe,
                 port_range=a.port_range,neighbors=a.neighbors,iterations=a.iterations,
                 minimum_vm_available_mib=a.minimum_vm_available_mib,
                 **{key:getattr(a,key) for key in DEFAULT_LOAD})
    if a.action=='prepare':
        print(prepare(a.out,a.clients[0],**options));return 0
    out=Path(a.out).resolve();out.mkdir(parents=True,exist_ok=False)
    reports=[]
    for n in a.clients:
        folder=prepare(out/f'n{n}',n,**options)
        result=run(folder);reports.append(dict(n=n,valid=result['valid'],stable_for_growth=result['stable_for_growth'],
            simulation_seconds=result.get('simulation_seconds'),position_per_second=result.get('position_per_second')))
        write_json(out/'summary.json',dict(cases=reports))
        if not result['stable_for_growth']:
            write_json(out/'stop.json',dict(n=n,reason='nivel no estable; no escalar',errors=result['errors']))
            break
    return 0 if all(r['valid'] for r in reports) else 1


if __name__=='__main__':
    raise SystemExit(main())
