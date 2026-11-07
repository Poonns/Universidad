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
	ErrClientConfig    = errors.New("clients: invalid configuration")
	ErrClientBusy      = errors.New("clients: previous operation is not complete")
	ErrClientPhase     = errors.New("clients: unexpected state")
	ErrClientTimestamp = errors.New("clients: invalid monotonic timestamp")
	ErrClientWrite     = errors.New("clients: socket write failed")
	ErrRetryExhausted  = errors.New("clients: retry attempts exhausted")
)

type ClientConfig struct {
	Academic                                                   bool // full START -> I iterations -> METRICS -> GROUP_DONE workflow
	OptimizeDataAdmission                                      bool // experimental, false retains historical pending admission
	Manifest                                                   *Manifest
	GeneratorID                                                uint32
	Server                                                     netip.AddrPort
	ServerEgressEndpoints                                      []netip.AddrPort // explicit RX sources only; nil preserves canonical-only validation
	Address                                                    string
	Workers, GOMAXPROCS                                        int
	InputCapacity, HighCapacity, ControlCapacity, DataCapacity int
	ConfirmationCapacity, CommandCapacity                      int
	WorkBudget, HighBurst, AckBurst                            int
	RegisterWindow, RegisterRate, PositionRate                 int
	PositionWindow, PositionBurst, RetryRate, RetryBurst       int
	RetryCapacity                                              int
	InitialRTO, MaxRTO                                         time.Duration
	MaxSends                                                   uint16
	JitterPercent                                              int
	JitterSeed                                                 uint64
	ReadBuffer, WriteBuffer                                    int
	WriteTimeout                                               time.Duration
}

func DefaultClientConfig(m *Manifest, id uint32, server netip.AddrPort) ClientConfig {
	return ClientConfig{Academic: true, Manifest: m, GeneratorID: id, Server: server, Address: "127.0.0.1:0",
		Workers: 6, GOMAXPROCS: 6, InputCapacity: 4096, HighCapacity: 1024, ControlCapacity: 128, DataCapacity: 2048,
		ConfirmationCapacity: 256, CommandCapacity: 256, WorkBudget: 64, HighBurst: 8, AckBurst: 4,
		RegisterWindow: 128, RegisterRate: 1000, PositionRate: 1000, PositionWindow: 128, PositionBurst: 1,
		RetryCapacity: 512, RetryRate: 100, RetryBurst: 1, InitialRTO: time.Second, MaxRTO: 5 * time.Second, MaxSends: 12, JitterPercent: 20, JitterSeed: 42,
		ReadBuffer: 4 * 1024 * 1024, WriteBuffer: 4 * 1024 * 1024, WriteTimeout: time.Second}
}

type ClientPhase uint8

const (
	ClientNew ClientPhase = iota
	ClientRegisterQueued
	ClientRegisterWait
	ClientReady
	ClientPositionQueued
	ClientPositionWait
	ClientComplete
	ClientFailed
	ClientStopped
	ClientWaitStart
	ClientMetricsQueued
	ClientMetricsWait
	ClientWaitGroupDone
	ClientDone
)
const (
	clientRegisterIntent uint16 = 1 << iota
	clientPositionIntent
	clientPendingIndex
	clientRegisterTxKnown
	clientRegisterSlotReleased
	clientPositionTxKnown
	clientRetryIntent
	clientAwaitConfirmation
	clientPositionSlotReleased
	clientAcademicStarted
	clientPendingStart
	clientMetricsIntent
	clientMetricsAccepted
	clientPendingControl
	clientPendingRetry
)

type ClientFailure uint8

const (
	ClientFailureNone ClientFailure = iota
	ClientRetryExhausted
	ClientWriteFailed
	ClientTimestampFailed
)

const clientTicketClaimed uint64 = 1 << 63

// ClientState is a compact, pointer-free value. Only its group's worker writes it.
// Nonce and ExpectedMask are initialized before launch. SentAt=-1 means unknown.
type ClientState struct {
	RegistrationNonce, MessageID                          uint64
	SentAt, LastUniqueACKAt                               int64
	RTTSum, LastRTTNS                                     uint64
	AckMask, ExpectedMask, Iteration, CompletedIterations uint32
	Position                                              PositionPayload
	Phase                                                 ClientPhase
	Registered                                            bool
	Flags                                                 uint16
	CurrentRTO, Deadline                                  int64
	SendTicket                                            uint64
	HeapIndex                                             int32
	SendCount, RetryCounter                               uint16
	Failure                                               ClientFailure
	PendingStartAttempt                                   uint16
}
type clientInbound struct {
	Packet     Packet
	ReceivedAt int64
}
type clientOutbound struct {
	Packet   Packet
	Ticket   uint64
	ClientID uint32
	Attempt  uint16
}
type clientCommand struct {
	ClientID, Iteration uint32
	Position            PositionPayload
}
type clientSendStatus uint8

const (
	clientSendOK clientSendStatus = iota
	clientSendFailed
)

type clientConfirmation struct {
	MessageID           uint64
	SentAt              int64
	Ticket              uint64
	ClientID, Iteration uint32
	Type                MessageType
	Status              clientSendStatus
	Attempt             uint16
}

