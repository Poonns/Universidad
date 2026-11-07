package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/bits"
	"sync"
	"testing"
)

func fixtureManifest(t testing.TB, v uint32) *Manifest {
	t.Helper()
	c := RunConfig{RunID: 42, N: 2 * v, V: v, I: 5}
	m, err := NewManifest(c, []GeneratorConfig{{7, 1, v, c}, {9, v + 1, v, c}})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func fixturePacket(typ MessageType) Packet {
	p := Packet{Header: Header{Type: typ, RunID: 42, GeneratorID: 7, GroupID: 1, OriginID: 1, MessageID: 99}}
	switch typ {
	case REGISTER:
		p.GroupID, p.OriginID, p.RegistrationNonce = 0, 0, 1
	case REGISTER_ACK:
		p.Flags, p.OriginID, p.TargetID, p.RegistrationNonce = ServerToClient, 0, 1, 1
		p.Registration = RegistrationPayload{25, 5}
	case START, METRICS_ACK, GROUP_DONE:
		p.Flags, p.OriginID, p.TargetID = ServerToClient, 0, 1
	case POSITION, RETRY_POSITION:
		p.Iteration = 1
		p.Position = PositionPayload{1, -2, 0.5}
		if typ == RETRY_POSITION {
			p.MissingMask, p.RetryCounter = 2, 1
		}
	case ACK:
		p.OriginID, p.TargetID, p.Iteration = 2, 1, 1
	case METRICS:
		p.Metrics = MetricsPayload{123456789, 5}
	case ERROR:
		p.Flags, p.OriginID, p.TargetID = ServerToClient, 0, 1
		p.ErrorCode = OperationTimeout
	}
	return p
}

func wire(t testing.TB, p Packet) []byte {
	t.Helper()
	var buf [MaxPacketSize]byte
	n, err := Encode(buf[:], &p)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Clone(buf[:n])
}

func TestGoldenPosition(t *testing.T) {
	p := fixturePacket(POSITION)
	p.RunID = 0x0102030405060708
	p.MessageID = 0x1112131415161718
	// Independent, explicit wire vector: all 64 header bytes, then IEEE-754 xyz.
	want, err := hex.DecodeString(
		"41524347" + "01" + "05" + "0000" + "0102030405060708" +
			"00000007" + "00000001" + "00000001" + "00000000" +
			"00000001" + "00000000" + "1112131415161718" +
			"0000000000000000" + "0000" + "000c" + "00000000" +
			"3f800000" + "c0000000" + "3f000000")
	if err != nil {
		t.Fatal(err)
	}
	if got := wire(t, p); !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
	got, err := Decode(want)
	if err != nil || got != p {
		t.Fatalf("decoded %+v, %v", got, err)
	}
}

func TestGoldenACK(t *testing.T) {
	p := fixturePacket(ACK)
	p.Flags = ServerToClient
	p.RetryCounter = 3
	want, err := hex.DecodeString(
		"41524347" + "01" + "07" + "0001" + "000000000000002a" +
			"00000007" + "00000001" + "00000002" + "00000001" +
			"00000001" + "00000000" + "0000000000000063" +
			"0000000000000000" + "0003" + "0000" + "00000000")
	if err != nil {
		t.Fatal(err)
	}
	if got := wire(t, p); !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
	got, err := Decode(want)
	if err != nil || got != p {
		t.Fatalf("decoded %+v, %v", got, err)
	}
}

func TestAllTypesRoundTrip(t *testing.T) {
	m := fixtureManifest(t, 25)
	wantLengths := [...]int{0, 64, 72, 64, 64, 76, 76, 64, 80, 64, 64, 64, 68}
	for typ := REGISTER; typ <= ERROR; typ++ {
		p := fixturePacket(typ)
		check := func(p Packet) {
			t.Helper()
			buf := wire(t, p)
			if len(buf) != wantLengths[typ] {
				t.Fatalf("type %d length %d", typ, len(buf))
			}
			got, err := Decode(buf)
			if err != nil || got != p {
				t.Fatalf("type %d roundtrip: %+v %v", typ, got, err)
			}
			if err := m.Validate(&got); err != nil {
				t.Fatalf("type %d manifest: %v", typ, err)
			}
			clear(buf) // Packet owns values; it cannot alias the receive buffer.
			if got != p {
				t.Fatalf("type %d aliases receive buffer", typ)
			}
		}
		check(p)
		if typ == POSITION || typ == RETRY_POSITION || typ == ACK {
			p.Flags = ServerToClient
			if typ != ACK {
				p.TargetID = 2
			}
			check(p)
		}
	}
	// ERROR before assignment echoes a nonce, even when that nonce was rejected.
	p := fixturePacket(ERROR)
	p.TargetID, p.GroupID, p.RegistrationNonce, p.ErrorCode = 0, 0, 26, InvalidRegistration
	d, err := Decode(wire(t, p))
	if err != nil || d != p {
		t.Fatalf("registration error: %+v %v", d, err)
	}
	if err := m.Validate(&p); err != nil {
		t.Fatal(err)
	}
}

func TestTruncationAndTrailingBytes(t *testing.T) {
	for typ := REGISTER; typ <= ERROR; typ++ {
		buf := wire(t, fixturePacket(typ))
		for n := 0; n < len(buf); n++ {
			got, err := Decode(buf[:n])
			if err == nil || got != (Packet{}) {
				t.Fatalf("type %d accepted prefix %d", typ, n)
			}
		}
		if _, err := Decode(append(buf, 0)); err != ErrLength {
			t.Fatalf("type %d trailing: %v", typ, err)
		}
	}
	if _, err := Decode(make([]byte, 65535)); err != ErrLength {
		t.Fatal(err)
	}
}

func TestInvalidWire(t *testing.T) {
	type mutation struct {
		name   string
		typ    MessageType
		offset int
		data   []byte
		want   error
	}
	u16 := func(n uint16) []byte { return binary.BigEndian.AppendUint16(nil, n) }
	u32 := func(n uint32) []byte { return binary.BigEndian.AppendUint32(nil, n) }
	u64 := func(n uint64) []byte { return binary.BigEndian.AppendUint64(nil, n) }
	cases := []mutation{
		{"magic", POSITION, 0, u32(0), ErrMagic},
		{"version", POSITION, 4, []byte{2}, ErrVersion},
		{"type-zero", ACK, 5, []byte{0}, ErrType},
		{"type-unknown", ACK, 5, []byte{255}, ErrType},
		{"flags", ACK, 6, u16(2), ErrFlags},
		{"run-zero", ACK, 8, u64(0), ErrRun},
		{"generator-zero", ACK, 16, u32(0), ErrIdentity},
		{"group-zero", ACK, 20, u32(0), ErrIdentity},
		{"origin-zero", ACK, 24, u32(0), ErrIdentity},
		{"target-zero", ACK, 28, u32(0), ErrIdentity},
		{"self-ack", ACK, 24, u32(1), ErrIdentity},
		{"iteration-zero", ACK, 32, u32(0), ErrIteration},
		{"control-iteration", START, 32, u32(1), ErrIteration},
		{"mask-on-ack", ACK, 36, u32(2), ErrMask},
		{"message-zero", ACK, 40, u64(0), ErrMessage},
		{"nonce-on-ack", ACK, 48, u64(1), ErrNonce},
		{"length", POSITION, 58, u16(0), ErrLength},
		{"length-max", ACK, 58, u16(65535), ErrLength},
		{"reserved", ACK, 60, u32(1), ErrReserved},
		{"nan-x", POSITION, 64, u32(0x7fc00000), ErrPayload},
		{"infinity-y", POSITION, 68, u32(0x7f800000), ErrPayload},
		{"negative-infinity-z", POSITION, 72, u32(0xff800000), ErrPayload},
		{"initial-retry", POSITION, 56, u16(1), ErrRetry},
		{"retry-zero", RETRY_POSITION, 56, u16(0), ErrRetry},
		{"retry-empty-mask", RETRY_POSITION, 36, u32(0), ErrMask},
		{"forward-missing-target", POSITION, 6, u16(ServerToClient), ErrIdentity},
		{"client-position-target", POSITION, 28, u32(2), ErrIdentity},
		{"register-direction", REGISTER, 6, u16(ServerToClient), ErrFlags},
		{"register-group", REGISTER, 20, u32(1), ErrIdentity},
		{"register-origin", REGISTER, 24, u32(1), ErrIdentity},
		{"register-target", REGISTER, 28, u32(1), ErrIdentity},
		{"register-nonce-zero", REGISTER, 48, u64(0), ErrNonce},
		{"register-ack-direction", REGISTER_ACK, 6, u16(0), ErrFlags},
		{"register-ack-origin", REGISTER_ACK, 24, u32(1), ErrIdentity},
		{"register-ack-target", REGISTER_ACK, 28, u32(0), ErrIdentity},
		{"register-ack-nonce", REGISTER_ACK, 48, u64(0), ErrNonce},
		{"register-ack-v-zero", REGISTER_ACK, 64, u32(0), ErrPayload},
		{"register-ack-v-large", REGISTER_ACK, 64, u32(33), ErrPayload},
		{"register-ack-i-zero", REGISTER_ACK, 68, u32(0), ErrPayload},
		{"start-direction", START, 6, u16(0), ErrFlags},
		{"start-origin", START, 24, u32(1), ErrIdentity},
		{"start-target", START, 28, u32(0), ErrIdentity},
		{"start-ack-direction", START_ACK, 6, u16(ServerToClient), ErrFlags},
		{"start-ack-origin", START_ACK, 24, u32(0), ErrIdentity},
		{"start-ack-target", START_ACK, 28, u32(1), ErrIdentity},
		{"metrics-zero-iterations", METRICS, 72, u32(0), ErrPayload},
		{"metrics-reserved", METRICS, 76, u32(1), ErrReserved},
		{"error-direction", ERROR, 6, u16(0), ErrFlags},
		{"error-origin", ERROR, 24, u32(1), ErrIdentity},
		{"error-group", ERROR, 20, u32(0), ErrIdentity},
		{"error-code-zero", ERROR, 64, u32(0), ErrPayload},
		{"error-code-unknown", ERROR, 64, u32(5), ErrPayload},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := wire(t, fixturePacket(tc.typ))
			copy(buf[tc.offset:], tc.data)
			got, err := Decode(buf)
			if err != tc.want || got != (Packet{}) {
				t.Fatalf("got %+v, %v; want %v", got, err, tc.want)
			}
		})
	}
}

