package protocol

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestServerEgressConfigurationAndDefault(t *testing.T) {
	for _, sockets := range []int{-1, 0, 3, 4} {
		cfg := serverConfig(t, 100, 25, 6, false)
		cfg.ServerWriterCount = 2
		cfg.ServerEgressSocketCount = sockets
		if _, err := newServer(cfg); err != ErrServerConfig {
			t.Fatal("invalid socket count accepted", sockets, err)
		}
	}
	cfg := serverConfig(t, 100, 25, 6, false)
	if cfg.ServerWriterCount != 1 || cfg.ServerEgressSocketCount != 1 {
		t.Fatal("defaults changed")
	}
	cfg.ServerEgressSocketCount = 2
	if _, err := newServer(cfg); err != ErrServerConfig {
		t.Fatal("two sockets require two writers")
	}
}

func assertEgressAccounting(t testing.TB, st ServerStats, sockets int) {
	t.Helper()
	var total uint64
	writers := 2
	if sockets == 3 {
		writers = 3
	}
	if len(st.Writers) != writers || len(st.Egress) != writers {
		t.Fatal("writer/egress rows", st.Writers, st.Egress)
	}
	for i, e := range st.Egress {
		w := st.Writers[i]
		if e.SocketIndex != i%sockets || e.WriterIndex != i || w.SocketIndex != e.SocketIndex ||
			e.Writes != w.Writes || e.HighWrites != w.HighWrites || e.ControlWrites != w.ControlWrites || e.DataWrites != w.DataWrites || e.SocketErrors != w.SocketErrors {
			t.Fatal("egress accounting mismatch", e, w)
		}
		total += e.Writes
	}
	if total != st.PacketsSent {
		t.Fatal("egress total mismatch", total, st.PacketsSent)
	}
}

func TestServerTransportAcademicFlowAndIngressOnly(t *testing.T) {
	for _, sockets := range []int{1, 2} {
		t.Run(fmt.Sprintf("2W/%dSocket", sockets), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cfg := simulationServerConfig(t, 100, 25, 5, 6)
			cfg.ServerWriterCount = 2
			cfg.ServerEgressSocketCount = sockets
			s, g, _ := simulationListen(t, ctx, cfg, nil, false, nil)
			endpoints := s.EgressEndpoints()
			if len(endpoints) != sockets || len(s.egress) != sockets {
				t.Fatal("socket count")
			}
			for i := range s.writers {
				w := &s.writers[i]
				if w.socketIndex != i%sockets || w.connection() != s.egress[i%sockets] || !reflect.DeepEqual(w.workers, []int{i, i + 2, i + 4}) {
					t.Fatal("writer/socket/worker mapping")
				}
			}
			if sockets == 1 {
				if s.egress[0] != s.conn || endpoints[0] != normalizeEndpoint(s.Addr()) {
					t.Fatal("shared-descriptor baseline changed")
				}
			} else {
				if s.egress[0] == s.egress[1] || s.egress[0] == s.conn || s.egress[1] == s.conn || endpoints[0] == endpoints[1] || endpoints[0] == s.Addr() || endpoints[1] == s.Addr() {
					t.Fatal("egress descriptor/endpoints not distinct")
				}
			}
			// The caller cannot change published endpoint identity/mapping.
			endpoints[0] = testEndpoint(20)
			if s.EgressEndpoints()[0] == endpoints[0] {
				t.Fatal("mutable endpoint slice escaped")
			}
			if err := s.WaitSimulationComplete(ctx); err != nil {
				t.Fatal(err, s.Stats(), g.Stats())
			}
			if err := g.WaitAllDone(ctx); err != nil {
				t.Fatal(err)
			}
			simulationWaitCondition(t, ctx, s, g, nil, "drain before inspecting TX-only sockets", func(st ServerStats, ct ClientStats, _ lossProxyStats) bool { return simulationQueuesEmpty(st, ct) })
			if g.cfg.Server != normalizeEndpoint(s.Addr()) {
				t.Fatal("client destination changed to egress")
			}
			if s.Stats().PacketsReceived != g.Stats().PacketsSent {
				t.Fatal("client sends missed ingress")
			}
			if sockets == 2 {
				var buf [MaxPacketSize + 1]byte
				for _, conn := range s.egress {
					conn.SetReadDeadline(time.Now().Add(2 * time.Millisecond))
					_, _, err := conn.ReadFromUDPAddrPort(buf[:])
					if e, ok := err.(net.Error); !ok || !e.Timeout() {
						t.Fatal("client traffic reached egress instead of canonical ingress", err)
					}
				}
			}
			st, ct, _ := simulationCompleteAndClose(t, ctx, s, g, nil)
			assertEgressAccounting(t, st, sockets)
			if st.PacketsSent != 24400 || st.PacketsReceived != 12900 || ct.PacketsSent != 12900 || ct.PacketsReceived != 24400 || st.StartRetries+st.GroupDoneRetries+ct.PositionRetries != 0 {
				t.Fatal("baseline routing/controls changed", st, ct)
			}
			for _, w := range st.Writers {
				if w.Writes != 12200 || w.HighWrites != 6100 || w.ControlWrites != 100 || w.DataWrites != 6000 {
					t.Fatal("not exact 50/50 workload", w)
				}
			}
		})
	}
}

