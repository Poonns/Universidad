# Errores corregidos, validación y resultados

Este documento reúne el diagnóstico, las correcciones comprobadas y las mediciones relevantes. La arquitectura vigente y los comandos están en [ARQUITECTURA.md](ARQUITECTURA.md); el contrato de mensajes y sus garantías están en [PROTOCOLO.md](PROTOCOLO.md). No se conservan documentos separados de fases ni gráficas.

## Errores y correcciones

| Problema | Corrección implementada y comprobación |
|---|---|
| Contar cualquier ACK como progreso | Conjunto de vecinos únicos por origen/iteración; destino, grupo y endpoint registrados. Duplicados y ACK tardíos no completan otra operación. |
| Perder mensajes de otra iteración al consumir una cola | Estado independiente por operación y callbacks que atienden posiciones/ACK desfasados correctamente. |
| Registro retransmitido asigna otro ID | Token estable por cliente y clave generador/token; REGISTRO_ACK autoritativo e idempotente. |
| INICIO antes de confirmar identidad o antes de preparar receptores | Retención/validación en C; rutas inmutables publicadas antes de abrir peers y readiness de todos antes de INICIO. |
| Aviso multihilo de INICIO adelantado por DATA | Autorización por cliente y timestamp del envío real; cuarentena acotada para mensajes válidos adelantados, rechazo de los prematuros. |
| Tormenta de retransmisiones | REQUEST compartido también para retries, subpresupuesto RETRY, ventana y backoff/jitter acotado; reenvíos POSICION selectivos. |
| METRICAS perdida | ARQ existente con confirmación, intentos y timeout acotados; retransmitir sin duplicar estadística. |
| METRICAS_ACK perdida y cierre prematuro | Exigir METRICAS_ACK explícito. Recordar GRUPO_FIN pero retener GRUPO_FIN_ACK hasta confirmar métricas. Una prueba de pérdida expuso la carrera; ahora se recupera antes de cerrar S. |
| Duplicar o contradecir métricas | S guarda una media por ID; un duplicado igual recibe ACK, uno contradictorio se rechaza y no altera la estadística. |
| Cliente/G rápido cierra antes de sus vecinos | Mantener endpoints hasta certificado global, FIN de todos los G y retención de respuestas terminales. |
| G confunde N local con N global | Manifiesto coherente: crea N local pero todos registran N global. Los IDs y grupos siguen el orden real de registro. |
| Depender de un resultado local de S en cinco PCs | Certificado y cierre por el mismo servicio UDP, correlacionados por sesión/instancia/nonce/secuencia/SHA. El JSON queda como evidencia local. |
| S espera métricas sin diagnóstico útil | Deadline finito, resultado incompleto y listado exacto de IDs registrados sin métricas/inicio/cierre, más cantidad no registrada. |
| Cancelar G levantando KeyboardInterrupt dentro del loop | SIGTERM solicita cancelación cooperativa de la tarea principal; asyncio termina cleanup y persiste parcial/usage. |
| Cancelar el benchmark o fallar Docker pierde el resultado | Checkpoint inicial inválido; captura de cancelación/excepciones; cleanup tolerante a errores y publicación final en `finally`. Un fallo de cleanup invalida la ejecución y conserva la causa original. |
| Eventos de cancelación inflan grupos en un parcial | Contar clientes realmente completados, no eventos despertados para detener tareas; aserciones de cero cierres en cancelación. |
| Checkpoint anuncia éxito antes de terminar | Estado `finalizing` siempre inválido; se publica éxito después de validación y cleanup. |
| JSON truncado o corrupto durante escritura | Temporal junto al destino, flush/fsync y reemplazo atómico. Serialización fallida conserva el JSON anterior y retira el temporal. |
| Mezclar CPU o duración total con DATA/RTT | Clocks separados, CPU por thread en el intervalo de simulación, muestras RTT individuales y tasas configuradas/emitidas/recibidas. |
| Atribuir RAM/puertos del laboratorio al máximo de S | Preflight cuantificado y niveles no ejecutados explícitos. Rango ampliado únicamente en namespaces G; sin sysctl global. |

