package protocol

import (
	"context"
	"errors"
	"math/bits"
	"net/netip"
	"testing"
	"time"
)

// Phase 4B exercises the public academic mode. There is intentionally no manual
// StartPosition call in this file: START and each previous operation's real ACKs
// are the only triggers for the I subsequent POSITION operations.
func simulationServerConfig(t testing.TB, n, v, iterations uint32, workers int) ServerConfig {
	t.Helper()
	run := RunConfig{RunID: 42, N: n, V: v, I: iterations}
	m, err := NewManifest(run, []GeneratorConfig{{GeneratorID: 7, FirstID: 1, Count: n, Run: run}})
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultServerConfig(m)
	cfg.Academic, cfg.Workers, cfg.GOMAXPROCS, cfg.Address = true, workers, 0, "127.0.0.1:0"
	cfg.InputCapacity, cfg.HighCapacity, cfg.DataCapacity = 512, 512, 512
	cfg.ControlCapacity, cfg.ConfirmationCapacity = 64, 64
	cfg.ControlRate, cfg.ControlBurst = 500, 1
	cfg.InitialControlRTO, cfg.MaxControlRTO = 200*time.Millisecond, 800*time.Millisecond
	cfg.ControlJitterPercent, cfg.ControlJitterSeed = 0, 0x415243
	cfg.ReadBuffer, cfg.WriteBuffer = 256*1024, 256*1024
	return cfg
}

func simulationClientConfig(m *Manifest, server netip.AddrPort, workers int) ClientConfig {
	cfg := DefaultClientConfig(m, 7, server)
	cfg.Academic, cfg.Workers, cfg.GOMAXPROCS = true, workers, 0
	cfg.InputCapacity, cfg.HighCapacity, cfg.ControlCapacity, cfg.DataCapacity = 512, 512, 128, 128
	cfg.ConfirmationCapacity, cfg.CommandCapacity, cfg.RetryCapacity = 128, 128, 128
	cfg.RegisterWindow, cfg.RegisterRate = 8, 500
	cfg.PositionWindow, cfg.PositionRate, cfg.PositionBurst = 16, 500, 1
	cfg.RetryRate, cfg.RetryBurst = 200, 8
	cfg.InitialRTO, cfg.MaxRTO = 200*time.Millisecond, 800*time.Millisecond
	cfg.JitterPercent, cfg.JitterSeed = 0, 0x415243
	cfg.ReadBuffer, cfg.WriteBuffer = 256*1024, 256*1024
	return cfg
}

func simulationListen(t *testing.T, ctx context.Context, scfg ServerConfig, rules []lossRule, proxy bool, mutate func(*ClientConfig)) (*Server, *ClientGenerator, *lossProxy) {
	t.Helper()
	s, err := ListenServer(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	address := s.Addr()
	var p *lossProxy
	if proxy {
		p = newLossProxy(t, ctx, address, rules, false, s.EgressEndpoints()...)
		address = p.Addr()
	}
	cfg := simulationClientConfig(scfg.Manifest, address, scfg.Workers)
	if !proxy {
		cfg.ServerEgressEndpoints = s.EgressEndpoints()
	}
	if mutate != nil {
		mutate(&cfg)
	}
	g, err := ListenClients(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	return s, g, p
}

// This polling helper observes specific counters/queue states; its tick does not
// decide completion. The production wait APIs decide the protocol's outcome.
func simulationWaitCondition(t *testing.T, ctx context.Context, s *Server, g *ClientGenerator, p *lossProxy, label string, condition func(ServerStats, ClientStats, lossProxyStats) bool) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		sst, gst := s.Stats(), g.Stats()
		var pst lossProxyStats
		if p != nil {
			pst = p.snapshot()
			if pst.Errors != 0 {
				t.Fatalf("%s: proxy failed: %+v", label, pst)
			}
		}
		if condition(sst, gst, pst) {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("%s: %v; server=%+v generator=%+v proxy=%+v", label, ctx.Err(), sst, gst, pst)
		case <-ticker.C:
		}
	}
}