func TestServerTwoEgressActualReceiptsAndIndependentDeadlines(t *testing.T) {
	cfg := academicServerConfig(t, 2, 1, 2)
	cfg.ServerWriterCount = 2
	cfg.ServerEgressSocketCount = 2
	s := bareServer(t, cfg)
	registerAll(t, s)
	for i := range s.workers {
		s.workers[i].stageStarts()
		s.workers[i].flushControlPending()
	}
	s.ctx = context.Background()
	s.conn = testSocket(t)
	s.address = s.conn.LocalAddr().(*net.UDPAddr).AddrPort()
	if err := s.openEgress(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.closeSockets)
	peer := testSocket(t)
	endpoint := peer.LocalAddr().(*net.UDPAddr).AddrPort()
	s.endpoints[0].address.Store(&endpoint)
	var buf [MaxPacketSize]byte
	// A deadline on socket0 must survive writer1 setting socket1's deadline.
	s.egress[0].SetWriteDeadline(time.Now().Add(-time.Second))
	for _, i := range []int{1, 0} {
		w := &s.writers[i]
		job, _, ok := w.takeControlOutbound(s.now())
		if !ok {
			t.Fatal("missing CONTROL")
		}
		group := &s.groups[job.GroupID-1]
		if group.ControlDeadline != -1 {
			t.Fatal("RTO before actual write")
		}
		before := s.now()
		if !w.writeOutbound(buf[:], outbound{}, job, true) {
			t.Fatal("write failed")
		}
		got := readTestPacket(t, peer, s.EgressEndpoints()[i])
		if got != job.Packet {
			t.Fatal("payload/identity changed")
		}
		owner, _ := s.WorkerForGroup(job.GroupID)
		if owner != i {
			t.Fatal("owner changed")
		}
		tx, ok := s.workers[owner].confirmations.pop()
		if !ok || tx.Ticket != job.Ticket || tx.GroupID != job.GroupID || tx.TargetBit != job.TargetBit || tx.Attempt != job.Attempt || tx.Type != job.Packet.Type || tx.Failed || tx.SentAt < before || tx.SentAt > s.now() {
			t.Fatal("receipt mismatch", tx, job)
		}
		if err := s.workers[1-owner].handleControlConfirmation(tx); err != ErrIdentity {
			t.Fatal("wrong owner accepted receipt")
		}
		if group.ControlDeadline != -1 {
			t.Fatal("writer armed RTO")
		}
		if err := s.workers[owner].handleControlConfirmation(tx); err != nil {
			t.Fatal(err)
		}
		if group.ControlDeadline != tx.SentAt+int64(cfg.InitialControlRTO) {
			t.Fatal("bad RTO timestamp")
		}
		if _, ok := s.workers[owner].confirmations.pop(); ok {
			t.Fatal("duplicate receipt")
		}
		if i == 1 {
			_, err := s.egress[0].WriteToUDPAddrPort([]byte{0}, endpoint)
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				t.Fatal("writer1 changed socket0 deadline", err)
			}
		}
	}
	for i := range s.writers {
		s.writers[i].publishCounts()
	}
	assertEgressAccounting(t, s.Stats(), 2)
}

