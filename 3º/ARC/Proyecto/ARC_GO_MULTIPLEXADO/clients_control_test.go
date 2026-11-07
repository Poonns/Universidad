package protocol

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

func academicClients(t testing.TB, n, v, iterations uint32) *ClientGenerator {
	t.Helper()
	cfg := clientsConfig(t, n, v, iterations, 1)
	cfg.Academic = true
	g := bareClients(t, cfg)
	readyClients(t, g)
	return g
}

func clientAcademicPacket(g *ClientGenerator, typ MessageType, id uint32) Packet {
	msg := clientStartMessageID
	if typ == GROUP_DONE {
		msg = clientGroupDoneMessageID
	}
	if typ == METRICS_ACK {
		msg = uint64(id-g.rangeConfig.FirstID) + uint64(g.cfg.Manifest.run.I) + 2
	}
	return Packet{Header: Header{Type: typ, Flags: ServerToClient, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, GroupID: 1 + (id-1)/g.cfg.Manifest.run.V, TargetID: id, MessageID: msg}}
}

func academicFinishIterations(t testing.TB, g *ClientGenerator, id uint32) (*clientWorker, clientOutbound) {
	t.Helper()
	owner, _ := g.WorkerForClient(id)
	w := &g.workers[owner]
	if err := clientDeliver(t, g, clientAcademicPacket(g, START, id), 100); err != nil {
		t.Fatal(err)
	}
	w.high.pop()
	var last clientOutbound
	for iteration := uint32(1); iteration <= g.cfg.Manifest.run.I; iteration++ {
		w.flushPending()
		job, ok := w.data.pop()
		if !ok || job.ClientID != id || job.Packet.Iteration != iteration || !g.claimOutbound(job) {
			t.Fatal("automatic iteration missing", job)
		}
		g.admitPosition()
		at := int64(iteration) * 1000
		if err := w.handleConfirmation(clientPositionConfirmation(job, at)); err != nil {
			t.Fatal(err)
		}
		base := (id-1)/g.cfg.Manifest.run.V*g.cfg.Manifest.run.V + 1
		for neighbor := base; neighbor < base+g.cfg.Manifest.run.V; neighbor++ {
			if neighbor == id {
				continue
			}
			if err := clientDeliver(t, g, clientForwardedACK(job.Packet, neighbor), at+500); err != nil {
				t.Fatal(err)
			}
		}
		last = job
	}
	return w, last
}

func academicMetricsJob(t testing.TB, g *ClientGenerator, id uint32) (*clientWorker, clientOutbound) {
	t.Helper()
	w, _ := academicFinishIterations(t, g, id)
	w.flushPending()
	job, ok := w.control.pop()
	if !ok || job.Packet.Type != METRICS || job.ClientID != id || !g.claimOutbound(job) {
		t.Fatal("METRICS missing", job)
	}
	return w, job
}

func TestClientAcademicStartBeforeRegisterACK(t *testing.T) {
	cfg := clientsConfig(t, 3, 3, 2, 1)
	cfg.Academic = true
	g := bareClients(t, cfg)
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	job, _ := w.control.pop()
	g.admitRegister()
	start := clientAcademicPacket(g, START, 1)
	start.RetryCounter = 2
	if err := clientDeliver(t, g, start, 100); err != nil {
		t.Fatal(err)
	}
	if g.clients[0].Registered || g.clients[0].Flags&clientPendingStart == 0 || w.high.stats().Depth != 0 || g.Stats().StartedClients != 0 {
		t.Fatal("START incorrectly accredited before registration")
	}
	// A legitimate neighbor can begin after the server's barrier while this
	// client's registration ACK is still missing. It must not invent authority.
	if err := clientDeliver(t, g, clientForwardedPosition(g, 2, 1), 110); err != nil {
		t.Fatal(err)
	}
	if w.high.stats().Depth != 0 || g.Stats().PreRegistrationData != 1 || g.Stats().InvalidPackets != 0 {
		t.Fatal("pre-registration data generated an ACK or logical error")
	}
	if err := clientDeliver(t, g, clientRegistrationReply(g, 1), 200); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	ack, ok := w.high.pop()
	if !ok || ack.Packet.Type != START_ACK || ack.Packet.RetryCounter != 2 || c.Phase != ClientPositionQueued || c.Iteration != 1 || c.Flags&clientPendingStart != 0 || g.Stats().StartedClients != 1 {
		t.Fatal("retained START was lost or restarted", *c, ack)
	}
	if err := w.handleConfirmation(arqConfirmation(job, 50)); err != nil {
		t.Fatal(err)
	}
	if c.Phase != ClientPositionQueued || c.HeapIndex != -1 || c.SentAt != -1 || g.registerInFlight.Load() != 0 {
		t.Fatal("late registration receipt corrupted automatic iteration", *c)
	}
}

