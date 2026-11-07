package protocol

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func writerServer(t testing.TB, n, v uint32, workers, writers int) *Server {
	t.Helper()
	cfg := academicServerConfig(t, n, v, workers)
	cfg.ServerWriterCount = writers
	if writers == 3 {
		cfg.ServerEgressSocketCount = 3
	}
	s := bareServer(t, cfg)
	registerAll(t, s)
	for i := range s.workers {
		s.workers[i].stageStarts()
		s.workers[i].flushControlPending()
	}
	return s
}

func TestServerWriterDefaultPartitionAndValidation(t *testing.T) {
	for _, workers := range []int{2, 5, 6} {
		for _, writers := range []int{1, 2} {
			cfg := serverConfig(t, 100, 25, workers, false)
			if cfg.ServerWriterCount != 1 {
				t.Fatal("default changed")
			}
			cfg.ServerWriterCount = writers
			s := bareServer(t, cfg)
			seen := make([]int, workers)
			for i := range s.writers {
				wr := &s.writers[i]
				for _, wi := range wr.workers {
					seen[wi]++
					if wi%writers != i || s.workers[wi].writerIndex != i {
						t.Fatal("wrong partition", wi, i)
					}
					// Unique consumer: a second writer cannot dequeue this worker.
					s.workers[wi].data.push(outbound{Packet: Packet{Header: Header{TargetID: uint32(wi + 1), Type: POSITION}}}, false)
				}
			}
			for _, n := range seen {
				if n != 1 {
					t.Fatal("worker missing or assigned twice", seen)
				}
			}
			for i := range s.writers {
				for _, want := range s.writers[i].workers {
					job, ok := s.writers[i].takeOutbound()
					if !ok || job.Packet.TargetID != uint32(want+1) {
						t.Fatal("round-robin changed", job, want)
					}
				}
				if _, ok := s.writers[i].takeOutbound(); ok {
					t.Fatal("consumed other writer's queue")
				}
			}
			if workers == 6 && writers == 2 && (!reflect.DeepEqual(s.writers[0].workers, []int{0, 2, 4}) || !reflect.DeepEqual(s.writers[1].workers, []int{1, 3, 5})) {
				t.Fatal("W6 partition")
			}
		}
	}
	for _, count := range []int{-1, 0, 3, 4, 7} {
		cfg := serverConfig(t, 100, 25, 6, false)
		cfg.ServerWriterCount = count
		if _, err := newServer(cfg); err != ErrServerConfig {
			t.Fatal("invalid writer count accepted", count, err)
		}
	}
	cfg := serverConfig(t, 25, 25, 1, false)
	cfg.ServerWriterCount = 2
	if _, err := newServer(cfg); err != ErrServerConfig {
		t.Fatal("more writers than workers accepted")
	}
}

func TestServerTwoWritersPriorityAndAllClassesProgress(t *testing.T) {
	s := writerServer(t, 150, 25, 6, 2)
	for i := range s.workers {
		for j := 0; j < 32; j++ {
			s.workers[i].high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
			s.workers[i].data.push(outbound{Packet: Packet{Header: Header{Type: POSITION}}}, false)
		}
	}
	for i := range s.writers {
		wr := &s.writers[i]
		streak := 0
		classes := [3]int{}
		for turn := 0; turn < 64; turn++ {
			job, ctl, isControl, _, ok := wr.takeServerOutbound(int64(turn) * 1e6)
			if !ok {
				t.Fatal("eligible work did not progress")
			}
			if isControl {
				owner, _ := s.WorkerForGroup(ctl.GroupID)
				if s.workers[owner].writerIndex != i || !s.controlBucket.take(int64(turn)*1e6) {
					t.Fatal("CONTROL ownership/token")
				}
				s.releaseControlLease()
				classes[1]++
				streak = 0
			} else if job.Packet.Type == ACK {
				classes[0]++
				streak++
				if streak > s.cfg.HighBurst {
					t.Fatal("HIGH starved CONTROL/DATA", streak)
				}
			} else {
				classes[2]++
				streak = 0
			}
		}
		for _, n := range classes {
			if n == 0 {
				t.Fatal("starved class", classes)
			}
		}
	}
}