La ruta DATA de recepción, validación y forwarding directo conserva la arquitectura previa: un proceso S, W threads/loops/sockets SO_REUSEPORT, sin cola central DATA. Se amplía la corrección de confirmación/cierre y persistencia, no se añade TCP ni otro servidor alternativo.

## Pruebas ejecutadas

Suite actual: **81/81 en macOS/Python3.9.6 (9.478 s)** y **81/81 en Linux/CPython3.14.10 free-threaded (9.198 s)**, GIL desactivado tras la suite Linux. [Logs y entorno](resultados/validacion).

El ajuste final también evita contar como cierres los eventos despertados al cancelar; cuenta clientes realmente completados. Un resumen interrumpido selecciona workers con tres referencias comparables por perfil. Se verificaron de nuevo cancelación y caída de G con25 clientes, sin más estrés.

Cobertura: codec y tipos inválidos, identidad/configuración, registro duplicado y pérdida de REGISTRO_ACK, INICIO adelantado/perdido y su ACK perdido, POSICION/ACK perdidos, ACK duplicados/tardíos, retry cancelado por ACK tardío, agotamiento ARQ, V1/V25 y parámetros V32, cliente rápido/lento, dos y cinco generadores y grupos mezclados, certificado de otra sesión/instancia rechazado, cierre remoto y pérdida de sus respuestas.

Se añadieron pruebas específicas para METRICAS ausente permanentemente, METRICAS_ACK temporal o permanentemente perdida, métricas duplicadas, diagnóstico del servidor multihilo, fallo de escritura atómica, cancelación antes del primer resultado, caída inesperada de G, timeout de Docker durante cleanup y checkpoint inválido durante la finalización. La política de producción conserva RTO/ventanas/presupuestos; las fixtures usan perfiles cortos para probar fallos sin esperar minutos.

La sensibilidad temporal del antiguo test_v25 procedía de combinar un timeout interior de fase de2 s con un deadline exterior mayor. El fixture de correctness usa fase10 s/deadline15 s, con RTO de producción; no se convierte un timeout en éxito ni se relajan las pruebas de pérdida persistente. Se conservó el diagnóstico histórico en el respaldo y la suite actual pasa con ese perfil.

Validaciones adicionales en Docker real:

- W6 y W8, N25/V25/I5/G5: protocolo completo, métricas explícitamente confirmadas, cero errores y cierre remoto.
- Cancelación controlada durante una simulación: resultado `cancelled`, inválido, JSON parciales de los cinco G y diagnóstico S persistidos.
- SIGKILL deliberado a G3: resultado `incomplete`, inválido; falta el JSON de G3, quedan su exit137, los otros informes y el diagnóstico S. `OOMKilled=false`: fue una caída inyectada, no falta de memoria.
- Contenedor de prueba con rango10240–65535, seguido por otro sin el ajuste: este último conservó32768–60999. Se comprobó el aislamiento del namespace antes de usarlo para capacidad.

Estas verificaciones funcionales no se mezclan con referencias de rendimiento. Un exit137 no se etiqueta automáticamente como OOM. Los timeouts no equivalen necesariamente a pérdidas físicas y un retry recuperado no invalida por sí solo el protocolo.

## Entorno y metodología de capacidad

MacBook Pro M4 Pro, 24 GiB físicos, macOS26.6.2 arm64. Docker29.9.2, Linux7.0.14-linuxkit aarch64, 12 CPU de VM. El límite anterior de ~7,75 GiB impedía seguir la campaña previa. Se guardó el ajuste anterior de Docker y se estableció `MemoryMiB=18432`: **18.830.053.376 bytes utilizables (~17,54 GiB)** observados después del reinicio. CPU, swap y protecciones OOM no se desactivaron. macOS partió con swap usado0.

Comparación vigente: solamente W6 frente a W8, un S y cinco G. Cuota S6CPU/2GiB; cada G1CPU/4GiB, idénticas entre W. Los 4GiB por G permiten más clientes que el antiguo límite2GiB y se aplican a toda la campaña nueva. Los máximos sumados no son reservas. No se atribuye a W un cambio de cuota.

Constantes: V25/I5, REQUEST global32000/s, ventanaPOSITION64, registro2000/s/ventana64, RETRY100/s, CONTROL2000/s, RTO.25→2 s, máximo8 envíos, jitter20 %, buffers solicitados S256KiB/G64KiB por dirección. Linux concede S512KiB/G128KiB; rmem_max/wmem_max4MiB se mantienen. nofile65536, deadline y timeoutG600 s. Misma imagen congelada para cada rol durante todas las referencias.