func TestClientAcademicAutomaticFiveIterations(t *testing.T) {
	for _, v := range []uint32{1, 25, 32} {
		t.Run(fmt.Sprintf("V%d", v), func(t *testing.T) {
			g := academicClients(t, v, v, 5)
			w, last := academicFinishIterations(t, g, 1)
			c := &g.clients[0]
			if c.CompletedIterations != 5 || c.Phase != ClientMetricsQueued || c.HeapIndex != -1 || g.positionInFlight.Load() != 0 || g.Stats().CompletedOperations != 5 || g.Stats().UniqueAcks != uint64(5*(v-1)) {
				t.Fatal("automatic simulation progression", *c, g.Stats())
			}
			w.flushPending()
			metrics, _ := w.control.pop()
			wantMean := uint64(500)
			if v == 1 {
				wantMean = 0
			}
			if metrics.Packet.Type != METRICS || metrics.Packet.MessageID != 7 || metrics.Packet.Metrics != (MetricsPayload{wantMean, 5}) || metrics.Packet.Iteration != 0 {
				t.Fatal("incorrect academic metrics", metrics)
			}
			if g.cfg.Manifest.Validate(&metrics.Packet) != nil {
				t.Fatal("METRICS violates v1")
			}
			if v > 1 {
				before := *c
				if err := clientDeliver(t, g, clientForwardedACK(last.Packet, 2), 9000); err != nil || *c != before {
					t.Fatal("late final ACK changed metrics", err)
				}
				old := last.Packet
				old.Iteration, old.MessageID = 1, 2
				if err := clientDeliver(t, g, clientForwardedACK(old, 2), 10000); err != nil || *c != before || g.Stats().StaleACKs != 1 {
					t.Fatal("previous-iteration ACK changed metrics", err)
				}
			}
		})
	}
}

func TestClientAcademicStartIdempotent(t *testing.T) {
	g := academicClients(t, 3, 3, 2)
	w := &g.workers[0]
	p := clientAcademicPacket(g, START, 1)
	if err := clientDeliver(t, g, p, 100); err != nil {
		t.Fatal(err)
	}
	w.high.pop()
	before := g.clients[0]
	p.RetryCounter = 4
	if err := clientDeliver(t, g, p, 200); err != nil || g.clients[0] != before {
		t.Fatal("duplicate START restarted client", err)
	}
	ack, _ := w.high.pop()
	if ack.Packet.Type != START_ACK || ack.Packet.RetryCounter != 4 || g.Stats().StartedClients != 1 || g.Stats().StartedOperations != 1 || g.Stats().DuplicateStarts != 1 {
		t.Fatal("START deduplication", g.Stats())
	}
	if err := g.StartPosition(1, 1, PositionPayload{}); err != ErrClientPhase {
		t.Fatal("manual API bypassed academic START", err)
	}
}

func TestClientAcademicOriginalReceiptDoesNotArmNextOperation(t *testing.T) {
	g := academicClients(t, 3, 3, 2)
	w := &g.workers[0]
	if err := clientDeliver(t, g, clientAcademicPacket(g, START, 1), 100); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	job, _ := w.data.pop()
	g.claimOutbound(job)
	g.admitPosition()
	for neighbor := uint32(2); neighbor <= 3; neighbor++ {
		if err := clientDeliver(t, g, clientForwardedACK(job.Packet, neighbor), 1500); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	if c.Iteration != 2 || c.Phase != ClientPositionQueued || c.HeapIndex != -1 || c.SentAt != -1 || c.SendCount != 0 || c.Flags&clientPositionIntent == 0 || g.Stats().InflightCurrent != 0 {
		t.Fatal("old receipt armed new operation's deadline", *c)
	}
	before := *c
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil || *c != before {
		t.Fatal("old receipt corrupted next operation", err)
	}
}

func TestClientAcademicMetricsReliableAndCanonical(t *testing.T) {
	g := academicClients(t, 3, 3, 1)
	w, job := academicMetricsJob(t, g, 1)
	c := &g.clients[0]
	if c.HeapIndex != -1 || w.processTimeouts(100*int64(time.Second)) {
		t.Fatal("METRICS deadline armed before Write confirmation")
	}
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		t.Fatal(err)
	}
	deadline := c.Deadline
	if !w.processTimeouts(deadline) {
		t.Fatal("METRICS timeout not processed")
	}
	w.flushPending()
	retry, ok := w.retry.pop()
	if !ok || retry.Packet != job.Packet || retry.Attempt != 1 || retry.Packet.RetryCounter != 0 || !bytes.Equal(wire(t, retry.Packet), wire(t, job.Packet)) {
		t.Fatal("METRICS retry changed operation/payload", retry)
	}
	if !g.claimOutbound(retry) {
		t.Fatal("METRICS retry cancellation gate")
	}
	if err := w.handleConfirmation(arqConfirmation(retry, deadline+1)); err != nil {
		t.Fatal(err)
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, METRICS_ACK, 1), deadline+2); err != nil {
		t.Fatal(err)
	}
	before := *c
	if c.Phase != ClientWaitGroupDone || c.HeapIndex != -1 || g.Stats().MetricsAccepted != 1 || g.Stats().ControlTimeouts != 1 {
		t.Fatal("METRICS_ACK did not settle ARQ", *c)
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, METRICS_ACK, 1), deadline+3); err != nil || *c != before || g.Stats().MetricsAccepted != 1 {
		t.Fatal("METRICS_ACK duplicated transition", err)
	}
	if err := w.handleConfirmation(arqConfirmation(retry, deadline+1)); err != nil || *c != before {
		t.Fatal("late METRICS receipt rearmed accepted operation", err)
	}
}

