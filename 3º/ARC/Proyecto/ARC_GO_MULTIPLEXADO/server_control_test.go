package protocol

import (
	"context"
	"math"
	"math/bits"
	"testing"
	"time"
	"unsafe"
)

func academicServerConfig(t testing.TB, n, v uint32, workers int) ServerConfig {
	cfg := serverConfig(t, n, v, workers, false)
	cfg.Academic = true
	cfg.ControlCapacity = 128
	cfg.ConfirmationCapacity = 128
	cfg.ControlRate = 1000
	cfg.ControlBurst = 32
	cfg.InitialControlRTO = 100 * time.Nanosecond
	cfg.MaxControlRTO = 500 * time.Nanosecond
	cfg.ControlJitterPercent = 0
	return cfg
}
func readyAcademicServer(t testing.TB, n, v uint32, workers int) *Server {
	t.Helper()
	s := bareServer(t, academicServerConfig(t, n, v, workers))
	registerAll(t, s)
	for i := range s.workers {
		s.workers[i].stageStarts()
	}
	return s
}
func collectControl(t testing.TB, s *Server) []serverControlOutbound {
	t.Helper()
	var out []serverControlOutbound
	for turn := 0; turn < 10000; turn++ {
		progress := false
		for i := range s.workers {
			w := &s.workers[i]
			if w.flushControlPending() {
				progress = true
			}
			for {
				job, ok := w.control.pop()
				if !ok {
					break
				}
				out = append(out, job)
				progress = true
			}
		}
		if !progress {
			return out
		}
	}
	t.Fatal("control did not drain")
	return nil
}
func confirmControl(t testing.TB, s *Server, job serverControlOutbound, sentAt int64) {
	t.Helper()
	owner, _ := s.WorkerForGroup(job.GroupID)
	if err := s.workers[owner].handleControlConfirmation(serverControlConfirmation{Ticket: job.Ticket, SentAt: sentAt, GroupID: job.GroupID,
		TargetBit: job.TargetBit, Attempt: job.Attempt, Type: job.Packet.Type}); err != nil {
		t.Fatal(err)
	}
}
func serverControlACK(job serverControlOutbound) Packet {
	typ := START_ACK
	if job.Packet.Type == GROUP_DONE {
		typ = GROUP_DONE_ACK
	}
	return Packet{Header: Header{Type: typ, RunID: job.Packet.RunID, GeneratorID: job.Packet.GeneratorID, GroupID: job.GroupID,
		OriginID: job.Packet.TargetID, MessageID: job.Packet.MessageID, RetryCounter: job.Packet.RetryCounter}}
}
func acknowledgeControl(t testing.TB, s *Server, job serverControlOutbound) {
	t.Helper()
	if err := deliver(t, s, serverControlACK(job), testEndpoint(job.GeneratorIndex)); err != nil {
		t.Fatal(err)
	}
}
func startAcademicGroups(t testing.TB, s *Server) {
	t.Helper()
	for _, job := range collectControl(t, s) {
		confirmControl(t, s, job, 100)
		acknowledgeControl(t, s, job)
	}
}
func serverMetricPacket(s *Server, id uint32, mean uint64) Packet {
	index := (id - 1) / s.cfg.Manifest.run.V
	gi := int(s.groups[index].GeneratorIndex)
	gen := s.cfg.Manifest.generators[gi]
	return Packet{Header: Header{Type: METRICS, RunID: s.cfg.Manifest.run.RunID, GeneratorID: gen.GeneratorID, GroupID: index + 1,
		OriginID: id, MessageID: uint64(id-gen.FirstID) + uint64(s.cfg.Manifest.run.I) + 2}, Metrics: MetricsPayload{mean, s.cfg.Manifest.run.I}}
}
func accreditServerFinalData(s *Server, id uint32) {
	c := &s.clients[id-1]
	c.Iteration = s.cfg.Manifest.run.I
	c.SeenACKMask = c.ExpectedMask
	s.groups[(id-1)/s.cfg.Manifest.run.V].FinalACKMask |= uint32(1) << ((id - 1) % s.cfg.Manifest.run.V)
}
func acceptServerMetric(t testing.TB, s *Server, id uint32, mean uint64) {
	t.Helper()
	accreditServerFinalData(s, id)
	p := serverMetricPacket(s, id, mean)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
}