El plan alterna W6/W8 entre tres repeticiones por N; la ejecución efectiva se detuvo por petición del usuario y sus recuentos se detallan debajo. La subida sigue80K/100K/120K/140K/160K y solo niveles superiores permitidos por recursos. No se repiten barridos de W1/2/4, tasas ni ventanas. Los perfiles abortados se guardan pero no aportan throughput válido.

El preflight proyecta RAM desde RSS medido y exige1GiB adicional disponible; comprueba puertos con reserva128 por G. Para niveles que no caben en28.232 puertos se configura55296 únicamente en namespaces G. S aborta ante VM MemAvailable<1GiB durante0,5 s; el controlador corta crecimiento sostenido de drops UDP durante≥5 s/≥1000 paquetes, macOS disponible<8 % o crecimiento de swap>256MiB. Se guardan muestras de VM/macOS, RSS/cgroup, throttling y OOM.

Cada referencia válida debe acreditar N registros/inicios/métricas/METRICAS_ACK/cierres, NI operaciones, NI(V−1) ACK únicos y FIN de cinco G. Para aumentar N se exige además cero drops observados/de cuarentena y retries/POSITION≤0,001. Se informa toda recuperación. Un nivel no estable impide subir; un preflight negativo es **no ejecutado**, no prueba fallida ni incapacidad de S.

Las tasas DATA usan primer INICIO→última METRICA de S; el wall time incluye registro, cierre, retención terminal, Docker y exportación. El rollout de INICIO se solapa con simulación. CPU S se pondera por el intervalo DATA; CPU G se obtiene de su CPU acumulada. p95/p99 se calculan con todas las muestras individuales de cada ejecución, no promediando percentiles de cinco G. Las medias entre repeticiones se acompañan de rango.

RAM conjunta cgroup es una estimación por muestreo dentro de la misma VM, conservando la última muestra disponible de cada proceso; puede sobreestimar la simultaneidad al terminar un G y no captura todos los microbursts. Se informa también el mínimo MemAvailable de la VM y la suma de máximos RSS registrados, sin asumir simultaneidad ni capturar toda la serialización final de S, sin equiparar estas medidas. La disponibilidad VM incluye efectos de kernel/cache y el supervisor host también consume recursos. No se pretende acreditar una NIC física ni paralelismo de cinco PCs a partir del bridge Docker.

## Resultados de capacidad

**Parada solicitada por el usuario después de finalizar la referencia en curso.** Se completaron **17/17 referencias válidas y estables**: seis a80K, seis a100K y cinco a120K. Se comprobaron8.400.000 operaciones y201.600.000 ACK únicos en el conjunto. Todas las referencias recibieron N métricas, sus confirmaciones explícitas, N cierres de grupo y FIN de los cinco generadores. Hubo cero drops observados, errores de protocolo/socket y OOM en estas17 referencias.

**Conclusión acreditada: podemos superar100.000 clientes; se han validado120.000.** W6 completó3/3 a120K y W8 completó2/2. La tercera W8/120K y N140K/160K/180K/200K **no se ejecutaron por petición de parada**, no por fallo de S ni por demostrar falta de RAM. El techo absoluto del servidor no se buscó.

La interrupción del controlador se solicitó tras persistir el resultado válido de la última W6/120K. Su JSON se certificó de nuevo y se incorporó al resumen; el registro original del controlador permanece en el respaldo. `stopped_by_user` describe la campaña interrumpida y `all_executed_valid=true` las17 referencias efectivamente ejecutadas. No se convierte una referencia fallida en válida ni se incluye una preparación sin ejecución.

Tabla: media [mínimo–máximo] de referencias completas. CPU100 % equivale a un core. RTT muestra las medias entre referencias de su media/p95/p99 globales; no representa percentiles de una mezcla ficticia de generadores. La columna recuperación suma las repeticiones de esa fila.

