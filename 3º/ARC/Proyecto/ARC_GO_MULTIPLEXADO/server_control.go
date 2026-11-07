package protocol

import (
	"context"
	"errors"
	"math"
	"math/bits"
	"sync/atomic"
	"time"
)

var (
	ErrControlRetryExhausted = errors.New("server: reliable control retry exhausted")
	ErrSimulationIncomplete  = errors.New("server: simulation incomplete")
	ErrControlWrite          = errors.New("server: reliable control write failed")
)

// These IDs name operations within a run/group, rather than datagrams/attempts.
const serverStartMessageID uint64 = 1
const serverGroupDoneMessageID uint64 = 2

type ServerGroupPhase uint8

const (
	ServerGroupIdle ServerGroupPhase = iota
	ServerGroupStarting
	ServerGroupRunning
	ServerGroupFinishing
	ServerGroupComplete
	ServerGroupFailed
)

type serverControlOutbound struct {
	Packet             Packet
	Ticket             uint64
	GeneratorIndex     int
	GroupID, TargetBit uint32
	Attempt            uint16
}
type serverControlConfirmation struct {
	Ticket             uint64
	SentAt             int64
	GroupID, TargetBit uint32
	Attempt            uint16
	Type               MessageType
	Failed             bool
}

// Writer uses only these immutable-job/atomic gates; it never reads GroupState.
// Cancellation is best effort: a write already selected may produce one harmless
// duplicate, but there is never another copy of a target in the same round.
type serverControlGate struct {
	ticket atomic.Uint64
	mask   atomic.Uint32
}

// Snapshots are published by the group owner. The sequence protects the 128-bit
// sum and the masks from tearing, without a global lock or a pointer allocation.
type serverGroupSnapshot struct {
	sequence                                  atomic.Uint64
	sumLo, sumHi                              atomic.Uint64
	registered, started, metrics, done, phase atomic.Uint32
}
type ServerGroupStats struct {
	GroupID, V                                                  uint32
	RegisteredMask, StartAckMask, MetricsMask, GroupDoneAckMask uint32
	MetricsClients                                              uint32
	MeanRTTNS                                                   uint64
	Phase                                                       ServerGroupPhase
	SumLo, SumHi                                                uint64
}

func (w *serverWorker) publishGroup(index uint32) {
	g := &w.server.groups[index]
	p := &w.server.groupSnapshots[index]
	p.sequence.Add(1)
	p.sumLo.Store(g.MetricsSumLo)
	p.sumHi.Store(g.MetricsSumHi)
	p.registered.Store(g.RegisteredMask)
	p.started.Store(g.StartAckMask)
	p.metrics.Store(g.MetricsMask)
	p.done.Store(g.GroupDoneAckMask)
	p.phase.Store(uint32(g.ControlPhase))
	p.sequence.Add(1)
}
func (s *Server) readGroup(index int) ServerGroupStats {
	p := &s.groupSnapshots[index]
	for {
		seq := p.sequence.Load()
		if seq&1 != 0 {
			continue
		}
		st := ServerGroupStats{GroupID: uint32(index) + 1, V: s.cfg.Manifest.run.V,
			RegisteredMask: p.registered.Load(), StartAckMask: p.started.Load(), MetricsMask: p.metrics.Load(), GroupDoneAckMask: p.done.Load(),
			SumLo: p.sumLo.Load(), SumHi: p.sumHi.Load(), Phase: ServerGroupPhase(p.phase.Load())}
		if p.sequence.Load() != seq {
			continue
		}
		st.MetricsClients = uint32(bits.OnesCount32(st.MetricsMask))
		// A partial mean is labelled by MetricsClients, never divided by all V.
		if st.MetricsClients != 0 {
			st.MeanRTTNS, _ = bits.Div64(st.SumHi, st.SumLo, uint64(st.MetricsClients))
		}
		return st
	}
}
func (s *Server) GroupStats() []ServerGroupStats {
	out := make([]ServerGroupStats, len(s.groups))
	for i := range out {
		out[i] = s.readGroup(i)
	}
	return out
}
func (s *Server) now() int64 { return time.Since(s.epoch).Nanoseconds() }
func (s *Server) WaitSimulationComplete(ctx context.Context) error {
	if ctx == nil {
		return ErrServerConfig
	}
	if !s.cfg.Academic {
		return ErrSimulationIncomplete
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.simulationDone:
		if s.simulationOutcome.Load() == 1 {
			return nil
		}
		if s.counters.controlRetryExhausted.Load() != 0 {
			return ErrControlRetryExhausted
		}
		return ErrControlWrite
	case <-s.done:
		if s.simulationOutcome.Load() == 1 {
			return nil
		}
		if s.simulationOutcome.Load() == 2 {
			if s.counters.controlRetryExhausted.Load() != 0 {
				return ErrControlRetryExhausted
			}
			return ErrControlWrite
		}
		if s.runErr != nil {
			return s.runErr
		}
		return ErrSimulationIncomplete
	}
}
func (s *Server) markControlFailure(exhausted bool) {
	s.counters.controlFailures.Add(1)
	if exhausted {
		s.counters.controlRetryExhausted.Add(1)
	}
	if s.simulationOutcome.CompareAndSwap(0, 2) {
		close(s.simulationDone)
	}
}
func (s *Server) checkSimulationComplete() {
	n := uint64(s.cfg.Manifest.run.N)
	if s.counters.registered.Load() != n || s.counters.startedClients.Load() != n || s.counters.metricsAccepted.Load() != n ||
		s.counters.doneClients.Load() != n || s.counters.completedGroups.Load() != uint64(len(s.groups)) || s.counters.controlFailures.Load() != 0 {
		return
	}
	if s.simulationOutcome.CompareAndSwap(0, 1) {
		s.simulationEndNS.Store(s.now())
		close(s.simulationDone)
	}
}

