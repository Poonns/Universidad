package protocol

import "math"

// clientDeadlineHeap has fixed storage and exactly one entry per active client.
// All methods, including peek/len, belong to the same group-owner worker as its
// ClientState values. There is no stale-event history, lock or per-client timer.
// ClientState.HeapIndex must be initialized to -1 before the worker starts.
type clientDeadlineHeap struct {
	clients []ClientState
	indices []uint32
	size    int
	maximum int
}

func newClientDeadlineHeap(clients []ClientState, capacity int) clientDeadlineHeap {
	return clientDeadlineHeap{clients: clients, indices: make([]uint32, capacity)}
}

func (h *clientDeadlineHeap) len() int { return h.size }

func (h *clientDeadlineHeap) peek() (index uint32, deadline int64, ok bool) {
	if h.size == 0 {
		return 0, 0, false
	}
	index = h.indices[0]
	return index, h.clients[index].Deadline, true
}

func (h *clientDeadlineHeap) less(a, b int) bool {
	ai, bi := h.indices[a], h.indices[b]
	ad, bd := h.clients[ai].Deadline, h.clients[bi].Deadline
	return ad < bd || ad == bd && ai < bi
}

func (h *clientDeadlineHeap) swap(a, b int) {
	h.indices[a], h.indices[b] = h.indices[b], h.indices[a]
	h.clients[h.indices[a]].HeapIndex = int32(a)
	h.clients[h.indices[b]].HeapIndex = int32(b)
}

func (h *clientDeadlineHeap) up(pos int) int {
	for pos > 0 {
		parent := (pos - 1) / 2
		if !h.less(pos, parent) {
			break
		}
		h.swap(pos, parent)
		pos = parent
	}
	return pos
}

func (h *clientDeadlineHeap) down(pos int) {
	for pos < h.size/2 {
		child := pos*2 + 1
		if child+1 < h.size && h.less(child+1, child) {
			child++
		}
		if !h.less(child, pos) {
			return
		}
		h.swap(pos, child)
		pos = child
	}
}

// set inserts a previously absent client or updates its existing entry in place.
// False means bad index/deadline, exhausted capacity or inconsistent ownership;
// it never grows storage or changes a rejected client's deadline.
func (h *clientDeadlineHeap) set(index uint32, deadline int64) bool {
	if uint64(index) >= uint64(len(h.clients)) || deadline < 0 {
		return false
	}
	c := &h.clients[index]
	pos := int(c.HeapIndex)
	if pos == -1 {
		if h.size == len(h.indices) || h.size > math.MaxInt32 {
			return false
		}
		pos = h.size
		h.indices[pos] = index
		c.HeapIndex = int32(pos)
		c.Deadline = deadline
		h.size++
		if h.size > h.maximum {
			h.maximum = h.size
		}
		h.up(pos)
		return true
	}
	if pos < 0 || pos >= h.size || h.indices[pos] != index {
		return false
	}
	previous := c.Deadline
	c.Deadline = deadline
	if deadline < previous {
		h.up(pos)
	} else if deadline > previous {
		h.down(pos)
	}
	return true
}

func (h *clientDeadlineHeap) remove(index uint32) bool {
	if uint64(index) >= uint64(len(h.clients)) {
		return false
	}
	c := &h.clients[index]
	pos := int(c.HeapIndex)
	if pos < 0 || pos >= h.size || h.indices[pos] != index {
		return false
	}
	h.size--
	if pos != h.size {
		h.indices[pos] = h.indices[h.size]
		h.clients[h.indices[pos]].HeapIndex = int32(pos)
	}
	h.indices[h.size] = 0
	c.HeapIndex, c.Deadline = -1, 0
	if pos < h.size {
		// One moved entry can need either direction, never a complete rebuild.
		if moved := h.up(pos); moved == pos {
			h.down(pos)
		}
	}
	return true
}

// clear cancels all active deadlines in O(active), preserving fixed capacity and
// the historical high-water mark. Only this heap's owner may call it at shutdown.
func (h *clientDeadlineHeap) clear() {
	for i := 0; i < h.size; i++ {
		c := &h.clients[h.indices[i]]
		c.HeapIndex, c.Deadline = -1, 0
		h.indices[i] = 0
	}
	h.size = 0
}

const clientTokenScale int64 = 1_000_000_000 // one token in fixed-point units

// clientTokenBucket is aggregate, writer-owned state, not a per-client object.
// Configuration validates rate in [1,1e9], burst in [1,MaxInt64/1e9], and a
// monotonic nonnegative clock. The initial burst is available immediately.
type clientTokenBucket struct {
	rate, capacity, balance, updatedAt int64
}

func newClientTokenBucket(rate, burst int, now int64) clientTokenBucket {
	capacity := int64(burst) * clientTokenScale
	return clientTokenBucket{rate: int64(rate), capacity: capacity, balance: capacity, updatedAt: now}
}

func clientCeilDivide(n, d int64) int64 {
	q := n / d
	if n%d != 0 {
		q++
	}
	return q
}

func (b *clientTokenBucket) refill(now int64) {
	if now <= b.updatedAt {
		return
	}
	elapsed := now - b.updatedAt
	missing := b.capacity - b.balance
	// Compare before multiplication: elapsed*rate would overflow after a long
	// idle period, but only the finite missing balance is useful.
	if elapsed >= clientCeilDivide(missing, b.rate) {
		b.balance = b.capacity
	} else {
		b.balance += elapsed * b.rate
	}
	b.updatedAt = now
}

// ready returns an absolute monotonic ns deadline; now means a token is ready.
// It updates refill accounting but never consumes a token. The ceiling prevents
// sub-nanosecond rounding from releasing an operation earlier than its budget.
func (b *clientTokenBucket) ready(now int64) int64 {
	b.refill(now)
	if b.balance >= clientTokenScale {
		return now
	}
	wait := clientCeilDivide(clientTokenScale-b.balance, b.rate)
	base := now
	if b.updatedAt > base {
		base = b.updatedAt
	}
	if wait > math.MaxInt64-base {
		return math.MaxInt64
	}
	return base + wait
}

func (b *clientTokenBucket) take(now int64) bool {
	b.refill(now)
	if b.balance < clientTokenScale {
		return false
	}
	b.balance -= clientTokenScale
	return true
}
