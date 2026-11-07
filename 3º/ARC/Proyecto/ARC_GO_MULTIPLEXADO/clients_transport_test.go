package protocol

import (
	"net/netip"
	"testing"
)

func TestClientAuthorizedServerSourcesAndCanonicalDestination(t *testing.T) {
	cfg := clientsConfig(t, 25, 25, 5, 6)
	ingress := cfg.Server
	first, second, third := testEndpoint(10), testEndpoint(11), testEndpoint(12)
	cfg.ServerEgressEndpoints = []netip.AddrPort{first, second}
	g := bareClients(t, cfg)
	p := clientRegistrationReply(g, 1)
	for _, source := range []netip.AddrPort{first, second} {
		if err := g.decodeDispatch(wire(t, p), source, 1); err != nil {
			t.Fatal("authorized endpoint rejected", source, err)
		}
		if g.cfg.Server != ingress {
			t.Fatal("RX changed canonical destination")
		}
		g.workers[0].input.pop()
	}
	for _, source := range []netip.AddrPort{third, ingress, netip.AddrPort{}} {
		if err := g.decodeDispatch(wire(t, p), source, 1); err != ErrEndpoint {
			t.Fatal("unauthorized endpoint accepted", source, err)
		}
	}
	bad := p
	bad.RunID++
	if err := g.decodeDispatch(wire(t, bad), first, 1); err == nil {
		t.Fatal("source list bypassed RunID validation")
	}
	badWire := wire(t, p)
	badWire[6], badWire[7] = 0, 0 // corrupt direction after Encode's own validation
	if err := g.decodeDispatch(badWire, first, 1); err == nil {
		t.Fatal("source list bypassed direction validation")
	}
	// Freeze authorization against caller mutation, including slice aliasing.
	cfg.ServerEgressEndpoints[0] = third
	if !g.acceptsServerSource(first) || g.acceptsServerSource(third) || g.cfg.ServerEgressEndpoints[0] != first {
		t.Fatal("authorization retained caller's mutable slice")
	}
	if testing.AllocsPerRun(1000, func() { g.acceptsServerSource(second) }) != 0 {
		t.Fatal("source validation allocated")
	}
	// Default shared-socket behavior remains canonical-only.
	cfg.ServerEgressEndpoints = nil
	legacy := bareClients(t, cfg)
	if !legacy.acceptsServerSource(ingress) || legacy.acceptsServerSource(first) {
		t.Fatal("canonical-only default weakened")
	}
}

func TestClientSourceConfigurationRejectsInvalidAndNormalizes(t *testing.T) {
	first := testEndpoint(10)
	mapped := netip.MustParseAddrPort("[::ffff:127.0.0.1]:21010")
	for _, sources := range [][]netip.AddrPort{
		{{}}, {netip.MustParseAddrPort("127.0.0.1:0")}, {netip.MustParseAddrPort("0.0.0.0:1234")},
		{netip.MustParseAddrPort("[::]:1234")}, {netip.MustParseAddrPort("224.0.0.1:1234")},
		{first, first}, {first, mapped}, {first, testEndpoint(11), testEndpoint(12), testEndpoint(13)},
	} {
		cfg := clientsConfig(t, 25, 25, 5, 6)
		cfg.ServerEgressEndpoints = sources
		if _, err := newClientGenerator(cfg); err != ErrClientConfig {
			t.Fatal("invalid source configuration accepted", sources, err)
		}
	}
	cfg := clientsConfig(t, 25, 25, 5, 6)
	cfg.ServerEgressEndpoints = []netip.AddrPort{mapped}
	g := bareClients(t, cfg)
	if !g.acceptsServerSource(first) || !g.acceptsServerSource(mapped) || g.serverSources[0] != first {
		t.Fatal("IPv4 mapped normalization")
	}
}
