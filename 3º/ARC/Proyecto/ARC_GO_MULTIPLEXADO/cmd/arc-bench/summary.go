package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type rangeStat struct{ Mean, Min, Max float64 }

func stats(values []float64) rangeStat {
	var r rangeStat
	if len(values) == 0 {
		return r
	}
	r.Min, r.Max = values[0], values[0]
	for _, v := range values {
		r.Mean += v
		if v < r.Min {
			r.Min = v
		}
		if v > r.Max {
			r.Max = v
		}
	}
	r.Mean /= float64(len(values))
	return r
}

type summary struct {
	Name                                                                                              string
	N, Rate, PositionWindow, Repetitions, Successful, Stable                                          int
	ServerWriterCount                                                                                 int
	ServerEgressSocketCount                                                                           int
	GeneratorHighBurst                                                                                int
	GeneratorHighCapacity                                                                             int
	GeneratorWriter                                                                                   generatorWriterSummary
	Egress                                                                                            []egressSummary
	Writers                                                                                           []writerSummary
	Profile                                                                                           bool
	SimulationSeconds, RTTMilliseconds, ServerCPU, ClientCPU, ServerRSSMiB, ClientRSSMiB              rangeStat
	ServerReceivedPS, ServerSentPS, ClientReceivedPS, ClientSentPS, TotalSent                         rangeStat
	ServerHeapMiB, ClientHeapMiB, ServerHeapSysMiB, ClientHeapSysMiB, ServerGC, ClientGC              rangeStat
	ServerQueues, ClientQueues                                                                        map[string]int
	RegisterRetries, PositionRetries, MetricsRetries, StartRetries, DoneRetries, Timeouts, QueueDrops uint64
	Errors                                                                                            []string
	PositionPS, DataGlobalPS, InflightMaximum                                                         rangeStat
	ServerDataMaximum, ClientDataMaximum                                                              rangeStat
	ServerData50Seconds, ServerData75Seconds, ServerData90Seconds                                     rangeStat
	ClientData50Seconds, ClientData75Seconds, ClientData90Seconds                                     rangeStat
	ServerDataFullAttempts, ClientDataFullAttempts, ServerDataFullPS, ClientDataFullPS                rangeStat
	ClientHighMaximum, ClientHigh50Seconds, ClientHigh75Seconds, ClientHigh90Seconds                  rangeStat
	ClientHighMaximumPercent                                                                          rangeStat
	ClientHighFullAttempts, ClientHighFullPS, ClientHighDrops                                         rangeStat
}

type writerSummary struct{ Writes, HighWrites, ControlWrites, DataWrites rangeStat }
type generatorWriterSummary struct{ Writes, HighWrites, ControlWrites, RetryWrites, DataWrites rangeStat }
type egressSummary struct {
	SocketIndex, WriterIndex                                    int
	Writes, HighWrites, ControlWrites, DataWrites, SocketErrors rangeStat
}

