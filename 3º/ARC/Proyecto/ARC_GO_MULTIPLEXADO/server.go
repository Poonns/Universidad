package protocol

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerConfig  = errors.New("server: invalid configuration")
	ErrEndpoint      = errors.New("server: generator endpoint mismatch")
	ErrNotRegistered = errors.New("server: registration barrier not reached")
	ErrConflict      = errors.New("server: conflicting operation")
	ErrQueueFull     = errors.New("server: bounded queue full")
	ErrProgress      = errors.New("server: operation cannot advance yet")
)

// DefaultServerConfig is the initial portable profile. GOMAXPROCS is a process
// setting, not a worker count: use zero to preserve it when embedding the server.
type ServerConfig struct {
	Manifest                                  *Manifest
	Address                                   string
	Workers, GOMAXPROCS                       int
	ServerWriterCount                         int
	ServerEgressSocketCount                   int
	InputCapacity, HighCapacity, DataCapacity int
	WorkBudget, HighBurst                     int
	ReadBuffer, WriteBuffer                   int
	WriteTimeout                              time.Duration
	Academic                                  bool
	ControlCapacity, ConfirmationCapacity     int
	ControlRate, ControlBurst                 int
	InitialControlRTO, MaxControlRTO          time.Duration
	MaxControlSends                           uint16
	ControlJitterPercent                      int
	ControlJitterSeed                         uint64
}

func DefaultServerConfig(m *Manifest) ServerConfig {
	return ServerConfig{Manifest: m, Address: "127.0.0.1:9000", Workers: 6, GOMAXPROCS: 6, ServerWriterCount: 1, ServerEgressSocketCount: 1,
		InputCapacity: 4096, HighCapacity: 1024, DataCapacity: 2048,
		WorkBudget: 64, HighBurst: 8, ReadBuffer: 4 * 1024 * 1024, WriteBuffer: 4 * 1024 * 1024,
		WriteTimeout: time.Second, Academic: true, ControlCapacity: 256, ConfirmationCapacity: 256,
		ControlRate: 1000, ControlBurst: 1, InitialControlRTO: time.Second, MaxControlRTO: 5 * time.Second,
		MaxControlSends: 12, ControlJitterPercent: 20, ControlJitterSeed: 42}
}

// QueueStats is an exact, individually consistent snapshot. FullAttempts includes
// deferred pushes; Drops includes only work explicitly discarded at that queue.
type QueueStats struct {
	Capacity, Depth, Maximum           int
	FullAttempts, Drops, ShutdownDrops uint64
}

// Fixed ring with a lock local to this queue, never held during I/O or processing.
// This keeps occupancy/high-water accounting exact, including concurrent snapshots.
// No queue shares a lock with another queue or with simulation state.
type boundedQueue[T any] struct {
	mu                    sync.Mutex
	slots                 []T
	head, size, maximum   int
	full, drops, shutdown uint64
}

func newQueue[T any](capacity int) *boundedQueue[T] {
	return &boundedQueue[T]{slots: make([]T, capacity)}
}
func (q *boundedQueue[T]) push(v T, discard bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == len(q.slots) {
		q.full++
		if discard {
			q.drops++
		}
		return false
	}
	q.slots[(q.head+q.size)%len(q.slots)] = v
	q.size++
	if q.size > q.maximum {
		q.maximum = q.size
	}
	return true
}
func (q *boundedQueue[T]) pop() (v T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		return v, false
	}
	v = q.slots[q.head]
	var zero T
	q.slots[q.head] = zero
	q.head = (q.head + 1) % len(q.slots)
	q.size--
	return v, true
}
func (q *boundedQueue[T]) peek() (v T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		return v, false
	}
	return q.slots[q.head], true
}
func (q *boundedQueue[T]) stats() QueueStats {
	q.mu.Lock()
	defer q.mu.Unlock()
	return QueueStats{len(q.slots), q.size, q.maximum, q.full, q.drops, q.shutdown}
}
func (q *boundedQueue[T]) discardAll() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := uint64(q.size)
	clear(q.slots)
	q.head = 0
	q.size = 0
	q.drops += n
	q.shutdown += n
	return n
}

const (
	clientRegistered uint16 = 1 << iota
	registerReplyPending
	clientScheduled
	metricsAccepted
	metricsReplyPending
)

// Only the owning group worker may access mutable client/group fields while live.
type ServerClientState struct {
	RegisterMessageID, MessageID                     uint64
	MeanRTTNS                                        uint64
	Position                                         PositionPayload
	Iteration, ExpectedMask, SeenACKMask             uint32
	PendingPositionMask, PendingRetryMask, RetryMask uint32
	RetryCounter, Flags                              uint16
}
type GroupState struct {
	GeneratorIndex                                    uint32 // immutable after initialization
	RegisteredMask, PositionStartedMask, FinalACKMask uint32
	StartAckMask, MetricsMask, GroupDoneAckMask       uint32
	MetricsSumLo, MetricsSumHi                        uint64
	ControlTicket                                     uint64
	ControlRTO, ControlDeadline, ControlLastSentAt    int64
	ControlPendingMask, ControlAwaitMask              uint32
	ControlWrittenMask, ControlRoundMask              uint32
	ControlHeapIndex                                  int32
	ControlAttempt, ControlSends                      uint16
	ControlPhase                                      ServerGroupPhase
	ControlScheduled                                  bool
}

type inbound struct {
	Packet         Packet
	Source         netip.AddrPort
	GeneratorIndex int
	ClientID       uint32 // authoritative assignment for REGISTER; logical origin otherwise
	GroupID        uint32 // includes assignment for REGISTER (wire GroupID is zero)
}
type outbound struct {
	Packet         Packet
	GeneratorIndex int
}
type generatorEndpoint struct {
	address atomic.Pointer[netip.AddrPort]
}

