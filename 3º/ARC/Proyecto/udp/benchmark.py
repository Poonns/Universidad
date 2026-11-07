"""Benchmark de capacidad W6/W8; controlador externo, sin participar en DATA."""
from __future__ import annotations

import argparse
from collections import Counter
import json
import math
from pathlib import Path
import signal
import statistics
import time

from lab import command,prepare,read,run,ROOT
import subprocess
import sys
import re
from observability import vm_memory
from runtime import percentiles,write_json



def cpu_window(samples,start,end,tids=None):
    if start is None or end is None or end<=start:
        return None
    total=covered=0.;previous=None
    for sample in samples:
        if previous is not None:
            overlap=max(0,min(end,sample['at'])-max(start,previous))
            values=sample['thread_cpu_percent']
            total+=sum(v for k,v in values.items() if tids is None or k in tids)*overlap
            covered+=overlap
        previous=sample['at']
    return total/covered if covered else None


def cpu_generator(samples,start,end):
    pairs=[(s['at'],s['cpu_seconds']) for s in samples]
    if len(pairs)<2 or start is None or end is None:
        return None
    left=max(start,pairs[0][0]);right=min(end,pairs[-1][0])
    if right<=left:return None
    def interpolate(at):
        for a,b in zip(pairs,pairs[1:]):
            if a[0]<=at<=b[0]:return a[1]+(b[1]-a[1])*(at-a[0])/(b[0]-a[0])
        raise ValueError('fuera del intervalo observado')
    return 100*(interpolate(right)-interpolate(left))/(right-left)


def metrics(folder,report):
    cfg=read(folder/'config.json');manifest=read(folder/'manifest.json');s=report.get('server') or {};gs=report.get('generators') or {}
    gs={key:g for key,g in gs.items() if g}
    valid=bool(report.get('valid'))
    duration=s.get('simulation_seconds') if valid else None
    a=s.get('simulation_clock',{});start,end=a.get('start'),a.get('end')
    telemetry=s.get('telemetry',{});samples=telemetry.get('samples',[]);workers=s.get('workers',[])
    diagnostic_end=end or (samples[-1]['at'] if samples else None)
    latency=[value for g in gs.values() for value in (g.get('latency_samples_ms') or [])]
    tx=Counter();rx=Counter()
    for g in gs.values():
        tx.update(g.get('counters',{}).get('tx_by_type',{}))
        rx.update(g.get('counters',{}).get('rx_by_type',{}))
    result=dict(case=folder.name,n=cfg['n'],workers=cfg['workers'],load=cfg['load'],
        valid=valid,stable=bool(report.get('stable_for_growth')),errors=report.get('errors',[]),
        server_error=s.get('error'),simulation_seconds=duration,wall_seconds=report.get('wall_seconds'),
        position_per_second=cfg['n']*manifest['i']/duration if duration else None,
        offered_position_per_second=tx['POSICION']/duration if duration else None,
        received_position_per_second=s.get('counters',{}).get('rx_by_type',{}).get('POSICION',0)/duration if duration else None,
        offered_data_per_second=(tx['POSICION']+tx['ACK'])/duration if duration else None,
        server_data_rx_per_second=s.get('data_rx_per_second') if valid else None,
        server_data_tx_per_second=s.get('data_tx_per_second') if valid else None,
        generator_data_rx_per_second=(rx['POSICION']+rx['ACK'])/duration if duration else None,
        latency=percentiles(latency),latency_samples=len(latency),
        server_cpu_simulation_percent=cpu_window(samples,start,end),
        server_cpu_observed_until_end_or_abort_percent=cpu_window(samples,start,diagnostic_end),
        worker_cpu_until_end_or_abort_percent=[cpu_window(samples,start,diagnostic_end,{str(w['native_thread_id'])}) for w in workers],
        peak_simultaneous_worker_cpu_percent=s.get('simultaneous_worker_cpu_peak_percent'),
        worker_cpu_simulation_percent=[cpu_window(samples,start,end,{str(w['native_thread_id'])}) for w in workers],
        generator_cpu_simulation_percent={key:cpu_generator(g.get('observations',[]),start,end) for key,g in gs.items()},
        server_cpu_lifetime_percent=telemetry.get('process_cpu_percent'),
        worker_lag=[w.get('event_loop_lag',{}) for w in workers],
        generator_lag={key:g.get('event_loop_lag',{}) for key,g in gs.items()},
        rx_distribution=[w.get('data_rx_percent') for w in workers],
        server_peak_rss_kib=telemetry.get('peak_rss_kib'),
        generator_peak_rss_kib={key:(read(folder/key/'usage.json') or {}).get('peak_rss_kib') for key in gs},
        server_cgroup_memory=telemetry.get('cgroup_memory',{}),
        generator_cgroup_memory={key:g.get('cgroup_memory',{}) for key,g in gs.items()},
        server_cpu_throttle=telemetry.get('cgroup_cpu_delta',{}),
        generator_cpu_throttle={key:g.get('cgroup_cpu_delta',{}) for key,g in gs.items()},
        limiter={key:g.get('limiter_observations',{}) for key,g in gs.items()},
        server_queue_peaks=[w.get('socket_observed',{}) for w in workers],
        transport_queue_peaks=[w.get('transport_queue_max_bytes') for w in workers],
        start_publication=s.get('start_authorization_publication'),phase_clock=s.get('phase_clock'),
        generator_phase_clock={key:g.get('phase_clock',{}) for key,g in gs.items()},
        retries=report.get('retries'),timeouts=report.get('timeouts'),drops_S=report.get('socket_drops_S'),
        drops_G=report.get('socket_rcvbuf_errors_G'),
        completed_iterations=sum(g.get('completed_iterations',0) for g in gs.values()),
        ack_unique=sum(g.get('counters',{}).get('ack_unique',0) for g in gs.values()),
        source=str(folder))
    if valid:
        if len(latency)!=cfg['n']*manifest['i']:raise ValueError('faltan muestras RTT individuales')
    vm_available=[row.get('vm',{}).get('memavailable_bytes') for row in samples]
    result['vm_available_min_bytes']=min((v for v in vm_available if v is not None),default=None)
    return result


