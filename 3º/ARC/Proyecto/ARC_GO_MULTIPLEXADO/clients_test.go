package protocol

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func clientsConfig(t testing.TB, n, v, iterations uint32, workers int) ClientConfig {
	t.Helper()
	run := RunConfig{42, n, v, iterations}
	m, err := NewManifest(run, []GeneratorConfig{{7, 1, n, run}})
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultClientConfig(m, 7, testEndpoint(0))
	cfg.Academic = false // Legacy transport fixtures deliberately start POSITION through the test API.
	cfg.Workers, cfg.GOMAXPROCS = workers, 0
	cfg.InputCapacity, cfg.HighCapacity, cfg.ControlCapacity, cfg.DataCapacity = 128, 128, 128, 128
	cfg.ConfirmationCapacity, cfg.CommandCapacity = 128, 128
	cfg.ReadBuffer, cfg.WriteBuffer = 256*1024, 256*1024
	return cfg
}
func bareClients(t testing.TB, cfg ClientConfig) *ClientGenerator {
	t.Helper()
	g, err := newClientGenerator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func clientRegistrationReply(g *ClientGenerator, nonce uint64) Packet {
	id, group, _ := g.cfg.Manifest.RegistrationIdentity(g.cfg.GeneratorID, nonce)
	return Packet{Header: Header{Type: REGISTER_ACK, Flags: ServerToClient, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, GroupID: group, TargetID: id, MessageID: nonce, RegistrationNonce: nonce},
		Registration: RegistrationPayload{g.cfg.Manifest.run.V, g.cfg.Manifest.run.I}}
}
func clientDeliver(t testing.TB, g *ClientGenerator, p Packet, at int64) error {
	t.Helper()
	if err := g.decodeDispatch(wire(t, p), g.cfg.Server, at); err != nil {
		return err
	}
	owner, err := g.WorkerForClient(p.TargetID)
	if err != nil {
		t.Fatal(err)
	}
	e, ok := g.workers[owner].input.pop()
	if !ok {
		t.Fatal("missing client event")
	}
	return g.workers[owner].handleIncoming(e)
}

// Pure, controlled state-machine fixture: no socket or artificial UDP retries.
// Exercise REGISTER, its write confirmation and the authoritative REGISTER_ACK.
func readyClients(t testing.TB, g *ClientGenerator) {
	t.Helper()
	for i := range g.workers {
		g.workers[i].stageRegistrations()
	}
	for turn := 0; turn < len(g.clients)+1; turn++ {
		for i := range g.workers {
			w := &g.workers[i]
			w.flushPending()
			for {
				job, ok := w.control.pop()
				if !ok {
					break
				}
				if job.Packet.Type != REGISTER || job.Packet.MessageID != job.Packet.RegistrationNonce {
					t.Fatal("incorrect REGISTER", job)
				}
				g.admitRegister()
				if err := w.handleConfirmation(clientConfirmation{MessageID: job.Packet.MessageID, SentAt: 10,
					ClientID: job.ClientID, Type: REGISTER, Ticket: job.Ticket, Attempt: job.Attempt}); err != nil {
					t.Fatal(err)
				}
				if err := clientDeliver(t, g, clientRegistrationReply(g, job.Packet.RegistrationNonce), 20); err != nil {
					t.Fatal(err)
				}
			}
		}
		if g.counters.registered.Load() == uint64(len(g.clients)) {
			return
		}
	}
	t.Fatal("registration fixture did not converge")
}
func clientBegin(t testing.TB, g *ClientGenerator, id, iteration uint32) (*clientWorker, clientOutbound) {
	t.Helper()
	owner, err := g.WorkerForClient(id)
	if err != nil {
		t.Fatal(err)
	}
	w := &g.workers[owner]
	if err = w.beginPosition(clientCommand{id, iteration, PositionPayload{1, -2, 0.5}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	job, ok := w.data.pop()
	if !ok || job.ClientID != id {
		t.Fatal("missing own POSITION")
	}
	g.admitPosition() // controlled fixture reserves the same first-send slot as writer
	return w, job
}
func clientPositionConfirmation(job clientOutbound, at int64) clientConfirmation {
	return clientConfirmation{MessageID: job.Packet.MessageID, SentAt: at, ClientID: job.ClientID,
		Iteration: job.Packet.Iteration, Type: job.Packet.Type, Ticket: job.Ticket, Attempt: job.Attempt}
}
func clientForwardedPosition(g *ClientGenerator, origin, target uint32) Packet {
	return Packet{Header: Header{Type: POSITION, Flags: ServerToClient, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, GroupID: 1 + (origin-1)/g.cfg.Manifest.run.V, OriginID: origin,
		TargetID: target, Iteration: 1, MessageID: 999}, Position: PositionPayload{2, 3, 4}}
}
func clientForwardedACK(position Packet, neighbor uint32) Packet {
	return Packet{Header: Header{Type: ACK, Flags: ServerToClient, RunID: position.RunID,
		GeneratorID: position.GeneratorID, GroupID: position.GroupID, OriginID: neighbor,
		TargetID: position.OriginID, Iteration: position.Iteration, MessageID: position.MessageID}}
}

func TestClientPreallocationAndGroupOwnership(t *testing.T) {
	run := RunConfig{42, 150, 25, 2}
	m, err := NewManifest(run, []GeneratorConfig{{7, 1, 75, run}, {9, 76, 75, run}})
	if err != nil {
		t.Fatal(err)
	}
	cfg := clientsConfig(t, 75, 25, 2, 2)
	cfg.Manifest, cfg.GeneratorID = m, 9
	g := bareClients(t, cfg)
	if len(g.clients) != 75 || cap(g.clients) != 75 {
		t.Fatal("state not allocated exactly for the local range")
	}
	for id := uint32(76); id <= 150; id++ {
		index, err := g.ClientIndex(id)
		owner, ownerErr := g.WorkerForClient(id)
		if err != nil || ownerErr != nil || index != int(id-76) || owner != int((id-76)/25%2) {
			t.Fatalf("id=%d index=%d owner=%d errors=%v/%v", id, index, owner, err, ownerErr)
		}
		c, err := g.workers[owner].client(id)
		if err != nil || c != &g.clients[index] || c.RegistrationNonce != uint64(index+1) || c.Registered || c.SentAt != -1 {
			t.Fatal("identity inference registered a client or made auxiliary state", id)
		}
		if _, err = g.workers[1-owner].client(id); err != ErrIdentity {
			t.Fatal("foreign worker accessed client", err)
		}
	}
	for _, id := range []uint32{0, 1, 75, 151, math.MaxUint32} {
		if _, err := g.ClientIndex(id); err != ErrIdentity {
			t.Fatal("out of range accepted", id, err)
		}
	}
	if g.workers[0].pending.stats().Capacity != 50 || g.workers[1].pending.stats().Capacity != 25 {
		t.Fatal("pending ring does not match owned clients")
	}
}

func TestClientRegisterAndAuthoritativeReply(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 25, 25, 2, 6))
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	job, ok := w.control.pop()
	if !ok || job.ClientID != 1 || job.Packet != (Packet{Header: Header{Type: REGISTER, RunID: 42,
		GeneratorID: 7, MessageID: 1, RegistrationNonce: 1}}) {
		t.Fatalf("REGISTER: %+v", job)
	}
	if err := g.cfg.Manifest.Validate(&job.Packet); err != nil {
		t.Fatal(err)
	}
	c := &g.clients[0]
	if c.Registered || w.beginPosition(clientCommand{1, 1, PositionPayload{}}) != ErrNotRegistered {
		t.Fatal("registration inferred without authority")
	}
	g.admitRegister()
	if err := w.handleConfirmation(clientConfirmation{MessageID: 1, SentAt: 100, ClientID: 1, Type: REGISTER, Ticket: job.Ticket}); err != nil {
		t.Fatal(err)
	}
	if c.Registered || c.Phase != ClientRegisterWait || g.registerInFlight.Load() != 1 {
		t.Fatal("write confirmation was mistaken for REGISTER_ACK")
	}
	reply := clientRegistrationReply(g, 1)
	// Reader only queues; it must not mutate ClientState.
	if err := g.decodeDispatch(wire(t, reply), g.cfg.Server, 110); err != nil {
		t.Fatal(err)
	}
	if c.Registered {
		t.Fatal("reader mutated client")
	}
	e, _ := w.input.pop()
	if err := w.handleIncoming(e); err != nil {
		t.Fatal(err)
	}
	if !c.Registered || c.Phase != ClientReady || g.counters.registered.Load() != 1 || g.registerInFlight.Load() != 0 {
		t.Fatal("registration not committed exactly once")
	}
	before := *c
	if err := clientDeliver(t, g, reply, 120); err != nil || *c != before || g.counters.registered.Load() != 1 || g.counters.duplicateRegisterAcks.Load() != 1 {
		t.Fatal("duplicate REGISTER_ACK changed progress", err)
	}
}

func TestClientRegisterACKBeforeWriteConfirmation(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 3, 3, 2, 1))
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	job, _ := w.control.pop()
	g.admitRegister()
	if err := clientDeliver(t, g, clientRegistrationReply(g, 1), 200); err != nil {
		t.Fatal(err)
	}
	if g.registerInFlight.Load() != 1 || !g.clients[0].Registered {
		t.Fatal("window released before confirming admission")
	}
	_, own := clientBegin(t, g, 1, 1)
	if err := w.handleConfirmation(clientConfirmation{MessageID: job.Packet.MessageID, SentAt: 100, ClientID: 1, Type: REGISTER, Ticket: job.Ticket}); err != nil {
		t.Fatal(err)
	}
	if g.clients[0].Phase != ClientPositionQueued || g.registerInFlight.Load() != 0 || g.clients[0].SentAt != -1 {
		t.Fatal("late REGISTER confirmation overwrote POSITION state")
	}
	if err := w.handleConfirmation(clientPositionConfirmation(own, 300)); err != nil {
		t.Fatal(err)
	}
}