func TestServerTwoWritersAggregateControlBudgetAndBusyLease(t *testing.T) {
	s := writerServer(t, 200, 25, 6, 2)
	s.controlBucket = newClientTokenBucket(1000, 1, 0)
	first, _, ok := s.writers[0].takeControlOutbound(0)
	if !ok || first.GroupID == 0 {
		t.Fatal("first control absent")
	}
	if _, _, ok := s.writers[1].takeControlOutbound(0); ok {
		t.Fatal("second writer acquired same token/lease")
	}
	s.workers[1].high.push(outbound{Packet: Packet{Header: Header{Type: ACK}}}, false)
	if job, _, ctl, _, ok := s.writers[1].takeServerOutbound(0); !ok || ctl || job.Packet.Type != ACK {
		t.Fatal("busy CONTROL blocked HIGH")
	}
	if !s.controlBucket.take(0) {
		t.Fatal("lost reserved token")
	}
	s.releaseControlLease()
	total := 1
	perWriter := [2]int{1, 0}
	for tick := 0; tick < 50; tick++ {
		for _, index := range []int{tick % 2, 1 - tick%2} {
			job, _, ok := s.writers[index].takeControlOutbound(int64(tick) * 1e6)
			if ok && job.GroupID != 0 {
				if !s.controlBucket.take(int64(tick) * 1e6) {
					t.Fatal("aggregate budget race")
				}
				s.releaseControlLease()
				total++
				perWriter[index]++
			}
		}
		if total != tick+1 {
			t.Fatal("rate was duplicated or work stalled", tick, total)
		}
	}
	if perWriter[0] == 0 || perWriter[1] == 0 {
		t.Fatal("writer monopolized control", perWriter)
	}
	if testing.AllocsPerRun(100, func() { s.writers[0].takeServerOutbound(0) }) != 0 {
		t.Fatal("selector allocated")
	}
}

func TestServerTwoWritersActualWriteConfirmationsAndDeadlines(t *testing.T) {
	s := writerServer(t, 2, 1, 2, 2)
	s.ctx = context.Background()
	s.conn = testSocket(t)
	s.address = s.conn.LocalAddr().(*net.UDPAddr).AddrPort()
	peer := testSocket(t)
	endpoint := peer.LocalAddr().(*net.UDPAddr).AddrPort()
	s.endpoints[0].address.Store(&endpoint)
	var buf [MaxPacketSize]byte
	for i := range s.writers {
		wr := &s.writers[i]
		job, _, ok := wr.takeControlOutbound(s.now())
		if !ok || job.GroupID != uint32(i+1) {
			t.Fatal("wrong writer/control", i, job)
		}
		group := &s.groups[job.GroupID-1]
		if group.ControlDeadline != -1 || s.controlHeapCurrent.Load() != 0 {
			t.Fatal("RTO armed on dequeue")
		}
		before := s.now()
		if !wr.writeOutbound(buf[:], outbound{}, job, true) {
			t.Fatal("write failed")
		}
		got := readTestPacket(t, peer, s.Addr()) // bare fixture Addr is assigned below
		if got != job.Packet {
			t.Fatal("datagram changed", got, job.Packet)
		}
		owner, _ := s.WorkerForGroup(job.GroupID)
		for other := range s.workers {
			want := 0
			if other == owner {
				want = 1
			}
			if s.workers[other].confirmations.stats().Depth != want {
				t.Fatal("confirmation went to wrong owner")
			}
		}
		tx, ok := s.workers[owner].confirmations.pop()
		if !ok || tx.GroupID != job.GroupID || tx.TargetBit != job.TargetBit || tx.Ticket != job.Ticket || tx.Attempt != job.Attempt || tx.Type != START || tx.Failed || tx.SentAt < before || tx.SentAt > s.now() {
			t.Fatal("incorrect actual-write receipt", tx, job)
		}
		if group.ControlDeadline != -1 {
			t.Fatal("writer mutated owner deadline")
		}
		if err := s.workers[1-owner].handleControlConfirmation(tx); err == nil {
			t.Fatal("wrong owner accepted receipt")
		}
		if err := s.workers[owner].handleControlConfirmation(tx); err != nil {
			t.Fatal(err)
		}
		if group.ControlDeadline != tx.SentAt+int64(s.cfg.InitialControlRTO) {
			t.Fatal("wrong deadline", group.ControlDeadline, tx)
		}
		if _, ok := s.workers[owner].confirmations.pop(); ok {
			t.Fatal("duplicated receipt")
		}
		s.workers[owner].removeControlDeadline(job.GroupID - 1)
	}
}

func TestServerTwoWritersLimiterProbeWakeup(t *testing.T) {
	s := writerServer(t, 50, 25, 2, 2)
	// Writer 0 can briefly hold the guard while finding no eligible job.
	// This release must wake writer 1 even though writer 0 wrote nothing.
	for i := 0; i < 1000; i++ {
		for {
			select {
			case <-s.writers[1].wake:
				continue
			default:
				goto drained
			}
		}
	drained:
		s.controlMu.Lock()
		if _, _, ok := s.writers[1].takeControlOutbound(0); ok {
			t.Fatal("busy guard consumed control")
		}
		s.releaseControlLease()
		select {
		case <-s.writers[1].wake:
		default:
			t.Fatal("lost CONTROL wakeup")
		}
		if s.controlWaiters.Load() != 0 {
			t.Fatal("waiter leaked")
		}
	}
	if !s.controlMu.TryLock() {
		t.Fatal("guard leaked")
	}
	s.releaseControlLease()
	select {
	case <-s.writers[0].wake: // startup notification may still be queued
	default:
	}
	select {
	case <-s.writers[0].wake:
		t.Fatal("idle release self-woke without a waiter")
	default:
	}
}