| N | W | Válidas | POSITION/s | Simulación (s) | RTT media / p95 / p99 (ms) | CPU S (%) | Retries / timeouts |
|---:|---:|---:|---|---|---|---:|---:|
| 80000 | 6 | 3/3 | 5098,077 [5057,270–5121,063] | 78,463 [78,109–79,094] | 11,435 / 18,174 / 24,367 | 533,4 | 23 / 137 |
| 80000 | 8 | 3/3 | 5025,674 [4938,805–5073,460] | 79,603 [78,842–80,991] | 11,544 / 19,036 / 27,354 | 560,7 | 99 / 702 |
| 100000 | 6 | 3/3 | 4859,317 [4821,461–4892,464] | 102,899 [102,198–103,703] | 12,008 / 18,991 / 26,545 | 529,4 | 423 / 1611 |
| 100000 | 8 | 3/3 | 4675,004 [4518,813–4842,280] | 107,037 [103,257–110,649] | 12,402 / 20,165 / 30,628 | 547,4 | 578 / 2229 |
| 120000 | 6 | 3/3 | 4422,385 [4032,216–4625,036] | 136,228 [129,729–148,802] | 13,189 / 20,818 / 29,953 | 512,6 | 1218 / 2945 |
| 120000 | 8 | 2/2 | 4598,120 [4564,103–4632,138] | 130,495 [129,530–131,461] | 12,621 / 20,413 / 29,033 | 545,7 | 749 / 1947 |

**Referencia operativa elegida: W6.** La comparación completa de mayor N con tres referencias por perfil es100K: W6 ofrece4859,317 frente a4675,004 POSITION/s, una diferencia observada del3,94 %, y usa menos CPU. A80K su ventaja media observada fue1,44 %. Bajo la misma cuota de seis CPU, W8 registra mucho más throttling; no se le atribuyen dos cores adicionales.

A120K **no hay un ganador concluyente**: W8 tiene dos referencias y W6 tres; una W6 válida obtuvo4032,216 POSITION/s, se conserva en la media y no se descarta como outlier. La media de W8 a120K supera la de W6 un3,97 %, con rangos que se solapan y recuentos diferentes. La selección de W6 utiliza la comparación completa a100K y su validación3/3 a120K, no oculta ese resultado. Las muestras escasas y el hardware compartido no acreditan significación estadística ni escalado lineal.

W cambia threads, loops, sockets y buffers agregados, manteniendo cada buffer individual y la cuota total. No se aisló el coste exclusivo de cada componente ni se perfilaron funciones durante las referencias.

### Recursos y duración completa

| N | W | Wall caso, media [mín–máx] (s) | Máximo RSS S observado (MiB) | Mayor RSS de un G (MiB) | Mínimo disponible VM (GiB) | Throttling S, nr_throttled mín–máx |
|---:|---:|---|---:|---:|---:|---:|
| 80000 | 6 | 208,327 [207,029–209,307] | 456,1 | 1467,7 | 8,977 | 1–1 |
| 80000 | 8 | 226,304 [225,736–226,786] | 616,6 | 1474,1 | 8,690 | 505–553 |
| 100000 | 6 | 252,875 [252,487–253,631] | 541,8 | 1817,6 | 7,059 | 8–13 |
| 100000 | 8 | 276,038 [273,490–278,455] | 708,6 | 1824,0 | 6,655 | 572–698 |
| 120000 | 6 | 301,747 [297,007–310,338] | 611,4 | 2173,5 | 5,066 | 2–17 |
| 120000 | 8 | 325,935 [325,388–326,483] | 721,7 | 2174,8 | 4,801 | 767–878 |

La RAM conjunta cgroup máxima estimada de la campaña fue12,107 GiB; el mínimo MemAvailable VM fue4,801 GiB. El indicador disponible de macOS llegó como mínimo al44 %; no se observó uso de swap en macOS ni en las muestras de la VM. No se activaron las guardas de RAM, swap o pérdida sostenida. Estos valores no convierten120K en un límite de memoria del laboratorio: la subida se detuvo voluntariamente.

CPU por worker a100K (media de tres): W6=91,6 /85,3 /85,4 /85,7 /85,6 /85,4 %; W8=76,8 /65,5 /65,9 /66,0 /66,4 /66,0 /65,4 /64,8 %. CPU de G durante DATA: intervalos exactos por referencia en el resumen. El consumo lifetime incluye otras fases y no sustituye esa medida.