func simulationQueuesEmpty(sst ServerStats, gst ClientStats) bool {
	for _, q := range sst.Queues {
		if q.Input.Depth+q.High.Depth+q.Data.Depth+q.PendingClients.Depth+q.Control.Depth+q.Confirmations.Depth+q.PendingGroups.Depth != 0 {
			return false
		}
	}
	for _, q := range gst.Queues {
		if q.Input.Depth+q.High.Depth+q.Control.Depth+q.Data.Depth+q.Retry.Depth+q.Confirmations.Depth+q.Commands.Depth+q.PendingClients.Depth != 0 {
			return false
		}
	}
	return true
}

func simulationCompleteAndClose(t *testing.T, ctx context.Context, s *Server, g *ClientGenerator, p *lossProxy) (ServerStats, ClientStats, lossProxyStats) {
	t.Helper()
	if err := s.WaitSimulationComplete(ctx); err != nil {
		t.Fatalf("server academic completion: %v; server=%+v generator=%+v", err, s.Stats(), g.Stats())
	}
	if err := g.WaitAllDone(ctx); err != nil {
		t.Fatal("generator academic completion", err, g.Stats())
	}
	// Keep both sockets alive until the server has credited the final ACK and
	// every local client is DONE; delayed duplicates are answered during drain.
	simulationWaitCondition(t, ctx, s, g, p, "bounded queues drained after real completion", func(sst ServerStats, gst ClientStats, pst lossProxyStats) bool {
		return sst.SimulationComplete && gst.DoneClients == uint64(len(g.clients)) &&
			sst.ControlHeapCurrent == 0 && gst.DeadlineHeapCurrent == 0 && gst.RegisterInFlight == 0 && gst.InflightCurrent == 0 &&
			pst.DelayCurrent == 0 && simulationQueuesEmpty(sst, gst)
	})
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	var pst lossProxyStats
	if p != nil {
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		pst = p.snapshot()
	}
	sst, gst := s.Stats(), g.Stats()
	run := s.cfg.Manifest.run
	n, operations, uniqueACKs := uint64(run.N), uint64(run.N)*uint64(run.I), uint64(run.N)*uint64(run.I)*uint64(run.V-1)
	if sst.RegisteredClients != n || sst.StartedClients != n || sst.MetricsAccepted != n || sst.DoneClients != n || sst.CompletedGroups != n/uint64(run.V) || !sst.SimulationComplete {
		t.Fatal("server did not complete every academic barrier", sst)
	}
	if gst.RegisteredClients != n || gst.StartedClients != n || gst.CompletedOperations != operations || gst.UniqueAcks != uniqueACKs || gst.RTTSamples != operations || gst.MetricsAccepted != n || gst.DoneClients != n {
		t.Fatal("generator logical counters do not match the academic workload", gst)
	}
	if gst.InvalidPackets+gst.QueueDrops+gst.SocketErrors+gst.SendDrops+gst.RejectedOperations+gst.InvalidConfirmations+gst.FailedRegistrations+gst.FailedOperations+gst.FailedMetrics+gst.RetryExhausted+gst.AbortedOperations != 0 ||
		sst.InvalidPackets+sst.QueueDrops+sst.SocketErrors+sst.SendDrops+sst.ControlFailures+sst.ControlRetryExhausted != 0 || pst.Errors != 0 {
		t.Fatalf("unexpected errors outside injector: server=%+v generator=%+v proxy=%+v", sst, gst, pst)
	}
	if sst.ControlHeapCurrent != 0 || gst.DeadlineHeapCurrent != 0 || sst.ControlTimers != 0 || sst.WriterTimers != 0 || gst.ARQTimers != 0 || gst.WriterTimers != 0 || !simulationQueuesEmpty(sst, gst) {
		t.Fatal("academic completion leaked shared scheduler resources", sst, gst)
	}
	if sst.ControlHeapMaximum > int64(run.N/run.V) || gst.DeadlineHeapMaximum > int64(run.N) || gst.InflightMaximum > int64(g.cfg.PositionWindow) || gst.MaximumRegisterInFlight > int64(g.cfg.RegisterWindow) {
		t.Fatal("academic scheduling exceeded its configured bounded heap/window", sst, gst)
	}
	var sumMean uint64
	mask, _ := GroupMask(run.V)
	for index := range g.clients {
		c := &g.clients[index]
		if !c.Registered || c.Phase != ClientDone || c.CompletedIterations != run.I || c.Iteration != run.I || c.AckMask != c.ExpectedMask || bits.OnesCount32(c.AckMask) != int(run.V)-1 || c.HeapIndex != -1 || c.Failure != ClientFailureNone {
			t.Fatalf("client %d did not run exactly I iterations then reach DONE: %+v", index+1, *c)
		}
		mean := c.RTTSum / uint64(run.I)
		if s.clients[index].MeanRTTNS != mean {
			t.Fatal("accepted metrics do not equal the client's integer RTT mean", index+1, mean, s.clients[index])
		}
		sumMean += mean
		if p != nil && (p.positionIterations[index] != uint32((uint64(1)<<run.I)-1) || !p.metricsSeen[index] || p.metrics[index].Metrics.MeanRTTNS != mean) {
			t.Fatal("proxy did not observe exactly the automatic iterations and stable metrics", index+1)
		}
	}
	for index, group := range s.groups {
		if group.RegisteredMask != mask || group.StartAckMask != mask || group.MetricsMask != mask || group.GroupDoneAckMask != mask || group.ControlHeapIndex != -1 {
			t.Fatal("group academic masks not complete", index+1, group)
		}
		var groupMeanSum uint64
		for id := index * int(run.V); id < (index+1)*int(run.V); id++ {
			groupMeanSum += s.clients[id].MeanRTTNS
		}
		if group.MetricsSumLo != groupMeanSum || group.MetricsSumHi != 0 {
			t.Fatal("group metrics were summed twice or changed", index+1, group)
		}
	}
	if sst.MeanRTTNS != sumMean/n {
		t.Fatal("global mean does not equal sum of accepted client means / N", sst.MeanRTTNS, sumMean/n)
	}
	return sst, gst, pst
}