type clientCounters struct {
	registered, packetsReceived, packetsSent                                                         atomic.Uint64
	registerSent, registerAckReceived, duplicateRegisterAcks                                         atomic.Uint64
	positionSent, positionReceived, ackSent, ackReceived, uniqueAcks, duplicateAcks                  atomic.Uint64
	started, completed, invalid, queueDrops, socketErrors, sendDrops                                 atomic.Uint64
	rejectedOperations, failedRegistrations, failedOperations, abortedOperations, shutdownIntentions atomic.Uint64
	confirmations, duplicateConfirmations, invalidConfirmations, rttSum, rttSamples                  atomic.Uint64
	registerRetries, positionRetries, retryNeighbors, timeouts, retryExhausted, recovered            atomic.Uint64
	admissionDelayed, retryBudgetDelayed, cancelledSends, staleConfirmations, admitted               atomic.Uint64
	startReceived, startAckSent, duplicateStarts, startedClients                                     atomic.Uint64
	metricsSent, metricsAckReceived, metricsRetries, metricsAccepted, duplicateMetricsAcks           atomic.Uint64
	groupDoneReceived, groupDoneAckSent, doneClients, duplicateGroupDone                             atomic.Uint64
	failedMetrics, controlTimeouts, controlRetryExhausted                                            atomic.Uint64
	preRegistrationData, staleACKs                                                                   atomic.Uint64
}
type ClientQueueStats struct {
	Input, High, Control, Data, Confirmations, Commands, PendingClients QueueStats
	Retry                                                               QueueStats
}
type ClientStats struct {
	LogicalClients, RegisteredClients, PacketsReceived, PacketsSent                                  uint64
	RegisterSent, RegisterAckReceived, DuplicateRegisterAcks                                         uint64
	PositionSent, PositionReceived, AckSent, AckReceived, UniqueAcks, DuplicateAcks                  uint64
	StartedOperations, CompletedOperations, InvalidPackets, QueueDrops, SocketErrors, SendDrops      uint64
	RejectedOperations, FailedRegistrations, FailedOperations, AbortedOperations, ShutdownIntentions uint64
	WriterConfirmations, DuplicateConfirmations, InvalidConfirmations, RTTSum, RTTSamples            uint64
	RegisterInFlight, MaximumRegisterInFlight                                                        int64
	RegisterRetries, PositionRetries, RetryPositionSent, RetryNeighborsRequested                     uint64
	Timeouts, RetryExhausted, RecoveredOperations, AdmissionDelayed, AdmissionRejected               uint64
	RetryBudgetDelayed, CancelledSends, StaleConfirmations, NewPositionAdmitted                      uint64
	InflightCurrent, InflightMaximum, DeadlineHeapCurrent, DeadlineHeapMaximum                       int64
	ARQTimers, WriterTimers                                                                          int64
	StartReceived, StartAckSent, DuplicateStarts, StartedClients                                     uint64
	MetricsSent, MetricsAckReceived, MetricsRetries, MetricsAccepted, DuplicateMetricsAcks           uint64
	GroupDoneReceived, GroupDoneAckSent, DoneClients, DuplicateGroupDone                             uint64
	FailedMetrics, ControlTimeouts, ControlRetryExhausted                                            uint64
	PreRegistrationData, StaleACKs                                                                   uint64
	Queues                                                                                           []ClientQueueStats
}
type clientWorker struct {
	generator                    *ClientGenerator
	index                        int
	input                        *boundedQueue[clientInbound]
	high, control, data          *boundedQueue[clientOutbound]
	confirmations                *boundedQueue[clientConfirmation]
	commands                     *boundedQueue[clientCommand]
	pending                      *boundedQueue[uint32] // local client indices, at most one per client
	pendingControl, pendingRetry int                   // owner-only conservative class counts; no extra queue
	retry                        *boundedQueue[clientOutbound]
	deadlines                    clientDeadlineHeap
	wake                         chan struct{}
}
type ClientGenerator struct {
	cfg                                                                       ClientConfig
	rangeConfig                                                               GeneratorConfig
	clients                                                                   []ClientState
	workers                                                                   []clientWorker
	counters                                                                  clientCounters
	registerInFlight, registerMaximum                                         atomic.Int64
	positionInFlight, positionMaximum, heapCurrent, heapMaximum               atomic.Int64
	activeARQTimers, activeWriterTimers                                       atomic.Int64
	sendGates                                                                 []atomic.Uint64 // writer metadata, never mutable ClientState
	conn                                                                      *net.UDPConn
	address                                                                   netip.AddrPort
	serverSources                                                             [3]netip.AddrPort // immutable, prevalidated; no per-packet map/allocation
	serverSourceCount                                                         int
	epoch                                                                     time.Time
	ctx                                                                       context.Context
	cancel                                                                    context.CancelFunc
	done                                                                      chan struct{}
	fatal                                                                     chan error
	writerWake                                                                chan struct{}
	wg                                                                        sync.WaitGroup
	runErr                                                                    error // supervisor writes before done closes
	highCursor, controlCursor, dataCursor, retryCursor, highStreak, ackStreak int   // writer-owned
	nextRegister                                                              int64
	positions, retries                                                        clientTokenBucket
	academicDone                                                              chan struct{} // one terminal signal per generator, never per client
	academicOutcome                                                           atomic.Uint32 // 0 pending, 1 done, 2+ClientFailure failed
}

func newClientGenerator(cfg ClientConfig) (*ClientGenerator, error) {
	if cfg.Manifest == nil || cfg.Manifest.run.Validate() != nil || cfg.Workers < 1 || uint64(cfg.Workers) > uint64(math.MaxUint32) ||
		cfg.GOMAXPROCS < 0 || cfg.InputCapacity < 1 || cfg.HighCapacity < 1 || cfg.ControlCapacity < 1 || cfg.DataCapacity < 1 ||
		cfg.ConfirmationCapacity < 1 || cfg.CommandCapacity < 1 || cfg.WorkBudget < 1 || cfg.HighBurst < 1 || cfg.AckBurst < 1 ||
		cfg.RegisterWindow < 1 || cfg.RegisterRate < 1 || cfg.PositionRate < 1 || cfg.RegisterRate > int(time.Second) || cfg.PositionRate > int(time.Second) ||
		cfg.PositionWindow < 1 || cfg.PositionBurst < 1 || int64(cfg.PositionBurst) > math.MaxInt64/int64(time.Second) ||
		cfg.RetryCapacity < 1 || cfg.RetryRate < 1 || cfg.RetryRate > int(time.Second) || cfg.RetryBurst < 1 || int64(cfg.RetryBurst) > math.MaxInt64/int64(time.Second) ||
		cfg.InitialRTO <= 0 || cfg.MaxRTO < cfg.InitialRTO || cfg.MaxRTO > time.Duration(math.MaxInt64/4) || cfg.MaxSends < 1 || cfg.JitterPercent < 0 || cfg.JitterPercent > 50 ||
		cfg.ReadBuffer < 0 || cfg.WriteBuffer < 0 || cfg.WriteTimeout <= 0 || !cfg.Server.IsValid() || cfg.Server.Port() == 0 || cfg.Server.Addr().IsUnspecified() {
		return nil, ErrClientConfig
	}
	gc, ok := cfg.Manifest.generator(cfg.GeneratorID)
	if !ok || uint64(gc.Count) > uint64(^uint(0)>>1) {
		return nil, ErrClientConfig
	}
	cfg.Server = normalizeEndpoint(cfg.Server)
	sources, count, err := clientServerSources(cfg.Server, cfg.ServerEgressEndpoints)
	if err != nil {
		return nil, err
	}
	if len(cfg.ServerEgressEndpoints) > 0 {
		cfg.ServerEgressEndpoints = append([]netip.AddrPort(nil), sources[:count]...)
	}
	g := &ClientGenerator{cfg: cfg, rangeConfig: gc, clients: make([]ClientState, int(gc.Count)), workers: make([]clientWorker, cfg.Workers), sendGates: make([]atomic.Uint64, int(gc.Count)),
		epoch: time.Now(), ctx: context.Background(), done: make(chan struct{}), fatal: make(chan error, 1), writerWake: make(chan struct{}, 1), academicDone: make(chan struct{})}
	g.serverSources, g.serverSourceCount = sources, count
	g.positions = newClientTokenBucket(cfg.PositionRate, cfg.PositionBurst, 0)
	g.retries = newClientTokenBucket(cfg.RetryRate, cfg.RetryBurst, 0)
	for i := range g.clients {
		c := &g.clients[i]
		c.RegistrationNonce = uint64(i) + 1
		c.ExpectedMask, _ = ExpectedMask(gc.FirstID+uint32(i), cfg.Manifest.run.V)
		c.SentAt, c.LastUniqueACKAt = -1, -1
		c.HeapIndex, c.Deadline = -1, -1
	}
	groups := int(gc.Count / cfg.Manifest.run.V)
	for i := range g.workers {
		owned := 0
		if i < groups {
			owned = 1 + (groups-1-i)/cfg.Workers
		}
		g.workers[i] = clientWorker{generator: g, index: i, input: newQueue[clientInbound](cfg.InputCapacity), high: newQueue[clientOutbound](cfg.HighCapacity),
			control: newQueue[clientOutbound](cfg.ControlCapacity), data: newQueue[clientOutbound](cfg.DataCapacity),
			confirmations: newQueue[clientConfirmation](cfg.ConfirmationCapacity), commands: newQueue[clientCommand](cfg.CommandCapacity),
			pending: newQueue[uint32](owned * int(cfg.Manifest.run.V)), retry: newQueue[clientOutbound](cfg.RetryCapacity),
			deadlines: newClientDeadlineHeap(g.clients, owned*int(cfg.Manifest.run.V)), wake: make(chan struct{}, 1)}
	}
	return g, nil
}