func TestClientRejectContradictoryRegistration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Packet)
	}{
		{"nonce", func(p *Packet) { p.RegistrationNonce = 2 }},
		{"message", func(p *Packet) { p.MessageID++ }},
		{"target", func(p *Packet) { p.TargetID = 2 }},
		{"group", func(p *Packet) { p.GroupID = 2 }},
		{"run", func(p *Packet) { p.RunID++ }},
		{"generator", func(p *Packet) { p.GeneratorID++ }},
		{"V", func(p *Packet) { p.Registration.V++ }},
		{"I", func(p *Packet) { p.Registration.I++ }},
		{"direction", func(p *Packet) { p.Flags = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bareClients(t, clientsConfig(t, 50, 25, 2, 2))
			readyClients(t, g)
			before := g.clients[0]
			p := clientRegistrationReply(g, 1)
			tc.change(&p)
			if err := g.workers[0].handleIncoming(clientInbound{p, 100}); err == nil || g.clients[0] != before {
				t.Fatal("contradictory REGISTER_ACK accepted", err)
			}
			if g.counters.registered.Load() != 50 {
				t.Fatal("contradictory registration changed progress")
			}
		})
	}
}

func TestClientNeighborPositionProducesRealACK(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 50, 25, 2, 2))
	readyClients(t, g)
	for _, typ := range []MessageType{POSITION, RETRY_POSITION} {
		p := clientForwardedPosition(g, 1, 25)
		p.Type = typ
		if typ == RETRY_POSITION {
			p.MissingMask, p.RetryCounter = 1<<24, 2
		}
		before := g.clients[24]
		if err := clientDeliver(t, g, p, 100); err != nil {
			t.Fatal(err)
		}
		job, ok := g.workers[0].high.pop()
		want := Packet{Header: Header{Type: ACK, RunID: 42, GeneratorID: 7, GroupID: 1,
			OriginID: 25, TargetID: 1, Iteration: 1, MessageID: p.MessageID, RetryCounter: p.RetryCounter}}
		if !ok || job.Packet != want || job.ClientID != 0 || g.workers[0].high.stats().Depth != 0 || g.clients[24] != before {
			t.Fatalf("neighbor ACK: %+v; own state changed=%v", job, g.clients[24] != before)
		}
		if err := g.cfg.Manifest.Validate(&job.Packet); err != nil {
			t.Fatal(err)
		}
	}
}

