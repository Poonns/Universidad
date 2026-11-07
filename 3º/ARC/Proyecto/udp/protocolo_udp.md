# 3. Protocolo de Aplicación (Versión UDP)

## 3.1 Estructura del Datagrama (Sin Framing)

A diferencia de la implementación sobre TCP, el protocolo **UDP es orientado a mensajes (*datagram-oriented*)** y preserva los límites de cada paquete en la capa de transporte. Por este motivo, **no requiere una cabecera explícita de encuadre (*framing*) ni prefijo de longitud** (`uint32`).

Cada datagrama transmitido contiene directamente el cuerpo del mensaje estructurado en formato **JSON codificado en UTF-8**:

```text
+-------------------------------------------------------------------+
|                     Cuerpo del Mensaje (JSON)                     |
|                  Payload del Datagrama UDP (UTF-8)                |
+-------------------------------------------------------------------+
```

---

## 3.2 Tipos de Mensajes y Esquemas

Todos los mensajes transmitidos en la simulación comparten la estructura básica en JSON con los campos comunes: `tipo` (identificador de mensaje), `origen` (ID del cliente emisor; `0` reservado para el servidor), `grupo` (identificador del grupo asignado) y campos de datos específicos.

| Tipo | Emisor | Destinatario | Descripción |
| :--- | :--- | :--- | :--- |
| **`REGISTRO`** | Cliente | Servidor | El cliente solicita asignación de identificador (`client_id`) y grupo (`group_id`) enviando un datagrama inicial al servidor. |
| **`REGISTRO_ACK`** | Servidor | Cliente | El servidor confirma la asignación devolviendo el `client_id` y `group_id` asignados. |
| **`INICIO`** | Servidor | Difusión (*Broadcast*) | Orden enviada por el servidor a todos los clientes registrados una vez completado el cupo de $N$ clientes. |
| **`POSICION`** | Cliente $x$ | Servidor $\rightarrow$ Vecinos | Notificación de coordenadas $(x, y, z)$ enviada al servidor, que la retransmite a los $V - 1$ vecinos de grupo. |
| **`ACK`** | Vecino | Servidor $\rightarrow$ Cliente $x$ | Confirmación de recepción de posición devuelta por el vecino hacia el servidor, que la redirige al emisor original. |
| **`METRICAS`** | Cliente | Servidor | Envío de la latencia media calculada por el cliente tras completar sus $I$ iteraciones. |

---

## 3.3 Ciclo de Vida Completo

El ciclo de vida del sistema comprende tres fases principales:
1. **Fase de Registro e Inicio:** Los clientes envían datagramas de `REGISTRO` sin necesidad de establecer conexiones persistentes. Una vez alcanzados los $N$ clientes requeridos, el servidor difunde el mensaje `INICIO`.
2. **Fase 1 (Iteraciones Asíncronas UDP):** Cada cliente envía de forma asíncrona sus coordenadas (`POSICION`), el servidor las enruta hacia los $V-1$ vecinos del grupo, y estos responden con un datagrama `ACK` redirigido al origen. Al reunir los $V-1$ ACKs, el cliente calcula el RTT de la iteración y avanza a la siguiente.
3. **Fase 2 (Recolección y Cierre):** Tras $I$ iteraciones, cada cliente envía su mensaje `METRICAS` y permanece activo respondiendo ACKs de sus vecinos. El servidor consolida las métricas grupales y globales al recibir la totalidad de los reportes.

### Diagrama de Secuencia

