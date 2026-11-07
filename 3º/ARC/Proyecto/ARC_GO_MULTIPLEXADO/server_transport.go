package protocol

import (
	"fmt"
	"net"
	"net/netip"
)

// Count=1 preserves the approved shared ingress/TX descriptor. Count=2/3 uses
// that many stable TX-only sockets; none gets a reader or receives requests.
func (s *Server) openEgress() error {
	if s.cfg.ServerEgressSocketCount == 1 {
		s.egress = []*net.UDPConn{s.conn}
	} else {
		bind := net.UDPAddrFromAddrPort(netip.AddrPortFrom(s.address.Addr(), 0))
		s.egress = make([]*net.UDPConn, 0, s.cfg.ServerEgressSocketCount)
		for i := 0; i < s.cfg.ServerEgressSocketCount; i++ {
			conn, err := net.ListenUDP("udp", bind)
			if err != nil {
				return fmt.Errorf("egress %d bind: %w", i, err)
			}
			s.egress = append(s.egress, conn) // include partial startup in cleanup
			if s.cfg.WriteBuffer > 0 {
				if err = conn.SetWriteBuffer(s.cfg.WriteBuffer); err != nil {
					return fmt.Errorf("egress %d write buffer: %w", i, err)
				}
			}
		}
	}
	for i := range s.writers {
		w := &s.writers[i]
		w.egressConn = s.egress[w.socketIndex]
	}
	return nil
}

func (s *Server) closeSockets() {
	if s.conn != nil {
		s.conn.Close()
	}
	for _, c := range s.egress {
		if c != s.conn {
			c.Close()
		}
	}
}

// Addr remains the sole ingress endpoint. EgressEndpoints returns a copy; the
// mapping and socket identities never change during a run or after Close.
func (s *Server) EgressEndpoints() []netip.AddrPort {
	out := make([]netip.AddrPort, len(s.egress))
	for i, c := range s.egress {
		out[i] = normalizeEndpoint(c.LocalAddr().(*net.UDPAddr).AddrPort())
	}
	return out
}

// One row per writer/socket relation. With a shared socket both rows use index
// zero; summing by SocketIndex yields physical per-egress counts without overlap.
type ServerEgressStats struct {
	SocketIndex, WriterIndex                                    int
	Endpoint                                                    netip.AddrPort
	Writes, HighWrites, ControlWrites, DataWrites, SocketErrors uint64
}

func (s *Server) egressStats(writers []ServerWriterStats) []ServerEgressStats {
	out := make([]ServerEgressStats, len(writers))
	endpoints := s.EgressEndpoints()
	for i, w := range writers {
		var endpoint netip.AddrPort
		if w.SocketIndex < len(endpoints) {
			endpoint = endpoints[w.SocketIndex]
		}
		out[i] = ServerEgressStats{SocketIndex: w.SocketIndex, WriterIndex: w.Index, Endpoint: endpoint, Writes: w.Writes,
			HighWrites: w.HighWrites, ControlWrites: w.ControlWrites, DataWrites: w.DataWrites, SocketErrors: w.SocketErrors}
	}
	return out
}