class SustainedLossGuard:
    def __init__(self):
        self.previous=0;self.growth_start=None;self.last_growth=None;self.begin_drops=0

    def __call__(self,folder,elapsed,inspect):
        live=read(folder/'server/live.json')
        if not live:return None
        at=live['at'];drops=sum(s.get('drops',0) for s in live.get('sockets',{}).values())
        if drops>self.previous:
            if self.growth_start is None or self.last_growth is None or at-self.last_growth>1.5:
                self.growth_start=at;self.begin_drops=self.previous
            self.last_growth=at
        if self.last_growth is not None and at-self.last_growth>1.5:
            self.growth_start=None
        self.previous=drops
        if self.growth_start is not None and at-self.growth_start>=5 and drops-self.begin_drops>=1000:
            return 'parada preventiva: descartes UDP sostenidos >=5 s y >=1000 paquetes'
        return None


def resource_forecast(n,available,capacity,previous,port_count=28232):
    """Estimación, no capacidad medida; protege VM y puertos antes de un nivel nuevo."""
    local=math.ceil(n/5)
    if local+128>port_count:
        return dict(allowed=False,reason='N local supera el rango efímero por IP',local_clients=local,ports=port_count)
    # Pendiente RSS conservadora a partir de las referencias ya ejecutadas.
    points=[r for r in previous if r.get('valid') and r.get('generator_peak_rss_kib')]
    slopes=[sum(r['generator_peak_rss_kib'].values())*1024/r['n'] for r in points
            if all(v is not None for v in r['generator_peak_rss_kib'].values())]
    groups={}
    for row in points:
        values=list(row['generator_peak_rss_kib'].values())
        if all(value is not None for value in values):
            groups.setdefault(row['n'],[]).append(sum(values)*1024)
    ordered=sorted((count,statistics.median(values)) for count,values in groups.items())
    if len(ordered)>=2:
        (n1,m1),(n2,m2)=ordered[-2:]
        per_client=max(0,(m2-m1)/(n2-n1));intercept=max(0,m2-per_client*n2)
    else:
        per_client=max(slopes,default=100*1024);intercept=0
    predicted=intercept+per_client*n+max(384*1024**2,7*1024*n+128*1024**2)
    return dict(allowed=predicted+1024**3<available,reason='RAM proyectada frente a MemAvailable y reserva 1 GiB',
        predicted_bytes=predicted,available_bytes=available,capacity_bytes=capacity,
        rss_bytes_per_client=per_client,intercept_bytes=intercept,model_points=ordered,
        local_clients=local,ports=port_count)


def read_image(tag):
    return json.loads(command(['docker','image','inspect',tag]).stdout)[0]['Id']