// Fixed, indexed heap: one entry per owner group, no obsolete event history.
type serverControlDeadlineHeap struct {
	groups        []GroupState
	indices       []uint32
	size, maximum int
}

func newServerControlDeadlineHeap(groups []GroupState, capacity int) serverControlDeadlineHeap {
	return serverControlDeadlineHeap{groups: groups, indices: make([]uint32, capacity)}
}
func (h *serverControlDeadlineHeap) len() int { return h.size }
func (h *serverControlDeadlineHeap) peek() (uint32, int64, bool) {
	if h.size == 0 {
		return 0, 0, false
	}
	i := h.indices[0]
	return i, h.groups[i].ControlDeadline, true
}
func (h *serverControlDeadlineHeap) less(a, b int) bool {
	ai, bi := h.indices[a], h.indices[b]
	ad, bd := h.groups[ai].ControlDeadline, h.groups[bi].ControlDeadline
	return ad < bd || ad == bd && ai < bi
}
func (h *serverControlDeadlineHeap) swap(a, b int) {
	h.indices[a], h.indices[b] = h.indices[b], h.indices[a]
	h.groups[h.indices[a]].ControlHeapIndex = int32(a)
	h.groups[h.indices[b]].ControlHeapIndex = int32(b)
}
func (h *serverControlDeadlineHeap) up(p int) int {
	for p > 0 {
		parent := (p - 1) / 2
		if !h.less(p, parent) {
			break
		}
		h.swap(p, parent)
		p = parent
	}
	return p
}
func (h *serverControlDeadlineHeap) down(p int) {
	for p < h.size/2 {
		c := p*2 + 1
		if c+1 < h.size && h.less(c+1, c) {
			c++
		}
		if !h.less(c, p) {
			return
		}
		h.swap(p, c)
		p = c
	}
}
func (h *serverControlDeadlineHeap) set(i uint32, deadline int64) bool {
	if uint64(i) >= uint64(len(h.groups)) || deadline < 0 {
		return false
	}
	g := &h.groups[i]
	p := int(g.ControlHeapIndex)
	if p == -1 {
		if h.size == len(h.indices) || h.size > math.MaxInt32 {
			return false
		}
		p = h.size
		h.indices[p] = i
		g.ControlHeapIndex = int32(p)
		g.ControlDeadline = deadline
		h.size++
		if h.size > h.maximum {
			h.maximum = h.size
		}
		h.up(p)
		return true
	}
	if p < 0 || p >= h.size || h.indices[p] != i {
		return false
	}
	old := g.ControlDeadline
	g.ControlDeadline = deadline
	if deadline < old {
		h.up(p)
	} else if deadline > old {
		h.down(p)
	}
	return true
}
func (h *serverControlDeadlineHeap) remove(i uint32) bool {
	if uint64(i) >= uint64(len(h.groups)) {
		return false
	}
	g := &h.groups[i]
	p := int(g.ControlHeapIndex)
	if p < 0 || p >= h.size || h.indices[p] != i {
		return false
	}
	h.size--
	if p != h.size {
		h.indices[p] = h.indices[h.size]
		h.groups[h.indices[p]].ControlHeapIndex = int32(p)
	}
	h.indices[h.size] = 0
	g.ControlHeapIndex = -1
	g.ControlDeadline = -1
	if p < h.size {
		if h.up(p) == p {
			h.down(p)
		}
	}
	return true
}
func (w *serverWorker) removeControlDeadline(index uint32) {
	if w.controlDeadlines.remove(index) {
		w.server.controlHeapCurrent.Add(-1)
	}
	w.server.groups[index].ControlDeadline = -1
}
func (w *serverWorker) armControlDeadline(index uint32) {
	s := w.server
	g := &s.groups[index]
	seed := s.cfg.ControlJitterSeed ^ s.cfg.Manifest.run.RunID ^ uint64(index+1)<<32 ^ g.ControlTicket ^ uint64(w.index)<<48
	delay := clientJitterRTO(g.ControlRTO, s.cfg.ControlJitterPercent, seed)
	deadline := int64(math.MaxInt64)
	if g.ControlLastSentAt <= math.MaxInt64-delay {
		deadline = g.ControlLastSentAt + delay
	}
	existing := g.ControlHeapIndex >= 0
	if !w.controlDeadlines.set(index, deadline) {
		s.reportFatal(ErrProgress)
		return
	}
	if !existing {
		depth := s.controlHeapCurrent.Add(1)
		for {
			old := s.controlHeapMaximum.Load()
			if depth <= old || s.controlHeapMaximum.CompareAndSwap(old, depth) {
				break
			}
		}
	}
}
func (w *serverWorker) groupMask() uint32 { m, _ := GroupMask(w.server.cfg.Manifest.run.V); return m }
func (g *GroupState) controlACKMask() uint32 {
	if g.ControlPhase == ServerGroupStarting {
		return g.StartAckMask
	}
	return g.GroupDoneAckMask
}
func (g *GroupState) controlType() MessageType {
	if g.ControlPhase == ServerGroupStarting {
		return START
	}
	return GROUP_DONE
}
func (w *serverWorker) scheduleControl(index uint32) {
	g := &w.server.groups[index]
	if g.ControlScheduled || g.ControlPendingMask == 0 {
		return
	}
	if !w.pendingGroups.push(index, false) {
		w.server.reportFatal(ErrProgress)
		return
	}
	g.ControlScheduled = true
}
func (w *serverWorker) beginControlRound(index uint32) {
	s := w.server
	g := &s.groups[index]
	g.ControlTicket++
	g.ControlRoundMask = w.groupMask() &^ g.controlACKMask()
	g.ControlPendingMask = g.ControlRoundMask
	g.ControlAwaitMask = 0
	g.ControlWrittenMask = 0
	g.ControlLastSentAt = -1
	s.controlGates[index].ticket.Store(g.ControlTicket)
	s.controlGates[index].mask.Store(g.ControlRoundMask)
	w.scheduleControl(index)
}
func (w *serverWorker) beginGroupControl(index uint32, phase ServerGroupPhase) {
	g := &w.server.groups[index]
	w.removeControlDeadline(index)
	g.ControlPhase = phase
	g.ControlAttempt = 0
	g.ControlSends = 0
	g.ControlRTO = int64(w.server.cfg.InitialControlRTO)
	if phase == ServerGroupStarting {
		w.server.counters.startedGroups.Add(1)
	}
	w.beginControlRound(index)
	w.publishGroup(index)
}

