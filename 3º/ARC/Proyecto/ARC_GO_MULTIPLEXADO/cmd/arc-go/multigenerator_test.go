package main

import (
	"io"
	"testing"
)

func TestCompleteManifestWiringAndLocalRange(t *testing.T) {
	ranges := "1:1:15000,2:15001:15000"
	base := []string{"-n", "30000", "-session", "multi", "-generator-ranges", ranges}
	_, manifest, err := parseOptions("server", append(append([]string{}, base...), "-result", "server.json"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id, first, count string
		nonce            uint64
		assigned         uint32
	}{
		{"1", "1", "15000", 15000, 15000}, {"2", "15001", "15000", 1, 15001},
	} {
		args := append(append([]string{}, base...), "-generator-id", tc.id, "-first-id", tc.first, "-count", tc.count, "-server-result", "server.json")
		o, m, err := parseOptions("client", args, io.Discard)
		if err != nil || m.Config() != manifest.Config() {
			t.Fatal(o, err)
		}
		id, _, err := m.RegistrationIdentity(uint32(o.GeneratorID), tc.nonce)
		if err != nil || id != tc.assigned {
			t.Fatal(id, err)
		}
		c, err := o.clientConfig(m)
		if err != nil || c.GeneratorID != uint32(o.GeneratorID) || c.Manifest.Config().N != 30000 {
			t.Fatal(c, err)
		}
	}
	for _, value := range []string{"1:1:15000", "1:1:15000,2:15000:15000", "1:1:14999,2:15000:15001", "1:1:15000,1:15001:15000", "1:1:15000,2:15001:15025", "1:1:15000,2:15026:14975", "bad", "1:1:4294967296", "1:1:30000,"} {
		if _, _, err := parseOptions("server", []string{"-n", "30000", "-session", "multi", "-result", "server.json", "-generator-ranges", value}, io.Discard); err == nil {
			t.Fatal("invalid manifest accepted", value)
		}
	}
	for _, extra := range [][]string{{"-generator-id", "3", "-count", "15000"}, {"-generator-id", "2", "-first-id", "15001", "-count", "14975"}, {"-generator-id", "2", "-count", "15000"}} {
		args := append(append([]string{}, base...), "-server-result", "server.json")
		if _, _, err := parseOptions("client", append(args, extra...), io.Discard); err == nil {
			t.Fatal("local mismatch accepted", extra)
		}
	}
}
