package protocol

import (
	"context"
	"errors"
	"fmt"
	"math/bits"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"
)

// Fault injection is deliberately confined to _test.go. Each proxy has exactly
// two UDP sockets, two readers and one scheduler/writer, independent of N. The
// server binds the generator to the stable backend endpoint, while the generator
// accepts replies only from the stable frontend endpoint. No production hook is
// required, and every delivered neighbor ACK is still a real UDP datagram.
type lossDirection uint8

const (
	lossToServer lossDirection = iota
	lossToClients
	lossProxyInputCapacity       = 1024
	lossProxyDelayCapacity       = 64
	lossProxyObservationCapacity = 128
)

type lossRule struct {
	Direction        lossDirection
	Type             MessageType
	Origin, Target   uint32 // zero matches any identity
	Group, Iteration uint32 // zero matches any group/iteration
	Nonce            uint64 // zero matches any nonce
	Counter          uint16
	AnyCounter       bool
	Limit            uint32 // zero means every match; otherwise only the first Limit matches
	Drop             bool
	Duplicate        bool
	Delay            time.Duration
	matched          uint32 // owned exclusively by the proxy scheduler
}

type lossEvent struct {
	Data       [MaxPacketSize]byte
	Length     uint8
	Direction  lossDirection
	Source     netip.AddrPort
	ReceivedAt int64
}

type lossDelayed struct {
	Event    lossEvent
	Due      int64
	Copies   uint8
	Occupied bool
}

type lossObservation struct {
	Initial                     Packet
	FirstRetry                  Packet
	InitialAt, FirstRetryAt     int64
	InitialSeen, FirstRetrySeen bool
	Retries                     uint32
}

type lossProxyStats struct {
	Received, Delivered, Dropped, Duplicated, Delayed uint64
	InitialDataObserved, InitialDataDropped           uint64
	RuleMatches                                       [24]uint32
	DelayCurrent, DelayMaximum                        int
	Errors                                            uint64
	Observed                                          [13][2]uint64
	StartBeforeRegisterACK                            uint64
}

type lossProxy struct {
	front, back             *net.UDPConn
	server                  netip.AddrPort
	serverSources           []netip.AddrPort // exact backend RX sources; TX destination remains server ingress
	epoch                   time.Time
	ctx                     context.Context
	cancel                  context.CancelFunc
	input                   chan lossEvent
	wg                      sync.WaitGroup
	closeOnce               sync.Once
	done                    chan struct{}
	mu                      sync.Mutex // test snapshots; never added to production
	stats                   lossProxyStats
	firstError              error
	rules                   [24]lossRule
	ruleCount               int
	percent                 bool
	observations            [lossProxyObservationCapacity]lossObservation
	registrations           [lossProxyObservationCapacity]Packet
	registrationSeen        [lossProxyObservationCapacity]bool
	registerACKs            [lossProxyObservationCapacity]Packet
	registerACKSeen         [lossProxyObservationCapacity]bool
	registerCopies          [lossProxyObservationCapacity]uint32
	registerACKCopies       [lossProxyObservationCapacity]uint32
	registerACKDelivered    [lossProxyObservationCapacity]bool
	positionIterations      [lossProxyObservationCapacity]uint32
	metrics                 [lossProxyObservationCapacity]Packet
	metricsSeen             [lossProxyObservationCapacity]bool
	metricsFirstAt          [lossProxyObservationCapacity]int64
	neighborACKAfterMetrics [lossProxyObservationCapacity]uint32
	groupDoneFirstAt        [lossProxyObservationCapacity]int64
	groupDoneSeen           [lossProxyObservationCapacity]bool
	client                  netip.AddrPort // scheduler-owned, learned from first client datagram
	delays                  [lossProxyDelayCapacity]lossDelayed
}

