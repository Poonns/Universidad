package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	arc "arc.local/multiplexado"
	m "arc.local/multiplexado/internal/measurement"
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 || (args[0] != "server" && args[0] != "client") {
		fmt.Fprintln(os.Stderr, "usage: arc-go server|client [flags]")
		return 2
	}
	role := args[0]
	o, manifest, err := parseOptions(role, args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	runtime.GOMAXPROCS(o.GOMAXPROCS)
	begin := time.Now()
	config, _ := json.Marshal(o)
	r := m.Report{Schema: 1, Event: "final", Role: role, Session: o.Session, PID: os.Getpid(), Run: manifest.Config(), Config: config,
		StartedUTC: begin.UTC(), StartRuntime: m.RuntimeSnapshot()}
	parent, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(parent, o.Timeout)
	defer cancel()
	stopProfile, err := cpuProfile(o.CPUProfile)
	if err == nil {
		if role == "server" {
			err = serve(ctx, o, manifest, &r, begin)
		} else {
			err = generate(ctx, o, manifest, &r, begin)
		}
		stopProfile()
	}
	r.FinishedUTC = time.Now().UTC()
	r.ProcessDurationNS = time.Since(begin).Nanoseconds()
	r.EndRuntime = m.RuntimeSnapshot()
	r.Success = err == nil
	if err != nil {
		r.Error = err.Error()
	}
	seconds := float64(r.ProcessDurationNS) / 1e9
	if r.Server != nil {
		s := r.Server
		r.PacketsReceivedPerSecond = float64(s.PacketsReceived) / seconds
		r.PacketsSentPerSecond = float64(s.PacketsSent) / seconds
		r.RTTMeanNS = s.MeanRTTNS
		if d := s.SimulationDuration.Seconds(); d > 0 {
			r.AcademicDataReceivedPerSecond = float64(s.PositionReceived+s.RetryPositionReceived+s.AckReceived) / d
			r.AcademicDataSentPerSecond = float64(s.PositionForwarded+s.RetryPositionForwarded+s.AckForwarded) / d
		}
	}
	if r.Client != nil {
		c := r.Client
		writer := m.GeneratorWriterSnapshot(*c)
		r.GeneratorWriter = &writer
		r.PacketsReceivedPerSecond = float64(c.PacketsReceived) / seconds
		r.PacketsSentPerSecond = float64(c.PacketsSent) / seconds
		if c.RTTSamples > 0 {
			r.RTTMeanNS = c.RTTSum / c.RTTSamples
		}
	}
	if writeErr := m.WriteJSON(o.Result, r); writeErr != nil {
		r.Success = false
		r.Error = fmt.Sprintf("%s result write: %v", r.Error, writeErr)
	}
	json.NewEncoder(os.Stdout).Encode(r)
	fmt.Fprintf(os.Stderr, "%s PID=%d success=%t elapsed=%s RTT=%s error=%s\n", role, r.PID, r.Success, time.Duration(r.ProcessDurationNS), time.Duration(r.RTTMeanNS), r.Error)
	if !r.Success {
		return 1
	}
	return 0
}

func ready(o options, role, address string, run arc.RunConfig, egress ...netip.AddrPort) error {
	v := m.Ready{Event: "ready", Role: role, Session: o.Session, Address: address, PID: os.Getpid(), Run: run}
	if role == "server" {
		v.IngressEndpoint = address
		for _, endpoint := range egress {
			v.ServerEgressEndpoints = append(v.ServerEgressEndpoints, endpoint.String())
		}
	}
	if err := m.WriteJSON(o.ReadyFile, v); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s PID=%d UDP=%s N=%d V=%d I=%d W=%d P=%d\n", role, v.PID, address, run.N, run.V, run.I, o.Workers, runtime.GOMAXPROCS(0))
	return json.NewEncoder(os.Stdout).Encode(v)
}

