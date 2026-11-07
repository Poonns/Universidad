package protocol

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGeneratorEndpointsCannotCrossOrMigrate(t *testing.T) {
	s := bareServer(t, serverConfig(t, 100, 25, 6, true))
	registerAll(t, s)
	for gi, id := range []uint32{1, 51} {
		other := testEndpoint(1 - gi)
		p := positionPacket(s, id)
		for _, forged := range []Packet{registerPacket(s, gi, 1), p, ackFor(p, id+1)} {
			if err := deliver(t, s, forged, other); err != ErrEndpoint {
				t.Fatalf("gid%d type%d cross-endpoint accepted: %v", p.GeneratorID, forged.Type, err)
			}
		}
		if err := deliver(t, s, p, testEndpoint(gi)); err != nil {
			t.Fatal("authorized endpoint no longer valid", err)
		}
		if current := s.endpoints[gi].address.Load(); current == nil || *current != testEndpoint(gi) {
			t.Fatal("endpoint migrated")
		}
	}
	if s.Stats().EndpointRejects != 6 {
		t.Fatal(s.Stats().EndpointRejects)
	}
}

func TestTwoGeneratorsAcademicGlobalBarrierAndCompleteRanges(t *testing.T) {
	for _, shards := range []int{2, 3} {
		t.Run(fmt.Sprintf("TXShards%d", shards), func(t *testing.T) {
			twoGeneratorsAcademicComplete(t, shards)
		})
	}
}

func twoGeneratorsAcademicComplete(t *testing.T, shards int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg := simulationServerConfig(t, 100, 25, 5, 6)
	run := cfg.Manifest.Config()
	var err error
	cfg.Manifest, err = NewManifest(run, []GeneratorConfig{{7, 1, 50, run}, {9, 51, 50, run}})
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerWriterCount, cfg.ServerEgressSocketCount = shards, shards
	cfg.InitialControlRTO, cfg.MaxControlRTO = time.Second, 3*time.Second
	s, err := ListenServer(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	makeGenerator := func(id uint32) *ClientGenerator {
		c := simulationClientConfig(cfg.Manifest, s.Addr(), 3)
		c.GeneratorID = id
		c.ServerEgressEndpoints = s.EgressEndpoints()
		c.InitialRTO, c.MaxRTO = time.Second, 3*time.Second
		g, e := ListenClients(ctx, c)
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() { g.Close() })
		return g
	}
	g1 := makeGenerator(7)
	simulationWaitCondition(t, ctx, s, g1, nil, "first range fully registered", func(st ServerStats, ct ClientStats, _ lossProxyStats) bool {
		return st.RegisteredClients == 50 && ct.RegisteredClients == 50
	})
	if st := s.Stats(); st.StartSent != 0 || st.RegistrationBarrierReached || st.SimulationComplete {
		t.Fatal("START before global registration", st)
	}
	g2 := makeGenerator(9)
	if g1.Addr() == g2.Addr() {
		t.Fatal("generator endpoints not independent")
	}
	if err = s.WaitSimulationComplete(ctx); err != nil {
		t.Fatal(err)
	}
	for _, g := range []*ClientGenerator{g1, g2} {
		if err = g.WaitAllDone(ctx); err != nil {
			t.Fatal(err)
		}
		st := g.Stats()
		if st.LogicalClients != 50 || st.RegisteredClients != 50 || st.StartedClients != 50 || st.CompletedOperations != 250 || st.UniqueAcks != 6000 || st.MetricsAccepted != 50 || st.DoneClients != 50 || st.GroupDoneAckSent != 50 || st.QueueDrops+st.SocketErrors+st.FailedOperations != 0 {
			t.Fatal(st)
		}
	}
	st := s.Stats()
	if st.RegisteredClients != 100 || st.StartedClients != 100 || st.MetricsAccepted != 100 || st.DoneClients != 100 || st.GroupDoneAckReceived != 100 || st.CompletedGroups != 4 || !st.SimulationComplete || st.EndpointRejects+st.QueueDrops+st.SocketErrors != 0 {
		t.Fatal(st)
	}
}