type serverCounters struct {
	packetsReceived, packetsSent                                             atomic.Uint64
	registerReceived, registerAckSent, positionReceived, positionForwarded   atomic.Uint64
	retryReceived, retryForwarded, ackReceived, ackForwarded                 atomic.Uint64
	invalid, duplicateRegisters, queueDrops, socketErrors, registered        atomic.Uint64
	endpointRejects, conflicts, duplicateACKs, sendDrops, shutdownIntentions atomic.Uint64
	startSent, startAckReceived, startRetries, startedClients, startedGroups atomic.Uint64
	metricsReceived, metricsAckSent, metricsAccepted, duplicateMetrics       atomic.Uint64
	groupDoneSent, groupDoneAckReceived, groupDoneRetries, doneClients       atomic.Uint64
	completedGroups, controlTimeouts, controlRetryExhausted, controlFailures atomic.Uint64
	duplicateStartACKs, duplicateGroupDoneACKs, stalePackets                 atomic.Uint64
	cancelledControls, staleControlConfirmations                             atomic.Uint64
}
type WorkerQueueStats struct{ Input, High, Data, PendingClients, Control, Confirmations, PendingGroups QueueStats }

// Writer counters are published on idle, every 256 writes and on exit. Live
// snapshots can lag; after Close these counts are exact. Formatting/allocation
// belongs to Stats, not to the per-datagram path.
type ServerWriterStats struct {
	Index                                         int
	SocketIndex                                   int
	Workers                                       []int
	Writes, HighWrites, ControlWrites, DataWrites uint64
	SocketErrors                                  uint64
}
type ServerStats struct {
	PacketsReceived, PacketsSent                                                    uint64
	RegisterReceived, RegisterAckSent                                               uint64
	PositionReceived, PositionForwarded                                             uint64
	RetryPositionReceived, RetryPositionForwarded                                   uint64
	AckReceived, AckForwarded                                                       uint64
	InvalidPackets, DuplicateRegisters, QueueDrops, SocketErrors, RegisteredClients uint64
	EndpointRejects, Conflicts, DuplicateACKs, SendDrops, ShutdownIntentions        uint64
	RegistrationBarrierReached                                                      bool
	Queues                                                                          []WorkerQueueStats
	StartSent, StartAckReceived, StartRetries, StartedClients, StartedGroups        uint64
	MetricsReceived, MetricsAckSent, MetricsAccepted, DuplicateMetrics              uint64
	GroupDoneSent, GroupDoneAckReceived, GroupDoneRetries, DoneClients              uint64
	CompletedGroups, ControlTimeouts, ControlRetryExhausted, ControlFailures        uint64
	DuplicateStartACKs, DuplicateGroupDoneACKs, StalePackets                        uint64
	CancelledControls, StaleControlConfirmations                                    uint64
	ControlHeapCurrent, ControlHeapMaximum, ControlTimers, WriterTimers             int64
	SimulationComplete                                                              bool
	N, V, I                                                                         uint32
	MeanRTTNS                                                                       uint64
	SimulationDuration                                                              time.Duration
	Writers                                                                         []ServerWriterStats
	Egress                                                                          []ServerEgressStats
}

type serverWorker struct {
	server           *Server
	index            int
	writerIndex      int // immutable: index % ServerWriterCount
	input            *boundedQueue[inbound]
	high, data       *boundedQueue[outbound]
	pending          *boundedQueue[uint32] // at most one entry for each owned client
	wake             chan struct{}         // coalesced notification, not a work queue
	control          *boundedQueue[serverControlOutbound]
	confirmations    *boundedQueue[serverControlConfirmation]
	pendingGroups    *boundedQueue[uint32]
	controlDeadlines serverControlDeadlineHeap
	startCursor      int
}

type Server struct {
	cfg                                     ServerConfig
	clients                                 []ServerClientState
	groups                                  []GroupState
	endpoints                               []generatorEndpoint
	workers                                 []serverWorker
	writers                                 []serverWriter
	counters                                serverCounters
	conn                                    *net.UDPConn
	egress                                  []*net.UDPConn // immutable after startup; shared ingress alias for count=1
	address                                 netip.AddrPort
	ctx                                     context.Context
	cancel                                  context.CancelFunc
	done                                    chan struct{}
	fatal                                   chan error
	wg                                      sync.WaitGroup
	runErr                                  error      // supervisor writes before closing done
	controlMu                               sync.Mutex // one aggregate CONTROL lease; never locks HIGH/DATA
	controlWaiters                          atomic.Uint32
	controlBucket                           clientTokenBucket
	controlGates                            []serverControlGate
	groupSnapshots                          []serverGroupSnapshot
	epoch                                   time.Time
	simulationDone                          chan struct{}
	simulationOutcome                       atomic.Uint32
	simulationStartNS, simulationEndNS      atomic.Int64
	controlHeapCurrent, controlHeapMaximum  atomic.Int64
	activeControlTimers, activeWriterTimers atomic.Int64
}

