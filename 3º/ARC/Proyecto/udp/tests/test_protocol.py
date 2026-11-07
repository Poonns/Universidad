import json
import math
import unittest
from support import *
from runtime import percentiles


class CodecTests(unittest.TestCase):
    def test_roundtrip_all_messages(self):
        control = dict(destino=1, token="token", instancia=INSTANCE)
        packets = [message(MSG_REGISTER, SESSION, token="token", generador="g1",
                    num_clientes=4, num_vecinos=2, total_iteraciones=3),
            message(MSG_REGISTER_ACK, SESSION, group=1, cliente_id=1, grupo_id=1,
                    token="token", instancia=INSTANCE, num_clientes=4,
                    num_vecinos=2, total_iteraciones=3),
            message(MSG_POSITION, SESSION, 1, 1, iteracion=1, coords=[1, 2, 3], pendientes=[2]),
            ack(), message(MSG_METRICS, SESSION, 1, 1, latencia_media_ms=2.5,
                            total_iteraciones=3, acks_unicos=3),
            message(MSG_ERROR, SESSION, token="token", mensaje="incompatible")]
        packets += [message(kind, SESSION, group=1, **control)
                    for kind in (MSG_START, MSG_METRICS_ACK, MSG_GROUP_DONE)]
        packets += [message(kind, SESSION, 1, 1, instancia=INSTANCE)
                    for kind in (MSG_START_ACK, MSG_GROUP_DONE_ACK)]
        for kind in GENERATOR_REQUESTS | GENERATOR_RESPONSES:
            body = dict(generador='g1', nonce='nonce', manifiesto='sha', secuencia=1)
            if kind == MSG_GENERATOR_HELLO:
                body.update(n=25, v=25, i=5, local=5)
            else:
                body['instancia'] = INSTANCE
            if kind == MSG_RUN_STATUS_ACK:
                body['completo'] = False
            packets.append(message(kind, SESSION, **body))
        self.assertEqual({p["tipo"] for p in packets}, MESSAGE_TYPES)
        for packet in packets:
            with self.subTest(kind=packet["tipo"]):
                self.assertEqual(decode_message(encode_message(packet)), packet)

    def test_invalid_datagrams(self):
        valid = encode_message(ack())
        invalid = [b"", valid[:-1], b"[]", b"null", b"{}", b"\xff",
            b"{" * 1200, valid + b"garbage", b" " * (MAX_DATAGRAM_SIZE + 1),
            valid.replace(b'"version":1', b'"version":true'),
            valid.replace(b'"origen":2', b'"origen":true'),
            valid.replace(b'"tipo":"ACK"', b'"tipo":"UNKNOWN"'),
            valid.replace(b'"destino":1', b'"destino":0'),
            valid.replace(b'"iteracion":1', b'"iteracion":-1'),
            valid.replace(b'"origen":2', b'"origen":2,"origen":3')]
        for packet in invalid:
            with self.subTest(packet=packet[:100]):
                with self.assertRaises(ProtocolError):
                    decode_message(packet)
        for value in (float("nan"), float("inf"), True, "x"):
            with self.assertRaises(ProtocolError):
                encode_message(message(MSG_POSITION, SESSION, 1, 1, iteracion=1, coords=[0, 1, value]))
        for targets in ([1, 1], [], [True], "2"):
            with self.assertRaises(ProtocolError):
                encode_message(message(MSG_POSITION, SESSION, 1, 1,
                                       iteracion=1, coords=[0, 0, 0], pendientes=targets))

    def test_configuration_and_identifiers(self):
        for values in ((0, 1, 1), (2, 0, 1), (3, 2, 1), (4, 2, 0), (True, 1, 1)):
            with self.assertRaises(ProtocolError):
                validate_configuration(*values)
        for session in ("", "../archivo", "a" * 65):
            with self.assertRaises(ProtocolError):
                identifier(session, "session")
        for v in (1, 25, 32):
            validate_configuration(v * 2, v, 5)
            self.assertEqual(len(members(2, v)), v)
            self.assertEqual(group_for(v + 1, v), 2)

    def test_percentiles_threshold_and_nearest_rank(self):
        self.assertIsNone(percentiles([])["mean_ms"])
        self.assertIsNone(percentiles(list(range(19)))["p95_ms"])
        self.assertEqual(percentiles(list(range(1, 21)))["p95_ms"], 19)
        self.assertIsNone(percentiles(list(range(99)))["p99_ms"])
        result = percentiles(list(range(1, 101)))
        self.assertEqual(result["mean_ms"], 50.5)
        self.assertEqual(result["p95_ms"], 95)
        self.assertEqual(result["p99_ms"], 99)

    def test_retry_policy_and_limiter_validation(self):
        from runtime import RateLimiter
        for args in ((0, 1, 3, 0), (1, .5, 3, 0), (1, 2, 0, 0), (1, 2, 3, 1)):
            with self.assertRaises(ValueError):
                RetryPolicy(*args)
        for rate in (0, -1, math.inf, math.nan):
            with self.assertRaises(ValueError):
                RateLimiter(rate)
