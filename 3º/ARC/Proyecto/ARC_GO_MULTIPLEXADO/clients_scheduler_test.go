package protocol

import (
	"math"
	"testing"
)

func deadlineClients(n int) []ClientState {
	clients := make([]ClientState, n)
	for i := range clients {
		clients[i].HeapIndex = -1
	}
	return clients
}

func checkDeadlineHeap(t *testing.T, h *clientDeadlineHeap) {
	t.Helper()
	if h.size < 0 || h.size > len(h.indices) || h.maximum < h.size {
		t.Fatalf("invalid size: size=%d capacity=%d maximum=%d", h.size, len(h.indices), h.maximum)
	}
	seen := make([]bool, len(h.clients))
	for pos := 0; pos < h.size; pos++ {
		index := h.indices[pos]
		if uint64(index) >= uint64(len(h.clients)) || seen[index] || h.clients[index].HeapIndex != int32(pos) {
			t.Fatalf("inconsistent indexed entry: pos=%d client=%d", pos, index)
		}
		seen[index] = true
		if pos > 0 && h.less(pos, (pos-1)/2) {
			t.Fatalf("heap order violated: child=%d parent=%d", pos, (pos-1)/2)
		}
	}
}

func TestClientDeadlineHeapOrdersAndUpdates(t *testing.T) {
	clients := deadlineClients(8)
	h := newClientDeadlineHeap(clients, len(clients))
	for _, entry := range []struct {
		index    uint32
		deadline int64
	}{{7, 700}, {6, 600}, {5, 500}, {4, 400}, {3, 300}, {2, 200}, {1, 100}, {0, 0}} {
		if !h.set(entry.index, entry.deadline) {
			t.Fatal("insert failed", entry)
		}
		checkDeadlineHeap(t, &h)
	}
	if !h.set(0, 900) || !h.set(7, 50) {
		t.Fatal("up/down update failed")
	}
	checkDeadlineHeap(t, &h)
	if index, deadline, ok := h.peek(); !ok || index != 7 || deadline != 50 || h.len() != 8 || h.maximum != 8 {
		t.Fatal("updated minimum", index, deadline, ok, h.len(), h.maximum)
	}
	// Equal deadlines have a stable local-index order independent of insertion.
	for index := range clients {
		if !h.set(uint32(index), 42) {
			t.Fatal("in-place update failed")
		}
		checkDeadlineHeap(t, &h)
	}
	for expected := uint32(0); expected < 8; expected++ {
		index, deadline, ok := h.peek()
		if !ok || index != expected || deadline != 42 || !h.remove(index) {
			t.Fatal("unstable tie/order", expected, index, deadline, ok)
		}
		if clients[index].HeapIndex != -1 || clients[index].Deadline != 0 {
			t.Fatal("removed entry retained deadline")
		}
		checkDeadlineHeap(t, &h)
	}
	if _, _, ok := h.peek(); ok || h.len() != 0 || h.maximum != 8 {
		t.Fatal("empty heap/high-water state")
	}
}

func TestClientDeadlineHeapBoundedAndInvalid(t *testing.T) {
	clients := deadlineClients(3)
	h := newClientDeadlineHeap(clients, 2)
	backing := &h.indices[0]
	if !h.set(0, 10) || !h.set(1, 20) {
		t.Fatal("valid insert failed")
	}
	before := clients[2]
	if h.set(2, 30) || clients[2] != before || len(h.indices) != 2 || cap(h.indices) != 2 || &h.indices[0] != backing {
		t.Fatal("full heap grew storage or changed rejected client")
	}
	// Updates are legal at capacity; they never create another event.
	for i := 0; i < 1000; i++ {
		if !h.set(0, int64(i)) || h.len() != 2 {
			t.Fatal("update accumulated stale entries")
		}
	}
	for _, index := range []uint32{3, math.MaxUint32} {
		if h.set(index, 1) || h.remove(index) {
			t.Fatal("invalid client index accepted", index)
		}
	}
	before = clients[0]
	if h.set(0, -1) || clients[0] != before || h.remove(2) {
		t.Fatal("invalid deadline/absent remove changed state")
	}
	if !h.remove(1) || h.remove(1) || !h.set(2, 0) {
		t.Fatal("slot was not reusable")
	}
	checkDeadlineHeap(t, &h)
	h.clear()
	if h.len() != 0 || h.maximum != 2 || len(h.indices) != 2 || &h.indices[0] != backing {
		t.Fatal("clear changed capacity/high-water mark")
	}
	for _, c := range clients {
		if c.HeapIndex != -1 || c.Deadline != 0 {
			t.Fatal("clear retained active deadline")
		}
	}
	zero := newClientDeadlineHeap(clients, 0)
	if zero.set(0, 1) || zero.remove(0) || zero.len() != 0 {
		t.Fatal("worker without clients acquired an entry")
	}
	zero.clear()
}

