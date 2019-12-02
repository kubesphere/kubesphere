package quictrace

import (
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

// A Tracer traces a QUIC connection
type Tracer interface {
	Trace(protocol.ConnectionID, Event)
	GetAllTraces() map[string][]byte
}

// EventType is the type of an event
type EventType uint8

const (
	// PacketSent means that a packet was sent
	PacketSent EventType = 1 + iota
	// PacketReceived means that a packet was received
	PacketReceived
	// PacketLost means that a packet was lost
	PacketLost
)

// Event is a quic-traceable event
type Event struct {
	Time      time.Time
	EventType EventType

	TransportState  *TransportState
	EncryptionLevel protocol.EncryptionLevel
	PacketNumber    protocol.PacketNumber
	PacketSize      protocol.ByteCount
	Frames          []wire.Frame
}

// TransportState contains some transport and congestion statistics
type TransportState struct {
	MinRTT      time.Duration
	SmoothedRTT time.Duration
	LatestRTT   time.Duration

	BytesInFlight    protocol.ByteCount
	CongestionWindow protocol.ByteCount
	InSlowStart      bool
	InRecovery       bool
}
