package protocol

import "testing"

var academicPositionSink PositionPayload

// These benchmarks include restoring the compact state and freeing fixed ring
// slots. Initialization, registration and all UDP I/O remain outside the timing.
func BenchmarkClientSTARTProcessing(b *testing.B) {
	g := academicClients(b, 25, 25, 5)
	w := &g.workers[0]
	initial := g.clients[0]
	p := clientAcademicPacket(g, START, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
		clientJobSink, _ = w.high.pop()
		w.pending.pop()
	}
	b.StopTimer()
	if clientErrSink != nil || clientJobSink.Packet.Type != START_ACK || g.clients[0].Iteration != 1 {
		b.Fatal(clientErrSink, clientJobSink)
	}
}

func BenchmarkClientMETRICSACKProcessing(b *testing.B) {
	g := academicClients(b, 25, 25, 1)
	w, job := academicMetricsJob(b, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		b.Fatal(err)
	}
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	p := clientAcademicPacket(g, METRICS_ACK, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
	}
	b.StopTimer()
	if clientErrSink != nil || g.clients[0].Phase != ClientWaitGroupDone {
		b.Fatal(clientErrSink, g.clients[0])
	}
}

func BenchmarkClientGROUPDONEProcessing(b *testing.B) {
	g := academicClients(b, 25, 25, 1)
	w, job := academicMetricsJob(b, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		b.Fatal(err)
	}
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	p := clientAcademicPacket(g, GROUP_DONE, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.clients[0] = initial
		g.counters.doneClients.Store(0) // one target, never fake global completion
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
		clientJobSink, _ = w.high.pop()
	}
	b.StopTimer()
	if clientErrSink != nil || g.clients[0].Phase != ClientDone || clientJobSink.Packet.Type != GROUP_DONE_ACK {
		b.Fatal(clientErrSink, clientJobSink)
	}
}

func TestClientAcademicControlHotPathAllocations(t *testing.T) {
	g := academicClients(t, 25, 25, 1)
	w := &g.workers[0]
	ready := g.clients[0]
	p := clientAcademicPacket(g, START, 1)
	if allocations := testing.AllocsPerRun(1000, func() {
		g.clients[0] = ready
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
		w.high.pop()
		w.pending.pop()
	}); allocations != 0 || clientErrSink != nil {
		t.Fatal("START hot path allocates", allocations, clientErrSink)
	}
	// Return to one genuine pending operation to create the METRICS fixture.
	g.clients[0] = ready
	w, job := academicMetricsJob(t, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		t.Fatal(err)
	}
	w.removeDeadline(1, &g.clients[0])
	initial := g.clients[0]
	p = clientAcademicPacket(g, METRICS_ACK, 1)
	if allocations := testing.AllocsPerRun(1000, func() {
		g.clients[0] = initial
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
	}); allocations != 0 || clientErrSink != nil {
		t.Fatal("METRICS_ACK hot path allocates", allocations, clientErrSink)
	}
	p = clientAcademicPacket(g, GROUP_DONE, 1)
	if allocations := testing.AllocsPerRun(1000, func() {
		g.clients[0] = initial
		g.counters.doneClients.Store(0)
		clientErrSink = w.handleAcademicControl(&p, &g.clients[0])
		w.high.pop()
	}); allocations != 0 || clientErrSink != nil {
		t.Fatal("GROUP_DONE hot path allocates", allocations, clientErrSink)
	}
	if allocations := testing.AllocsPerRun(1000, func() { academicPositionSink = automaticPosition(42, 9, 1, 1) }); allocations != 0 {
		t.Fatal("coordinates allocate", allocations)
	}
}