func newLossProxy(t testing.TB, ctx context.Context, server netip.AddrPort, rules []lossRule, percent bool, sources ...netip.AddrPort) *lossProxy {
	t.Helper()
	if len(rules) > 24 {
		t.Fatal("fault rule limit exceeded")
	}
	listen := func() *net.UDPConn {
		c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		if err := c.SetReadBuffer(256 * 1024); err != nil {
			c.Close()
			t.Fatal(err)
		}
		if err := c.SetWriteBuffer(256 * 1024); err != nil {
			c.Close()
			t.Fatal(err)
		}
		return c
	}
	front := listen()
	t.Cleanup(func() { front.Close() })
	back := listen()
	pctx, cancel := context.WithCancel(ctx)
	p := &lossProxy{front: front, back: back, server: normalizeEndpoint(server), epoch: time.Now(),
		ctx: pctx, cancel: cancel, input: make(chan lossEvent, lossProxyInputCapacity), done: make(chan struct{}),
		ruleCount: len(rules), percent: percent}
	copy(p.rules[:], rules)
	if len(sources) == 0 {
		sources = []netip.AddrPort{server}
	}
	for _, source := range sources {
		p.serverSources = append(p.serverSources, normalizeEndpoint(source))
	}
	p.wg.Add(3)
	go p.read(front, lossToServer)
	go p.read(back, lossToClients)
	go p.run()
	t.Cleanup(func() { p.Close() })
	return p
}

func (p *lossProxy) Addr() netip.AddrPort { return p.front.LocalAddr().(*net.UDPAddr).AddrPort() }

func (p *lossProxy) Close() error {
	p.closeOnce.Do(func() {
		p.cancel()
		p.front.Close()
		p.back.Close()
		p.wg.Wait()
		close(p.done)
	})
	<-p.done
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.firstError
}

func (p *lossProxy) recordError(err error) {
	p.mu.Lock()
	p.stats.Errors++
	if p.firstError == nil {
		p.firstError = err
	}
	p.mu.Unlock()
	p.cancel()
	p.front.Close()
	p.back.Close()
}

func (p *lossProxy) read(c *net.UDPConn, direction lossDirection) {
	defer p.wg.Done()
	var buf [MaxPacketSize + 1]byte
	for {
		n, addr, err := c.ReadFromUDPAddrPort(buf[:])
		at := time.Since(p.epoch).Nanoseconds()
		if err != nil {
			if p.ctx.Err() == nil {
				p.recordError(err)
			}
			return
		}
		if n > MaxPacketSize {
			p.recordError(ErrLength)
			return
		}
		e := lossEvent{Length: uint8(n), Direction: direction, Source: normalizeEndpoint(addr), ReceivedAt: at}
		copy(e.Data[:], buf[:n])
		select {
		case p.input <- e:
		case <-p.ctx.Done():
			return
		}
	}
}

func (r *lossRule) matches(e lossEvent, packet Packet) bool {
	return r.Direction == e.Direction && r.Type == packet.Type &&
		(r.Origin == 0 || r.Origin == packet.OriginID) && (r.Target == 0 || r.Target == packet.TargetID) &&
		(r.Group == 0 || r.Group == packet.GroupID) && (r.Iteration == 0 || r.Iteration == packet.Iteration) &&
		(r.Nonce == 0 || r.Nonce == packet.RegistrationNonce) && (r.AnyCounter || r.Counter == packet.RetryCounter) &&
		(r.Limit == 0 || r.matched < r.Limit)
}