func TestClientRejectWrongNeighborPosition(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 50, 25, 2, 2))
	readyClients(t, g)
	for _, change := range []func(*Packet){
		func(p *Packet) { p.TargetID = 51 }, func(p *Packet) { p.OriginID = 26 },
		func(p *Packet) { p.GroupID = 2 }, func(p *Packet) { p.OriginID = p.TargetID },
		func(p *Packet) { p.Iteration = 0 }, func(p *Packet) { p.Iteration = 3 },
		func(p *Packet) { p.MessageID = 0 }, func(p *Packet) { p.RunID++ },
		func(p *Packet) { p.GeneratorID = 9 }, func(p *Packet) { p.Flags = 0 },
		func(p *Packet) { p.Position.X = float32(math.NaN()) },
	} {
		p := clientForwardedPosition(g, 1, 2)
		change(&p)
		if err := g.workers[0].handleIncoming(clientInbound{p, 100}); err == nil {
			t.Fatal("wrong POSITION accepted", p)
		}
	}
	if g.workers[0].high.stats().Depth != 0 {
		t.Fatal("invalid POSITION generated ACK")
	}
	// Correct global packet addressed to another generator's valid range.
	m := fixtureManifest(t, 25)
	cfg := clientsConfig(t, 25, 25, 5, 1)
	cfg.Manifest = m
	other := bareClients(t, cfg)
	p := clientForwardedPosition(other, 26, 27)
	p.GeneratorID = 9
	if err := other.decodeDispatch(wire(t, p), other.cfg.Server, 100); err != ErrIdentity {
		t.Fatal("another generator was accepted", err)
	}
	unregistered := bareClients(t, clientsConfig(t, 25, 25, 1, 1))
	if err := clientDeliver(t, unregistered, clientForwardedPosition(unregistered, 1, 2), 100); err != ErrNotRegistered {
		t.Fatal(err)
	}
}

func TestClientACKMasksAndIterationBarrier(t *testing.T) {
	for _, v := range []uint32{1, 25, 32} {
		t.Run(fmt.Sprintf("V%d", v), func(t *testing.T) {
			g := bareClients(t, clientsConfig(t, v, v, 2, 1))
			readyClients(t, g)
			w, job := clientBegin(t, g, 1, 1)
			c := &g.clients[0]
			if c.SentAt != -1 || c.AckMask != 0 || c.Phase != ClientPositionQueued {
				t.Fatal("RTT started when enqueued")
			}
			if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil {
				t.Fatal(err)
			}
			if v == 1 {
				if c.ExpectedMask != 0 || c.Phase != ClientComplete || c.LastRTTNS != 0 || c.CompletedIterations != 1 {
					t.Fatal("V1 waited for nonexistent neighbor", *c)
				}
			} else {
				for neighbor := uint32(2); neighbor < v; neighbor++ {
					p := clientForwardedACK(job.Packet, neighbor)
					if err := clientDeliver(t, g, p, 2000+int64(neighbor)); err != nil {
						t.Fatal(err)
					}
					if c.AckMask&(uint32(1)<<(neighbor-1)) == 0 || c.Phase == ClientComplete {
						t.Fatal("ACK bit or premature completion", *c)
					}
					before := *c
					if err := clientDeliver(t, g, p, 999999); err != nil || *c != before {
						t.Fatal("duplicate ACK changed state/timestamps", err)
					}
				}
				if bits.OnesCount32(c.AckMask) != int(v-2) || w.beginPosition(clientCommand{1, 2, PositionPayload{}}) != ErrClientBusy {
					t.Fatal("one missing ACK did not hold iteration barrier")
				}
				missing, err := MissingMask(c.ExpectedMask, c.AckMask)
				if err != nil || missing != uint32(1)<<(v-1) {
					t.Fatal("missing mask", missing, err)
				}
				if err := clientDeliver(t, g, clientForwardedACK(job.Packet, v), 3000); err != nil {
					t.Fatal(err)
				}
				if c.Phase != ClientComplete || c.AckMask != c.ExpectedMask || bits.OnesCount32(c.AckMask) != int(v-1) || c.LastRTTNS != 2000 {
					t.Fatal("last unique ACK did not complete operation", *c)
				}
				if v == 32 && c.AckMask&(uint32(1)<<31) == 0 {
					t.Fatal("bit 31 lost")
				}
			}
			_, next := clientBegin(t, g, 1, 2)
			if next.Packet.MessageID == job.Packet.MessageID || c.AckMask != 0 || c.SentAt != -1 {
				t.Fatal("next operation did not get independent correlation/t0")
			}
			if v != 1 {
				if err := clientDeliver(t, g, clientForwardedACK(job.Packet, 2), 5000); err == nil || c.AckMask != 0 {
					t.Fatal("late ACK from previous iteration accepted")
				}
			}
		})
	}
}

func TestClientRejectUncorrelatedACK(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 50, 25, 2, 2))
	readyClients(t, g)
	w, job := clientBegin(t, g, 1, 1)
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Packet){
		func(p *Packet) { p.OriginID = 1 }, func(p *Packet) { p.OriginID = 26 },
		func(p *Packet) { p.TargetID = 2 }, func(p *Packet) { p.GroupID = 2 },
		func(p *Packet) { p.MessageID++ }, func(p *Packet) { p.Iteration = 2 },
		func(p *Packet) { p.RunID++ }, func(p *Packet) { p.Flags = 0 },
	} {
		p := clientForwardedACK(job.Packet, 2)
		change(&p)
		before := g.clients[0]
		// Some contradictions (self-ACK) are rejected by Encode itself. Exercise
		// the owner's defensive validation without requiring an invalid encoding.
		if err := w.handleIncoming(clientInbound{p, 2000}); err == nil || g.clients[0] != before {
			t.Fatal("uncorrelated ACK modified operation", err)
		}
	}
	if g.counters.uniqueAcks.Load() != 0 {
		t.Fatal("invalid ACK counted as progress")
	}
}

