package main

import (
	"io"
	"reflect"
	"testing"
)

func TestExperimentalDataAdmissionFlagOnlyChangesAdmission(t *testing.T) {
	args := []string{"-session", "test", "-server-result", "server.json"}
	o, m, e := parseOptions("client", args, io.Discard)
	if e != nil {
		t.Fatal(e)
	}
	a, e := o.clientConfig(m)
	if e != nil || a.OptimizeDataAdmission {
		t.Fatal("historical default changed")
	}
	o, _, e = parseOptions("client", append(args, "-optimize-data-admission"), io.Discard)
	if e != nil {
		t.Fatal(e)
	}
	b, e := o.clientConfig(m)
	if e != nil || !b.OptimizeDataAdmission {
		t.Fatal("flag not applied")
	}
	b.OptimizeDataAdmission = false
	if !reflect.DeepEqual(a, b) {
		t.Fatal("flag changed another client knob")
	}
	if _, _, e = parseOptions("server", []string{"-session", "test", "-result", "server.json", "-optimize-data-admission"}, io.Discard); e == nil {
		t.Fatal("client flag exposed on server")
	}
}
