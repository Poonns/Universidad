package protocol

import "math"

// Existing retries take precedence even while waiting for their own budget.
// The single writer samples bounded queues; it never reads mutable client state.
func (g *ClientGenerator) hasRetryBacklog() bool {
	for i := range g.workers {
		if g.workers[i].retry.stats().Depth > 0 {
			return true
		}
	}
	return false
}

func nextClientRTO(current, maximum int64) int64 {
	if current >= maximum/2+maximum%2 {
		return maximum
	}
	return current * 2
}

// Stateless SplitMix64 jitter: no global random source, lock or per-client RNG.
// The nominal RTO follows backoff; actual wait is in [(100-p)% RTO, RTO].
func clientJitterRTO(rto int64, percent int, seed uint64) int64 {
	span := rto / 100 * int64(percent)
	seed += 0x9e3779b97f4a7c15
	seed = (seed ^ (seed >> 30)) * 0xbf58476d1ce4e5b9
	seed = (seed ^ (seed >> 27)) * 0x94d049bb133111eb
	seed ^= seed >> 31
	return rto - span + int64(seed%uint64(span+1))
}

func (w *clientWorker) nextSend(id uint32, c *ClientState) {
	// At most MaxSends per REGISTER/iteration: uint32 I * uint16 attempts fits
	// below the reserved claim bit. Tickets are never reset between operations.
	c.SendTicket++
	index, _ := w.generator.ClientIndex(id)
	w.generator.sendGates[index].Store(c.SendTicket)
}

func (g *ClientGenerator) claimOutbound(job clientOutbound) bool {
	if job.Ticket == 0 { // synthetic selector fixtures only; real jobs always have a ticket
		return true
	}
	index, err := g.ClientIndex(job.ClientID)
	return err == nil && g.sendGates[index].CompareAndSwap(job.Ticket, job.Ticket|clientTicketClaimed)
}

func (g *ClientGenerator) outboundCancelled(job clientOutbound) bool {
	if job.Ticket == 0 {
		return false
	}
	index, err := g.ClientIndex(job.ClientID)
	return err != nil || g.sendGates[index].Load() != job.Ticket
}

func (w *clientWorker) cancelSend(id uint32, c *ClientState) {
	index, _ := w.generator.ClientIndex(id)
	w.generator.sendGates[index].Store(0)
	c.Flags &^= clientRegisterIntent | clientPositionIntent | clientRetryIntent | clientMetricsIntent
	notify(w.generator.writerWake)
}

func (w *clientWorker) armDeadline(id uint32, c *ClientState, sentAt int64) {
	g := w.generator
	index, _ := g.ClientIndex(id)
	seed := g.cfg.JitterSeed ^ g.cfg.Manifest.run.RunID ^ uint64(id)<<32 ^ c.SendTicket ^ uint64(w.index)<<48
	delay := clientJitterRTO(c.CurrentRTO, g.cfg.JitterPercent, seed)
	deadline := int64(math.MaxInt64)
	if sentAt <= math.MaxInt64-delay {
		deadline = sentAt + delay
	}
	wasActive := c.HeapIndex >= 0
	if !w.deadlines.set(uint32(index), deadline) {
		g.reportFatal(ErrClientPhase)
		return
	}
	if !wasActive {
		depth := g.heapCurrent.Add(1)
		for {
			old := g.heapMaximum.Load()
			if depth <= old || g.heapMaximum.CompareAndSwap(old, depth) {
				break
			}
		}
	}
}

func (w *clientWorker) removeDeadline(id uint32, c *ClientState) {
	index, _ := w.generator.ClientIndex(id)
	if w.deadlines.remove(uint32(index)) {
		w.generator.heapCurrent.Add(-1)
	}
	c.Deadline = -1
}

func (g *ClientGenerator) admitPosition() {
	depth := g.positionInFlight.Add(1)
	g.counters.admitted.Add(1)
	for {
		old := g.positionMaximum.Load()
		if depth <= old || g.positionMaximum.CompareAndSwap(old, depth) {
			break
		}
	}
}

func (w *clientWorker) releasePositionSlot(c *ClientState) {
	if c.Flags&clientPositionTxKnown != 0 && c.Flags&clientPositionSlotReleased == 0 {
		c.Flags |= clientPositionSlotReleased
		w.generator.positionInFlight.Add(-1)
		notify(w.generator.writerWake)
	}
}

func (w *clientWorker) failClient(id uint32, c *ClientState, reason ClientFailure) {
	g := w.generator
	if c.Phase == ClientFailed || c.Phase == ClientDone {
		return
	}
	metrics := c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait
	w.removeDeadline(id, c)
	w.cancelSend(id, c)
	c.Failure, c.Phase = reason, ClientFailed
	if !c.Registered {
		g.counters.failedRegistrations.Add(1)
		w.releaseRegisterSlot(c)
	} else {
		g.counters.failedOperations.Add(1)
		w.releasePositionSlot(c)
	}
	if reason == ClientRetryExhausted {
		g.counters.retryExhausted.Add(1)
		if metrics {
			g.counters.controlRetryExhausted.Add(1)
		}
	}
	if metrics {
		g.counters.failedMetrics.Add(1)
	}
	g.publishAcademicFailure(reason)
}

// Only the owner calls this. There is no live deadline while a retry intention
// or write confirmation is pending, so expiration cannot duplicate that job.
func (w *clientWorker) processTimeouts(now int64) bool {
	g := w.generator
	progress := false
	for i := 0; i < g.cfg.WorkBudget; i++ {
		index, deadline, ok := w.deadlines.peek()
		if !ok || deadline > now {
			break
		}
		id := g.rangeConfig.FirstID + index
		c := &g.clients[index]
		w.removeDeadline(id, c)
		progress = true
		metrics := c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait
		if c.Phase == ClientComplete || c.Phase == ClientFailed || c.Phase == ClientStopped || c.Phase == ClientDone || c.Phase == ClientWaitGroupDone || c.Registered && !metrics && c.AckMask == c.ExpectedMask {
			continue
		}
		g.counters.timeouts.Add(1)
		if metrics {
			g.counters.controlTimeouts.Add(1)
		}
		if c.SendCount >= g.cfg.MaxSends {
			w.failClient(id, c, ClientRetryExhausted)
			continue
		}
		c.CurrentRTO = nextClientRTO(c.CurrentRTO, int64(g.cfg.MaxRTO))
		c.RetryCounter++
		c.Flags |= clientRetryIntent
		w.nextSend(id, c)
		w.schedule(id, c)
	}
	return progress
}
