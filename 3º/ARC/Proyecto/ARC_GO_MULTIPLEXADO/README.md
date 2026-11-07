# ARC_GO_MULTIPLEXADO

## 1. Qué es

Simulación académica UDP escrita en Go. Clientes lógicos multiplexados ejecutan registro autoritativo → START → POSITION a vecinos mediante el servidor → ACK reales → METRICS → cierre fiable por grupo. Cada cliente guarda estado compacto y comparte sockets, workers y timers.

## 2. Resultado principal

**160.000 clientes locales, V=25 e I=5, estables en 3/3 ejecuciones:** aproximadamente 149,251 s, 5.360 POSITION completas/s, RTT de 22,651 ms y RSS total de 78,318 MiB. Cada ejecución transportó 59.680.000 datagramas, sin descartes, retransmisiones naturales, timeouts ni errores de socket.

Configuración validada: un servidor con seis workers y tres shards TX; dos generadores con tres workers cada uno. La ventana POSITION agregada es 128. Este perfil se configura explícitamente; los valores por defecto conservan el perfil inicial.

## 3. Estructura

```text
ARC_GO_MULTIPLEXADO/
├── README.md / go.mod
├── *.go                 biblioteca de protocolo, servidor y generadores
├── *_test.go            pruebas de la biblioteca
├── cmd/                 ejecutables arc-go y supervisor arc-bench
├── internal/            medición y certificados
├── docs/                cinco documentos principales y assets/
└── archivo/             histórico comprimido y SHA256SUMS.txt
```

Los fuentes de la raíz mantienen su ubicación porque pertenecen al mismo paquete Go. Solo se utiliza la biblioteca estándar. El histórico completo queda dentro del archivo comprimido.

## 4. Compilar

Desde este directorio, con Go ≥1.23.0; validación local realizada con Go 1.27.1 en macOS arm64:

```bash
mkdir -p /tmp/arc-bin
go build -o /tmp/arc-bin/arc-go ./cmd/arc-go
go build -o /tmp/arc-bin/arc-bench ./cmd/arc-bench
```

## 5. Prueba pequeña

El supervisor actual utiliza herramientas de medición de macOS. Este comando ejecuta 100 clientes con la topología local validada y un directorio nuevo de resultados:

```bash
/tmp/arc-bin/arc-bench -binary /tmp/arc-bin/arc-go \
  -out "/tmp/arc-pequena-$(date +%Y%m%d-%H%M%S)" -n 100 \
  -rate 20000 -position-window 128 -generators 2 \
  -server-writers 3 -server-egress-sockets 3 \
  -generator-high-burst 24 -generator-high-capacity 2048 -repetitions 1
```

La admisión experimental permanece desactivada. El supervisor reparte rangos y presupuestos sin dividir grupos. Para comprobar el módulo: `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...` y `go build ./...`.

## 6. Dónde leer

- [ARQUITECTURA](docs/ARQUITECTURA.md): procesos, ownership, sockets, colas y recursos.
- [PROTOCOLO_Y_FLUJOS](docs/PROTOCOLO_Y_FLUJOS.md): formato binario, mensajes, estados, RTT y ARQ.
- [CODIGO](docs/CODIGO.md): diez fuentes para entender el proyecto y sus relaciones.
- [FASES_Y_RESULTADOS](docs/FASES_Y_RESULTADOS.md): decisiones experimentales, escalabilidad y cuatro gráficas.
- [DISTRIBUIDO](docs/DISTRIBUIDO.md): adaptaciones pendientes para varios PCs.