```mermaid
sequenceDiagram
    autonumber
    participant C1 as Cliente 1 (Grupo 1)
    participant C2 as Cliente 2 (Grupo 1)
    participant S as Servidor Central (UDP)

    Note over S: Servidor a la escucha en PUERTO UDP

    C1->>S: Datagrama REGISTRO (tipo: "REGISTRO")
    C2->>S: Datagrama REGISTRO (tipo: "REGISTRO")
    
    S-->>C1: REGISTRO_ACK (ID: 1, Grupo: 1)
    S-->>C2: REGISTRO_ACK (ID: 2, Grupo: 1)

    Note over S: Se han registrado los N clientes requeridos

    S->>C1: INICIO (Difusión a Cliente 1)
    S->>C2: INICIO (Difusión a Cliente 2)

    Note over C1, S: === FASE 1: Iteraciones Asíncronas (UDP) ===
    C1->>S: POSICION (origen: 1, iter: 1, coords)
    S->>C2: POSICION (retransmitido a vecino C2)
    C2->>S: ACK (para: 1, iter: 1)
    S->>C1: ACK (entregado a origen C1)
    Note over C1: Recibidos V-1 ACKs -> Calcula RTT -> Pasa a iteración 2

    Note over C1, S: === FASE 2: Recolección y Cierre ===
    Note over C1: Finaliza I iteraciones (sigue activo respondiendo ACKs)
    Note over C2: Finaliza I iteraciones
    C1->>S: METRICAS (latencia_media, total_iteraciones)
    C2->>S: METRICAS (latencia_media, total_iteraciones)
    Note over S: Calcula promedios grupales y global.<br/>Guarda métricas en log y consola
```

---

## 3.4 Diagrama de Flujo General del Sistema

El siguiente diagrama de flujo muestra la lógica completa de la simulación desde el arranque del servidor hasta la finalización del último cliente. Se aprecian las tres fases principales (Registro, Iteraciones y Métricas) y la interacción entre los roles de **servidor**, **cliente activo** (el que publica posición) y **cliente reactivo** (el que responde con ACK).

```mermaid
flowchart TD
    A([Inicio del Sistema]) --> B[Servidor abre socket UDP<br/>en HOST:PUERTO]
    B --> C[Servidor queda a la<br/>escucha de datagramas]

    C --> D{¿Llega un<br/>datagrama?}
    D -- No --> C
    D -- Sí --> E{¿Tipo de<br/>mensaje?}

    %% --- REGISTRO ---
    E -- REGISTRO --> F[Asignar client_id y group_id]
    F --> G[Enviar REGISTRO_ACK<br/>al cliente]
    G --> H{¿N clientes<br/>registrados?}
    H -- No --> C
    H -- Sí --> I[Enviar INICIO por difusión<br/>a todos los clientes]
    I --> C

    %% --- POSICION ---
    E -- POSICION --> J[Identificar grupo del emisor]
    J --> K[Reenviar datagrama a<br/>los V-1 vecinos del grupo]
    K --> C

    %% --- ACK ---
    E -- ACK --> L[Leer campo 'destino'<br/>del mensaje]
    L --> M[Reenviar ACK al<br/>cliente destino original]
    M --> C

    %% --- METRICAS ---
    E -- METRICAS --> N[Almacenar latencia media<br/>del cliente]
    N --> O{¿Métricas de<br/>todos los N<br/>clientes?}
    O -- No --> C
    O -- Sí --> P[Calcular promedios<br/>grupales y global]
    P --> Q([Mostrar Resumen Final<br/>y cerrar servidor])

```

