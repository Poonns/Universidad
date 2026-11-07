package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	m "arc.local/multiplexado/internal/measurement"
)

type generatorPlan struct {
	GeneratorID, FirstID, Count, Workers, P                               int
	PositionWindow, RegisterWindow, PositionRate, RegisterRate, RetryRate int
}
type generatorResult struct {
	generatorPlan
	Ready  m.Ready
	Report m.Report
	OS     external
}

// A fixed two-way experiment, not an adaptive partitioner. Budgets remain P6,
// W6, POSITION20000, REGISTER40000, RETRY100 and aggregate windows128.
func shardPlan(n, rate, window int) ([]generatorPlan, error) {
	if n < 50 || n%50 != 0 || rate < 2 || rate%2 != 0 || window < 2 || window%2 != 0 {
		return nil, fmt.Errorf("two generators require N divisible by 2*V25 and evenly divisible rate/window")
	}
	return []generatorPlan{
		{1, 1, n / 2, 3, 3, window / 2, 64, rate / 2, rate, 50},
		{2, n/2 + 1, n / 2, 3, 3, window / 2, 64, rate / 2, rate, 50},
	}, nil
}

func oneSharded(ctx context.Context, binary, dir string, n, rate, window, writers, sockets, highBurst, highCapacity, rep int, profile, optimizedAdmission bool, timeout time.Duration) experiment {
	x := experiment{Directory: dir, N: n, Rate: rate, RegisterRate: rate * 2, ControlRate: rate * 2, PositionWindow: window,
		ServerWriterCount: writers, ServerEgressSocketCount: sockets, GeneratorHighBurst: highBurst, GeneratorHighCapacity: highCapacity,
		Repetition: rep, Profile: profile, GeneratorCount: 2}
	x.OptimizeDataAdmission = optimizedAdmission
	plans, err := shardPlan(n, rate, window)
	if err != nil {
		x.Error = err.Error()
		return x
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		x.Error = err.Error()
		return x
	}
	runID := uint64(time.Now().UnixNano())
	session := fmt.Sprintf("5i-%d", runID)
	ranges := fmt.Sprintf("1:1:%d,2:%d:%d", n/2, n/2+1, n/2)
	common := []string{"-n", strconv.Itoa(n), "-v", "25", "-i", "5", "-run-id", strconv.FormatUint(runID, 10),
		"-session", session, "-generator-ranges", ranges, "-workers", "6", "-gomaxprocs", "6",
		"-register-rate", strconv.Itoa(rate * 2), "-position-rate", strconv.Itoa(rate), "-control-rate", strconv.Itoa(rate * 2), "-timeout", timeout.String()}
	serverArgs := append(append([]string{}, common...), "-address", "127.0.0.1:0", "-server-writers", strconv.Itoa(writers), "-server-egress-sockets", strconv.Itoa(sockets),
		"-ready-file", filepath.Join(dir, "server.ready.json"), "-result", filepath.Join(dir, "server.json"))
	if profile {
		serverArgs = append(serverArgs, "-cpuprofile", filepath.Join(dir, "server.cpu.pprof"))
	}
	s, err := start(binary, "server", dir, serverArgs)
	if err != nil {
		x.Error = err.Error()
		return x
	}
	defer cleanup(s)
	x.ServerReady, err = ready(ctx, s, filepath.Join(dir, "server.ready.json"))
	if err != nil {
		x.Error = err.Error()
		return x
	}
	ingress, sources, err := serverTransport(x.ServerReady, sockets)
	if err != nil {
		x.Error = err.Error()
		return x
	}
	processes := []*process{s}
	dirs := []string{dir}
	x.Generators = make([]generatorResult, len(plans))
	for i, p := range plans {
		local := filepath.Join(dir, fmt.Sprintf("generator%d", p.GeneratorID))
		if err = os.Mkdir(local, 0755); err != nil {
			x.Error = err.Error()
			return x
		}
		args := append(append([]string{}, common...), "-generator-id", strconv.Itoa(p.GeneratorID), "-first-id", strconv.Itoa(p.FirstID), "-count", strconv.Itoa(p.Count),
			"-workers", strconv.Itoa(p.Workers), "-gomaxprocs", strconv.Itoa(p.P), "-position-window", strconv.Itoa(p.PositionWindow),
			"-register-window", strconv.Itoa(p.RegisterWindow), "-position-rate", strconv.Itoa(p.PositionRate), "-register-rate", strconv.Itoa(p.RegisterRate), "-retry-rate", strconv.Itoa(p.RetryRate),
			"-server", ingress, "-server-egress-endpoints", sources, "-ready-file", filepath.Join(local, "client.ready.json"),
			"-server-result", filepath.Join(dir, "server.json"), "-result", filepath.Join(local, "client.json"))
		args = append(args, generatorPriorityArgs(highBurst)...)
		args = append(args, generatorHighCapacityArgs(highCapacity)...)
		args = append(args, generatorAdmissionArgs(optimizedAdmission)...)
		if profile {
			args = append(args, "-cpuprofile", filepath.Join(local, "client.cpu.pprof"))
		}
		c, startErr := start(binary, "client", local, args)
		if startErr != nil {
			x.Error = startErr.Error()
			return x
		}
		defer cleanup(c)
		processes, dirs = append(processes, c), append(dirs, local)
		x.Generators[i].generatorPlan = p
		x.Generators[i].Ready, err = ready(ctx, c, filepath.Join(local, "client.ready.json"))
		if err != nil {
			x.Error = err.Error()
			return x
		}
	}
	ps, err := os.Create(filepath.Join(dir, "ps.csv"))
	if err != nil {
		x.Error = err.Error()
		return x
	}
	defer ps.Close()
	fmt.Fprintln(ps, "UTC,PID,CPUPercent,RSSBytes")
	if err = waitProcesses(ctx, processes, &x, ps); err != nil {
		x.Error += err.Error() + ";"
	}
	if err = loadReport(filepath.Join(dir, "server.json"), &x.Server); err != nil {
		x.Error += err.Error() + ";"
	}
	if err = parseTime(filepath.Join(dir, "server.time.txt"), &x.ServerOS); err != nil {
		x.Error += err.Error() + ";"
	}
	for i := range x.Generators {
		g := &x.Generators[i]
		if err = loadReport(filepath.Join(dirs[i+1], "client.json"), &g.Report); err != nil {
			x.Error += err.Error() + ";"
		}
		if err = parseTime(filepath.Join(dirs[i+1], "client.time.txt"), &g.OS); err != nil {
			x.Error += err.Error() + ";"
		}
	}
	certifySharded(&x)
	if err = m.WriteJSON(filepath.Join(dir, "experiment.json"), x); err != nil {
		x.Error += err.Error()
		x.Success, x.Stable = false, false
	}
	return x
}

