// Package protocol implements the ARC binary v1 wire format and pure validation.
// It does not register clients, authenticate UDP endpoints, or run state machines.
package protocol

import (
	"encoding/binary"
	"errors"
	"math"
	"slices"
)

const (
	Magic          uint32 = 0x41524347 // ASCII ARCG
	Version        byte   = 1
	HeaderSize            = 64
	MaxPacketSize         = 80
	ServerToClient uint16 = 1
)

type MessageType uint8

const (
	REGISTER MessageType = 1 + iota
	REGISTER_ACK
	START
	START_ACK
	POSITION
	RETRY_POSITION
	ACK
	METRICS
	METRICS_ACK
	GROUP_DONE
	GROUP_DONE_ACK
	ERROR
)

type ErrorCode uint32

const (
	InvalidRegistration ErrorCode = 1 + iota
	ConflictingOperation
	OperationTimeout
	InternalError
)

// Sentinel errors avoid constructing diagnostic strings on packet paths.
var (
	ErrLength    = errors.New("protocol: invalid datagram length")
	ErrBuffer    = errors.New("protocol: destination buffer too short")
	ErrMagic     = errors.New("protocol: invalid magic")
	ErrVersion   = errors.New("protocol: unsupported version")
	ErrType      = errors.New("protocol: unknown message type")
	ErrFlags     = errors.New("protocol: invalid flags or direction")
	ErrReserved  = errors.New("protocol: nonzero reserved field")
	ErrRun       = errors.New("protocol: invalid or mismatched run")
	ErrIdentity  = errors.New("protocol: invalid identity, group or destination")
	ErrIteration = errors.New("protocol: invalid iteration")
	ErrMask      = errors.New("protocol: invalid neighbor mask")
	ErrMessage   = errors.New("protocol: invalid or mismatched message ID")
	ErrNonce     = errors.New("protocol: invalid registration nonce")
	ErrRetry     = errors.New("protocol: invalid retry counter")
	ErrPayload   = errors.New("protocol: invalid or inactive payload")
	ErrConfig    = errors.New("protocol: invalid run configuration")
	ErrManifest  = errors.New("protocol: invalid generator ranges")
)

// Header is a logical representation, not a memory image of the wire header.
// Magic/version/length/reserved bytes are controlled by the codec, not callers.
type Header struct {
	Type              MessageType
	Flags             uint16
	RunID             uint64
	GeneratorID       uint32
	GroupID           uint32
	OriginID          uint32
	TargetID          uint32
	Iteration         uint32
	MissingMask       uint32
	MessageID         uint64
	RegistrationNonce uint64
	RetryCounter      uint16
}

type PositionPayload struct{ X, Y, Z float32 }
type RegistrationPayload struct{ V, I uint32 }
type MetricsPayload struct {
	MeanRTTNS           uint64
	CompletedIterations uint32
}

// Packet owns its values: Decode does not retain or alias its source buffer.
// Only the payload corresponding to Type may be nonzero.
type Packet struct {
	Header
	Position     PositionPayload
	Registration RegistrationPayload
	Metrics      MetricsPayload
	ErrorCode    ErrorCode
}

func payloadSize(t MessageType) (int, error) {
	switch t {
	case POSITION, RETRY_POSITION:
		return 12, nil
	case REGISTER_ACK:
		return 8, nil
	case METRICS:
		return 16, nil
	case ERROR:
		return 4, nil
	case REGISTER, START, START_ACK, ACK, METRICS_ACK, GROUP_DONE, GROUP_DONE_ACK:
		return 0, nil
	default:
		return 0, ErrType
	}
}

func finite(f float32) bool { return math.Float32bits(f)&0x7f800000 != 0x7f800000 }