// ListenClients starts one socket, reader, writer, W workers and one supervisor.
// Registration is automatic and paced. Academic mode advances only after START.
func ListenClients(ctx context.Context, cfg ClientConfig) (*ClientGenerator, error) {
	if ctx == nil {
		return nil, ErrClientConfig
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g, err := newClientGenerator(cfg)
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
	fail := func(err error) (*ClientGenerator, error) { conn.Close(); return nil, err }
	if cfg.ReadBuffer > 0 {
		if err = conn.SetReadBuffer(cfg.ReadBuffer); err != nil {
			return fail(fmt.Errorf("client read buffer: %w", err))
		}
	}
	if cfg.WriteBuffer > 0 {
		if err = conn.SetWriteBuffer(cfg.WriteBuffer); err != nil {
			return fail(fmt.Errorf("client write buffer: %w", err))
		}
	}
	if cfg.GOMAXPROCS > 0 {
		runtime.GOMAXPROCS(cfg.GOMAXPROCS)
	}
	g.conn = conn
	g.address = conn.LocalAddr().(*net.UDPAddr).AddrPort()
	g.ctx, g.cancel = context.WithCancel(ctx)
	g.wg.Add(len(g.workers) + 2)
	for i := range g.workers {
		go g.workers[i].run()
	}
	go g.reader()
	go g.writer()
	go g.supervise()
	return g, nil
}
func (g *ClientGenerator) Addr() netip.AddrPort { return g.address }
func (g *ClientGenerator) Wait() error          { <-g.done; return g.runErr }
func (g *ClientGenerator) Close() error         { g.cancel(); return g.Wait() }
func (g *ClientGenerator) now() int64           { return time.Since(g.epoch).Nanoseconds() }
func (g *ClientGenerator) reportFatal(err error) {
	select {
	case g.fatal <- err:
	default:
	}
}
func (g *ClientGenerator) supervise() {
	select {
	case <-g.ctx.Done():
	case g.runErr = <-g.fatal:
	}
	g.cancel()
	g.conn.Close()
	g.wg.Wait()
	g.registerInFlight.Store(0) // cancelled capacity; no writer/owner can still update it
	g.positionInFlight.Store(0)
	for i := range g.workers {
		w := &g.workers[i]
		n := w.input.discardAll() + w.high.discardAll() + w.control.discardAll() + w.data.discardAll() + w.retry.discardAll() + w.confirmations.discardAll() + w.commands.discardAll()
		g.counters.queueDrops.Add(n)
		w.pending.discardAll()
	}
	close(g.done)
}
func (g *ClientGenerator) ClientIndex(id uint32) (int, error) {
	if id < g.rangeConfig.FirstID || uint64(id) >= uint64(g.rangeConfig.FirstID)+uint64(g.rangeConfig.Count) {
		return 0, ErrIdentity
	}
	return int(id - g.rangeConfig.FirstID), nil
}
func (g *ClientGenerator) WorkerForClient(id uint32) (int, error) {
	index, err := g.ClientIndex(id)
	if err != nil {
		return 0, err
	}
	return int((uint32(index) / g.cfg.Manifest.run.V) % uint32(len(g.workers))), nil
}

// StartPosition submits a bounded asynchronous command, not a completion promise.
// The owner validates registration/phase/sequence. Rejections are explicit in
// RejectedOperations; full queues and invalid static arguments return immediately.
func (g *ClientGenerator) StartPosition(id, iteration uint32, position PositionPayload) error {
	if err := g.ctx.Err(); err != nil {
		return err
	}
	if g.cfg.Academic {
		return ErrClientPhase
	}
	owner, err := g.WorkerForClient(id)
	if err != nil {
		return err
	}
	if iteration == 0 || iteration > g.cfg.Manifest.run.I {
		return ErrIteration
	}
	if !finite(position.X) || !finite(position.Y) || !finite(position.Z) {
		return ErrPayload
	}
	w := &g.workers[owner]
	if err := g.enqueueCommand(w.commands, clientCommand{id, iteration, position}); err != nil {
		return err
	}
	notify(w.wake)
	return nil
}

// Check cancellation under the SAME local queue lock used by shutdown discard.
// A command either precedes that discard or observes cancellation; it cannot
// repopulate a queue after the supervisor has joined workers and emptied it.
func (g *ClientGenerator) enqueueCommand(q *boundedQueue[clientCommand], cmd clientCommand) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := g.ctx.Err(); err != nil {
		return err
	}
	if q.size == len(q.slots) {
		q.full++
		q.drops++
		g.counters.queueDrops.Add(1)
		return ErrQueueFull
	}
	q.slots[(q.head+q.size)%len(q.slots)] = cmd
	q.size++
	if q.size > q.maximum {
		q.maximum = q.size
	}
	return nil
}
func (g *ClientGenerator) invalidPacket(err error) error { g.counters.invalid.Add(1); return err }
func (g *ClientGenerator) decodeDispatch(data []byte, source netip.AddrPort, receivedAt int64) error {
	g.counters.packetsReceived.Add(1)
	if !g.acceptsServerSource(source) {
		return g.invalidPacket(ErrEndpoint)
	}
	p, err := Decode(data)
	if err != nil {
		return g.invalidPacket(err)
	}
	if err = g.cfg.Manifest.Validate(&p); err != nil {
		return g.invalidPacket(err)
	}
	if p.Flags != ServerToClient {
		return g.invalidPacket(ErrFlags)
	}
	if p.GeneratorID != g.cfg.GeneratorID {
		return g.invalidPacket(ErrIdentity)
	}
	if receivedAt < 0 {
		return g.invalidPacket(ErrClientTimestamp)
	}
	switch p.Type {
	case REGISTER_ACK:
		g.counters.registerAckReceived.Add(1)
	case POSITION, RETRY_POSITION:
		g.counters.positionReceived.Add(1)
	case ACK:
		g.counters.ackReceived.Add(1)
	case START:
		if !g.cfg.Academic {
			return g.invalidPacket(ErrType)
		}
		g.counters.startReceived.Add(1)
	case METRICS_ACK:
		if !g.cfg.Academic {
			return g.invalidPacket(ErrType)
		}
		g.counters.metricsAckReceived.Add(1)
	case GROUP_DONE:
		if !g.cfg.Academic {
			return g.invalidPacket(ErrType)
		}
		g.counters.groupDoneReceived.Add(1)
	default:
		return g.invalidPacket(ErrType)
	}
	owner, err := g.WorkerForClient(p.TargetID)
	if err != nil {
		return g.invalidPacket(err)
	}
	w := &g.workers[owner]
	if !w.input.push(clientInbound{p, receivedAt}, true) {
		g.counters.queueDrops.Add(1)
		return ErrQueueFull
	}
	notify(w.wake)
	return nil
}
func (g *ClientGenerator) reader() {
	defer g.wg.Done()
	var buf [MaxPacketSize + 1]byte
	for {
		n, addr, err := g.conn.ReadFromUDPAddrPort(buf[:])
		at := g.now()
		if err != nil {
			if g.ctx.Err() == nil {
				g.counters.socketErrors.Add(1)
				g.reportFatal(err)
			}
			return
		}
		g.decodeDispatch(buf[:n], addr, at)
	}
}
func (w *clientWorker) client(id uint32) (*ClientState, error) {
	g := w.generator
	index, err := g.ClientIndex(id)
	if err != nil {
		return nil, err
	}
	owner := int((uint32(index) / g.cfg.Manifest.run.V) % uint32(len(g.workers)))
	if owner != w.index {
		return nil, ErrIdentity
	}
	return &g.clients[index], nil
}
func (w *clientWorker) schedule(id uint32, c *ClientState) {
	if c.Flags&clientPendingIndex != 0 {
		if w.generator.cfg.OptimizeDataAdmission {
			w.notePendingClasses(c)
		}
		return
	}
	if c.Flags&(clientRegisterIntent|clientPositionIntent|clientRetryIntent|clientMetricsIntent) == 0 {
		return
	}
	index, err := w.generator.ClientIndex(id)
	if err != nil {
		w.generator.reportFatal(err)
		return
	}
	if !w.pending.push(uint32(index), false) {
		w.generator.reportFatal(errors.New("clients: pending ownership invariant"))
		return
	}
	c.Flags |= clientPendingIndex
	if w.generator.cfg.OptimizeDataAdmission {
		w.notePendingClasses(c)
	}
}
func (w *clientWorker) stageRegistrations() {
	g := w.generator
	v := int(g.cfg.Manifest.run.V)
	groups := len(g.clients) / v
	for group := w.index; group < groups; group += len(g.workers) {
		for index := group * v; index < (group+1)*v; index++ {
			c := &g.clients[index]
			c.Flags |= clientRegisterIntent
			c.CurrentRTO = int64(g.cfg.InitialRTO)
			w.nextSend(g.rangeConfig.FirstID+uint32(index), c)
			w.schedule(g.rangeConfig.FirstID+uint32(index), c)
		}
	}
}
func (w *clientWorker) beginPosition(cmd clientCommand) error {
	g := w.generator
	reject := func(err error) error { g.counters.rejectedOperations.Add(1); return err }
	c, err := w.client(cmd.ClientID)
	if err != nil {
		return reject(err)
	}
	if !c.Registered {
		return reject(ErrNotRegistered)
	}
	if g.cfg.Academic && c.Flags&clientAcademicStarted == 0 {
		return reject(ErrClientPhase)
	}
	if c.Phase != ClientReady && c.Phase != ClientComplete {
		return reject(ErrClientBusy)
	}
	if c.Phase == ClientComplete && (c.AckMask != c.ExpectedMask || c.Flags&clientPositionTxKnown == 0) {
		return reject(ErrClientBusy)
	}
	if cmd.Iteration == 0 || cmd.Iteration > g.cfg.Manifest.run.I || cmd.Iteration != c.CompletedIterations+1 {
		return reject(ErrIteration)
	}
	if !finite(cmd.Position.X) || !finite(cmd.Position.Y) || !finite(cmd.Position.Z) {
		return reject(ErrPayload)
	}
	c.MessageID = c.RegistrationNonce + uint64(cmd.Iteration)
	c.Iteration = cmd.Iteration
	c.Position = cmd.Position
	c.AckMask = 0
	c.SentAt, c.LastUniqueACKAt = -1, -1
	c.LastRTTNS = 0
	c.SendCount, c.RetryCounter, c.Failure = 0, 0, ClientFailureNone
	c.CurrentRTO = int64(g.cfg.InitialRTO)
	c.Flags &^= clientPositionTxKnown | clientPositionSlotReleased | clientAwaitConfirmation | clientRetryIntent
	c.Flags |= clientPositionIntent
	c.Phase = ClientPositionQueued
	w.nextSend(cmd.ClientID, c)
	g.counters.started.Add(1)
	w.schedule(cmd.ClientID, c)
	return nil
}
func (w *clientWorker) releaseRegisterSlot(c *ClientState) {
	if c.Flags&clientRegisterTxKnown != 0 && c.Flags&clientRegisterSlotReleased == 0 {
		c.Flags |= clientRegisterSlotReleased
		w.generator.registerInFlight.Add(-1)
		notify(w.generator.writerWake)
	}
}
func (w *clientWorker) finishOperation(c *ClientState) error {
	if c.Phase == ClientComplete {
		return nil
	}
	if c.Flags&clientPositionTxKnown == 0 || c.AckMask != c.ExpectedMask {
		return nil
	}
	if c.Phase != ClientPositionQueued && c.Phase != ClientPositionWait {
		return ErrClientPhase
	}
	var rtt uint64
	id := w.generator.rangeConfig.FirstID + uint32(c.RegistrationNonce-1)
	if c.ExpectedMask != 0 {
		if c.LastUniqueACKAt < c.SentAt {
			w.failClient(id, c, ClientTimestampFailed)
			return ErrClientTimestamp
		}
		rtt = uint64(c.LastUniqueACKAt - c.SentAt)
	}
	c.LastRTTNS = rtt
	if c.RTTSum > math.MaxUint64-rtt {
		w.failClient(id, c, ClientTimestampFailed)
		return ErrClientTimestamp
	}
	c.RTTSum += rtt
	c.CompletedIterations++
	c.Phase = ClientComplete
	w.removeDeadline(id, c)
	w.cancelSend(id, c)
	w.releasePositionSlot(c)
	if c.RetryCounter != 0 {
		w.generator.counters.recovered.Add(1)
	}
	w.generator.counters.rttSum.Add(rtt)
	w.generator.counters.rttSamples.Add(1)
	w.generator.counters.completed.Add(1)
	if w.generator.cfg.Academic {
		return w.advanceAcademic(id, c)
	}
	return nil
}
func (w *clientWorker) handleIncoming(e clientInbound) error {
	g := w.generator
	p := &e.Packet
	if err := g.cfg.Manifest.Validate(p); err != nil {
		return g.invalidPacket(err)
	}
	if p.Flags != ServerToClient {
		return g.invalidPacket(ErrFlags)
	}
	if p.GeneratorID != g.cfg.GeneratorID {
		return g.invalidPacket(ErrIdentity)
	}
	c, err := w.client(p.TargetID)
	if err != nil {
		return g.invalidPacket(err)
	}
	switch p.Type {
	case REGISTER_ACK:
		if p.RegistrationNonce != c.RegistrationNonce {
			return g.invalidPacket(ErrNonce)
		}
		if p.MessageID != c.RegistrationNonce {
			return g.invalidPacket(ErrMessage)
		}
		if c.Registered {
			g.counters.duplicateRegisterAcks.Add(1)
			return nil
		}
		if c.Phase != ClientRegisterQueued && c.Phase != ClientRegisterWait {
			return g.invalidPacket(ErrClientPhase)
		}
		c.Registered = true
		c.Phase = ClientReady
		if g.cfg.Academic {
			c.Phase = ClientWaitStart
		}
		w.removeDeadline(p.TargetID, c)
		w.cancelSend(p.TargetID, c)
		w.releaseRegisterSlot(c)
		g.counters.registered.Add(1)
		if g.cfg.Academic && c.Flags&clientPendingStart != 0 {
			return w.startAcademic(p.TargetID, c, c.PendingStartAttempt)
		}
		return nil
	case START, METRICS_ACK, GROUP_DONE:
		if !g.cfg.Academic {
			return g.invalidPacket(ErrType)
		}
		return w.handleAcademicControl(p, c)
	case POSITION, RETRY_POSITION:
		if !c.Registered {
			if g.cfg.Academic {
				// Server barrier can precede the authoritative local reply.
				// No ACK before registration; the source ARQ recovers later.
				g.counters.preRegistrationData.Add(1)
				return nil
			}
			return g.invalidPacket(ErrNotRegistered)
		}
		// Neighbor service does not depend on its own operation/iteration/RTT state.
		ack := Packet{Header: Header{Type: ACK, RunID: p.RunID, GeneratorID: p.GeneratorID, GroupID: p.GroupID,
			OriginID: p.TargetID, TargetID: p.OriginID, Iteration: p.Iteration, MessageID: p.MessageID, RetryCounter: p.RetryCounter}}
		if !w.high.push(clientOutbound{Packet: ack}, true) {
			g.counters.queueDrops.Add(1)
			return ErrQueueFull
		}
		notify(g.writerWake)
		return nil
	case ACK:
		if !c.Registered {
			return g.invalidPacket(ErrNotRegistered)
		}
		if g.cfg.Academic && p.Iteration < c.Iteration && p.MessageID == c.RegistrationNonce+uint64(p.Iteration) {
			g.counters.staleACKs.Add(1)
			return nil
		}
		if g.cfg.Academic && (c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait || c.Phase == ClientWaitGroupDone || c.Phase == ClientDone) {
			_, duplicate, ackErr := g.cfg.Manifest.ApplyACK(p, Operation{g.cfg.Manifest.run.RunID, c.RegistrationNonce + uint64(c.Iteration), p.TargetID, c.Iteration}, c.AckMask)
			if ackErr != nil {
				return g.invalidPacket(ackErr)
			}
			if !duplicate {
				return g.invalidPacket(ErrClientPhase)
			}
			g.counters.duplicateAcks.Add(1)
			return nil
		}
		if c.Phase != ClientPositionQueued && c.Phase != ClientPositionWait && c.Phase != ClientComplete {
			return g.invalidPacket(ErrClientPhase)
		}
		if e.ReceivedAt < 0 {
			return g.invalidPacket(ErrClientTimestamp)
		}
		mask, duplicate, err := g.cfg.Manifest.ApplyACK(p, Operation{g.cfg.Manifest.run.RunID, c.MessageID, p.TargetID, c.Iteration}, c.AckMask)
		if err != nil {
			return g.invalidPacket(err)
		}
		if duplicate {
			g.counters.duplicateAcks.Add(1)
			return nil
		}
		c.AckMask = mask
		if e.ReceivedAt > c.LastUniqueACKAt {
			c.LastUniqueACKAt = e.ReceivedAt
		}
		g.counters.uniqueAcks.Add(1)
		if err = w.finishOperation(c); err != nil {
			return g.invalidPacket(err)
		}
		return nil
	default:
		return g.invalidPacket(ErrType)
	}
}
func (w *clientWorker) handleConfirmation(tx clientConfirmation) error {
	g := w.generator
	invalid := func(err error) error { g.counters.invalidConfirmations.Add(1); return err }
	c, err := w.client(tx.ClientID)
	if err != nil {
		return invalid(err)
	}
	if tx.Status != clientSendOK && tx.Status != clientSendFailed {
		return invalid(ErrClientWrite)
	}
	if tx.Type != REGISTER && tx.Type != POSITION && tx.Type != RETRY_POSITION && tx.Type != METRICS {
		return invalid(ErrType)
	}
	if tx.Type == METRICS {
		return w.handleMetricsConfirmation(tx, c)
	}
	if g.cfg.Academic && tx.Type != REGISTER && c.CompletedIterations == g.cfg.Manifest.run.I && tx.Ticket < c.SendTicket && tx.MessageID == c.RegistrationNonce+uint64(tx.Iteration) && tx.Iteration == c.Iteration {
		g.counters.staleConfirmations.Add(1)
		return nil
	}
	// A claimed old retry can finish after a new operation has been admitted.
	// Its receipt cannot overwrite the new operation or rearm an old deadline.
	if tx.Type != REGISTER && tx.Ticket != 0 && tx.Ticket < c.SendTicket && tx.Iteration < c.Iteration && tx.MessageID == c.RegistrationNonce+uint64(tx.Iteration) {
		g.counters.staleConfirmations.Add(1)
		return nil
	}
	switch tx.Type {
	case REGISTER:
		if tx.MessageID != c.RegistrationNonce || tx.Iteration != 0 {
			return invalid(ErrMessage)
		}
		if tx.Ticket == 0 || tx.Ticket > c.SendTicket || tx.Attempt >= g.cfg.MaxSends {
			return invalid(ErrMessage)
		}
		if tx.Status == clientSendOK && tx.SentAt < 0 {
			return invalid(ErrClientTimestamp)
		}
		if tx.Attempt == 0 && c.Flags&clientRegisterTxKnown != 0 {
			g.counters.duplicateConfirmations.Add(1)
			return nil
		}
		if c.Registered {
			if tx.Attempt == 0 {
				c.Flags |= clientRegisterTxKnown
				w.releaseRegisterSlot(c)
				g.counters.confirmations.Add(1)
			} else {
				g.counters.staleConfirmations.Add(1)
			}
			return nil
		}
		if !c.Registered && c.Phase != ClientRegisterQueued {
			if c.Phase != ClientRegisterWait {
				return invalid(ErrClientPhase)
			}
		}
		if tx.Status == clientSendOK && tx.SentAt < 0 {
			return invalid(ErrClientTimestamp)
		}
		if tx.Ticket != c.SendTicket || tx.Attempt != c.RetryCounter {
			if tx.Ticket < c.SendTicket {
				g.counters.staleConfirmations.Add(1)
				return nil
			}
			return invalid(ErrMessage)
		}
		if c.Flags&clientAwaitConfirmation == 0 {
			g.counters.duplicateConfirmations.Add(1)
			return nil
		}
		c.Flags &^= clientAwaitConfirmation
		c.SendCount++
		if tx.Attempt == 0 {
			c.Flags |= clientRegisterTxKnown
		}
		if tx.Status == clientSendFailed {
			w.failClient(tx.ClientID, c, ClientWriteFailed)
		} else {
			c.Phase = ClientRegisterWait
			w.armDeadline(tx.ClientID, c, tx.SentAt)
		}
	case POSITION, RETRY_POSITION:
		if tx.MessageID != c.MessageID {
			return invalid(ErrMessage)
		}
		if tx.Iteration != c.Iteration {
			return invalid(ErrIteration)
		}
		if tx.Ticket != c.SendTicket || tx.Attempt != c.RetryCounter {
			if tx.Ticket < c.SendTicket {
				g.counters.staleConfirmations.Add(1)
				return nil
			}
			return invalid(ErrMessage)
		}
		if c.Flags&clientAwaitConfirmation == 0 {
			g.counters.duplicateConfirmations.Add(1)
			return nil
		}
		if c.Phase != ClientPositionQueued && c.Phase != ClientPositionWait && c.Phase != ClientComplete {
			return invalid(ErrClientPhase)
		}
		if tx.Status == clientSendOK && tx.SentAt < 0 {
			return invalid(ErrClientTimestamp)
		}
		c.Flags &^= clientAwaitConfirmation
		c.SendCount++
		if tx.Attempt == 0 {
			c.Flags |= clientPositionTxKnown
			c.SentAt = tx.SentAt
		}
		if tx.Status == clientSendFailed {
			if c.Phase != ClientComplete {
				w.failClient(tx.ClientID, c, ClientWriteFailed)
			}
		} else {
			if c.Phase != ClientComplete {
				c.Phase = ClientPositionWait
			}
			if err = w.finishOperation(c); err != nil {
				return invalid(err)
			}
			if c.Phase == ClientPositionWait && c.SendTicket == tx.Ticket {
				w.armDeadline(tx.ClientID, c, tx.SentAt)
			}
		}
	default:
		return invalid(ErrType)
	}
	g.counters.confirmations.Add(1)
	return nil
}
func (w *clientWorker) registerJob(id uint32, c *ClientState) clientOutbound {
	g := w.generator
	return clientOutbound{Packet: Packet{Header: Header{Type: REGISTER, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, MessageID: c.RegistrationNonce, RegistrationNonce: c.RegistrationNonce}}, ClientID: id, Ticket: c.SendTicket, Attempt: c.RetryCounter}
}
func (w *clientWorker) positionJob(id uint32, c *ClientState) clientOutbound {
	g := w.generator
	return clientOutbound{Packet: Packet{Header: Header{Type: POSITION, RunID: g.cfg.Manifest.run.RunID,
		GeneratorID: g.cfg.GeneratorID, GroupID: 1 + (id-1)/g.cfg.Manifest.run.V, OriginID: id, Iteration: c.Iteration, MessageID: c.MessageID}, Position: c.Position}, ClientID: id, Ticket: c.SendTicket, Attempt: c.RetryCounter}
}
func (w *clientWorker) flushPending() bool {
	if w.generator.cfg.OptimizeDataAdmission {
		return w.flushPendingOptimized()
	}
	return w.flushPendingHistorical()
}