func TestSimulationEndToEnd100V25I5(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	begin := time.Now()
	s, g, _ := simulationListen(t, ctx, simulationServerConfig(t, 100, 25, 5, 6), nil, false, nil)
	sst, gst, _ := simulationCompleteAndClose(t, ctx, s, g, nil)
	if gst.PositionSent != 500 || gst.MetricsSent != 100 || gst.PositionRetries+gst.RegisterRetries+gst.MetricsRetries+sst.StartRetries+sst.GroupDoneRetries != 0 || sst.StartSent != 100 || sst.GroupDoneSent != 100 || sst.AckForwarded != 12000 || gst.AckSent != 12000 {
		t.Fatal("loss-free workload emitted unexpected data/control retries", sst, gst)
	}
	if sst.StartAckReceived != 100 || sst.MetricsReceived != 100 || sst.MetricsAckSent != 100 || sst.GroupDoneAckReceived != 100 || gst.StartReceived != 100 || gst.StartAckSent != 100 || gst.MetricsAckReceived != 100 || gst.GroupDoneReceived != 100 || gst.GroupDoneAckSent != 100 {
		t.Fatal("loss-free control flow did not exchange exactly one of each control per client", sst, gst)
	}
	t.Logf("N100/V25/I5 automatic: registered100 started100 complete500 uniqueACK12000 metrics100 DONE100 serverDONEACK100 failed0 drops0 socketErrors0 simulationComplete=true; elapsed=%s academic=%s meanRTT=%s; serverControlHeapMax=%d clientHeapMax=%d", time.Since(begin), sst.SimulationDuration, time.Duration(sst.MeanRTTNS), sst.ControlHeapMaximum, gst.DeadlineHeapMaximum)
}

