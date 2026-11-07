package main

import (
	m "arc.local/multiplexado/internal/measurement"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGeneratorPriorityUsesExistingClientKnob(t *testing.T) {
	if generatorPriorityArgs(8) != nil {
		t.Fatal("default invocation changed")
	}
	if got := generatorPriorityArgs(24); !reflect.DeepEqual(got, []string{"-high-burst", "24"}) {
		t.Fatal("wrong generator-only knob", got)
	}
}

func TestGeneratorHighCapacityUsesExistingClientKnob(t *testing.T) {
	if generatorHighCapacityArgs(1024) != nil {
		t.Fatal("default invocation changed")
	}
	if got := generatorHighCapacityArgs(2048); !reflect.DeepEqual(got, []string{"-high-capacity", "2048"}) {
		t.Fatal("wrong generator-only capacity knob", got)
	}
}

func TestBSDTimeParsing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "time.txt")
	data := "server PID=123 success=true\n        2.50 real         1.75 user         3.25 sys\n            16777216  maximum resident set size\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	var e external
	if err := parseTime(path, &e); err != nil || e.CPUPercent != 200 || e.PeakRSSBytes != 16777216 {
		t.Fatal(e, err)
	}
	os.WriteFile(path, []byte("invalid incomplete measurement"), 0600)
	if parseTime(path, &e) == nil {
		t.Fatal("missing measurements silently accepted")
	}
}

func TestReadyTransportExactSourceAuthorization(t *testing.T) {
	shared := m.Ready{Address: "127.0.0.1:20000", IngressEndpoint: "127.0.0.1:20000", ServerEgressEndpoints: []string{"127.0.0.1:20000"}}
	ingress, sources, err := serverTransport(shared, 1)
	if err != nil || ingress != shared.Address || sources != shared.Address {
		t.Fatal(ingress, sources, err)
	}
	separate := shared
	separate.ServerEgressEndpoints = []string{"127.0.0.1:20001", "127.0.0.1:20002"}
	ingress, sources, err = serverTransport(separate, 2)
	if err != nil || ingress != shared.Address || sources != "127.0.0.1:20001,127.0.0.1:20002" {
		t.Fatal(ingress, sources, err)
	}
	for _, bad := range []m.Ready{
		{Address: shared.Address, IngressEndpoint: "127.0.0.1:20003", ServerEgressEndpoints: separate.ServerEgressEndpoints},
		{Address: shared.Address, ServerEgressEndpoints: []string{shared.Address, "127.0.0.1:20002"}},
		{Address: shared.Address, ServerEgressEndpoints: []string{"127.0.0.1:20001", "127.0.0.1:20001"}},
		{Address: shared.Address, ServerEgressEndpoints: []string{"127.0.0.1:20001"}},
		{Address: shared.Address, ServerEgressEndpoints: []string{"0.0.0.0:20001", "127.0.0.1:20002"}},
	} {
		if _, _, err := serverTransport(bad, 2); err == nil {
			t.Fatal("invalid READY accepted", bad)
		}
	}
	old := m.Ready{Address: shared.Address}
	if _, sources, err := serverTransport(old, 1); err != nil || sources != "" {
		t.Fatal("old default binary incompatible")
	}
	if _, _, err := serverTransport(old, 2); err == nil {
		t.Fatal("two egress accepted without explicit READY sources")
	}
}
