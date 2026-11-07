"""Mide un generador sin modificarlo ni crear otro proceso generador."""
import argparse
import os
import resource
import sys
import time
from run_clients import main as generator_main
from runtime import write_json


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--usage', required=True)
    args, rest = parser.parse_known_args()
    if rest and rest[0] == '--':
        rest = rest[1:]
    start = time.perf_counter()
    before = resource.getrusage(resource.RUSAGE_SELF)
    code = 1
    try:
        code = generator_main(rest)
        return code
    finally:
        elapsed = time.perf_counter() - start
        after = resource.getrusage(resource.RUSAGE_SELF)
        cpu = after.ru_utime + after.ru_stime - before.ru_utime - before.ru_stime
        rss = after.ru_maxrss / 1024 if sys.platform == 'darwin' else after.ru_maxrss
        write_json(args.usage, {'pid': os.getpid(), 'cpu_seconds': cpu,
            'cpu_percent': 100 * cpu / elapsed, 'peak_rss_kib': rss,
            'elapsed_seconds': elapsed, 'exit_code': code,
            'python': sys.version, 'platform': sys.platform,
            'fd_limits': list(resource.getrlimit(resource.RLIMIT_NOFILE))})


if __name__ == '__main__':
    raise SystemExit(main())