def environment(image):
    probe='import json;from pathlib import Path;from observability import vm_memory;print(json.dumps(dict(vm=vm_memory(),ports=Path("/proc/sys/net/ipv4/ip_local_port_range").read_text().split(),rmem_max=Path("/proc/sys/net/core/rmem_max").read_text(),wmem_max=Path("/proc/sys/net/core/wmem_max").read_text())))'
    data=json.loads(command(['docker','run','--rm','--entrypoint','python3.14t',image,'-c',probe]).stdout)
    info=json.loads(command(['docker','info','--format','{{json .}}']).stdout)
    data['docker']={k:info[k] for k in ('NCPU','MemTotal','ServerVersion','KernelVersion','Architecture')}
    return data


def host_memory():
    if sys.platform!='darwin':return {}
    result={}
    for key,argv in [('swap',['sysctl','vm.swapusage']),('pressure',['memory_pressure'])]:
        p=subprocess.run(argv,capture_output=True,text=True,timeout=5)
        if key=='swap':
            m=re.search(r'used = ([0-9.]+)M',p.stdout)
            if m:result['swap_used_bytes']=int(float(m[1])*1024**2)
        else:
            m=re.search(r'System-wide memory free percentage: (\d+)%',p.stdout)
            if m:result['available_percent']=int(m[1])
    return result


def extra_metrics(report):
    s=report.get('server') or {};gs=report.get('generators') or {}
    counters=[s.get('counters',{})]+[g.get('counters',{}) for g in gs.values() if g]
    events=[]
    for name,source in [('server',s.get('telemetry',{}).get('samples',[]))]+[
            (name,g.get('observations',[])) for name,g in gs.items() if g]:
        events.extend((r['at'],name,r.get('memory',{}).get('current_bytes',0) or 0) for r in source)
    current={};peak=0
    for _,name,value in sorted(events):
        current[name]=value
        peak=max(peak,sum(current.values()))
    return dict(retries=sum(c.get('retransmissions',0) for c in counters),
        timeouts=sum(c.get('timeouts',0) for c in counters),
        protocol_errors=sum(c.get('protocol_errors',0) for c in counters),
        socket_errors=sum(c.get('socket_errors',0) for c in counters),
        observed_combined_cgroup_peak_bytes=peak,
        metrics_received=s.get('completed_clients'),metrics_confirmed=sum(g.get('metrics_confirmed_clients',0) for g in gs.values() if g),
        group_done_acks=s.get('group_done_acks'),generators_released=len((s.get('distributed') or {}).get('released_generators',[])))


class ResourceGuard(SustainedLossGuard):
    def __init__(self,folder):
        super().__init__();self.folder=folder;self.host=[];self.last_host=-1;self.base_swap=None
    def __call__(self,folder,elapsed,inspect):
        reason=super().__call__(folder,elapsed,inspect)
        if elapsed-self.last_host>=5:
            memory=host_memory();self.last_host=elapsed
            if self.base_swap is None:self.base_swap=memory.get('swap_used_bytes',0)
            self.host.append(dict(elapsed=elapsed,**memory))
            write_json(folder/'host-memory.json',self.host)
            if memory.get('available_percent',100)<8:
                return 'parada preventiva: memoria disponible macOS <8%'
            if memory.get('swap_used_bytes',0)-self.base_swap>256*1024**2:
                return 'parada preventiva: crecimiento swap macOS >256 MiB'
        return reason


def summary_lines(rows,tests='consultar resultados/validacion/tests-*.log'):
    valid=[r for r in rows if r['valid']]
    if not valid:return ['Tests: '+tests,'Resultado: INCOMPLETO; ninguna ejecución válida en esta campaña.']
    maximum=max(r['n'] for r in valid)
    comparable=[n for n in {r['n'] for r in valid}
                if all(sum(r['n']==n and r['workers']==w for r in valid)>=3 for w in (6,8))]
    comparison=max(comparable) if comparable else None
    if comparison is None:
        return ['Tests: '+tests,f'Clientes validados: {maximum}',
                'Workers: sin selección concluyente; faltan tres referencias comparables por perfil.']
    top=[r for r in valid if r['n']==comparison]
    means={w:statistics.mean(r['position_per_second'] for r in top if r['workers']==w)
           for w in (6,8) if any(r['workers']==w for r in top)}
    selected=max(means,key=means.get);best=max((r for r in valid if r['workers']==selected),key=lambda r:r['position_per_second'])
    lat=best['latency'];cpu=best['server_cpu_simulation_percent']
    return ['Tests: '+tests,f'Workers de referencia: {selected} (comparación N={comparison}, tres referencias por perfil; cuota S=6 CPU)',
        f'Clientes validados: {maximum}; referencias válidas: {len(valid)}/{len(rows)}',
        f'Mejor ejecución válida: {best["case"]}',f'Throughput: {best["position_per_second"]:.3f} POSITION/s',
        f'RTT medio/p95/p99: {lat["mean_ms"]:.3f}/{lat["p95_ms"]:.3f}/{lat["p99_ms"]:.3f} ms',
        f'RAM conjunta estimada: máximo campaña {max(r["observed_combined_cgroup_peak_bytes"] for r in valid)/1024**3:.3f} GiB; mejor caso {best["observed_combined_cgroup_peak_bytes"]/1024**3:.3f} GiB',
        f'CPU S: {cpu:.1f}%; CPU G: '+', '.join(f'{v:.1f}%' for v in best['generator_cpu_simulation_percent'].values()),
        f'Retransmisiones/timeouts: {best["retries"]}/{best["timeouts"]}',
        f'Duración simulación/caso: {best["simulation_seconds"]:.3f}/{best["wall_seconds"]:.3f} s',
        'Límites: consultar stop/preflight; laboratorio Docker compartido, PCs físicos pendientes.']


