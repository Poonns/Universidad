// Package measurement records experiments outside the protocol hot path.
package measurement

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	arc "arc.local/multiplexado"
)

type Runtime struct {
	GoVersion, GOOS, GOARCH         string
	GOMAXPROCS, NumCPU, Goroutines  int
	HeapAlloc, HeapSys, HeapObjects uint64
	NumGC                           uint32
	PauseTotalNs                    uint64
}

func RuntimeSnapshot() Runtime {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Runtime{runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0),
		runtime.NumCPU(), runtime.NumGoroutine(), m.HeapAlloc, m.HeapSys, m.HeapObjects, m.NumGC, m.PauseTotalNs}
}

type QueuePeak struct {
	Capacity, Maximum, SumWorkerHighWaters int
	FullAttempts, Drops                    uint64
	Above50NS, Above75NS, Above90NS        int64
}

// Sample records the worst worker per class. Exposure is an approximation using
// the previous sampled occupancy; exact high waters come from the library.
type Sampler struct {
	Peaks    map[string]QueuePeak
	previous map[string]int
	last     time.Time
	Samples  uint64
}

func NewSampler() *Sampler {
	return &Sampler{Peaks: make(map[string]QueuePeak), previous: make(map[string]int), last: time.Now()}
}
func (s *Sampler) Observe(queues map[string][]arc.QueueStats) {
	now := time.Now()
	dt := now.Sub(s.last).Nanoseconds()
	for name, workers := range queues {
		p := s.Peaks[name]
		if s.previous[name] >= 50 {
			p.Above50NS += dt
		}
		if s.previous[name] >= 75 {
			p.Above75NS += dt
		}
		if s.previous[name] >= 90 {
			p.Above90NS += dt
		}
		sum, full, drops, exposure := 0, uint64(0), uint64(0), 0
		for _, q := range workers {
			if q.Capacity > p.Capacity {
				p.Capacity = q.Capacity
			}
			if q.Maximum > p.Maximum {
				p.Maximum = q.Maximum
			}
			sum += q.Maximum
			full += q.FullAttempts
			drops += q.Drops
			if q.Capacity > 0 && 100*q.Depth/q.Capacity > exposure {
				exposure = 100 * q.Depth / q.Capacity
			}
		}
		p.SumWorkerHighWaters, p.FullAttempts, p.Drops = sum, full, drops
		s.Peaks[name], s.previous[name] = p, exposure
	}
	s.last = now
	s.Samples++
}

func ServerQueues(st arc.ServerStats) map[string][]arc.QueueStats {
	out := make(map[string][]arc.QueueStats)
	for _, q := range st.Queues {
		out["INPUT"] = append(out["INPUT"], q.Input)
		out["HIGH"] = append(out["HIGH"], q.High)
		out["CONTROL"] = append(out["CONTROL"], q.Control)
		out["DATA"] = append(out["DATA"], q.Data)
		out["CONFIRMATIONS"] = append(out["CONFIRMATIONS"], q.Confirmations)
		out["PENDING_CLIENTS"] = append(out["PENDING_CLIENTS"], q.PendingClients)
		out["PENDING_GROUPS"] = append(out["PENDING_GROUPS"], q.PendingGroups)
	}
	return out
}
func ClientQueues(st arc.ClientStats) map[string][]arc.QueueStats {
	out := make(map[string][]arc.QueueStats)
	for _, q := range st.Queues {
		out["INPUT"] = append(out["INPUT"], q.Input)
		out["HIGH"] = append(out["HIGH"], q.High)
		out["CONTROL"] = append(out["CONTROL"], q.Control)
		out["DATA"] = append(out["DATA"], q.Data)
		out["RETRY"] = append(out["RETRY"], q.Retry)
		out["CONFIRMATIONS"] = append(out["CONFIRMATIONS"], q.Confirmations)
		out["COMMANDS"] = append(out["COMMANDS"], q.Commands)
		out["PENDING_CLIENTS"] = append(out["PENDING_CLIENTS"], q.PendingClients)
	}
	return out
}
func Empty(queues map[string][]arc.QueueStats) bool {
	for _, workers := range queues {
		for _, q := range workers {
			if q.Depth != 0 {
				return false
			}
		}
	}
	return true
}

type Ready struct {
	Event, Role, Session, Address string
	IngressEndpoint               string   `json:",omitempty"`
	ServerEgressEndpoints         []string `json:",omitempty"`
	PID                           int
	Run                           arc.RunConfig
}
type Report struct {
	Schema                                                   int
	Event, Role, Session                                     string
	PID                                                      int
	Run                                                      arc.RunConfig
	Config                                                   json.RawMessage
	Success                                                  bool
	Error                                                    string
	StartedUTC, FinishedUTC                                  time.Time
	ProcessDurationNS, LogicalDoneNS, CoordinationWaitNS     int64
	StartRuntime, EndRuntime                                 Runtime
	QueuePeaks                                               map[string]QueuePeak
	QueueSamples                                             uint64
	Server                                                   *arc.ServerStats      `json:",omitempty"`
	Client                                                   *arc.ClientStats      `json:",omitempty"`
	GeneratorWriter                                          *GeneratorWriterStats `json:",omitempty"`
	RTTMeanNS                                                uint64
	PacketsReceivedPerSecond, PacketsSentPerSecond           float64
	AcademicDataReceivedPerSecond, AcademicDataSentPerSecond float64
}

func ServerCertificate(st arc.ServerStats, run arc.RunConfig) error {
	n := uint64(run.N)
	if st.RegisteredClients != n || st.StartedClients != n || st.MetricsAccepted != n ||
		st.DoneClients != n || st.CompletedGroups != n/uint64(run.V) || !st.SimulationComplete ||
		st.ControlFailures != 0 || st.ControlRetryExhausted != 0 {
		return fmt.Errorf("incomplete server academic barriers")
	}
	if st.InvalidPackets+st.SocketErrors+st.SendDrops+st.Conflicts+st.EndpointRejects != 0 {
		return fmt.Errorf("server validation/socket errors")
	}
	return nil
}
func ClientCertificate(st arc.ClientStats, run arc.RunConfig, count uint32) error {
	n := uint64(count)
	if st.LogicalClients != n || st.RegisteredClients != n || st.StartedClients != n || st.DoneClients != n ||
		st.CompletedOperations != n*uint64(run.I) || st.UniqueAcks != n*uint64(run.I)*uint64(run.V-1) || st.MetricsAccepted != n ||
		st.RTTSamples != n*uint64(run.I) {
		return fmt.Errorf("incomplete generator academic counters")
	}
	if st.FailedRegistrations+st.FailedOperations+st.FailedMetrics+st.RetryExhausted+st.InvalidPackets+
		st.SocketErrors+st.SendDrops+st.InvalidConfirmations+st.RejectedOperations+st.AbortedOperations != 0 {
		return fmt.Errorf("generator terminal/validation/socket errors")
	}
	return nil
}
func ValidateProof(proof Report, run arc.RunConfig, session string, started time.Time) error {
	if proof.Schema != 1 || proof.Role != "server" || proof.Session != session || proof.Run != run ||
		proof.FinishedUTC.Before(started) {
		return fmt.Errorf("stale or mismatched server result")
	}
	if !proof.Success || proof.Server == nil {
		return fmt.Errorf("server failed: %s", proof.Error)
	}
	return ServerCertificate(*proof.Server, run)
}
func WriteJSON(path string, value any) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".arc-result-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
