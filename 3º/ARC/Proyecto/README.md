# Arquitectura de Redes de Computadores — simulación RAC

El repositorio conserva implementaciones distintas. El trabajo actual se centra en **Python UDP en `udp/`**, con tres documentos principales:

- [Arquitectura y ejecución UDP](udp/ARQUITECTURA.md): instalación, servidor free-threaded, generadores, Docker y estrés.
- [Protocolo UDP](udp/PROTOCOLO.md): registro, datos, métricas confirmadas, retransmisiones y cierre remoto.
- [Errores corregidos y resultados UDP](udp/CAMBIOS_UDP.md): validación, comparación W6/W8, recursos, límites y respaldo de evidencia.
- [TCP original: diseño](DESIGN.md), [registro](REGISTRO.md), [posición/ACK](POSICION_Y_ACKS.md) y [métricas](METRICAS.md). Corresponden a los fuentes Python de la raíz.
- TCP optimizado y UDP Go anterior: referencias históricas que no están incluidas actualmente en esta copia del proyecto.
- [Go multiplexado](ARC_GO_MULTIPLEXADO/README.md): etapa local independiente, validada en su propia configuración.

Los resultados de una variante no acreditan capacidad de otra. Para ejecutar o modificar Python UDP, comenzar por ARQUITECTURA; sus decisiones y evidencias están dentro de `udp/`.
