// arc-bench supervises two independent executables. It does not embed endpoints.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	m "arc.local/multiplexado/internal/measurement"
)

type external struct {
	RealSeconds, UserSeconds, SystemSeconds, CPUPercent float64
	PeakRSSBytes                                        uint64
	PSMaximumCPUPercent                                 float64
	PSSamples                                           int
	PSSampledPeakRSSBytes                               uint64
}
type experiment struct {
	Directory                          string
	N, Rate, RegisterRate, ControlRate int
	PositionWindow                     int
	ServerWriterCount                  int
	ServerEgressSocketCount            int
	GeneratorHighBurst                 int
	GeneratorHighCapacity              int
	OptimizeDataAdmission              bool
	ServerReady                        m.Ready
	Repetition                         int
	Profile                            bool
	Success, Stable                    bool
	Error                              string
	Server, Client                     m.Report
	ServerOS, ClientOS                 external
	TotalSent                          uint64
	GeneratorCount                     int
	Generators                         []generatorResult `json:",omitempty"`
}
type process struct {
	cmd            *exec.Cmd
	done           chan error
	stdout, stderr *os.File
	pid            int
	reaped         bool // supervisor-owned
}

func start(binary, role, dir string, args []string) (*process, error) {
	out, err := os.Create(filepath.Join(dir, role+".stdout.jsonl"))
	if err != nil {
		return nil, err
	}
	log, err := os.Create(filepath.Join(dir, role+".time.txt"))
	if err != nil {
		out.Close()
		return nil, err
	}
	cmd := exec.Command("/usr/bin/time", append([]string{"-l", binary, role}, args...)...)
	cmd.Stdout, cmd.Stderr = out, log
	if err = cmd.Start(); err != nil {
		out.Close()
		log.Close()
		return nil, err
	}
	p := &process{cmd: cmd, done: make(chan error, 1), stdout: out, stderr: log}
	go func() { err := cmd.Wait(); out.Close(); log.Close(); p.done <- err }()
	return p, nil
}
func ready(ctx context.Context, p *process, path string) (m.Ready, error) {
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			var r m.Ready
			if err = json.Unmarshal(data, &r); err != nil {
				return r, err
			}
			p.pid = r.PID
			return r, nil
		}
		select {
		case err := <-p.done:
			p.reaped = true
			return m.Ready{}, fmt.Errorf("process exited before ready: %v", err)
		case <-ctx.Done():
			return m.Ready{}, ctx.Err()
		case <-tick.C:
		}
	}
}
func interrupt(p *process) {
	if p == nil {
		return
	}
	if p.pid > 0 {
		if real, err := os.FindProcess(p.pid); err == nil {
			real.Signal(os.Interrupt)
		}
	} else {
		p.cmd.Process.Signal(os.Interrupt)
	}
}
func cleanup(p *process) {
	if p == nil || p.reaped {
		return
	}
	interrupt(p)
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
		p.reaped = true
	case <-timer.C:
		if p.pid > 0 {
			real, _ := os.FindProcess(p.pid)
			if real != nil {
				real.Kill()
			}
		}
		p.cmd.Process.Kill()
		<-p.done
		p.reaped = true
	}
}
func samplePS(p *process, e *external, f *os.File) {
	if p == nil || p.pid == 0 {
		return
	}
	data, err := exec.Command("ps", "-p", strconv.Itoa(p.pid), "-o", "%cpu=,rss=").Output()
	if err != nil {
		return
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return
	}
	cpu, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return
	}
	rss, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return
	}
	if cpu > e.PSMaximumCPUPercent {
		e.PSMaximumCPUPercent = cpu
	}
	if rss*1024 > e.PSSampledPeakRSSBytes {
		e.PSSampledPeakRSSBytes = rss * 1024
	}
	e.PSSamples++
	fmt.Fprintf(f, "%s,%d,%.3f,%d\n", time.Now().UTC().Format(time.RFC3339Nano), p.pid, cpu, rss*1024)
}