// Percentage losses use stable logical identities, not arrival order or global
// randomness. The original operation's V25 neighbor ordinal is 0..23. Distinct
// residues exercise the four data/ACK links. Recovery attempts are delivered,
// making this approximately 1% of observed INITIAL data datagrams, reproducibly.
func lossPercentDrop(e lossEvent, packet Packet) bool {
	if packet.RetryCounter != 0 {
		return false
	}
	if packet.Type == POSITION && e.Direction == lossToServer {
		return (packet.OriginID-1)%100 == 99
	}
	origin, neighbor := packet.OriginID, packet.TargetID
	residue := uint32(7)
	if packet.Type == ACK {
		origin, neighbor = packet.TargetID, packet.OriginID
		residue = 31
		if e.Direction == lossToClients {
			residue = 63
		}
	} else if packet.Type != POSITION || e.Direction != lossToClients {
		return false
	}
	slot := (neighbor - 1) % 25
	if neighbor > origin {
		slot-- // exclude the origin's slot from the 24-neighbor ordinal
	}
	return ((origin-1)*24+slot)%100 == residue
}

func TestLossPatternIdentityBasedOnePercent(t *testing.T) {
	var drops [4]int
	for origin := uint32(1); origin <= 100; origin++ {
		p := Packet{Header: Header{Type: POSITION, OriginID: origin}}
		if lossPercentDrop(lossEvent{Direction: lossToServer}, p) {
			drops[0]++
		}
		first := ((origin-1)/25)*25 + 1
		for neighbor := first; neighbor < first+25; neighbor++ {
			if neighbor == origin {
				continue
			}
			p.TargetID = neighbor
			if lossPercentDrop(lossEvent{Direction: lossToClients}, p) {
				drops[1]++
			}
			ack := Packet{Header: Header{Type: ACK, OriginID: neighbor, TargetID: origin}}
			if lossPercentDrop(lossEvent{Direction: lossToServer}, ack) {
				drops[2]++
			}
			if lossPercentDrop(lossEvent{Direction: lossToClients}, ack) {
				drops[3]++
			}
			ack.RetryCounter = 1
			if lossPercentDrop(lossEvent{Direction: lossToServer}, ack) || lossPercentDrop(lossEvent{Direction: lossToClients}, ack) {
				t.Fatal("percentage schedule unexpectedly drops recovery ACK")
			}
		}
	}
	if drops != [4]int{1, 24, 24, 24} {
		t.Fatal("identity schedule is not exactly 1% of the 7300 theoretical initial datagrams", drops)
	}
}

