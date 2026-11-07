package measurement

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	arc "arc.local/multiplexado"
)

func complete(run arc.RunConfig) (arc.ServerStats, arc.ClientStats) {
	n := uint64(run.N)
	ops := n * uint64(run.I)
	return arc.ServerStats{RegisteredClients: n, StartedClients: n, MetricsAccepted: n, DoneClients: n, CompletedGroups: n / uint64(run.V), SimulationComplete: true},
		arc.ClientStats{LogicalClients: n, RegisteredClients: n, StartedClients: n, DoneClients: n, CompletedOperations: ops, UniqueAcks: ops * uint64(run.V-1), MetricsAccepted: n, RTTSamples: ops}
}
func TestCertificatesRejectPartialAndFailedRuns(t *testing.T) {
	run := arc.RunConfig{RunID: 1, N: 100, V: 25, I: 5}
	s, c := complete(run)
	if ServerCertificate(s, run) != nil || ClientCertificate(c, run, 100) != nil {
		t.Fatal("valid completion rejected")
	}
	for _, mutate := range []func(*arc.ServerStats){func(s *arc.ServerStats) { s.DoneClients-- }, func(s *arc.ServerStats) { s.MetricsAccepted-- },
		func(s *arc.ServerStats) { s.ControlFailures = 1 }, func(s *arc.ServerStats) { s.SimulationComplete = false }} {
		bad := s
		mutate(&bad)
		if ServerCertificate(bad, run) == nil {
			t.Fatal("partial server certified")
		}
	}
	for _, mutate := range []func(*arc.ClientStats){func(c *arc.ClientStats) { c.UniqueAcks-- }, func(c *arc.ClientStats) { c.CompletedOperations-- },
		func(c *arc.ClientStats) { c.DoneClients-- }, func(c *arc.ClientStats) { c.FailedMetrics = 1 }, func(c *arc.ClientStats) { c.InvalidConfirmations = 1 }} {
		bad := c
		mutate(&bad)
		if ClientCertificate(bad, run, 100) == nil {
			t.Fatal("partial/failed client certified")
		}
	}
}
func TestProofRequiresCurrentMatchingSuccessfulServer(t *testing.T) {
	run := arc.RunConfig{RunID: 42, N: 25, V: 25, I: 5}
	s, _ := complete(run)
	start := time.Now()
	r := Report{Schema: 1, Role: "server", Session: "unique", Run: run, Success: true, FinishedUTC: start.Add(time.Second), Server: &s}
	if ValidateProof(r, run, "unique", start) != nil {
		t.Fatal("valid proof rejected")
	}
	for _, mutate := range []func(*Report){func(r *Report) { r.Session = "old" }, func(r *Report) { r.Run.RunID++ },
		func(r *Report) { r.Success = false }, func(r *Report) { r.FinishedUTC = start.Add(-time.Second) }, func(r *Report) { r.Server = nil }} {
		bad := r
		mutate(&bad)
		if ValidateProof(bad, run, "unique", start) == nil {
			t.Fatal("stale/mismatched failure proof accepted")
		}
	}
}
func TestAtomicJSONAndQueueMeasurements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result.json")
	if err := WriteJSON(path, Report{Schema: 1, Success: true}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result Report
	if json.Unmarshal(data, &result) != nil || !result.Success {
		t.Fatal("invalid published JSON")
	}
	s := NewSampler()
	s.Observe(map[string][]arc.QueueStats{"DATA": {{Capacity: 10, Depth: 9, Maximum: 10}, {Capacity: 10, Maximum: 7}}})
	s.last = time.Now().Add(-time.Second)
	s.Observe(map[string][]arc.QueueStats{"DATA": {{Capacity: 10, Maximum: 10}, {Capacity: 10, Maximum: 7}}})
	p := s.Peaks["DATA"]
	if p.Maximum != 10 || p.SumWorkerHighWaters != 17 || p.Above90NS < int64(time.Second) {
		t.Fatal("high waters/exposure", p)
	}
}
