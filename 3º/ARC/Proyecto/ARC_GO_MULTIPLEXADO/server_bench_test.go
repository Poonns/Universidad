package protocol

import "testing"

var serverJobSink outbound
var serverErrSink error

func BenchmarkServerDecodeDispatch(b *testing.B) {
	s := bareServer(b, serverConfig(b, 25, 25, 1, false))
	src := testEndpoint(0)
	data := wire(b, positionPacket(s, 1))
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serverErrSink = s.decodeDispatch(data, src)
		s.workers[0].input.pop() // release the bounded slot; included in measurement
	}
	b.StopTimer()
	if serverErrSink != nil {
		b.Fatal(serverErrSink)
	}
}

func benchmarkServerFanout(b *testing.B, retryMask uint32) {
	s := bareServer(b, serverConfig(b, 25, 25, 1, false))
	registerAll(b, s)
	p := positionPacket(s, 1)
	if retryMask != 0 {
		p.Type = RETRY_POSITION
		p.RetryCounter = 1
		p.MissingMask = retryMask
	}
	e := inbound{p, testEndpoint(0), 0, 1, 1}
	w := &s.workers[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serverErrSink = w.handle(e)
		for w.pending.stats().Depth > 0 {
			w.flushPending()
		}
		for {
			job, ok := w.data.pop()
			if !ok {
				break
			}
			serverJobSink = job
		}
	}
	b.StopTimer()
	if serverErrSink != nil {
		b.Fatal(serverErrSink)
	}
}
func BenchmarkServerPOSITIONFanout24(b *testing.B) { benchmarkServerFanout(b, 0) }
func BenchmarkServerRETRYFanout(b *testing.B) {
	b.Run("1", func(b *testing.B) { benchmarkServerFanout(b, 2) })
	b.Run("4", func(b *testing.B) { benchmarkServerFanout(b, 2|4|8|16) })
	b.Run("24", func(b *testing.B) { benchmarkServerFanout(b, 0x01fffffe) })
}
func BenchmarkServerACKOutbound(b *testing.B) {
	s := bareServer(b, serverConfig(b, 25, 25, 1, false))
	registerAll(b, s)
	p := positionPacket(s, 1)
	if err := deliver(b, s, p, testEndpoint(0)); err != nil {
		b.Fatal(err)
	}
	collectOutbound(b, s)
	e := inbound{ackFor(p, 2), testEndpoint(0), 0, 2, 1}
	w := &s.workers[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.clients[0].SeenACKMask = 0 // reset fixture so each sample is a unique real ACK
		serverErrSink = w.handle(e)
		serverJobSink, _ = w.high.pop()
	}
	b.StopTimer()
	if serverErrSink != nil {
		b.Fatal(serverErrSink)
	}
}
func TestServerHotPathAllocations(t *testing.T) {
	s := bareServer(t, serverConfig(t, 25, 25, 1, false))
	registerAll(t, s)
	p := positionPacket(s, 1)
	src := testEndpoint(0)
	data := wire(t, p)
	w := &s.workers[0]
	check := func(name string, fn func()) {
		t.Helper()
		if allocs := testing.AllocsPerRun(1000, fn); allocs != 0 {
			t.Fatalf("%s: %v allocs/op", name, allocs)
		}
		if serverErrSink != nil {
			t.Fatal(name, serverErrSink)
		}
	}
	check("decode-dispatch", func() { serverErrSink = s.decodeDispatch(data, src); w.input.pop() })
	e := inbound{p, src, 0, 1, 1}
	check("position-fanout", func() {
		serverErrSink = w.handle(e)
		for w.pending.stats().Depth > 0 {
			w.flushPending()
		}
		for {
			job, ok := w.data.pop()
			if !ok {
				break
			}
			serverJobSink = job
		}
	})
	e.Packet = ackFor(p, 2)
	e.ClientID = 2
	check("ack-outbound", func() { s.clients[0].SeenACKMask = 0; serverErrSink = w.handle(e); serverJobSink, _ = w.high.pop() })
	e.Packet = p
	e.Packet.Type = RETRY_POSITION
	e.Packet.RetryCounter = 1
	e.Packet.MissingMask = 2
	e.ClientID = 1
	check("selective-retry", func() { serverErrSink = w.handle(e); w.flushPending(); serverJobSink, _ = w.data.pop() })
}