// validateShape checks all rules independent of a run's manifest/state.
func (p *Packet) validateShape() error {
	if p == nil {
		return ErrPayload
	}
	if _, err := payloadSize(p.Type); err != nil {
		return err
	}
	if p.Flags & ^ServerToClient != 0 {
		return ErrFlags
	}
	if p.RunID == 0 {
		return ErrRun
	}
	if p.GeneratorID == 0 {
		return ErrIdentity
	}
	if p.MessageID == 0 {
		return ErrMessage
	}
	data := p.Type == POSITION || p.Type == RETRY_POSITION || p.Type == ACK
	if data {
		if p.Iteration == 0 {
			return ErrIteration
		}
	} else if p.Iteration != 0 {
		return ErrIteration
	}
	if p.Type != RETRY_POSITION && p.MissingMask != 0 {
		return ErrMask
	}
	if p.Type != REGISTER && p.Type != REGISTER_ACK && !(p.Type == ERROR && p.TargetID == 0) && p.RegistrationNonce != 0 {
		return ErrNonce
	}
	// Comparing float bits also rejects negative zero in an inactive payload.
	if p.Type != POSITION && p.Type != RETRY_POSITION &&
		(math.Float32bits(p.Position.X)|math.Float32bits(p.Position.Y)|math.Float32bits(p.Position.Z)) != 0 {
		return ErrPayload
	}
	if p.Type != REGISTER_ACK && p.Registration != (RegistrationPayload{}) {
		return ErrPayload
	}
	if p.Type != METRICS && p.Metrics != (MetricsPayload{}) {
		return ErrPayload
	}
	if p.Type != ERROR && p.ErrorCode != 0 {
		return ErrPayload
	}

	switch p.Type {
	case REGISTER:
		if p.Flags != 0 {
			return ErrFlags
		}
		if p.OriginID != 0 || p.TargetID != 0 || p.GroupID != 0 {
			return ErrIdentity
		}
		if p.RegistrationNonce == 0 {
			return ErrNonce
		}
	case REGISTER_ACK:
		if p.Flags != ServerToClient {
			return ErrFlags
		}
		if p.OriginID != 0 || p.TargetID == 0 || p.GroupID == 0 {
			return ErrIdentity
		}
		if p.RegistrationNonce == 0 {
			return ErrNonce
		}
		if p.Registration.V < 1 || p.Registration.V > 32 || p.Registration.I == 0 {
			return ErrPayload
		}
	case POSITION, RETRY_POSITION:
		if p.OriginID == 0 || p.GroupID == 0 {
			return ErrIdentity
		}
		if p.Flags == 0 {
			if p.TargetID != 0 {
				return ErrIdentity
			}
		} else if p.TargetID == 0 || p.TargetID == p.OriginID {
			return ErrIdentity
		}
		if !finite(p.Position.X) || !finite(p.Position.Y) || !finite(p.Position.Z) {
			return ErrPayload
		}
		if p.Type == POSITION && p.RetryCounter != 0 {
			return ErrRetry
		}
		if p.Type == RETRY_POSITION {
			if p.RetryCounter == 0 {
				return ErrRetry
			}
			if p.MissingMask == 0 {
				return ErrMask
			}
		}
	case ACK:
		if p.OriginID == 0 || p.TargetID == 0 || p.OriginID == p.TargetID || p.GroupID == 0 {
			return ErrIdentity
		}
	case START, METRICS_ACK, GROUP_DONE:
		if p.Flags != ServerToClient {
			return ErrFlags
		}
		if p.OriginID != 0 || p.TargetID == 0 || p.GroupID == 0 {
			return ErrIdentity
		}
	case START_ACK, METRICS, GROUP_DONE_ACK:
		if p.Flags != 0 {
			return ErrFlags
		}
		if p.OriginID == 0 || p.TargetID != 0 || p.GroupID == 0 {
			return ErrIdentity
		}
		if p.Type == METRICS && p.Metrics.CompletedIterations == 0 {
			return ErrPayload
		}
	case ERROR:
		if p.Flags != ServerToClient {
			return ErrFlags
		}
		if p.OriginID != 0 {
			return ErrIdentity
		}
		if p.TargetID == 0 {
			// A registration failure has no assigned client yet. Echo its nonce.
			if p.GroupID != 0 {
				return ErrIdentity
			}
			if p.RegistrationNonce == 0 {
				return ErrNonce
			}
		} else if p.GroupID == 0 {
			return ErrIdentity
		}
		if p.ErrorCode < InvalidRegistration || p.ErrorCode > InternalError {
			return ErrPayload
		}
	}
	return nil
}