func (p *lossProxy) capture(e lossEvent, packet Packet) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.Received++
	p.stats.Observed[packet.Type][e.Direction]++
	if packet.RetryCounter == 0 && (packet.Type == POSITION || packet.Type == ACK) {
		p.stats.InitialDataObserved++
	}
	if e.Direction == lossToServer && packet.Type == REGISTER {
		index := packet.RegistrationNonce - 1
		if index >= lossProxyObservationCapacity {
			return ErrIdentity
		}
		if !p.registrationSeen[index] {
			p.registrations[index] = packet
			p.registrationSeen[index] = true
		} else {
			if p.registrations[index] != packet {
				return fmt.Errorf("REGISTER changed across attempts: old=%+v new=%+v", p.registrations[index], packet)
			}
		}
		p.registerCopies[index]++
	}
	if e.Direction == lossToClients && packet.Type == REGISTER_ACK {
		index := packet.RegistrationNonce - 1
		if index >= lossProxyObservationCapacity {
			return ErrIdentity
		}
		if p.registerACKSeen[index] && p.registerACKs[index] != packet {
			return fmt.Errorf("REGISTER_ACK changed across attempts: old=%+v new=%+v", p.registerACKs[index], packet)
		}
		p.registerACKs[index], p.registerACKSeen[index] = packet, true
		p.registerACKCopies[index]++
	}
	if e.Direction == lossToServer && (packet.Type == POSITION || packet.Type == RETRY_POSITION) {
		if packet.OriginID == 0 || packet.OriginID > lossProxyObservationCapacity {
			return ErrIdentity
		}
		o := &p.observations[packet.OriginID-1]
		if packet.Type == POSITION {
			if o.InitialSeen && packet.Iteration == o.Initial.Iteration+1 {
				// Phase 4B executes successive operations automatically. Keep only
				// the latest operation per identity, still in fixed test storage.
				*o = lossObservation{}
			}
			if o.InitialSeen {
				return errors.New("original POSITION repeated instead of selective retry")
			}
			if packet.Iteration > 32 {
				return ErrIteration
			}
			p.positionIterations[packet.OriginID-1] |= uint32(1) << (packet.Iteration - 1)
			o.Initial, o.InitialAt, o.InitialSeen = packet, e.ReceivedAt, true
		} else {
			if !o.InitialSeen || packet.MessageID != o.Initial.MessageID || packet.Iteration != o.Initial.Iteration || packet.Position != o.Initial.Position || packet.RunID != o.Initial.RunID {
				return errors.New("retry changed the position operation")
			}
			if !o.FirstRetrySeen {
				o.FirstRetry, o.FirstRetryAt, o.FirstRetrySeen = packet, e.ReceivedAt, true
			}
			o.Retries++
		}
	}
	if e.Direction == lossToServer && packet.Type == METRICS {
		if packet.OriginID == 0 || packet.OriginID > lossProxyObservationCapacity {
			return ErrIdentity
		}
		index := packet.OriginID - 1
		if p.metricsSeen[index] && p.metrics[index] != packet {
			return errors.New("METRICS changed across attempts")
		}
		if !p.metricsSeen[index] {
			p.metrics[index], p.metricsSeen[index], p.metricsFirstAt[index] = packet, true, e.ReceivedAt
		}
	}
	if e.Direction == lossToServer && packet.Type == ACK && packet.OriginID <= lossProxyObservationCapacity && p.metricsSeen[packet.OriginID-1] {
		p.neighborACKAfterMetrics[packet.OriginID-1] |= uint32(1) << ((packet.TargetID - 1) % 32)
	}
	if e.Direction == lossToClients && packet.Type == START && !p.registerACKDelivered[packet.TargetID-1] {
		p.stats.StartBeforeRegisterACK++
	}
	if e.Direction == lossToClients && packet.Type == GROUP_DONE && !p.groupDoneSeen[packet.GroupID-1] {
		p.groupDoneFirstAt[packet.GroupID-1], p.groupDoneSeen[packet.GroupID-1] = e.ReceivedAt, true
	}
	return nil
}

func (p *lossProxy) send(e lossEvent, copies uint8) error {
	c, destination := p.back, p.server
	if e.Direction == lossToClients {
		c, destination = p.front, p.client
	}
	if !destination.IsValid() {
		return ErrEndpoint
	}
	for i := uint8(0); i < copies; i++ {
		n, err := c.WriteToUDPAddrPort(e.Data[:int(e.Length)], destination)
		if err != nil {
			if p.ctx.Err() != nil {
				return nil
			}
			return err
		}
		if n != int(e.Length) {
			return ErrLength
		}
		p.mu.Lock()
		p.stats.Delivered++
		if e.Direction == lossToClients && e.Data[5] == byte(REGISTER_ACK) {
			packet, err := Decode(e.Data[:int(e.Length)])
			if err != nil || packet.TargetID == 0 || packet.TargetID > lossProxyObservationCapacity {
				p.mu.Unlock()
				return ErrIdentity
			}
			p.registerACKDelivered[packet.TargetID-1] = true
		}
		p.mu.Unlock()
	}
	return nil
}