func newServer(cfg ServerConfig) (*Server, error) {
	if cfg.Manifest == nil || cfg.Manifest.run.Validate() != nil || len(cfg.Manifest.generators) == 0 ||
		cfg.Workers < 1 || cfg.ServerWriterCount < 1 || cfg.ServerWriterCount > 3 || cfg.ServerWriterCount > cfg.Workers ||
		cfg.ServerEgressSocketCount < 1 || cfg.ServerEgressSocketCount > 3 || cfg.ServerEgressSocketCount > cfg.ServerWriterCount ||
		(cfg.ServerWriterCount == 3 && cfg.ServerEgressSocketCount != 3) ||
		cfg.GOMAXPROCS < 0 || cfg.InputCapacity < 1 || cfg.HighCapacity < 1 || cfg.DataCapacity < 1 ||
		cfg.WorkBudget < 1 || cfg.HighBurst < 1 || cfg.ReadBuffer < 0 || cfg.WriteBuffer < 0 || cfg.WriteTimeout <= 0 {
		return nil, ErrServerConfig
	}
	if cfg.Academic && (cfg.ControlCapacity < 1 || cfg.ConfirmationCapacity < 1 || cfg.ControlRate < 1 ||
		cfg.ControlRate > 1_000_000_000 || cfg.ControlBurst < 1 || uint64(cfg.ControlBurst) > uint64(math.MaxInt64/1_000_000_000) ||
		cfg.InitialControlRTO <= 0 || cfg.MaxControlRTO < cfg.InitialControlRTO || cfg.MaxControlRTO > time.Duration(math.MaxInt64/4) ||
		cfg.MaxControlSends < 1 || cfg.ControlJitterPercent < 0 || cfg.ControlJitterPercent > 50) {
		return nil, ErrServerConfig
	}
	// Check sizes before converting to platform int (notably on 32-bit builds).
	maxInt := uint64(^uint(0) >> 1)
	if uint64(cfg.Manifest.run.N) > maxInt || uint64(cfg.Workers) > uint64(math.MaxUint32) {
		return nil, ErrServerConfig
	}
	s := &Server{cfg: cfg, clients: make([]ServerClientState, int(cfg.Manifest.run.N)),
		groups:    make([]GroupState, int(cfg.Manifest.run.N/cfg.Manifest.run.V)),
		endpoints: make([]generatorEndpoint, len(cfg.Manifest.generators)), workers: make([]serverWorker, cfg.Workers),
		writers: make([]serverWriter, cfg.ServerWriterCount),
		done:    make(chan struct{}), fatal: make(chan error, 1),
		epoch: time.Now(), simulationDone: make(chan struct{})}
	for i := range s.writers {
		s.writers[i] = serverWriter{server: s, index: i, socketIndex: i % cfg.ServerEgressSocketCount, wake: make(chan struct{}, 1), lowerControlTurn: true}
		for worker := i; worker < cfg.Workers; worker += cfg.ServerWriterCount {
			s.writers[i].workers = append(s.writers[i].workers, worker)
		}
	}
	s.simulationStartNS.Store(-1)
	s.simulationEndNS.Store(-1)
	s.controlGates = make([]serverControlGate, len(s.groups))
	s.groupSnapshots = make([]serverGroupSnapshot, len(s.groups))
	if cfg.Academic {
		s.controlBucket = newClientTokenBucket(cfg.ControlRate, cfg.ControlBurst, 0)
	}
	for i := range s.groups {
		s.groups[i].ControlHeapIndex, s.groups[i].ControlDeadline = -1, -1
	}
	for index, g := range cfg.Manifest.generators {
		first := (g.FirstID - 1) / cfg.Manifest.run.V
		count := g.Count / cfg.Manifest.run.V
		for j := uint32(0); j < count; j++ {
			s.groups[first+j].GeneratorIndex = uint32(index)
		}
	}
	for i := range s.workers {
		owned := 0
		if i < len(s.groups) {
			owned = 1 + (len(s.groups)-1-i)/cfg.Workers
		}
		s.workers[i] = serverWorker{server: s, index: i, writerIndex: i % cfg.ServerWriterCount, input: newQueue[inbound](cfg.InputCapacity), high: newQueue[outbound](cfg.HighCapacity),
			data: newQueue[outbound](cfg.DataCapacity), pending: newQueue[uint32](owned * int(cfg.Manifest.run.V)), wake: make(chan struct{}, 1), startCursor: i}
		w := &s.workers[i]
		controlCap, confirmCap := cfg.ControlCapacity, cfg.ConfirmationCapacity
		if !cfg.Academic {
			controlCap, confirmCap = 1, 1
		}
		w.control = newQueue[serverControlOutbound](controlCap)
		w.confirmations = newQueue[serverControlConfirmation](confirmCap)
		w.pendingGroups = newQueue[uint32](owned)
		w.controlDeadlines = newServerControlDeadlineHeap(s.groups, owned)
	}
	return s, nil
}