func TestClientACKBeforeWriterConfirmationRTT(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 25, 25, 2, 1))
	readyClients(t, g)
	w, job := clientBegin(t, g, 1, 1)
	c := &g.clients[0]
	// Deliver in reverse timestamp order: completion uses max UNIQUE receive time.
	for neighbor := uint32(2); neighbor <= 25; neighbor++ {
		if err := clientDeliver(t, g, clientForwardedACK(job.Packet, neighbor), 4000-int64(neighbor)); err != nil {
			t.Fatal(err)
		}
	}
	if c.AckMask != c.ExpectedMask || c.SentAt != -1 || c.CompletedIterations != 0 || c.Phase == ClientComplete || c.LastUniqueACKAt != 3998 {
		t.Fatal("ACK-before-confirmation race completed RTT without t0", *c)
	}
	if err := clientDeliver(t, g, clientForwardedACK(job.Packet, 2), 10000); err != nil || c.LastUniqueACKAt != 3998 {
		t.Fatal("duplicate retimed last unique ACK", err)
	}
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != nil {
		t.Fatal(err)
	}
	if c.Phase != ClientComplete || c.LastRTTNS != 2998 || c.RTTSum != 2998 || g.counters.rttSamples.Load() != 1 {
		t.Fatal("RTT correlation failed", *c)
	}
	before := *c
	if err := w.handleConfirmation(clientPositionConfirmation(job, 500)); err != nil || *c != before {
		t.Fatal("duplicate confirmation restarted t0", err)
	}
	if err := clientDeliver(t, g, clientForwardedACK(job.Packet, 25), 20000); err != nil || *c != before || g.counters.completed.Load() != 1 {
		t.Fatal("completed operation progressed twice", err)
	}
}

func TestClientFailedAndInvalidWriteConfirmations(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 3, 3, 2, 1))
	readyClients(t, g)
	w, job := clientBegin(t, g, 1, 1)
	for _, change := range []func(*clientConfirmation){
		func(c *clientConfirmation) { c.MessageID++ }, func(c *clientConfirmation) { c.Iteration++ },
		func(c *clientConfirmation) { c.ClientID = 4 }, func(c *clientConfirmation) { c.Status = 99 },
		func(c *clientConfirmation) { c.Type = ACK }, func(c *clientConfirmation) { c.SentAt = -1 },
	} {
		tx := clientPositionConfirmation(job, 1000)
		change(&tx)
		before := g.clients[0]
		if err := w.handleConfirmation(tx); err == nil || g.clients[0] != before {
			t.Fatal("invalid confirmation changed state", err)
		}
	}
	tx := clientPositionConfirmation(job, 1000)
	tx.Status = clientSendFailed
	if err := w.handleConfirmation(tx); err != nil || g.clients[0].Phase != ClientFailed || g.counters.failedOperations.Load() != 1 {
		t.Fatal("failed write not explicit", err)
	}
	if err := clientDeliver(t, g, clientForwardedACK(job.Packet, 2), 2000); err == nil {
		t.Fatal("failed operation accepted progress")
	}
	reg := bareClients(t, clientsConfig(t, 1, 1, 1, 1))
	reg.workers[0].stageRegistrations()
	reg.workers[0].flushPending()
	reg.workers[0].control.pop()
	reg.admitRegister()
	if err := reg.workers[0].handleConfirmation(clientConfirmation{MessageID: 1, SentAt: -1, ClientID: 1, Type: REGISTER, Status: clientSendFailed, Ticket: reg.clients[0].SendTicket}); err != nil {
		t.Fatal(err)
	}
	if reg.clients[0].Registered || reg.clients[0].Phase != ClientFailed || reg.registerInFlight.Load() != 0 || reg.counters.failedRegistrations.Load() != 1 {
		t.Fatal("failed REGISTER was accepted or leaked window")
	}
}

func TestClientRejectImpossibleRTT(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 3, 3, 1, 1))
	readyClients(t, g)
	w, job := clientBegin(t, g, 1, 1)
	for _, id := range []uint32{2, 3} {
		if err := clientDeliver(t, g, clientForwardedACK(job.Packet, id), 100); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.handleConfirmation(clientPositionConfirmation(job, 1000)); err != ErrClientTimestamp || g.clients[0].Phase != ClientFailed || g.counters.completed.Load() != 0 {
		t.Fatal("impossible timestamp produced underflow/false completion", err)
	}
}

func TestClientBoundedQueuesAndPendingBackpressure(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 1, 1)
	cfg.InputCapacity, cfg.HighCapacity, cfg.DataCapacity, cfg.CommandCapacity, cfg.ControlCapacity = 1, 1, 1, 1, 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w := &g.workers[0]
	p := clientForwardedPosition(g, 1, 2)
	data := wire(t, p)
	if err := g.decodeDispatch(data, cfg.Server, 100); err != nil {
		t.Fatal(err)
	}
	beforeG := runtime.NumGoroutine()
	allocs := testing.AllocsPerRun(1000, func() {
		if err := g.decodeDispatch(data, cfg.Server, 100); err != ErrQueueFull {
			panic("full inbound queue accepted packet")
		}
	})
	if allocs != 0 || runtime.NumGoroutine() > beforeG+2 || w.input.stats().Capacity != 1 || w.input.stats().Maximum != 1 {
		t.Fatal("full input grew memory/goroutines", allocs)
	}
	w.input.pop()
	if err := w.handleIncoming(clientInbound{p, 100}); err != nil {
		t.Fatal(err)
	}
	if err := w.handleIncoming(clientInbound{p, 100}); err != ErrQueueFull || w.high.stats().Drops != 1 {
		t.Fatal("reactive ACK overflow not counted", err)
	}
	if err := g.StartPosition(1, 1, PositionPayload{}); err != nil {
		t.Fatal(err)
	}
	if err := g.StartPosition(2, 1, PositionPayload{}); err != ErrQueueFull {
		t.Fatal("command overflow not immediate", err)
	}
	w.commands.pop()
	if err := w.beginPosition(clientCommand{1, 1, PositionPayload{}}); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	if err := w.beginPosition(clientCommand{2, 1, PositionPayload{}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		w.flushPending()
	}
	if w.pending.stats().Depth != 1 || w.data.stats().Drops != 0 || w.data.stats().FullAttempts != 1000 || g.clients[1].Flags&clientPositionIntent == 0 {
		t.Fatal("full DATA did not retain one bounded intention", w.data.stats(), w.pending.stats())
	}
	w.data.pop()
	w.flushPending()
	job, ok := w.data.pop()
	if !ok || job.ClientID != 2 || w.pending.stats().Depth != 0 {
		t.Fatal("pending POSITION not resumed after backpressure")
	}
}