func TestSimulationAutomaticVEdges(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    uint32
	}{{"V1", 1}, {"V32", 32}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s, g, _ := simulationListen(t, ctx, simulationServerConfig(t, tc.v, tc.v, 2, 2), nil, false, nil)
			sst, gst, _ := simulationCompleteAndClose(t, ctx, s, g, nil)
			if gst.PositionSent != uint64(tc.v)*2 || gst.UniqueAcks != uint64(tc.v)*2*uint64(tc.v-1) || sst.CompletedGroups != 1 {
				t.Fatal("automatic V edge omitted an operation, neighbor bit or control member", sst, gst)
			}
			if tc.v == 1 && sst.MeanRTTNS != 0 {
				t.Fatal("V1 fabricated data-plane ACK latency")
			}
		})
	}
}

func TestSimulationMixedLoss100V25I5(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	rules := []lossRule{
		{Direction: lossToServer, Type: REGISTER, Nonce: 1, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 2, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: START, Target: 3, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: START_ACK, Origin: 4, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: POSITION, Origin: 5, Iteration: 1, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: POSITION, Origin: 6, Target: 7, Iteration: 2, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: ACK, Origin: 9, Target: 8, Iteration: 3, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: ACK, Origin: 11, Target: 10, Iteration: 4, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: METRICS, Origin: 12, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: METRICS_ACK, Target: 13, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: GROUP_DONE, Target: 14, Limit: 1, Drop: true},
		{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 15, Limit: 1, Drop: true},
		{Direction: lossToClients, Type: START, Target: 16, Limit: 1, Duplicate: true},
		{Direction: lossToServer, Type: METRICS, Origin: 17, Limit: 1, Duplicate: true},
		{Direction: lossToClients, Type: GROUP_DONE, Target: 18, Limit: 1, Duplicate: true},
		{Direction: lossToClients, Type: START, Target: 19, Limit: 1, Delay: 250 * time.Millisecond},
		{Direction: lossToClients, Type: ACK, Origin: 21, Target: 20, Iteration: 1, Limit: 1, Duplicate: true},
	}
	begin := time.Now()
	s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 100, 25, 5, 6), rules, true, nil)
	sst, gst, pst := simulationCompleteAndClose(t, ctx, s, g, p)
	for index := range rules {
		if pst.RuleMatches[index] != 1 {
			t.Fatal("mixed fault was not exercised exactly once", index, pst)
		}
	}
	if gst.RegisterRetries == 0 || gst.PositionRetries == 0 || gst.MetricsRetries == 0 || sst.StartRetries == 0 || sst.GroupDoneRetries == 0 || gst.DuplicateStarts == 0 || gst.DuplicateGroupDone == 0 || sst.DuplicateMetrics == 0 || pst.Dropped != 12 || pst.Duplicated != 4 || pst.Delayed != 1 {
		t.Fatal("mixed schedule did not exercise all reliability mechanisms", sst, gst, pst)
	}
	t.Logf("N100/V25/I5 mixed: 100registered/100started/500complete/12000uniqueACK/100metrics/100DONE/100uniqueServerDONEACK (rawDONEACK=%d); failed0; elapsed=%s; retries REGISTER=%d POSITION=%d METRICS=%d START=%d GROUP_DONE=%d requestedNeighbors=%d; proxy=%+v", sst.GroupDoneAckReceived, time.Since(begin), gst.RegisterRetries, gst.PositionRetries, gst.MetricsRetries, sst.StartRetries, sst.GroupDoneRetries, gst.RetryNeighborsRequested, pst)
}