func summarize(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var all []experiment
		if err = json.Unmarshal(data, &all); err != nil {
			return err
		}
		if len(all) == 0 {
			continue
		}
		for _, x := range all {
			if x.GeneratorCount > 1 {
				return fmt.Errorf("%s: multi-generator reports require the FASE_5I aggregate auditor; legacy summary only supports one generator", file)
			}
		}
		s := summary{Name: filepath.Dir(file), N: all[0].N, Rate: all[0].Rate, PositionWindow: all[0].PositionWindow, Profile: all[0].Profile, Repetitions: len(all), ServerQueues: map[string]int{}, ClientQueues: map[string]int{}}
		s.GeneratorHighBurst = all[0].GeneratorHighBurst
		s.GeneratorHighCapacity = all[0].GeneratorHighCapacity
		if s.GeneratorHighCapacity == 0 { // Historical reports retain the actual queue capacity.
			s.GeneratorHighCapacity = all[0].Client.QueuePeaks["HIGH"].Capacity
		}
		if s.GeneratorHighBurst == 0 {
			var cfg struct{ HighBurst int }
			if err := json.Unmarshal(all[0].Client.Config, &cfg); err != nil {
				return err
			}
			s.GeneratorHighBurst = cfg.HighBurst
		}
		s.ServerWriterCount = all[0].ServerWriterCount
		if s.ServerWriterCount == 0 {
			s.ServerWriterCount = 1
		}
		s.Writers = make([]writerSummary, s.ServerWriterCount)
		s.ServerEgressSocketCount = all[0].ServerEgressSocketCount
		if s.ServerEgressSocketCount == 0 {
			s.ServerEgressSocketCount = 1
		}
		if all[0].Server.Server != nil {
			s.Egress = make([]egressSummary, len(all[0].Server.Server.Egress))
		}
		if s.PositionWindow == 0 { // Historical 5A records predate the supervisor flag.
			var cfg struct{ PositionWindow int }
			if err := json.Unmarshal(all[0].Client.Config, &cfg); err != nil {
				return err
			}
			s.PositionWindow = cfg.PositionWindow
		}
		metric := make(map[string][]float64)
		add := func(name string, v float64) { metric[name] = append(metric[name], v) }
		for _, x := range all {
			if x.Success {
				s.Successful++
			}
			if x.Stable {
				s.Stable++
			}
			if x.Error != "" {
				s.Errors = append(s.Errors, x.Error)
			}
			add("scpu", x.ServerOS.CPUPercent)
			add("ccpu", x.ClientOS.CPUPercent)
			add("srss", float64(x.ServerOS.PeakRSSBytes)/1048576)
			add("crss", float64(x.ClientOS.PeakRSSBytes)/1048576)
			add("srx", x.Server.PacketsReceivedPerSecond)
			add("stx", x.Server.PacketsSentPerSecond)
			add("crx", x.Client.PacketsReceivedPerSecond)
			add("ctx", x.Client.PacketsSentPerSecond)
			add("sent", float64(x.TotalSent))
			if w := x.Client.GeneratorWriter; w != nil {
				add("gwTotal", float64(w.Writes))
				add("gwHigh", float64(w.HighWrites))
				add("gwControl", float64(w.ControlWrites))
				add("gwRetry", float64(w.RetryWrites))
				add("gwData", float64(w.DataWrites))
			}
			add("sheap", float64(x.Server.EndRuntime.HeapAlloc)/1048576)
			add("cheap", float64(x.Client.EndRuntime.HeapAlloc)/1048576)
			add("ssys", float64(x.Server.EndRuntime.HeapSys)/1048576)
			add("csys", float64(x.Client.EndRuntime.HeapSys)/1048576)
			add("sgc", float64(x.Server.EndRuntime.NumGC-x.Server.StartRuntime.NumGC))
			add("cgc", float64(x.Client.EndRuntime.NumGC-x.Client.StartRuntime.NumGC))
			if st := x.Server.Server; st != nil {
				for _, e := range st.Egress {
					key := fmt.Sprintf("e%dw%d", e.SocketIndex, e.WriterIndex)
					add(key+"Total", float64(e.Writes))
					add(key+"High", float64(e.HighWrites))
					add(key+"Control", float64(e.ControlWrites))
					add(key+"Data", float64(e.DataWrites))
					add(key+"Errors", float64(e.SocketErrors))
				}
				for _, w := range st.Writers {
					add(fmt.Sprintf("w%dTotal", w.Index), float64(w.Writes))
					add(fmt.Sprintf("w%dHigh", w.Index), float64(w.HighWrites))
					add(fmt.Sprintf("w%dControl", w.Index), float64(w.ControlWrites))
					add(fmt.Sprintf("w%dData", w.Index), float64(w.DataWrites))
				}
				add("sim", st.SimulationDuration.Seconds())
				add("rtt", float64(st.MeanRTTNS)/1e6)
				s.StartRetries += st.StartRetries
				s.DoneRetries += st.GroupDoneRetries
				s.Timeouts += st.ControlTimeouts
				s.QueueDrops += st.QueueDrops
			}
			if ct := x.Client.Client; ct != nil {
				s.RegisterRetries += ct.RegisterRetries
				s.PositionRetries += ct.PositionRetries
				s.MetricsRetries += ct.MetricsRetries
				s.Timeouts += ct.Timeouts
				s.QueueDrops += ct.QueueDrops
				add("inflight", float64(ct.InflightMaximum))
			}
			if st, ct := x.Server.Server, x.Client.Client; st != nil && ct != nil && st.SimulationDuration > 0 {
				duration := st.SimulationDuration.Seconds()
				q := x.Client.QueuePeaks["HIGH"]
				add("chighMax", float64(q.Maximum))
				if q.Capacity > 0 {
					add("chighMaxPercent", 100*float64(q.Maximum)/float64(q.Capacity))
				}
				add("chigh50", float64(q.Above50NS)/1e9)
				add("chigh75", float64(q.Above75NS)/1e9)
				add("chigh90", float64(q.Above90NS)/1e9)
				add("chighFull", float64(q.FullAttempts))
				add("chighFullPS", float64(q.FullAttempts)/duration)
				add("chighDrops", float64(q.Drops))
				add("positions", float64(ct.CompletedOperations)/duration)
				add("data", x.Server.AcademicDataReceivedPerSecond+x.Server.AcademicDataSentPerSecond)
				for _, role := range []string{"server", "client"} {
					q := x.Server.QueuePeaks["DATA"]
					if role == "client" {
						q = x.Client.QueuePeaks["DATA"]
					}
					add(role+"dataMax", float64(q.Maximum))
					add(role+"data50", float64(q.Above50NS)/1e9)
					add(role+"data75", float64(q.Above75NS)/1e9)
					add(role+"data90", float64(q.Above90NS)/1e9)
					add(role+"dataFull", float64(q.FullAttempts))
					// DATA work belongs to the simulation; use the same denominator for both endpoints.
					add(role+"dataFullPS", float64(q.FullAttempts)/duration)
				}
			}
			for k, v := range x.Server.QueuePeaks {
				if old, exists := s.ServerQueues[k]; !exists || v.Maximum > old {
					s.ServerQueues[k] = v.Maximum
				}
			}
			for k, v := range x.Client.QueuePeaks {
				if old, exists := s.ClientQueues[k]; !exists || v.Maximum > old {
					s.ClientQueues[k] = v.Maximum
				}
			}
		}
		s.SimulationSeconds = stats(metric["sim"])
		s.RTTMilliseconds = stats(metric["rtt"])
		s.ServerCPU = stats(metric["scpu"])
		s.ClientCPU = stats(metric["ccpu"])
		s.ServerRSSMiB = stats(metric["srss"])
		s.ClientRSSMiB = stats(metric["crss"])
		s.ServerReceivedPS = stats(metric["srx"])
		s.ServerSentPS = stats(metric["stx"])
		s.ClientReceivedPS = stats(metric["crx"])
		s.ClientSentPS = stats(metric["ctx"])
		s.TotalSent = stats(metric["sent"])
		s.GeneratorWriter = generatorWriterSummary{stats(metric["gwTotal"]), stats(metric["gwHigh"]), stats(metric["gwControl"]), stats(metric["gwRetry"]), stats(metric["gwData"])}
		s.ClientHighMaximum, s.ClientHigh50Seconds, s.ClientHigh75Seconds, s.ClientHigh90Seconds = stats(metric["chighMax"]), stats(metric["chigh50"]), stats(metric["chigh75"]), stats(metric["chigh90"])
		s.ClientHighMaximumPercent = stats(metric["chighMaxPercent"])
		s.ClientHighFullAttempts, s.ClientHighFullPS, s.ClientHighDrops = stats(metric["chighFull"]), stats(metric["chighFullPS"]), stats(metric["chighDrops"])
		s.PositionPS, s.DataGlobalPS, s.InflightMaximum = stats(metric["positions"]), stats(metric["data"]), stats(metric["inflight"])
		s.ServerDataMaximum, s.ClientDataMaximum = stats(metric["serverdataMax"]), stats(metric["clientdataMax"])
		s.ServerData50Seconds, s.ServerData75Seconds, s.ServerData90Seconds = stats(metric["serverdata50"]), stats(metric["serverdata75"]), stats(metric["serverdata90"])
		s.ClientData50Seconds, s.ClientData75Seconds, s.ClientData90Seconds = stats(metric["clientdata50"]), stats(metric["clientdata75"]), stats(metric["clientdata90"])
		s.ServerDataFullAttempts, s.ClientDataFullAttempts = stats(metric["serverdataFull"]), stats(metric["clientdataFull"])
		s.ServerDataFullPS, s.ClientDataFullPS = stats(metric["serverdataFullPS"]), stats(metric["clientdataFullPS"])
		s.ServerHeapMiB = stats(metric["sheap"])
		s.ClientHeapMiB = stats(metric["cheap"])
		s.ServerHeapSysMiB = stats(metric["ssys"])
		s.ClientHeapSysMiB = stats(metric["csys"])
		s.ServerGC = stats(metric["sgc"])
		s.ClientGC = stats(metric["cgc"])
		for i := range s.Writers {
			s.Writers[i] = writerSummary{stats(metric[fmt.Sprintf("w%dTotal", i)]), stats(metric[fmt.Sprintf("w%dHigh", i)]), stats(metric[fmt.Sprintf("w%dControl", i)]), stats(metric[fmt.Sprintf("w%dData", i)])}
		}
		for i := range s.Egress {
			e := all[0].Server.Server.Egress[i]
			key := fmt.Sprintf("e%dw%d", e.SocketIndex, e.WriterIndex)
			s.Egress[i] = egressSummary{e.SocketIndex, e.WriterIndex, stats(metric[key+"Total"]), stats(metric[key+"High"]), stats(metric[key+"Control"]), stats(metric[key+"Data"]), stats(metric[key+"Errors"])}
		}
		if err = json.NewEncoder(os.Stdout).Encode(s); err != nil {
			return err
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("no experiment files match %s", pattern)
	}
	return nil
}