func TestClientFailedPushesConsumeWorkBudget(t *testing.T) {
	cfg := clientsConfig(t, 100, 25, 1, 1)
	cfg.ControlCapacity, cfg.WorkBudget = 1, 4
	g := bareClients(t, cfg)
	w := &g.workers[0]
	w.stageRegistrations()
	w.flushPending()
	before := w.control.stats().FullAttempts
	if w.flushPending() || w.control.stats().FullAttempts-before != 4 || w.pending.stats().Depth != 99 {
		t.Fatal("failed pushes escaped work budget or duplicated intentions")
	}
}

func TestClientConfirmationReservation(t *testing.T) {
	cfg := clientsConfig(t, 3, 3, 1, 1)
	cfg.ConfirmationCapacity = 1
	g := bareClients(t, cfg)
	readyClients(t, g)
	w, _ := clientBegin(t, g, 1, 1)
	// Place another own POSITION back in DATA for the writer selector.
	w.data.push(w.positionJob(1, &g.clients[0]), false)
	w.confirmations.push(clientConfirmation{MessageID: 99}, false)
	if _, _, ok, _ := g.takeOutbound(100); ok || w.data.stats().Depth != 1 {
		t.Fatal("writer attempted send with no space for t0 confirmation")
	}
	w.high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	job, _, ok, _ := g.takeOutbound(100)
	if !ok || job.Packet.Type != ACK {
		t.Fatal("confirmation backpressure blocked reactive ACK")
	}
	w.confirmations.pop()
	job, _, ok, _ = g.takeOutbound(100)
	if !ok || job.Packet.Type != POSITION || !g.publishConfirmation(0, clientPositionConfirmation(job, 101)) {
		t.Fatal("reserved confirmation could not be published")
	}
	if w.confirmations.stats().Drops != 0 || w.confirmations.stats().Maximum != 1 {
		t.Fatal("confirmation discarded silently")
	}
}

func TestClientWriterPriorityPacingAndWindow(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 25, 25, 1, 1))
	w := &g.workers[0]
	for i := 0; i < 18; i++ {
		w.high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	}
	for i := 0; i < 4; i++ {
		w.control.push(clientOutbound{Packet: Packet{Header: Header{Type: REGISTER}}}, false)
	}
	w.data.push(clientOutbound{Packet: Packet{Header: Header{Type: POSITION}}}, false)
	for i := 0; i < 9; i++ {
		job, _, ok, _ := g.takeOutbound(100)
		if !ok || (i == 4 && job.Packet.Type != REGISTER) || (i == 8 && job.Packet.Type != POSITION) {
			t.Fatal("ACK/control priority or DATA fairness violated", i, job.Packet.Type)
		}
	}
	// Empty all queues, then show paced REGISTER cannot hold ACK at its head.
	w.high.discardAll()
	w.control.discardAll()
	w.data.discardAll()
	w.control.push(clientOutbound{Packet: Packet{Header: Header{Type: REGISTER}}}, false)
	g.nextRegister = 200
	if _, _, ok, wait := g.takeOutbound(100); ok || wait != 200 {
		t.Fatal("REGISTER pacing ignored", ok, wait)
	}
	w.high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	if job, _, ok, _ := g.takeOutbound(100); !ok || job.Packet.Type != ACK {
		t.Fatal("paced control delayed ACK")
	}
	g.registerInFlight.Store(int64(g.cfg.RegisterWindow))
	if _, _, ok, _ := g.takeOutbound(300); ok {
		t.Fatal("REGISTER window bypassed")
	}
	g.registerInFlight.Add(-1)
	if job, _, ok, _ := g.takeOutbound(300); !ok || job.Packet.Type != REGISTER {
		t.Fatal("REGISTER not resumed when window released")
	}
	w.data.push(clientOutbound{Packet: Packet{Header: Header{Type: POSITION}}}, false)
	g.positions.take(300)
	wantReady := g.positions.ready(400)
	if _, _, ok, wait := g.takeOutbound(400); ok || wait != wantReady {
		t.Fatal("own POSITION pacing ignored", ok, wait)
	}
}

func TestClientWriterRoundRobin(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 12, 3, 1, 2))
	for i := range g.workers {
		for j := 0; j < 3; j++ {
			g.workers[i].high.push(clientOutbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
		}
	}
	for i := 0; i < 6; i++ {
		_, owner, ok, _ := g.takeOutbound(0)
		if !ok || owner != i%2 {
			t.Fatal("writer did not alternate eligible workers", i, owner)
		}
	}
}