func TestSimulationControlLossRecovery(t *testing.T) {
	for _, tc := range []struct {
		name string
		rule lossRule
	}{
		{"START-lost", lossRule{Direction: lossToClients, Type: START, Target: 1, Drop: true}},
		{"START_ACK-lost", lossRule{Direction: lossToServer, Type: START_ACK, Origin: 1, Drop: true}},
		{"START-duplicated", lossRule{Direction: lossToClients, Type: START, Target: 1, Duplicate: true}},
		{"START-delayed", lossRule{Direction: lossToClients, Type: START, Target: 1, Delay: 250 * time.Millisecond}},
		{"METRICS-lost", lossRule{Direction: lossToServer, Type: METRICS, Origin: 1, Drop: true}},
		{"METRICS_ACK-lost", lossRule{Direction: lossToClients, Type: METRICS_ACK, Target: 1, Drop: true}},
		{"METRICS-duplicated", lossRule{Direction: lossToServer, Type: METRICS, Origin: 1, Duplicate: true}},
		{"GROUP_DONE-lost", lossRule{Direction: lossToClients, Type: GROUP_DONE, Target: 1, Drop: true}},
		{"GROUP_DONE_ACK-lost", lossRule{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 1, Drop: true}},
		{"GROUP_DONE-duplicated", lossRule{Direction: lossToClients, Type: GROUP_DONE, Target: 1, Duplicate: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tc.rule.Limit = 1
			s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 6, 3, 2, 2), []lossRule{tc.rule}, true, nil)
			if tc.rule.Delay > 0 {
				// Proxy DelayCurrent=0 means sendto returned, not that the late
				// START and its reactive ACK have passed both owner workers.
				// Observe the actual round trip before attempting terminal drain.
				simulationWaitCondition(t, ctx, s, g, p, "delayed START acknowledged", func(st ServerStats, ct ClientStats, _ lossProxyStats) bool {
					return ct.DuplicateStarts > 0 && st.DuplicateStartACKs > 0
				})
			}
			sst, gst, pst := simulationCompleteAndClose(t, ctx, s, g, p)
			if pst.RuleMatches[0] != 1 {
				t.Fatal("directed control fault was not injected", pst)
			}
			t.Logf("complete6 operations12 uniqueACK24; retries START=%d METRICS=%d GROUP_DONE=%d duplicatedSTART=%d duplicatedDONE=%d duplicatedMETRICS=%d", sst.StartRetries, gst.MetricsRetries, sst.GroupDoneRetries, gst.DuplicateStarts, gst.DuplicateGroupDone, sst.DuplicateMetrics)
		})
	}
}

func TestSimulationSTARTBeforeLocalREGISTER_ACK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rules := []lossRule{{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 1, Limit: 1, Drop: true}}
	s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 3, 3, 2, 1), rules, true, nil)
	_, gst, pst := simulationCompleteAndClose(t, ctx, s, g, p)
	if pst.StartBeforeRegisterACK == 0 || gst.RegisterRetries == 0 || gst.StartedClients != 3 || gst.CompletedOperations != 6 {
		t.Fatal("START before local REGISTER_ACK was lost or prematurely acknowledged", gst, pst)
	}
	t.Log("START observed before local REGISTER_ACK, retained until valid registration; all clients then execute exactly I iterations")
}

func TestSimulationGROUP_DONEBeforeMETRICS_ACK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Every explicit METRICS_ACK for client 1 is lost. GROUP_DONE provides the
	// approved implicit acceptance and must cancel its METRICS retry instead.
	rules := []lossRule{{Direction: lossToClients, Type: METRICS_ACK, Target: 1, AnyCounter: true, Drop: true}}
	s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 3, 3, 2, 1), rules, true, nil)
	_, gst, pst := simulationCompleteAndClose(t, ctx, s, g, p)
	if pst.RuleMatches[0] == 0 || gst.MetricsAckReceived >= 3 || gst.MetricsAccepted != 3 || gst.DoneClients != 3 || gst.FailedMetrics != 0 {
		t.Fatal("GROUP_DONE did not implicitly accept the lost METRICS_ACK", gst, pst)
	}
	t.Log("GROUP_DONE acknowledges accepted metrics implicitly, cancels METRICS ARQ, reaches DONE and generates GROUP_DONE_ACK")
}