func TestServerTwoEgressReliableFlowUnderLoss(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cfg := simulationServerConfig(t, 100, 25, 5, 6)
	cfg.ServerWriterCount = 2
	cfg.ServerEgressSocketCount = 2
	rules := []lossRule{
		{Direction: lossToServer, Type: REGISTER, Nonce: 1, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 26, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: START, Target: 1, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: START_ACK, Origin: 26, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: POSITION, Origin: 1, Target: 2, Iteration: 2, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 27, Target: 26, Iteration: 3, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: METRICS, Origin: 1, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: GROUP_DONE, Target: 1, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 26, Limit: 1, Drop: true},
	}
	s, g, p := simulationListen(t, ctx, cfg, rules, true, nil)
	st, ct, ps := simulationCompleteAndClose(t, ctx, s, g, p)
	if ps.Dropped != uint64(len(rules)) || ct.RegisterRetries == 0 || ct.PositionRetries == 0 || ct.MetricsRetries == 0 || st.StartRetries == 0 || st.GroupDoneRetries == 0 {
		t.Fatal("loss recovery not exercised", ps, st, ct)
	}
	for _, n := range ps.RuleMatches[:len(rules)] {
		if n != 1 {
			t.Fatal("loss not exercised", ps)
		}
	}
	assertEgressAccounting(t, st, 2)
	if g.cfg.Server != normalizeEndpoint(p.Addr()) || p.server != normalizeEndpoint(s.Addr()) {
		t.Fatal("proxy redirected client sends to source")
	}
}

func TestServerEgressShutdownRebindAndGoroutines(t *testing.T) {
	for _, sockets := range []int{1, 2} {
		for _, n := range []uint32{100, 1000} {
			cfg := simulationServerConfig(t, n, 25, 5, 6)
			cfg.ServerWriterCount = 2
			cfg.ServerEgressSocketCount = sockets
			before := runtime.NumGoroutine()
			s, err := ListenServer(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			delta := runtime.NumGoroutine() - before
			addresses := append([]netip.AddrPort{s.Addr()}, s.EgressEndpoints()...)
			if delta != 10 {
				s.Close()
				t.Fatal("goroutines changed with socket count/N", sockets, n, delta)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			for i, endpoint := range addresses {
				if sockets == 1 && i == 1 {
					continue
				} // alias of ingress, closed once
				conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(endpoint))
				if err != nil {
					t.Fatal("socket leaked", endpoint, err)
				}
				conn.Close()
			}
			st := s.Stats()
			if st.WriterTimers != 0 || st.ControlTimers != 0 || st.ControlHeapCurrent != 0 {
				t.Fatal("timer leaked")
			}
			t.Logf("N%d W6 writers2 egress%d: own goroutines=%d; released %d unique sockets", n, sockets, delta, 1+2*(sockets-1))
		}
	}
	cfg := simulationServerConfig(t, 100, 25, 5, 6)
	cfg.ServerWriterCount = 2
	cfg.ServerEgressSocketCount = 2
	s, err := ListenServer(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	addresses := append([]netip.AddrPort{s.Addr()}, s.EgressEndpoints()...)
	s.conn.Close() // fatal ingress read must also close both egress sockets
	select {
	case <-s.done:
	case <-time.After(3 * time.Second):
		t.Fatal("fatal shutdown hung")
	}
	if s.Wait() == nil {
		t.Fatal("fatal read not reported")
	}
	for _, endpoint := range addresses {
		conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(endpoint))
		if err != nil {
			t.Fatal("fatal shutdown leaked socket", err)
		}
		conn.Close()
	}
}

func TestServerTwoEgressControlBudgetUnchanged(t *testing.T) {
	cfg := academicServerConfig(t, 200, 25, 6)
	cfg.ServerWriterCount = 2
	cfg.ServerEgressSocketCount = 2
	s := bareServer(t, cfg)
	registerAll(t, s)
	for i := range s.workers {
		s.workers[i].stageStarts()
		s.workers[i].flushControlPending()
	}
	s.controlBucket = newClientTokenBucket(40000, 1, 0)
	// Fill all 200 control jobs, rather than only the first WorkBudget pass.
	for pass := 0; pass < 4; pass++ {
		for i := range s.workers {
			s.workers[i].flushControlPending()
		}
	}
	count := 0
	for tick := 0; tick < 100; tick++ {
		now := int64(tick) * 25000 // one aggregate token every 25 us
		for _, index := range []int{tick % 2, 1 - tick%2} {
			job, _, ok := s.writers[index].takeControlOutbound(now)
			if ok && job.GroupID != 0 {
				if !s.controlBucket.take(now) {
					t.Fatal("duplicated aggregate budget")
				}
				s.releaseControlLease()
				count++
			}
		}
		if count != tick+1 {
			t.Fatal("CONTROL rate doubled/stalled", count, tick)
		}
	}
}
