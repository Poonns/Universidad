"""Lecturas procfs/cgroup de baja frecuencia. No crea procesos, threads o sockets."""
from pathlib import Path
import sys


def scalar(path):
    try:
        value=Path(path).read_text().strip()
        return int(value) if value!='max' else None
    except (OSError,ValueError):
        return None


def vm_memory():
    result={}
    try:
        for row in Path('/proc/meminfo').read_text().splitlines():
            name,value,*_=row.split()
            if name in ('MemTotal:','MemAvailable:','SwapTotal:','SwapFree:'):
                result[name[:-1].lower()+'_bytes']=int(value)*1024
    except (OSError,ValueError):
        pass
    return result


def memory_accounting():
    if sys.platform!='linux':
        return {}
    return dict(current_bytes=scalar('/sys/fs/cgroup/memory.current'),
        peak_bytes=scalar('/sys/fs/cgroup/memory.peak'),
        limit_bytes=scalar('/sys/fs/cgroup/memory.max'),
        swap_current_bytes=scalar('/sys/fs/cgroup/memory.swap.current'))
