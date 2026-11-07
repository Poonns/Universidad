package protocol

import (
	"bytes"
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func serverConfig(t testing.TB, n, v uint32, workers int, two bool) ServerConfig {
	t.Helper()
	run := RunConfig{42, n, v, 1}
	gens := []GeneratorConfig{{7, 1, n, run}}
	if two {
		gens = []GeneratorConfig{{7, 1, n / 2, run}, {9, n/2 + 1, n / 2, run}}
	}
	m, err := NewManifest(run, gens)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultServerConfig(m)
	cfg.Academic = false // Preserve the phase 2 routing fixture without automatic academic controls.
	cfg.Workers = workers
	cfg.GOMAXPROCS = 0
	cfg.Address = "127.0.0.1:0"
	cfg.InputCapacity = 128
	cfg.HighCapacity = 128
	cfg.DataCapacity = 128
	cfg.ReadBuffer = 256 * 1024
	cfg.WriteBuffer = 256 * 1024
	return cfg
}
func bareServer(t testing.TB, cfg ServerConfig) *Server {
	t.Helper()
	s, err := newServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func testEndpoint(index int) netip.AddrPort {
	return netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(21000+index))
}
func registerPacket(s *Server, gi int, nonce uint64) Packet {
	return Packet{Header: Header{Type: REGISTER, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: s.cfg.Manifest.generators[gi].GeneratorID, RegistrationNonce: nonce, MessageID: nonce}}
}
func deliver(t testing.TB, s *Server, p Packet, source netip.AddrPort) error {
	t.Helper()
	if err := s.decodeDispatch(wire(t, p), source); err != nil {
		return err
	}
	group := p.GroupID
	if p.Type == REGISTER {
		_, group, _ = s.cfg.Manifest.RegistrationIdentity(p.GeneratorID, p.RegistrationNonce)
	}
	index, err := s.WorkerForGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	e, ok := s.workers[index].input.pop()
	if !ok {
		t.Fatal("missing dispatched event")
	}
	return s.workers[index].handle(e)
}
func collectOutbound(t testing.TB, s *Server) []outbound {
	t.Helper()
	var jobs []outbound
	for attempt := 0; attempt < 10000; attempt++ {
		progress := false
		for i := range s.workers {
			if s.workers[i].flushPending() {
				progress = true
			}
		}
		for i := range s.writers {
			for {
				job, ok := s.writers[i].takeOutbound()
				if !ok {
					break
				}
				jobs = append(jobs, job)
				progress = true
			}
		}
		if !progress {
			return jobs
		}
	}
	t.Fatal("pending drain did not converge")
	return nil
}
func registerAll(t testing.TB, s *Server) {
	t.Helper()
	for gi, g := range s.cfg.Manifest.generators {
		for nonce := uint64(1); nonce <= uint64(g.Count); nonce++ {
			if err := deliver(t, s, registerPacket(s, gi, nonce), testEndpoint(gi)); err != nil {
				t.Fatal(err)
			}
			jobs := collectOutbound(t, s)
			if len(jobs) != 1 || jobs[0].Packet.Type != REGISTER_ACK {
				t.Fatalf("registration outputs %+v", jobs)
			}
			if err := s.cfg.Manifest.Validate(&jobs[0].Packet); err != nil {
				t.Fatal(err)
			}
		}
	}
}
func positionPacket(s *Server, id uint32) Packet {
	group := 1 + (id-1)/s.cfg.Manifest.run.V
	gi := s.groups[group-1].GeneratorIndex
	return Packet{Header: Header{Type: POSITION, RunID: 42, GeneratorID: s.cfg.Manifest.generators[gi].GeneratorID,
		GroupID: group, OriginID: id, Iteration: 1, MessageID: 1000 + uint64(id)}, Position: PositionPayload{1, -2, 0.5}}
}
func ackFor(p Packet, neighbor uint32) Packet {
	return Packet{Header: Header{Type: ACK, RunID: p.RunID, GeneratorID: p.GeneratorID, GroupID: p.GroupID,
		OriginID: neighbor, TargetID: p.OriginID, Iteration: p.Iteration, MessageID: p.MessageID, RetryCounter: p.RetryCounter}}
}

func TestServerRegistrationAndEndpoints(t *testing.T) {
	s := bareServer(t, serverConfig(t, 50, 25, 6, true))
	p := registerPacket(s, 0, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	first := collectOutbound(t, s)
	if len(first) != 1 || first[0].Packet.TargetID != 1 || first[0].Packet.GroupID != 1 {
		t.Fatal(first)
	}
	p.RetryCounter = 9
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	duplicate := collectOutbound(t, s)
	if len(duplicate) != 1 || !bytes.Equal(wire(t, first[0].Packet), wire(t, duplicate[0].Packet)) {
		t.Fatal("REGISTER_ACK not byte-identical")
	}
	if s.Stats().RegisteredClients != 1 || s.Stats().DuplicateRegisters != 1 || s.groups[0].RegisteredMask != 1 {
		t.Fatal(s.Stats())
	}
	p.MessageID++
	if err := deliver(t, s, p, testEndpoint(0)); err != ErrConflict {
		t.Fatal(err)
	}
	p.MessageID--
	if err := deliver(t, s, p, testEndpoint(1)); err != ErrEndpoint {
		t.Fatal(err)
	}
	p = registerPacket(s, 1, 1)
	if err := deliver(t, s, p, testEndpoint(1)); err != nil {
		t.Fatal(err)
	}
	jobs := collectOutbound(t, s)
	if len(jobs) != 1 || jobs[0].Packet.TargetID != 26 || jobs[0].Packet.GroupID != 2 {
		t.Fatal(jobs)
	}
	p.RegistrationNonce = 26
	if err := deliver(t, s, p, testEndpoint(1)); err != ErrNonce {
		t.Fatal(err)
	}
	p = registerPacket(s, 0, 2)
	p.RunID++
	if err := deliver(t, s, p, testEndpoint(0)); err != ErrRun {
		t.Fatal(err)
	}
	p = positionPacket(s, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != ErrNotRegistered {
		t.Fatal(err)
	}
	if s.Stats().RegisteredClients != 2 {
		t.Fatal("invalid registration changed state")
	}
}

func TestServerPositionFanoutAndRealACK(t *testing.T) {
	s := bareServer(t, serverConfig(t, 25, 25, 6, false))
	registerAll(t, s)
	p := positionPacket(s, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs := collectOutbound(t, s)
	if len(jobs) != 24 {
		t.Fatalf("got %d outputs", len(jobs))
	}
	var seen uint32
	for _, job := range jobs {
		q := job.Packet
		if q.Type != POSITION || q.TargetID == 1 || q.OriginID != 1 || q.GroupID != 1 || q.MessageID != p.MessageID || q.Iteration != 1 || q.Position != p.Position || q.Flags != ServerToClient {
			t.Fatal(q)
		}
		if err := s.cfg.Manifest.Validate(&q); err != nil {
			t.Fatal(err)
		}
		bit := uint32(1) << (q.TargetID - 1)
		if seen&bit != 0 {
			t.Fatal("duplicate fanout target")
		}
		seen |= bit
	}
	if seen != 0x01fffffe {
		t.Fatalf("fanout %08x", seen)
	}
	if s.clients[0].SeenACKMask != 0 || s.Stats().AckReceived != 0 {
		t.Fatal("synthetic ACK progress")
	}
	ack := ackFor(p, 2)
	ack.RetryCounter = 4
	if err := deliver(t, s, ack, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs = collectOutbound(t, s)
	ack.Flags = ServerToClient
	if len(jobs) != 1 || jobs[0].Packet != ack {
		t.Fatalf("ACK not forwarded faithfully: %+v", jobs)
	}
	ack.Flags = 0
	if err := deliver(t, s, ack, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if len(collectOutbound(t, s)) != 1 || s.Stats().DuplicateACKs != 1 || s.clients[0].SeenACKMask != 2 {
		t.Fatal("duplicate ACK progress")
	}
}

func TestServerRejectsInvalidACKs(t *testing.T) {
	s := bareServer(t, serverConfig(t, 50, 25, 3, false))
	registerAll(t, s)
	p := positionPacket(s, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	collectOutbound(t, s)
	for _, tc := range []struct {
		name   string
		change func(*Packet)
		source netip.AddrPort
		want   error
	}{
		{"wrong-neighbor", func(p *Packet) { p.OriginID = 26 }, testEndpoint(0), ErrIdentity},
		{"wrong-group", func(p *Packet) { p.GroupID = 2 }, testEndpoint(0), ErrIdentity},
		{"wrong-message", func(p *Packet) { p.MessageID++ }, testEndpoint(0), ErrMessage},
		{"wrong-operation-target", func(p *Packet) { p.TargetID = 3 }, testEndpoint(0), ErrIteration},
		{"wrong-run", func(p *Packet) { p.RunID++ }, testEndpoint(0), ErrRun},
		{"wrong-endpoint", func(p *Packet) {}, testEndpoint(1), ErrEndpoint},
		{"wrong-generator", func(p *Packet) { p.GeneratorID = 9 }, testEndpoint(0), ErrIdentity},
		{"server-direction", func(p *Packet) { p.Flags = ServerToClient }, testEndpoint(0), ErrFlags},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ack := ackFor(p, 2)
			tc.change(&ack)
			if err := deliver(t, s, ack, tc.source); err != tc.want {
				t.Fatalf("%v != %v", err, tc.want)
			}
			if len(collectOutbound(t, s)) != 0 || s.clients[0].SeenACKMask != 0 {
				t.Fatal("invalid ACK produced progress")
			}
		})
	}
	// Self-ACK is rejected by the codec before it can reach routing.
	ack := ackFor(p, 1)
	var buf [MaxPacketSize]byte
	if _, err := Encode(buf[:], &ack); err != ErrIdentity {
		t.Fatal(err)
	}
}

func TestServerSelectiveRetryAndConflict(t *testing.T) {
	s := bareServer(t, serverConfig(t, 25, 25, 2, false))
	registerAll(t, s)
	p := positionPacket(s, 1)
	// A selective retry can be the first observed datagram if initial POSITION was lost.
	p.Type = RETRY_POSITION
	p.RetryCounter = 1
	p.MissingMask = 1 << 24
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs := collectOutbound(t, s)
	if len(jobs) != 1 || jobs[0].Packet.TargetID != 25 || jobs[0].Packet.Type != RETRY_POSITION || jobs[0].Packet.MissingMask != p.MissingMask {
		t.Fatal(jobs)
	}
	p.RetryCounter++
	p.MissingMask = (1 << 1) | (1 << 4) | (1 << 24)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs = collectOutbound(t, s)
	if len(jobs) != 3 {
		t.Fatal(len(jobs))
	}
	var got uint32
	for _, job := range jobs {
		got |= uint32(1) << (job.Packet.TargetID - 1)
		if err := s.cfg.Manifest.Validate(&job.Packet); err != nil {
			t.Fatal(err)
		}
	}
	if got != p.MissingMask {
		t.Fatal(got)
	}
	for _, mask := range []uint32{1, 1 << 25} {
		bad := p
		bad.MissingMask = mask
		if err := deliver(t, s, bad, testEndpoint(0)); err != ErrMask {
			t.Fatal(err)
		}
	}
	bad := p
	bad.Position.X++
	if err := deliver(t, s, bad, testEndpoint(0)); err != ErrConflict {
		t.Fatal(err)
	}
	bad = p
	bad.MessageID++
	if err := deliver(t, s, bad, testEndpoint(0)); err != ErrConflict {
		t.Fatal(err)
	}
	bad = p
	bad.MissingMask = 2
	if err := deliver(t, s, bad, testEndpoint(0)); err != ErrConflict {
		t.Fatal("same retry attempt mutated", err)
	}
	bad = p
	bad.RetryCounter--
	if err := deliver(t, s, bad, testEndpoint(0)); err != ErrRetry {
		t.Fatal(err)
	}
	if len(collectOutbound(t, s)) != 0 {
		t.Fatal("invalid retry forwarded")
	}
}

func TestServerIterationProgress(t *testing.T) {
	cfg := serverConfig(t, 3, 3, 2, false)
	c := cfg.Manifest.run
	c.I = 2
	cfg.Manifest, _ = NewManifest(c, []GeneratorConfig{{7, 1, 3, c}})
	s := bareServer(t, cfg)
	registerAll(t, s)
	p := positionPacket(s, 1)
	future := p
	future.Iteration = 2
	future.MessageID++
	if err := deliver(t, s, future, testEndpoint(0)); err != ErrProgress {
		t.Fatal(err)
	}
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	collectOutbound(t, s)
	if err := deliver(t, s, future, testEndpoint(0)); err != ErrProgress {
		t.Fatal(err)
	}
	for _, neighbor := range []uint32{2, 3} {
		if err := deliver(t, s, ackFor(p, neighbor), testEndpoint(0)); err != nil {
			t.Fatal(err)
		}
	}
	collectOutbound(t, s)
	if err := deliver(t, s, future, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	if len(collectOutbound(t, s)) != 2 {
		t.Fatal("next iteration fanout")
	}
	if err := deliver(t, s, ackFor(p, 2), testEndpoint(0)); err != ErrIteration {
		t.Fatal(err)
	}
}

func TestServerQueueBoundsAndRecovery(t *testing.T) {
	cfg := serverConfig(t, 25, 25, 1, false)
	cfg.InputCapacity = 1
	cfg.HighCapacity = 1
	cfg.DataCapacity = 1
	s := bareServer(t, cfg)
	data := wire(t, registerPacket(s, 0, 1))
	before := runtime.NumGoroutine()
	if err := s.decodeDispatch(data, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10000; i++ {
		if err := s.decodeDispatch(data, testEndpoint(0)); err != ErrQueueFull {
			t.Fatal(err)
		}
	}
	if runtime.NumGoroutine() != before {
		t.Fatal("queue pressure created goroutines")
	}
	q := s.Stats().Queues[0].Input
	if q.Depth != 1 || q.Maximum != 1 || q.FullAttempts != 10000 || q.Drops != 10000 || s.Stats().RegisteredClients != 0 {
		t.Fatal(q)
	}
	if allocs := testing.AllocsPerRun(100, func() { s.decodeDispatch(data, testEndpoint(0)) }); allocs != 0 {
		t.Fatal("full ingress allocates", allocs)
	}
	e, _ := s.workers[0].input.pop()
	if err := s.workers[0].handle(e); err != nil {
		t.Fatal(err)
	}
	collectOutbound(t, s)
	registerAll(t, s) // includes a duplicate of client 1
	w := &s.workers[0]
	dummy := w.controlReply(1, &s.clients[0])
	w.high.push(dummy, false)
	p := registerPacket(s, 0, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		w.flushPending()
	}
	if w.pending.stats().Depth != 1 || w.high.stats().Depth != 1 || w.high.stats().Drops != 0 || w.high.stats().FullAttempts == 0 {
		t.Fatal("register reply was lost or unbounded")
	}
	if len(collectOutbound(t, s)) != 2 {
		t.Fatal("deferred register reply not recovered")
	}
	p = positionPacket(s, 1)
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	w.flushPending()
	if w.data.stats().Depth != 1 || s.clients[0].PendingPositionMask == 0 {
		t.Fatal("fanout not retained")
	}
	jobs := collectOutbound(t, s)
	if len(jobs) != 24 || w.data.stats().Maximum != 1 || w.data.stats().Drops != 0 {
		t.Fatal("fanout lost under pressure", len(jobs))
	}
	w.high.push(dummy, false)
	ack := ackFor(p, 2)
	if err := deliver(t, s, ack, testEndpoint(0)); err != ErrQueueFull {
		t.Fatal(err)
	}
	if w.high.stats().Drops != 1 {
		t.Fatal("ACK drop not counted")
	}
	collectOutbound(t, s)
	// Recovery requires another REAL ACK; no cached/synthetic response is generated.
	if err := deliver(t, s, ack, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs = collectOutbound(t, s)
	if len(jobs) != 1 || jobs[0].Packet.Type != ACK {
		t.Fatal(jobs)
	}
	if w.pending.stats().Depth != 0 {
		t.Fatal("pending reference leaked")
	}
}

func TestServerRetrySupersedesOnlyPendingIntentions(t *testing.T) {
	cfg := serverConfig(t, 4, 4, 1, false)
	cfg.DataCapacity = 1
	s := bareServer(t, cfg)
	registerAll(t, s)
	p := positionPacket(s, 1)
	p.Type = RETRY_POSITION
	p.RetryCounter = 1
	p.MissingMask = 14
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	s.workers[0].flushPending() // target 2 already queued, 3/4 remain pending
	p.RetryCounter = 2
	p.MissingMask = 8 // only target 4 still needed
	if err := deliver(t, s, p, testEndpoint(0)); err != nil {
		t.Fatal(err)
	}
	jobs := collectOutbound(t, s)
	if len(jobs) != 2 || jobs[0].Packet.TargetID != 2 || jobs[1].Packet.TargetID != 4 || jobs[1].Packet.MissingMask != 8 {
		t.Fatal(jobs)
	}
}

func TestServerWriterPriorityAndFairness(t *testing.T) {
	s := bareServer(t, serverConfig(t, 6, 3, 2, false))
	s.cfg.HighBurst = 2
	for i := range s.workers {
		for j := 0; j < 6; j++ {
			s.workers[i].high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}, GeneratorIndex: i}, false)
			s.workers[i].data.push(outbound{Packet: Packet{Header: Header{Type: POSITION}}, GeneratorIndex: i}, false)
		}
	}
	for cycle := 0; cycle < 6; cycle++ {
		for step := 0; step < 3; step++ {
			job, ok := s.takeOutbound()
			if !ok {
				t.Fatal("writer empty")
			}
			want := ACK
			if step == 2 {
				want = POSITION
			}
			if job.Packet.Type != want {
				t.Fatalf("step %d type %d", step, job.Packet.Type)
			}
			if step == 2 && job.GeneratorIndex != cycle%2 {
				t.Fatal("data worker starvation")
			}
		}
	}
}

func TestServerGroupOwnershipAndVEdges(t *testing.T) {
	s := bareServer(t, serverConfig(t, 100, 25, 3, false))
	for id := uint32(1); id <= 100; id++ {
		group := 1 + (id-1)/25
		owner, err := s.WorkerForGroup(group)
		if err != nil || owner != int((group-1)%3) {
			t.Fatal(owner, err)
		}
		if _, err := s.workers[(owner+1)%3].client(id); err != ErrIdentity {
			t.Fatal("wrong worker accessed client")
		}
	}
	for _, group := range []uint32{0, 5, math.MaxUint32} {
		if _, err := s.WorkerForGroup(group); err != ErrIdentity {
			t.Fatal(err)
		}
	}
	for _, v := range []uint32{1, 32} {
		s := bareServer(t, serverConfig(t, v, v, 3, false))
		registerAll(t, s)
		p := positionPacket(s, 1)
		if err := deliver(t, s, p, testEndpoint(0)); err != nil {
			t.Fatal(err)
		}
		jobs := collectOutbound(t, s)
		if len(jobs) != int(v)-1 {
			t.Fatal(v, len(jobs))
		}
		if v == 1 && s.groups[0].FinalACKMask != 1 {
			t.Fatal("V=1 did not complete empty ACK set")
		}
		if v == 32 && jobs[len(jobs)-1].Packet.TargetID != 32 {
			t.Fatal("slot 31 omitted")
		}
	}
}

func TestServerMalformedAndUnsupported(t *testing.T) {
	s := bareServer(t, serverConfig(t, 25, 25, 2, false))
	valid := wire(t, registerPacket(s, 0, 1))
	for n := 0; n < len(valid); n++ {
		if err := s.decodeDispatch(valid[:n], testEndpoint(0)); err == nil {
			t.Fatal(n)
		}
	}
	for _, data := range [][]byte{make([]byte, MaxPacketSize+1), make([]byte, 65535), []byte("invalid")} {
		if err := s.decodeDispatch(data, testEndpoint(0)); err == nil {
			t.Fatal("malformed accepted")
		}
	}
	p := fixturePacket(START_ACK)
	if err := s.decodeDispatch(wire(t, p), testEndpoint(0)); err != ErrType {
		t.Fatal(err)
	}
	if err := s.decodeDispatch(valid, netip.AddrPort{}); err != ErrEndpoint {
		t.Fatal(err)
	}
	if s.Stats().RegisteredClients != 0 {
		t.Fatal("reader modified registration state")
	}
}

func TestServerInvalidConfiguration(t *testing.T) {
	for _, change := range []func(*ServerConfig){
		func(c *ServerConfig) { c.Manifest = nil }, func(c *ServerConfig) { c.Workers = 0 },
		func(c *ServerConfig) { c.GOMAXPROCS = -1 }, func(c *ServerConfig) { c.InputCapacity = 0 },
		func(c *ServerConfig) { c.HighCapacity = 0 }, func(c *ServerConfig) { c.DataCapacity = 0 },
		func(c *ServerConfig) { c.WorkBudget = 0 }, func(c *ServerConfig) { c.HighBurst = 0 },
		func(c *ServerConfig) { c.ReadBuffer = -1 }, func(c *ServerConfig) { c.WriteBuffer = -1 },
		func(c *ServerConfig) { c.WriteTimeout = 0 },
	} {
		cfg := serverConfig(t, 25, 25, 6, false)
		change(&cfg)
		if _, err := newServer(cfg); err != ErrServerConfig {
			t.Fatal(err)
		}
	}
	cfg := serverConfig(t, 3, 3, 1, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ListenServer(ctx, cfg); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := ListenServer(nil, cfg); err != ErrServerConfig {
		t.Fatal(err)
	}
}

// Real UDP helpers exist ONLY in tests; no production generator is implemented.
func testSocket(t testing.TB) *net.UDPConn {
	t.Helper()
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}
func sendTestPacket(c *net.UDPConn, to netip.AddrPort, p Packet) error {
	var buf [MaxPacketSize]byte
	n, err := Encode(buf[:], &p)
	if err != nil {
		return err
	}
	_, err = c.WriteToUDPAddrPort(buf[:n], to)
	return err
}
func receiveTestPacket(c *net.UDPConn, server netip.AddrPort) (Packet, error) {
	if err := c.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return Packet{}, err
	}
	var buf [MaxPacketSize + 1]byte
	n, addr, err := c.ReadFromUDPAddrPort(buf[:])
	if err != nil {
		return Packet{}, err
	}
	if normalizeEndpoint(addr) != normalizeEndpoint(server) {
		return Packet{}, ErrEndpoint
	}
	return Decode(buf[:n])
}
func readTestPacket(t testing.TB, c *net.UDPConn, server netip.AddrPort) Packet {
	t.Helper()
	p, err := receiveTestPacket(c, server)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestServerLoopbackConcurrentGroups(t *testing.T) {
	cfg := serverConfig(t, 12, 3, 4, true)
	s, err := ListenServer(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	sockets := []*net.UDPConn{testSocket(t), testSocket(t)}
	var wg sync.WaitGroup
	// Two goroutines per test generator deliberately send duplicate registrations concurrently.
	for gi, g := range cfg.Manifest.generators {
		for attempt := 0; attempt < 2; attempt++ {
			wg.Add(1)
			go func(gi int, g GeneratorConfig) {
				defer wg.Done()
				for nonce := uint64(1); nonce <= uint64(g.Count); nonce++ {
					if err := sendTestPacket(sockets[gi], s.Addr(), registerPacket(s, gi, nonce)); err != nil {
						t.Error(err)
						return
					}
				}
			}(gi, g)
		}
	}
	wg.Wait()
	// Duplicate replies may coalesce before entering the writer. Wait for all unique
	// assignments, checking byte-stable contents whenever a duplicate is observed.
	for gi, g := range cfg.Manifest.generators {
		seen := make(map[uint32]Packet)
		for len(seen) < int(g.Count) {
			p := readTestPacket(t, sockets[gi], s.Addr())
			if p.Type != REGISTER_ACK || cfg.Manifest.Validate(&p) != nil {
				t.Fatal(p)
			}
			if old, ok := seen[p.TargetID]; ok && old != p {
				t.Fatal("different duplicate assignment")
			}
			seen[p.TargetID] = p
		}
	}
	if s.Stats().RegisteredClients != 12 || !s.Stats().RegistrationBarrierReached {
		t.Fatal(s.Stats())
	}
	// Pull stats concurrently while owner workers and writer operate.
	statsDone := make(chan struct{})
	statsStopped := make(chan struct{})
	go func() {
		defer close(statsStopped)
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-statsDone:
				return
			case <-ticker.C:
				s.Stats()
			}
		}
	}()
	for gi, g := range cfg.Manifest.generators {
		wg.Add(1)
		go func(gi int, g GeneratorConfig) {
			defer wg.Done()
			received := uint32(0)
			// Send one operation per logical client, including groups on different workers.
			for id := g.FirstID; id < g.FirstID+g.Count; id++ {
				if err := sendTestPacket(sockets[gi], s.Addr(), positionPacket(s, id)); err != nil {
					t.Error(err)
					return
				}
			}
			masks := make([]uint32, g.Count)
			for received < g.Count*(cfg.Manifest.run.V-1) {
				p, err := receiveTestPacket(sockets[gi], s.Addr())
				if err != nil {
					t.Error(err)
					return
				}
				if err := cfg.Manifest.Validate(&p); err != nil {
					t.Error(err)
					return
				}
				switch p.Type {
				case REGISTER_ACK:
					continue // leftover identical responses
				case POSITION:
					if err := sendTestPacket(sockets[gi], s.Addr(), ackFor(p, p.TargetID)); err != nil {
						t.Error(err)
						return
					}
				case ACK:
					if p.TargetID < g.FirstID || p.TargetID >= g.FirstID+g.Count {
						t.Error("wrong endpoint destination")
						return
					}
					bit := uint32(1) << ((p.OriginID - 1) % cfg.Manifest.run.V)
					if masks[p.TargetID-g.FirstID]&bit != 0 {
						t.Error("duplicate ACK in loss-free test")
						return
					}
					masks[p.TargetID-g.FirstID] |= bit
					received++
				default:
					t.Error("unexpected type", p.Type)
					return
				}
			}
		}(gi, g)
	}
	wg.Wait()
	close(statsDone)
	<-statsStopped
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	st := s.Stats()
	t.Logf("loopback: registered=%d duplicateRegisters=%d positions=%d ACK_received=%d ACK_forwarded=%d invalid=%d drops=%d socketErrors=%d", st.RegisteredClients, st.DuplicateRegisters, st.PositionForwarded, st.AckReceived, st.AckForwarded, st.InvalidPackets, st.QueueDrops, st.SocketErrors)
	if st.RegisteredClients != 12 || st.DuplicateRegisters != 12 || st.PositionForwarded != 24 || st.AckReceived != 24 || st.AckForwarded != 24 || st.QueueDrops != 0 || st.InvalidPackets != 0 || st.SocketErrors != 0 {
		t.Fatalf("stats %+v", st)
	}
	for i, c := range s.clients {
		if c.SeenACKMask != c.ExpectedMask {
			t.Fatalf("client %d incomplete", i+1)
		}
	}
	for _, g := range s.groups {
		if g.RegisteredMask != 7 || g.PositionStartedMask != 7 || g.FinalACKMask != 7 {
			t.Fatal(g)
		}
	}
}

func TestServerCleanCloseAndFatalRead(t *testing.T) {
	for _, fatal := range []bool{false, true} {
		cfg := serverConfig(t, 3, 3, 2, false)
		ctx, cancel := context.WithCancel(context.Background())
		s, err := ListenServer(ctx, cfg)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		if fatal {
			s.conn.Close()
		} else {
			cancel()
		}
		select {
		case <-s.done:
		case <-time.After(3 * time.Second):
			t.Fatal("server did not stop")
		}
		cancel()
		if err := s.Wait(); fatal != (err != nil) {
			t.Fatalf("fatal=%v error=%v", fatal, err)
		}
		if err := s.Close(); fatal != (err != nil) {
			t.Fatal("Close not idempotent", err)
		}
		c, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(s.Addr()))
		if err != nil {
			t.Fatal("socket not released", err)
		}
		c.Close()
	}
}

func TestServerStateSizes(t *testing.T) {
	t.Logf("client=%d group=%d inbound=%d outbound=%d worker=%d endpoint=%d", unsafe.Sizeof(ServerClientState{}), unsafe.Sizeof(GroupState{}), unsafe.Sizeof(inbound{}), unsafe.Sizeof(outbound{}), unsafe.Sizeof(serverWorker{}), unsafe.Sizeof(generatorEndpoint{}))
}

func TestServerCongestedFlushHasFiniteWorkBudget(t *testing.T) {
	cfg := serverConfig(t, 100, 1, 1, false)
	cfg.WorkBudget = 4
	cfg.HighCapacity = 1
	s := bareServer(t, cfg)
	registerAll(t, s)
	w := &s.workers[0]
	if !w.high.push(w.controlReply(1, &s.clients[0]), false) {
		t.Fatal("fixture full")
	}
	for nonce := uint64(1); nonce <= 100; nonce++ {
		if err := deliver(t, s, registerPacket(s, 0, nonce), testEndpoint(0)); err != nil {
			t.Fatal(err)
		}
	}
	before := w.high.stats().FullAttempts
	if w.flushPending() {
		t.Fatal("unexpected progress")
	}
	if attempts := w.high.stats().FullAttempts - before; attempts != 4 {
		t.Fatalf("attempts=%d, budget=4", attempts)
	}
	if w.pending.stats().Depth != 100 {
		t.Fatal("pending registration was lost")
	}
}

func TestServerOwnerCleansOnlyItsOwnGroups(t *testing.T) {
	s := bareServer(t, serverConfig(t, 6, 3, 2, false))
	registerAll(t, s)
	for _, id := range []uint32{1, 4} {
		c := &s.clients[id-1]
		c.PendingPositionMask = 6
		c.Flags |= registerReplyPending
		s.workers[(id-1)/3].schedule(id, c)
	}
	other := s.clients[3]
	s.workers[0].finish()
	if s.clients[3] != other {
		t.Fatal("worker changed another owner's state")
	}
	if s.clients[0].PendingPositionMask != 0 || s.clients[0].Flags&(registerReplyPending|clientScheduled) != 0 {
		t.Fatal("owner did not clean state")
	}
	if s.Stats().ShutdownIntentions != 3 {
		t.Fatal(s.Stats().ShutdownIntentions)
	}
	s.workers[1].finish()
	if s.Stats().ShutdownIntentions != 6 {
		t.Fatal(s.Stats().ShutdownIntentions)
	}
}
