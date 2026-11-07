package protocol

import (
	"context"
	"net"
	"runtime"
	"testing"
	"time"
)

func admissionClients(t testing.TB, n uint32, optimized bool, capacity, budget int) *ClientGenerator {
	t.Helper()
	cfg := clientsConfig(t, n, 25, 1, 1)
	cfg.OptimizeDataAdmission, cfg.DataCapacity, cfg.WorkBudget = optimized, capacity, budget
	g := bareClients(t, cfg)
	readyClients(t, g)
	return g
}

func TestDataAdmissionRetainsUniquePendingAndResumesFIFO(t *testing.T) {
	g := admissionClients(t, 25, true, 1, 64)
	w := &g.workers[0]
	for id := uint32(1); id <= 3; id++ {
		if err := w.beginPosition(clientCommand{ClientID: id, Iteration: 1}); err != nil {
			t.Fatal(err)
		}
	}
	w.flushPending()
	first, _ := w.data.peek()
	if first.ClientID != 1 || w.pending.stats().Depth != 2 {
		t.Fatal("bad initial admission")
	}
	before := g.clients[1]
	for i := 0; i < 1000; i++ {
		w.schedule(2, &g.clients[1])
		if w.flushPending() {
			t.Fatal("reported progress against full DATA")
		}
	}
	if g.clients[1] != before || w.pending.stats().Depth != 2 || w.data.stats().FullAttempts != 0 {
		t.Fatal("lost/duplicated pending or futile pushes")
	}
	w.data.pop()
	for _, id := range []uint32{2, 3} {
		if !w.flushPending() {
			t.Fatal("space did not resume admission")
		}
		job, ok := w.data.pop()
		if !ok || job.ClientID != id || job.Packet.Type != POSITION || job.Packet.GroupID != 1 || job.Ticket != g.clients[id-1].SendTicket {
			t.Fatal("FIFO/identity/ownership", job)
		}
	}
	if w.pending.stats().Depth != 0 || w.flushPending() || w.pendingControl+w.pendingRetry != 0 {
		t.Fatal("intention duplicated")
	}
}

func TestDataAdmissionMixedClassesBoundedAndNotStarved(t *testing.T) {
	g := admissionClients(t, 1000, true, 1, 4)
	w := &g.workers[0]
	for id := uint32(1); id <= 1000; id++ {
		if err := w.beginPosition(clientCommand{ClientID: id, Iteration: 1}); err != nil {
			t.Fatal(err)
		}
	}
	w.flushPending() // one DATA admitted; all others retained
	// Replace a queued intention far behind DATA. schedule must update marks
	// without adding another pending entry for this already-pending client.
	c := &g.clients[999]
	w.cancelSend(1000, c)
	c.Phase, c.CompletedIterations, c.MessageID = ClientMetricsQueued, 1, c.RegistrationNonce+2
	c.Flags |= clientMetricsIntent
	w.nextSend(1000, c)
	w.schedule(1000, c)
	r := &g.clients[899]
	w.cancelSend(900, r)
	r.Phase, r.RetryCounter, r.AckMask = ClientPositionWait, 1, r.ExpectedMask&^1
	r.Flags |= clientRetryIntent
	w.nextSend(900, r)
	w.schedule(900, r)
	if w.pendingControl != 1 || w.pendingRetry != 1 || w.pending.stats().Depth != 999 {
		t.Fatal("mixed mark/unique slot")
	}
	for pass := 0; pass < 260 && w.pendingControl+w.pendingRetry > 0; pass++ {
		head := w.pending.head
		progress := w.flushPending()
		visited := (w.pending.head - head + len(w.pending.slots)) % len(w.pending.slots)
		if visited > 4 {
			t.Fatal("scanned beyond bounded work budget", visited)
		}
		if !progress && w.pendingControl+w.pendingRetry > 0 {
			t.Fatal("useful class hidden behind full DATA")
		}
	}
	metric, mok := w.control.pop()
	retry, rok := w.retry.pop()
	if !mok || !rok || metric.ClientID != 1000 || metric.Packet.Type != METRICS || retry.ClientID != 900 || retry.Packet.Type != RETRY_POSITION || retry.Packet.MissingMask != 1 || retry.Attempt != 1 {
		t.Fatal("control/retry starvation or identity", metric, retry)
	}
	if w.pendingControl+w.pendingRetry != 0 || w.data.stats().FullAttempts != 0 || w.flushPending() {
		t.Fatal("busy loop with only full DATA")
	}
	// Stale class marks are cleared on visit; they cannot force an endless scan.
	c = &g.clients[500]
	c.Flags |= clientRetryIntent
	w.schedule(501, c)
	w.cancelSend(501, c)
	for pass := 0; pass < 260 && w.pendingRetry > 0; pass++ {
		w.flushPending()
	}
	if w.pendingRetry != 0 || w.flushPending() {
		t.Fatal("cancelled class left a busy loop")
	}
}

func TestDataAdmissionFullUsefulQueuesWaitAndWakeWithoutLostSignal(t *testing.T) {
	g := admissionClients(t, 25, true, 1, 64)
	w := &g.workers[0]
	for id := uint32(1); id <= 3; id++ {
		w.beginPosition(clientCommand{ClientID: id, Iteration: 1})
	}
	w.flushPending()
	c := &g.clients[2]
	w.cancelSend(3, c)
	c.Flags |= clientRegisterIntent
	w.nextSend(3, c)
	w.schedule(3, c)
	for w.control.stats().Depth < w.control.stats().Capacity {
		w.control.push(clientOutbound{}, false)
	}
	if w.flushPending() {
		t.Fatal("busy loop while DATA/CONTROL full")
	}
	for {
		select {
		case <-w.wake:
			continue
		default:
			goto empty
		}
	}
empty:
	var wait int64
	job, owner, ok := g.popOutbound(2, 0, &wait)
	if !ok || owner != 0 || job.ClientID != 1 {
		t.Fatal("writer failed to free DATA", job)
	}
	select {
	case <-w.wake:
	default:
		t.Fatal("space wakeup lost before owner waits")
	}
	if !w.flushPending() {
		t.Fatal("freed capacity did not progress")
	}
	if next, _ := w.data.peek(); next.ClientID != 2 {
		t.Fatal("pending order changed", next)
	}
}