func TestClientMalformedPacketsAndSource(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 25, 25, 1, 1))
	valid := wire(t, clientRegistrationReply(g, 1))
	for n := 0; n < len(valid); n++ {
		if err := g.decodeDispatch(valid[:n], g.cfg.Server, 100); err == nil {
			t.Fatal("truncated REGISTER_ACK accepted", n)
		}
	}
	for _, data := range [][]byte{make([]byte, MaxPacketSize+1), make([]byte, 65535), []byte("garbage")} {
		if err := g.decodeDispatch(data, g.cfg.Server, 100); err == nil {
			t.Fatal("malformed accepted")
		}
	}
	for _, offset := range []int{4, 5, 6, 60} {
		buf := append([]byte(nil), valid...)
		buf[offset] = 255
		if err := g.decodeDispatch(buf, g.cfg.Server, 100); err == nil {
			t.Fatal("invalid version/type/flags/reserved field accepted", offset)
		}
	}
	buf := append([]byte(nil), valid...)
	binary.BigEndian.PutUint64(buf[8:16], 43)
	if err := g.decodeDispatch(buf, g.cfg.Server, 100); err == nil {
		t.Fatal("wrong RunID accepted")
	}
	for _, source := range []netip.AddrPort{testEndpoint(1), {}} {
		if err := g.decodeDispatch(valid, source, 100); err != ErrEndpoint {
			t.Fatal(err)
		}
	}
	if err := g.decodeDispatch(valid, g.cfg.Server, -1); err != ErrClientTimestamp {
		t.Fatal(err)
	}
	for _, typ := range []MessageType{START, METRICS_ACK, GROUP_DONE} {
		p := fixturePacket(typ)
		if err := g.decodeDispatch(wire(t, p), g.cfg.Server, 100); err != ErrType {
			t.Fatal("future control was accepted in phase 3", typ, err)
		}
	}
	if g.workers[0].input.stats().Depth != 0 || g.clients[0].Registered {
		t.Fatal("reader mutated state/queued rejected packet")
	}
}

func TestClientInvalidConfigAndCommands(t *testing.T) {
	for _, change := range []func(*ClientConfig){
		func(c *ClientConfig) { c.Manifest = nil }, func(c *ClientConfig) { c.GeneratorID = 0 },
		func(c *ClientConfig) { c.Workers = 0 }, func(c *ClientConfig) { c.GOMAXPROCS = -1 },
		func(c *ClientConfig) { c.InputCapacity = 0 }, func(c *ClientConfig) { c.HighCapacity = 0 },
		func(c *ClientConfig) { c.ControlCapacity = 0 }, func(c *ClientConfig) { c.DataCapacity = 0 },
		func(c *ClientConfig) { c.ConfirmationCapacity = 0 }, func(c *ClientConfig) { c.CommandCapacity = 0 },
		func(c *ClientConfig) { c.WorkBudget = 0 }, func(c *ClientConfig) { c.HighBurst = 0 },
		func(c *ClientConfig) { c.AckBurst = 0 }, func(c *ClientConfig) { c.RegisterWindow = 0 },
		func(c *ClientConfig) { c.RegisterRate = 0 }, func(c *ClientConfig) { c.PositionRate = 0 },
		func(c *ClientConfig) { c.RegisterRate = int(time.Second) + 1 }, func(c *ClientConfig) { c.PositionRate = int(time.Second) + 1 },
		func(c *ClientConfig) { c.ReadBuffer = -1 }, func(c *ClientConfig) { c.WriteBuffer = -1 },
		func(c *ClientConfig) { c.WriteTimeout = 0 }, func(c *ClientConfig) { c.Server = netip.AddrPort{} },
		func(c *ClientConfig) { c.Server = netip.MustParseAddrPort("127.0.0.1:0") },
		func(c *ClientConfig) { c.Server = netip.MustParseAddrPort("0.0.0.0:9000") },
	} {
		cfg := clientsConfig(t, 3, 3, 1, 1)
		change(&cfg)
		if _, err := newClientGenerator(cfg); err != ErrClientConfig {
			t.Fatal("bad configuration accepted", err)
		}
	}
	cfg := clientsConfig(t, 3, 3, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ListenClients(ctx, cfg); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := ListenClients(nil, cfg); err != ErrClientConfig {
		t.Fatal(err)
	}
	g := bareClients(t, cfg)
	for _, tc := range []struct {
		id, iteration uint32
		pos           PositionPayload
		want          error
	}{
		{0, 1, PositionPayload{}, ErrIdentity}, {4, 1, PositionPayload{}, ErrIdentity},
		{1, 0, PositionPayload{}, ErrIteration}, {1, 2, PositionPayload{}, ErrIteration},
		{1, 1, PositionPayload{X: float32(math.Inf(1))}, ErrPayload},
	} {
		if err := g.StartPosition(tc.id, tc.iteration, tc.pos); err != tc.want {
			t.Fatal(err, tc.want)
		}
	}
	readyClients(t, g)
	if err := g.workers[0].beginPosition(clientCommand{1, 0, PositionPayload{}}); err != ErrIteration {
		t.Fatal(err)
	}
	if err := g.workers[0].beginPosition(clientCommand{1, 1, PositionPayload{Z: float32(math.NaN())}}); err != ErrPayload {
		t.Fatal(err)
	}
}

// Model cancellation just AFTER StartPosition's initial context observation.
// Err's first result was read before cancel: this ordering can occur in real use.
type clientCancelAfterObservation struct {
	context.Context
	cancel context.CancelFunc
	checks atomic.Uint32
}

func (c *clientCancelAfterObservation) Err() error {
	err := c.Context.Err()
	if c.checks.Add(1) == 1 {
		c.cancel()
	}
	return err
}

func TestClientCommandCannotRepopulateAfterCancellation(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 3, 3, 1, 1))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g.ctx = &clientCancelAfterObservation{Context: ctx, cancel: cancel}
	if err := g.StartPosition(1, 1, PositionPayload{}); !errors.Is(err, context.Canceled) {
		t.Fatal("command admitted after cancellation raced the first check", err)
	}
	if st := g.workers[0].commands.stats(); st.Depth != 0 || st.Drops != 0 {
		t.Fatal("cancelled submission repopulated the queue or counted packet loss", st)
	}
}

func TestClientOwnerCleansOnlyItsGroups(t *testing.T) {
	g := bareClients(t, clientsConfig(t, 12, 3, 1, 2))
	readyClients(t, g)
	for _, id := range []uint32{1, 4, 7, 10} {
		clientBegin(t, g, id, 1)
	}
	before := append([]ClientState(nil), g.clients...)
	g.workers[0].finish()
	for i := range g.clients {
		owner, _ := g.WorkerForClient(uint32(i + 1))
		if owner == 1 && g.clients[i] != before[i] {
			t.Fatal("shutdown owner changed another worker's state", i)
		}
		if owner == 0 && before[i].Phase == ClientPositionQueued && g.clients[i].Phase != ClientStopped {
			t.Fatal("own unfinished operation not marked stopped", i)
		}
	}
	if g.counters.abortedOperations.Load() != 2 {
		t.Fatal("aborted operations not accounted per owner")
	}
}

