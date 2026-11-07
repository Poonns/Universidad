package protocol

import (
	"fmt"
	"testing"
)

// Fixture setup/registration is excluded. Cyclic cases include dequeues and
// restoring each consumed intention in the same bounded ring, for both paths.
func BenchmarkClientDataAdmission(b *testing.B) {
	for _, optimized := range []bool{false, true} {
		name := "historical"
		if optimized {
			name = "optimized"
		}
		for _, scenario := range []string{"Full", "ManyPendingFull", "Space", "FillDrain"} {
			b.Run(fmt.Sprintf("%s/%s", name, scenario), func(b *testing.B) {
				n, capacity := uint32(25), 1
				if scenario != "Full" {
					n = 5000
					capacity = 2048
				}
				if scenario == "FillDrain" {
					capacity = 8
				}
				g := admissionClients(b, n, optimized, capacity, 64)
				w := &g.workers[0]
				for id := uint32(1); id <= n; id++ {
					if err := w.beginPosition(clientCommand{ClientID: id, Iteration: 1}); err != nil {
						b.Fatal(err)
					}
				}
				if scenario == "Full" || scenario == "ManyPendingFull" {
					for i := 0; i < capacity; i++ {
						w.data.push(clientOutbound{}, false)
					}
				}
				before := w.data.stats().FullAttempts
				var admitted uint64
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					w.flushPending()
					if scenario == "Space" || scenario == "FillDrain" {
						if scenario == "FillDrain" {
							w.flushPending()
						} // second pass sees full DATA
						for {
							job, ok := w.data.pop()
							if !ok {
								break
							}
							admitted++
							c := &g.clients[job.ClientID-1]
							c.Flags &^= clientAwaitConfirmation
							c.Flags |= clientPositionIntent
							w.schedule(job.ClientID, c)
						}
					}
				}
				b.StopTimer()
				full := w.data.stats().FullAttempts - before
				b.ReportMetric(float64(full)/float64(b.N), "full/op")
				b.ReportMetric(float64(full+admitted)/float64(b.N), "push/op")
				b.ReportMetric(float64(admitted)/float64(b.N), "admitted/op")
				if admitted > 0 {
					b.ReportMetric(float64(full+admitted)/float64(admitted), "push/admit")
				}
				if optimized && full != 0 {
					b.Fatal("known-full DATA push")
				}
				if w.pending.stats().Depth != int(n) || len(w.pending.slots) != int(n) || w.pendingControl+w.pendingRetry != 0 {
					b.Fatal("benchmark changed bounded fixture")
				}
			})
		}
	}
}

func TestDataAdmissionHotPathAllocations(t *testing.T) {
	g := admissionClients(t, 25, true, 1, 64)
	w := &g.workers[0]
	w.beginPosition(clientCommand{ClientID: 1, Iteration: 1})
	w.data.push(clientOutbound{}, false)
	if n := testing.AllocsPerRun(1000, func() { w.flushPending() }); n != 0 {
		t.Fatal("full admission allocates", n)
	}
	w.data.pop()
	if n := testing.AllocsPerRun(1000, func() {
		w.flushPending()
		job, ok := w.data.pop()
		if !ok {
			panic("admission missing")
		}
		c := &g.clients[job.ClientID-1]
		c.Flags &^= clientAwaitConfirmation
		c.Flags |= clientPositionIntent
		w.schedule(job.ClientID, c)
	}); n != 0 {
		t.Fatal("space admission allocates", n)
	}
}

func TestDataAdmissionOwnershipAcrossWorkers(t *testing.T) {
	cfg := clientsConfig(t, 150, 25, 1, 3)
	cfg.OptimizeDataAdmission = true
	g := bareClients(t, cfg)
	readyClients(t, g)
	for id := uint32(1); id <= 150; id++ {
		owner, _ := g.WorkerForClient(id)
		if _, err := g.workers[(owner+1)%3].client(id); err != ErrIdentity {
			t.Fatal("foreign ownership accepted")
		}
		if err := g.workers[owner].beginPosition(clientCommand{ClientID: id, Iteration: 1}); err != nil {
			t.Fatal(err)
		}
	}
	for i := range g.workers {
		w := &g.workers[i]
		w.flushPending()
		count := 0
		for {
			job, ok := w.data.pop()
			if !ok {
				break
			}
			owner, _ := g.WorkerForClient(job.ClientID)
			if owner != i || int((job.Packet.GroupID-1)%3) != i {
				t.Fatal("admission moved ownership", job)
			}
			count++
		}
		if count != 50 || w.pending.stats().Depth != 0 {
			t.Fatal("owned jobs lost", i, count)
		}
	}
}
