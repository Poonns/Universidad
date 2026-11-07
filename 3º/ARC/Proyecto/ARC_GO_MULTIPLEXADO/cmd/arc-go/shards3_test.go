package main

import (
	"io"
	"net/netip"
	"testing"
)

func TestThreeTXShardsCLIAndUnchangedDefaults(t *testing.T) {
	base := []string{"-session", "test", "-result", "server.json"}
	o, _, err := parseOptions("server", base, io.Discard)
	if err != nil || o.ServerWriterCount != 1 || o.ServerEgressSocketCount != 1 {
		t.Fatal("public defaults changed", o, err)
	}
	o, m, err := parseOptions("server", append(base, "-server-writers", "3", "-server-egress-sockets", "3"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	c := o.serverConfig(m)
	if c.ServerWriterCount != 3 || c.ServerEgressSocketCount != 3 || c.Workers != 6 || c.GOMAXPROCS != 6 {
		t.Fatal(c)
	}
	for _, sockets := range []string{"1", "2", "4"} {
		if _, _, err := parseOptions("server", append(append([]string{}, base...), "-server-writers", "3", "-server-egress-sockets", sockets), io.Discard); err == nil {
			t.Fatal("nonexclusive three-writer mode accepted", sockets)
		}
	}
	o, m, err = parseOptions("client", []string{"-session", "test", "-server-result", "server.json", "-server", "127.0.0.1:20000", "-server-egress-endpoints", "127.0.0.1:20001,127.0.0.1:20002,127.0.0.1:20003"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	g, err := o.clientConfig(m)
	if err != nil || len(g.ServerEgressEndpoints) != 3 || g.ServerEgressEndpoints[2] != netip.MustParseAddrPort("127.0.0.1:20003") || g.Server.Port() != 20000 {
		t.Fatal("three sources altered canonical destination", g, err)
	}
}