// Global admission to START is discovered by each owner from the atomic barrier.
// Scanning is budgeted, as is subsequent fanout, so no worker creates N packets.
func (w *serverWorker) stageStarts() bool {
	s := w.server
	if !s.cfg.Academic || s.counters.registered.Load() != uint64(s.cfg.Manifest.run.N) {
		return false
	}
	progressed := false
	for n := 0; n < s.cfg.WorkBudget && w.startCursor < len(s.groups); n++ {
		i := w.startCursor
		w.startCursor += len(s.workers)
		if s.groups[i].ControlPhase == ServerGroupIdle {
			w.beginGroupControl(uint32(i), ServerGroupStarting)
			progressed = true
		}
	}
	return progressed
}
func (w *serverWorker) cancelGroupRound(index uint32) {
	g := &w.server.groups[index]
	w.removeControlDeadline(index)
	g.ControlPendingMask = 0
	g.ControlAwaitMask = 0
	w.server.controlGates[index].mask.Store(0)
	w.notifyWriter()
}
func (w *serverWorker) completeGroupControl(index uint32) {
	s := w.server
	g := &s.groups[index]
	if g.controlACKMask() != w.groupMask() {
		return
	}
	w.cancelGroupRound(index)
	switch g.ControlPhase {
	case ServerGroupStarting:
		g.ControlPhase = ServerGroupRunning
		if g.MetricsMask == w.groupMask() {
			w.beginGroupControl(index, ServerGroupFinishing)
		}
	case ServerGroupFinishing:
		g.ControlPhase = ServerGroupComplete
		s.counters.completedGroups.Add(1)
	}
	w.publishGroup(index)
	s.checkSimulationComplete()
}
func (w *serverWorker) failGroupControl(index uint32, exhausted bool) {
	g := &w.server.groups[index]
	if g.ControlPhase == ServerGroupFailed || g.ControlPhase == ServerGroupComplete {
		return
	}
	w.cancelGroupRound(index)
	g.ControlPhase = ServerGroupFailed
	w.publishGroup(index)
	w.server.markControlFailure(exhausted)
}
func (w *serverWorker) handleControlACK(p *Packet) error {
	s := w.server
	index := p.GroupID - 1
	g := &s.groups[index]
	bit := uint32(1) << ((p.OriginID - 1) % s.cfg.Manifest.run.V)
	switch p.Type {
	case START_ACK:
		if p.MessageID != serverStartMessageID {
			return s.invalid(ErrMessage)
		}
		if g.ControlPhase == ServerGroupIdle {
			return s.invalid(ErrProgress)
		}
		if g.StartAckMask&bit != 0 {
			s.counters.duplicateStartACKs.Add(1)
			return nil
		}
		// A late START_ACK remains useful while START is active; retries do not change identity.
		if g.ControlPhase != ServerGroupStarting {
			return s.invalid(ErrProgress)
		}
		g.StartAckMask |= bit
		s.counters.startedClients.Add(1)
	case GROUP_DONE_ACK:
		if p.MessageID != serverGroupDoneMessageID {
			return s.invalid(ErrMessage)
		}
		if g.GroupDoneAckMask&bit != 0 {
			s.counters.duplicateGroupDoneACKs.Add(1)
			return nil
		}
		if g.ControlPhase != ServerGroupFinishing || g.MetricsMask != w.groupMask() {
			return s.invalid(ErrProgress)
		}
		g.GroupDoneAckMask |= bit
		s.counters.doneClients.Add(1)
	default:
		return s.invalid(ErrType)
	}
	g.ControlPendingMask &^= bit
	g.ControlAwaitMask &^= bit
	s.controlGates[index].mask.And(^bit)
	w.publishGroup(index)
	if g.controlACKMask() == w.groupMask() {
		w.completeGroupControl(index)
	} else {
		w.finishControlWrites(index)
	}
	return nil
}
func (w *serverWorker) handleMetrics(p *Packet, c *ServerClientState) error {
	s := w.server
	index := p.GroupID - 1
	g := &s.groups[index]
	gi := int(g.GeneratorIndex)
	gen := s.cfg.Manifest.generators[gi]
	nonce := uint64(p.OriginID-gen.FirstID) + 1
	if p.MessageID != nonce+uint64(s.cfg.Manifest.run.I)+1 {
		return s.invalid(ErrMessage)
	}
	if p.Metrics.CompletedIterations != s.cfg.Manifest.run.I {
		return s.invalid(ErrIteration)
	}
	if c.Flags&metricsAccepted != 0 {
		if c.MeanRTTNS != p.Metrics.MeanRTTNS {
			return s.invalid(ErrConflict)
		}
		s.counters.duplicateMetrics.Add(1)
	} else {
		if c.Iteration != s.cfg.Manifest.run.I || c.SeenACKMask != c.ExpectedMask || g.FinalACKMask&(uint32(1)<<((p.OriginID-1)%s.cfg.Manifest.run.V)) == 0 {
			return s.invalid(ErrProgress)
		}
		if g.ControlPhase == ServerGroupIdle || g.ControlPhase == ServerGroupFailed {
			return s.invalid(ErrProgress)
		}
		c.MeanRTTNS = p.Metrics.MeanRTTNS
		c.Flags |= metricsAccepted
		g.MetricsMask |= uint32(1) << ((p.OriginID - 1) % s.cfg.Manifest.run.V)
		var carry uint64
		g.MetricsSumLo, carry = bits.Add64(g.MetricsSumLo, c.MeanRTTNS, 0)
		g.MetricsSumHi, _ = bits.Add64(g.MetricsSumHi, 0, carry)
		s.counters.metricsAccepted.Add(1)
		w.publishGroup(index)
	}
	c.Flags |= metricsReplyPending
	w.schedule(p.OriginID, c)
	if g.MetricsMask == w.groupMask() && g.ControlPhase == ServerGroupRunning {
		w.beginGroupControl(index, ServerGroupFinishing)
	}
	return nil
}
func (w *serverWorker) metricsReply(id uint32, c *ServerClientState) outbound {
	s := w.server
	v := s.cfg.Manifest.run.V
	index := (id - 1) / v
	gi := int(s.groups[index].GeneratorIndex)
	gen := s.cfg.Manifest.generators[gi]
	return outbound{Packet: Packet{Header: Header{Type: METRICS_ACK, Flags: ServerToClient, RunID: s.cfg.Manifest.run.RunID, GeneratorID: gen.GeneratorID,
		GroupID: index + 1, TargetID: id, MessageID: uint64(id-gen.FirstID) + uint64(s.cfg.Manifest.run.I) + 2}}, GeneratorIndex: gi}
}
func (w *serverWorker) controlJob(index, bit uint32) serverControlOutbound {
	s := w.server
	g := &s.groups[index]
	gi := int(g.GeneratorIndex)
	target := index*s.cfg.Manifest.run.V + uint32(bits.TrailingZeros32(bit)) + 1
	typ, msg := g.controlType(), serverStartMessageID
	if typ == GROUP_DONE {
		msg = serverGroupDoneMessageID
	}
	return serverControlOutbound{Packet: Packet{Header: Header{Type: typ, Flags: ServerToClient, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: s.cfg.Manifest.generators[gi].GeneratorID, GroupID: index + 1, TargetID: target, MessageID: msg, RetryCounter: g.ControlAttempt}},
		Ticket: g.ControlTicket, GeneratorIndex: gi, GroupID: index + 1, TargetBit: bit, Attempt: g.ControlAttempt}
}
func (w *serverWorker) flushControlPending() bool {
	count := w.pendingGroups.stats().Depth
	budget := w.server.cfg.WorkBudget
	progress := false
	for turn := 0; turn < count && budget > 0; turn++ {
		index, ok := w.pendingGroups.pop()
		if !ok {
			break
		}
		g := &w.server.groups[index]
		g.ControlScheduled = false
		quota := 8
		if quota > budget {
			quota = budget
		}
		for g.ControlPendingMask != 0 && quota > 0 {
			bit := uint32(1) << bits.TrailingZeros32(g.ControlPendingMask)
			budget--
			quota--
			if !w.control.push(w.controlJob(index, bit), false) {
				break
			}
			g.ControlPendingMask &^= bit
			g.ControlAwaitMask |= bit
			progress = true
			w.notifyWriter()
		}
		w.scheduleControl(index)
	}
	return progress
}
func (w *serverWorker) finishControlWrites(index uint32) {
	g := &w.server.groups[index]
	if (g.ControlPhase != ServerGroupStarting && g.ControlPhase != ServerGroupFinishing) || g.ControlPendingMask != 0 || g.ControlAwaitMask != 0 || g.controlACKMask() == w.groupMask() || g.ControlHeapIndex >= 0 {
		return
	}
	// Only a successful Write can start a network RTO. ACK-before-confirmation may
	// cancel some targets, but at least one still-pending target has been written.
	if g.ControlWrittenMask&^g.controlACKMask() == 0 || g.ControlLastSentAt < 0 {
		return
	}
	g.ControlSends = g.ControlAttempt + 1
	w.armControlDeadline(index)
}
func (w *serverWorker) handleControlConfirmation(tx serverControlConfirmation) error {
	s := w.server
	owner, err := s.WorkerForGroup(tx.GroupID)
	if err != nil || owner != w.index {
		return ErrIdentity
	}
	index := tx.GroupID - 1
	g := &s.groups[index]
	if tx.Ticket < g.ControlTicket {
		s.counters.staleControlConfirmations.Add(1)
		return nil
	}
	if tx.Ticket != g.ControlTicket || tx.Attempt != g.ControlAttempt || tx.TargetBit == 0 || tx.TargetBit&(tx.TargetBit-1) != 0 || tx.TargetBit&^w.groupMask() != 0 {
		return ErrMessage
	}
	if g.ControlPhase != ServerGroupStarting && g.ControlPhase != ServerGroupFinishing {
		s.counters.staleControlConfirmations.Add(1)
		return nil
	}
	if tx.Type != g.controlType() {
		return ErrType
	}
	if !tx.Failed && tx.SentAt < 0 {
		return ErrProgress
	}
	if g.ControlAwaitMask&tx.TargetBit == 0 {
		s.counters.staleControlConfirmations.Add(1)
		return nil
	}
	g.ControlAwaitMask &^= tx.TargetBit
	if tx.Failed {
		w.failGroupControl(index, false)
		return nil
	}
	g.ControlWrittenMask |= tx.TargetBit
	if tx.SentAt > g.ControlLastSentAt {
		g.ControlLastSentAt = tx.SentAt
	}
	w.finishControlWrites(index)
	return nil
}
func (w *serverWorker) processControlTimeouts(now int64) bool {
	progress := false
	s := w.server
	for n := 0; n < s.cfg.WorkBudget; n++ {
		index, deadline, ok := w.controlDeadlines.peek()
		if !ok || deadline > now {
			break
		}
		g := &s.groups[index]
		w.removeControlDeadline(index)
		progress = true
		if g.ControlPhase != ServerGroupStarting && g.ControlPhase != ServerGroupFinishing {
			continue
		}
		if g.controlACKMask() == w.groupMask() {
			w.completeGroupControl(index)
			continue
		}
		s.counters.controlTimeouts.Add(1)
		if g.ControlSends >= s.cfg.MaxControlSends {
			w.failGroupControl(index, true)
			continue
		}
		g.ControlRTO = nextClientRTO(g.ControlRTO, int64(s.cfg.MaxControlRTO))
		g.ControlAttempt++
		w.beginControlRound(index)
	}
	return progress
}
func (s *Server) controlCancelled(job serverControlOutbound) bool {
	if job.GroupID == 0 || uint64(job.GroupID) > uint64(len(s.controlGates)) {
		return true
	}
	gate := &s.controlGates[job.GroupID-1]
	return gate.ticket.Load() != job.Ticket || gate.mask.Load()&job.TargetBit == 0
}
func (wr *serverWriter) takeControlOutbound(now int64) (serverControlOutbound, int64, bool) {
	s := wr.server
	if !s.controlMu.TryLock() {
		// Register before retrying: release may occur between the failed
		// probe and registration. The retry closes that lost-wakeup window.
		s.controlWaiters.Or(uint32(1) << wr.index)
		if !s.controlMu.TryLock() {
			return serverControlOutbound{}, 0, false
		}
	}
	s.controlWaiters.And(^(uint32(1) << wr.index))
	leased := false
	defer func() {
		if !leased {
			s.releaseControlLease()
		}
	}()
	wait := int64(0)
	for n := 0; n < len(wr.workers); n++ {
		i := (wr.controlCursor + n) % len(wr.workers)
		w := &s.workers[wr.workers[i]]
		for discarded := 0; discarded < s.cfg.WorkBudget; discarded++ {
			head, ok := w.control.peek()
			if !ok || !s.controlCancelled(head) {
				break
			}
			w.control.pop()
			s.counters.cancelledControls.Add(1)
			notify(w.wake)
		}
		head, ok := w.control.peek()
		if !ok {
			continue
		}
		if s.controlCancelled(head) { // More cancelled heads than this pass's budget.
			return serverControlOutbound{}, now, true
		}
		q := w.confirmations.stats()
		if q.Depth >= q.Capacity {
			continue
		}
		at := s.controlBucket.ready(now)
		if at > now {
			wait = at
			continue
		}
		job, ok := w.control.pop()
		if ok {
			wr.controlCursor = (i + 1) % len(wr.workers)
			leased = true // writeOutbound releases this aggregate CONTROL lease
			notify(w.wake)
			return job, wait, true
		}
	}
	return serverControlOutbound{}, wait, false
}