func TestServerTwoWritersAcademicRoutingAndReliableControl(t *testing.T) {
	for _, faults := range []bool{false, true} {
		t.Run(fmt.Sprintf("loss=%v", faults), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cfg := simulationServerConfig(t, 100, 25, 5, 6)
			cfg.ServerWriterCount = 2
			var rules []lossRule
			if faults {
				rules = []lossRule{
					{Direction: lossToClients, Type: START, Target: 1, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: START_ACK, Origin: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: GROUP_DONE, Target: 1, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: POSITION, Origin: 27, Target: 28, Iteration: 2, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: ACK, Origin: 2, Target: 1, Iteration: 3, Limit: 1, Drop: true},
				}
			}
			s, g, p := simulationListen(t, ctx, cfg, rules, faults, nil)
			st, ct, ps := simulationCompleteAndClose(t, ctx, s, g, p)
			if ct.UniqueAcks != 12000 || ct.CompletedOperations != 500 || st.GroupDoneAckReceived < 100 {
				t.Fatal("routing/control incomplete")
			}
			var writes uint64
			for _, wr := range st.Writers {
				if wr.HighWrites == 0 || wr.ControlWrites == 0 || wr.DataWrites == 0 || wr.Writes != wr.HighWrites+wr.ControlWrites+wr.DataWrites {
					t.Fatal("writer/class did not run", wr)
				}
				writes += wr.Writes
				if !faults && (wr.Writes != 12200 || wr.HighWrites != 6100 || wr.ControlWrites != 100 || wr.DataWrites != 6000) {
					t.Fatal("duplicate/missing datagrams", wr)
				}
			}
			if writes != st.PacketsSent {
				t.Fatal("lost writer accounting", writes, st.PacketsSent)
			}
			if faults {
				if ps.Dropped != 6 || st.StartRetries == 0 || st.GroupDoneRetries == 0 || ct.PositionRetries == 0 {
					t.Fatal("losses not recovered", ps, st, ct)
				}
				for _, n := range ps.RuleMatches[:len(rules)] {
					if n != 1 {
						t.Fatal("fault not exercised")
					}
				}
			} else if st.PositionForwarded != 12000 || st.AckForwarded != 12000 || ct.PositionRetries+st.StartRetries+st.GroupDoneRetries != 0 {
				t.Fatal("loss-free route changed")
			}
		})
	}
}

func TestServerTwoWritersReleaseLeaseOnCancelAndWriteFailure(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		s := writerServer(t, 2, 1, 2, 2)
		s.ctx = context.Background()
		s.conn = testSocket(t)
		job, _, ok := s.writers[0].takeControlOutbound(s.now())
		if !ok {
			t.Fatal("missing control")
		}
		if cancelled {
			s.controlGates[job.GroupID-1].mask.Store(0)
		} else {
			s.conn.Close()
		}
		var buf [MaxPacketSize]byte
		if !s.writers[0].writeOutbound(buf[:], outbound{}, job, true) {
			t.Fatal("write loop failed")
		}
		if !s.controlMu.TryLock() {
			t.Fatal("CONTROL lease leaked")
		}
		s.controlMu.Unlock()
		if cancelled {
			if s.workers[0].confirmations.stats().Depth != 0 || s.Stats().PacketsSent != 0 {
				t.Fatal("cancelled control wrote/confirmed")
			}
		} else {
			tx, ok := s.workers[0].confirmations.pop()
			if !ok || !tx.Failed || tx.Ticket != job.Ticket || tx.GroupID != job.GroupID {
				t.Fatal("missing failure receipt", tx)
			}
			if err := s.workers[0].handleControlConfirmation(tx); err != nil {
				t.Fatal(err)
			}
			if s.Stats().ControlFailures != 1 || s.groups[0].ControlDeadline != -1 {
				t.Fatal("failed write armed RTO")
			}
		}
	}
}

func TestServerWriterGoroutinesIndependentOfNAndCleanClose(t *testing.T) {
	for _, writers := range []int{1, 2} {
		for _, n := range []uint32{100, 1000} {
			cfg := simulationServerConfig(t, n, 25, 5, 6)
			cfg.ServerWriterCount = writers
			before := runtime.NumGoroutine()
			s, err := ListenServer(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			delta := runtime.NumGoroutine() - before
			if delta != 6+writers+2 {
				s.Close()
				t.Fatal("goroutines grew with N or wrong writer count", n, writers, delta)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal("non-idempotent close", err)
			}
			st := s.Stats()
			if st.WriterTimers != 0 || st.ControlTimers != 0 || st.ControlHeapCurrent != 0 {
				t.Fatal("timers leaked", st)
			}
			rebound, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(s.Addr()))
			if err != nil {
				t.Fatal("socket leaked", err)
			}
			rebound.Close()
			t.Logf("N%d W6 writers%d: %d own goroutines, joined", n, writers, delta)
		}
	}
}