func waitProcesses(ctx context.Context, ps []*process, x *experiment, log *os.File) error {
	type completion struct {
		index int
		err   error
	}
	done := make(chan completion, len(ps))
	for i, p := range ps {
		go func(i int, p *process) { done <- completion{i, <-p.done} }(i, p)
	}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	var kill <-chan time.Time
	deadline := ctx.Done()
	var failure error
	abort := func(err error) {
		if failure == nil {
			failure = err
		}
		for _, p := range ps {
			if !p.reaped {
				interrupt(p)
			}
		}
		if timer == nil {
			timer = time.NewTimer(5 * time.Second)
			kill = timer.C
		}
	}
	for left := len(ps); left > 0; {
		select {
		case c := <-done:
			ps[c.index].reaped = true
			left--
			if c.err != nil {
				abort(fmt.Errorf("process%d: %w", c.index, c.err))
			}
		case <-tick.C:
			for i, p := range ps {
				if p.reaped {
					continue
				}
				e := &x.ServerOS
				if i > 0 {
					e = &x.Generators[i-1].OS
				}
				samplePS(p, e, log)
			}
		case <-deadline:
			abort(fmt.Errorf("supervisor timeout: %w", ctx.Err()))
			deadline = nil
		case <-kill:
			for _, p := range ps {
				if p.reaped {
					continue
				}
				if real, err := os.FindProcess(p.pid); err == nil {
					real.Kill()
				}
				p.cmd.Process.Kill()
			}
			kill = nil
		}
	}
	return failure
}

func certifySharded(x *experiment) {
	x.Success, x.Stable = false, false
	if x.Error != "" || !x.Server.Success || x.Server.Server == nil || len(x.Generators) != 2 {
		return
	}
	st := x.Server.Server
	if err := m.ServerCertificate(*st, x.Server.Run); err != nil {
		x.Error += err.Error()
		return
	}
	stable := st.QueueDrops+st.ControlTimeouts+st.StartRetries+st.GroupDoneRetries+st.SocketErrors+st.ControlFailures+st.ControlRetryExhausted == 0
	var ops, acks, clients uint64
	x.TotalSent = st.PacketsSent
	for _, g := range x.Generators {
		r, ct := g.Report, g.Report.Client
		if !r.Success || ct == nil || r.Run != x.Server.Run || r.Session != x.Server.Session {
			x.Error += "generator identity/report mismatch;"
			return
		}
		if err := m.ClientCertificate(*ct, r.Run, uint32(g.Count)); err != nil {
			x.Error += err.Error()
			return
		}
		stable = stable && ct.QueueDrops+ct.RegisterRetries+ct.PositionRetries+ct.MetricsRetries+ct.Timeouts+ct.ControlTimeouts+ct.SocketErrors+
			ct.FailedRegistrations+ct.FailedOperations+ct.FailedMetrics+ct.AbortedOperations+ct.RetryExhausted+ct.ControlRetryExhausted == 0
		clients += ct.RegisteredClients
		ops += ct.CompletedOperations
		acks += ct.UniqueAcks
		x.TotalSent += ct.PacketsSent
	}
	if clients != uint64(x.N) || ops != uint64(x.N)*5 || acks != uint64(x.N)*5*24 {
		x.Error += "global sum mismatch;"
		return
	}
	x.Success, x.Stable = true, stable
}