var timesRE = regexp.MustCompile(`(?m)^\s*([0-9.]+)\s+real\s+([0-9.]+)\s+user\s+([0-9.]+)\s+sys\s*$`)
var rssRE = regexp.MustCompile(`(?m)^\s*([0-9]+)\s+maximum resident set size\s*$`)

func parseTime(path string, e *external) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	t := timesRE.FindStringSubmatch(string(data))
	rss := rssRE.FindStringSubmatch(string(data))
	if len(t) != 4 || len(rss) != 2 {
		return fmt.Errorf("missing BSD time CPU/RSS measurements in %s", path)
	}
	e.RealSeconds, _ = strconv.ParseFloat(t[1], 64)
	e.UserSeconds, _ = strconv.ParseFloat(t[2], 64)
	e.SystemSeconds, _ = strconv.ParseFloat(t[3], 64)
	e.PeakRSSBytes, _ = strconv.ParseUint(rss[1], 10, 64)
	if e.RealSeconds > 0 {
		e.CPUPercent = 100 * (e.UserSeconds + e.SystemSeconds) / e.RealSeconds
	}
	return nil
}
func loadReport(path string, r *m.Report) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, r)
}
func one(ctx context.Context, binary, dir string, n, rate, window, writers, sockets, highBurst, highCapacity, rep int, profile, optimizedAdmission bool, timeout time.Duration) experiment {
	x := experiment{Directory: dir, N: n, Rate: rate, RegisterRate: rate * 2, ControlRate: rate * 2, PositionWindow: window, ServerWriterCount: writers, ServerEgressSocketCount: sockets, GeneratorHighBurst: highBurst, GeneratorHighCapacity: highCapacity, Repetition: rep, Profile: profile}
	x.OptimizeDataAdmission = optimizedAdmission
	if err := os.MkdirAll(dir, 0755); err != nil {
		x.Error = err.Error()
		return x
	}
	runID := uint64(time.Now().UnixNano())
	session := fmt.Sprintf("5a-%d", runID)
	common := []string{"-n", strconv.Itoa(n), "-v", "25", "-i", "5", "-run-id", strconv.FormatUint(runID, 10),
		"-session", session, "-workers", "6", "-gomaxprocs", "6", "-register-rate", strconv.Itoa(rate * 2),
		"-position-rate", strconv.Itoa(rate), "-control-rate", strconv.Itoa(rate * 2), "-timeout", timeout.String()}
	serverArgs := append(append([]string{}, common...), "-address", "127.0.0.1:0", "-ready-file", filepath.Join(dir, "server.ready.json"), "-result", filepath.Join(dir, "server.json"))
	// Do not pass the new flag to historical 5A/5B binaries for the default case.
	if writers != 1 {
		serverArgs = append(serverArgs, "-server-writers", strconv.Itoa(writers))
	}
	if sockets != 1 {
		serverArgs = append(serverArgs, "-server-egress-sockets", strconv.Itoa(sockets))
	}
	if profile {
		serverArgs = append(serverArgs, "-cpuprofile", filepath.Join(dir, "server.cpu.pprof"))
		if n >= 30000 {
			serverArgs = append(serverArgs, "-memprofile", filepath.Join(dir, "server.heap.pprof"))
		}
	}
	s, err := start(binary, "server", dir, serverArgs)
	if err != nil {
		x.Error = err.Error()
		return x
	}
	defer cleanup(s)
	sr, err := ready(ctx, s, filepath.Join(dir, "server.ready.json"))
	if err != nil {
		interrupt(s)
		x.Error = err.Error()
		return x
	}
	x.ServerReady = sr
	ingress, sources, err := serverTransport(sr, sockets)
	if err != nil {
		x.Error = err.Error()
		return x
	}
	clientArgs := append(append([]string{}, common...), "-position-window", strconv.Itoa(window), "-server", ingress, "-ready-file", filepath.Join(dir, "client.ready.json"),
		"-server-result", filepath.Join(dir, "server.json"), "-result", filepath.Join(dir, "client.json"))
	if sources != "" {
		clientArgs = append(clientArgs, "-server-egress-endpoints", sources)
	}
	clientArgs = append(clientArgs, generatorPriorityArgs(highBurst)...)
	clientArgs = append(clientArgs, generatorHighCapacityArgs(highCapacity)...)
	clientArgs = append(clientArgs, generatorAdmissionArgs(optimizedAdmission)...)
	if profile {
		clientArgs = append(clientArgs, "-cpuprofile", filepath.Join(dir, "client.cpu.pprof"))
		if n >= 30000 {
			clientArgs = append(clientArgs, "-memprofile", filepath.Join(dir, "client.heap.pprof"))
		}
	}
	c, err := start(binary, "client", dir, clientArgs)
	if err != nil {
		interrupt(s)
		x.Error = err.Error()
		return x
	}
	defer cleanup(c)
	_, err = ready(ctx, c, filepath.Join(dir, "client.ready.json"))
	if err != nil {
		interrupt(s)
		interrupt(c)
		x.Error = err.Error()
		return x
	}
	psfile, err := os.Create(filepath.Join(dir, "ps.csv"))
	if err != nil {
		interrupt(s)
		interrupt(c)
		x.Error = err.Error()
		return x
	}
	defer psfile.Close()
	fmt.Fprintln(psfile, "UTC,PID,CPUPercent,RSSBytes")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	sdone, cdone := false, false
	var sc, cc <-chan error = s.done, c.done
	deadline := ctx.Done()
	var killTimer *time.Timer
	var kill <-chan time.Time
	for !sdone || !cdone {
		select {
		case err := <-sc:
			s.reaped = true
			sc = nil
			sdone = true
			if err != nil {
				x.Error += fmt.Sprintf(" server: %v;", err)
				interrupt(c)
			}
		case err := <-cc:
			c.reaped = true
			cc = nil
			cdone = true
			if err != nil {
				x.Error += fmt.Sprintf(" client: %v;", err)
				interrupt(s)
			}
		case <-ticker.C:
			if !sdone {
				samplePS(s, &x.ServerOS, psfile)
			}
			if !cdone {
				samplePS(c, &x.ClientOS, psfile)
			}
		case <-deadline:
			x.Error += " supervisor timeout;"
			interrupt(s)
			interrupt(c)
			deadline = nil
			killTimer = time.NewTimer(5 * time.Second)
			kill = killTimer.C
		case <-kill:
			for _, p := range []*process{s, c} {
				if p.pid > 0 {
					real, _ := os.FindProcess(p.pid)
					if real != nil {
						real.Kill()
					}
				}
				p.cmd.Process.Kill()
			}
			kill = nil
		}
	}
	if killTimer != nil {
		killTimer.Stop()
	}
	for _, v := range []struct {
		path string
		dst  *m.Report
	}{{"server.json", &x.Server}, {"client.json", &x.Client}} {
		if err := loadReport(filepath.Join(dir, v.path), v.dst); err != nil {
			x.Error += err.Error() + ";"
		}
	}
	if err := parseTime(filepath.Join(dir, "server.time.txt"), &x.ServerOS); err != nil {
		x.Error += err.Error() + ";"
	}
	if err := parseTime(filepath.Join(dir, "client.time.txt"), &x.ClientOS); err != nil {
		x.Error += err.Error() + ";"
	}
	x.Success = x.Error == "" && x.Server.Success && x.Client.Success
	if x.Server.Server != nil && x.Client.Client != nil {
		st, ct := x.Server.Server, x.Client.Client
		if len(st.Writers) > 0 || writers != 1 {
			var written uint64
			for _, w := range st.Writers {
				written += w.Writes
			}
			if len(st.Writers) != writers || written != st.PacketsSent {
				x.Error += "writer count/accounting mismatch;"
				x.Success = false
			}
		}
		if len(st.Egress) > 0 || sockets != 1 {
			var written uint64
			for _, e := range st.Egress {
				written += e.Writes
			}
			if len(st.Egress) != writers || written != st.PacketsSent {
				x.Error += "egress count/accounting mismatch;"
				x.Success = false
			}
		}
		x.TotalSent = st.PacketsSent + ct.PacketsSent
		x.Stable = x.Success && st.QueueDrops == 0 && ct.QueueDrops == 0 && st.ControlTimeouts == 0 && ct.Timeouts == 0 && ct.ControlTimeouts == 0 &&
			st.StartRetries+st.GroupDoneRetries+ct.RegisterRetries+ct.PositionRetries+ct.MetricsRetries == 0 && st.SocketErrors+ct.SocketErrors == 0
	}
	m.WriteJSON(filepath.Join(dir, "experiment.json"), x)
	return x
}
func main() {
	runtime.GOMAXPROCS(1)
	summaryPattern := flag.String("summarize", "", "glob of experiments.json files to summarize")
	binary := flag.String("binary", "", "compiled arc-go executable")
	out := flag.String("out", "", "new result directory (must not exist)")
	n := flag.Int("n", 1000, "logical clients")
	rate := flag.Int("rate", 1000, "POSITION/s, register/control=2x")
	window := flag.Int("position-window", 128, "generator POSITION inflight window; passed to the existing executable flag")
	writers := flag.Int("server-writers", 1, "server TX writers (1/2; experimental 3 requires 3 exclusive egress)")
	sockets := flag.Int("server-egress-sockets", 1, "server TX descriptors (1 shared or 2/3 independent)")
	highBurst := flag.Int("generator-high-burst", 8, "generator-only existing HIGH/CONTROL quota before RETRY/DATA")
	highCapacity := flag.Int("generator-high-capacity", 1024, "generator-only existing HIGH capacity per worker")
	generators := flag.Int("generators", 1, "generator processes (1 baseline or 2 with divided P6/rates/windows)")
	optimizedAdmission := flag.Bool("optimize-data-admission", false, "experimental generator DATA admission; default historical")
	reps := flag.Int("repetitions", 3, "independent sequential repetitions")
	profile := flag.Bool("profile", false, "selected profiling run")
	timeout := flag.Duration("timeout", 10*time.Minute, "deadline per pair")
	flag.Parse()
	if *summaryPattern != "" {
		if err := summarize(*summaryPattern); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *binary == "" || *out == "" || *reps < 1 || *n < 1 || *rate < 1 || *rate > 500000000 || *window < 1 || *writers < 1 || *writers > 3 || *sockets < 1 || *sockets > 3 || *sockets > *writers || (*writers == 3 && *sockets != 3) || *highBurst < 1 || *highCapacity < 1 || (*generators != 1 && *generators != 2) {
		fmt.Fprintln(os.Stderr, "invalid experiment flags")
		os.Exit(2)
	}
	if *generators == 2 {
		if _, err := shardPlan(*n, *rate, *window); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	var err error
	*binary, err = filepath.Abs(*binary)
	if err != nil {
		panic(err)
	}
	*out, err = filepath.Abs(*out)
	if err != nil {
		panic(err)
	}
	if err = os.Mkdir(*out, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "output must be new:", err)
		os.Exit(2)
	}
	all := make([]experiment, 0, *reps)
	success := true
	for rep := 1; rep <= *reps; rep++ {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout+10*time.Second)
		var x experiment
		if *generators == 1 {
			x = one(ctx, *binary, filepath.Join(*out, fmt.Sprintf("r%d", rep)), *n, *rate, *window, *writers, *sockets, *highBurst, *highCapacity, rep, *profile, *optimizedAdmission, *timeout)
			x.GeneratorCount = 1
		} else {
			x = oneSharded(ctx, *binary, filepath.Join(*out, fmt.Sprintf("r%d", rep)), *n, *rate, *window, *writers, *sockets, *highBurst, *highCapacity, rep, *profile, *optimizedAdmission, *timeout)
		}
		cancel()
		all = append(all, x)
		success = success && x.Success
		json.NewEncoder(os.Stdout).Encode(x)
	}
	m.WriteJSON(filepath.Join(*out, "experiments.json"), all)
	if !success {
		os.Exit(1)
	}
}