func TestEncodeFailuresLeaveBufferUntouched(t *testing.T) {
	p := fixturePacket(POSITION)
	for n := 0; n < 76; n++ {
		buf := bytes.Repeat([]byte{0xa5}, MaxPacketSize)
		if size, err := Encode(buf[:n], &p); size != 0 || err != ErrBuffer {
			t.Fatalf("size %d: %d %v", n, size, err)
		}
		if !bytes.Equal(buf, bytes.Repeat([]byte{0xa5}, MaxPacketSize)) {
			t.Fatal("changed buffer on failure")
		}
	}
	for _, tc := range []struct {
		name   string
		change func(*Packet)
		want   error
	}{
		{"inactive-position", func(p *Packet) { p.Position.X = 1 }, ErrPayload},
		{"inactive-negative-zero", func(p *Packet) { p.Position.X = math.Float32frombits(0x80000000) }, ErrPayload},
		{"inactive-registration", func(p *Packet) { p.Registration.V = 25 }, ErrPayload},
		{"inactive-metrics", func(p *Packet) { p.Metrics.MeanRTTNS = 1 }, ErrPayload},
		{"inactive-error", func(p *Packet) { p.ErrorCode = InternalError }, ErrPayload},
		{"type", func(p *Packet) { p.Type = 0 }, ErrType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixturePacket(ACK)
			tc.change(&p)
			buf := bytes.Repeat([]byte{0xa5}, MaxPacketSize)
			before := bytes.Clone(buf)
			if n, err := Encode(buf, &p); n != 0 || err != tc.want {
				t.Fatalf("%d %v", n, err)
			}
			if !bytes.Equal(buf, before) {
				t.Fatal("changed buffer on failure")
			}
		})
	}
	if _, err := Encode(nil, nil); err != ErrPayload {
		t.Fatal(err)
	}
	var buf [MaxPacketSize]byte
	for i := range buf {
		buf[i] = 0xff
	}
	p = fixturePacket(ACK)
	n, err := Encode(buf[:], &p)
	if err != nil {
		t.Fatal(err)
	}
	if buf[n] != 0xff {
		t.Fatal("wrote beyond datagram")
	}
	if _, err := Decode(buf[:n]); err != nil {
		t.Fatal(err)
	}
}

