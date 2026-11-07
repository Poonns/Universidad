package protocol

import (
	"fmt"
	"math/bits"
	"sync"
	"testing"
	"time"
)

// These fixtures drive the production owner/writer handoff explicitly. They do
// not manufacture ACK progress: every bit still passes through ApplyACK.
func arqClients(t testing.TB, v, iterations uint32) *ClientGenerator {
	t.Helper()
	cfg := clientsConfig(t, v, v, iterations, 1)
	cfg.InitialRTO, cfg.MaxRTO = time.Second, 5*time.Second
	cfg.MaxSends, cfg.JitterPercent, cfg.JitterSeed = 12, 20, 42
	g := bareClients(t, cfg)
	readyClients(t, g)
	return g
}

func arqConfirmation(job clientOutbound, at int64) clientConfirmation {
	return clientConfirmation{MessageID: job.Packet.MessageID, SentAt: at,
		ClientID: job.ClientID, Iteration: job.Packet.Iteration, Type: job.Packet.Type,
		Ticket: job.Ticket, Attempt: job.Attempt}
}

func arqStart(t testing.TB, g *ClientGenerator, id, iteration uint32, at int64) (*clientWorker, clientOutbound) {
	t.Helper()
	owner, err := g.WorkerForClient(id)
	if err != nil {
		t.Fatal(err)
	}
	w := &g.workers[owner]
	if err := w.beginPosition(clientCommand{id, iteration, PositionPayload{1, -2, 0.5}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	job, ok := w.data.pop()
	if !ok || job.ClientID != id || job.Packet.Type != POSITION {
		t.Fatal("missing original POSITION", job)
	}
	if !g.claimOutbound(job) {
		t.Fatal("writer could not claim original POSITION")
	}
	g.admitPosition()
	if err := w.handleConfirmation(arqConfirmation(job, at)); err != nil {
		t.Fatal(err)
	}
	return w, job
}

func arqExpireAndPop(t testing.TB, w *clientWorker, id uint32) clientOutbound {
	t.Helper()
	c, err := w.client(id)
	if err != nil || c.HeapIndex < 0 || c.Deadline <= 0 {
		t.Fatal("missing live deadline", err, c)
	}
	if !w.processTimeouts(c.Deadline) {
		t.Fatal("due deadline was not processed")
	}
	w.flushPending()
	job, ok := w.retry.pop()
	if !ok || job.ClientID != id {
		t.Fatal("missing retry", job)
	}
	return job
}

func arqACK(t testing.TB, g *ClientGenerator, original Packet, neighbor uint32, at int64, attempt uint16) {
	t.Helper()
	p := clientForwardedACK(original, neighbor)
	p.RetryCounter = attempt
	if err := clientDeliver(t, g, p, at); err != nil {
		t.Fatal(err)
	}
}

func TestClientARQSelectiveMasks(t *testing.T) {
	for _, missing := range []uint32{1 << 24, (1 << 1) | (1 << 24), 2 | 4 | 8 | 16, 0x01fffffe} {
		t.Run(fmt.Sprintf("missing%d", bits.OnesCount32(missing)), func(t *testing.T) {
			g := arqClients(t, 25, 1)
			w, original := arqStart(t, g, 1, 1, 1000)
			c := &g.clients[0]
			for neighbor := uint32(2); neighbor <= 25; neighbor++ {
				if missing&(uint32(1)<<(neighbor-1)) == 0 {
					arqACK(t, g, original.Packet, neighbor, 2000+int64(neighbor), 0)
				}
			}
			retry := arqExpireAndPop(t, w, 1)
			p := retry.Packet
			if p.Type != RETRY_POSITION || p.MissingMask != missing || p.RetryCounter != 1 ||
				p.RunID != original.Packet.RunID || p.GeneratorID != original.Packet.GeneratorID ||
				p.GroupID != original.Packet.GroupID || p.OriginID != original.Packet.OriginID ||
				p.MessageID != original.Packet.MessageID || p.Iteration != original.Packet.Iteration || p.Position != original.Packet.Position {
				t.Fatal("selective retry changed operation or requested unnecessary neighbors", retry)
			}
			if c.SentAt != 1000 || c.SendCount != 1 || c.HeapIndex != -1 {
				t.Fatal("queueing retry retimed RTT, counted an unsent datagram or retained deadline", *c)
			}
			if err := g.cfg.Manifest.Validate(&p); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClientARQBackoffAndPermanentT0(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, original := arqStart(t, g, 1, 1, 1000)
	c := &g.clients[0]
	for _, want := range []int64{int64(time.Second), 2 * int64(time.Second), 4 * int64(time.Second), 5 * int64(time.Second), 5 * int64(time.Second)} {
		if c.CurrentRTO != want {
			t.Fatalf("RTO=%s, want %s", time.Duration(c.CurrentRTO), time.Duration(want))
		}
		deadline := c.Deadline
		retry := arqExpireAndPop(t, w, 1)
		if !g.claimOutbound(retry) {
			t.Fatal("valid retry could not be claimed")
		}
		if err := w.handleConfirmation(arqConfirmation(retry, deadline+1)); err != nil {
			t.Fatal(err)
		}
		if c.SentAt != 1000 || c.MessageID != original.Packet.MessageID || c.HeapIndex < 0 || c.Deadline <= deadline {
			t.Fatal("retry changed original t0 or failed to rearm unique deadline", *c)
		}
	}
	at := c.Deadline - 1
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		arqACK(t, g, original.Packet, neighbor, at, 0) // delayed ACK of the ORIGINAL attempt
	}
	if c.Phase != ClientComplete || c.LastRTTNS != uint64(at-1000) || c.LastRTTNS <= uint64(time.Second) || c.HeapIndex != -1 {
		t.Fatal("RTT excluded recovery time or old-attempt ACK did not finish", *c)
	}
	if g.Stats().CompletedOperations != 1 || g.Stats().RecoveredOperations != 1 || g.Stats().InflightCurrent != 0 {
		t.Fatal("completion/slot counters", g.Stats())
	}
}

func TestClientARQJitterDeterministicAndBounded(t *testing.T) {
	for _, nominal := range []int64{1, 1000, int64(time.Second), 5 * int64(time.Second)} {
		for seed := uint64(0); seed < 1000; seed++ {
			a := clientJitterRTO(nominal, 20, seed)
			b := clientJitterRTO(nominal, 20, seed)
			// Integer rounding can differ by one ns at very small nominal values.
			lower := nominal - nominal/5
			if a != b || a <= 0 || a < lower-1 || a > nominal {
				t.Fatalf("nominal=%d seed=%d jitter=%d/%d outside [%d,%d]", nominal, seed, a, b, lower, nominal)
			}
		}
		if got := clientJitterRTO(nominal, 0, 42); got != nominal {
			t.Fatal("disabled jitter altered nominal RTO", got)
		}
	}
	if nextClientRTO(int64(time.Second), 5*int64(time.Second)) != 2*int64(time.Second) ||
		nextClientRTO(4*int64(time.Second), 5*int64(time.Second)) != 5*int64(time.Second) ||
		nextClientRTO(5*int64(time.Second), 5*int64(time.Second)) != 5*int64(time.Second) {
		t.Fatal("backoff does not cap at MaxRTO")
	}
}

func TestClientARQTimeoutAndACKInterleavings(t *testing.T) {
	for _, timeoutFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("timeoutFirst%v", timeoutFirst), func(t *testing.T) {
			g := arqClients(t, 25, 1)
			w, original := arqStart(t, g, 1, 1, 1000)
			c := &g.clients[0]
			for neighbor := uint32(2); neighbor < 25; neighbor++ {
				arqACK(t, g, original.Packet, neighbor, 2000, 0)
			}
			deadline := c.Deadline
			if timeoutFirst {
				if !w.processTimeouts(deadline) {
					t.Fatal("timeout not observed")
				}
			}
			arqACK(t, g, original.Packet, 25, deadline, 0)
			if !timeoutFirst && w.processTimeouts(deadline) {
				t.Fatal("cancelled deadline was processed")
			}
			w.flushPending()
			if c.Phase != ClientComplete || c.CompletedIterations != 1 || c.HeapIndex != -1 ||
				w.retry.stats().Depth != 0 || w.pending.stats().Depth != 0 || g.Stats().CompletedOperations != 1 {
				t.Fatal("timeout/last ACK created duplicate progress or redundant retry", *c, g.Stats())
			}
		})
	}
}

func TestClientARQQueuedRetryCancelledBeforeClaim(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, original := arqStart(t, g, 1, 1, 1000)
	for neighbor := uint32(2); neighbor < 25; neighbor++ {
		arqACK(t, g, original.Packet, neighbor, 2000, 0)
	}
	retry := arqExpireAndPop(t, w, 1)
	c := &g.clients[0]
	arqACK(t, g, original.Packet, 25, 2*int64(time.Second), 0)
	if g.claimOutbound(retry) {
		t.Fatal("writer claimed a retry cancelled by the last ACK")
	}
	if c.Phase != ClientComplete || c.HeapIndex != -1 || c.SentAt != 1000 || c.SendCount != 1 || g.Stats().CompletedOperations != 1 {
		t.Fatal("cancelled retry altered completed operation", *c)
	}
	for i := 0; i < 100; i++ {
		w.processTimeouts(100 * int64(time.Second))
		w.flushPending()
	}
	if w.pending.stats().Depth != 0 || w.retry.stats().Depth != 0 || g.Stats().InflightCurrent != 0 {
		t.Fatal("cancelled retry generated unlimited work")
	}
}

func TestClientARQClaimedRetryConfirmationAfterNextOperation(t *testing.T) {
	g := arqClients(t, 25, 2)
	w, original := arqStart(t, g, 1, 1, 1000)
	retry := arqExpireAndPop(t, w, 1)
	if !g.claimOutbound(retry) {
		t.Fatal("retry claim failed")
	}
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		arqACK(t, g, original.Packet, neighbor, 2*int64(time.Second), 0)
	}
	// The first operation finished while its last retry was already claimed. The
	// writer may publish that receipt after the owner has accepted iteration 2.
	if err := w.beginPosition(clientCommand{1, 2, PositionPayload{4, 5, 6}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	next, ok := w.data.pop()
	if !ok || next.Ticket == retry.Ticket || next.Packet.MessageID == retry.Packet.MessageID {
		t.Fatal("new operation reused retry identity/ticket", next)
	}
	before := g.clients[0]
	if err := w.handleConfirmation(arqConfirmation(retry, 3*int64(time.Second))); err != nil || g.clients[0] != before {
		t.Fatal("legitimate stale receipt changed newer operation", err, g.clients[0])
	}
	if !g.claimOutbound(next) {
		t.Fatal("old writer receipt erased newer ticket")
	}
	g.admitPosition()
	if err := w.handleConfirmation(arqConfirmation(next, 4*int64(time.Second))); err != nil {
		t.Fatal(err)
	}
	if g.clients[0].SentAt != 4*int64(time.Second) || g.clients[0].SendCount != 1 || g.Stats().InflightCurrent != 1 {
		t.Fatal("old receipt rearmed deadline/retimed/released current operation", g.clients[0], g.Stats())
	}
	if err := clientDeliver(t, g, clientForwardedACK(original.Packet, 2), 5*int64(time.Second)); err == nil || g.clients[0].AckMask != 0 {
		t.Fatal("old operation ACK credited new operation", err)
	}
}

func TestClientARQPendingRetryHasNoSecondDeadline(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, _ := arqStart(t, g, 1, 1, 1000)
	c := &g.clients[0]
	deadline := c.Deadline
	if !w.processTimeouts(deadline) || c.HeapIndex != -1 || w.pending.stats().Depth != 1 {
		t.Fatal("timeout did not create one pending intention")
	}
	for i := 0; i < 1000; i++ {
		if w.processTimeouts(deadline + 100*int64(time.Second)) {
			t.Fatal("timeout recreated an already pending retry")
		}
	}
	w.flushPending()
	if w.retry.stats().Depth != 1 || w.pending.stats().Depth != 0 || c.SendCount != 1 {
		t.Fatal("pending retry was duplicated or counted as sent", w.retry.stats(), *c)
	}
}

func TestClientARQRetryExhaustedAfterTwelveWrites(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, _ := arqStart(t, g, 1, 1, 1000)
	c := &g.clients[0]
	for count := uint16(2); count <= 12; count++ {
		deadline := c.Deadline
		retry := arqExpireAndPop(t, w, 1)
		if !g.claimOutbound(retry) {
			t.Fatal("retry claim failed", count)
		}
		if err := w.handleConfirmation(arqConfirmation(retry, deadline+1)); err != nil || c.SendCount != count || c.Phase == ClientFailed {
			t.Fatal("retry failed before the final receipt deadline", count, err, *c)
		}
	}
	if c.SendCount != 12 || c.Phase != ClientPositionWait {
		t.Fatal("twelfth write was not allowed its ACK wait", *c)
	}
	if !w.processTimeouts(c.Deadline) || c.Phase != ClientFailed || c.Failure != ClientRetryExhausted || c.HeapIndex != -1 {
		t.Fatal("retry exhaustion not explicit", *c)
	}
	w.flushPending()
	st := g.Stats()
	if st.RetryExhausted != 1 || st.FailedOperations != 1 || st.CompletedOperations != 0 || st.InflightCurrent != 0 ||
		w.retry.stats().Depth != 0 || w.pending.stats().Depth != 0 || c.SendCount != 12 {
		t.Fatal("exhaustion leaked slot or queued a thirteenth write", st)
	}
	for i := 0; i < 100; i++ {
		w.processTimeouts(1000 * int64(time.Second))
	}
	if g.Stats().RetryExhausted != 1 {
		t.Fatal("failed operation exhausted more than once")
	}
}

func TestClientARQRegisterRetryIsByteStableAndWindowSingleSlot(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 1, 1, 1, 1))
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	original, ok := w.control.pop()
	if !ok || !g.claimOutbound(original) {
		t.Fatal("REGISTER not claimable")
	}
	g.admitRegister()
	if err := w.handleConfirmation(arqConfirmation(original, 1000)); err != nil {
		t.Fatal(err)
	}
	deadline := g.clients[0].Deadline
	retry := arqExpireAndPop(t, w, 1)
	if retry.Packet != original.Packet || retry.Attempt == original.Attempt || retry.Ticket == original.Ticket || retry.Packet.RetryCounter != 0 {
		t.Fatal("REGISTER retry changed wire identity or reused attempt", original, retry)
	}
	// Retrying this REGISTER must stay eligible even with its window full.
	g.cfg.RegisterWindow = 1
	w.retry.push(retry, false)
	selected, _, eligible, _ := g.takeOutbound(deadline + 1)
	if !eligible || selected.Packet.Type != REGISTER || selected.Attempt != retry.Attempt {
		t.Fatal("full register window prevented retry of its existing slot", selected)
	}
	if !g.claimOutbound(selected) {
		t.Fatal("REGISTER retry not claimable")
	}
	if err := w.handleConfirmation(arqConfirmation(selected, deadline+1)); err != nil || g.registerInFlight.Load() != 1 {
		t.Fatal("REGISTER retry consumed a second slot", err, g.registerInFlight.Load())
	}
	if err := clientDeliver(t, g, clientRegistrationReply(g, 1), deadline+2); err != nil {
		t.Fatal(err)
	}
	before := g.clients[0]
	if err := clientDeliver(t, g, clientRegistrationReply(g, 1), deadline+3); err != nil || g.clients[0] != before {
		t.Fatal("REGISTER_ACK duplicate mutated registered client", err)
	}
	if g.clients[0].HeapIndex != -1 || g.Stats().RegisteredClients != 1 || g.registerInFlight.Load() != 0 {
		t.Fatal("REGISTER retry leaked deadline/window or registered twice", g.clients[0], g.Stats())
	}
}

func TestClientARQCompletionAndOwnerCloseEmptyHeap(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, original := arqStart(t, g, 1, 1, 1000)
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		arqACK(t, g, original.Packet, neighbor, 2000, 0)
	}
	if g.Stats().DeadlineHeapCurrent != 0 || g.clients[0].HeapIndex != -1 {
		t.Fatal("completion leaked heap entry")
	}
	for id := uint32(2); id <= 25; id++ {
		arqStart(t, g, id, 1, 3000)
	}
	if g.Stats().DeadlineHeapCurrent != 24 {
		t.Fatal("one live heap entry per operation was not maintained", g.Stats())
	}
	w.finish()
	if g.Stats().DeadlineHeapCurrent != 0 || g.Stats().InflightCurrent != 0 {
		t.Fatal("owner shutdown leaked active deadlines/slots", g.Stats())
	}
	for _, c := range g.clients {
		if c.HeapIndex != -1 {
			t.Fatal("shutdown left stale indexed heap position", c)
		}
	}
}

func TestClientARQNewPositionBucketAndReactiveACK(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.PositionRate, cfg.PositionBurst = 1, 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w := &g.workers[0]
	for id := uint32(1); id <= 2; id++ {
		if err := w.beginPosition(clientCommand{id, 1, PositionPayload{1, 2, 3}}); err != nil {
			t.Fatal(err)
		}
	}
	w.flushPending()
	first, _, ok, _ := g.takeOutbound(0)
	if !ok || first.Packet.Type != POSITION || !g.positions.take(0) || !g.claimOutbound(first) {
		t.Fatal("first burst token did not admit original POSITION")
	}
	g.admitPosition()
	if err := w.handleConfirmation(arqConfirmation(first, 1)); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, wait := g.takeOutbound(0); ok || wait != int64(time.Second) || w.data.stats().Depth != 1 {
		t.Fatal("empty token bucket admitted another POSITION", ok, wait)
	}
	if err := w.handleIncoming(clientInbound{clientForwardedPosition(g, 2, 1), 100}); err != nil {
		t.Fatal(err)
	}
	if ack, _, ok, _ := g.takeOutbound(0); !ok || ack.Packet.Type != ACK {
		t.Fatal("empty new POSITION bucket held indispensable reactive ACK")
	}
	if _, _, ok, _ := g.takeOutbound(int64(time.Second) - 1); ok {
		t.Fatal("bucket replenished before its configured interval")
	}
	if next, _, ok, _ := g.takeOutbound(int64(time.Second)); !ok || next.Packet.Type != POSITION || !g.positions.take(int64(time.Second)) {
		t.Fatal("second POSITION did not resume after token refill")
	}
}

func TestClientARQRetryPriorityAndSeparateBudget(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.PositionRate, cfg.PositionBurst, cfg.RetryRate, cfg.RetryBurst = 1000, 2, 1, 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w, _ := arqStart(t, g, 1, 1, 1000)
	deadline := g.clients[0].Deadline
	if !w.processTimeouts(deadline) {
		t.Fatal("existing operation did not time out")
	}
	if err := w.beginPosition(clientCommand{2, 1, PositionPayload{1, 2, 3}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	if w.data.stats().Depth != 1 || w.retry.stats().Depth != 1 {
		t.Fatal("both priority classes were not eligible")
	}
	retry, _, ok, _ := g.takeOutbound(deadline)
	if !ok || retry.Packet.Type != RETRY_POSITION || !g.retries.take(deadline) {
		t.Fatal("new POSITION outran eligible recovery work", retry)
	}
	if !g.positions.take(deadline) {
		t.Fatal("retry consumed the independent new POSITION token bucket")
	}
	if g.retries.take(deadline) {
		t.Fatal("one retry token permitted an unlimited burst")
	}
	// A second existing retry is delayed by its own budget, while ACK service is
	// unaffected. This duplicate is an isolated bounded selector fixture.
	w.retry.push(retry, false)
	w.data.discardAll()
	if _, _, ok, wait := g.takeOutbound(deadline); ok || wait <= deadline || w.retry.stats().Depth != 1 {
		t.Fatal("retry budget was bypassed or job was dropped", ok, wait)
	}
	w.high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	if ack, _, ok, _ := g.takeOutbound(deadline); !ok || ack.Packet.Type != ACK {
		t.Fatal("retry budget delayed reactive ACK")
	}
}

func TestClientARQPositionWindowPreservesPendingAndReleasesOnce(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.PositionWindow, cfg.PositionBurst = 1, 25
	g := bareClients(t, cfg)
	readyClients(t, g)
	w, original := arqStart(t, g, 1, 1, 1000)
	for id := uint32(2); id <= 25; id++ {
		if err := w.beginPosition(clientCommand{id, 1, PositionPayload{1, 2, 3}}); err != nil {
			t.Fatal(err)
		}
	}
	w.flushPending()
	if _, _, ok, _ := g.takeOutbound(2000); ok || w.data.stats().Depth != 24 || g.Stats().InflightCurrent != 1 || g.Stats().InflightMaximum != 1 {
		t.Fatal("full inflight window admitted another operation or lost pending DATA", g.Stats())
	}
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		arqACK(t, g, original.Packet, neighbor, 3000, 0)
	}
	if g.Stats().InflightCurrent != 0 {
		t.Fatal("completion did not release operation slot")
	}
	arqACK(t, g, original.Packet, 25, 4000, 0)
	if g.Stats().InflightCurrent != 0 {
		t.Fatal("duplicate ACK released slot twice")
	}
	next, _, ok, _ := g.takeOutbound(5000)
	if !ok || next.Packet.Type != POSITION || next.ClientID != 2 || !g.claimOutbound(next) {
		t.Fatal("released operation window did not resume pending POSITION")
	}
	g.admitPosition()
	if err := w.handleConfirmation(arqConfirmation(next, 5000)); err != nil || g.Stats().InflightCurrent != 1 || g.Stats().InflightMaximum != 1 {
		t.Fatal("new operation exceeded window after resumption", err, g.Stats())
	}
}

func TestClientARQDeadlineWaitsForOriginalReceiptWithEarlyACKs(t *testing.T) {
	g := arqClients(t, 25, 1)
	w := &g.workers[0]
	if err := w.beginPosition(clientCommand{1, 1, PositionPayload{1, -2, 0.5}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	job, ok := w.data.pop()
	if !ok || !g.claimOutbound(job) {
		t.Fatal("original POSITION not claimable")
	}
	g.admitPosition()
	c := &g.clients[0]
	if c.HeapIndex != -1 || c.Deadline > 0 || c.SentAt != -1 || w.processTimeouts(100*int64(time.Second)) {
		t.Fatal("deadline or RTT started before first writer receipt", *c)
	}
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		arqACK(t, g, job.Packet, neighbor, 2000+int64(neighbor), 0)
	}
	if c.Phase == ClientComplete || c.HeapIndex != -1 || g.Stats().InflightCurrent != 1 || c.SendCount != 0 {
		t.Fatal("ACK-before-receipt completed or prematurely released admission", *c, g.Stats())
	}
	if err := w.handleConfirmation(arqConfirmation(job, 1000)); err != nil {
		t.Fatal(err)
	}
	if c.Phase != ClientComplete || c.LastRTTNS != 1025 || c.HeapIndex != -1 || g.Stats().InflightCurrent != 0 || g.Stats().CompletedOperations != 1 {
		t.Fatal("initial receipt did not atomically settle early ACKs and slot", *c, g.Stats())
	}
}

func TestClientARQRejectsInvalidConfiguration(t *testing.T) {
	for _, mutate := range []func(*ClientConfig){
		func(c *ClientConfig) { c.InitialRTO = 0 },
		func(c *ClientConfig) { c.MaxRTO = c.InitialRTO - 1 },
		func(c *ClientConfig) { c.MaxSends = 0 },
		func(c *ClientConfig) { c.JitterPercent = -1 },
		func(c *ClientConfig) { c.JitterPercent = 51 },
		func(c *ClientConfig) { c.PositionWindow = 0 },
		func(c *ClientConfig) { c.PositionBurst = 0 },
		func(c *ClientConfig) { c.RetryCapacity = 0 },
		func(c *ClientConfig) { c.RetryRate = 0 },
		func(c *ClientConfig) { c.RetryBurst = 0 },
	} {
		cfg := clientsConfig(t, 25, 25, 1, 1)
		mutate(&cfg)
		if _, err := newClientGenerator(cfg); err != ErrClientConfig {
			t.Fatal("invalid ARQ/admission configuration accepted", err)
		}
	}
}

func TestClientARQShutdownCountsIntentionsBeforeCancellation(t *testing.T) {
	t.Run("REGISTER", func(t *testing.T) {
		g := bareClients(t, clientsConfig(t, 3, 3, 1, 1))
		w := &g.workers[0]
		w.stageRegistrations()
		w.finish()
		if st := g.Stats(); st.ShutdownIntentions != 3 || st.RegisteredClients != 0 || st.DeadlineHeapCurrent != 0 {
			t.Fatal("shutdown erased REGISTER intention accounting", st)
		}
	})
	t.Run("POSITION", func(t *testing.T) {
		g := arqClients(t, 3, 1)
		w := &g.workers[0]
		if err := w.beginPosition(clientCommand{1, 1, PositionPayload{1, 2, 3}}); err != nil {
			t.Fatal(err)
		}
		w.finish()
		if st := g.Stats(); st.ShutdownIntentions != 1 || st.AbortedOperations != 1 || st.InflightCurrent != 0 {
			t.Fatal("shutdown erased POSITION intention accounting", st)
		}
	})
	t.Run("RETRY", func(t *testing.T) {
		g := arqClients(t, 3, 1)
		w, _ := arqStart(t, g, 1, 1, 1000)
		if !w.processTimeouts(g.clients[0].Deadline) {
			t.Fatal("retry not staged")
		}
		w.finish()
		if st := g.Stats(); st.ShutdownIntentions != 1 || st.AbortedOperations != 1 || st.InflightCurrent != 0 || st.DeadlineHeapCurrent != 0 {
			t.Fatal("shutdown erased RETRY intention accounting", st)
		}
	})
}

func TestClientARQLateRegisterReceiptRejectsFutureAndInvalidTime(t *testing.T) {
	for _, future := range []bool{false, true} {
		t.Run(fmt.Sprintf("future%v", future), func(t *testing.T) {
			g := arqClients(t, 3, 1)
			w := &g.workers[0]
			if err := w.beginPosition(clientCommand{1, 1, PositionPayload{1, 2, 3}}); err != nil {
				t.Fatal(err)
			}
			c := &g.clients[0]
			before := *c
			tx := clientConfirmation{MessageID: c.RegistrationNonce, ClientID: 1, Type: REGISTER, SentAt: 1000, Ticket: c.SendTicket - 1}
			if future {
				tx.Ticket = c.SendTicket + 1
			} else {
				tx.SentAt = -1
			}
			if err := w.handleConfirmation(tx); err == nil || *c != before || g.Stats().InvalidConfirmations != 1 {
				t.Fatal("contradictory late REGISTER receipt accepted or changed POSITION", err, *c)
			}
		})
	}
}

func TestClientARQPermanentRegisterLossExhaustsConfiguredWrites(t *testing.T) {
	cfg := clientsConfig(t, 1, 1, 1, 1)
	cfg.MaxSends = 3
	g := bareClients(t, cfg)
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	original, ok := w.control.pop()
	if !ok || !g.claimOutbound(original) {
		t.Fatal("original REGISTER not claimable")
	}
	g.admitRegister()
	if err := w.handleConfirmation(arqConfirmation(original, 1000)); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	for count := uint16(2); count <= cfg.MaxSends; count++ {
		deadline := c.Deadline
		retry := arqExpireAndPop(t, w, 1)
		if retry.Packet != original.Packet || !g.claimOutbound(retry) {
			t.Fatal("REGISTER retry changed identity or lost claim", retry)
		}
		if err := w.handleConfirmation(arqConfirmation(retry, deadline+1)); err != nil || c.SendCount != count || g.registerInFlight.Load() != 1 {
			t.Fatal("REGISTER attempt accounting/window", count, err, *c)
		}
	}
	if !w.processTimeouts(c.Deadline) || c.Phase != ClientFailed || c.Failure != ClientRetryExhausted || c.Registered {
		t.Fatal("permanent REGISTER loss did not fail explicitly", *c)
	}
	w.flushPending()
	if st := g.Stats(); st.RetryExhausted != 1 || st.FailedRegistrations != 1 || st.FailedOperations != 0 || st.RegisterInFlight != 0 ||
		st.DeadlineHeapCurrent != 0 || w.retry.stats().Depth != 0 || w.pending.stats().Depth != 0 || c.SendCount != cfg.MaxSends {
		t.Fatal("REGISTER exhaustion leaked slot/deadline or generated unlimited work", st)
	}
}

func TestClientARQPartialACKBeforeFlushRefreshesMissingMask(t *testing.T) {
	g := arqClients(t, 25, 1)
	w, original := arqStart(t, g, 1, 1, 1000)
	c := &g.clients[0]
	deadline := c.Deadline
	if !w.processTimeouts(deadline) {
		t.Fatal("timeout not processed")
	}
	arqACK(t, g, original.Packet, 2, deadline+1, 0)
	w.flushPending()
	retry, ok := w.retry.pop()
	if !ok || retry.Packet.MissingMask != c.ExpectedMask&^2 || bits.OnesCount32(retry.Packet.MissingMask) != 23 {
		t.Fatal("retry snapshot ignored newly received partial ACK", retry)
	}
}

func TestClientARQBoundedRetryQueueRetainsOneIntention(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.RetryCapacity = 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w, _ := arqStart(t, g, 1, 1, 1000)
	arqStart(t, g, 2, 1, 1000)
	// Equal deadlines give a deterministic owner order; this only changes the
	// indexed fixture's test clock, not retry scheduling or network behavior.
	if !w.deadlines.set(0, 10000) || !w.deadlines.set(1, 10000) || !w.processTimeouts(10000) {
		t.Fatal("fixture could not expire both operations")
	}
	w.flushPending()
	if w.retry.stats().Depth != 1 || w.pending.stats().Depth != 1 {
		t.Fatal("full retry queue did not retain exactly one owner intention")
	}
	if n := testing.AllocsPerRun(1000, func() { w.flushPending() }); n != 0 {
		t.Fatal("retry backpressure allocated", n)
	}
	if w.retry.stats().Maximum != 1 || w.retry.stats().Drops != 0 || w.pending.stats().Depth != 1 ||
		g.clients[0].SendCount != 1 || g.clients[1].SendCount != 1 || w.processTimeouts(100*int64(time.Second)) {
		t.Fatal("backpressure grew retries, counted unsent attempts, or retained a second timer")
	}
	first, ok := w.retry.pop()
	if !ok || first.ClientID != 1 {
		t.Fatal("deterministic first retry missing", first)
	}
	w.flushPending()
	second, ok := w.retry.pop()
	if !ok || second.ClientID != 2 || w.pending.stats().Depth != 0 || w.retry.stats().Depth != 0 {
		t.Fatal("bounded retry intention did not resume after space became available", second)
	}
}

func TestClientARQWriterClaimRacesOwnerLastACK(t *testing.T) {
	for turn := 0; turn < 100; turn++ {
		g := arqClients(t, 3, 1)
		w, original := arqStart(t, g, 1, 1, 1000)
		arqACK(t, g, original.Packet, 2, 2000, 0)
		retry := arqExpireAndPop(t, w, 1)
		start := make(chan struct{}) // fixed TEST actors, never per-client production channels
		var wg sync.WaitGroup
		wg.Add(2)
		claimed := false
		var ackErr error
		go func() {
			defer wg.Done()
			<-start
			claimed = g.claimOutbound(retry) // writer reads only atomic metadata
		}()
		go func() {
			defer wg.Done()
			<-start
			ackErr = w.handleIncoming(clientInbound{clientForwardedACK(original.Packet, 3), 2 * int64(time.Second)})
		}()
		close(start)
		wg.Wait()
		if ackErr != nil || g.clients[0].Phase != ClientComplete || g.Stats().CompletedOperations != 1 || g.Stats().InflightCurrent != 0 {
			t.Fatal("writer claim/owner completion race changed progress", ackErr, g.clients[0])
		}
		if claimed {
			// A claim won before cancellation: the write may already be effective.
			// Its eventual receipt is harmless and must not rearm the operation.
			if err := w.handleConfirmation(arqConfirmation(retry, int64(time.Second))); err != nil {
				t.Fatal(err)
			}
		}
		if g.clients[0].HeapIndex != -1 || g.clients[0].SentAt != 1000 || g.Stats().CompletedOperations != 1 ||
			g.sendGates[0].Load() != 0 || g.claimOutbound(retry) {
			t.Fatal("late claimed retry rearmed, retimed, or escaped cancellation")
		}
	}
}

func TestClientARQRetryBacklogStopsNewAdmissionUntilBudgetRefills(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.PositionBurst, cfg.RetryRate, cfg.RetryBurst = 2, 1, 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w, _ := arqStart(t, g, 1, 1, 1000)
	now := g.clients[0].Deadline
	retry := arqExpireAndPop(t, w, 1)
	if !g.retries.take(now) || !w.retry.push(retry, false) {
		t.Fatal("fixture could not exhaust retry budget with a remaining retry")
	}
	if err := w.beginPosition(clientCommand{2, 1, PositionPayload{1, 2, 3}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	if g.positions.ready(now) != now || w.data.stats().Depth != 1 || w.retry.stats().Depth != 1 {
		t.Fatal("fixture does not have eligible new work and existing recovery")
	}
	if _, _, ok, wait := g.takeOutbound(now); ok || wait <= now || w.data.stats().Depth != 1 || w.retry.stats().Depth != 1 {
		t.Fatal("retry-budget pressure admitted newer work or dropped recovery", ok, wait)
	}
	w.high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	if ack, _, ok, _ := g.takeOutbound(now); !ok || ack.Packet.Type != ACK || w.data.stats().Depth != 1 {
		t.Fatal("retry backlog prevented ACK service")
	}
	ready := g.retries.ready(now)
	job, _, ok, _ := g.takeOutbound(ready)
	if !ok || job.Packet.Type != RETRY_POSITION || !g.retries.take(ready) || !g.claimOutbound(job) {
		t.Fatal("recovery did not leave first when its budget replenished", job)
	}
	if err := w.handleConfirmation(arqConfirmation(job, ready)); err != nil {
		t.Fatal(err)
	}
	if next, _, ok, _ := g.takeOutbound(ready); !ok || next.Packet.Type != POSITION || next.ClientID != 2 {
		t.Fatal("new admission did not resume after draining retry backlog", next)
	}
}