func TestClientDeadlineHeapSharedStatesKeepOwnerIsolation(t *testing.T) {
	clients := deadlineClients(6)
	a, b := newClientDeadlineHeap(clients, 3), newClientDeadlineHeap(clients, 3)
	for _, index := range []uint32{0, 2, 4} {
		if !a.set(index, int64(index)+100) {
			t.Fatal("owner A failed")
		}
	}
	for _, index := range []uint32{1, 3, 5} {
		if !b.set(index, int64(index)+200) {
			t.Fatal("owner B failed")
		}
	}
	before := clients[1]
	if a.set(1, 0) || a.remove(1) || clients[1] != before {
		t.Fatal("foreign active client modified by wrong heap")
	}
	a.clear()
	if b.len() != 3 || clients[1] != before || clients[3].HeapIndex == -1 || clients[5].HeapIndex == -1 {
		t.Fatal("owner clear changed another owner's state")
	}
	checkDeadlineHeap(t, &b)
	b.clear()
}

func TestClientDeadlineHeapDeterministicMixedOperations(t *testing.T) {
	const n = 32
	clients := deadlineClients(n)
	h := newClientDeadlineHeap(clients, n)
	var active [n]bool
	var deadlines [n]int64
	state := uint32(17)
	for turn := 0; turn < 2000; turn++ {
		// Fixed LCG test stream, no random generator or nondeterministic clock.
		state = state*1664525 + 1013904223
		index := (state >> 16) % n
		state = state*1664525 + 1013904223
		if (state>>24)&3 == 0 {
			if h.remove(index) != active[index] {
				t.Fatal("remove/reference mismatch", turn, index)
			}
			active[index] = false
		} else {
			deadline := int64(state % 101)
			if !h.set(index, deadline) {
				t.Fatal("set failed", turn, index)
			}
			active[index], deadlines[index] = true, deadline
		}
		checkDeadlineHeap(t, &h)
		expected, count := uint32(0), 0
		for candidate, exists := range active {
			if exists {
				if count == 0 || deadlines[candidate] < deadlines[expected] || deadlines[candidate] == deadlines[expected] && uint32(candidate) < expected {
					expected = uint32(candidate)
				}
				count++
			}
		}
		index, deadline, ok := h.peek()
		if h.len() != count || ok != (count != 0) || ok && (index != expected || deadline != deadlines[expected]) {
			t.Fatalf("reference mismatch turn=%d got=%d/%d/%t expected=%d count=%d", turn, index, deadline, ok, expected, count)
		}
	}
}

func TestClientTokenBucketBurstRateAndRefill(t *testing.T) {
	b := newClientTokenBucket(4, 3, 100)
	for i := 0; i < 3; i++ {
		if b.ready(100) != 100 || !b.take(100) {
			t.Fatal("initial burst unavailable", i)
		}
	}
	if b.take(100) || b.ready(100) != 250000100 || b.take(250000099) {
		t.Fatal("bucket exceeded burst/rate or rounded down")
	}
	if !b.take(250000100) || b.take(250000100) || b.ready(250000100) != 500000100 {
		t.Fatal("next token not replenished exactly at rate")
	}
	// A long idle restores only burst, not unlimited accumulated work.
	for i := 0; i < 3; i++ {
		if !b.take(10_000_000_100) {
			t.Fatal("burst did not replenish", i)
		}
	}
	if b.take(10_000_000_100) {
		t.Fatal("idle accumulated more than burst")
	}
}