func TestManifestExample160K(t *testing.T) {
	c := RunConfig{RunID: 123, N: 160000, V: 25, I: 5}
	g := []GeneratorConfig{{4, 120001, 40000, c}, {2, 40001, 40000, c}, {1, 1, 40000, c}, {3, 80001, 40000, c}}
	m, err := NewManifest(c, g)
	if err != nil {
		t.Fatal(err)
	}
	// Caller changes cannot invalidate the immutable validated snapshot.
	g[0].Count = 1
	for gen := uint32(1); gen <= 4; gen++ {
		for _, nonce := range []uint64{1, 40000} {
			id, group, err := m.RegistrationIdentity(gen, nonce)
			want := (gen-1)*40000 + uint32(nonce)
			if err != nil || id != want || group != 1+(want-1)/25 {
				t.Fatalf("%d/%d -> %d/%d %v", gen, nonce, id, group, err)
			}
			id2, group2, err2 := m.RegistrationIdentity(gen, nonce)
			if id2 != id || group2 != group || err2 != nil {
				t.Fatal("unstable assignment")
			}
		}
	}
	if m.Config() != c {
		t.Fatal("config mismatch")
	}
}

func TestInvalidManifests(t *testing.T) {
	base := RunConfig{42, 50, 25, 5}
	for _, c := range []RunConfig{{0, 50, 25, 5}, {42, 0, 25, 5}, {42, 50, 0, 5}, {42, 66, 33, 5}, {42, 50, 25, 0}, {42, 51, 25, 5}} {
		if _, err := NewManifest(c, nil); err != ErrConfig {
			t.Fatalf("%+v: %v", c, err)
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func([]GeneratorConfig) []GeneratorConfig
		want   error
	}{
		{"empty", func(g []GeneratorConfig) []GeneratorConfig { return nil }, ErrManifest},
		{"zero-id", func(g []GeneratorConfig) []GeneratorConfig { g[0].GeneratorID = 0; return g }, ErrManifest},
		{"duplicate-id", func(g []GeneratorConfig) []GeneratorConfig { g[1].GeneratorID = 7; return g }, ErrManifest},
		{"zero-first", func(g []GeneratorConfig) []GeneratorConfig { g[0].FirstID = 0; return g }, ErrManifest},
		{"zero-count", func(g []GeneratorConfig) []GeneratorConfig { g[0].Count = 0; return g }, ErrManifest},
		{"unaligned-first", func(g []GeneratorConfig) []GeneratorConfig { g[1].FirstID = 25; return g }, ErrManifest},
		{"split-group", func(g []GeneratorConfig) []GeneratorConfig {
			g[0].Count = 24
			g[1].FirstID = 25
			g[1].Count = 26
			return g
		}, ErrManifest},
		{"overlap", func(g []GeneratorConfig) []GeneratorConfig { g[1].FirstID = 1; return g }, ErrManifest},
		{"gap", func(g []GeneratorConfig) []GeneratorConfig { g[1].FirstID = 51; return g }, ErrManifest},
		{"missing-last-range", func(g []GeneratorConfig) []GeneratorConfig { return g[:1] }, ErrManifest},
		{"past-N", func(g []GeneratorConfig) []GeneratorConfig { g[1].Count = 50; return g }, ErrManifest},
		{"run", func(g []GeneratorConfig) []GeneratorConfig { g[1].Run.RunID++; return g }, ErrRun},
		{"N", func(g []GeneratorConfig) []GeneratorConfig { g[1].Run.N++; return g }, ErrRun},
		{"V", func(g []GeneratorConfig) []GeneratorConfig { g[1].Run.V++; return g }, ErrRun},
		{"I", func(g []GeneratorConfig) []GeneratorConfig { g[1].Run.I++; return g }, ErrRun},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.mutate([]GeneratorConfig{{7, 1, 25, base}, {9, 26, 25, base}})
			if _, err := NewManifest(base, g); err != tc.want {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	// uint32 maximum: the exclusive end must not wrap to zero.
	c := RunConfig{42, math.MaxUint32, 1, 5}
	m, err := NewManifest(c, []GeneratorConfig{{1, 1, math.MaxUint32, c}})
	if err != nil {
		t.Fatal(err)
	}
	id, group, err := m.RegistrationIdentity(1, math.MaxUint32)
	if err != nil || id != math.MaxUint32 || group != math.MaxUint32 {
		t.Fatalf("%d/%d %v", id, group, err)
	}
	if _, err := NewManifest(c, []GeneratorConfig{{1, 1, math.MaxUint32 - 1, c}, {2, math.MaxUint32, 2, c}}); err != ErrManifest {
		t.Fatal(err)
	}
}

func TestManifestPacketRejection(t *testing.T) {
	m := fixtureManifest(t, 25)
	for _, tc := range []struct {
		name   string
		typ    MessageType
		change func(*Packet)
		want   error
	}{
		{"wrong-run", POSITION, func(p *Packet) { p.RunID++ }, ErrRun},
		{"unknown-generator", POSITION, func(p *Packet) { p.GeneratorID = 8 }, ErrIdentity},
		{"outside-N", POSITION, func(p *Packet) { p.OriginID = 51 }, ErrIdentity},
		{"wrong-generator-range", POSITION, func(p *Packet) { p.OriginID = 26; p.GroupID = 2 }, ErrIdentity},
		{"wrong-group", POSITION, func(p *Packet) { p.GroupID = 2 }, ErrIdentity},
		{"cross-group-ack", ACK, func(p *Packet) { p.OriginID = 26 }, ErrIdentity},
		{"iteration-too-large", POSITION, func(p *Packet) { p.Iteration = 6 }, ErrIteration},
		{"nonce-too-large", REGISTER, func(p *Packet) { p.RegistrationNonce = 26 }, ErrNonce},
		{"nonce-overflow", REGISTER, func(p *Packet) { p.RegistrationNonce = math.MaxUint64 }, ErrNonce},
		{"assignment-mismatch", REGISTER_ACK, func(p *Packet) { p.TargetID = 2 }, ErrIdentity},
		{"wrong-V", REGISTER_ACK, func(p *Packet) { p.Registration.V = 24 }, ErrPayload},
		{"wrong-I", REGISTER_ACK, func(p *Packet) { p.Registration.I = 4 }, ErrPayload},
		{"self-mask", RETRY_POSITION, func(p *Packet) { p.MissingMask = 1 }, ErrMask},
		{"outside-V-mask", RETRY_POSITION, func(p *Packet) { p.MissingMask = 1 << 25 }, ErrMask},
		{"unrequested-target", RETRY_POSITION, func(p *Packet) { p.Flags = ServerToClient; p.TargetID = 3 }, ErrMask},
		{"metrics-incomplete", METRICS, func(p *Packet) { p.Metrics.CompletedIterations = 4 }, ErrIteration},
		{"metrics-excess", METRICS, func(p *Packet) { p.Metrics.CompletedIterations = 6 }, ErrIteration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixturePacket(tc.typ)
			tc.change(&p)
			// Structurally valid bytes can still be invalid for this run.
			d, err := Decode(wire(t, p))
			if err != nil {
				t.Fatal(err)
			}
			if err := m.Validate(&d); err != tc.want {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	for _, nonce := range []uint64{0, 26, math.MaxUint64} {
		if _, _, err := m.RegistrationIdentity(7, nonce); err != ErrNonce {
			t.Fatal(err)
		}
	}
	if _, _, err := m.RegistrationIdentity(8, 1); err != ErrIdentity {
		t.Fatal(err)
	}
	var nilManifest *Manifest
	p := fixturePacket(POSITION)
	if err := nilManifest.Validate(&p); err != ErrRun {
		t.Fatal(err)
	}
	if _, _, err := nilManifest.RegistrationIdentity(1, 1); err != ErrIdentity {
		t.Fatal(err)
	}
}

func TestMasksAndACKsV1V25V32(t *testing.T) {
	for _, v := range []uint32{1, 25, 32} {
		m := fixtureManifest(t, v)
		groupMask, err := GroupMask(v)
		if err != nil || bits.OnesCount32(groupMask) != int(v) {
			t.Fatalf("V=%d mask %08x %v", v, groupMask, err)
		}
		for slot := uint32(0); slot < v; slot++ {
			origin := slot + 1
			expected, err := ExpectedMask(origin, v)
			if err != nil || bits.OnesCount32(expected) != int(v)-1 || expected&(1<<slot) != 0 {
				t.Fatalf("V=%d expected %08x %v", v, expected, err)
			}
			op := Operation{42, 99, origin, 1}
			var received uint32
			// Reverse order, including slot 31 for V=32; real IDs, not just bits.
			for neighbor := v; neighbor > 0; neighbor-- {
				if neighbor == origin {
					continue
				}
				p := fixturePacket(ACK)
				p.Flags = ServerToClient
				p.OriginID = neighbor
				p.TargetID = origin
				before := received
				var duplicate bool
				received, duplicate, err = m.ApplyACK(&p, op, received)
				if err != nil || duplicate || bits.OnesCount32(received) != bits.OnesCount32(before)+1 {
					t.Fatalf("ACK %d: %08x %v %v", neighbor, received, duplicate, err)
				}
				again, duplicate, err := m.ApplyACK(&p, op, received)
				if err != nil || !duplicate || again != received {
					t.Fatal("duplicate created progress")
				}
				missing, err := MissingMask(expected, received)
				if err != nil || missing&received != 0 || (missing|received) != expected {
					t.Fatal("invalid missing mask")
				}
				if missing != 0 {
					retry := fixturePacket(RETRY_POSITION)
					retry.OriginID = origin
					retry.MissingMask = missing
					if err := m.Validate(&retry); err != nil {
						t.Fatalf("selective retry: %v", err)
					}
				}
			}
			missing, err := MissingMask(expected, received)
			if err != nil || missing != 0 || received != expected {
				t.Fatalf("V=%d incomplete %08x", v, received)
			}
		}
		// Both complete groups, nonce assignment, data and registration payload.
		p := fixturePacket(POSITION)
		p.GeneratorID = 9
		p.OriginID = v + 1
		p.GroupID = 2
		if err := m.Validate(&p); err != nil {
			t.Fatal(err)
		}
		p = fixturePacket(REGISTER_ACK)
		p.Registration.V = v
		if err := m.Validate(&p); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []uint32{0, 33, math.MaxUint32} {
		if _, err := GroupMask(v); err != ErrMask {
			t.Fatal(err)
		}
	}
	if _, err := ExpectedMask(0, 25); err != ErrIdentity {
		t.Fatal(err)
	}
	if _, err := ExpectedMask(1, 0); err != ErrMask {
		t.Fatal(err)
	}
	if _, err := MissingMask(2, 1); err != ErrMask {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, v, want uint32 }{{1, 1, 0}, {1, 25, 0x01fffffe}, {25, 25, 0x00ffffff}, {1, 32, 0xfffffffe}, {32, 32, 0x7fffffff}} {
		got, err := ExpectedMask(tc.id, tc.v)
		if err != nil || got != tc.want {
			t.Fatalf("%+v: %08x %v", tc, got, err)
		}
	}
}

func TestACKCorrelationRejectsWithoutProgress(t *testing.T) {
	m := fixtureManifest(t, 25)
	op := Operation{42, 99, 1, 1}
	for _, tc := range []struct {
		name   string
		change func(*Packet)
		want   error
	}{
		{"old-run", func(p *Packet) { p.RunID = 41 }, ErrRun},
		{"old-iteration", func(p *Packet) { p.Iteration = 1 }, ErrIteration},
		{"future-iteration", func(p *Packet) { p.Iteration = 3 }, ErrIteration},
		{"other-operation", func(p *Packet) { p.MessageID++ }, ErrMessage},
		{"wrong-target", func(p *Packet) { p.TargetID = 3 }, ErrIdentity},
		{"wrong-group", func(p *Packet) { p.GroupID = 2 }, ErrIdentity},
		{"self", func(p *Packet) { p.OriginID = 1 }, ErrIdentity},
		{"unforwarded", func(p *Packet) { p.Flags = 0 }, ErrFlags},
		{"wrong-type", func(p *Packet) { *p = fixturePacket(START) }, ErrType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixturePacket(ACK)
			p.Flags = ServerToClient
			p.Iteration = 2
			active := op
			active.Iteration = 2
			tc.change(&p)
			got, duplicate, err := m.ApplyACK(&p, active, 4)
			if err != tc.want || got != 4 || duplicate {
				t.Fatalf("got %x %v %v want %v", got, duplicate, err, tc.want)
			}
		})
	}
	p := fixturePacket(ACK)
	p.Flags = ServerToClient
	for _, active := range []Operation{{41, 99, 1, 1}, {42, 0, 1, 1}, {42, 99, 0, 1}, {42, 99, 1, 0}} {
		got, _, err := m.ApplyACK(&p, active, 4)
		if err == nil || got != 4 {
			t.Fatal("invalid expected operation accepted")
		}
	}
	if got, _, err := m.ApplyACK(&p, op, 1); got != 1 || err != ErrMask {
		t.Fatal("invalid received mask accepted")
	}
	// ACK from any attempt of the active operation is useful, including an old attempt.
	p.RetryCounter = math.MaxUint16
	if got, duplicate, err := m.ApplyACK(&p, op, 0); err != nil || duplicate || got != 2 {
		t.Fatalf("%x %v %v", got, duplicate, err)
	}
	// V=1 permits no neighbor ACK or selective retry.
	one := fixtureManifest(t, 1)
	p = fixturePacket(ACK)
	p.Flags = ServerToClient
	if _, _, err := one.ApplyACK(&p, op, 0); err == nil {
		t.Fatal("V=1 accepted neighbor")
	}
	p = fixturePacket(RETRY_POSITION)
	if err := one.Validate(&p); err != ErrMask {
		t.Fatal(err)
	}
}

func TestFiniteFloatBitPatterns(t *testing.T) {
	for _, n := range []uint32{0, 0x80000000, 1, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff} {
		p := fixturePacket(POSITION)
		p.Position.X = math.Float32frombits(n)
		d, err := Decode(wire(t, p))
		if err != nil || math.Float32bits(d.Position.X) != n {
			t.Fatalf("float %08x: %v", n, err)
		}
	}
}

func TestMultipleGroupsInOneGenerator(t *testing.T) {
	c := RunConfig{42, 100, 25, 5}
	m, err := NewManifest(c, []GeneratorConfig{{7, 1, 50, c}, {9, 51, 50, c}})
	if err != nil {
		t.Fatal(err)
	}
	id, group, err := m.RegistrationIdentity(7, 26)
	if err != nil || id != 26 || group != 2 {
		t.Fatalf("second group assignment %d/%d %v", id, group, err)
	}
	p := fixturePacket(ACK)
	p.OriginID, p.TargetID, p.GroupID, p.Flags = 27, 26, 2, ServerToClient
	op := Operation{42, 99, 26, 1}
	if mask, _, err := m.ApplyACK(&p, op, 0); err != nil || mask != 2 {
		t.Fatalf("local slot in second group: %x %v", mask, err)
	}
	p.OriginID = 2 // Same generator, but a different group: not a neighbor.
	if mask, _, err := m.ApplyACK(&p, op, 0); err != ErrIdentity || mask != 0 {
		t.Fatalf("cross-group ACK counted: %x %v", mask, err)
	}
	p = fixturePacket(REGISTER_ACK)
	p.TargetID, p.GroupID, p.RegistrationNonce = 26, 2, 26
	d, err := Decode(wire(t, p))
	if err != nil || m.Validate(&d) != nil {
		t.Fatalf("registration in second group: %+v %v", d, err)
	}
}

func TestRoundTripGroupBoundaries(t *testing.T) {
	for _, v := range []uint32{1, 25, 32} {
		m := fixtureManifest(t, v)
		for _, typ := range []MessageType{REGISTER_ACK, POSITION, RETRY_POSITION, ACK, METRICS} {
			if v == 1 && (typ == RETRY_POSITION || typ == ACK) {
				continue // No neighbor exchange exists for V=1.
			}
			p := fixturePacket(typ)
			p.GeneratorID, p.GroupID = 9, 2
			if p.OriginID != 0 {
				p.OriginID = v + 1
			}
			switch typ {
			case REGISTER_ACK:
				p.Registration.V, p.RegistrationNonce, p.TargetID = v, uint64(v), 2*v
			case RETRY_POSITION:
				p.Flags, p.TargetID, p.MissingMask = ServerToClient, 2*v, uint32(1)<<(v-1)
			case ACK:
				p.Flags, p.OriginID, p.TargetID = ServerToClient, 2*v, v+1
			}
			d, err := Decode(wire(t, p))
			if err != nil || d != p {
				t.Fatalf("V=%d type=%d roundtrip: %v", v, typ, err)
			}
			if err := m.Validate(&d); err != nil {
				t.Fatalf("V=%d type=%d validation: %v", v, typ, err)
			}
		}
	}
}

func TestConcurrentCodecAndManifest(t *testing.T) {
	m := fixtureManifest(t, 25)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := fixturePacket(POSITION)
			var buf [MaxPacketSize]byte
			for j := 0; j < 1000; j++ {
				n, err := Encode(buf[:], &p)
				if err != nil {
					t.Error(err)
					return
				}
				d, err := Decode(buf[:n])
				if err != nil {
					t.Error(err)
					return
				}
				if err := m.Validate(&d); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