A120K, tras la última METRICA, los ACK de grupo tardaron47,112–55,630 s con W6 y82,526–83,338 s con W8; permanece CONTROL=2000/s y retención terminal13,75 s. La duración completa media fue301,747 s con W6 y325,935 s con W8. Se distinguen de la simulación y del RTT; no se atribuye ese cierre a exportación de JSON.

### Mejor ejecución válida por throughput

`n80000-w6-r2`: **5121,063 POSITION/s**, N80000,400000 operaciones y9600000 ACK únicos. RTT medio/p95/p99=11,399/18,083/24,566 ms; simulación=78,109 s, caso completo=209,307 s. CPU S=535,2 %, G=91,1–93,0 %. RAM conjunta cgroup estimada=8,074 GiB; RSS S observado=456,1 MiB y G=1449,4–1461,5 MiB por proceso. Retries/timeouts=3/26, drops/errores=0.

DATA S RX=128026,765/s y TX=245811,327/s. DATA TX global G+S=373838,092 datagramas/s, contando cada envío una vez; no se suman de nuevo sus recepciones. Son contadores de aplicación/kernel, no captura física de NIC. **Estas cifras corresponden a80K y no se presentan como rendimiento medido a120K.**

### Recuperación y límites de la interpretación

La campaña suma3090 retransmisiones y9571 timeouts, todos contabilizados. Un timeout puede cancelarse al llegar un ACK tardío antes de obtener permiso de retry, por eso no se exige igualdad entre ambas cifras. El criterio heredado de subida, retries/POSITION≤0,001, se cumplió en las17 referencias. Nunca se saltó REQUEST/RETRY para recuperar pérdida.

Ejemplo120K/W6/r1:371 retries POSICION selectivos en600000 operaciones, METRICAS enviadas exactamente120000 veces, cero drops observados. Los generadores tuvieron picos de lag durante DATA de504–808 ms, superiores al RTO inicial250 ms. Es compatible con respuestas tardías por planificación; no demuestra que el kernel perdiese esos paquetes ni permite atribuir la causa a GC sin profiling. Generación, estado retenido por cliente, cuotas, servicio S y pacing CONTROL forman parte del perfil medido.

N140K+ queda **sin medir**, al igual que el rendimiento en cinco PCs físicos. La ampliación namespaced de puertos se validó por configuración y aislamiento; no se ejecutaron160K ni32K clientes por G en esta entrega. No se proclama una capacidad física de NIC ni un máximo universal de CPython/S.

### Evidencia y reproducción

[Resumen de todas las referencias](resultados/resumen.json), [configuración congelada](resultados/configuracion.json), [índice de evidencia y respaldo bruto](resultados/evidencia.json) y [auditoría final](resultados/verificacion.json). Se conservan informes comprimidos completos de las seis referencias100K, las cinco120K y la mejor80K; el resto de brutos queda recuperable fuera del repositorio. El resumen conserva todas las filas, sus rangos y recuperaciones.

Las17 mediciones usaron las imágenes ahora conservadas con etiquetas `:medido`, fijadas por Image ID. La entrega `:final` incorpora después únicamente la corrección del contador de grupos en parciales y del resumen para seleccionar workers con tres referencias comparables; no cambia DATA, ARQ, codec, registro o cierre de red. `server_parallel.py` se conserva byte a byte respecto al inicio de la consolidación. La generación del informe G ocurre después de DATA/certificado/FIN; las diferencias de versiones se registran, sin afirmar que los binarios son idénticos.

Los fuentes medidos completos y sus hashes están respaldados en `/Users/hectorpons/universidad/3º/ARC/respaldos-udp/fuentes-consolidacion-medidos.tar.gz`; las validaciones previas al último ajuste también están archivadas. Los IDs y hashes quedan en el índice de validación. Las cifras de wall/RSS lifetime pertenecen a la versión medida.

Para repetir el plan con las imágenes medidas conservadas en este Docker:

```bash
python3 benchmark.py --out resultados/repeticion-nueva \
  --clients 80000 100000 120000 --repeats 3 \
  --server-image arc-udp-server:medido --generator-image arc-udp-generator:medido
```