func TestClientStructureSizes(t *testing.T) {
	state, in, out, worker := unsafe.Sizeof(ClientState{}), unsafe.Sizeof(clientInbound{}), unsafe.Sizeof(clientOutbound{}), unsafe.Sizeof(clientWorker{})
	confirmation, command, queue := unsafe.Sizeof(clientConfirmation{}), unsafe.Sizeof(clientCommand{}), unsafe.Sizeof(boundedQueue[clientInbound]{})
	if state != 120 {
		t.Fatalf("ClientState budget changed: %d bytes", state)
	}
	t.Logf("ClientState=%d inbound=%d outbound=%d worker=%d confirmation=%d command=%d queueMetadata=%d", state, in, out, worker, confirmation, command, queue)
	// Arithmetic estimate only: no 160K generator/network run is created.
	n, w := uint64(160000), uint64(6)
	stateBytes := n * uint64(state)
	pendingBytes := n * 4
	queueBytes := w * (4096*uint64(in) + (1024+128+2048+512)*uint64(out) + 256*uint64(confirmation) + 256*uint64(command))
	metadataBytes := w * (uint64(worker) + 8*uint64(queue))
	heapBytes, gateBytes := n*4, n*uint64(unsafe.Sizeof(atomic.Uint64{}))
	t.Logf("N160000 W6: state=%d pending=%d deadlines=%d gates=%d packet/confirmation/commandQueues=%d worker/queueMetadata=%d subtotal=%d bytes (not RSS)",
		stateBytes, pendingBytes, heapBytes, gateBytes, queueBytes, metadataBytes, stateBytes+pendingBytes+heapBytes+gateBytes+queueBytes+metadataBytes)
}

// Poll observable progress; recovery belongs to production ARQ, never the test.
func waitClientCondition(t *testing.T, ctx context.Context, label string, g *ClientGenerator, condition func(ClientStats) bool) ClientStats {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		st := g.Stats()
		if st.InvalidPackets+st.QueueDrops+st.SocketErrors+st.SendDrops+st.RejectedOperations+st.FailedRegistrations+st.FailedOperations+st.InvalidConfirmations > 0 {
			t.Fatalf("%s: explicit error/loss, no test retries: %+v", label, st)
		}
		if condition(st) {
			return st
		}
		select {
		case <-ctx.Done():
			t.Fatalf("%s: incomplete: %+v: %v", label, st, ctx.Err())
		case <-g.done:
			t.Fatalf("%s: generator stopped: %v", label, g.Wait())
		case <-ticker.C:
		}
	}
}
func checkCompletedClients(t *testing.T, g *ClientGenerator, v uint32) {
	t.Helper()
	// Called ONLY after Close/Wait has joined every worker.
	var sum uint64
	for i := range g.clients {
		c := &g.clients[i]
		if !c.Registered || c.Phase != ClientComplete || c.CompletedIterations != 1 || c.AckMask != c.ExpectedMask || bits.OnesCount32(c.AckMask) != int(v-1) || c.SentAt < 0 || c.LastUniqueACKAt < c.SentAt {
			t.Fatalf("client %d incomplete: %+v", g.rangeConfig.FirstID+uint32(i), *c)
		}
		if c.LastRTTNS != uint64(c.LastUniqueACKAt-c.SentAt) {
			t.Fatal("socket-based RTT not measured correctly")
		}
		sum += c.RTTSum
	}
	if st := g.Stats(); st.RTTSum != sum || st.RTTSamples != uint64(len(g.clients)) || st.AbortedOperations != 0 {
		t.Fatal("RTT/cleanup counters", st)
	}
}

