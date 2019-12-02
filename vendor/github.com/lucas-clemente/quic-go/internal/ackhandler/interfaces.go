package ackhandler

import (
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"
	"github.com/lucas-clemente/quic-go/quictrace"
)

// A Packet is a packet
type Packet struct {
	PacketNumber    protocol.PacketNumber
	Frames          []Frame
	LargestAcked    protocol.PacketNumber // InvalidPacketNumber if the packet doesn't contain an ACK
	Length          protocol.ByteCount
	EncryptionLevel protocol.EncryptionLevel
	SendTime        time.Time

	includedInBytesInFlight bool
}

// SentPacketHandler handles ACKs received for outgoing packets
type SentPacketHandler interface {
	// SentPacket may modify the packet
	SentPacket(packet *Packet)
	ReceivedAck(ackFrame *wire.AckFrame, withPacketNumber protocol.PacketNumber, encLevel protocol.EncryptionLevel, recvTime time.Time) error
	DropPackets(protocol.EncryptionLevel)
	ResetForRetry() error

	// The SendMode determines if and what kind of packets can be sent.
	SendMode() SendMode
	// TimeUntilSend is the time when the next packet should be sent.
	// It is used for pacing packets.
	TimeUntilSend() time.Time
	// ShouldSendNumPackets returns the number of packets that should be sent immediately.
	// It always returns a number greater or equal than 1.
	// A number greater than 1 is returned when the pacing delay is smaller than the minimum pacing delay.
	// Note that the number of packets is only calculated based on the pacing algorithm.
	// Before sending any packet, SendingAllowed() must be called to learn if we can actually send it.
	ShouldSendNumPackets() int

	// only to be called once the handshake is complete
	GetLowestPacketNotConfirmedAcked() protocol.PacketNumber
	QueueProbePacket() bool /* was a packet queued */

	PeekPacketNumber(protocol.EncryptionLevel) (protocol.PacketNumber, protocol.PacketNumberLen)
	PopPacketNumber(protocol.EncryptionLevel) protocol.PacketNumber

	GetLossDetectionTimeout() time.Time
	OnLossDetectionTimeout() error

	// report some congestion statistics. For tracing only.
	GetStats() *quictrace.TransportState
}

// ReceivedPacketHandler handles ACKs needed to send for incoming packets
type ReceivedPacketHandler interface {
	ReceivedPacket(pn protocol.PacketNumber, encLevel protocol.EncryptionLevel, rcvTime time.Time, shouldInstigateAck bool)
	IgnoreBelow(protocol.PacketNumber)
	DropPackets(protocol.EncryptionLevel)

	GetAlarmTimeout() time.Time
	GetAckFrame(protocol.EncryptionLevel) *wire.AckFrame
}