func TestClientAcademicGroupDoneBeforeMetricsACK(t *testing.T) {
	g := academicClients(t, 3, 3, 1)
	w, job := academicMetricsJob(t, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		t.Fatal(err)
	}
	w.processTimeouts(g.clients[0].Deadline)
	w.flushPending()
	retry, ok := clientQueuePeek(w.retry)
	if !ok {
		t.Fatal("METRICS retry fixture missing")
	}
	p := clientAcademicPacket(g, GROUP_DONE, 1)
	if err := clientDeliver(t, g, p, 10000); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	if c.Phase != ClientDone || c.Flags&clientMetricsAccepted == 0 || c.HeapIndex != -1 || g.claimOutbound(retry) || g.Stats().MetricsAccepted != 1 || g.Stats().DoneClients != 1 {
		t.Fatal("GROUP_DONE did not prove metric acceptance/cancel retry", *c)
	}
	ack, _ := w.high.pop()
	if ack.Packet.Type != GROUP_DONE_ACK {
		t.Fatal("missing terminal ACK", ack)
	}
	before := *c
	if err := clientDeliver(t, g, p, 10001); err != nil || *c != before || g.Stats().DoneClients != 1 {
		t.Fatal("GROUP_DONE not idempotent", err)
	}
	w.high.pop()
	if err := clientDeliver(t, g, clientAcademicPacket(g, METRICS_ACK, 1), 10002); err != nil || *c != before {
		t.Fatal("late METRICS_ACK altered DONE", err)
	}
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil || *c != before {
		t.Fatal("late metric writer receipt altered DONE", err)
	}
	// DONE remains a live logical participant. It helps its slow neighbor and
	// answers repeated START without scheduling another own operation.
	if err := clientDeliver(t, g, clientForwardedPosition(g, 2, 1), 11000); err != nil {
		t.Fatal(err)
	}
	neighborACK, _ := w.high.pop()
	if neighborACK.Packet.Type != ACK || neighborACK.Packet.OriginID != 1 || neighborACK.Packet.TargetID != 2 || *c != before {
		t.Fatal("DONE stopped helping neighbor")
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, START, 1), 11001); err != nil || *c != before {
		t.Fatal("DONE restarted on late START", err)
	}
	lateStartACK, _ := w.high.pop()
	if lateStartACK.Packet.Type != START_ACK {
		t.Fatal("DONE did not answer late START")
	}
}

func TestClientAcademicMetricsACKBeforeWriterReceipt(t *testing.T) {
	g := academicClients(t, 3, 3, 1)
	w, job := academicMetricsJob(t, g, 1)
	if err := clientDeliver(t, g, clientAcademicPacket(g, METRICS_ACK, 1), 9000); err != nil {
		t.Fatal(err)
	}
	before := g.clients[0]
	if err := w.handleConfirmation(arqConfirmation(job, 8000)); err != nil || g.clients[0] != before || g.Stats().DeadlineHeapCurrent != 0 {
		t.Fatal("early METRICS_ACK lost to local receipt", err)
	}
}