def campaign(args):
    out=Path(args.out).resolve();out.mkdir(parents=True,exist_ok=False)
    rows=[];status=dict(status='running',valid=False,rows=rows,not_executed=[],errors=[])
    write_json(out/'summary.json',status)
    history=read(ROOT/'resultados/referencia.json') or []
    try:
        images=dict(server_image=read_image(args.server_image),generator_image=read_image(args.generator_image))
        write_json(out/'methodology.json',dict(images=images,workers=[6,8],clients=args.clients,repetitions=args.repeats,
            quota_cpu_server=6,quota_cpu_each_generator=1,server_memory='2g',generator_memory='4g',
            request_rate=32000,position_window=64,registration_rate=2000,registration_window=64,retry_rate=100,
            control_rate=2000,rto=.25,max_rto=2,max_attempts=8,jitter=.2,deadline=600,
            source_sha256={p.name:__import__('hashlib').sha256(p.read_bytes()).hexdigest() for p in ROOT.glob('*.py')}))
        for index,n in enumerate(args.clients):
            env=environment(images['server_image']);host=host_memory()
            port_range='10240 65535' if math.ceil(n/5)+128>28232 else None
            ports=65535-10240+1 if port_range else int(env['ports'][1])-int(env['ports'][0])+1
            forecast=resource_forecast(n,env['vm']['memavailable_bytes'],env['vm']['memtotal_bytes'],history+rows,ports)
            write_json(out/f'preflight-n{n}.json',dict(environment=env,host=host,forecast=forecast,port_range=port_range))
            if not forecast['allowed']:
                status['not_executed']=args.clients[index:];status['stop']=forecast;break
            level=[]
            for rep in range(1,args.repeats+1):
                for w in ((6,8) if rep%2 else (8,6)):
                    folder=out/f'n{n}-w{w}-r{rep}'
                    prepare(folder,n,**images,workers=w,deadline=600,observe=True,generator_memory='4g',
                        minimum_vm_available_mib=1024,port_range=port_range,
                        request_rate=32000,registration_rate=2000,retry_rate=100,position_window=64,registration_window=64)
                    print(json.dumps(dict(event='start',case=folder.name)),flush=True)
                    guard=ResourceGuard(folder);report=run(folder,stop_check=guard)
                    row=metrics(folder,report);row.update(extra_metrics(report));row['repetition']=rep
                    row=enrich_rows(folder.parent,[row])[0]
                    row['host_memory']=guard.host;row['status']=report['status']
                    rows.append(row);level.append(row);write_json(folder/'metrics.json',row)
                    write_json(out/'summary.json',status)
                    print(json.dumps(dict(event='done',case=folder.name,valid=row['valid'],stable=row['stable'],
                        position_per_second=row['position_per_second'],errors=row['errors'])),flush=True)
                    if report.get('cancelled'):
                        raise KeyboardInterrupt
                    if any('macOS' in e or 'RAM disponible' in e for e in row['errors']+[row.get('server_error') or '']):
                        status['not_executed']=args.clients[index+1:];status['stop']='protección de memoria';return status
            if not all(r['valid'] and r['stable'] for r in level):
                status['not_executed']=args.clients[index+1:];status['stop']='nivel no estable; no aumentar N';break
        status['valid']=bool(rows) and all(r['valid'] for r in rows)
    except KeyboardInterrupt:
        status['errors'].append('campaña cancelada de forma controlada')
    except Exception as exc:
        status['errors'].append(f'{type(exc).__name__}: {exc}')
    finally:
        status['status']='complete' if status['valid'] else 'incomplete'
        write_json(out/'summary.json',status)
        print('\n'.join(summary_lines(rows)),flush=True)
    return status


