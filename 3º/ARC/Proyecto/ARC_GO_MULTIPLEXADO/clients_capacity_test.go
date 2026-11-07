package protocol

import (
	"testing"
	"unsafe"
)

func TestClientHighCapacityBoundedAllocation(t *testing.T) {
	fixture := clientsConfig(t, 25, 25, 1, 6)
	if DefaultClientConfig(fixture.Manifest, fixture.GeneratorID, fixture.Server).HighCapacity != 1024 {
		t.Fatal("public HIGH default changed")
	}
	for _, capacity := range []int{1024, 2048} {
		cfg := fixture
		cfg.HighCapacity = capacity
		g := bareClients(t, cfg)
		for _, w := range g.workers {
			for i := 0; i < capacity; i++ {
				if !w.high.push(clientOutbound{}, true) {
					t.Fatal("queue rejected before fixed capacity")
				}
			}
			if w.high.push(clientOutbound{}, true) || len(w.high.slots) != capacity {
				t.Fatal("queue grew or accepted beyond capacity")
			}
			q := w.high.stats()
			if q.Capacity != capacity || q.Depth != capacity || q.Maximum != capacity || q.FullAttempts != 1 || q.Drops != 1 {
				t.Fatal("bounded admission accounting", q)
			}
			allocs := testing.AllocsPerRun(100, func() {
				if _, ok := w.high.pop(); !ok || !w.high.push(clientOutbound{}, true) {
					t.Fatal("fixed ring reuse failed")
				}
			})
			if allocs != 0 || len(w.high.slots) != capacity {
				t.Fatal("ring reuse allocated or resized", allocs)
			}
			w.high.discardAll()
			if w.high.stats().Depth != 0 || len(w.high.slots) != capacity {
				t.Fatal("cleanup did not drain fixed ring")
			}
		}
	}
	t.Logf("HIGH default1024 and explicit2048, fixed rings and 0 allocs per reuse; extra slot storage W6: %d bytes (slot=%d bytes)", 6*1024*unsafe.Sizeof(clientOutbound{}), unsafe.Sizeof(clientOutbound{}))
}
