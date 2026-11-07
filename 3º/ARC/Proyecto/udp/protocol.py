<<<<<<< HEAD
"""Contrato UDP JSON v1. Cada datagrama es exactamente un mensaje, sin framing TCP."""
from __future__ import annotations

import json
import math
import re
from typing import Any, Dict, Optional, Tuple

MSG_REGISTER = "REGISTRO"
MSG_REGISTER_ACK = "REGISTRO_ACK"
MSG_START = "INICIO"
MSG_START_ACK = "INICIO_ACK"
=======
import asyncio
import json
from typing import Any, Dict, Optional, Tuple

# Message Type Constants (matching DESIGN.md specification)
MSG_REGISTER = "REGISTRO"
MSG_REGISTER_ACK = "REGISTRO_ACK"
MSG_START = "INICIO"
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
MSG_POSITION = "POSICION"
MSG_ACK = "ACK"
MSG_METRICS = "METRICAS"
MSG_METRICS_ACK = "METRICAS_ACK"
<<<<<<< HEAD
MSG_GROUP_DONE = "GRUPO_FIN"
MSG_GROUP_DONE_ACK = "GRUPO_FIN_ACK"
MSG_ERROR = "ERROR"
MSG_GENERATOR_HELLO = "GENERADOR_HOLA"
MSG_GENERATOR_HELLO_ACK = "GENERADOR_HOLA_ACK"
MSG_RUN_STATUS = "ESTADO_RUN"
MSG_RUN_STATUS_ACK = "ESTADO_RUN_ACK"
MSG_GENERATOR_DONE = "GENERADOR_FIN"
MSG_GENERATOR_DONE_ACK = "GENERADOR_FIN_ACK"
GENERATOR_REQUESTS = frozenset((MSG_GENERATOR_HELLO, MSG_RUN_STATUS, MSG_GENERATOR_DONE))
GENERATOR_RESPONSES = frozenset((MSG_GENERATOR_HELLO_ACK, MSG_RUN_STATUS_ACK, MSG_GENERATOR_DONE_ACK))
VERSION = 1
MAX_DATAGRAM_SIZE = 8192
MESSAGE_TYPES = frozenset((MSG_REGISTER, MSG_REGISTER_ACK, MSG_START, MSG_START_ACK,
    MSG_POSITION, MSG_ACK, MSG_METRICS, MSG_METRICS_ACK, MSG_GROUP_DONE,
    MSG_GROUP_DONE_ACK, MSG_ERROR)) | GENERATOR_REQUESTS | GENERATOR_RESPONSES
IDENTIFIER = re.compile(r"[A-Za-z0-9_-]{1,64}\Z")
Address = Tuple[str, int]
Message = Dict[str, Any]


class ProtocolError(ValueError):
    """Datagrama o combinación de parámetros inválidos."""


def identifier(value: Any, name: str) -> str:
    if not isinstance(value, str) or IDENTIFIER.fullmatch(value) is None:
        raise ProtocolError(f"{name}: se requiere identificador de 1..64 caracteres ASCII")
    return value


def integer(value: Any, name: str, minimum: int = 0) -> int:
    if type(value) is not int or value < minimum:
        raise ProtocolError(f"{name}: se requiere entero >= {minimum}")
    return value


def finite(value: Any, name: str) -> float:
    if type(value) not in (int, float):
        raise ProtocolError(f"{name}: se requiere número finito")
    try:
        result = float(value)
    except (OverflowError, ValueError) as exc:
        raise ProtocolError(f"{name}: número no representable") from exc
    if not math.isfinite(result):
        raise ProtocolError(f"{name}: se requiere número finito")
    return result


def validate_configuration(n: int, v: int, iterations: int) -> None:
    integer(n, "N", 1)
    integer(v, "V", 1)
    integer(iterations, "I", 1)
    if n % v:
        raise ProtocolError("N debe ser divisible por V")


def group_for(client_id: int, v: int) -> int:
    return (client_id - 1) // v + 1


def members(group: int, v: int) -> set[int]:
    first = (group - 1) * v + 1
    return set(range(first, first + v))


def message(kind: str, session: str, origin: int = 0, group: int = 0,
            **body: Any) -> Message:
    return {"version": VERSION, "sesion": session, "tipo": kind,
            "origen": origin, "grupo": group, "datos": body}