func TestSimulationLastGROUP_DONE_ACKKeepsGeneratorAlive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rules := []lossRule{{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 3, Limit: 1, Drop: true}}
	s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 3, 3, 1, 1), rules, true, nil)
	if err := g.WaitAllDone(ctx); err != nil {
		t.Fatal(err)
	}
	simulationWaitCondition(t, ctx, s, g, p, "final ACK intentionally lost", func(sst ServerStats, gst ClientStats, pst lossProxyStats) bool {
		return gst.DoneClients == 3 && pst.RuleMatches[0] == 1
	})
	if s.Stats().SimulationComplete {
		t.Fatal("server declared completion before the lost final ACK was recovered")
	}
	select {
	case <-g.done:
		t.Fatal("DONE clients closed their shared socket before server confirmation")
	default:
	}
	sst, gst, _ := simulationCompleteAndClose(t, ctx, s, g, p)
	if sst.GroupDoneRetries != 1 || gst.DuplicateGroupDone == 0 || gst.GroupDoneAckSent != 4 {
		t.Fatal("last ACK recovery did not use the still-active DONE client", sst, gst)
	}
}

func TestSimulationFastClientKeepsHelpingSlowNeighbor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rules := []lossRule{{Direction: lossToServer, Type: POSITION, Origin: 2, Iteration: 1, Limit: 1, Delay: 200 * time.Millisecond}}
	scfg := simulationServerConfig(t, 3, 3, 2, 1)
	s, g, p := simulationListen(t, ctx, scfg, rules, true, func(cfg *ClientConfig) {
		cfg.InitialRTO, cfg.MaxRTO = time.Second, 2*time.Second // delay is below RTO, so only client 2's first data operation is slowed
	})
	simulationWaitCondition(t, ctx, s, g, p, "fast clients report metrics while slow neighbor is outstanding", func(sst ServerStats, gst ClientStats, pst lossProxyStats) bool {
		return sst.MetricsAccepted == 2 && pst.DelayCurrent == 1
	})
	if s.Stats().GroupDoneSent != 0 || g.Stats().DoneClients != 0 || s.Stats().SimulationComplete {
		t.Fatal("GROUP_DONE/cancellation occurred before the slow client's metrics")
	}
	_, _, _ = simulationCompleteAndClose(t, ctx, s, g, p)
	if p.neighborACKAfterMetrics[0]&(1<<1) == 0 || p.metricsFirstAt[0] >= p.metricsFirstAt[1] || p.groupDoneFirstAt[0] <= p.metricsFirstAt[1] {
		t.Fatal("fast client stopped acknowledging the slow neighbor after sending METRICS")
	}
	t.Log("fast client 1 sends METRICS first and then sends real UDP ACKs for slow client 2; GROUP_DONE follows all three accepted metrics")
}

func TestSimulationGROUP_DONEIsPerGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// One server-bound operation delays group B; group A has no dependency on B.
	rules := []lossRule{{Direction: lossToServer, Type: POSITION, Origin: 4, Iteration: 1, Limit: 1, Delay: 200 * time.Millisecond}}
	s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 6, 3, 2, 2), rules, true, func(cfg *ClientConfig) {
		cfg.InitialRTO, cfg.MaxRTO = time.Second, 2*time.Second
	})
	simulationWaitCondition(t, ctx, s, g, p, "group A DONE while group B still runs", func(sst ServerStats, gst ClientStats, pst lossProxyStats) bool {
		return sst.CompletedGroups == 1 && gst.DoneClients == 3 && pst.DelayCurrent == 1
	})
	if s.Stats().SimulationComplete || s.Stats().MetricsAccepted == 6 {
		t.Fatal("global completion preceded slow group B")
	}
	_, _, _ = simulationCompleteAndClose(t, ctx, s, g, p)
	if !p.groupDoneSeen[0] || !p.groupDoneSeen[1] || p.groupDoneFirstAt[0] >= p.metricsFirstAt[3] || p.groupDoneFirstAt[1] <= p.metricsFirstAt[3] {
		t.Fatal("GROUP_DONE for A waited for B or DONE for B preceded B metrics")
	}
	t.Log("group A receives/acknowledges GROUP_DONE while group B is still delayed; global completion waits for B")
}