// ListenServer validates/allocates before binding one ingress and optional
// TX-only egress sockets. Only the ingress has a reader. It starts
// W workers, one reader, K writers and one supervisor (W+K+2 goroutines).
// Academic mode adds reliable START/GROUP_DONE control to those same workers;
// Academic=false preserves the isolated routing API used by the approved phases.
func ListenServer(ctx context.Context, cfg ServerConfig) (*Server, error) {
	if ctx == nil {
		return nil, ErrServerConfig
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s, err := newServer(cfg)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", cfg.Address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	s.conn = conn
	s.address = conn.LocalAddr().(*net.UDPAddr).AddrPort()
	fail := func(err error) (*Server, error) { s.closeSockets(); return nil, err }
	if cfg.ReadBuffer > 0 {
		if err = conn.SetReadBuffer(cfg.ReadBuffer); err != nil {
			return fail(fmt.Errorf("read buffer: %w", err))
		}
	}
	if cfg.WriteBuffer > 0 {
		if err = conn.SetWriteBuffer(cfg.WriteBuffer); err != nil {
			return fail(fmt.Errorf("write buffer: %w", err))
		}
	}
	if err = s.openEgress(); err != nil {
		return fail(err)
	}
	if cfg.GOMAXPROCS > 0 {
		runtime.GOMAXPROCS(cfg.GOMAXPROCS)
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.wg.Add(len(s.workers) + len(s.writers) + 1)
	for i := range s.workers {
		go s.workers[i].run()
	}
	go s.reader()
	for i := range s.writers {
		go s.writers[i].run()
	}
	go s.supervise()
	return s, nil
}
func (s *Server) Addr() netip.AddrPort { return s.address }
func (s *Server) Wait() error          { <-s.done; return s.runErr }
func (s *Server) Close() error         { s.cancel(); return s.Wait() }
func (s *Server) reportFatal(err error) {
	select {
	case s.fatal <- err:
	default:
	}
}
func notify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (s *Server) supervise() {
	select {
	case <-s.ctx.Done():
	case s.runErr = <-s.fatal:
	}
	s.cancel()
	s.closeSockets()
	s.wg.Wait()
	// Simulation state has already been cleaned up by its owner workers.
	for i := range s.workers {
		w := &s.workers[i]
		n := w.input.discardAll() + w.high.discardAll() + w.data.discardAll() + w.control.discardAll()
		s.counters.queueDrops.Add(n)
		w.pending.discardAll() // references, not additional datagrams
		w.pendingGroups.discardAll()
		w.confirmations.discardAll()
	}
	close(s.done)
}

// WorkerForGroup checks before subtracting/indexing; registration uses its derived group.
func (s *Server) WorkerForGroup(groupID uint32) (int, error) {
	if groupID == 0 || uint64(groupID) > uint64(len(s.groups)) {
		return 0, ErrIdentity
	}
	return int((groupID - 1) % uint32(len(s.workers))), nil
}
func (s *Server) generatorIndex(id uint32) int {
	g := s.cfg.Manifest.generators
	lo, hi := 0, len(g)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if g[mid].GeneratorID < id {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo == len(g) || g[lo].GeneratorID != id {
		return -1
	}
	return lo
}
func (s *Server) invalid(err error) error {
	s.counters.invalid.Add(1)
	if err == ErrEndpoint {
		s.counters.endpointRejects.Add(1)
	}
	if err == ErrConflict {
		s.counters.conflicts.Add(1)
	}
	return err
}
func normalizeEndpoint(a netip.AddrPort) netip.AddrPort {
	return netip.AddrPortFrom(a.Addr().Unmap(), a.Port())
}

// decodeDispatch is reader work only: no access to mutable simulation state.
func (s *Server) decodeDispatch(data []byte, source netip.AddrPort) error {
	s.counters.packetsReceived.Add(1)
	p, err := Decode(data)
	if err != nil {
		return s.invalid(err)
	}
	if err = s.cfg.Manifest.Validate(&p); err != nil {
		return s.invalid(err)
	}
	if p.Flags != 0 {
		return s.invalid(ErrFlags)
	}
	if !source.IsValid() || source.Port() == 0 {
		return s.invalid(ErrEndpoint)
	}
	source = normalizeEndpoint(source)
	var id, group uint32
	switch p.Type {
	case REGISTER:
		s.counters.registerReceived.Add(1)
		id, group, err = s.cfg.Manifest.RegistrationIdentity(p.GeneratorID, p.RegistrationNonce)
	case POSITION:
		s.counters.positionReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	case RETRY_POSITION:
		s.counters.retryReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	case ACK:
		s.counters.ackReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	case START_ACK:
		if !s.cfg.Academic {
			return s.invalid(ErrType)
		}
		s.counters.startAckReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	case METRICS:
		if !s.cfg.Academic {
			return s.invalid(ErrType)
		}
		s.counters.metricsReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	case GROUP_DONE_ACK:
		if !s.cfg.Academic {
			return s.invalid(ErrType)
		}
		s.counters.groupDoneAckReceived.Add(1)
		id, group = p.OriginID, p.GroupID
	default:
		return s.invalid(ErrType)
	}
	if err != nil {
		return s.invalid(err)
	}
	index, err := s.WorkerForGroup(group)
	if err != nil {
		return s.invalid(err)
	}
	gi := s.generatorIndex(p.GeneratorID)
	if gi < 0 {
		return s.invalid(ErrIdentity)
	}
	w := &s.workers[index]
	if !w.input.push(inbound{p, source, gi, id, group}, true) {
		s.counters.queueDrops.Add(1)
		return ErrQueueFull
	}
	notify(w.wake)
	return nil
}
func (s *Server) reader() {
	defer s.wg.Done()
	var buf [MaxPacketSize + 1]byte // detects oversize UDP datagrams even if truncated by the OS
	for {
		n, addr, err := s.conn.ReadFromUDPAddrPort(buf[:])
		if err != nil {
			if s.ctx.Err() == nil {
				s.counters.socketErrors.Add(1)
				s.reportFatal(err)
			}
			return
		}
		s.decodeDispatch(buf[:n], addr)
	}
}

func (w *serverWorker) client(id uint32) (*ServerClientState, error) {
	if id == 0 || uint64(id) > uint64(len(w.server.clients)) {
		return nil, ErrIdentity
	}
	group := 1 + (id-1)/w.server.cfg.Manifest.run.V
	owner, err := w.server.WorkerForGroup(group)
	if err != nil || owner != w.index {
		return nil, ErrIdentity
	}
	return &w.server.clients[id-1], nil
}
func (w *serverWorker) bindEndpoint(index int, addr netip.AddrPort) error {
	e := &w.server.endpoints[index]
	current := e.address.Load()
	if current == nil {
		candidate := new(netip.AddrPort)
		*candidate = addr
		if e.address.CompareAndSwap(nil, candidate) {
			return nil
		}
		current = e.address.Load()
	}
	if *current != addr {
		return ErrEndpoint
	}
	return nil
}
func (w *serverWorker) schedule(id uint32, c *ServerClientState) {
	if c.Flags&clientScheduled != 0 {
		return
	}
	if c.Flags&(registerReplyPending|metricsReplyPending) == 0 && c.PendingPositionMask == 0 && c.PendingRetryMask == 0 {
		return
	}
	if !w.pending.push(id, false) { // impossible with one entry per owned client
		w.server.reportFatal(errors.New("server: pending ownership invariant"))
		return
	}
	c.Flags |= clientScheduled
}
func samePosition(a, b PositionPayload) bool {
	return math.Float32bits(a.X) == math.Float32bits(b.X) && math.Float32bits(a.Y) == math.Float32bits(b.Y) && math.Float32bits(a.Z) == math.Float32bits(b.Z)
}

// handle is called only by the owner worker, after decodeDispatch validation.
func (w *serverWorker) handle(e inbound) error {
	s := w.server
	p := &e.Packet
	if e.GeneratorIndex < 0 || e.GeneratorIndex >= len(s.endpoints) {
		return s.invalid(ErrIdentity)
	}
	owner, err := s.WorkerForGroup(e.GroupID)
	if err != nil || owner != w.index {
		return s.invalid(ErrIdentity)
	}
	c, err := w.client(e.ClientID)
	if err != nil {
		return s.invalid(err)
	}
	if p.Type == REGISTER {
		if err = w.bindEndpoint(e.GeneratorIndex, e.Source); err != nil {
			return s.invalid(err)
		}
		if c.Flags&clientRegistered != 0 {
			if c.RegisterMessageID != p.MessageID {
				return s.invalid(ErrConflict)
			}
			s.counters.duplicateRegisters.Add(1)
		} else {
			c.RegisterMessageID = p.MessageID
			c.Flags |= clientRegistered
			c.ExpectedMask, _ = ExpectedMask(e.ClientID, s.cfg.Manifest.run.V)
			s.groups[e.GroupID-1].RegisteredMask |= uint32(1) << ((e.ClientID - 1) % s.cfg.Manifest.run.V)
			registered := s.counters.registered.Add(1)
			if s.cfg.Academic {
				w.publishGroup(e.GroupID - 1)
				if registered == uint64(s.cfg.Manifest.run.N) {
					s.simulationStartNS.Store(s.now())
					for i := range s.workers {
						notify(s.workers[i].wake)
					}
				}
			}
		}
		c.Flags |= registerReplyPending
		w.schedule(e.ClientID, c)
		return nil
	}
	endpoint := s.endpoints[e.GeneratorIndex].address.Load()
	if endpoint == nil || *endpoint != e.Source {
		return s.invalid(ErrEndpoint)
	}
	if c.Flags&clientRegistered == 0 || s.counters.registered.Load() != uint64(s.cfg.Manifest.run.N) {
		return s.invalid(ErrNotRegistered)
	}
	switch p.Type {
	case START_ACK, GROUP_DONE_ACK:
		return w.handleControlACK(p)
	case METRICS:
		return w.handleMetrics(p, c)
	case POSITION, RETRY_POSITION:
		if s.cfg.Academic {
			g := &s.groups[e.GroupID-1]
			if g.ControlPhase == ServerGroupIdle || g.ControlPhase == ServerGroupFailed {
				return s.invalid(ErrProgress)
			}
			// Stable academic IDs identify old datagrams without storing a history.
			if p.Iteration < c.Iteration && p.MessageID == c.RegisterMessageID+uint64(p.Iteration) {
				s.counters.stalePackets.Add(1)
				return nil
			}
		}
		if c.Iteration == 0 || p.Iteration != c.Iteration {
			pendingBlocks := !s.cfg.Academic && (c.PendingPositionMask != 0 || c.PendingRetryMask != 0)
			if p.Iteration != c.Iteration+1 || c.Iteration != 0 && (c.SeenACKMask != c.ExpectedMask || pendingBlocks) {
				return s.invalid(ErrProgress)
			}
			// The origin's next operation proves it received the previous ACK set.
			// Cancel unsent old fanout; already-enqueued immutable jobs remain safe.
			c.PendingPositionMask, c.PendingRetryMask = 0, 0
			c.Iteration = p.Iteration
			c.MessageID = p.MessageID
			c.Position = p.Position
			c.SeenACKMask, c.RetryMask, c.RetryCounter = 0, 0, 0
		} else if p.MessageID != c.MessageID || !samePosition(p.Position, c.Position) {
			return s.invalid(ErrConflict)
		}
		if p.Type == POSITION {
			c.PendingPositionMask |= c.ExpectedMask
		} else {
			if p.RetryCounter < c.RetryCounter {
				if s.cfg.Academic {
					s.counters.stalePackets.Add(1)
					return nil
				}
				return s.invalid(ErrRetry)
			}
			if p.RetryCounter == c.RetryCounter && p.MissingMask != c.RetryMask {
				return s.invalid(ErrConflict)
			}
			// The newest retry snapshot supersedes only unsent intentions. Jobs already
			// queued remain immutable. Every resulting destination was actually requested.
			c.RetryCounter = p.RetryCounter
			c.RetryMask = p.MissingMask
			c.PendingRetryMask = p.MissingMask
		}
		g := &s.groups[e.GroupID-1]
		slot := uint32(1) << ((e.ClientID - 1) % s.cfg.Manifest.run.V)
		g.PositionStartedMask |= slot
		if c.Iteration == s.cfg.Manifest.run.I && c.SeenACKMask == c.ExpectedMask {
			g.FinalACKMask |= slot
		}
		w.schedule(e.ClientID, c)
		return nil
	case ACK:
		origin, err := w.client(p.TargetID)
		if err != nil {
			return s.invalid(err)
		}
		if origin.Flags&clientRegistered == 0 {
			return s.invalid(ErrNotRegistered)
		}
		if s.cfg.Academic && p.Iteration < origin.Iteration && p.MessageID == origin.RegisterMessageID+uint64(p.Iteration) {
			s.counters.stalePackets.Add(1)
			return nil
		}
		if origin.Iteration != p.Iteration {
			return s.invalid(ErrIteration)
		}
		if origin.MessageID != p.MessageID {
			return s.invalid(ErrMessage)
		}
		bit := uint32(1) << ((p.OriginID - 1) % s.cfg.Manifest.run.V)
		if origin.ExpectedMask&bit == 0 {
			return s.invalid(ErrIdentity)
		}
		if origin.SeenACKMask&bit != 0 {
			s.counters.duplicateACKs.Add(1)
		}
		origin.SeenACKMask |= bit
		if origin.Iteration == s.cfg.Manifest.run.I && origin.SeenACKMask == origin.ExpectedMask {
			s.groups[e.GroupID-1].FinalACKMask |= uint32(1) << ((p.TargetID - 1) % s.cfg.Manifest.run.V)
		}
		forwarded := *p
		forwarded.Flags = ServerToClient
		if !w.high.push(outbound{forwarded, e.GeneratorIndex}, true) {
			s.counters.queueDrops.Add(1)
			return ErrQueueFull
		}
		w.notifyWriter()
		return nil
	default:
		return s.invalid(ErrType)
	}
}

func (w *serverWorker) controlReply(id uint32, c *ServerClientState) outbound {
	s := w.server
	v := s.cfg.Manifest.run.V
	group := 1 + (id-1)/v
	gi := int(s.groups[group-1].GeneratorIndex)
	g := s.cfg.Manifest.generators[gi]
	return outbound{Packet{Header: Header{Type: REGISTER_ACK, Flags: ServerToClient, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: g.GeneratorID, GroupID: group, TargetID: id, MessageID: c.RegisterMessageID,
		RegistrationNonce: uint64(id-g.FirstID) + 1}, Registration: RegistrationPayload{v, s.cfg.Manifest.run.I}}, gi}
}
func (w *serverWorker) positionJob(id uint32, c *ServerClientState, target uint32, retry bool) outbound {
	s := w.server
	v := s.cfg.Manifest.run.V
	group := 1 + (id-1)/v
	gi := int(s.groups[group-1].GeneratorIndex)
	p := Packet{Header: Header{Type: POSITION, Flags: ServerToClient, RunID: s.cfg.Manifest.run.RunID,
		GeneratorID: s.cfg.Manifest.generators[gi].GeneratorID, GroupID: group, OriginID: id, TargetID: target,
		Iteration: c.Iteration, MessageID: c.MessageID}, Position: c.Position}
	if retry {
		p.Type = RETRY_POSITION
		p.MissingMask = c.RetryMask
		p.RetryCounter = c.RetryCounter
	}
	return outbound{p, gi}
}

// flushPending takes a finite snapshot of the pending ring, so a blocked client
// cannot busy-spin. Writer dequeue notifications wake the owner to try again.
func (w *serverWorker) flushPending() bool {
	count := w.pending.stats().Depth
	budget := w.server.cfg.WorkBudget
	progress := false
	for turn := 0; turn < count && budget > 0; turn++ {
		id, ok := w.pending.pop()
		if !ok {
			break
		}
		c, err := w.client(id)
		if err != nil {
			w.server.reportFatal(err)
			break
		}
		c.Flags &^= clientScheduled
		// Bound one client's service to 8 writes, then rotate through other clients.
		quota := 8
		if quota > budget {
			quota = budget
		}
		if c.Flags&registerReplyPending != 0 {
			quota--
			budget--
			if w.high.push(w.controlReply(id, c), false) {
				c.Flags &^= registerReplyPending
				progress = true
				w.notifyWriter()
			}
		}
		if c.Flags&metricsReplyPending != 0 && quota > 0 {
			quota--
			budget--
			if w.high.push(w.metricsReply(id, c), false) {
				c.Flags &^= metricsReplyPending
				progress = true
				w.notifyWriter()
			}
		}
		for kind := 0; kind < 2 && quota > 0; kind++ {
			mask := &c.PendingPositionMask
			if kind == 1 {
				mask = &c.PendingRetryMask
			}
			for *mask != 0 && quota > 0 {
				bit := uint32(1) << bits.TrailingZeros32(*mask)
				target := ((id-1)/w.server.cfg.Manifest.run.V)*w.server.cfg.Manifest.run.V + uint32(bits.TrailingZeros32(bit)) + 1
				quota--
				budget--
				if !w.data.push(w.positionJob(id, c, target, kind == 1), false) {
					break
				}
				*mask &^= bit
				progress = true
				w.notifyWriter()
			}
		}
		w.schedule(id, c)
	}
	return progress
}
func (w *serverWorker) run() {
	defer w.server.wg.Done()
	defer w.finish()
	var timer *time.Timer
	if w.server.cfg.Academic {
		timer = time.NewTimer(time.Hour)
		stopServerTimer(timer)
		w.server.activeControlTimers.Add(1)
		defer func() { stopServerTimer(timer); w.server.activeControlTimers.Add(-1) }()
	}
	for {
		if w.server.ctx.Err() != nil {
			return
		}
		for i := 0; i < w.server.cfg.WorkBudget; i++ {
			tx, ok := w.confirmations.pop()
			if !ok {
				break
			}
			if err := w.handleControlConfirmation(tx); err != nil {
				w.server.reportFatal(err)
				return
			}
			w.notifyWriter()
		}
		for i := 0; i < w.server.cfg.WorkBudget; i++ {
			e, ok := w.input.pop()
			if !ok {
				break
			}
			w.handle(e)
		}
		progress := w.stageStarts()
		if w.server.cfg.Academic && w.processControlTimeouts(w.server.now()) {
			progress = true
		}
		if w.flushPending() {
			progress = true
		}
		if w.server.cfg.Academic && w.flushControlPending() {
			progress = true
		}
		if w.input.stats().Depth > 0 || w.confirmations.stats().Depth > 0 || progress {
			continue
		}
		var timerC <-chan time.Time
		if timer != nil {
			if _, deadline, ok := w.controlDeadlines.peek(); ok {
				delay := deadline - w.server.now()
				if delay < 0 {
					delay = 0
				}
				timer.Reset(time.Duration(delay))
				timerC = timer.C
			}
		}
		select {
		case <-w.server.ctx.Done():
			return
		case <-w.wake:
		case <-timerC:
		}
		if timer != nil {
			stopServerTimer(timer)
		}
	}
}

// Even shutdown preserves ownership: the supervisor never writes client state.
func (w *serverWorker) finish() {
	s := w.server
	v := int(s.cfg.Manifest.run.V)
	for group := w.index; group < len(s.groups); group += len(s.workers) {
		g := &s.groups[group]
		s.counters.shutdownIntentions.Add(uint64(bits.OnesCount32(g.ControlPendingMask)))
		w.cancelGroupRound(uint32(group))
		g.ControlScheduled = false
		for id := group * v; id < (group+1)*v; id++ {
			c := &s.clients[id]
			count := bits.OnesCount32(c.PendingPositionMask) + bits.OnesCount32(c.PendingRetryMask)
			if c.Flags&registerReplyPending != 0 {
				count++
			}
			if c.Flags&metricsReplyPending != 0 {
				count++
			}
			s.counters.shutdownIntentions.Add(uint64(count))
			c.PendingPositionMask, c.PendingRetryMask = 0, 0
			c.Flags &^= registerReplyPending | metricsReplyPending | clientScheduled
		}
	}
}

func (wr *serverWriter) popClass(high bool) (outbound, bool) {
	s := wr.server
	cursor := &wr.dataCursor
	if high {
		cursor = &wr.highCursor
	}
	for n := 0; n < len(wr.workers); n++ {
		i := (*cursor + n) % len(wr.workers)
		w := &s.workers[wr.workers[i]]
		q := w.data
		if high {
			q = w.high
		}
		if job, ok := q.pop(); ok {
			*cursor = (i + 1) % len(wr.workers)
			notify(w.wake)
			return job, true
		}
	}
	return outbound{}, false
}
func (wr *serverWriter) takeOutbound() (outbound, bool) {
	if wr.highStreak >= wr.server.cfg.HighBurst {
		if job, ok := wr.popClass(false); ok {
			wr.highStreak = 0
			return job, true
		}
	}
	if job, ok := wr.popClass(true); ok {
		if wr.highStreak < wr.server.cfg.HighBurst {
			wr.highStreak++
		}
		return job, true
	}
	if job, ok := wr.popClass(false); ok {
		wr.highStreak = 0
		return job, true
	}
	return outbound{}, false
}
func (wr *serverWriter) run() {
	s := wr.server
	defer s.wg.Done()
	defer wr.publishCounts()
	var buf [MaxPacketSize]byte
	var timer *time.Timer
	if s.cfg.Academic {
		timer = time.NewTimer(time.Hour)
		stopServerTimer(timer)
		s.activeWriterTimers.Add(1)
		defer func() { stopServerTimer(timer); s.activeWriterTimers.Add(-1) }()
	}
	for {
		if s.ctx.Err() != nil {
			return
		}
		var job outbound
		var control serverControlOutbound
		var isControl, ok bool
		var wait int64
		if s.cfg.Academic {
			job, control, isControl, wait, ok = wr.takeServerOutbound(s.now())
		} else {
			job, ok = wr.takeOutbound()
		}
		if !ok {
			wr.publishCounts()
			var timerC <-chan time.Time
			if timer != nil && wait > 0 {
				delay := wait - s.now()
				if delay < 0 {
					delay = 0
				}
				timer.Reset(time.Duration(delay))
				timerC = timer.C
			}
			select {
			case <-s.ctx.Done():
				return
			case <-wr.wake:
			case <-timerC:
			}
			if timer != nil {
				stopServerTimer(timer)
			}
			continue
		}
		if isControl && control.GroupID == 0 {
			continue
		} // cancelled-head cleanup, no lease
		if !wr.writeOutbound(buf[:], job, control, isControl) {
			return
		}
	}
}

// CONTROL holds only its pacing lease. HIGH/DATA use their writer's assigned
// descriptor. Deadlines are shared with one socket, independent with two/three.
// Cancellation closes ingress/egress and joins every writer.
func (wr *serverWriter) writeOutbound(buf []byte, job outbound, control serverControlOutbound, isControl bool) bool {
	s := wr.server
	if isControl {
		defer s.releaseControlLease()
		if s.controlCancelled(control) {
			return true
		}
		job.Packet, job.GeneratorIndex = control.Packet, control.GeneratorIndex
	}
	n, err := Encode(buf, &job.Packet)
	if err != nil {
		s.counters.sendDrops.Add(1)
		s.reportFatal(fmt.Errorf("outbound invariant: %w", err))
		return false
	}
	endpoint := s.endpoints[job.GeneratorIndex].address.Load()
	if endpoint == nil {
		s.counters.sendDrops.Add(1)
		s.reportFatal(ErrEndpoint)
		return false
	}
	conn := wr.connection()
	err = conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	if isControl && s.controlCancelled(control) {
		s.counters.cancelledControls.Add(1)
		return true
	}
	if isControl && !s.controlBucket.take(s.now()) {
		s.reportFatal(ErrProgress)
		return false
	}
	sentAt := s.now()
	if err == nil {
		var sent int
		sent, err = conn.WriteToUDPAddrPort(buf[:n], *endpoint)
		if err == nil && sent != n {
			err = io.ErrShortWrite
		}
	}
	if err != nil {
		s.counters.sendDrops.Add(1)
		if s.ctx.Err() != nil {
			return false
		}
		s.counters.socketErrors.Add(1)
		wr.socketErrors++
		if isControl {
			s.publishControlConfirmation(control, sentAt, true)
		}
		return true
	}
	// CONTROL receipt follows the effective write, including socket lock wait.
	if isControl {
		sentAt = s.now()
	}
	if isControl && !s.publishControlConfirmation(control, sentAt, false) {
		return false
	}
	s.counters.packetsSent.Add(1)
	wr.recordWrite(job.Packet.Type != POSITION && job.Packet.Type != RETRY_POSITION && !isControl, isControl)
	switch job.Packet.Type {
	case REGISTER_ACK:
		s.counters.registerAckSent.Add(1)
	case METRICS_ACK:
		s.counters.metricsAckSent.Add(1)
	case START:
		s.counters.startSent.Add(1)
		if control.Attempt > 0 {
			s.counters.startRetries.Add(1)
		}
	case GROUP_DONE:
		s.counters.groupDoneSent.Add(1)
		if control.Attempt > 0 {
			s.counters.groupDoneRetries.Add(1)
		}
	case POSITION:
		s.counters.positionForwarded.Add(1)
	case RETRY_POSITION:
		s.counters.retryForwarded.Add(1)
	case ACK:
		s.counters.ackForwarded.Add(1)
	}
	return true
}

// Stats allocates only for the on-demand snapshot, never in packet processing.
// Individual counters/queues are race-safe; the complete snapshot is not atomic.
func (s *Server) Stats() ServerStats {
	c := &s.counters
	st := ServerStats{PacketsReceived: c.packetsReceived.Load(), PacketsSent: c.packetsSent.Load(),
		RegisterReceived: c.registerReceived.Load(), RegisterAckSent: c.registerAckSent.Load(),
		PositionReceived: c.positionReceived.Load(), PositionForwarded: c.positionForwarded.Load(),
		RetryPositionReceived: c.retryReceived.Load(), RetryPositionForwarded: c.retryForwarded.Load(),
		AckReceived: c.ackReceived.Load(), AckForwarded: c.ackForwarded.Load(), InvalidPackets: c.invalid.Load(),
		DuplicateRegisters: c.duplicateRegisters.Load(), QueueDrops: c.queueDrops.Load(), SocketErrors: c.socketErrors.Load(),
		RegisteredClients: c.registered.Load(), EndpointRejects: c.endpointRejects.Load(), Conflicts: c.conflicts.Load(),
		DuplicateACKs: c.duplicateACKs.Load(), SendDrops: c.sendDrops.Load(), ShutdownIntentions: c.shutdownIntentions.Load(),
		StartSent: c.startSent.Load(), StartAckReceived: c.startAckReceived.Load(), StartRetries: c.startRetries.Load(),
		StartedClients: c.startedClients.Load(), StartedGroups: c.startedGroups.Load(),
		MetricsReceived: c.metricsReceived.Load(), MetricsAckSent: c.metricsAckSent.Load(), MetricsAccepted: c.metricsAccepted.Load(), DuplicateMetrics: c.duplicateMetrics.Load(),
		GroupDoneSent: c.groupDoneSent.Load(), GroupDoneAckReceived: c.groupDoneAckReceived.Load(), GroupDoneRetries: c.groupDoneRetries.Load(), DoneClients: c.doneClients.Load(),
		CompletedGroups: c.completedGroups.Load(), ControlTimeouts: c.controlTimeouts.Load(), ControlRetryExhausted: c.controlRetryExhausted.Load(), ControlFailures: c.controlFailures.Load(),
		DuplicateStartACKs: c.duplicateStartACKs.Load(), DuplicateGroupDoneACKs: c.duplicateGroupDoneACKs.Load(), StalePackets: c.stalePackets.Load(),
		CancelledControls: c.cancelledControls.Load(), StaleControlConfirmations: c.staleControlConfirmations.Load(),
		ControlHeapCurrent: s.controlHeapCurrent.Load(), ControlHeapMaximum: s.controlHeapMaximum.Load(), ControlTimers: s.activeControlTimers.Load(), WriterTimers: s.activeWriterTimers.Load(),
		SimulationComplete: s.simulationOutcome.Load() == 1, N: s.cfg.Manifest.run.N, V: s.cfg.Manifest.run.V, I: s.cfg.Manifest.run.I,
		Queues: make([]WorkerQueueStats, len(s.workers)), Writers: s.writerStats()}
	st.Egress = s.egressStats(st.Writers)
	st.RegistrationBarrierReached = st.RegisteredClients == uint64(s.cfg.Manifest.run.N)
	var sumLo, sumHi, accepted uint64
	for i := range s.groups {
		g := s.readGroup(i)
		var carry uint64
		sumLo, carry = bits.Add64(sumLo, g.SumLo, 0)
		sumHi, _ = bits.Add64(sumHi, g.SumHi, carry)
		accepted += uint64(g.MetricsClients)
	}
	if accepted != 0 {
		st.MeanRTTNS, _ = bits.Div64(sumHi, sumLo, accepted)
	}
	if start := s.simulationStartNS.Load(); start >= 0 {
		end := s.simulationEndNS.Load()
		if end < 0 {
			end = s.now()
		}
		st.SimulationDuration = time.Duration(end - start)
	}
	for i := range s.workers {
		w := &s.workers[i]
		st.Queues[i] = WorkerQueueStats{Input: w.input.stats(), High: w.high.stats(), Data: w.data.stats(), PendingClients: w.pending.stats(),
			Control: w.control.stats(), Confirmations: w.confirmations.stats(), PendingGroups: w.pendingGroups.stats()}
	}
	return st
}