> **Lectura del diagrama:** El servidor opera como un *bucle de eventos* no bloqueante (reactor pattern). Cada datagrama recibido se clasifica por su campo `tipo` y se procesa de forma síncrona. El diagrama refleja fielmente la función [`handle_datagram()`](file:///Users/hectorpons/proyectos/arq_redes_comp/udp/server.py#L72-L152) del servidor.

---

## 3.5 Diagrama de Estados del Cliente

Cada cliente recorre un ciclo de vida bien definido con estados claramente diferenciados. El diagrama de estados inferior modela las transiciones del autómata finito del cliente, incluyendo los flujos de error por timeout.

```mermaid
stateDiagram-v2
    [*] --> Desconectado

    Desconectado --> Registrando : Crear socket UDP<br/>Enviar REGISTRO

    Registrando --> Registrado : Recibir REGISTRO_ACK<br/>(asigna client_id, group_id)
    Registrando --> Error_Timeout : Timeout sin respuesta

    Registrado --> Esperando_Inicio : Esperar señal INICIO<br/>del servidor

    Esperando_Inicio --> Simulando : Recibir mensaje INICIO
    Esperando_Inicio --> Error_Timeout : Timeout sin INICIO

    state Simulando {
        [*] --> Publicar_Posicion

        Publicar_Posicion --> Esperando_ACKs : Enviar POSICION<br/>(registrar t₀)
        
        Esperando_ACKs --> Esperando_ACKs : Recibir ACK parcial<br/>(acks_recibidos < V-1)
        Esperando_ACKs --> Calcular_RTT : Recibir último ACK<br/>(acks_recibidos = V-1)
        Esperando_ACKs --> Error_Timeout : Timeout esperando ACKs
        
        Calcular_RTT --> Publicar_Posicion : RTT registrado<br/>iter < I
        Calcular_RTT --> [*] : iter = I<br/>(todas las iteraciones completadas)
    }

    note right of Simulando
        Durante la simulación el cliente
        también actúa en ROL REACTIVO:
        responde ACK a las POSICIONES
        recibidas de sus vecinos.
    end note

    Simulando --> Enviando_Metricas : Completadas I iteraciones

    Enviando_Metricas --> Gracia : Enviar METRICAS al servidor<br/>(latencia_media_ms)

    Gracia --> Finalizado : Período de gracia (1.5s)<br/>respondiendo ACKs rezagados

    Finalizado --> [*] : Cerrar socket UDP

    Error_Timeout --> [*] : Cerrar socket y salir con error
```

> **Nota sobre el rol dual:** Mientras un cliente está en el estado `Simulando`, opera simultáneamente en dos roles. Como **emisor activo** envía su posición y espera ACKs; como **receptor reactivo** responde con ACK a las posiciones que recibe de vecinos. Esto se implementa mediante el método [`handle_datagram()`](file:///Users/hectorpons/proyectos/arq_redes_comp/udp/client.py#L74-L115) que procesa datagramas de forma concurrente al bucle de iteraciones.

---

## 3.6 Diagrama de Estados del Servidor

El servidor UDP sigue un modelo de estados más simple, ya que actúa como un *router de mensajes* sin estado de sesión complejo (al contrario que TCP, no mantiene conexiones persistentes).

```mermaid
stateDiagram-v2
    [*] --> Inicializando

    Inicializando --> Escuchando : Abrir socket UDP<br/>en HOST:PUERTO

    state Escuchando {
        [*] --> Fase_Registro

        Fase_Registro --> Fase_Registro : Recibir REGISTRO<br/>(clientes < N)
        Fase_Registro --> Fase_Simulacion : N clientes registrados<br/>→ Difundir INICIO

        Fase_Simulacion --> Fase_Simulacion : Recibir POSICION<br/>→ Reenviar a V-1 vecinos
        Fase_Simulacion --> Fase_Simulacion : Recibir ACK<br/>→ Reenviar a destino
        Fase_Simulacion --> Fase_Recoleccion : Recibir primera METRICA

        Fase_Recoleccion --> Fase_Recoleccion : Recibir METRICAS parciales<br/>(también sigue reenviando<br/>POSICION y ACK)
        Fase_Recoleccion --> [*] : Todas las N métricas recibidas
    }

    Escuchando --> Resumen_Final : Todas las métricas recopiladas

    Resumen_Final --> [*] : Mostrar promedios<br/>grupales y global<br/>→ Cerrar socket
```

> **Observación clave:** A diferencia de un servidor TCP, aquí **no existe estado de conexión por cliente**. El servidor identifica a cada cliente únicamente por su dirección `(IP, puerto)` del datagrama recibido, almacenada en el diccionario [`addr_to_id`](file:///Users/hectorpons/proyectos/arq_redes_comp/udp/server.py#L63). Esto simplifica el modelo de estados pero exige manejar posibles datagramas duplicados de registro.

---

## 3.7 Diagrama de Flujo: Procesamiento de Datagramas en el Cliente

Este diagrama detalla la lógica interna del cliente al recibir un datagrama del servidor, mostrando cómo se bifurca el procesamiento según el tipo de mensaje y los dos roles (activo/reactivo).

```mermaid
flowchart TD
    A([Datagrama recibido<br/>del servidor]) --> B{¿JSON<br/>válido?}
    B -- No --> C([Descartar<br/>datagrama])
    B -- Sí --> D{Tipo de<br/>mensaje}

    D -- REGISTRO_ACK --> E[Extraer client_id,<br/>group_id, num_vecinos]
    E --> F[Guardar identidad<br/>del cliente]
    F --> G[Señalizar<br/>registered_event]
    G --> H([Fin del procesamiento])

    D -- INICIO --> I[Imprimir señal<br/>de inicio recibida]
    I --> J[Señalizar<br/>start_event]
    J --> H

    D -- POSICION --> K["ROL REACTIVO"]
    K --> L[Leer origen e<br/>iteración del mensaje]
    L --> M[Construir mensaje ACK<br/>con destino = origen]
    M --> N[Enviar ACK al servidor<br/>via send_message]
    N --> H

    D -- ACK --> O["ROL ACTIVO"]
    O --> P[Encolar ACK en<br/>ack_queue interna]
    P --> Q[El bucle de simulación<br/>consume el ACK]
    Q --> H

```

> **Detalle de implementación:** El cliente utiliza una cola asíncrona [`ack_queue`](file:///Users/hectorpons/proyectos/arq_redes_comp/udp/client.py#L68) para desacoplar la recepción de ACKs (callback no bloqueante en `handle_datagram`) del consumo secuencial en el bucle de iteraciones. Los mensajes de tipo `POSICION` se responden de forma inmediata sin pasar por la cola, garantizando baja latencia en el rol reactivo.

---

## 3.8 Diagrama de Flujo: Iteración de Simulación del Cliente

Este diagrama muestra en detalle lo que ocurre dentro de **cada iteración** del bucle principal del cliente activo, desde la publicación de la posición hasta el cálculo del RTT.

```mermaid
flowchart TD
    A([Inicio de iteración i]) --> B["Registrar timestamp t₀<br/>(time.perf_counter)"]
    B --> C[Construir mensaje POSICION<br/>con coords e iteración i]
    C --> D[Enviar datagrama POSICION<br/>al servidor]
    D --> E[Inicializar contador:<br/>acks_recibidos = 0]

    E --> F{¿acks_recibidos<br/>< V-1?}
    F -- Sí --> G["Esperar ACK de ack_queue<br/>(con timeout)"]
    G --> H{¿ACK corresponde<br/>a iteración i?}
    H -- No --> G
    H -- Sí --> I[acks_recibidos += 1]
    I --> F

    F -- No --> J["Calcular RTT:<br/>RTT = (t₁ - t₀) × 1000 ms"]
    J --> K[Almacenar RTT en<br/>lista de latencias]
    K --> L{¿i < I?}
    L -- Sí --> M([Avanzar a iteración i+1])
    L -- No --> N["Calcular latencia_media =<br/>Σ(latencias) / I"]
    N --> O[Enviar mensaje METRICAS<br/>al servidor]
    O --> P["Período de gracia (1.5s)<br/>respondiendo ACKs vecinos"]
    P --> Q([Simulación finalizada])

    G -. Timeout .-> R([Error: Timeout<br/>esperando ACKs])

```

> **¿Por qué se filtran ACKs por iteración?** Un ACK puede llegar desordenado o pertenecer a una iteración anterior. La condición `ack.datos.iteracion == i` (línea del [bucle de simulación](file:///Users/hectorpons/proyectos/arq_redes_comp/udp/client.py#L162-L165)) garantiza que sólo se contabilizan los ACKs correspondientes a la iteración actual, descartando los obsoletos.