// Eight HIGH then one lower-class turn; alternate eligible CONTROL/DATA turns.
// CONTROL has one aggregate bucket. A paced CONTROL never blocks reactive HIGH.
func (wr *serverWriter) takeServerOutbound(now int64) (outbound, serverControlOutbound, bool, int64, bool) {
	wait := int64(0)
	lower := func() (outbound, serverControlOutbound, bool, bool) {
		for attempt := 0; attempt < 2; attempt++ {
			useControl := wr.lowerControlTurn
			wr.lowerControlTurn = !wr.lowerControlTurn
			if useControl {
				job, at, ok := wr.takeControlOutbound(now)
				if at > 0 {
					wait = at
				}
				if ok {
					return outbound{}, job, true, true
				}
			} else if job, ok := wr.popClass(false); ok {
				return job, serverControlOutbound{}, false, true
			}
		}
		return outbound{}, serverControlOutbound{}, false, false
	}
	if wr.highStreak >= wr.server.cfg.HighBurst {
		if normal, control, isControl, ok := lower(); ok {
			wr.highStreak = 0
			return normal, control, isControl, wait, true
		}
	}
	if job, ok := wr.popClass(true); ok {
		if wr.highStreak < wr.server.cfg.HighBurst {
			wr.highStreak++
		}
		return job, serverControlOutbound{}, false, wait, true
	}
	if normal, control, isControl, ok := lower(); ok {
		wr.highStreak = 0
		return normal, control, isControl, wait, true
	}
	return outbound{}, serverControlOutbound{}, false, wait, false
}
func stopServerTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func (s *Server) publishControlConfirmation(job serverControlOutbound, sentAt int64, failed bool) bool {
	owner, _ := s.WorkerForGroup(job.GroupID)
	tx := serverControlConfirmation{Ticket: job.Ticket, SentAt: sentAt, GroupID: job.GroupID, TargetBit: job.TargetBit, Attempt: job.Attempt, Type: job.Packet.Type, Failed: failed}
	if !s.workers[owner].confirmations.push(tx, false) {
		s.reportFatal(ErrProgress)
		return false
	}
	notify(s.workers[owner].wake)
	return true
}
