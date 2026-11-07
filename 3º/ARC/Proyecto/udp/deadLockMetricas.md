# Resolución del Deadlock en la Recolección de Métricas (UDP)

## 1. El Problema: Bloqueo Indefinido (Deadlock)

En la implementación inicial sobre UDP, el servidor centralizaba el cierre de la simulación mediante un evento asíncrono:

```python
if len(self.metrics) == self.num_clients:
    self.all_metrics_event.set()
```

Y en su ciclo principal ejecutaba:

```python
await self.all_metrics_event.wait()
```

Al ser **UDP un protocolo de transporte no confiable (best-effort)**, los datagramas pueden descartarse por congestión, saturación de buffers o pérdida de enlace. Además, un cliente puede fallar o cerrarse abruptamente. Si tan solo un datagrama `METRICAS` no alcanzaba al servidor, `len(self.metrics)` nunca igualaba `self.num_clients`, provocando un **deadlock permanente**: el proceso del servidor permanecía bloqueado indefinidamente a la espera de un mensaje que nunca llegaría.

---

## 2. ¿Por qué Stop-and-Wait ARQ en lugar de un Híbrido TCP/UDP?

Una alternativa intuitiva pero arquitectónicamente deficiente consistía en abrir un socket TCP únicamente para el envío del mensaje final de métricas. Esta opción fue descartada por los siguientes motivos:

1. **Sobrecarga desproporcionada (Overhead):**
   Establecer una conexión TCP para transmitir un payload JSON de aproximadamente 50 bytes requiere el saludo de tres vías (`SYN`, `SYN-ACK`, `ACK`), intercambio de paquetes de finalización (`FIN`, `ACK`) y gestión de estados de kernel como `TIME_WAIT`. Esto introduce latencia adicional y consumo innecesario de descriptores de archivo.
2. **Violación del modelo de capas y coherencia arquitectónica:**
   El objetivo del desarrollo es operar sobre UDP. En la arquitectura de redes, cuando se utiliza un transporte no confiable por requerimientos de baja latencia o difusión, la confiabilidad de control se resuelve en la **capa de aplicación** mediante protocolos ARQ (Automatic Repeat reQuest), siguiendo el estándar de protocolos consolidados como TFTP, CoAP o QUIC.
3. **Complejidad operativa:**
   Obliga a clientes y servidores a gestionar dos interfaces de red, dos puertos de escucha y dos máquinas de estado concurrentes con diferentes semánticas de fallo.

---

## 3. ¿Por qué Stop-and-Wait ARQ en lugar de un Timeout Defensivo Aislado?

Un simple timeout en el servidor (`asyncio.wait_for`) sin retransmisión resuelve el bloqueo, pero introduce pérdida de datos:

- **Degradación silenciosa:** Si se pierden datagramas en la red, el servidor simplemente se rinde y descarta las métricas de los clientes afectados, distorsionando el cálculo de latencia media global.
- **Garantía activa de entrega:** Stop-and-Wait implementa reintentos del lado emisor (cliente) con confirmaciones explícitas (`METRICAS_ACK`) desde el receptor (servidor). Si un datagrama o su ACK se pierden, el cliente retransmite hasta 5 veces con intervalos de 1.0 s, garantizando la recolección íntegra aún bajo condiciones adversas de red.
- **Idempotencia:** En el servidor, la recepción reiterada del mismo reporte de métricas actualiza el registro y reenvía el `METRICAS_ACK` correspondiente sin alterar los contadores de clientes registrados ni duplicar registros en consola.

---

## 4. Funcionamiento del Timeout de Guarda Dinámico en el Servidor

Para complementar el protocolo ARQ y prevenir bloqueos en caso de fallo catastrófico (un cliente que cae físicamente y no puede retransmitir), se implementó un timeout de guarda en el servidor.

### El problema de un timeout estático desde $t = 0$

Fijar un timeout estático al iniciar el servidor es un antipatrón: la duración total de la simulación depende del número de iteraciones ($I$), el número de clientes ($N$) y la latencia de red. Un timeout fijo corre el riesgo de expirar prematuramente mientras los clientes aún están ejecutando sus iteraciones normales.

### Solución: Desacoplamiento en Dos Fases

El servidor desacopla la fase de simulación de la fase de recolección de métricas:

```text
[Inicio Servidor] ──(Sin timeout: Iteraciones activas)──> [Llega 1ª Métrica] ──(Ventana acotada)──> [Cierre / Timeout]
```

1. **Fase de Simulación (Abierta):**  
   El servidor espera la llegada de la primera métrica mediante un evento dedicado:

   ```python
   await self.first_metric_event.wait()
   ```

   No existe temporizador activo durante esta fase; los clientes pueden demorar el tiempo que sea necesario en completar sus $I$ iteraciones.

2. **Fase de Recolección (Acotada):**  
   Al recibir la primera métrica, se constata que la simulación ha concluido para al menos un nodo. Dado que todos los clientes operan de manera concurrente en la misma simulación, los demás nodos (muy probablemente) se encuentran en su última iteración o enviando ya sus reportes. En este instante se activa la guarda:
   ```python
   await asyncio.wait_for(self.all_metrics_event.wait(), timeout=self.metrics_timeout)
   ```

### Fórmula de Cálculo Dinámico

El valor de `metrics_timeout` se parametriza dinámicamente en función del número total de clientes $N$:

$$\text{metrics\_timeout} = \max\left(5.0,\, 5.0 + 0.05 \times N\right)$$

- **Término base (5.0 s):** Corresponde exactamente a la cota máxima del mecanismo Stop-and-Wait del cliente (5 reintentos $\times$ 1.0 s de espera).
- **Término de escala ($0.05 \times N$):** Compensa el encolamiento en el bucle de eventos (`asyncio`) y la contención en el buffer del socket UDP cuando $N$ crece ($10.0\text{ s}$ para 100 clientes, $30.0\text{ s}$ para 500 clientes).
- **Recuperación elegante:** Si el temporizador vence, el servidor notifica los identificadores de los clientes faltantes, genera el resumen con las métricas efectivamente recibidas y cierra los sockets de manera limpia sin congelar el proceso.
