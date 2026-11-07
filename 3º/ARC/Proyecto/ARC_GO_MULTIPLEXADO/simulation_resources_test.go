package protocol

import (
	"context"
	"net"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

// This test computes the 160K budget; it never allocates or starts a 160K run.
func TestSimulationStructureSizesAndTheoreticalMemory(t *testing.T) {
	cs, ss, gs := uint64(unsafe.Sizeof(ClientState{})), uint64(unsafe.Sizeof(ServerClientState{})), uint64(unsafe.Sizeof(GroupState{}))
	co, cf := uint64(unsafe.Sizeof(serverControlOutbound{})), uint64(unsafe.Sizeof(serverControlConfirmation{}))
	cw, sw := uint64(unsafe.Sizeof(clientWorker{})), uint64(unsafe.Sizeof(serverWorker{}))
	gate, snapshot := uint64(unsafe.Sizeof(serverControlGate{})), uint64(unsafe.Sizeof(serverGroupSnapshot{}))
	t.Logf("4B sizes: ClientState=%d (4A120), ServerClientState=%d (4A56), GroupState=%d (4A16), controlOutbound=%d controlConfirmation=%d clientWorker=%d serverWorker=%d groupGate=%d groupSnapshot=%d", cs, ss, gs, co, cf, cw, sw, gate, snapshot)
	n, groups, workers := uint64(160000), uint64(6400), uint64(6)
	clientQueues := workers * (4096*uint64(unsafe.Sizeof(clientInbound{})) + (1024+128+2048+512)*uint64(unsafe.Sizeof(clientOutbound{})) + 256*uint64(unsafe.Sizeof(clientConfirmation{})) + 256*uint64(unsafe.Sizeof(clientCommand{})))
	clientAux := n * (4 + 4 + uint64(unsafe.Sizeof(atomic.Uint64{})))
	clientMeta := workers * (cw + 8*uint64(unsafe.Sizeof(boundedQueue[clientInbound]{})))
	serverQueues := workers * (4096*uint64(unsafe.Sizeof(inbound{})) + (1024+2048)*uint64(unsafe.Sizeof(outbound{})) + 256*co + 256*cf)
	serverAux := n*4 + groups*(4+4+gate+snapshot)
	serverMeta := workers * (sw + 7*uint64(unsafe.Sizeof(boundedQueue[inbound]{})))
	t.Logf("N160000/V25/G6400/W6 arithmetic only: clientState=%d (+%d vs4A), clientAux=%d clientQueues=%d clientMetadata=%d clientSubtotal=%d; serverClients=%d serverGroups=%d serverAux=%d serverQueues=%d serverMetadata=%d serverSubtotal=%d; excludes runtime/allocator/kernel, not RSS",
		n*cs, n*(cs-120), clientAux, clientQueues, clientMeta, n*cs+clientAux+clientQueues+clientMeta,
		n*ss, groups*gs, serverAux, serverQueues, serverMeta, n*ss+groups*gs+serverAux+serverQueues+serverMeta)
}

// The N=1000 case exercises startup and shutdown, not a throughput/scale run.
// Registration pacing at 1/s holds the global barrier below N throughout.
func TestSimulationResourcesScaleWithWorkersAndCleanClose(t *testing.T) {
	warm := testSocket(t)
	// A live, silent peer keeps REGISTER deadlines armed while we measure
	// resources. Immediate REGISTER_ACK replies make a polling assertion miss
	// the deliberately short-lived heap entries nondeterministically.
	defer warm.Close()
	silentPeer := warm.LocalAddr().(*net.UDPAddr).AddrPort()
	var serverDeltas, clientDeltas [2]int
	for sample, n := range []uint32{100, 1000} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		scfg := serverConfig(t, n, 25, 6, false)
		scfg.Academic = true
		before := runtime.NumGoroutine()
		s, err := ListenServer(ctx, scfg)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		serverDeltas[sample] = runtime.NumGoroutine() - before
		ccfg := clientsConfig(t, n, 25, 1, 6)
		ccfg.Academic, ccfg.Server, ccfg.RegisterRate = true, silentPeer, 1
		before = runtime.NumGoroutine()
		g, err := ListenClients(ctx, ccfg)
		if err != nil {
			s.Close()
			cancel()
			t.Fatal(err)
		}
		clientDeltas[sample] = runtime.NumGoroutine() - before
		tick := time.NewTicker(time.Millisecond)
		for {
			st, ct := s.Stats(), g.Stats()
			if st.ControlTimers == 6 && st.WriterTimers == 1 && ct.ARQTimers == 6 && ct.WriterTimers == 1 && ct.DeadlineHeapCurrent > 0 {
				break
			}
			select {
			case <-ctx.Done():
				tick.Stop()
				g.Close()
				s.Close()
				cancel()
				t.Fatal("shared schedulers did not start", st, ct)
			case <-tick.C:
			}
		}
		tick.Stop()
		if s.Stats().StartSent != 0 {
			g.Close()
			s.Close()
			cancel()
			t.Fatal("START crossed incomplete global barrier")
		}
		cancel()
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		if err := s.Wait(); err != nil {
			t.Fatal(err)
		}
		st, ct := s.Stats(), g.Stats()
		if st.ControlTimers != 0 || st.WriterTimers != 0 || st.ControlHeapCurrent != 0 || ct.ARQTimers != 0 || ct.WriterTimers != 0 || ct.DeadlineHeapCurrent != 0 || ct.RegisterInFlight != 0 || ct.InflightCurrent != 0 {
			t.Fatal("shutdown leaked shared resources", st, ct)
		}
		for _, q := range st.Queues {
			if q.Input.Depth+q.High.Depth+q.Data.Depth+q.PendingClients.Depth+q.Control.Depth+q.Confirmations.Depth+q.PendingGroups.Depth != 0 {
				t.Fatal("server queues not drained", q)
			}
		}
		for _, q := range ct.Queues {
			if q.Input.Depth+q.High.Depth+q.Control.Depth+q.Data.Depth+q.Retry.Depth+q.Confirmations.Depth+q.Commands.Depth+q.PendingClients.Depth != 0 {
				t.Fatal("client queues not drained", q)
			}
		}
		for _, addr := range []*net.UDPAddr{net.UDPAddrFromAddrPort(s.Addr()), net.UDPAddrFromAddrPort(g.Addr())} {
			rebound, err := net.ListenUDP("udp", addr)
			if err != nil {
				t.Fatal("socket not released", err)
			}
			rebound.Close()
		}
	}
	if serverDeltas[1] > serverDeltas[0]+4 || clientDeltas[1] > clientDeltas[0]+4 {
		t.Fatal("goroutines grew with N", serverDeltas, clientDeltas)
	}
	t.Logf("W6/N100,N1000 startup-only: server goroutine deltas=%v, generator deltas=%v; own count W+3=9 per extreme; each W shared ARQ/control timers +1 writer pacing; all stopped/joined", serverDeltas, clientDeltas)
}