func TestDataAdmissionPositionWindowUnchanged(t *testing.T) {
	g := admissionClients(t, 25, true, 2, 64)
	w := &g.workers[0]
	g.cfg.PositionWindow = 1
	for id := uint32(1); id <= 2; id++ {
		w.beginPosition(clientCommand{ClientID: id, Iteration: 1})
	}
	w.flushPending()
	var wait int64
	first, _, ok := g.popOutbound(2, 0, &wait)
	if !ok || !g.claimOutbound(first) {
		t.Fatal("first operation missing")
	}
	g.admitPosition()
	if err := w.handleConfirmation(clientPositionConfirmation(first, 1000)); err != nil {
		t.Fatal(err)
	}
	if _, _, ok = g.popOutbound(2, 1000000, &wait); ok {
		t.Fatal("window exceeded")
	}
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		if err := clientDeliver(t, g, clientForwardedACK(first.Packet, neighbor), 2000); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, ok = g.popOutbound(2, 2000000, &wait); !ok || g.Stats().InflightMaximum > 1 {
		t.Fatal("window did not resume")
	}
}

func TestDataAdmissionAcademicFlowAndDirectedLoss(t *testing.T) {
	for _, loss := range []bool{false, true} {
		name := "direct"
		if loss {
			name = "directed-loss"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cfg := simulationServerConfig(t, 100, 25, 5, 6)
			cfg.ServerWriterCount, cfg.ServerEgressSocketCount = 3, 3
			var rules []lossRule
			if loss {
				rules = []lossRule{
					{Direction: lossToServer, Type: REGISTER, Nonce: 1, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: START, Target: 51, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: START_ACK, Origin: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: POSITION, Origin: 1, Target: 2, Iteration: 2, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: ACK, Origin: 52, Target: 51, Iteration: 3, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: METRICS, Origin: 1, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: GROUP_DONE, Target: 51, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 26, Limit: 1, Drop: true},
				}
			}
			s, g, p := simulationListen(t, ctx, cfg, rules, loss, func(c *ClientConfig) { c.Workers = 3; c.DataCapacity = 8; c.OptimizeDataAdmission = true })
			st, ct, ps := simulationCompleteAndClose(t, ctx, s, g, p)
			if loss {
				if ps.Dropped != uint64(len(rules)) || ct.PositionRetries == 0 || ct.RegisterRetries == 0 || ct.MetricsRetries == 0 || st.StartRetries == 0 || st.GroupDoneRetries == 0 {
					t.Fatal("recovery not exercised")
				}
				for _, n := range ps.RuleMatches[:len(rules)] {
					if n != 1 {
						t.Fatal("unexercised loss", ps)
					}
				}
			} else if ct.PositionRetries+ct.RegisterRetries+ct.MetricsRetries+st.StartRetries+st.GroupDoneRetries != 0 {
				t.Fatal("unexpected recovery")
			}
			for i := range g.workers {
				if g.workers[i].pendingControl+g.workers[i].pendingRetry != 0 || ct.Queues[i].Data.FullAttempts != 0 {
					t.Fatal("marks leaked or full DATA push")
				}
			}
		})
	}
}

func TestDataAdmissionShutdownBoundedResources(t *testing.T) {
	for _, n := range []uint32{100, 1000} {
		peer := testSocket(t)
		cfg := clientsConfig(t, n, 25, 1, 3)
		cfg.Server = peer.LocalAddr().(*net.UDPAddr).AddrPort()
		cfg.OptimizeDataAdmission = true
		cfg.RegisterRate, cfg.RegisterWindow, cfg.ControlCapacity = 1, 1, 1
		before := runtime.NumGoroutine()
		g, err := ListenClients(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		if delta := runtime.NumGoroutine() - before; delta != 6 {
			g.Close()
			t.Fatal("goroutines grow with N", delta)
		}
		deadline := time.Now().Add(time.Second)
		for g.Stats().Queues[0].PendingClients.Maximum == 0 && time.Now().Before(deadline) {
			runtime.Gosched()
		}
		if g.Stats().Queues[0].PendingClients.Maximum == 0 {
			g.Close()
			t.Fatal("pending shutdown not exercised")
		}
		if err := g.Close(); err != nil {
			t.Fatal(err)
		}
		st := g.Stats()
		for i := range g.workers {
			w := &g.workers[i]
			q := st.Queues[i]
			if w.pendingControl+w.pendingRetry != 0 || q.PendingClients.Depth+q.Input.Depth+q.High.Depth+q.Control.Depth+q.Data.Depth+q.Retry.Depth+q.Confirmations.Depth+q.Commands.Depth != 0 {
				t.Fatal("shutdown not clean")
			}
		}
		for _, c := range g.clients {
			if c.Flags&(clientPendingIndex|clientPendingControl|clientPendingRetry|clientRegisterIntent|clientPositionIntent|clientMetricsIntent|clientRetryIntent) != 0 {
				t.Fatal("pending flags survived shutdown")
			}
		}
		if st.ARQTimers+st.WriterTimers+st.DeadlineHeapCurrent+st.InflightCurrent+st.RegisterInFlight != 0 {
			t.Fatal("scheduler leaked")
		}
	}
}
