package protocol

import "testing"

var serverControlJobSink serverControlOutbound
var serverControlMaskSink uint32

// These are owner/queue microbenchmarks. No sockets, readers, concurrent workers
// or network pacing run in a fixture. Initial registration and fixed allocation
// are outside the measured loop; restoring the same value and freeing ring slots
// are included, consistently with the existing phase 2/4A benchmarks.
func controlBenchServer(t testing.TB) *Server {
	t.Helper()
	s := bareServer(t, simulationServerConfig(t, 25, 25, 1, 1))
	registerAll(t, s)
	return s
}

type controlACKBenchFixture struct {
	s     *Server
	w     *serverWorker
	event inbound
	group GroupState
	mask  uint32
}

func newControlACKBenchFixture(t testing.TB, typ MessageType) controlACKBenchFixture {
	t.Helper()
	s := controlBenchServer(t)
	w := &s.workers[0]
	phase, message := ServerGroupStarting, serverStartMessageID
	if typ == GROUP_DONE_ACK {
		phase, message = ServerGroupFinishing, serverGroupDoneMessageID
	}
	w.beginGroupControl(0, phase)
	w.pendingGroups.pop() // isolate ACK handling from pending fanout, outside timing
	g := &s.groups[0]
	g.ControlScheduled = false
	mask, _ := GroupMask(25)
	if typ == GROUP_DONE_ACK {
		g.StartAckMask, g.MetricsMask = mask, mask
		for i := range s.clients {
			s.clients[i].Flags |= metricsAccepted
		}
	}
	p := Packet{Header: Header{Type: typ, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: 7, GroupID: 1, OriginID: 1, MessageID: message}}
	return controlACKBenchFixture{s: s, w: w, event: inbound{Packet: p, Source: testEndpoint(0), GeneratorIndex: 0, ClientID: 1, GroupID: 1}, group: *g, mask: mask}
}

func (f *controlACKBenchFixture) process() {
	// Each sample accepts a UNIQUE member bit, never only a duplicate fast path.
	f.s.groups[0] = f.group
	f.s.controlGates[0].mask.Store(f.mask)
	serverErrSink = f.w.handle(f.event)
}

func BenchmarkServerSTARTACKProcessing(b *testing.B) {
	f := newControlACKBenchFixture(b, START_ACK)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.process()
	}
	b.StopTimer()
	if serverErrSink != nil || f.s.groups[0].StartAckMask != 1 {
		b.Fatal("START_ACK did not credit one unique bit", serverErrSink)
	}
}

func BenchmarkServerGROUPDONEACKProcessing(b *testing.B) {
	f := newControlACKBenchFixture(b, GROUP_DONE_ACK)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.process()
	}
	b.StopTimer()
	if serverErrSink != nil || f.s.groups[0].GroupDoneAckMask != 1 {
		b.Fatal("GROUP_DONE_ACK did not credit one unique bit", serverErrSink)
	}
}

type controlMetricsBenchFixture struct {
	s     *Server
	w     *serverWorker
	event inbound
	group GroupState
	c     ServerClientState
}

func newControlMetricsBenchFixture(t testing.TB) controlMetricsBenchFixture {
	t.Helper()
	s := controlBenchServer(t)
	w := &s.workers[0]
	g := &s.groups[0]
	mask, _ := GroupMask(25)
	g.ControlPhase, g.StartAckMask, g.FinalACKMask = ServerGroupRunning, mask, 1
	c := &s.clients[0]
	c.Iteration, c.SeenACKMask = 1, c.ExpectedMask
	p := Packet{Header: Header{Type: METRICS, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: 7, GroupID: 1, OriginID: 1, MessageID: 3}, Metrics: MetricsPayload{MeanRTTNS: 123456789, CompletedIterations: 1}}
	return controlMetricsBenchFixture{s: s, w: w, event: inbound{Packet: p, Source: testEndpoint(0), GeneratorIndex: 0, ClientID: 1, GroupID: 1}, group: *g, c: *c}
}

func (f *controlMetricsBenchFixture) process() {
	// Accept, accumulate and publish one NEW metric, build/enqueue its logical
	// reply into HIGH, then release that fixed ring slot for the next sample.
	f.s.groups[0], f.s.clients[0] = f.group, f.c
	serverErrSink = f.w.handle(f.event)
	f.w.flushPending()
	serverJobSink, _ = f.w.high.pop()
}

func BenchmarkServerMETRICSProcessing(b *testing.B) {
	f := newControlMetricsBenchFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.process()
	}
	b.StopTimer()
	if serverErrSink != nil || f.s.groups[0].MetricsMask != 1 || f.s.groups[0].MetricsSumLo != f.event.Packet.Metrics.MeanRTTNS || serverJobSink.Packet.Type != METRICS_ACK {
		b.Fatal("METRICS did not accept once and enqueue its reactive reply", serverErrSink)
	}
}