func (p *lossProxy) process(e lossEvent) error {
	if e.Direction == lossToServer {
		if !p.client.IsValid() {
			p.client = e.Source
		} else if p.client != e.Source {
			return ErrEndpoint
		}
	} else {
		accepted := false
		for _, source := range p.serverSources {
			if e.Source == source {
				accepted = true
				break
			}
		}
		if !accepted {
			return ErrEndpoint
		}
	}
	packet, err := Decode(e.Data[:int(e.Length)])
	if err != nil {
		return err
	}
	if err := p.capture(e, packet); err != nil {
		return err
	}
	copies := uint8(1)
	delay := time.Duration(0)
	drop := p.percent && lossPercentDrop(e, packet)
	for i := 0; i < p.ruleCount; i++ {
		r := &p.rules[i]
		if !r.matches(e, packet) {
			continue
		}
		r.matched++
		p.mu.Lock()
		p.stats.RuleMatches[i]++
		p.mu.Unlock()
		drop = r.Drop
		delay = r.Delay
		if r.Duplicate {
			copies = 2
			p.mu.Lock()
			p.stats.Duplicated++
			p.mu.Unlock()
		}
		break
	}
	if drop {
		p.mu.Lock()
		p.stats.Dropped++
		if packet.RetryCounter == 0 && (packet.Type == POSITION || packet.Type == ACK) {
			p.stats.InitialDataDropped++
		}
		p.mu.Unlock()
		return nil
	}
	if delay == 0 {
		return p.send(e, copies)
	}
	for i := range p.delays {
		if !p.delays[i].Occupied {
			p.delays[i] = lossDelayed{Event: e, Due: time.Since(p.epoch).Nanoseconds() + int64(delay), Copies: copies, Occupied: true}
			p.mu.Lock()
			p.stats.Delayed++
			p.stats.DelayCurrent++
			if p.stats.DelayCurrent > p.stats.DelayMaximum {
				p.stats.DelayMaximum = p.stats.DelayCurrent
			}
			p.mu.Unlock()
			return nil
		}
	}
	return errors.New("bounded fault-proxy delay array is full")
}

func (p *lossProxy) flushDue(now int64) (int64, error) {
	var next int64
	for i := range p.delays {
		d := &p.delays[i]
		if !d.Occupied {
			continue
		}
		if d.Due <= now {
			if err := p.send(d.Event, d.Copies); err != nil {
				return 0, err
			}
			*d = lossDelayed{}
			p.mu.Lock()
			p.stats.DelayCurrent--
			p.mu.Unlock()
		} else if next == 0 || d.Due < next {
			next = d.Due
		}
	}
	return next, nil
}

func (p *lossProxy) run() {
	defer p.wg.Done()
	timer := time.NewTimer(time.Hour)
	stopClientTimer(timer)
	defer timer.Stop()
	for {
		now := time.Since(p.epoch).Nanoseconds()
		next, err := p.flushDue(now)
		if err != nil {
			p.recordError(err)
			return
		}
		var tick <-chan time.Time
		if next != 0 {
			timer.Reset(time.Duration(next - now))
			tick = timer.C
		}
		select {
		case <-p.ctx.Done():
			return
		case <-tick:
		case e := <-p.input:
			if err := p.process(e); err != nil {
				p.recordError(err)
				return
			}
		}
		stopClientTimer(timer)
	}
}

func (p *lossProxy) snapshot() lossProxyStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

func waitLossCondition(t *testing.T, ctx context.Context, g *ClientGenerator, p *lossProxy, label string, allowFailure bool, condition func(ClientStats) bool) ClientStats {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		st := g.Stats()
		if st.InvalidPackets+st.QueueDrops+st.SocketErrors+st.SendDrops+st.RejectedOperations+st.InvalidConfirmations != 0 || !allowFailure && st.FailedRegistrations+st.FailedOperations+st.RetryExhausted != 0 {
			t.Fatalf("%s: unexpected generator error: %+v; proxy=%+v", label, st, p.snapshot())
		}
		if pst := p.snapshot(); pst.Errors != 0 {
			t.Fatalf("%s: proxy error: %+v", label, pst)
		}
		if condition(st) {
			return st
		}
		select {
		case <-ctx.Done():
			t.Fatalf("%s: incomplete: %+v; proxy=%+v; %v", label, st, p.snapshot(), ctx.Err())
		case <-g.done:
			t.Fatalf("%s: generator stopped: %v", label, g.Wait())
		case <-ticker.C:
		}
	}
}

