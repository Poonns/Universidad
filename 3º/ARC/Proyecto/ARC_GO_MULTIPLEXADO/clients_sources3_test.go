package protocol

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

func TestClientsAuthorizeExactlyThreeServerSources(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 5, 3)
	ingress := cfg.Server
	sources := []netip.AddrPort{testEndpoint(10), testEndpoint(11), testEndpoint(12)}
	fourth := testEndpoint(13)
	cfg.ServerEgressEndpoints = sources
	g := bareClients(t, cfg)
	packet := clientRegistrationReply(g, 1)
	for _, source := range sources {
		if err := g.decodeDispatch(wire(t, packet), source, 1); err != nil {
			t.Fatal("published source rejected", source, err)
		}
		g.workers[0].input.pop()
	}
	for _, source := range []netip.AddrPort{fourth, ingress, netip.AddrPort{}} {
		if err := g.decodeDispatch(wire(t, packet), source, 1); err != ErrEndpoint {
			t.Fatal("unpublished source accepted", source, err)
		}
	}
	sources[0] = fourth
	if g.acceptsServerSource(fourth) || !g.acceptsServerSource(testEndpoint(10)) || g.cfg.Server != ingress {
		t.Fatal("mutable authorization or changed canonical destination")
	}
	if testing.AllocsPerRun(1000, func() { g.acceptsServerSource(testEndpoint(12)) }) != 0 {
		t.Fatal("source check allocates")
	}
	bad := packet
	bad.RunID++
	if err := g.decodeDispatch(wire(t, bad), testEndpoint(12), 1); err == nil {
		t.Fatal("third source bypassed protocol validation")
	}
}

func TestServerThreeTXShardsAcademicIngressAndLossRecovery(t *testing.T) {
	for _, loss := range []bool{false, true} {
		label := "direct"
		if loss {
			label = "directed-loss"
		}
		t.Run(label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cfg := simulationServerConfig(t, 100, 25, 5, 6)
			cfg.ServerWriterCount, cfg.ServerEgressSocketCount = 3, 3
			var rules []lossRule
			if loss {
				rules = []lossRule{
					{Direction: lossToServer, Type: REGISTER, Nonce: 1, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: REGISTER_ACK, Nonce: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: START, Target: 51, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: START_ACK, Origin: 26, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: POSITION, Origin: 1, Target: 2, Iteration: 2, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: ACK, Origin: 52, Target: 51, Iteration: 3, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: METRICS, Origin: 1, Limit: 1, Drop: true},
					{Direction: lossToClients, Type: GROUP_DONE, Target: 51, Limit: 1, Drop: true},
					{Direction: lossToServer, Type: GROUP_DONE_ACK, Origin: 26, Limit: 1, Drop: true},
				}
			}
			s, g, p := simulationListen(t, ctx, cfg, rules, loss, nil)
			endpoints := s.EgressEndpoints()
			if len(endpoints) != 3 {
				t.Fatal("missing READY sources")
			}
			for i := range s.writers {
				if s.writers[i].connection() != s.egress[i] || s.egress[i] == s.conn || endpoints[i] == s.Addr() {
					t.Fatal("nonexclusive TX descriptor")
				}
				for j := 0; j < i; j++ {
					if s.egress[i] == s.egress[j] || endpoints[i] == endpoints[j] {
						t.Fatal("duplicate TX socket/endpoint")
					}
				}
			}
			st, ct, ps := simulationCompleteAndClose(t, ctx, s, g, p)
			assertEgressAccounting(t, st, 3)
			if loss {
				if ps.Dropped != uint64(len(rules)) || ct.RegisterRetries == 0 || ct.PositionRetries == 0 || ct.MetricsRetries == 0 || st.StartRetries == 0 || st.GroupDoneRetries == 0 {
					t.Fatal("directed recovery not exercised", ps, ct, st)
				}
				for _, n := range ps.RuleMatches[:len(rules)] {
					if n != 1 {
						t.Fatal("unexercised loss rule", ps)
					}
				}
				if g.cfg.Server != normalizeEndpoint(p.Addr()) || p.server != normalizeEndpoint(s.Addr()) {
					t.Fatal("recovery redirected client traffic to egress")
				}
			} else {
				if g.cfg.Server != normalizeEndpoint(s.Addr()) || st.PacketsReceived != ct.PacketsSent || ct.PacketsReceived != st.PacketsSent {
					t.Fatal("canonical ingress or packet accounting changed")
				}
				if st.PacketsSent != 24400 || st.PacketsReceived != 12900 || st.StartRetries+st.GroupDoneRetries+ct.PositionRetries+ct.RegisterRetries+ct.MetricsRetries != 0 {
					t.Fatal("unexpected traffic/retries in direct run")
				}
				// Four groups: modulo assigns 2,1,1 groups to the three writers.
				for i, w := range st.Writers {
					groups := uint64(1)
					if i == 0 {
						groups = 2
					}
					if w.Writes != groups*6100 || w.HighWrites != groups*3050 || w.ControlWrites != groups*50 || w.DataWrites != groups*3000 {
						t.Fatal("unexplained academic TX distribution", w)
					}
				}
			}
		})
	}
}
