package protocol

import (
	"context"
	"net"
	"net/netip"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestServerThreeTXShardsPartitionAndValidation(t *testing.T) {
	cfg := serverConfig(t, 150, 25, 6, false)
	if cfg.ServerWriterCount != 1 || cfg.ServerEgressSocketCount != 1 {
		t.Fatal("public defaults changed")
	}
	cfg.ServerWriterCount, cfg.ServerEgressSocketCount = 3, 3
	s := bareServer(t, cfg)
	seen := [6]int{}
	for i := range s.writers {
		wr := &s.writers[i]
		if wr.socketIndex != i || !reflect.DeepEqual(wr.workers, []int{i, i + 3}) {
			t.Fatal("unbalanced W6 partition", wr.workers)
		}
		for _, wi := range wr.workers {
			seen[wi]++
			if s.workers[wi].writerIndex != i {
				t.Fatal("owner/writer mismatch")
			}
			s.workers[wi].data.push(outbound{Packet: Packet{Header: Header{TargetID: uint32(wi + 1), Type: POSITION}}}, false)
		}
	}
	for i := range s.writers {
		for _, wi := range s.writers[i].workers {
			job, ok := s.writers[i].takeOutbound()
			if !ok || job.Packet.TargetID != uint32(wi+1) {
				t.Fatal("consumed wrong worker queue")
			}
		}
		if _, ok := s.writers[i].takeOutbound(); ok {
			t.Fatal("consumed another writer's queue")
		}
	}
	if seen != [6]int{1, 1, 1, 1, 1, 1} {
		t.Fatal("overlap or missing owner", seen)
	}
	for _, sockets := range []int{1, 2, 4} {
		cfg.ServerEgressSocketCount = sockets
		if _, err := newServer(cfg); err != ErrServerConfig {
			t.Fatal("three writers must have three exclusive sockets", sockets, err)
		}
	}
}

func TestServerThreeTXShardsAggregateControlAndPriority(t *testing.T) {
	s := writerServer(t, 300, 25, 6, 3)
	for pass := 0; pass < 5; pass++ {
		for i := range s.workers {
			s.workers[i].flushControlPending()
		}
	}
	s.controlBucket = newClientTokenBucket(40000, 1, 0)
	job, _, ok := s.writers[0].takeControlOutbound(0)
	if !ok || job.GroupID == 0 {
		t.Fatal("missing CONTROL")
	}
	// Both waiting writers skip the busy lease and can still service HIGH.
	for _, i := range []int{1, 2} {
		if _, _, ok := s.writers[i].takeControlOutbound(0); ok {
			t.Fatal("CONTROL lease shared")
		}
		s.workers[i].high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
		if p, _, ctl, _, ok := s.writers[i].takeServerOutbound(0); !ok || ctl || p.Packet.Type != ACK {
			t.Fatal("CONTROL contention blocked HIGH")
		}
	}
	if !s.controlBucket.take(0) {
		t.Fatal("lost aggregate token")
	}
	s.releaseControlLease()
	for _, i := range []int{1, 2} {
		select {
		case <-s.writers[i].wake:
		default:
			t.Fatal("CONTROL waiter not awakened", i)
		}
	}
	count, perWriter := 1, [3]int{1, 0, 0}
	for tick := 0; tick < 120; tick++ {
		now := int64(tick) * 25000 // 40000/s TOTAL, not per writer
		for turn := 0; turn < 3; turn++ {
			i := (tick + turn) % 3
			ctl, _, ok := s.writers[i].takeControlOutbound(now)
			if ok && ctl.GroupID != 0 {
				owner, _ := s.WorkerForGroup(ctl.GroupID)
				if s.workers[owner].writerIndex != i || !s.controlBucket.take(now) {
					t.Fatal("duplicated budget or wrong owner")
				}
				s.releaseControlLease()
				count++
				perWriter[i]++
			}
		}
		if count != tick+1 {
			t.Fatal("CONTROL rate changed", tick, count)
		}
	}
	for _, n := range perWriter {
		if n == 0 {
			t.Fatal("writer starved", perWriter)
		}
	}
	// Exercise unchanged HIGH quota and CONTROL/DATA progress on each shard.
	s = writerServer(t, 150, 25, 6, 3)
	for i := range s.workers {
		for j := 0; j < 32; j++ {
			s.workers[i].high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
			s.workers[i].data.push(outbound{Packet: Packet{Header: Header{Type: POSITION}}}, false)
		}
	}
	for i := range s.writers {
		streak, classes := 0, [3]int{}
		for turn := 0; turn < 64; turn++ {
			p, ctl, control, _, ok := s.writers[i].takeServerOutbound(int64(turn) * 1e6)
			if !ok {
				t.Fatal("eligible work stalled")
			}
			if control {
				if ctl.GroupID == 0 || !s.controlBucket.take(int64(turn)*1e6) {
					t.Fatal("CONTROL selection")
				}
				s.releaseControlLease()
				classes[1]++
				streak = 0
			} else if p.Packet.Type == ACK {
				classes[0]++
				streak++
				if streak > s.cfg.HighBurst {
					t.Fatal("HIGH quota changed")
				}
			} else {
				classes[2]++
				streak = 0
			}
		}
		if classes[0] == 0 || classes[1] == 0 || classes[2] == 0 {
			t.Fatal("class starved", classes)
		}
	}
}

func TestServerThreeTXShardsActualReceiptsAndIndependentDeadlines(t *testing.T) {
	s := writerServer(t, 6, 1, 6, 3)
	s.controlBucket = newClientTokenBucket(1_000_000_000, 1, 0)
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
	for _, conn := range s.egress {
		conn.SetWriteDeadline(time.Now().Add(-time.Second))
	}
	var buf [MaxPacketSize]byte
	for i := range s.writers {
		wr := &s.writers[i]
		for range wr.workers {
			job, _, ok := wr.takeControlOutbound(s.now())
			if !ok {
				t.Fatal("missing CONTROL")
			}
			owner, _ := s.WorkerForGroup(job.GroupID)
			group := &s.groups[job.GroupID-1]
			if s.workers[owner].writerIndex != i || group.ControlDeadline != -1 {
				t.Fatal("wrong owner or premature RTO")
			}
			before := s.now()
			if !wr.writeOutbound(buf[:], outbound{}, job, true) {
				t.Fatal("actual write failed")
			}
			if got := readTestPacket(t, peer, s.EgressEndpoints()[i]); got != job.Packet {
				t.Fatal("datagram identity changed")
			}
			for j := range s.workers {
				want := 0
				if j == owner {
					want = 1
				}
				if s.workers[j].confirmations.stats().Depth != want {
					t.Fatal("confirmation not exclusive to owner")
				}
			}
			tx, ok := s.workers[owner].confirmations.pop()
			if !ok || tx.GroupID != job.GroupID || tx.TargetBit != job.TargetBit || tx.Attempt != job.Attempt || tx.Ticket != job.Ticket || tx.Type != job.Packet.Type || tx.Failed || tx.SentAt < before || tx.SentAt > s.now() {
				t.Fatal("wrong write receipt", tx, job)
			}
			if group.ControlDeadline != -1 {
				t.Fatal("writer armed RTO")
			}
			if err := s.workers[(owner+1)%6].handleControlConfirmation(tx); err != ErrIdentity {
				t.Fatal("wrong owner accepted receipt", err)
			}
			if err := s.workers[owner].handleControlConfirmation(tx); err != nil || group.ControlDeadline != tx.SentAt+int64(s.cfg.InitialControlRTO) {
				t.Fatal("owner confirmation/RTO", err)
			}
			if _, ok := s.workers[owner].confirmations.pop(); ok {
				t.Fatal("duplicate confirmation")
			}
		}
		for j := i + 1; j < 3; j++ {
			if _, err := s.egress[j].WriteToUDPAddrPort([]byte{0}, endpoint); err == nil {
				t.Fatal("writer changed another FD deadline")
			} else if e, ok := err.(net.Error); !ok || !e.Timeout() {
				t.Fatal(err)
			}
		}
	}
}

func TestServerThreeTXShardsShutdownAndGoroutines(t *testing.T) {
	for _, n := range []uint32{100, 1000} {
		cfg := simulationServerConfig(t, n, 25, 5, 6)
		cfg.ServerWriterCount, cfg.ServerEgressSocketCount = 3, 3
		before := runtime.NumGoroutine()
		s, err := ListenServer(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		delta := runtime.NumGoroutine() - before
		addresses := append([]netip.AddrPort{s.Addr()}, s.EgressEndpoints()...)
		if delta != 11 { // W6 + reader1 + writers3 + supervisor1, independent of N
			s.Close()
			t.Fatal("goroutines not O(W+writers)", n, delta)
		}
		if n == 1000 {
			s.conn.Close() // fatal ingress failure must also release all egress
			select {
			case <-s.done:
			case <-time.After(3 * time.Second):
				t.Fatal("fatal shutdown stalled")
			}
			if s.Wait() == nil {
				t.Fatal("fatal read not reported")
			}
		} else if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); n == 100 && err != nil {
			t.Fatal("close not idempotent", err)
		}
		seen := make(map[netip.AddrPort]bool)
		for _, endpoint := range addresses {
			if seen[endpoint] {
				t.Fatal("shared endpoint/FD")
			}
			seen[endpoint] = true
			conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(endpoint))
			if err != nil {
				t.Fatal("socket leaked", endpoint, err)
			}
			conn.Close()
		}
		if st := s.Stats(); st.WriterTimers+st.ControlTimers+st.ControlHeapCurrent != 0 {
			t.Fatal("scheduler resources leaked")
		}
		t.Logf("N%d W6/3TX: own goroutines=%d; ingress+3 egress released", n, delta)
	}
}
