package protocol

import "testing"

var packetSink Packet
var errorSink error
var sizeSink int
var byteSink byte

func benchmarkEncode(b *testing.B, typ MessageType) {
	p := fixturePacket(typ)
	var buf [MaxPacketSize]byte
	n, err := Encode(buf[:], &p)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sizeSink, errorSink = Encode(buf[:], &p)
	}
	b.StopTimer()
	byteSink = buf[0]
	if errorSink != nil {
		b.Fatal(errorSink)
	}
}

func benchmarkDecode(b *testing.B, typ MessageType) {
	buf := wire(b, fixturePacket(typ))
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		packetSink, errorSink = Decode(buf)
	}
	b.StopTimer()
	if errorSink != nil {
		b.Fatal(errorSink)
	}
}

func BenchmarkEncodePOSITION(b *testing.B) { benchmarkEncode(b, POSITION) }
func BenchmarkDecodePOSITION(b *testing.B) { benchmarkDecode(b, POSITION) }
func BenchmarkEncodeACK(b *testing.B)      { benchmarkEncode(b, ACK) }
func BenchmarkDecodeACK(b *testing.B)      { benchmarkDecode(b, ACK) }

func TestHotPathZeroAllocations(t *testing.T) {
	m := fixtureManifest(t, 25)
	for _, typ := range []MessageType{POSITION, ACK} {
		p := fixturePacket(typ)
		var buf [MaxPacketSize]byte
		n, err := Encode(buf[:], &p)
		if err != nil {
			t.Fatal(err)
		}
		if got := testing.AllocsPerRun(1000, func() { sizeSink, errorSink = Encode(buf[:], &p) }); got != 0 {
			t.Fatalf("Encode %d: %v allocs/op", typ, got)
		}
		if got := testing.AllocsPerRun(1000, func() { packetSink, errorSink = Decode(buf[:n]) }); got != 0 {
			t.Fatalf("Decode %d: %v allocs/op", typ, got)
		}
		if got := testing.AllocsPerRun(1000, func() { errorSink = m.Validate(&p) }); got != 0 {
			t.Fatalf("Validate %d: %v allocs/op", typ, got)
		}
	}
	p := fixturePacket(ACK)
	p.Flags = ServerToClient
	op := Operation{42, 99, 1, 1}
	if got := testing.AllocsPerRun(1000, func() { _, _, errorSink = m.ApplyACK(&p, op, 0) }); got != 0 {
		t.Fatalf("ApplyACK: %v allocs/op", got)
	}
}
