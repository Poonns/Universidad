package protocol

import "testing"

// Timed samples include restoring the fixed fixture and releasing its ring slot.
// Initialization, authoritative registration and socket I/O remain outside them.
func BenchmarkClientTimeoutProcessing(b *testing.B) {
	g := arqClients(b, 25, 1)
	w, original := arqStart(b, g, 1, 1, 1000)
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial
		g.sendGates[0].Store(original.Ticket | clientTicketClaimed)
		w.armDeadline(1, &g.clients[0], 1000)
		w.processTimeouts(g.clients[0].Deadline)
		w.pending.pop() // next sample replaces flags and reuses this same bounded slot
	}
	b.StopTimer()
	if g.clients[0].RetryCounter != 1 || g.clients[0].HeapIndex != -1 {
		b.Fatal("benchmark did not process an actual timeout", g.clients[0])
	}
}

func benchmarkClientSelectiveRetry(b *testing.B, missing uint32) {
	g := arqClients(b, 25, 1)
	w, original := arqStart(b, g, 1, 1, 1000)
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	initial.AckMask = initial.ExpectedMask &^ missing
	initial.LastUniqueACKAt = 2000
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial
		g.sendGates[0].Store(original.Ticket | clientTicketClaimed)
		w.armDeadline(1, &g.clients[0], 1000)
		w.processTimeouts(g.clients[0].Deadline)
		w.flushPending()
		clientJobSink, _ = w.retry.pop()
	}
	b.StopTimer()
	if clientJobSink.Packet.Type != RETRY_POSITION || clientJobSink.Packet.MissingMask != missing || clientJobSink.Packet.RetryCounter != 1 {
		b.Fatal("benchmark did not produce an exact selective retry", clientJobSink)
	}
}

func BenchmarkClientSelectiveRetry(b *testing.B) {
	b.Run("1", func(b *testing.B) { benchmarkClientSelectiveRetry(b, 1<<24) })
	b.Run("4", func(b *testing.B) { benchmarkClientSelectiveRetry(b, 2|4|8|16) })
	b.Run("24", func(b *testing.B) { benchmarkClientSelectiveRetry(b, 0x01fffffe) })
}

func TestClientARQHotPathAllocations(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, original := arqStart(t, g, 1, 1, 1000)
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	initial.AckMask = initial.ExpectedMask &^ (1 << 24)
	initial.LastUniqueACKAt = 2000
	allocations := testing.AllocsPerRun(1000, func() {
		g.clients[0] = initial
		g.sendGates[0].Store(original.Ticket | clientTicketClaimed)
		w.armDeadline(1, &g.clients[0], 1000)
		w.processTimeouts(g.clients[0].Deadline)
		w.flushPending()
		clientJobSink, _ = w.retry.pop()
	})
	if allocations != 0 || clientJobSink.Packet.MissingMask != 1<<24 {
		t.Fatalf("ARQ timeout/selective retry: allocs/op=%g retry=%+v", allocations, clientJobSink)
	}
}