func TestServerControlGlobalRegistrationBarrier(t *testing.T) {
	s := bareServer(t, academicServerConfig(t, 3, 3, 2))
	for nonce := uint64(1); nonce < 3; nonce++ {
		if err := deliver(t, s, registerPacket(s, 0, nonce), testEndpoint(0)); err != nil {
			t.Fatal(err)
		}
		collectOutbound(t, s)
		for i := range s.workers {
			if s.workers[i].stageStarts() {
				t.Fatal("START before N")
			}
		}
		if jobs := collectControl(t, s); len(jobs) != 0 {
			t.Fatal(jobs)
		}
	}
	if err := deliver(t, s, registerPacket(s, 0, 3), testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	for i := range s.workers {
		s.workers[i].stageStarts()
	}
	jobs := collectControl(t, s)
	if len(jobs) != 3 || s.Stats().StartedGroups != 1 {
		t.Fatal(len(jobs), s.Stats())
	}
	var mask uint32
	for _, job := range jobs {
		mask |= job.TargetBit
		if job.Packet.Type != START || job.Packet.MessageID != 1 || job.Attempt != 0 || s.cfg.Manifest.Validate(&job.Packet) != nil {
			t.Fatal(job)
		}
	}
	if mask != 7 || s.controlHeapCurrent.Load() != 0 || s.groups[0].ControlAwaitMask != 7 {
		t.Fatal(s.groups[0])
	}
}

func TestServerControlDeadlinesWaitForActualRoundWrites(t *testing.T) {
	cfg := academicServerConfig(t, 25, 25, 1)
	cfg.ControlCapacity = 1
	s := bareServer(t, cfg)
	registerAll(t, s)
	w := &s.workers[0]
	w.stageStarts()
	for target := 0; target < 25; target++ {
		w.flushControlPending()
		job, ok := w.control.pop()
		if !ok {
			t.Fatal(target)
		}
		if s.controlHeapCurrent.Load() != 0 {
			t.Fatal("deadline while targets await queue/write")
		}
		confirmControl(t, s, job, 1000+int64(target))
		if target < 24 && s.controlHeapCurrent.Load() != 0 {
			t.Fatal("deadline armed before last receipt")
		}
	}
	g := &s.groups[0]
	if g.ControlWrittenMask != 0x1ffffff || g.ControlAwaitMask != 0 || g.ControlPendingMask != 0 || g.ControlSends != 1 ||
		g.ControlDeadline != 1124 || s.controlHeapCurrent.Load() != 1 {
		t.Fatal(*g)
	}
	if w.processControlTimeouts(1123) {
		t.Fatal("early timeout")
	}
	if !w.processControlTimeouts(1124) || g.ControlAttempt != 1 || g.ControlRTO != 200 || s.controlHeapCurrent.Load() != 0 {
		t.Fatal(*g)
	}
	for i := 0; i < 100; i++ {
		w.processControlTimeouts(math.MaxInt64)
	}
	if g.ControlAttempt != 1 || s.Stats().ControlTimeouts != 1 {
		t.Fatal("queue waiting generates repeated retries", s.Stats())
	}
}

func TestServerControlSelectiveRetryAndACKBeforeConfirmation(t *testing.T) {
	s := readyAcademicServer(t, 25, 25, 1)
	w := &s.workers[0]
	jobs := collectControl(t, s)
	for _, job := range jobs {
		confirmControl(t, s, job, 100)
	}
	for _, job := range jobs[:24] {
		acknowledgeControl(t, s, job)
	}
	g := &s.groups[0]
	w.processControlTimeouts(g.ControlDeadline)
	retry := collectControl(t, s)
	if len(retry) != 1 || retry[0].TargetBit != 1<<24 || retry[0].Packet.TargetID != 25 || retry[0].Attempt != 1 {
		t.Fatal(retry)
	}
	// A real original ACK remains valid while the retry waits for its local receipt.
	acknowledgeControl(t, s, jobs[24])
	if g.ControlPhase != ServerGroupRunning || g.StartAckMask != 0x1ffffff || !s.controlCancelled(retry[0]) || w.controlDeadlines.len() != 0 {
		t.Fatal(*g)
	}
	confirmControl(t, s, retry[0], 999)
	if w.controlDeadlines.len() != 0 || g.ControlPhase != ServerGroupRunning || s.Stats().StartedClients != 25 {
		t.Fatal(*g, s.Stats())
	}
	acknowledgeControl(t, s, jobs[24])
	if s.Stats().StartedClients != 25 || s.Stats().DuplicateStartACKs != 1 {
		t.Fatal(s.Stats())
	}
}

func TestServerControlMetricsWaitForMissingSTARTACK(t *testing.T) {
	s := readyAcademicServer(t, 3, 3, 1)
	jobs := collectControl(t, s)
	for _, job := range jobs {
		confirmControl(t, s, job, 100)
	}
	acknowledgeControl(t, s, jobs[0])
	acknowledgeControl(t, s, jobs[1])
	for id := uint32(1); id <= 3; id++ {
		acceptServerMetric(t, s, id, uint64(id)*10)
	}
	if s.groups[0].ControlPhase != ServerGroupStarting || s.groups[0].MetricsMask != 7 || len(collectControl(t, s)) != 0 {
		t.Fatal(s.groups[0])
	}
	acknowledgeControl(t, s, jobs[2])
	done := collectControl(t, s)
	if len(done) != 3 || done[0].Packet.Type != GROUP_DONE || done[0].Packet.MessageID != 2 || s.controlHeapCurrent.Load() != 0 {
		t.Fatal(done, s.groups[0])
	}
	// Receipt from START cannot start or retime GROUP_DONE's deadline.
	confirmControl(t, s, jobs[2], 1000)
	if s.controlHeapCurrent.Load() != 0 || s.groups[0].ControlWrittenMask != 0 {
		t.Fatal(s.groups[0])
	}
}

func TestServerControlMetricsValidationIdempotencyAnd128BitMean(t *testing.T) {
	s := readyAcademicServer(t, 3, 3, 1)
	startAcademicGroups(t, s)
	p := serverMetricPacket(s, 1, math.MaxUint64)
	if err := deliver(t, s, p, testEndpoint(0)); err != ErrProgress {
		t.Fatal("METRICS without observed final ACK", err)
	}
	accreditServerFinalData(s, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	conflict := p
	conflict.Metrics.MeanRTTNS--
	if err := deliver(t, s, conflict, testEndpoint(0)); err != ErrConflict {
		t.Fatal(err)
	}
	bad := p
	bad.MessageID++
	if err := deliver(t, s, bad, testEndpoint(0)); err != ErrMessage {
		t.Fatal(err)
	}
	if s.Stats().MetricsAccepted != 1 || s.Stats().DuplicateMetrics != 1 || s.clients[0].MeanRTTNS != math.MaxUint64 {
		t.Fatal(s.Stats())
	}
	acceptServerMetric(t, s, 2, math.MaxUint64)
	acceptServerMetric(t, s, 3, math.MaxUint64)
	st := s.GroupStats()[0]
	if st.MetricsClients != 3 || st.MeanRTTNS != math.MaxUint64 || st.SumHi != 2 || st.SumLo != math.MaxUint64-2 || s.Stats().MeanRTTNS != math.MaxUint64 {
		t.Fatal(st, s.Stats())
	}
	replies := collectOutbound(t, s)
	if len(replies) != 3 {
		t.Fatal(replies)
	}
	for _, job := range replies {
		if job.Packet.Type != METRICS_ACK || job.Packet.MessageID != serverMetricPacket(s, job.Packet.TargetID, 0).MessageID {
			t.Fatal(job)
		}
	}
	for _, job := range collectControl(t, s) {
		confirmControl(t, s, job, 200)
		acknowledgeControl(t, s, job)
		acknowledgeControl(t, s, job)
	}
	if err := s.WaitSimulationComplete(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !s.Stats().SimulationComplete || s.Stats().DoneClients != 3 || s.Stats().CompletedGroups != 1 || s.Stats().DuplicateGroupDoneACKs != 3 || s.controlHeapCurrent.Load() != 0 {
		t.Fatal(s.Stats())
	}
}

func TestServerControlGROUPDONEIndependentPerGroup(t *testing.T) {
	s := readyAcademicServer(t, 4, 2, 2)
	startAcademicGroups(t, s)
	acceptServerMetric(t, s, 1, 10)
	acceptServerMetric(t, s, 2, 20)
	jobs := collectControl(t, s)
	if len(jobs) != 2 || jobs[0].GroupID != 1 || jobs[1].GroupID != 1 {
		t.Fatal(jobs)
	}
	for _, job := range jobs {
		confirmControl(t, s, job, 200)
		acknowledgeControl(t, s, job)
	}
	if s.Stats().CompletedGroups != 1 || s.Stats().SimulationComplete || s.groups[1].ControlPhase != ServerGroupRunning {
		t.Fatal(s.Stats())
	}
	acceptServerMetric(t, s, 3, 30)
	acceptServerMetric(t, s, 4, 40)
	for _, job := range collectControl(t, s) {
		confirmControl(t, s, job, 300)
		acknowledgeControl(t, s, job)
	}
	if err := s.WaitSimulationComplete(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.Stats().MeanRTTNS != 25 || s.Stats().CompletedGroups != 2 {
		t.Fatal(s.Stats())
	}
}

func TestServerControlPermanentACKExhaustsFiniteWrites(t *testing.T) {
	for _, done := range []bool{false, true} {
		name := "START"
		if done {
			name = "GROUP_DONE"
		}
		t.Run(name, func(t *testing.T) {
			cfg := academicServerConfig(t, 3, 3, 1)
			cfg.MaxControlSends = 3
			s := bareServer(t, cfg)
			registerAll(t, s)
			w := &s.workers[0]
			w.stageStarts()
			if done {
				startAcademicGroups(t, s)
				for id := uint32(1); id <= 3; id++ {
					acceptServerMetric(t, s, id, 10)
				}
			}
			written := 0
			for round := 0; round < 3; round++ {
				jobs := collectControl(t, s)
				if len(jobs) != 3 {
					t.Fatal(round, jobs)
				}
				for _, job := range jobs {
					confirmControl(t, s, job, 1000+int64(round)*1000)
					written++
				}
				if s.groups[0].ControlSends != uint16(round+1) || w.controlDeadlines.len() != 1 {
					t.Fatal(s.groups[0])
				}
				w.processControlTimeouts(s.groups[0].ControlDeadline)
			}
			if written != 9 || len(collectControl(t, s)) != 0 || s.Stats().ControlRetryExhausted != 1 || s.Stats().ControlTimeouts != 3 ||
				s.groups[0].ControlPhase != ServerGroupFailed || s.Stats().SimulationComplete || w.controlDeadlines.len() != 0 {
				t.Fatal(written, s.Stats())
			}
			if err := s.WaitSimulationComplete(context.Background()); err != ErrControlRetryExhausted {
				t.Fatal(err)
			}
			// Both completion signals are ready after shutdown. Either select path
			// must preserve the terminal failure rather than returning "incomplete".
			close(s.done)
			for n := 0; n < 100; n++ {
				if err := s.WaitSimulationComplete(context.Background()); err != ErrControlRetryExhausted {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestServerAcademicNextIterationCancelsObsoleteIntentions(t *testing.T) {
	cfg := academicServerConfig(t, 3, 3, 1)
	run := cfg.Manifest.run
	run.I = 2
	cfg.Manifest, _ = NewManifest(run, []GeneratorConfig{{7, 1, 3, run}})
	s := bareServer(t, cfg)
	registerAll(t, s)
	s.workers[0].stageStarts()
	startAcademicGroups(t, s)
	p := positionPacket(s, 1)
	p.MessageID = 2
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	// Pending fanout has not yet been enqueued; real ACKs of a retry can finish it.
	for _, neighbor := range []uint32{2, 3} {
		if err := deliver(t, s, ackFor(p, neighbor), testEndpoint(0)); err != nil {
			t.Fatal(err)
		}
	}
	s.clients[0].PendingRetryMask = 6
	next := p
	next.Iteration = 2
	next.MessageID = 3
	if err := deliver(t, s, next, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if s.clients[0].PendingRetryMask != 0 || s.clients[0].PendingPositionMask != 6 || s.clients[0].Iteration != 2 {
		t.Fatal(s.clients[0])
	}
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if err := deliver(t, s, ackFor(p, 2), testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if s.Stats().StalePackets != 2 || s.Stats().InvalidPackets != 0 {
		t.Fatal(s.Stats())
	}
}

func TestServerControlHeapOwnershipAndNoGrowth(t *testing.T) {
	s := bareServer(t, academicServerConfig(t, 12, 3, 2))
	h := &s.workers[0].controlDeadlines
	if !h.set(0, 30) || !h.set(2, 20) || !h.set(0, 10) {
		t.Fatal("heap insert/update")
	}
	i, d, ok := h.peek()
	if !ok || i != 0 || d != 10 || h.len() != 2 || cap(h.indices) != 2 {
		t.Fatal(i, d, ok, h)
	}
	if h.set(1, 40) {
		t.Fatal("heap grew beyond owner capacity")
	}
	if !h.remove(0) || !h.set(0, 50) || !h.remove(2) || !h.remove(0) || h.len() != 0 {
		t.Fatal("heap remove/reuse")
	}
	for _, w := range s.workers {
		if w.controlDeadlines.len() != 0 {
			t.Fatal("cross-owner heap changed")
		}
	}
}

func TestServerControlBoundedQueueAndPriority(t *testing.T) {
	cfg := academicServerConfig(t, 3, 3, 1)
	cfg.ControlCapacity = 1
	cfg.HighBurst = 2
	cfg.ControlBurst = 1
	s := bareServer(t, cfg)
	registerAll(t, s)
	w := &s.workers[0]
	w.stageStarts()
	w.flushControlPending()
	for n := 0; n < 20; n++ {
		w.flushControlPending()
	}
	if w.control.stats().Depth != 1 || w.pendingGroups.stats().Depth != 1 || w.control.stats().Drops != 0 || bits.OnesCount32(s.groups[0].ControlPendingMask) != 2 {
		t.Fatal(s.groups[0])
	}
	w.high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	w.high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	w.data.push(outbound{Packet: Packet{Header: Header{Type: POSITION}}}, false)
	s.writers[0].highStreak = 0 // registration draining consumed the previous HIGH quota
	for n := 0; n < 2; n++ {
		job, _, ctl, _, ok := s.takeServerOutbound(0)
		if !ok || ctl || job.Packet.Type != ACK {
			t.Fatal(job, ctl)
		}
	}
	_, ctl, isControl, _, ok := s.takeServerOutbound(0)
	if !ok || !isControl || ctl.Packet.Type != START {
		t.Fatal(ctl)
	}
	if !s.controlBucket.take(0) {
		t.Fatal("bucket fixture")
	}
	s.releaseControlLease() // emulate completion of the selected CONTROL write
	w.flushControlPending()
	w.high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	job, _, isControl, _, ok := s.takeServerOutbound(0)
	if !ok || isControl || job.Packet.Type != ACK {
		t.Fatal("reactive blocked by CONTROL bucket", job)
	}
	job, _, isControl, _, ok = s.takeServerOutbound(0)
	if !ok || isControl || job.Packet.Type != POSITION {
		t.Fatal("DATA starved", job)
	}
	_, _, _, wait, ok := s.takeServerOutbound(0)
	if ok || wait != int64(time.Millisecond) {
		t.Fatal(wait, ok)
	}
}

func TestServerControlConfigurationAndWriteFailure(t *testing.T) {
	for _, change := range []func(*ServerConfig){
		func(c *ServerConfig) { c.ControlCapacity = 0 }, func(c *ServerConfig) { c.ConfirmationCapacity = 0 }, func(c *ServerConfig) { c.ControlRate = 0 },
		func(c *ServerConfig) { c.ControlBurst = 0 }, func(c *ServerConfig) { c.ControlRate = 1_000_000_001 }, func(c *ServerConfig) { c.MaxControlSends = 0 },
		func(c *ServerConfig) { c.InitialControlRTO = 0 }, func(c *ServerConfig) { c.MaxControlRTO = c.InitialControlRTO - 1 },
		func(c *ServerConfig) { c.ControlJitterPercent = 51 }, func(c *ServerConfig) { c.ControlJitterPercent = -1 },
	} {
		cfg := academicServerConfig(t, 3, 3, 1)
		change(&cfg)
		if _, err := newServer(cfg); err != ErrServerConfig {
			t.Fatal(err)
		}
	}
	s := readyAcademicServer(t, 3, 3, 1)
	job := collectControl(t, s)[0]
	tx := serverControlConfirmation{Ticket: job.Ticket, GroupID: 1, TargetBit: job.TargetBit, Type: START, Failed: true}
	if err := s.workers[0].handleControlConfirmation(tx); err != nil {
		t.Fatal(err)
	}
	if s.Stats().ControlFailures != 1 || s.Stats().ControlRetryExhausted != 0 || s.Stats().SimulationComplete {
		t.Fatal(s.Stats())
	}
	if err := s.WaitSimulationComplete(context.Background()); err != ErrControlWrite {
		t.Fatal(err)
	}
	s.workers[0].finish()
	if s.controlHeapCurrent.Load() != 0 || s.controlGates[0].mask.Load() != 0 {
		t.Fatal("control cleanup")
	}
}

func TestServerControlStructureSizes(t *testing.T) {
	t.Logf("ServerClientState=%d GroupState=%d controlOutbound=%d controlConfirmation=%d worker=%d gate=%d snapshot=%d heap=%d",
		unsafe.Sizeof(ServerClientState{}), unsafe.Sizeof(GroupState{}), unsafe.Sizeof(serverControlOutbound{}), unsafe.Sizeof(serverControlConfirmation{}),
		unsafe.Sizeof(serverWorker{}), unsafe.Sizeof(serverControlGate{}), unsafe.Sizeof(serverGroupSnapshot{}), unsafe.Sizeof(serverControlDeadlineHeap{}))
}

func TestServerAcademicOlderRetryIsStaleWithoutMutatingCurrent(t *testing.T) {
	s := readyAcademicServer(t, 3, 3, 1)
	startAcademicGroups(t, s)
	p := positionPacket(s, 1)
	p.MessageID = 2
	p.Type = RETRY_POSITION
	p.RetryCounter = 2
	p.MissingMask = 6
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	before := s.clients[0]
	old := p
	old.RetryCounter = 1
	old.MissingMask = 2
	if err := deliver(t, s, old, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if s.clients[0] != before || s.Stats().StalePackets != 1 || s.Stats().InvalidPackets != 0 {
		t.Fatal(s.clients[0], before, s.Stats())
	}
}