func TestSimulationPermanentControlLossExhausts(t *testing.T) {
	for _, typ := range []MessageType{START_ACK, GROUP_DONE_ACK} {
		t.Run(map[MessageType]string{START_ACK: "START_ACK", GROUP_DONE_ACK: "GROUP_DONE_ACK"}[typ], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			scfg := simulationServerConfig(t, 3, 3, 1, 1)
			scfg.InitialControlRTO, scfg.MaxControlRTO, scfg.MaxControlSends = 20*time.Millisecond, 40*time.Millisecond, 3
			rules := []lossRule{{Direction: lossToServer, Type: typ, Origin: 1, AnyCounter: true, Drop: true}}
			s, g, p := simulationListen(t, ctx, scfg, rules, true, nil)
			if err := s.WaitSimulationComplete(ctx); err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("permanent control loss did not produce bounded explicit failure", err)
			}
			sst := s.Stats()
			if sst.SimulationComplete || sst.ControlRetryExhausted != 1 || sst.ControlHeapCurrent != 0 || p.snapshot().RuleMatches[0] != 3 {
				t.Fatal("permanent control loss exceeded MaxControlSends or leaked deadline", sst, p.snapshot())
			}
			if typ == START_ACK && (sst.StartRetries != 2 || sst.StartedClients != 2) || typ == GROUP_DONE_ACK && (sst.GroupDoneRetries != 2 || sst.DoneClients != 2 || g.Stats().DoneClients != 3) {
				t.Fatal("server repeated credited members or declared completion without final ACK", sst, g.Stats())
			}
			t.Logf("permanent type=%d: MaxControlSends3, retries2 for only missing member, exhausted1, SimulationComplete=false", typ)
		})
	}
}

func TestSimulationPermanentMETRICSLossExhausts(t *testing.T) {
	for _, ackLost := range []bool{false, true} {
		name := "METRICS"
		if ackLost {
			name = "METRICS_ACK-and-implicit-GROUP_DONE"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			rules := []lossRule{{Direction: lossToServer, Type: METRICS, Origin: 1, AnyCounter: true, Drop: true}}
			if ackLost {
				// GROUP_DONE is a valid alternative receipt. Suppress both receipts
				// to demonstrate METRICS exhaustion rather than implicit recovery.
				rules = []lossRule{{Direction: lossToClients, Type: METRICS_ACK, Target: 1, AnyCounter: true, Drop: true},
					{Direction: lossToClients, Type: GROUP_DONE, Target: 1, AnyCounter: true, Drop: true}}
			}
			s, g, p := simulationListen(t, ctx, simulationServerConfig(t, 3, 3, 1, 1), rules, true, func(cfg *ClientConfig) {
				cfg.InitialRTO, cfg.MaxRTO, cfg.MaxSends = 20*time.Millisecond, 40*time.Millisecond, 3
			})
			if err := g.WaitAllDone(ctx); err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("permanent METRICS loss did not produce bounded explicit client failure", err)
			}
			gst, sst, pst := g.Stats(), s.Stats(), p.snapshot()
			if gst.FailedMetrics != 1 || gst.RetryExhausted != 1 || gst.MetricsRetries != 2 || gst.DeadlineHeapCurrent != 0 || sst.SimulationComplete || pst.RuleMatches[0] != 3 {
				t.Fatal("METRICS exhaustion exceeded MaxSends or failed ambiguously", gst, sst, pst)
			}
			t.Log("METRICS original + 2 identical retries, one explicit FailedMetrics/RetryExhausted; global simulation remains incomplete")
		})
	}
}