func (w *clientWorker) flushPendingHistorical() bool {
	g := w.generator
	count := w.pending.stats().Depth
	budget := g.cfg.WorkBudget
	progress := false
	for turn := 0; turn < count && budget > 0; turn++ {
		index, ok := w.pending.pop()
		if !ok {
			break
		}
		id := g.rangeConfig.FirstID + index
		c, err := w.client(id)
		if err != nil {
			g.reportFatal(err)
			break
		}
		c.Flags &^= clientPendingIndex
		if c.Flags&clientRegisterIntent != 0 {
			budget--
			if w.control.push(w.registerJob(id, c), false) {
				c.Flags &^= clientRegisterIntent
				c.Phase = ClientRegisterQueued
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientPositionIntent != 0 && budget > 0 {
			budget--
			if w.data.push(w.positionJob(id, c), false) {
				c.Flags &^= clientPositionIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientMetricsIntent != 0 && budget > 0 {
			budget--
			if w.control.push(w.metricsJob(id, c), false) {
				c.Flags &^= clientMetricsIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		if c.Flags&clientRetryIntent != 0 && budget > 0 {
			budget--
			job := w.registerJob(id, c)
			if c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait {
				job = w.metricsJob(id, c)
			} else if c.Registered {
				job = w.positionJob(id, c)
				job.Packet.Type, job.Packet.RetryCounter = RETRY_POSITION, c.RetryCounter
				job.Packet.MissingMask = c.ExpectedMask &^ c.AckMask
				if job.Packet.MissingMask == 0 {
					c.Flags &^= clientRetryIntent
					w.finishOperation(c)
					w.schedule(id, c)
					continue
				}
			}
			if w.retry.push(job, false) {
				c.Flags &^= clientRetryIntent
				c.Flags |= clientAwaitConfirmation
				progress = true
				notify(g.writerWake)
			}
		}
		w.schedule(id, c)
	}
	return progress
}
func (w *clientWorker) run() {
	defer w.generator.wg.Done()
	defer w.finish()
	w.stageRegistrations()
	timer := time.NewTimer(time.Hour)
	stopClientTimer(timer)
	w.generator.activeARQTimers.Add(1)
	defer func() {
		timer.Stop()
		w.generator.activeARQTimers.Add(-1)
	}()
	for {
		if w.generator.ctx.Err() != nil {
			return
		}
		budget := w.generator.cfg.WorkBudget
		// Confirmations are independent of output space and always get service first.
		for i := 0; i < budget; i++ {
			tx, ok := w.confirmations.pop()
			if !ok {
				break
			}
			notify(w.generator.writerWake)
			w.handleConfirmation(tx)
		}
		for i := 0; i < budget; i++ {
			e, ok := w.input.pop()
			if !ok {
				break
			}
			w.handleIncoming(e)
		}
		for i := 0; i < budget; i++ {
			cmd, ok := w.commands.pop()
			if !ok {
				break
			}
			w.beginPosition(cmd)
		}
		progress := w.processTimeouts(w.generator.now())
		if w.flushPending() {
			progress = true
		}
		if progress || w.input.stats().Depth > 0 || w.confirmations.stats().Depth > 0 || w.commands.stats().Depth > 0 {
			continue
		}
		var tick <-chan time.Time
		if _, deadline, ok := w.deadlines.peek(); ok {
			delay := time.Duration(deadline - w.generator.now())
			if delay < 0 {
				delay = 0
			}
			timer.Reset(delay)
			tick = timer.C
		}
		select {
		case <-w.generator.ctx.Done():
			return
		case <-w.wake:
		case <-tick:
		}
		stopClientTimer(timer)
	}
}
func (w *clientWorker) finish() {
	g := w.generator
	v := int(g.cfg.Manifest.run.V)
	groups := len(g.clients) / v
	for group := w.index; group < groups; group += len(g.workers) {
		for index := group * v; index < (group+1)*v; index++ {
			c := &g.clients[index]
			if c.Flags&clientRegisterIntent != 0 {
				g.counters.shutdownIntentions.Add(1)
			}
			if c.Flags&clientPositionIntent != 0 {
				g.counters.shutdownIntentions.Add(1)
			}
			if c.Flags&clientRetryIntent != 0 {
				g.counters.shutdownIntentions.Add(1)
			}
			if c.Flags&clientMetricsIntent != 0 {
				g.counters.shutdownIntentions.Add(1)
			}
			w.removeDeadline(g.rangeConfig.FirstID+uint32(index), c)
			w.cancelSend(g.rangeConfig.FirstID+uint32(index), c)
			w.releasePositionSlot(c)
			if c.Phase == ClientPositionQueued || c.Phase == ClientPositionWait {
				g.counters.abortedOperations.Add(1)
				c.Phase = ClientStopped
			}
			if !c.Registered && c.Phase != ClientFailed {
				c.Phase = ClientStopped
			}
			if c.Phase == ClientMetricsQueued || c.Phase == ClientMetricsWait || c.Phase == ClientWaitStart || c.Phase == ClientWaitGroupDone {
				c.Phase = ClientStopped
			}
			c.Flags &^= clientRegisterIntent | clientPositionIntent | clientRetryIntent | clientMetricsIntent | clientPendingIndex | clientPendingControl | clientPendingRetry
		}
	}
	w.pendingControl, w.pendingRetry = 0, 0
}

// Peek is safe here because only this writer removes outbound heads. Producers
// only append, and shutdown clears queues only after the writer has stopped.
func clientQueuePeek[T any](q *boundedQueue[T]) (v T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		return v, false
	}
	return q.slots[q.head], true
}

// One writer produces confirmations, so room checked BEFORE a write cannot be
// consumed by another producer while the syscall runs. Workers only free slots.
func clientConfirmationRoom(q *boundedQueue[clientConfirmation]) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == len(q.slots) {
		q.full++
		return false
	}
	return true
}
func earliestClientTime(a, b int64) int64 {
	if a == 0 || b < a {
		return b
	}
	return a
}
func (g *ClientGenerator) popOutbound(class int, now int64, wait *int64) (clientOutbound, int, bool) {
	cursor := &g.highCursor
	if class == 1 {
		cursor = &g.controlCursor
	}
	if class == 2 {
		cursor = &g.dataCursor
	}
	if class == 3 {
		cursor = &g.retryCursor
	}
	for n := 0; n < len(g.workers); n++ {
		i := (*cursor + n) % len(g.workers)
		w := &g.workers[i]
		q := w.high
		if class == 1 {
			q = w.control
		}
		if class == 2 {
			q = w.data
		}
		if class == 3 {
			q = w.retry
		}
		for cleared := 0; cleared < g.cfg.WorkBudget; cleared++ {
			stale, present := clientQueuePeek(q)
			if !present || !g.outboundCancelled(stale) {
				break
			}
			q.pop()
			g.counters.cancelledSends.Add(1)
			notify(w.wake)
		}
		job, ok := clientQueuePeek(q)
		if !ok {
			continue
		}
		if class != 0 {
			if !clientConfirmationRoom(w.confirmations) {
				continue
			}
			at := g.positions.ready(now)
			if class == 1 {
				if job.Packet.Type == REGISTER && g.registerInFlight.Load() >= int64(g.cfg.RegisterWindow) {
					continue
				}
				at = now
				if job.Packet.Type == REGISTER {
					at = g.nextRegister
				}
			}
			if class == 2 && (g.hasRetryBacklog() || g.positionInFlight.Load() >= int64(g.cfg.PositionWindow)) {
				g.counters.admissionDelayed.Add(1)
				continue
			}
			if class == 3 {
				at = g.retries.ready(now)
			}
			if at > now {
				*wait = earliestClientTime(*wait, at)
				if class == 2 {
					g.counters.admissionDelayed.Add(1)
				}
				if class == 3 {
					g.counters.retryBudgetDelayed.Add(1)
				}
				continue
			}
		}
		if job, ok = q.pop(); ok {
			*cursor = (i + 1) % len(g.workers)
			notify(w.wake)
			return job, i, true
		}
	}
	return clientOutbound{}, 0, false
}
func (g *ClientGenerator) popHigh(now int64, wait *int64) (clientOutbound, int, bool) {
	if g.ackStreak >= g.cfg.AckBurst {
		if job, owner, ok := g.popOutbound(1, now, wait); ok {
			g.ackStreak = 0
			return job, owner, true
		}
	}
	if job, owner, ok := g.popOutbound(0, now, wait); ok {
		if g.ackStreak < g.cfg.AckBurst {
			g.ackStreak++
		}
		return job, owner, true
	}
	if job, owner, ok := g.popOutbound(1, now, wait); ok {
		g.ackStreak = 0
		return job, owner, true
	}
	return clientOutbound{}, 0, false
}
func (g *ClientGenerator) takeOutbound(now int64) (job clientOutbound, owner int, ok bool, wait int64) {
	if g.highStreak >= g.cfg.HighBurst {
		if job, owner, ok = g.popOutbound(3, now, &wait); ok {
			g.highStreak = 0
			return
		}
		if job, owner, ok = g.popOutbound(2, now, &wait); ok {
			g.highStreak = 0
			return
		}
	}
	if job, owner, ok = g.popHigh(now, &wait); ok {
		if g.highStreak < g.cfg.HighBurst {
			g.highStreak++
		}
		return
	}
	if job, owner, ok = g.popOutbound(3, now, &wait); ok {
		g.highStreak = 0
		return
	}
	if job, owner, ok = g.popOutbound(2, now, &wait); ok {
		g.highStreak = 0
		return
	}
	return
}
func (g *ClientGenerator) admitRegister() {
	depth := g.registerInFlight.Add(1)
	for {
		old := g.registerMaximum.Load()
		if depth <= old || g.registerMaximum.CompareAndSwap(old, depth) {
			break
		}
	}
}
func (g *ClientGenerator) publishConfirmation(owner int, tx clientConfirmation) bool {
	w := &g.workers[owner]
	if !w.confirmations.push(tx, true) {
		// A second producer would violate the reserved-room contract. Never silently
		// discard t0: terminate explicitly instead of producing an incorrect RTT.
		g.counters.queueDrops.Add(1)
		g.reportFatal(errors.New("clients: confirmation reservation invariant"))
		return false
	}
	notify(w.wake)
	return true
}
func stopClientTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
func (g *ClientGenerator) writer() {
	defer g.wg.Done()
	var buf [MaxPacketSize]byte
	timer := time.NewTimer(time.Hour)
	stopClientTimer(timer)
	g.activeWriterTimers.Add(1)
	defer func() {
		timer.Stop()
		g.activeWriterTimers.Add(-1)
	}()
	for {
		if g.ctx.Err() != nil {
			return
		}
		job, owner, ok, wait := g.takeOutbound(g.now())
		if !ok {
			var tick <-chan time.Time
			if wait != 0 {
				delay := time.Duration(wait - g.now())
				if delay < 0 {
					delay = 0
				}
				timer.Reset(delay)
				tick = timer.C
			}
			select {
			case <-g.ctx.Done():
				stopClientTimer(timer)
				return
			case <-g.writerWake:
			case <-tick:
			}
			stopClientTimer(timer)
			continue
		}
		p := &job.Packet
		needsConfirmation := p.Type == REGISTER || p.Type == POSITION || p.Type == RETRY_POSITION || p.Type == METRICS
		n, err := Encode(buf[:], p)
		if err != nil {
			g.counters.sendDrops.Add(1)
			g.reportFatal(fmt.Errorf("client outbound invariant: %w", err))
			return
		}
		sentAt := int64(-1)
		err = g.conn.SetWriteDeadline(time.Now().Add(g.cfg.WriteTimeout))
		if needsConfirmation && !g.claimOutbound(job) {
			g.counters.cancelledSends.Add(1)
			continue
		}
		admissionAt := g.now()
		if job.Attempt > 0 {
			if !g.retries.take(admissionAt) {
				g.reportFatal(ErrClientPhase)
				return
			}
		} else if p.Type == POSITION {
			if !g.positions.take(admissionAt) {
				g.reportFatal(ErrClientPhase)
				return
			}
			g.admitPosition()
		} else if p.Type == REGISTER {
			g.admitRegister()
		}
		if err == nil {
			sentAt = g.now() // AFTER encoding, pacing and deadline; immediately before Write.
			var sent int
			sent, err = g.conn.WriteToUDPAddrPort(buf[:n], g.cfg.Server)
			if err == nil && sent != n {
				err = io.ErrShortWrite
			}
		}
		attemptedAt := sentAt
		if attemptedAt < 0 {
			attemptedAt = g.now()
		}
		if p.Type == REGISTER && job.Attempt == 0 {
			g.nextRegister = attemptedAt + int64(time.Second)/int64(g.cfg.RegisterRate)
		}
		status := clientSendOK
		if err != nil {
			status = clientSendFailed
			g.counters.sendDrops.Add(1)
			if g.ctx.Err() == nil {
				g.counters.socketErrors.Add(1)
			}
		} else {
			g.counters.packetsSent.Add(1)
			switch p.Type {
			case REGISTER:
				g.counters.registerSent.Add(1)
				if job.Attempt > 0 {
					g.counters.registerRetries.Add(1)
				}
			case POSITION:
				g.counters.positionSent.Add(1)
			case RETRY_POSITION:
				g.counters.positionRetries.Add(1)
				g.counters.retryNeighbors.Add(uint64(bits.OnesCount32(p.MissingMask)))
			case ACK:
				g.counters.ackSent.Add(1)
			case START_ACK:
				g.counters.startAckSent.Add(1)
			case METRICS:
				g.counters.metricsSent.Add(1)
				if job.Attempt > 0 {
					g.counters.metricsRetries.Add(1)
				}
			case GROUP_DONE_ACK:
				g.counters.groupDoneAckSent.Add(1)
			}
		}
		if needsConfirmation {
			if !g.publishConfirmation(owner, clientConfirmation{MessageID: p.MessageID, SentAt: sentAt, ClientID: job.ClientID,
				Iteration: p.Iteration, Type: p.Type, Status: status, Ticket: job.Ticket, Attempt: job.Attempt}) {
				return
			}
		}
	}
}

// Stats is on-demand and race-safe; it allocates the W queue snapshots, not per
// packet. Individual counters/queues are consistent, not one global transaction.
func (g *ClientGenerator) Stats() ClientStats {
	c := &g.counters
	st := ClientStats{LogicalClients: uint64(len(g.clients)), RegisteredClients: c.registered.Load(),
		PacketsReceived: c.packetsReceived.Load(), PacketsSent: c.packetsSent.Load(), RegisterSent: c.registerSent.Load(),
		RegisterAckReceived: c.registerAckReceived.Load(), DuplicateRegisterAcks: c.duplicateRegisterAcks.Load(),
		PositionSent: c.positionSent.Load(), PositionReceived: c.positionReceived.Load(), AckSent: c.ackSent.Load(), AckReceived: c.ackReceived.Load(),
		UniqueAcks: c.uniqueAcks.Load(), DuplicateAcks: c.duplicateAcks.Load(), StartedOperations: c.started.Load(), CompletedOperations: c.completed.Load(),
		InvalidPackets: c.invalid.Load(), QueueDrops: c.queueDrops.Load(), SocketErrors: c.socketErrors.Load(), SendDrops: c.sendDrops.Load(),
		RejectedOperations: c.rejectedOperations.Load(), FailedRegistrations: c.failedRegistrations.Load(), FailedOperations: c.failedOperations.Load(),
		AbortedOperations: c.abortedOperations.Load(), ShutdownIntentions: c.shutdownIntentions.Load(), WriterConfirmations: c.confirmations.Load(),
		DuplicateConfirmations: c.duplicateConfirmations.Load(), InvalidConfirmations: c.invalidConfirmations.Load(), RTTSum: c.rttSum.Load(), RTTSamples: c.rttSamples.Load(),
		RegisterInFlight: g.registerInFlight.Load(), MaximumRegisterInFlight: g.registerMaximum.Load(), Queues: make([]ClientQueueStats, len(g.workers))}
	st.RegisterRetries, st.PositionRetries = c.registerRetries.Load(), c.positionRetries.Load()
	st.RetryPositionSent, st.RetryNeighborsRequested = st.PositionRetries, c.retryNeighbors.Load()
	st.Timeouts, st.RetryExhausted, st.RecoveredOperations = c.timeouts.Load(), c.retryExhausted.Load(), c.recovered.Load()
	st.InflightCurrent, st.InflightMaximum = g.positionInFlight.Load(), g.positionMaximum.Load()
	st.DeadlineHeapCurrent, st.DeadlineHeapMaximum = g.heapCurrent.Load(), g.heapMaximum.Load()
	st.ARQTimers, st.WriterTimers = g.activeARQTimers.Load(), g.activeWriterTimers.Load()
	st.AdmissionDelayed, st.AdmissionRejected = c.admissionDelayed.Load(), c.rejectedOperations.Load()
	st.RetryBudgetDelayed, st.CancelledSends, st.StaleConfirmations, st.NewPositionAdmitted = c.retryBudgetDelayed.Load(), c.cancelledSends.Load(), c.staleConfirmations.Load(), c.admitted.Load()
	st.StartReceived, st.StartAckSent, st.DuplicateStarts, st.StartedClients = c.startReceived.Load(), c.startAckSent.Load(), c.duplicateStarts.Load(), c.startedClients.Load()
	st.MetricsSent, st.MetricsAckReceived, st.MetricsRetries, st.MetricsAccepted, st.DuplicateMetricsAcks = c.metricsSent.Load(), c.metricsAckReceived.Load(), c.metricsRetries.Load(), c.metricsAccepted.Load(), c.duplicateMetricsAcks.Load()
	st.GroupDoneReceived, st.GroupDoneAckSent, st.DoneClients, st.DuplicateGroupDone = c.groupDoneReceived.Load(), c.groupDoneAckSent.Load(), c.doneClients.Load(), c.duplicateGroupDone.Load()
	st.FailedMetrics, st.ControlTimeouts, st.ControlRetryExhausted = c.failedMetrics.Load(), c.controlTimeouts.Load(), c.controlRetryExhausted.Load()
	st.PreRegistrationData, st.StaleACKs = c.preRegistrationData.Load(), c.staleACKs.Load()
	for i := range g.workers {
		w := &g.workers[i]
		st.Queues[i] = ClientQueueStats{Input: w.input.stats(), High: w.high.stats(), Control: w.control.stats(), Data: w.data.stats(),
			Confirmations: w.confirmations.stats(), Commands: w.commands.stats(), PendingClients: w.pending.stats(), Retry: w.retry.stats()}
	}
	return st
}