func BenchmarkServerControlDeadlineInsertUpdateRemove(b *testing.B) {
	// Same 99 active-entry fixture as the client heap benchmark, without
	// constructing or running a 2500-client server to obtain 100 group records.
	groups := make([]GroupState, 100)
	for i := range groups {
		groups[i].ControlHeapIndex = -1
	}
	h := newServerControlDeadlineHeap(groups, len(groups))
	for i := uint32(1); i < 100; i++ {
		h.set(i, int64(i)*1000)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.set(0, 50000)
		h.set(0, 1)
		h.set(0, 99999)
		h.remove(0)
	}
	b.StopTimer()
	if h.len() != 99 || groups[0].ControlHeapIndex != -1 {
		b.Fatal("control heap accumulated duplicate/stale entries")
	}
}

type controlTimeoutBenchFixture struct {
	s     *Server
	w     *serverWorker
	group GroupState
}

func newControlTimeoutBenchFixture(t testing.TB, missing uint32) controlTimeoutBenchFixture {
	t.Helper()
	s := controlBenchServer(t)
	w := &s.workers[0]
	w.beginGroupControl(0, ServerGroupStarting)
	w.pendingGroups.pop()
	g := &s.groups[0]
	mask, _ := GroupMask(25)
	g.StartAckMask = mask &^ missing
	g.ControlPendingMask, g.ControlAwaitMask = 0, 0
	g.ControlWrittenMask, g.ControlRoundMask = mask, mask
	g.ControlLastSentAt, g.ControlSends, g.ControlScheduled = 1000, 1, false
	return controlTimeoutBenchFixture{s: s, w: w, group: *g}
}

func (f *controlTimeoutBenchFixture) expire() {
	f.s.groups[0] = f.group
	f.w.armControlDeadline(0)
	f.w.processControlTimeouts(f.s.groups[0].ControlDeadline)
}

func BenchmarkServerControlTimeout(b *testing.B) {
	f := newControlTimeoutBenchFixture(b, 1<<24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.expire()
		f.w.pendingGroups.pop() // release the single scheduling reference
	}
	b.StopTimer()
	if f.s.groups[0].ControlAttempt != 1 || f.s.groups[0].ControlPendingMask != 1<<24 || f.s.groups[0].ControlHeapIndex != -1 || f.s.controlHeapCurrent.Load() != 0 {
		b.Fatal("timeout did not remove its deadline and create one selective retry")
	}
}

func controlBenchDrainFanout(w *serverWorker) uint32 {
	for w.pendingGroups.stats().Depth != 0 {
		w.flushControlPending()
	}
	mask := uint32(0)
	for {
		job, ok := w.control.pop()
		if !ok {
			break
		}
		serverControlJobSink = job
		mask |= job.TargetBit
	}
	return mask
}

func BenchmarkServerControlFanoutV25(b *testing.B) {
	s := controlBenchServer(b)
	w, initial := &s.workers[0], s.groups[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.groups[0] = initial
		w.beginGroupControl(0, ServerGroupStarting)
		serverControlMaskSink = controlBenchDrainFanout(w)
	}
	b.StopTimer()
	if serverControlMaskSink != 0x01ffffff || serverControlJobSink.Packet.Type != START || s.groups[0].ControlAwaitMask != 0x01ffffff {
		b.Fatal("initial control round did not produce all 25 distinct targets")
	}
}

func benchmarkServerControlSelectiveRetry(b *testing.B, missing uint32) {
	f := newControlTimeoutBenchFixture(b, missing)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.expire()
		serverControlMaskSink = controlBenchDrainFanout(f.w)
	}
	b.StopTimer()
	if serverControlMaskSink != missing || serverControlJobSink.Packet.Type != START || serverControlJobSink.Packet.MessageID != serverStartMessageID || serverControlJobSink.Packet.RetryCounter != 1 || f.s.groups[0].ControlAwaitMask != missing || f.s.groups[0].ControlHeapIndex != -1 {
		b.Fatal("control retry did not send exactly the missing members with stable identity")
	}
}

func BenchmarkServerControlSelectiveRetry(b *testing.B) {
	b.Run("1", func(b *testing.B) { benchmarkServerControlSelectiveRetry(b, 1<<24) })
	b.Run("24", func(b *testing.B) { benchmarkServerControlSelectiveRetry(b, 0x01fffffe) })
}

func TestServerControlHotPathZeroAllocations(t *testing.T) {
	for _, typ := range []MessageType{START_ACK, GROUP_DONE_ACK} {
		f := newControlACKBenchFixture(t, typ)
		if n := testing.AllocsPerRun(1000, f.process); n != 0 || serverErrSink != nil {
			t.Fatal("control ACK owner path allocates or rejects a valid ACK", typ, n, serverErrSink)
		}
	}
	m := newControlMetricsBenchFixture(t)
	if n := testing.AllocsPerRun(1000, m.process); n != 0 || serverErrSink != nil || serverJobSink.Packet.Type != METRICS_ACK {
		t.Fatal("METRICS owner/reply path allocates or rejects valid metrics", n, serverErrSink)
	}
	f := newControlTimeoutBenchFixture(t, 1<<24)
	if n := testing.AllocsPerRun(1000, func() {
		f.expire()
		serverControlMaskSink = controlBenchDrainFanout(f.w)
	}); n != 0 || serverControlMaskSink != 1<<24 {
		t.Fatal("control timeout/selective fanout allocates or sends wrong targets", n, serverControlMaskSink)
	}
}