func TestClientAcademicMetricsExhaustionAndWaiters(t *testing.T) {
	g := academicClients(t, 3, 3, 1)
	g.cfg.MaxSends = 3
	w, job := academicMetricsJob(t, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	for send := uint16(2); send <= g.cfg.MaxSends; send++ {
		at := c.Deadline
		w.processTimeouts(at)
		w.flushPending()
		retry, ok := w.retry.pop()
		if !ok || !g.claimOutbound(retry) {
			t.Fatal("bounded METRICS retry missing")
		}
		if err := w.handleConfirmation(arqConfirmation(retry, at+1)); err != nil {
			t.Fatal(err)
		}
	}
	w.processTimeouts(c.Deadline)
	w.flushPending()
	if c.Phase != ClientFailed || c.SendCount != 3 || c.HeapIndex != -1 || w.retry.stats().Depth != 0 || g.Stats().FailedMetrics != 1 || g.Stats().ControlRetryExhausted != 1 {
		t.Fatal("METRICS exhaustion not finite/explicit", *c, g.Stats())
	}
	for waiter := 0; waiter < 2; waiter++ {
		if err := g.WaitAllDone(context.Background()); err != ErrRetryExhausted {
			t.Fatal("WaitAllDone failed to unblock every waiter", err)
		}
	}
}

func TestClientAcademicWaitAllDoneDoesNotCloseSocket(t *testing.T) {
	g := academicClients(t, 1, 1, 1)
	w, job := academicMetricsJob(t, g, 1)
	if err := w.handleConfirmation(arqConfirmation(job, 5000)); err != nil {
		t.Fatal(err)
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, GROUP_DONE, 1), 6000); err != nil {
		t.Fatal(err)
	}
	if err := g.WaitAllDone(context.Background()); err != nil {
		t.Fatal(err)
	}
	if g.ctx.Err() != nil || w.high.stats().Depth != 1 || g.Stats().GroupDoneAckSent != 0 {
		t.Fatal("logical DONE cancelled generator or assumed delivery of last ACK")
	}
	select {
	case <-g.done:
		t.Fatal("generator closed at logical completion")
	default:
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, GROUP_DONE, 1), 7000); err != nil || w.high.stats().Depth != 2 {
		t.Fatal("last-ACK recovery impossible after WaitAllDone", err)
	}
	if err := g.WaitAllDone(nil); err != ErrClientConfig {
		t.Fatal(err)
	}
}

func TestClientAcademicRejectsContradictoryControl(t *testing.T) {
	g := academicClients(t, 3, 3, 1)
	for _, typ := range []MessageType{START, METRICS_ACK, GROUP_DONE} {
		p := clientAcademicPacket(g, typ, 1)
		p.MessageID += 10
		before := g.clients[0]
		if err := clientDeliver(t, g, p, 100); err == nil || g.clients[0] != before {
			t.Fatal("uncorrelated control changed client", typ, err)
		}
	}
	if err := clientDeliver(t, g, clientAcademicPacket(g, GROUP_DONE, 1), 101); err != ErrNotRegistered {
		t.Fatal("GROUP_DONE before START accepted", err)
	}
}

func TestClientAcademicCoordinatesAndRTTOverflow(t *testing.T) {
	a := automaticPosition(42, 9, 1, 1)
	if a != automaticPosition(42, 9, 1, 1) || a == automaticPosition(42, 9, 1, 2) || a == automaticPosition(42, 9, 2, 1) {
		t.Fatal("coordinates not deterministic or identity-sensitive")
	}
	for id := uint32(1); id < 1000; id++ {
		p := automaticPosition(math.MaxUint64, math.MaxUint64, id, math.MaxUint32)
		for _, component := range [3]float32{p.X, p.Y, p.Z} {
			if !finite(component) || component < -1000 || component >= 1000 {
				t.Fatal("invalid automatic coordinate", p)
			}
		}
	}
	g := academicClients(t, 3, 3, 1)
	w := &g.workers[0]
	if err := clientDeliver(t, g, clientAcademicPacket(g, START, 1), 100); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	job, _ := w.data.pop()
	g.claimOutbound(job)
	g.admitPosition()
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil {
		t.Fatal(err)
	}
	g.clients[0].RTTSum = math.MaxUint64 - 1
	clientDeliver(t, g, clientForwardedACK(job.Packet, 2), 1002)
	if err := clientDeliver(t, g, clientForwardedACK(job.Packet, 3), 1002); err != ErrClientTimestamp || g.clients[0].Phase != ClientFailed || g.clients[0].RTTSum != math.MaxUint64-1 {
		t.Fatal("RTT accumulation wrapped", err, g.clients[0])
	}
}