func serve(ctx context.Context, o options, manifest *arc.Manifest, r *m.Report, begin time.Time) error {
	s, err := arc.ListenServer(ctx, o.serverConfig(manifest))
	if err != nil {
		return err
	}
	defer s.Close()
	sample := m.NewSampler()
	observe := func() arc.ServerStats { st := s.Stats(); sample.Observe(m.ServerQueues(st)); return st }
	if err = ready(o, "server", s.Addr().String(), manifest.Config(), s.EgressEndpoints()...); err != nil {
		return err
	}
	finished := make(chan error, 1)
	go func() { finished <- s.WaitSimulationComplete(ctx) }()
	tick := time.NewTicker(o.SampleInterval)
	defer tick.Stop()
	waiting := true
	for waiting {
		select {
		case err = <-finished:
			waiting = false
		case <-tick.C:
			observe()
		case <-ctx.Done():
			err = ctx.Err()
			waiting = false
		}
	}
	r.LogicalDoneNS = time.Since(begin).Nanoseconds()
	if err == nil {
		// Real pending work, not a grace sleep. The generator is still serving UDP.
		for !m.Empty(m.ServerQueues(observe())) {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-tick.C:
			}
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		err = heapProfile(o.MemProfile)
	}
	closeErr := s.Close()
	st := observe()
	r.Server = &st
	r.QueuePeaks = sample.Peaks
	r.QueueSamples = sample.Samples
	runtime.KeepAlive(s)
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return m.ServerCertificate(st, manifest.Config())
}

func generate(ctx context.Context, o options, manifest *arc.Manifest, r *m.Report, begin time.Time) error {
	cfg, err := o.clientConfig(manifest)
	if err != nil {
		return err
	}
	g, err := arc.ListenClients(ctx, cfg)
	if err != nil {
		return err
	}
	defer g.Close()
	sample := m.NewSampler()
	observe := func() arc.ClientStats { st := g.Stats(); sample.Observe(m.ClientQueues(st)); return st }
	if err = ready(o, "client", g.Addr().String(), manifest.Config()); err != nil {
		return err
	}
	finished := make(chan error, 1)
	go func() { finished <- g.WaitAllDone(ctx) }()
	tick := time.NewTicker(o.SampleInterval)
	defer tick.Stop()
	waiting := true
	for waiting {
		select {
		case err = <-finished:
			waiting = false
		case <-tick.C:
			observe()
		case <-ctx.Done():
			err = ctx.Err()
			waiting = false
		}
	}
	r.LogicalDoneNS = time.Since(begin).Nanoseconds()
	proofBegin := time.Now()
	if err == nil {
		// Out-of-band supervisor proof only: no new UDP message or finite "last ACK" guess.
		for {
			data, readErr := os.ReadFile(o.ServerResult)
			if readErr == nil {
				var proof m.Report
				if err = json.Unmarshal(data, &proof); err != nil {
					break
				}
				err = m.ValidateProof(proof, manifest.Config(), o.Session, begin)
				break
			}
			if !errors.Is(readErr, os.ErrNotExist) {
				err = readErr
				break
			}
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-tick.C:
				observe()
			}
			if err != nil {
				break
			}
		}
	}
	r.CoordinationWaitNS = time.Since(proofBegin).Nanoseconds()
	if err == nil {
		for !m.Empty(m.ClientQueues(observe())) {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-tick.C:
			}
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		err = heapProfile(o.MemProfile)
	}
	closeErr := g.Close()
	st := observe()
	r.Client = &st
	r.QueuePeaks = sample.Peaks
	r.QueueSamples = sample.Samples
	runtime.KeepAlive(g)
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return m.ClientCertificate(st, manifest.Config(), uint32(o.Count))
}

func cpuProfile(path string) (func(), error) {
	if path == "" {
		return func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err = pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	return func() { pprof.StopCPUProfile(); f.Close() }, nil
}
func heapProfile(path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	runtime.GC()
	err = pprof.WriteHeapProfile(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
