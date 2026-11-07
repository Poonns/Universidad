package protocol

import (
	"net"
	"sync/atomic"
)

// Each worker has exactly one outbound consumer and confirmation producer.
// Mutable cursors/quotas/buffer/timer belong to that writer, never to a group.
type serverWriter struct {
	server                                            *Server
	index                                             int
	socketIndex                                       int
	egressConn                                        *net.UDPConn
	workers                                           []int // immutable, worker index % writer count == index
	wake                                              chan struct{}
	highCursor, dataCursor, controlCursor, highStreak int
	lowerControlTurn                                  bool
	writes                                            uint64
	classWrites                                       [3]uint64 // HIGH, CONTROL, DATA; writer-owned
	published                                         [3]atomic.Uint64
	socketErrors                                      uint64 // writer-owned, only incremented on failure
	publishedSocketErrors                             atomic.Uint64
}

func (w *serverWorker) notifyWriter() { notify(w.server.writers[w.writerIndex].wake) }

// A CONTROL selection owns the aggregate bucket until its actual write. Another
// writer skips busy CONTROL and continues servicing HIGH/DATA. No queue mutex is
// held during I/O. Unlock also wakes a writer that slept while CONTROL was busy.
func (s *Server) releaseControlLease() {
	s.controlMu.Unlock()
	waiters := s.controlWaiters.Swap(0)
	for i := range s.writers {
		if waiters&(uint32(1)<<i) != 0 {
			notify(s.writers[i].wake)
		}
	}
}

func (w *serverWriter) publishCounts() {
	for i := range w.classWrites {
		w.published[i].Store(w.classWrites[i])
	}
	w.publishedSocketErrors.Store(w.socketErrors)
}

func (w *serverWriter) connection() *net.UDPConn {
	if w.egressConn != nil {
		return w.egressConn
	}
	return w.server.conn // bare single-socket unit fixtures; live writers are assigned before startup
}

func (w *serverWriter) recordWrite(high, control bool) {
	class := 2
	if high {
		class = 0
	}
	if control {
		class = 1
	}
	w.classWrites[class]++
	w.writes++
	if w.writes&255 == 0 {
		w.publishCounts()
	}
}

// These private single-writer selectors retain the phase 2/4B test fixtures.
func (s *Server) takeOutbound() (outbound, bool) { return s.writers[0].takeOutbound() }
func (s *Server) takeServerOutbound(now int64) (outbound, serverControlOutbound, bool, int64, bool) {
	return s.writers[0].takeServerOutbound(now)
}

func (s *Server) writerStats() []ServerWriterStats {
	out := make([]ServerWriterStats, len(s.writers))
	for i := range s.writers {
		w := &s.writers[i]
		h, c, d := w.published[0].Load(), w.published[1].Load(), w.published[2].Load()
		out[i] = ServerWriterStats{Index: i, SocketIndex: w.socketIndex, Workers: append([]int(nil), w.workers...), Writes: h + c + d,
			HighWrites: h, ControlWrites: c, DataWrites: d, SocketErrors: w.publishedSocketErrors.Load()}
	}
	return out
}