func TestClientsEndToEnd100V25(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	scfg := serverConfig(t, 100, 25, 6, false)
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	cfg := clientsConfig(t, 100, 25, 1, 6)
	cfg.Manifest, cfg.Server = scfg.Manifest, s.Addr()
	cfg.RegisterWindow, cfg.RegisterRate, cfg.PositionRate = 8, 500, 100
	begin := time.Now()
	g, err := ListenClients(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	registered := waitClientCondition(t, ctx, "100 real registrations", g, func(st ClientStats) bool {
		return st.RegisteredClients == 100 && st.RegisterInFlight == 0
	})
	registrationTime := time.Since(begin)
	if !s.Stats().RegistrationBarrierReached || registered.RegisterSent != 100 || registered.RegisterAckReceived != 100 || registered.MaximumRegisterInFlight > 8 {
		t.Fatal("registration barrier/window", registered)
	}
	positionsBegin := time.Now()
	for id := uint32(1); id <= 100; id++ {
		if err := g.StartPosition(id, 1, PositionPayload{float32(id), -1, 0.5}); err != nil {
			t.Fatal(err)
		}
	}
	completed := waitClientCondition(t, ctx, "100 real POSITION fanouts", g, func(st ClientStats) bool {
		return st.CompletedOperations == 100 && st.PositionSent == 100 && st.AckSent == 2400 && st.WriterConfirmations == 200
	})
	positionsTime := time.Since(positionsBegin)
	if completed.UniqueAcks != 2400 || completed.AckReceived != 2400 || completed.PositionReceived != 2400 || completed.DuplicateAcks != 0 || completed.RegisterSent != 100 {
		t.Fatal("real neighbor progress counts", completed)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	checkCompletedClients(t, g, 25)
	sst := s.Stats()
	if sst.RegisteredClients != 100 || sst.PositionReceived != 100 || sst.PositionForwarded != 2400 || sst.AckReceived != 2400 || sst.AckForwarded != 2400 || sst.QueueDrops+sst.InvalidPackets+sst.SocketErrors+sst.SendDrops != 0 {
		t.Fatal("server did not preserve the complete real ACK path", sst)
	}
	t.Logf("N=100 V=25 I=1 W=6: registered=100/100 completed=100/100 uniqueACK=2400/2400 drops=0 errors=0 register=%s positions=%s; correctness pacing POSITION=100/s, not throughput benchmark", registrationTime, positionsTime)
	t.Logf("generator queue maxima: %+v", completed.Queues)
}

func TestClientsEndToEndConcurrentGroups(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	scfg := serverConfig(t, 12, 3, 3, true)
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	var generators []*ClientGenerator
	for _, gc := range scfg.Manifest.generators {
		cfg := clientsConfig(t, 12, 3, 1, 2)
		cfg.Manifest, cfg.GeneratorID, cfg.Server = scfg.Manifest, gc.GeneratorID, s.Addr()
		// Budgets are PER generator: totals <=128 outstanding and 1000 REGISTER/s.
		cfg.RegisterWindow, cfg.RegisterRate, cfg.PositionRate = 64, 500, 50
		g, err := ListenClients(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		generators = append(generators, g)
		t.Cleanup(func() { g.Close() })
	}
	for _, g := range generators {
		waitClientCondition(t, ctx, "multigroup register", g, func(st ClientStats) bool { return st.RegisteredClients == 6 && st.RegisterInFlight == 0 })
	}
	if !s.Stats().RegistrationBarrierReached {
		t.Fatal("server barrier not reached")
	}
	var wg sync.WaitGroup
	errs := make(chan error, len(generators)) // per GENERATOR test coordinator, not client state
	for _, g := range generators {
		wg.Add(1)
		go func(g *ClientGenerator) {
			defer wg.Done()
			for offset := uint32(0); offset < g.rangeConfig.Count; offset++ {
				if err := g.StartPosition(g.rangeConfig.FirstID+offset, 1, PositionPayload{1, 2, 3}); err != nil {
					errs <- err
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	for _, g := range generators {
		st := waitClientCondition(t, ctx, "multigroup positions", g, func(st ClientStats) bool {
			return st.CompletedOperations == 6 && st.PositionSent == 6 && st.AckSent == 12 && st.WriterConfirmations == 12
		})
		if st.UniqueAcks != 12 || st.PositionReceived != 12 || st.DuplicateAcks != 0 {
			t.Fatal(st)
		}
		if err := g.Close(); err != nil {
			t.Fatal(err)
		}
		checkCompletedClients(t, g, 3)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	st := s.Stats()
	if st.PositionForwarded != 24 || st.AckForwarded != 24 || st.QueueDrops+st.InvalidPackets+st.SocketErrors+st.SendDrops != 0 {
		t.Fatal(st)
	}
	t.Log("N=12 V=3 I=1, 4 groups, 2 generators, 2 workers/generator: 12/12 completed, 24/24 unique ACK, 3 UDP sockets total, zero drops/errors")
}

func TestClientGoroutinesScaleWithWorkersAndCleanClose(t *testing.T) {
	remote := testSocket(t) // Warm the network runtime before comparing deltas.
	var deltas []int
	for _, scenario := range []struct {
		n            uint32
		workers      int
		highCapacity int
	}{{25, 2, 128}, {250, 2, 128}, {25, 6, 128}, {25, 6, 2048}} {
		cfg := clientsConfig(t, scenario.n, 25, 1, scenario.workers)
		cfg.HighCapacity = scenario.highCapacity
		cfg.Server = remote.LocalAddr().(*net.UDPAddr).AddrPort()
		cfg.RegisterRate = 1 // This is a lifecycle test, not a registration-completion test.
		ctx, cancel := context.WithCancel(context.Background())
		before := runtime.NumGoroutine()
		g, err := ListenClients(ctx, cfg)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		// Wait for an active deadline and the W+1 shared timers before cancellation.
		deadline := time.After(2 * time.Second)
		for g.counters.confirmations.Load() == 0 || g.activeARQTimers.Load() != int64(cfg.Workers) || g.activeWriterTimers.Load() != 1 {
			select {
			case <-deadline:
				g.Close()
				cancel()
				t.Fatal("writer did not start")
			default:
				runtime.Gosched()
			}
		}
		if st := g.Stats(); st.DeadlineHeapCurrent == 0 || st.ARQTimers != int64(cfg.Workers) || st.WriterTimers != 1 {
			t.Fatal("scheduler resources did not start", st)
		}
		deltas = append(deltas, runtime.NumGoroutine()-before)
		cancel()
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		if err := g.Close(); err != nil { // idempotent
			t.Fatal(err)
		}
		if err := g.StartPosition(1, 1, PositionPayload{}); !errors.Is(err, context.Canceled) {
			t.Fatal("closed generator admitted command", err)
		}
		if st := g.Stats(); st.RegisterInFlight != 0 || st.InflightCurrent != 0 || st.DeadlineHeapCurrent != 0 || st.ARQTimers != 0 || st.WriterTimers != 0 {
			t.Fatal("shutdown leaked windows, heap or timers", st)
		}
		for _, q := range g.Stats().Queues {
			if q.High.Capacity != scenario.highCapacity {
				t.Fatal("HIGH capacity did not survive lifecycle", q.High)
			}
			if q.Input.Depth+q.High.Depth+q.Control.Depth+q.Data.Depth+q.Retry.Depth+q.Confirmations.Depth+q.Commands.Depth+q.PendingClients.Depth != 0 {
				t.Fatal("shutdown did not drain bounded queues", q)
			}
		}
		// The port is released, and no old worker can touch state after Wait.
		rebound, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(g.Addr()))
		if err != nil {
			t.Fatal("socket was not released", err)
		}
		rebound.Close()
		for _, c := range g.clients {
			if c.Phase != ClientStopped || c.Registered {
				t.Fatal("pending registration was not stopped", c)
			}
		}
	}
	if deltas[1] > deltas[0]+4 || deltas[1] > 2+3+6 {
		t.Fatal("goroutines grew with clients", deltas)
	}
	if deltas[3] > deltas[2]+4 {
		t.Fatal("HIGH2048 introduced extra goroutines", deltas)
	}
	t.Logf("W2/N25, W2/N250, W6/N25/H128, W6/N25/H2048 observed goroutine deltas=%v; own count W+3; ARQ timers W, writer timer 1; active deadlines cancelled, timers stopped, socket released and all queues drained", deltas)
}
