package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	arc "arc.local/multiplexado"
)

type options struct {
	N, V, I, GeneratorID, FirstID, Count                                                                             uint
	RunID                                                                                                            uint64
	Address, Server, Session, Result, ServerResult, ReadyFile, CPUProfile, MemProfile                                string
	GeneratorRanges                                                                                                  string
	Workers, GOMAXPROCS                                                                                              int
	ServerWriterCount                                                                                                int
	ServerEgressSocketCount                                                                                          int
	ServerEgressEndpoints                                                                                            string
	RegisterRate, RegisterWindow, PositionRate, PositionBurst, PositionWindow, RetryRate, RetryBurst                 int
	ControlRate, ControlBurst                                                                                        int
	InputCapacity, HighCapacity, ControlCapacity, DataCapacity, RetryCapacity, ConfirmationCapacity, CommandCapacity int
	WorkBudget, HighBurst, AckBurst, ReadBuffer, WriteBuffer                                                         int
	InitialRTO, MaxRTO, WriteTimeout, Timeout, SampleInterval                                                        time.Duration
	MaxSends                                                                                                         uint
	JitterPercent                                                                                                    int
	Seed                                                                                                             uint64
	Academic                                                                                                         bool
	OptimizeDataAdmission                                                                                            bool
}

func parseOptions(role string, args []string, output io.Writer) (options, *arc.Manifest, error) {
	var o options
	fs := flag.NewFlagSet(role, flag.ContinueOnError)
	fs.SetOutput(output)
	fs.UintVar(&o.N, "n", 1000, "total logical clients")
	fs.UintVar(&o.V, "v", 25, "members per group")
	fs.UintVar(&o.I, "i", 5, "iterations per client")
	fs.Uint64Var(&o.RunID, "run-id", 1, "nonzero run ID")
	fs.UintVar(&o.GeneratorID, "generator-id", 1, "manifest generator ID")
	fs.UintVar(&o.FirstID, "first-id", 1, "first logical client ID")
	fs.UintVar(&o.Count, "count", 0, "logical clients in generator (0 means N); baseline uses one full range")
	fs.StringVar(&o.GeneratorRanges, "generator-ranges", "", "complete manifest: generatorID:firstID:count,...; shared by server and every client")
	fs.StringVar(&o.Address, "address", "127.0.0.1:0", "local UDP listen address; server defaults to port 9000")
	fs.StringVar(&o.Server, "server", "127.0.0.1:9000", "server UDP host:port")
	fs.StringVar(&o.Session, "session", "", "unique supervisor session (required)")
	fs.StringVar(&o.Result, "result", "", "atomic final JSON file; required for server")
	fs.StringVar(&o.ServerResult, "server-result", "", "server result proof file; required for client")
	fs.StringVar(&o.ReadyFile, "ready-file", "", "optional startup JSON file")
	fs.StringVar(&o.CPUProfile, "cpuprofile", "", "optional CPU pprof file")
	fs.StringVar(&o.MemProfile, "memprofile", "", "optional live heap pprof file after forced GC, before close")
	fs.IntVar(&o.Workers, "workers", 6, "group owner workers")
	if role == "server" {
		fs.IntVar(&o.ServerWriterCount, "server-writers", 1, "server TX writers (1/2; experimental 3 requires 3 exclusive egress)")
		fs.IntVar(&o.ServerEgressSocketCount, "server-egress-sockets", 1, "server TX descriptors: shared ingress (1) or independent TX-only sockets (2/3)")
	} else {
		fs.StringVar(&o.ServerEgressEndpoints, "server-egress-endpoints", "", "comma-separated exact authorized server RX sources; default canonical ingress only")
		fs.BoolVar(&o.OptimizeDataAdmission, "optimize-data-admission", false, "experimental DATA admission; false preserves historical flushPending")
	}
	fs.IntVar(&o.GOMAXPROCS, "gomaxprocs", 6, "process CPU scheduler budget")
	fs.IntVar(&o.RegisterRate, "register-rate", 1000, "initial REGISTER/s, per generator")
	fs.IntVar(&o.RegisterWindow, "register-window", 128, "registration inflight limit")
	fs.IntVar(&o.PositionRate, "position-rate", 1000, "new POSITION/s, per generator")
	fs.IntVar(&o.PositionBurst, "position-burst", 1, "POSITION token burst")
	fs.IntVar(&o.PositionWindow, "position-window", 128, "POSITION inflight limit")
	fs.IntVar(&o.RetryRate, "retry-rate", 100, "retry requests/s")
	fs.IntVar(&o.RetryBurst, "retry-burst", 1, "retry token burst")
	fs.IntVar(&o.ControlRate, "control-rate", 1000, "server START/GROUP_DONE/s, aggregated")
	fs.IntVar(&o.ControlBurst, "control-burst", 1, "server control token burst")
	fs.IntVar(&o.InputCapacity, "input-capacity", 4096, "input slots per worker")
	fs.IntVar(&o.HighCapacity, "high-capacity", 1024, "HIGH slots per worker")
	controlDefault := 128
	if role == "server" {
		controlDefault = 256
	}
	fs.IntVar(&o.ControlCapacity, "control-capacity", controlDefault, "CONTROL slots per worker")
	fs.IntVar(&o.DataCapacity, "data-capacity", 2048, "DATA slots per worker")
	fs.IntVar(&o.RetryCapacity, "retry-capacity", 512, "client RETRY slots per worker")
	fs.IntVar(&o.ConfirmationCapacity, "confirmation-capacity", 256, "writer confirmation slots per worker")
	fs.IntVar(&o.CommandCapacity, "command-capacity", 256, "client command slots per worker")
	fs.IntVar(&o.WorkBudget, "work-budget", 64, "owner work budget")
	fs.IntVar(&o.HighBurst, "high-burst", 8, "writer HIGH quota")
	fs.IntVar(&o.AckBurst, "ack-burst", 4, "client ACK quota before control")
	fs.IntVar(&o.ReadBuffer, "read-buffer", 4*1024*1024, "requested socket read bytes")
	fs.IntVar(&o.WriteBuffer, "write-buffer", 4*1024*1024, "requested socket write bytes")
	fs.DurationVar(&o.InitialRTO, "initial-rto", time.Second, "initial RTO; server control or client ARQ")
	fs.DurationVar(&o.MaxRTO, "max-rto", 5*time.Second, "maximum RTO")
	fs.UintVar(&o.MaxSends, "max-sends", 12, "maximum sends including initial")
	fs.IntVar(&o.JitterPercent, "jitter-percent", 20, "deterministic RTO jitter percentage")
	fs.Uint64Var(&o.Seed, "seed", 42, "deterministic jitter/coordinates seed")
	fs.DurationVar(&o.WriteTimeout, "write-timeout", time.Second, "socket write deadline")
	fs.DurationVar(&o.Timeout, "timeout", 10*time.Minute, "whole-run deadline including registration and proof")
	fs.DurationVar(&o.SampleInterval, "sample-interval", 250*time.Millisecond, "queue observation interval")
	fs.BoolVar(&o.Academic, "academic", true, "full approved academic flow (must be true)")
	if err := fs.Parse(args); err != nil {
		return o, nil, err
	}
	if fs.NArg() != 0 {
		return o, nil, fmt.Errorf("unexpected positional arguments")
	}
	if o.Count == 0 {
		o.Count = o.N
	}
	if o.N > math.MaxUint32 || o.V > math.MaxUint32 || o.I > math.MaxUint32 || o.Count > math.MaxUint32 ||
		o.FirstID > math.MaxUint32 || o.GeneratorID > math.MaxUint32 || o.MaxSends == 0 || o.MaxSends > math.MaxUint16 ||
		o.Session == "" || !o.Academic || o.Timeout <= 0 || o.SampleInterval <= 0 || o.GOMAXPROCS < 1 ||
		(role == "server" && (o.Result == "" || o.ServerWriterCount < 1 || o.ServerWriterCount > 3 || o.ServerWriterCount > o.Workers ||
			o.ServerEgressSocketCount < 1 || o.ServerEgressSocketCount > 3 || o.ServerEgressSocketCount > o.ServerWriterCount ||
			(o.ServerWriterCount == 3 && o.ServerEgressSocketCount != 3))) || (role == "client" && o.ServerResult == "") {
		return o, nil, fmt.Errorf("invalid CLI values: require Academic=true, unique -session and server -result / client -server-result")
	}
	if role == "server" {
		visited := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "address" {
				visited = true
			}
		})
		if !visited {
			o.Address = "127.0.0.1:9000"
		}
	}
	run := arc.RunConfig{RunID: o.RunID, N: uint32(o.N), V: uint32(o.V), I: uint32(o.I)}
	gens := []arc.GeneratorConfig{{GeneratorID: uint32(o.GeneratorID), FirstID: uint32(o.FirstID), Count: uint32(o.Count), Run: run}}
	if o.GeneratorRanges != "" {
		var err error
		gens, err = parseGeneratorRanges(o.GeneratorRanges, run)
		if err != nil {
			return o, nil, err
		}
		if role == "client" {
			found := false
			for _, g := range gens {
				if g.GeneratorID == uint32(o.GeneratorID) {
					found = g.FirstID == uint32(o.FirstID) && g.Count == uint32(o.Count)
				}
			}
			if !found {
				return o, nil, fmt.Errorf("local generator ID/range must match complete manifest")
			}
		}
	}
	m, err := arc.NewManifest(run, gens)
	return o, m, err
}