func lossClientConfig(t testing.TB, scfg ServerConfig, proxy *lossProxy) ClientConfig {
	t.Helper()
	cfg := clientsConfig(t, scfg.Manifest.run.N, scfg.Manifest.run.V, scfg.Manifest.run.I, 6)
	cfg.Manifest, cfg.Server = scfg.Manifest, proxy.Addr()
	cfg.RegisterWindow, cfg.RegisterRate = 8, 500
	cfg.PositionRate, cfg.PositionBurst, cfg.PositionWindow = 100, 1, 16
	cfg.RetryRate, cfg.RetryBurst = 100, 32
	cfg.InitialRTO, cfg.MaxRTO = 100*time.Millisecond, 400*time.Millisecond
	cfg.JitterPercent, cfg.JitterSeed = 0, 0x415243
	return cfg
}

func checkLossCompletion(t *testing.T, g *ClientGenerator, s *Server, p *lossProxy, expectedOperations, expectedACKs uint64) {
	t.Helper()
	st, sst, pst := g.Stats(), s.Stats(), p.snapshot()
	if st.RegisteredClients != uint64(len(g.clients)) || st.CompletedOperations != expectedOperations || st.UniqueAcks != expectedACKs || st.RTTSamples != expectedOperations || st.RetryExhausted+st.FailedRegistrations+st.FailedOperations+st.AbortedOperations != 0 {
		t.Fatalf("logical completion: %+v", st)
	}
	if st.InvalidPackets+st.QueueDrops+st.SocketErrors+st.SendDrops+st.RejectedOperations+st.InvalidConfirmations != 0 || sst.QueueDrops+sst.InvalidPackets+sst.SocketErrors+sst.SendDrops != 0 || pst.Errors != 0 {
		t.Fatalf("unexpected loss outside injector: generator=%+v server=%+v proxy=%+v", st, sst, pst)
	}
	if st.InflightCurrent != 0 || st.DeadlineHeapCurrent != 0 || st.RegisterInFlight != 0 || st.InflightMaximum > int64(g.cfg.PositionWindow) || st.MaximumRegisterInFlight > int64(g.cfg.RegisterWindow) {
		t.Fatalf("ARQ/window leak or exceeded bound: %+v", st)
	}
	if expectedOperations == uint64(len(g.clients)) {
		checkCompletedClients(t, g, g.cfg.Manifest.run.V)
	}
	for i := range g.clients {
		c := &g.clients[i]
		if c.Phase != ClientComplete {
			continue
		}
		o := &p.observations[i]
		if !o.InitialSeen {
			t.Fatal("completed without initial write through UDP", i)
		}
		// Both clocks are monotonic-relative epochs from this process. If retry
		// confirmation reset SentAt, it would fall AFTER the first proxy reception.
		initialT0 := g.epoch.Sub(p.epoch).Nanoseconds() + c.SentAt
		if initialT0 > o.InitialAt {
			t.Fatalf("retry replaced first t0: id=%d t0=%d firstPOSITION=%d", i+1, initialT0, o.InitialAt)
		}
		minimumRTO := uint64(g.cfg.InitialRTO) * uint64(100-g.cfg.JitterPercent) / 100
		if o.FirstRetrySeen && (o.FirstRetryAt <= o.InitialAt || c.LastRTTNS < minimumRTO*9/10) {
			t.Fatalf("RTT excluded recovery: id=%d RTT=%s observation=%+v", i+1, time.Duration(c.LastRTTNS), *o)
		}
	}
}

