package protocol

import "context"

// The approved v1 controls use stable operation IDs within the RunID/group key.
// REGISTER/POSITION/METRICS have their own per-client operation identity.
const clientStartMessageID uint64 = 1
const clientGroupDoneMessageID uint64 = 2

// WaitAllDone waits for logical completion, not socket closure or delivery of
// the final GROUP_DONE_ACK. The caller keeps the generator alive until BOTH
// this condition and the server's WaitSimulationComplete have succeeded.
// One close-only signal serves every waiter and every client of this generator.
func (g *ClientGenerator) WaitAllDone(ctx context.Context) error {
	if ctx == nil {
		return ErrClientConfig
	}
	if !g.cfg.Academic {
		return ErrClientPhase
	}
	select {
	case <-g.academicDone:
		return g.academicResult()
	case <-ctx.Done():
		return ctx.Err()
	case <-g.done:
		if g.academicOutcome.Load() != 0 {
			return g.academicResult()
		}
		if g.runErr != nil {
			return g.runErr
		}
		return context.Canceled
	}
}

func (g *ClientGenerator) academicResult() error {
	switch g.academicOutcome.Load() {
	case 1:
		return nil
	case 2 + uint32(ClientRetryExhausted):
		return ErrRetryExhausted
	case 2 + uint32(ClientTimestampFailed):
		return ErrClientTimestamp
	default:
		return ErrClientWrite
	}
}

func (g *ClientGenerator) publishAcademicFailure(reason ClientFailure) {
	if g.cfg.Academic && g.academicOutcome.CompareAndSwap(0, 2+uint32(reason)) {
		close(g.academicDone)
	}
}

// Stateless SplitMix64 produces finite float32 coordinates in [-1000,1000).
// Each component uses the next fixed mixing step from RunID ^ seed ^ (id<<32)
// ^ iteration. Integer quantization to 1/1024 makes generation reproducible.
func automaticPosition(runID, seed uint64, id, iteration uint32) PositionPayload {
	s := runID ^ seed ^ uint64(id)<<32 ^ uint64(iteration)
	var xyz [3]float32
	for component := range xyz {
		s += 0x9e3779b97f4a7c15
		v := (s ^ (s >> 30)) * 0xbf58476d1ce4e5b9
		v = (v ^ (v >> 27)) * 0x94d049bb133111eb
		v ^= v >> 31
		xyz[component] = float32(int32(v%2_048_000)-1_024_000) / 1024
	}
	return PositionPayload{xyz[0], xyz[1], xyz[2]}
}

func (w *clientWorker) reactiveControl(p *Packet, typ MessageType) error {
	ack := Packet{Header: Header{Type: typ, RunID: p.RunID, GeneratorID: p.GeneratorID,
		GroupID: p.GroupID, OriginID: p.TargetID, MessageID: p.MessageID, RetryCounter: p.RetryCounter}}
	if !w.high.push(clientOutbound{Packet: ack}, true) {
		w.generator.counters.queueDrops.Add(1)
		return ErrQueueFull
	}
	notify(w.generator.writerWake)
	return nil
}

func (w *clientWorker) startAcademic(id uint32, c *ClientState, attempt uint16) error {
	g := w.generator
	if !c.Registered {
		return g.invalidPacket(ErrNotRegistered)
	}
	p := Packet{Header: Header{Type: START, RunID: g.cfg.Manifest.run.RunID, GeneratorID: g.cfg.GeneratorID,
		GroupID: 1 + (id-1)/g.cfg.Manifest.run.V, TargetID: id, MessageID: clientStartMessageID, RetryCounter: attempt}}
	c.Flags &^= clientPendingStart
	if c.Flags&clientAcademicStarted != 0 {
		g.counters.duplicateStarts.Add(1)
		return w.reactiveControl(&p, START_ACK)
	}
	if c.Phase != ClientWaitStart {
		return g.invalidPacket(ErrClientPhase)
	}
	c.Flags |= clientAcademicStarted
	c.Phase = ClientReady
	g.counters.startedClients.Add(1)
	// Admission can be deferred by the existing bounded pending index, DATA
	// queue, token bucket and inflight window. This happens strictly before t0.
	err := w.advanceAcademic(id, c)
	ackErr := w.reactiveControl(&p, START_ACK)
	if err != nil {
		return err
	}
	return ackErr
}

func (w *clientWorker) advanceAcademic(id uint32, c *ClientState) error {
	g := w.generator
	if c.CompletedIterations < g.cfg.Manifest.run.I {
		next := c.CompletedIterations + 1
		return w.beginPosition(clientCommand{id, next, automaticPosition(g.cfg.Manifest.run.RunID, g.cfg.JitterSeed, id, next)})
	}
	// No additional RTT sample or mutable metric storage: RTTSum/I is stable
	// after exactly I completed operations, including while helping neighbors.
	c.MessageID = c.RegistrationNonce + uint64(g.cfg.Manifest.run.I) + 1
	c.SendCount, c.RetryCounter = 0, 0
	c.CurrentRTO = int64(g.cfg.InitialRTO)
	c.Flags &^= clientAwaitConfirmation | clientRetryIntent
	c.Flags |= clientMetricsIntent
	c.Phase = ClientMetricsQueued
	w.nextSend(id, c)
	w.schedule(id, c)
	return nil
}