def validate_message(msg: Any) -> Message:
    if not isinstance(msg, dict):
        raise ProtocolError("el mensaje debe ser un objeto JSON")
    if type(msg.get("version")) is not int or msg["version"] != VERSION:
        raise ProtocolError("versión no soportada")
    identifier(msg.get("sesion"), "sesion")
    kind = msg.get("tipo")
    if not isinstance(kind, str) or kind not in MESSAGE_TYPES:
        raise ProtocolError("tipo de mensaje desconocido")
    integer(msg.get("origen"), "origen")
    integer(msg.get("grupo"), "grupo")
    body = msg.get("datos")
    if not isinstance(body, dict):
        raise ProtocolError("datos debe ser un objeto")
    if kind in GENERATOR_REQUESTS | GENERATOR_RESPONSES:
        if msg['origen'] or msg['grupo']:
            raise ProtocolError('control de generador exige origen/grupo cero')
        for key in ('generador', 'nonce', 'manifiesto'):
            identifier(body.get(key), key)
        integer(body.get('secuencia'), 'secuencia', 1)
        if kind == MSG_GENERATOR_HELLO:
            validate_configuration(body.get('n'), body.get('v'), body.get('i'))
            integer(body.get('local'), 'N local', 1)
        else:
            identifier(body.get('instancia'), 'instancia')
        if kind == MSG_RUN_STATUS_ACK:
            if type(body.get('completo')) is not bool:
                raise ProtocolError('completo debe ser booleano')
            if body['completo'] and not isinstance(body.get('certificado'), dict):
                raise ProtocolError('falta certificado remoto')
    elif kind == MSG_REGISTER:
        identifier(body.get("token"), "token")
        identifier(body.get("generador"), "generador")
        validate_configuration(body.get("num_clientes"), body.get("num_vecinos"),
                               body.get("total_iteraciones"))
    elif kind == MSG_REGISTER_ACK:
        identifier(body.get("token"), "token")
        identifier(body.get("instancia"), "instancia")
        integer(body.get("cliente_id"), "cliente_id", 1)
        integer(body.get("grupo_id"), "grupo_id", 1)
        validate_configuration(body.get("num_clientes"), body.get("num_vecinos"),
                               body.get("total_iteraciones"))
    elif kind in (MSG_START, MSG_METRICS_ACK, MSG_GROUP_DONE):
        integer(body.get("destino"), "destino", 1)
        identifier(body.get("token"), "token")
        identifier(body.get("instancia"), "instancia")
    elif kind in (MSG_START_ACK, MSG_GROUP_DONE_ACK):
        identifier(body.get("instancia"), "instancia")
    elif kind == MSG_POSITION:
        integer(body.get("iteracion"), "iteracion", 1)
        coords = body.get("coords")
        if not isinstance(coords, list) or len(coords) != 3:
            raise ProtocolError("coords debe contener tres coordenadas")
        for value in coords:
            finite(value, "coordenada")
        if "pendientes" in body:
            targets = body["pendientes"]
            if not isinstance(targets, list) or not targets:
                raise ProtocolError("pendientes debe ser una lista no vacía")
            for value in targets:
                integer(value, "vecino pendiente", 1)
            if len(set(targets)) != len(targets):
                raise ProtocolError("vecinos pendientes duplicados")
    elif kind == MSG_ACK:
        integer(body.get("destino"), "destino", 1)
        integer(body.get("iteracion"), "iteracion", 1)
    elif kind == MSG_METRICS:
        if finite(body.get("latencia_media_ms"), "latencia_media_ms") < 0:
            raise ProtocolError("latencia negativa")
        integer(body.get("total_iteraciones"), "total_iteraciones", 1)
        integer(body.get("acks_unicos"), "acks_unicos")
    elif kind == MSG_ERROR:
        identifier(body.get("token"), "token")
        if not isinstance(body.get("mensaje"), str):
            raise ProtocolError("mensaje de error inválido")
    return msg


def _object(pairs: list) -> dict:
    result = {}
    for key, value in pairs:
        if key in result:
            raise ProtocolError(f"clave JSON duplicada: {key}")
        result[key] = value
    return result


def _constant(value: str) -> None:
    raise ProtocolError(f"constante JSON inválida: {value}")


def encode_message(msg: Message) -> bytes:
    validate_message(msg)
    try:
        data = json.dumps(msg, separators=(",", ":"), allow_nan=False).encode("utf-8")
    except (ValueError, TypeError, RecursionError) as exc:
        raise ProtocolError("JSON no codificable") from exc
    if len(data) > MAX_DATAGRAM_SIZE:
        raise ProtocolError("datagrama demasiado grande")
    return data


def decode_message(data: bytes) -> Message:
    if not data or len(data) > MAX_DATAGRAM_SIZE:
        raise ProtocolError("longitud de datagrama inválida")
    try:
        return validate_message(json.loads(data.decode("utf-8"), object_pairs_hook=_object,
                                           parse_constant=_constant))
    except (UnicodeError, ValueError, TypeError, RecursionError) as exc:
        raise ProtocolError(str(exc)) from exc


def send_message(transport: Any, msg: Message, addr: Optional[Address] = None) -> None:
    """sendto acepta el envío local; no acredita entrega remota."""
    data = encode_message(msg)
    if addr is None:
        transport.sendto(data)
    else:
        transport.sendto(data, addr)
=======


def encode_message(message: Dict[str, Any]) -> bytes:
    return json.dumps(message).encode("utf-8")


def decode_message(data: bytes) -> Dict[str, Any]:
    return json.loads(data.decode("utf-8"))


def send_message(
    transport: asyncio.DatagramTransport, #creado en connection_made
    message: Dict[str, Any],
    addr: Optional[Tuple[str, int]] = None,
) -> None:
    data = encode_message(message)
    if addr is not None:
        transport.sendto(data, addr)
    else:
        transport.sendto(data)
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