Ese comando completaría tres referencias por perfil si no se interrumpe. En la campaña entregada el usuario pidió parar antes de la tercera W8/120K; no se simula que esa referencia exista. Para probar el código de entrega usar las imágenes `:final` y los comandos de ARQUITECTURA. Comparar solamente perfiles de la misma campaña/entorno.

Los informes retenidos se leen con `gzip` de stdlib:

```bash
python3 - <<'PYTEST'
import gzip, json
from pathlib import Path
for path in sorted(Path('resultados/evidencia').glob('*/result.json.gz')):
    report = json.loads(gzip.open(path, 'rb').read())
    assert report['valid'] and report['server']['simulation_complete']
    n = report['n']
    assert report['server']['completed_clients'] == n
    assert sum(g['metrics_confirmed_clients'] for g in report['generators'].values()) == n
    print(path.parent.name, 'válida', n)
PYTEST
```

El [resumen final de consola](resultados/resumen-final.txt) ocupa12 líneas y distingue máximo N de la mejor referencia por throughput.

```bash
cat resultados/resumen-final.txt
```

## Referencia histórica relevante

Antes de esta consolidación se validaron60K/V25/I5/G5 tres veces, W6, con simulación57,254–59,233 s y un retry/timeout selectivo en una referencia. N10K dio mejor media5912,951 POSITION/s con W6 frente a5682,100 con W8, bajo cuotaS6CPU; W8 registró throttling. El mejor caso individual fue5962,843 POSITION/s. W1/2 o ventanas mayores con W4 no sostuvieron la carga elegida. El limitadorREQUEST2000 previo restringía fuertemente la carga ofrecida; cambiarlo multiplicó throughput sin reescribir recepción.

A60K, confirmar cierres después de la última métrica añadía38,306–41,207 s, además de retención13,75 s. Esta observación motivó separar DATA/control/wall, no esconder el cierre como overhead de exportación. Los niveles80K/100K/160K de aquella campaña no se ejecutaron por RAM/puertos; no fueron simulaciones fallidas. Las cifras históricas pertenecen a sus imágenes/confirmación de métricas anteriores y no reemplazan la validación del código consolidado.

Las primeras validaciones locales de100/250/1000 y la posterior distribución10K/20K/30K/40K demostraron progresivamente fiabilidad y endpoints/remoto. No se mezclan con la comparación actual por cambiar entorno, runtime y perfil. Las cifras Go/TCP pertenecen a implementaciones diferentes.

## Conservación y limpieza

Antes de editar o eliminar se archivaron **1525 archivos** de UDP y el README raíz, con verificación SHA-256 de cada miembro del tar. Respaldo recuperable fuera del repositorio:

`/Users/hectorpons/universidad/3º/ARC/respaldos-udp/antes-consolidacion-20261107-214745.tar.gz`

SHA-256: `233b29a4fc9cac69944296c3d6761051a86c9e5a92c3a54d7a2f78e8cbd2a517`. El manifiesto junto al archivo permite verificar todos los originales. El respaldo del ajuste Docker también queda en ese directorio. No se borran originales externos ni variantes TCP/Go.

Se retiran README/DECISIONES y los documentos FASE_2/3/4 de `udp/`, tras integrar aquí y en ARQUITECTURA/PROTOCOLO su contenido relevante. Se retiran `bench_workers.py`, `bench_phase4.py`, `phase4_analysis.py` y `plot_phase4.py`: hay un único `benchmark.py` funcional y sin dependencia de gráficos. Se conserva `server.py` porque worker0 reutiliza sus clases; `run_test.py` sigue siendo necesario para integración automatizada.

La lista exacta de archivos retirados y sus hashes está en [limpieza.json](resultados/limpieza.json). El [índice de evidencia](resultados/evidencia.json) identifica los informes retenidos, la ruta del respaldo bruto externo y su SHA-256. Los auxiliares de validación anteriores al último ajuste están archivados antes de retirarlos. No quedan PNG/SVG ni directorios de fases en `udp/`.

El repositorio final conserva código, tests, Docker, estos tres Markdown y evidencia mínima: validación, configuración/hashes, repeticiones imprescindibles y datos de la mejor ejecución. Las referencias brutas retiradas tienen respaldo externo. La lista exacta de archivos eliminados se entrega en el manifiesto de limpieza; no se anuncia una eliminación antes de ejecutarla.