func (w *clientWorker) metricsJob(id uint32, c *ClientState) clientOutbound {
	g := w.generator
	return clientOutbound{Packet: Packet{Header: Header{Type: METRICS, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, GroupID: 1 + (id-1)/g.cfg.Manifest.run.V, OriginID: id, MessageID: c.MessageID},
		Metrics: MetricsPayload{MeanRTTNS: c.RTTSum / uint64(g.cfg.Manifest.run.I), CompletedIterations: g.cfg.Manifest.run.I}},
		ClientID: id, Ticket: c.SendTicket, Attempt: c.RetryCounter}
}

func (w *clientWorker) acceptMetrics(id uint32, c *ClientState) {
	if c.Flags&clientMetricsAccepted != 0 {
		return
	}
	c.Flags |= clientMetricsAccepted
	w.removeDeadline(id, c)
	w.cancelSend(id, c)
	c.Flags &^= clientAwaitConfirmation
	w.generator.counters.metricsAccepted.Add(1)
}

func (w *clientWorker) handleAcademicControl(p *Packet, c *ClientState) error {
	g := w.generator
	switch p.Type {
	case START:
		if p.MessageID != clientStartMessageID {
			return g.invalidPacket(ErrMessage)
		}
		if !c.Registered {
			if c.Phase != ClientNew && c.Phase != ClientRegisterQueued && c.Phase != ClientRegisterWait {
				return g.invalidPacket(ErrClientPhase)
			}
			// Never ACK before authoritative registration. The retained START is
			// processed only after a valid REGISTER_ACK; REGISTER ARQ continues.
			c.Flags |= clientPendingStart
			c.PendingStartAttempt = p.RetryCounter
			return nil
		}
		return w.startAcademic(p.TargetID, c, p.RetryCounter)
	case METRICS_ACK:
		if !c.Registered || c.Flags&clientAcademicStarted == 0 {
			return g.invalidPacket(ErrNotRegistered)
		}
		if p.MessageID != c.RegistrationNonce+uint64(g.cfg.Manifest.run.I)+1 {
			return g.invalidPacket(ErrMessage)
		}
		if c.Flags&clientMetricsAccepted != 0 {
			g.counters.duplicateMetricsAcks.Add(1)
			return nil
		}
		if c.Phase != ClientMetricsQueued && c.Phase != ClientMetricsWait {
			return g.invalidPacket(ErrClientPhase)
		}
		w.acceptMetrics(p.TargetID, c)
		c.Phase = ClientWaitGroupDone
		return nil
	case GROUP_DONE:
		if !c.Registered || c.Flags&clientAcademicStarted == 0 {
			return g.invalidPacket(ErrNotRegistered)
		}
		if p.MessageID != clientGroupDoneMessageID {
			return g.invalidPacket(ErrMessage)
		}
		if c.Phase == ClientDone {
			g.counters.duplicateGroupDone.Add(1)
			return w.reactiveControl(p, GROUP_DONE_ACK)
		}
		if c.CompletedIterations != g.cfg.Manifest.run.I || c.Phase != ClientMetricsQueued && c.Phase != ClientMetricsWait && c.Phase != ClientWaitGroupDone {
			return g.invalidPacket(ErrClientPhase)
		}
		// Group completion proves acceptance even when METRICS_ACK was lost.
		w.acceptMetrics(p.TargetID, c)
		c.Phase = ClientDone
		count := g.counters.doneClients.Add(1)
		ackErr := w.reactiveControl(p, GROUP_DONE_ACK)
		if count == uint64(len(g.clients)) && g.academicOutcome.CompareAndSwap(0, 1) {
			close(g.academicDone)
		}
		return ackErr
	default:
		return g.invalidPacket(ErrType)
	}
}

func (w *clientWorker) handleMetricsConfirmation(tx clientConfirmation, c *ClientState) error {
	g := w.generator
	invalid := func(err error) error { g.counters.invalidConfirmations.Add(1); return err }
	if !g.cfg.Academic {
		return invalid(ErrType)
	}
	if tx.MessageID != c.RegistrationNonce+uint64(g.cfg.Manifest.run.I)+1 || tx.Iteration != 0 || tx.Ticket == 0 || tx.Ticket > c.SendTicket || tx.Attempt >= g.cfg.MaxSends {
		return invalid(ErrMessage)
	}
	if tx.Status == clientSendOK && tx.SentAt < 0 {
		return invalid(ErrClientTimestamp)
	}
	if c.Flags&clientMetricsAccepted != 0 || tx.Ticket < c.SendTicket {
		g.counters.staleConfirmations.Add(1)
		return nil
	}
	if c.Phase != ClientMetricsQueued && c.Phase != ClientMetricsWait {
		return invalid(ErrClientPhase)
	}
	if tx.Attempt != c.RetryCounter {
		return invalid(ErrMessage)
	}
	if c.Flags&clientAwaitConfirmation == 0 {
		g.counters.duplicateConfirmations.Add(1)
		return nil
	}
	c.Flags &^= clientAwaitConfirmation
	c.SendCount++
	if tx.Status == clientSendFailed {
		w.failClient(tx.ClientID, c, ClientWriteFailed)
	} else {
		c.Phase = ClientMetricsWait
		w.armDeadline(tx.ClientID, c, tx.SentAt)
	}
	g.counters.confirmations.Add(1)
	return nil
}
