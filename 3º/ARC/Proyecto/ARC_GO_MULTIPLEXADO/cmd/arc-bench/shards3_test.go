package main

import (
	m "arc.local/multiplexado/internal/measurement"
	"strings"
	"testing"
)

func TestReadyThreeTXShardsExactSources(t *testing.T) {
	r := m.Ready{Address: "127.0.0.1:20000", IngressEndpoint: "127.0.0.1:20000", ServerEgressEndpoints: []string{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003"}}
	in, sources, err := serverTransport(r, 3)
	if err != nil || in != r.Address || sources != strings.Join(r.ServerEgressEndpoints, ",") {
		t.Fatal("invalid three-source wiring", in, sources, err)
	}
	for _, bad := range [][]string{
		{r.Address, "127.0.0.1:20002", "127.0.0.1:20003"},
		{"127.0.0.1:20001", "127.0.0.1:20001", "127.0.0.1:20003"},
		{"127.0.0.1:20001", "127.0.0.1:20002"},
		{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003", "127.0.0.1:20004"},
	} {
		r.ServerEgressEndpoints = bad
		if _, _, err := serverTransport(r, 3); err == nil {
			t.Fatal("invalid READY accepted", bad)
		}
	}
}