func TestClientTokenBucketFractionalRateAndQueries(t *testing.T) {
	// Burst 2 leaves room for fractions; balance above the configured burst is
	// deliberately discarded by a token bucket when it saturates.
	b := newClientTokenBucket(3, 2, 0)
	if !b.take(0) || !b.take(0) || b.ready(0) != 333333334 {
		t.Fatal("fractional next deadline must round up")
	}
	if b.take(333333333) || b.ready(333333333) != 333333334 {
		t.Fatal("token released before its fractional deadline")
	}
	if !b.take(333333334) || b.balance != 2 {
		t.Fatal("fractional balance was lost", b.balance)
	}
	for i := 0; i < 1000; i++ {
		if b.ready(333333334) != 666666667 {
			t.Fatal("readiness query consumed/lost a token", b)
		}
	}
	if !b.take(666666667) || b.balance != 1 {
		t.Fatal("fraction was not carried to next token", b)
	}
	if b.take(100) || b.updatedAt != 666666667 || b.ready(100) != 1_000_000_000 {
		t.Fatal("backwards clock replenished credits/changed epoch", b)
	}
}

func TestClientTokenBucketLongIdleAndDeadlineOverflow(t *testing.T) {
	b := newClientTokenBucket(1_000_000_000, 3, 0)
	if !b.take(0) || !b.take(0) || !b.take(0) || !b.take(math.MaxInt64) || b.balance != 2*clientTokenScale {
		t.Fatal("large elapsed refill overflowed instead of saturating", b)
	}
	// A next deadline beyond representable ns saturates explicitly.
	c := newClientTokenBucket(1, 1, math.MaxInt64-10)
	if !c.take(math.MaxInt64-10) || c.ready(math.MaxInt64-10) != math.MaxInt64 || c.take(math.MaxInt64) {
		t.Fatal("future deadline wrapped or generated a premature token", c)
	}
	// Largest allowed capacity exercises multiplication protection as well.
	d := newClientTokenBucket(1_000_000_000, int(math.MaxInt64/clientTokenScale), 0)
	for i := 0; i < 3; i++ {
		if !d.take(0) {
			t.Fatal("maximum valid burst unavailable")
		}
	}
	if d.ready(math.MaxInt64) != math.MaxInt64 || d.balance != d.capacity {
		t.Fatal("maximum bucket capacity refill overflowed")
	}
}

func TestClientSchedulerHotPathZeroAllocations(t *testing.T) {
	clients := deadlineClients(16)
	h := newClientDeadlineHeap(clients, len(clients))
	for index := uint32(1); index < 16; index++ {
		h.set(index, int64(index)*1000)
	}
	if n := testing.AllocsPerRun(1000, func() {
		if !h.set(0, 5000) || !h.set(0, 1) || !h.set(0, 9000) || !h.remove(0) {
			panic("deadline invariant")
		}
	}); n != 0 {
		t.Fatal("heap insert/update/remove allocates", n)
	}
	b := newClientTokenBucket(1_000_000_000, 1, 0)
	var now int64
	if n := testing.AllocsPerRun(1000, func() {
		if b.ready(now) != now || !b.take(now) {
			panic("token invariant")
		}
		now++
	}); n != 0 {
		t.Fatal("aggregate bucket allocates", n)
	}
}

func BenchmarkClientDeadlineInsertUpdateRemove(b *testing.B) {
	clients := deadlineClients(100)
	h := newClientDeadlineHeap(clients, len(clients))
	for index := uint32(1); index < 100; index++ {
		h.set(index, int64(index)*1000)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.set(0, 50000)
		h.set(0, 1)
		h.set(0, 99999)
		h.remove(0)
	}
	b.StopTimer()
	if h.len() != 99 {
		b.Fatal("heap accumulated stale entries")
	}
}

func BenchmarkClientTokenBucket(b *testing.B) {
	bucket := newClientTokenBucket(1_000_000_000, 1, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bucket.ready(int64(i)) != int64(i) || !bucket.take(int64(i)) {
			b.Fatal("token budget invariant")
		}
	}
}