// Encode writes one canonical v1 datagram into caller-owned storage. On error it
// writes nothing. dst must have sufficient length (not merely capacity).
<<<<<<< HEAD
func Encode(dst []byte, p *Packet) (int, error) {
=======
func Encode(dst []byte, p *Packet) (int, error) { //codifica 
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
	if err := p.validateShape(); err != nil {
		return 0, err
	}
	pl, _ := payloadSize(p.Type)
	n := HeaderSize + pl
	if len(dst) < n {
		return 0, ErrBuffer
	}
	binary.BigEndian.PutUint32(dst[0:4], Magic)
	dst[4], dst[5] = Version, byte(p.Type)
	binary.BigEndian.PutUint16(dst[6:8], p.Flags)
	binary.BigEndian.PutUint64(dst[8:16], p.RunID)
	binary.BigEndian.PutUint32(dst[16:20], p.GeneratorID)
	binary.BigEndian.PutUint32(dst[20:24], p.GroupID)
	binary.BigEndian.PutUint32(dst[24:28], p.OriginID)
	binary.BigEndian.PutUint32(dst[28:32], p.TargetID)
	binary.BigEndian.PutUint32(dst[32:36], p.Iteration)
	binary.BigEndian.PutUint32(dst[36:40], p.MissingMask)
	binary.BigEndian.PutUint64(dst[40:48], p.MessageID)
	binary.BigEndian.PutUint64(dst[48:56], p.RegistrationNonce)
	binary.BigEndian.PutUint16(dst[56:58], p.RetryCounter)
	binary.BigEndian.PutUint16(dst[58:60], uint16(pl))
	binary.BigEndian.PutUint32(dst[60:64], 0)
	switch p.Type {
	case POSITION, RETRY_POSITION:
		binary.BigEndian.PutUint32(dst[64:68], math.Float32bits(p.Position.X))
		binary.BigEndian.PutUint32(dst[68:72], math.Float32bits(p.Position.Y))
		binary.BigEndian.PutUint32(dst[72:76], math.Float32bits(p.Position.Z))
	case REGISTER_ACK:
		binary.BigEndian.PutUint32(dst[64:68], p.Registration.V)
		binary.BigEndian.PutUint32(dst[68:72], p.Registration.I)
	case METRICS:
		binary.BigEndian.PutUint64(dst[64:72], p.Metrics.MeanRTTNS)
		binary.BigEndian.PutUint32(dst[72:76], p.Metrics.CompletedIterations)
		binary.BigEndian.PutUint32(dst[76:80], 0)
	case ERROR:
		binary.BigEndian.PutUint32(dst[64:68], uint32(p.ErrorCode))
	}
	return n, nil
}

// Decode rejects truncation, trailing bytes, noncanonical fields and invalid
// shapes. Errors return a zero Packet. A successful decode still needs Manifest
// validation and then endpoint/phase/operation checks by the future state owner.
func Decode(src []byte) (Packet, error) {
	if len(src) < HeaderSize || len(src) > MaxPacketSize {
		return Packet{}, ErrLength
	}
	if binary.BigEndian.Uint32(src[0:4]) != Magic {
		return Packet{}, ErrMagic
	}
	if src[4] != Version {
		return Packet{}, ErrVersion
	}
	t := MessageType(src[5])
	pl, err := payloadSize(t)
	if err != nil {
		return Packet{}, err
	}
	if int(binary.BigEndian.Uint16(src[58:60])) != pl || len(src) != HeaderSize+pl {
		return Packet{}, ErrLength
	}
	if binary.BigEndian.Uint32(src[60:64]) != 0 {
		return Packet{}, ErrReserved
	}
<<<<<<< HEAD
	p := Packet{Header: Header{
=======
	p := Packet{Header: Header{ //unpack 
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
		Type:              t,
		Flags:             binary.BigEndian.Uint16(src[6:8]),
		RunID:             binary.BigEndian.Uint64(src[8:16]),
		GeneratorID:       binary.BigEndian.Uint32(src[16:20]),
		GroupID:           binary.BigEndian.Uint32(src[20:24]),
		OriginID:          binary.BigEndian.Uint32(src[24:28]),
		TargetID:          binary.BigEndian.Uint32(src[28:32]),
		Iteration:         binary.BigEndian.Uint32(src[32:36]),
		MissingMask:       binary.BigEndian.Uint32(src[36:40]),
		MessageID:         binary.BigEndian.Uint64(src[40:48]),
		RegistrationNonce: binary.BigEndian.Uint64(src[48:56]),
		RetryCounter:      binary.BigEndian.Uint16(src[56:58]),
	}}
<<<<<<< HEAD
	switch t {
=======
	switch t { //en base al tipo decido que decodifico
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd
	case POSITION, RETRY_POSITION:
		p.Position = PositionPayload{
			math.Float32frombits(binary.BigEndian.Uint32(src[64:68])),
			math.Float32frombits(binary.BigEndian.Uint32(src[68:72])),
			math.Float32frombits(binary.BigEndian.Uint32(src[72:76])),
		}
	case REGISTER_ACK:
		p.Registration = RegistrationPayload{binary.BigEndian.Uint32(src[64:68]), binary.BigEndian.Uint32(src[68:72])}
	case METRICS:
		if binary.BigEndian.Uint32(src[76:80]) != 0 {
			return Packet{}, ErrReserved
		}
		p.Metrics = MetricsPayload{binary.BigEndian.Uint64(src[64:72]), binary.BigEndian.Uint32(src[72:76])}
	case ERROR:
		p.ErrorCode = ErrorCode(binary.BigEndian.Uint32(src[64:68]))
	}
	if err := p.validateShape(); err != nil {
		return Packet{}, err
	}
	return p, nil
}

type RunConfig struct {
	RunID   uint64
	N, V, I uint32
}

func (c RunConfig) Validate() error {
	if c.RunID == 0 || c.N == 0 || c.V < 1 || c.V > 32 || c.I == 0 || c.N%c.V != 0 {
		return ErrConfig
	}
	return nil
}

// Each generator repeats the run tuple so startup rejects mismatched manifests.
// Its inclusive range is FirstID .. FirstID+Count-1.
type GeneratorConfig struct {
	GeneratorID, FirstID, Count uint32
	Run                         RunConfig
}

// Manifest is an immutable snapshot after NewManifest; safe for concurrent reads.
// Sorting and allocation happen only at startup, not per datagram.
type Manifest struct {
	run        RunConfig
	generators []GeneratorConfig // sorted by GeneratorID
}

func NewManifest(run RunConfig, generators []GeneratorConfig) (*Manifest, error) {
	if err := run.Validate(); err != nil {
		return nil, err
	}
	if len(generators) == 0 {
		return nil, ErrManifest
	}
	g := slices.Clone(generators)
	slices.SortFunc(g, func(a, b GeneratorConfig) int {
		if a.FirstID < b.FirstID {
			return -1
		}
		if a.FirstID > b.FirstID {
			return 1
		}
		return 0
	})
	next := uint64(1)
	for _, v := range g {
		if v.Run != run {
			return nil, ErrRun
		}
		if v.GeneratorID == 0 || v.FirstID == 0 || v.Count == 0 ||
			(v.FirstID-1)%run.V != 0 || v.Count%run.V != 0 || uint64(v.FirstID) != next {
			return nil, ErrManifest
		}
		next = uint64(v.FirstID) + uint64(v.Count)
		if next > uint64(run.N)+1 {
			return nil, ErrManifest
		}
	}
	if next != uint64(run.N)+1 {
		return nil, ErrManifest
	}
	slices.SortFunc(g, func(a, b GeneratorConfig) int {
		if a.GeneratorID < b.GeneratorID {
			return -1
		}
		if a.GeneratorID > b.GeneratorID {
			return 1
		}
		return 0
	})
	for i := 1; i < len(g); i++ {
		if g[i-1].GeneratorID == g[i].GeneratorID {
			return nil, ErrManifest
		}
	}
	return &Manifest{run: run, generators: g}, nil
}

func (m *Manifest) Config() RunConfig { return m.run }

func (m *Manifest) generator(id uint32) (GeneratorConfig, bool) {
	if m == nil {
		return GeneratorConfig{}, false
	}
	lo, hi := 0, len(m.generators)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if m.generators[mid].GeneratorID < id {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo == len(m.generators) || m.generators[lo].GeneratorID != id {
		return GeneratorConfig{}, false
	}
	return m.generators[lo], true
}

func registrationIdentity(g GeneratorConfig, nonce uint64, v uint32) (uint32, uint32, error) {
	if nonce == 0 || nonce > uint64(g.Count) {
		return 0, 0, ErrNonce
	}
	id := g.FirstID + uint32(nonce-1) // range arithmetic was checked in uint64 at startup
	return id, 1 + (id-1)/v, nil
}

// RegistrationIdentity computes the server's authoritative assignment. This pure
// calculation grants no registered status: the generator must receive REGISTER_ACK.
func (m *Manifest) RegistrationIdentity(generatorID uint32, nonce uint64) (uint32, uint32, error) {
	g, ok := m.generator(generatorID)
	if !ok {
		return 0, 0, ErrIdentity
	}
	return registrationIdentity(g, nonce, m.run.V)
}

// Validate checks a packet against a validated manifest without changing state.
// Endpoint ownership, registration status, phase and idempotency require the
// future server/generator state owner; this method cannot attest any of them.
func (m *Manifest) Validate(p *Packet) error {
	if err := p.validateShape(); err != nil {
		return err
	}
	if m == nil || m.run.RunID == 0 || p.RunID != m.run.RunID {
		return ErrRun
	}
	g, ok := m.generator(p.GeneratorID)
	if !ok {
		return ErrIdentity
	}
	for _, id := range [2]uint32{p.OriginID, p.TargetID} {
		if id == 0 {
			continue
		}
		if id < g.FirstID || uint64(id) >= uint64(g.FirstID)+uint64(g.Count) || p.GroupID != 1+(id-1)/m.run.V {
			return ErrIdentity
		}
	}
	if p.Iteration > m.run.I {
		return ErrIteration
	}
	switch p.Type {
	case REGISTER, REGISTER_ACK:
		id, group, err := registrationIdentity(g, p.RegistrationNonce, m.run.V)
		if err != nil {
			return err
		}
		if p.Type == REGISTER_ACK {
			if p.TargetID != id || p.GroupID != group {
				return ErrIdentity
			}
			if p.Registration != (RegistrationPayload{m.run.V, m.run.I}) {
				return ErrPayload
			}
		}
	case RETRY_POSITION:
		expected, _ := ExpectedMask(p.OriginID, m.run.V)
		if p.MissingMask & ^expected != 0 {
			return ErrMask
		}
		if p.Flags == ServerToClient && p.MissingMask&(uint32(1)<<((p.TargetID-1)%m.run.V)) == 0 {
			return ErrMask
		}
	case METRICS:
		if p.Metrics.CompletedIterations != m.run.I {
			return ErrIteration
		}
	}
	return nil
}

func GroupMask(v uint32) (uint32, error) {
	if v < 1 || v > 32 {
		return 0, ErrMask
	}
	return uint32((uint64(1) << v) - 1), nil
}

func ExpectedMask(clientID, v uint32) (uint32, error) {
	mask, err := GroupMask(v)
	if err != nil {
		return 0, err
	}
	if clientID == 0 {
		return 0, ErrIdentity
	}
	return mask &^ (uint32(1) << ((clientID - 1) % v)), nil
}

func MissingMask(expected, received uint32) (uint32, error) {
	if received & ^expected != 0 {
		return 0, ErrMask
	}
	return expected &^ received, nil
}

// Operation identifies the POSITION awaiting real neighbor ACKs at its origin.
type Operation struct {
	RunID, MessageID    uint64
	OriginID, Iteration uint32
}

// ApplyACK is a pure mask operation for a forwarded ACK at the origin. On error
// received is returned unchanged. Duplicates never create additional progress.
// It deliberately does not require matching RetryCounter: a delayed real ACK from
// any attempt of this same operation is useful. Endpoint checking stays external.
func (m *Manifest) ApplyACK(p *Packet, op Operation, received uint32) (updated uint32, duplicate bool, err error) {
	if err = m.Validate(p); err != nil {
		return received, false, err
	}
	if p.Type != ACK {
		return received, false, ErrType
	}
	if p.Flags != ServerToClient {
		return received, false, ErrFlags
	}
	if p.RunID != op.RunID {
		return received, false, ErrRun
	}
	if p.TargetID != op.OriginID {
		return received, false, ErrIdentity
	}
	if p.Iteration != op.Iteration {
		return received, false, ErrIteration
	}
	if p.MessageID != op.MessageID {
		return received, false, ErrMessage
	}
	expected, _ := ExpectedMask(op.OriginID, m.run.V)
	if _, err = MissingMask(expected, received); err != nil {
		return received, false, err
	}
	bit := uint32(1) << ((p.OriginID - 1) % m.run.V)
	return received | bit, received&bit != 0, nil
}
