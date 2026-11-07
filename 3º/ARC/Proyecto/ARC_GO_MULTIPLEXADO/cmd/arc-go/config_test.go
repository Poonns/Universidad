package main

import (
	"io"
	"net/netip"
	"reflect"
	"strconv"
	"testing"
)

func TestManifestAndFlagBounds(t *testing.T) {
	good := []string{"-n", "1000", "-session", "test", "-result", "server.json"}
	o, m, err := parseOptions("server", good, io.Discard)
	if err != nil || m.Config().N != 1000 || o.Address != "127.0.0.1:9000" || o.Count != 1000 {
		t.Fatal(o, err)
	}
	for _, extra := range [][]string{{"-first-id", "2"}, {"-count", "975"}, {"-n", "1001"}, {"-n", "4294967296"}, {"-max-sends", "65536"},
		{"-academic=false"}, {"-timeout", "0"}, {"-sample-interval", "0"}, {"-session", ""}} {
		args := append(append([]string{}, good...), extra...)
		if _, _, err := parseOptions("server", args, io.Discard); err == nil {
			t.Fatal("invalid accepted", args)
		}
	}
	client := []string{"-session", "test", "-server-result", "server.json", "-position-rate", "2500", "-position-window", "256"}
	o, m, err = parseOptions("client", client, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	c, err := o.clientConfig(m)
	if err != nil || c.PositionRate != 2500 || c.PositionWindow != 256 || c.Workers != 6 || !c.Academic {
		t.Fatal(c, err)
	}
}

func TestGeneratorHighBurstExistingConfiguration(t *testing.T) {
	for _, burst := range []string{"8", "24"} {
		o, m, err := parseOptions("client", []string{"-session", "test", "-server-result", "server.json", "-high-burst", burst}, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		c, err := o.clientConfig(m)
		if err != nil || c.HighBurst != o.HighBurst || c.AckBurst != 4 || c.HighCapacity != 1024 || c.DataCapacity != 2048 || c.PositionWindow != 128 {
			t.Fatal("priority knob changed other configuration", c, err)
		}
	}
}

func TestGeneratorHighCapacityExistingConfiguration(t *testing.T) {
	base := []string{"-session", "test", "-server-result", "server.json", "-high-burst", "24"}
	o, m, err := parseOptions("client", base, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := o.clientConfig(m)
	if err != nil || baseline.HighCapacity != 1024 {
		t.Fatal("HIGH default changed", baseline, err)
	}
	for _, capacity := range []int{1024, 2048} {
		o, m, err := parseOptions("client", append(append([]string{}, base...), "-high-capacity", strconv.Itoa(capacity)), io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		c, err := o.clientConfig(m)
		if err != nil || c.HighCapacity != capacity {
			t.Fatal("HIGH capacity not applied", c, err)
		}
		c.HighCapacity, c.Manifest = baseline.HighCapacity, baseline.Manifest
		if !reflect.DeepEqual(c, baseline) {
			t.Fatal("HIGH capacity changed another client setting")
		}
	}
}

func TestEgressFlagsSourcesAndCanonicalIngress(t *testing.T) {
	base := []string{"-session", "test", "-result", "server.json", "-server-writers", "2"}
	for _, count := range []string{"1", "2"} {
		o, m, err := parseOptions("server", append(append([]string{}, base...), "-server-egress-sockets", count), io.Discard)
		if err != nil || o.serverConfig(m).ServerEgressSocketCount != o.ServerEgressSocketCount {
			t.Fatal(o, err)
		}
	}
	for _, count := range []string{"0", "3", "4", "-1"} {
		if _, _, err := parseOptions("server", append(append([]string{}, base...), "-server-egress-sockets", count), io.Discard); err == nil {
			t.Fatal("invalid accepted", count)
		}
	}
	if _, _, err := parseOptions("server", []string{"-session", "test", "-result", "server.json", "-server-egress-sockets", "2"}, io.Discard); err == nil {
		t.Fatal("two sockets accepted with one writer")
	}
	client := []string{"-session", "test", "-server-result", "server.json", "-server", "127.0.0.1:20000"}
	o, m, err := parseOptions("client", append(append([]string{}, client...), "-server-egress-endpoints", "127.0.0.1:20001,127.0.0.1:20002"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	c, err := o.clientConfig(m)
	canonical := netip.AddrPortFrom(c.Server.Addr().Unmap(), c.Server.Port()) // ListenClients normalizes net.ResolveUDPAddr's IPv4-mapped result
	if err != nil || canonical != netip.MustParseAddrPort("127.0.0.1:20000") || len(c.ServerEgressEndpoints) != 2 || c.ServerEgressEndpoints[1] != netip.MustParseAddrPort("127.0.0.1:20002") {
		t.Fatal(c, err)
	}
	for _, sources := range []string{"any", "127.0.0.1:0", "0.0.0.0:1234", "127.0.0.1:20001,127.0.0.1:20001", "127.0.0.1:1,127.0.0.1:2,127.0.0.1:3,127.0.0.1:4", "127.0.0.1:1,"} {
		o, m, err := parseOptions("client", append(append([]string{}, client...), "-server-egress-endpoints", sources), io.Discard)
		if err == nil {
			if _, err = o.clientConfig(m); err == nil {
				t.Fatal("invalid sources accepted", sources)
			}
		}
	}
}
func TestUsageExitCodes(t *testing.T) {
	if run(nil) != 2 || run([]string{"unknown"}) != 2 || run([]string{"server", "-help"}) != 0 {
		t.Fatal("bad usage exit code")
	}
}

func TestServerWriterFlagAndUnchangedClient(t *testing.T) {
	base := []string{"-session", "test", "-result", "server.json"}
	for _, count := range []string{"1", "2"} {
		o, m, err := parseOptions("server", append(append([]string{}, base...), "-server-writers", count), io.Discard)
		if err != nil || o.serverConfig(m).ServerWriterCount != o.ServerWriterCount || o.ServerWriterCount < 1 || o.ServerWriterCount > 2 {
			t.Fatal(o, err)
		}
	}
	for _, count := range []string{"0", "3", "4", "-1"} {
		if _, _, err := parseOptions("server", append(append([]string{}, base...), "-server-writers", count), io.Discard); err == nil {
			t.Fatal("invalid accepted", count)
		}
	}
	if _, _, err := parseOptions("client", []string{"-session", "test", "-server-result", "server.json", "-server-writers", "2"}, io.Discard); err == nil {
		t.Fatal("server flag accepted by generator")
	}
}