func TestClientsLossDirected100V25(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	scfg := serverConfig(t, 100, 25, 6, false)
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	rules := []lossRule{
		{Direction: lossToServer, Type: REGISTER, Nonce: 1, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 2, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: POSITION, Origin: 3, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: POSITION, Origin: 4, Target: 5, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: ACK, Origin: 7, Target: 6, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 9, Target: 8, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 11, Target: 10, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 12, Target: 10, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 14, Target: 13, Limit: 1, Duplicate: true},
		{Direction: lossToClients, Type: ACK, Origin: 16, Target: 15, Limit: 1, Delay: 250 * time.Millisecond},
		{Direction: lossToClients, Type: ACK, Origin: 16, Target: 15, Counter: 1, Limit: 1, Drop: true},
	}
	p := newLossProxy(t, ctx, s.Addr(), rules, false)
	cfg := lossClientConfig(t, scfg, p)
	begin := time.Now()
	g, err := ListenClients(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	waitLossCondition(t, ctx, g, p, "directed registration", false, func(st ClientStats) bool {
		return st.RegisteredClients == 100 && st.RegisterInFlight == 0
	})
	registrationElapsed := time.Since(begin)
	positionBegin := time.Now()
	for id := uint32(1); id <= 100; id++ {
		if err := g.StartPosition(id, 1, PositionPayload{float32(id), -1, 0.5}); err != nil {
			t.Fatal(err)
		}
	}
	st := waitLossCondition(t, ctx, g, p, "directed data recovery", false, func(st ClientStats) bool {
		return st.CompletedOperations == 100 && st.UniqueAcks == 2400 && st.InflightCurrent == 0 && st.DeadlineHeapCurrent == 0 && p.snapshot().DelayCurrent == 0
	})
	positionElapsed := time.Since(positionBegin)
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	checkLossCompletion(t, g, s, p, 100, 2400)
	pst := p.snapshot()
	for i := range rules {
		if pst.RuleMatches[i] != 1 {
			t.Fatalf("directed fault %d was not exercised exactly once: %+v", i, pst)
		}
	}
	if p.registerCopies[0] != 2 || p.registerCopies[1] != 2 || p.registerACKCopies[0] != 1 || p.registerACKCopies[1] != 2 || st.RegisterRetries != 2 || s.Stats().DuplicateRegisters != 1 {
		t.Fatalf("idempotent register recovery: retries=%d REGISTER1/2=%d/%d REGISTER_ACK1/2=%d/%d server=%+v", st.RegisterRetries, p.registerCopies[0], p.registerCopies[1], p.registerACKCopies[0], p.registerACKCopies[1], s.Stats())
	}
	for _, tc := range []struct{ id, missing uint32 }{
		{3, 0x01fffffb}, {4, 1 << 4}, {6, 1 << 6}, {8, 1 << 8}, {10, (1 << 10) | (1 << 11)}, {15, 1 << 15},
	} {
		o := p.observations[tc.id-1]
		if !o.FirstRetrySeen || o.FirstRetry.MissingMask != tc.missing || o.FirstRetry.RetryCounter != 1 || o.Retries != 1 {
			t.Fatalf("retry not selective for id=%d: expected=%08x observation=%+v", tc.id, tc.missing, o)
		}
	}
	if st.PositionRetries != 6 || st.RetryNeighborsRequested != 30 || st.RecoveredOperations != 6 || st.DuplicateAcks != 1 || pst.Dropped != 9 || pst.Duplicated != 1 || pst.Delayed != 1 {
		t.Fatalf("unique/selective recovery counts: generator=%+v proxy=%+v", st, pst)
	}
	if g.clients[14].LastRTTNS < uint64(250*time.Millisecond) {
		t.Fatal("old RetryCounter ACK did not complete recovery with original RTT", g.clients[14])
	}
	t.Logf("N100/V25/I1 directed: registered100 complete100 uniqueACK2400 failed0; registration=%s positions=%s registerRetries=%d positionRetries=%d requestedNeighbors=%d recovered=%d duplicateACK=%d; proxy=%+v", registrationElapsed, positionElapsed, st.RegisterRetries, st.PositionRetries, st.RetryNeighborsRequested, st.RecoveredOperations, st.DuplicateAcks, pst)
}

func TestClientsLossDeterministicOnePercent100V25(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	scfg := serverConfig(t, 100, 25, 6, false)
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	p := newLossProxy(t, ctx, s.Addr(), nil, true)
	cfg := lossClientConfig(t, scfg, p)
	cfg.JitterPercent = 20
	g, err := ListenClients(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	waitLossCondition(t, ctx, g, p, "percent registration", false, func(st ClientStats) bool { return st.RegisteredClients == 100 && st.RegisterInFlight == 0 })
	begin := time.Now()
	for id := uint32(1); id <= 100; id++ {
		if err := g.StartPosition(id, 1, PositionPayload{float32(id), 2, 3}); err != nil {
			t.Fatal(err)
		}
	}
	st := waitLossCondition(t, ctx, g, p, "percent data recovery", false, func(st ClientStats) bool {
		return st.CompletedOperations == 100 && st.UniqueAcks == 2400 && st.InflightCurrent == 0 && st.DeadlineHeapCurrent == 0
	})
	elapsed := time.Since(begin)
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	checkLossCompletion(t, g, s, p, 100, 2400)
	pst := p.snapshot()
	percent := 100 * float64(pst.InitialDataDropped) / float64(pst.InitialDataObserved)
	if percent < 0.8 || percent > 1.2 || pst.InitialDataDropped < 60 || pst.InitialDataDropped > 80 || st.PositionRetries == 0 || st.RecoveredOperations == 0 {
		t.Fatalf("deterministic ~1%% schedule not exercised: %.3f%%; generator=%+v proxy=%+v", percent, st, pst)
	}
	for _, o := range p.observations {
		if o.FirstRetrySeen && (o.FirstRetry.MissingMask == 0 || bits.OnesCount32(o.FirstRetry.MissingMask) > 24) {
			t.Fatal("invalid selective mask", o)
		}
	}
	t.Logf("N100/V25/I1 identity-based initial-data loss: %.3f%% (%d/%d), complete100 uniqueACK2400 failed0, positions=%s retries=%d requestedNeighbors=%d recovered=%d duplicates=%d proxy=%+v", percent, pst.InitialDataDropped, pst.InitialDataObserved, elapsed, st.PositionRetries, st.RetryNeighborsRequested, st.RecoveredOperations, st.DuplicateAcks, pst)
}

func TestClientsLossPermanentACKExhaustsBoundedAttempts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	scfg := serverConfig(t, 3, 3, 1, false)
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	p := newLossProxy(t, ctx, s.Addr(), []lossRule{{Direction: lossToClients, Type: ACK, Target: 1, AnyCounter: true, Drop: true}}, false)
	cfg := lossClientConfig(t, scfg, p)
	cfg.InitialRTO, cfg.MaxRTO, cfg.MaxSends = 20*time.Millisecond, 40*time.Millisecond, 3
	g, err := ListenClients(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	waitLossCondition(t, ctx, g, p, "permanent-loss registration", false, func(st ClientStats) bool { return st.RegisteredClients == 3 && st.RegisterInFlight == 0 })
	if err := g.StartPosition(1, 1, PositionPayload{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	st := waitLossCondition(t, ctx, g, p, "permanent-loss explicit failure", true, func(st ClientStats) bool {
		return st.RetryExhausted == 1 && st.FailedOperations == 1 && st.InflightCurrent == 0 && st.DeadlineHeapCurrent == 0
	})
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if st.CompletedOperations != 0 || st.PositionSent != 1 || st.PositionRetries != 2 || g.clients[0].Phase != ClientFailed || g.clients[0].AckMask != 0 || p.observations[0].Retries != 2 || p.snapshot().Dropped != 6 {
		t.Fatalf("retry bound did not fail exactly once: generator=%+v client=%+v proxy=%+v", st, g.clients[0], p.snapshot())
	}
	t.Log("permanent ACK loss: original + 2 selective retries, one explicit Failed/RetryExhausted, inflight/deadline heap=0, no retry loop")
}
