package protocol

import "testing"

var clientJobSink clientOutbound
var clientErrSink error

func BenchmarkClientDecodeDispatch(b *testing.B) {
	g := bareClients(b, clientsConfig(b, 25, 25, 1, 1))
	data := wire(b, clientForwardedPosition(g, 1, 2))
	source := g.cfg.Server
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientErrSink = g.decodeDispatch(data, source, 2000)
		g.workers[0].input.pop() // release the bounded slot; included in measurement
	}
	b.StopTimer()
	if clientErrSink != nil {
		b.Fatal(clientErrSink)
	}
}

func BenchmarkClientNeighborPOSITIONToACK(b *testing.B) {
	g := bareClients(b, clientsConfig(b, 25, 25, 1, 1))
	readyClients(b, g)
	w := &g.workers[0]
	e := clientInbound{clientForwardedPosition(g, 1, 2), 2000}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientErrSink = w.handleIncoming(e)
		clientJobSink, _ = w.high.pop()
	}
	b.StopTimer()
	if clientErrSink != nil || clientJobSink.Packet.Type != ACK {
		b.Fatal(clientErrSink)
	}
}

func BenchmarkClientACKToMask(b *testing.B) {
	g := bareClients(b, clientsConfig(b, 25, 25, 1, 1))
	readyClients(b, g)
	w, own := clientBegin(b, g, 1, 1)
	if err := w.handleConfirmation(clientPositionConfirmation(own, 1000)); err != nil {
		b.Fatal(err)
	}
	e := clientInbound{clientForwardedACK(own.Packet, 2), 2000}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0].AckMask = 0 // unique ACK every sample, not just duplicate handling
		g.clients[0].LastUniqueACKAt = -1
		clientErrSink = w.handleIncoming(e)
	}
	b.StopTimer()
	if clientErrSink != nil || g.clients[0].AckMask != 2 {
		b.Fatal(clientErrSink)
	}
}

func BenchmarkClientOwnPOSITIONToOutbound(b *testing.B) {
	g := bareClients(b, clientsConfig(b, 25, 25, 1, 1))
	readyClients(b, g)
	w := &g.workers[0]
	initial := g.clients[0]
	cmd := clientCommand{1, 1, PositionPayload{1, 2, 3}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial // restore a registered READY fixture
		clientErrSink = w.beginPosition(cmd)
		w.flushPending()
		clientJobSink, _ = w.data.pop()
	}
	b.StopTimer()
	if clientErrSink != nil || clientJobSink.Packet.Type != POSITION {
		b.Fatal(clientErrSink)
	}
}

func BenchmarkClientWriterConfirmationToSentAt(b *testing.B) {
	g := bareClients(b, clientsConfig(b, 25, 25, 1, 1))
	readyClients(b, g)
	w, own := clientBegin(b, g, 1, 1)
	initial := g.clients[0]
	tx := clientPositionConfirmation(own, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		heapIndex := g.clients[0].HeapIndex
		g.clients[0] = initial
		g.clients[0].HeapIndex = heapIndex // restore fixture without orphaning the existing heap entry
		clientErrSink = w.handleConfirmation(tx)
	}
	b.StopTimer()
	if clientErrSink != nil || g.clients[0].SentAt != 1000 {
		b.Fatal(clientErrSink)
	}
}

// Deterministic allocation budget assertions complement the timed benchmarks.
func TestClientHotPathAllocations(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 25, 25, 1, 1))
	readyClients(t, g)
	w := &g.workers[0]
	neighbor := clientInbound{clientForwardedPosition(g, 1, 2), 2000}
	data := wire(t, neighbor.Packet)
	ready := g.clients[0]
	_, own := clientBegin(t, g, 1, 1)
	queued := g.clients[0]
	tx := clientPositionConfirmation(own, 1000)
	ack := clientInbound{clientForwardedACK(own.Packet, 2), 2000}
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"decode-dispatch", func() { clientErrSink = g.decodeDispatch(data, g.cfg.Server, 2000); w.input.pop() }},
		{"neighbor-position-ack", func() { clientErrSink = w.handleIncoming(neighbor); w.high.pop() }},
		{"ack-mask", func() {
			w.removeDeadline(1, &g.clients[0])
			g.clients[0] = queued
			clientErrSink = w.handleIncoming(ack)
		}},
		{"own-position-outbound", func() {
			w.removeDeadline(1, &g.clients[0])
			g.clients[0] = ready
			clientErrSink = w.beginPosition(clientCommand{1, 1, PositionPayload{1, 2, 3}})
			w.flushPending()
			w.data.pop()
		}},
		{"writer-confirmation-sentat", func() {
			heapIndex := g.clients[0].HeapIndex
			g.clients[0] = queued
			g.clients[0].HeapIndex = heapIndex
			clientErrSink = w.handleConfirmation(tx)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if n := testing.AllocsPerRun(1000, tc.fn); n != 0 || clientErrSink != nil {
				t.Fatalf("allocs/op=%g error=%v", n, clientErrSink)
			}
		})
	}
}
