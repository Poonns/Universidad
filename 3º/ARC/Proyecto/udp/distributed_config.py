"""Manifiesto de cantidades/presupuestos; nunca asigna IDs o grupos de clientes."""
from __future__ import annotations

from dataclasses import dataclass
import hashlib
import json
import math
from pathlib import Path

from protocol import identifier, integer, validate_configuration
from runtime import positive

DEFAULT_LOAD = dict(request_rate=2000, registration_rate=2000, retry_rate=100,
                    position_window=64, registration_window=64)


@dataclass(frozen=True)
class GeneratorSpec:
    id: str
    clients: int
    limits: tuple

    def load(self):
        return dict(self.limits)


@dataclass(frozen=True)
class RunManifest:
    session: str
    n: int
    v: int
    i: int
    generators: tuple
    digest: str
    document: dict

    def generator(self, name):
        for spec in self.generators:
            if spec.id == name:
                return spec
        raise ValueError('generador ausente del manifiesto: ' + name)

    @classmethod
    def parse(cls, doc):
        n, v, i = (doc[k] for k in ('n', 'v', 'i'))
        validate_configuration(n, v, i)
        session = identifier(doc['session'], 'sesion')
        rows = doc['generators']
        if not isinstance(rows, list) or not rows:
            raise ValueError('se requieren generadores')
        names = [identifier(r['id'], 'generador') for r in rows]
        counts = [integer(r['clients'], 'N local', 1) for r in rows]
        if len(set(names)) != len(names) or sum(counts) != n:
            raise ValueError('IDs de generador únicos y suma N locales == N global obligatorios')
        global_load = dict(DEFAULT_LOAD, **doc.get('load', {}))
        if set(global_load) != set(DEFAULT_LOAD):
            raise ValueError('parámetro de carga desconocido')
        shares = [dict() for _ in rows]
        for key, total in global_load.items():
            positive(total, key)
            if key.endswith('_window'):
                integer(total, key, len(rows))
                # Al menos un permiso por generador; el resto, por peso N local.
                exact = [(total-len(rows))*c/n for c in counts]
                values = [1 + int(x) for x in exact]
                order = sorted(range(len(rows)), key=lambda j: (-(exact[j] % 1), j))
                for j in order[:total-sum(values)]:
                    values[j] += 1
            else:
                values = [total*c/n for c in counts]
            for j, row in enumerate(rows):
                shares[j][key] = row.get('limits', {}).get(key, values[j])
                if key.endswith('_window'):
                    integer(shares[j][key], key, 1)
                else:
                    positive(shares[j][key], key)
            actual = sum(s[key] for s in shares)
            if not math.isclose(actual, total, rel_tol=1e-12, abs_tol=1e-9):
                raise ValueError('presupuestos locales no suman el global: ' + key)
        for row, limits in zip(rows, shares):
            if set(row.get('limits', {})) - set(DEFAULT_LOAD):
                raise ValueError('límite local desconocido')
            if limits['retry_rate'] > limits['request_rate']:
                raise ValueError('retry_rate local supera request_rate')
        normalized = dict(session=session, n=n, v=v, i=i, load=global_load,
            generators=[dict(id=name, clients=count, limits=limits)
                        for name, count, limits in zip(names, counts, shares)])
        digest = hashlib.sha256(json.dumps(normalized, sort_keys=True, separators=(',', ':')).encode()).hexdigest()
        return cls(session, n, v, i, tuple(GeneratorSpec(name, count, tuple(limits.items()))
            for name, count, limits in zip(names, counts, shares)), digest, normalized)

    @classmethod
    def read(cls, path):
        try:
            return cls.parse(json.loads(Path(path).read_text()))
        except (KeyError, TypeError) as exc:
            raise ValueError('manifiesto incompleto o mal formado') from exc


def make_manifest(n, generators=5, session='laboratorio-001', *, neighbors=25,iterations=5,**load):
    integer(generators, 'generadores', 1)
    if n < generators:
        raise ValueError('N debe permitir al menos un cliente por generador')
    counts = [n//generators + (j < n % generators) for j in range(generators)]
    return RunManifest.parse(dict(session=session,n=n,v=neighbors,i=iterations,load=load,
        generators=[dict(id=f'g{j+1}',clients=c) for j,c in enumerate(counts)]))


if __name__ == '__main__':
    import argparse
    import uuid
    from runtime import write_json
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--clients',type=int,required=True)
    parser.add_argument('--generators',type=int,default=5)
    parser.add_argument('-V','--neighbors',type=int,default=25)
    parser.add_argument('-I','--iterations',type=int,default=5)
    parser.add_argument('--session')
    parser.add_argument('--out',required=True)
    for key,value in DEFAULT_LOAD.items():
        parser.add_argument('--'+key.replace('_','-'),type=int if key.endswith('_window') else float,default=value)
    args=parser.parse_args()
    manifest=make_manifest(args.clients,args.generators,args.session or uuid.uuid4().hex,
        neighbors=args.neighbors,iterations=args.iterations,
        **{key:getattr(args,key) for key in DEFAULT_LOAD})
    if Path(args.out).exists():
        parser.error('el manifiesto ya existe; usar otro nombre/sesión')
    write_json(args.out,manifest.document)
    print(manifest.digest)
