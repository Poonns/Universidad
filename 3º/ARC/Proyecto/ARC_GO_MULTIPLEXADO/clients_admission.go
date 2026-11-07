package protocol

// These marks describe pending ring entries, not live network operations. A
// cancelled entry may retain a mark until visited; a newly added class is always
// marked by schedule even if that client already owns a pending ring slot.
func (w *clientWorker) notePendingClasses(c *ClientState) {
	if c.Flags&(clientRegisterIntent|clientMetricsIntent) != 0 && c.Flags&clientPendingControl == 0 {
		c.Flags |= clientPendingControl
		w.pendingControl++
	}
	if c.Flags&clientRetryIntent != 0 && c.Flags&clientPendingRetry == 0 {
		c.Flags |= clientPendingRetry
		w.pendingRetry++
	}
}

func (w *clientWorker) clearPendingClasses(c *ClientState) {
	if c.Flags&clientPendingControl != 0 {
		w.pendingControl--
	}
	if c.Flags&clientPendingRetry != 0 {
		w.pendingRetry--
	}
	c.Flags &^= clientPendingIndex | clientPendingControl | clientPendingRetry
}

// Only this worker produces DATA; its writer only removes entries. A snapshot of
// free slots is therefore a conservative admission credit, even during dequeue.
// Existing writer/reader/confirmation wakeups resume the owner; no polling timer.
func (w *clientWorker) flushPendingOptimized() bool {
	g := w.generator
	count := w.pending.stats().Depth
	if count == 0 {
		return false
	}
	q := w.data.stats()
	room := q.Capacity - q.Depth
	if room == 0 && w.pendingControl == 0 && w.pendingRetry == 0 {
		return false
	}
	budget, progress := g.cfg.WorkBudget, false
	turn := 0
	for ; turn < count && turn < g.cfg.WorkBudget && budget > 0; turn++ {
		index, ok := w.pending.pop()
		if !ok {
			break
		}
		id := g.rangeConfig.FirstID + index
		c, err := w.client(id)
		if err != nil {
			g.reportFatal(err)
			break
		}
		w.clearPendingClasses(c)
		if c.Flags&clientRegisterIntent != 0 {
			budget--
			if w.control.push(w.registerJob(id, c), false) {
				c.Flags &^= clientRegisterIntent
				c.Phase = ClientRegisterQueued
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientPositionIntent != 0 && budget > 0 {
			budget-- // skipped intentions still consume the same work budget
			if room > 0 && w.data.push(w.positionJob(id, c), false) {
				room--
				c.Flags &^= clientPositionIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientMetricsIntent != 0 && budget > 0 {
			budget--
			if w.control.push(w.metricsJob(id, c), false) {
				c.Flags &^= clientMetricsIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientRetryIntent != 0 && budget > 0 {
			budget--
			job := w.registerJob(id, c)
			if c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait {
				job = w.metricsJob(id, c)
			} else if c.Registered {
				job = w.positionJob(id, c)
				job.Packet.Type, job.Packet.RetryCounter = RETRY_POSITION, c.RetryCounter
				job.Packet.MissingMask = c.ExpectedMask &^ c.AckMask
				if job.Packet.MissingMask == 0 {
					c.Flags &^= clientRetryIntent
					w.finishOperation(c)
					w.schedule(id, c)
					continue
				}
			}
			if w.retry.push(job, false) {
				c.Flags &^= clientRetryIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		w.schedule(id, c)
	}
	if progress {
		return true
	}
	// If a useful class is beyond this bounded rotation, give it another turn.
	// Once its queue is full, wait for the existing dequeue notification instead.
	if turn < count {
		if w.pendingControl > 0 {
			q := w.control.stats()
			if q.Depth < q.Capacity {
				return true
			}
		}
		if w.pendingRetry > 0 {
			q := w.retry.stats()
			if q.Depth < q.Capacity {
				return true
			}
		}
	}
	return false
}