// Only startup wiring: NewManifest retains authority over coverage, alignment
// and identities; no per-datagram representation or validation is changed.
func parseGeneratorRanges(value string, run arc.RunConfig) ([]arc.GeneratorConfig, error) {
	var gens []arc.GeneratorConfig
	for _, item := range strings.Split(value, ",") {
		fields := strings.Split(strings.TrimSpace(item), ":")
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid generator range %q", item)
		}
		var nums [3]uint32
		for i, field := range fields {
			n, err := strconv.ParseUint(field, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid generator range %q: %w", item, err)
			}
			nums[i] = uint32(n)
		}
		gens = append(gens, arc.GeneratorConfig{GeneratorID: nums[0], FirstID: nums[1], Count: nums[2], Run: run})
	}
	return gens, nil
}
func (o options) serverConfig(m *arc.Manifest) arc.ServerConfig {
	c := arc.DefaultServerConfig(m)
	c.Address, c.Workers, c.GOMAXPROCS = o.Address, o.Workers, o.GOMAXPROCS
	c.ServerWriterCount = o.ServerWriterCount
	c.ServerEgressSocketCount = o.ServerEgressSocketCount
	c.InputCapacity, c.HighCapacity, c.ControlCapacity, c.DataCapacity = o.InputCapacity, o.HighCapacity, o.ControlCapacity, o.DataCapacity
	c.ConfirmationCapacity, c.WorkBudget, c.HighBurst = o.ConfirmationCapacity, o.WorkBudget, o.HighBurst
	c.ControlRate, c.ControlBurst = o.ControlRate, o.ControlBurst
	c.InitialControlRTO, c.MaxControlRTO, c.MaxControlSends = o.InitialRTO, o.MaxRTO, uint16(o.MaxSends)
	c.ControlJitterPercent, c.ControlJitterSeed = o.JitterPercent, o.Seed
	c.ReadBuffer, c.WriteBuffer, c.WriteTimeout = o.ReadBuffer, o.WriteBuffer, o.WriteTimeout
	return c
}
func (o options) clientConfig(m *arc.Manifest) (arc.ClientConfig, error) {
	addr, err := net.ResolveUDPAddr("udp", o.Server)
	if err != nil {
		return arc.ClientConfig{}, err
	}
	endpoint := netip.AddrPortFrom(addr.AddrPort().Addr(), addr.AddrPort().Port())
	c := arc.DefaultClientConfig(m, uint32(o.GeneratorID), endpoint)
	c.OptimizeDataAdmission = o.OptimizeDataAdmission
	if o.ServerEgressEndpoints != "" {
		parts := strings.Split(o.ServerEgressEndpoints, ",")
		if len(parts) > 3 {
			return arc.ClientConfig{}, fmt.Errorf("expected one to three exact authorized server egress endpoints")
		}
		for _, s := range parts {
			e, err := netip.ParseAddrPort(strings.TrimSpace(s))
			if err != nil || e.Port() == 0 {
				return arc.ClientConfig{}, fmt.Errorf("invalid authorized egress endpoint %q", s)
			}
			e = netip.AddrPortFrom(e.Addr().Unmap(), e.Port())
			if e.Addr().IsUnspecified() || e.Addr().IsMulticast() {
				return arc.ClientConfig{}, fmt.Errorf("invalid authorized source %q", s)
			}
			for _, previous := range c.ServerEgressEndpoints {
				if e == previous {
					return arc.ClientConfig{}, fmt.Errorf("duplicate authorized egress endpoint")
				}
			}
			c.ServerEgressEndpoints = append(c.ServerEgressEndpoints, e)
		}
	}
	c.Address, c.Workers, c.GOMAXPROCS = o.Address, o.Workers, o.GOMAXPROCS
	c.InputCapacity, c.HighCapacity, c.ControlCapacity, c.DataCapacity = o.InputCapacity, o.HighCapacity, o.ControlCapacity, o.DataCapacity
	c.ConfirmationCapacity, c.CommandCapacity, c.RetryCapacity = o.ConfirmationCapacity, o.CommandCapacity, o.RetryCapacity
	c.RegisterRate, c.RegisterWindow = o.RegisterRate, o.RegisterWindow
	c.PositionRate, c.PositionBurst, c.PositionWindow = o.PositionRate, o.PositionBurst, o.PositionWindow
	c.RetryRate, c.RetryBurst = o.RetryRate, o.RetryBurst
	c.InitialRTO, c.MaxRTO, c.MaxSends = o.InitialRTO, o.MaxRTO, uint16(o.MaxSends)
	c.JitterPercent, c.JitterSeed = o.JitterPercent, o.Seed
	c.WorkBudget, c.HighBurst, c.AckBurst = o.WorkBudget, o.HighBurst, o.AckBurst
	c.ReadBuffer, c.WriteBuffer, c.WriteTimeout = o.ReadBuffer, o.WriteBuffer, o.WriteTimeout
	return c, nil
}
