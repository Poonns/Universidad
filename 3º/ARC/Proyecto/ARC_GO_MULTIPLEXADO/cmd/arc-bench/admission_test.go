package main

import (
	"reflect"
	"testing"
)

func TestGeneratorAdmissionFlagDefaultsHistorical(t *testing.T) {
	if generatorAdmissionArgs(false) != nil || !reflect.DeepEqual(generatorAdmissionArgs(true), []string{"-optimize-data-admission"}) {
		t.Fatal("bad generator-only experiment flag")
	}
}