def main(argv=None):
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out',required=True)
    parser.add_argument('--clients',type=int,nargs='+',default=[80000,100000,120000,140000,160000,180000,200000])
    parser.add_argument('--repeats',type=int,default=3)
    parser.add_argument('--server-image',default='arc-udp-server:final')
    parser.add_argument('--generator-image',default='arc-udp-generator:final')
    args=parser.parse_args(argv)
    if args.repeats<1 or any(n<=0 or n%25 for n in args.clients):parser.error('N múltiplo de 25 y repeticiones positivas')
    previous={}
    def cancel(*_):raise KeyboardInterrupt
    for sig in (signal.SIGINT,signal.SIGTERM):previous[sig]=signal.signal(sig,cancel)
    try:return 0 if campaign(args)['valid'] else 1
    finally:
        for sig,handler in previous.items():signal.signal(sig,handler)




def enrich_rows(folder,source_rows):
    rows=[]
    for source in source_rows:
        row=dict(source)
        report=json.loads((Path(folder)/row['case']/'result.json').read_text())
        server=report.get('server') or {}
        generators={k:v for k,v in (report.get('generators') or {}).items() if v}
        sc=server.get('counters',{});gc=[g.get('counters',{}) for g in generators.values()]
        counters=gc+[sc]
        row['retries']=sum(c.get('retransmissions',0) for c in counters)
        row['timeouts']=sum(c.get('timeouts',0) for c in counters)
        row['drops_S']=sum(w.get('socket_observed',{}).get('drops',0) for w in server.get('workers',[]))
        row['drops_G']={k:g.get('udp_namespace_delta',{}).get('RcvbufErrors',0) for k,g in generators.items()}
        clock=server.get('phase_clock') or {};begin=clock.get('first_start')
        ends=[clock.get('finished')]+[g.get('phase_clock',{}).get('finished') for g in generators.values()]
        end=max((v for v in ends if v is not None),default=None)
        seconds=end-begin if begin is not None and end is not None and end>begin else None
        emitted=sum(c.get('tx_by_type',{}).get('POSICION',0) for c in gc) if len(gc)==5 else None
        received=sc.get('rx_by_type',{}).get('POSICION',0)
        row['raw']=dict(generator_reports_available=len(gc),position_emitted=emitted,position_received=received,
            data_emitted=sum(c.get('tx_by_type',{}).get('POSICION',0)+c.get('tx_by_type',{}).get('ACK',0) for c in gc),
            data_received=sum(w['data_received'] for w in server.get('workers',[]))
                if server.get('workers') and all('data_received' in w for w in server['workers']) else None,
            registered_clients=server.get('registered_clients'),
            metrics=server.get('completed_clients'),group_done_acks=server.get('group_done_acks'),
            completed_iterations=sum(g.get('completed_iterations',0) for g in generators.values()),
            ack_unique=sum(c.get('ack_unique',0) for c in gc),
            protocol_errors=sum(c.get('protocol_errors',0) for c in counters),
            socket_errors=sum(c.get('socket_errors',0) for c in counters),failures=sum(c.get('failures',0) for c in counters),
            observed_interval_seconds=seconds,
            observed_position_emitted_per_second=emitted/seconds if emitted is not None and seconds else None,
            observed_position_received_per_second=received/seconds if seconds else None,
            interval_definition='primer INICIO a último cierre de proceso observado; no SimulationDuration si abortada')
        samples=server.get('telemetry',{}).get('samples',[]);raw_end=clock.get('finished')
        row['diagnostic_cpu_percent']=cpu_window(samples,begin,raw_end)
        row['diagnostic_worker_cpu_percent']=[cpu_window(samples,begin,raw_end,{str(w['native_thread_id'])})
            for w in server.get('workers',[])]
        row['diagnostic_generator_cpu_percent']={k:cpu_generator(g.get('observations',[]),begin,raw_end)
            for k,g in generators.items()}
        if row['valid']:
            assert row['raw']['completed_iterations']==row['n']*5
            assert row['raw']['ack_unique']==row['n']*120
            assert row['raw']['protocol_errors']==row['raw']['socket_errors']==0
        else:
            assert row['position_per_second'] is None
        rows.append(row)
    return rows



if __name__=='__main__':raise SystemExit(main())
