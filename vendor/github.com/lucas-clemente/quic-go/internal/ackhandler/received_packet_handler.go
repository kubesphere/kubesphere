package ackhandler

import (
	"fmt"
	"time"

	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

const (
	// initial maximum number of ack-eliciting packets received before sending an ack.
	initialAckElicitingPacketsBeforeAck = 2
	// number of ack-eliciting that an ACK is sent for
	ackElicitingPacketsBeforeAck = 10
	// 1/5 RTT delay when doing ack decimation
	ackDecimationDelay = 1.0 / 4
	// 1/8 RTT delay when doing ack decimation
	shortAckDecimationDelay = 1.0 / 8
	// Minimum number of packets received before ack decimation is enabled.
	// This intends to avoid the beginning of slow start, when CWNDs may be
	// rapidly increasing.
	minReceivedBeforeAckDecimation = 100
	// Maximum number of packets to ack immediately after a missing packet for
	// fast retransmission to kick in at the sender.  This limit is created to
	// reduce the number of acks sent that have no benefit for fast retransmission.
	// Set to the number of nacks needed for fast retransmit plus one for protection
	// against an ack loss
	maxPacketsAfterNewMissing = 4
)

type receivedPacketHandler struct {
	initialPackets   *receivedPacketTracker
	handshakePackets *receivedPacketTracker
	oneRTTPackets    *receivedPacketTracker
}

var _ ReceivedPacketHandler = &receivedPacketHandler{}

// NewReceivedPacketHandler creates a new receivedPacketHandler
func NewReceivedPacketHandler(
	rttStats *congestion.RTTStats,
	logger utils.Logger,
	version protocol.VersionNumber,
) ReceivedPacketHandler {
	return &receivedPacketHandler{
		initialPackets:   newReceivedPacketTracker(rttStats, logger, version),
		handshakePackets: newReceivedPacketTracker(rttStats, logger, version),
		oneRTTPackets:    newReceivedPacketTracker(rttStats, logger, version),
	}
}

func (h *receivedPacketHandler) ReceivedPacket(
	pn protocol.PacketNumber,
	encLevel protocol.EncryptionLevel,
	rcvTime time.Time,
	shouldInstigateAck bool,
) {
	switch encLevel {
	case protocol.EncryptionInitial:
		h.initialPackets.ReceivedPacket(pn, rcvTime, shouldInstigateAck)
	case protocol.EncryptionHandshake:
		h.handshakePackets.ReceivedPacket(pn, rcvTime, shouldInstigateAck)
	case protocol.Encryption1RTT:
		h.oneRTTPackets.ReceivedPacket(pn, rcvTime, shouldInstigateAck)
	default:
		panic(fmt.Sprintf("received packet with unknown encryption level: %s", encLevel))
	}
}

// only to be used with 1-RTT packets
func (h *receivedPacketHandler) IgnoreBelow(pn protocol.PacketNumber) {
	h.oneRTTPackets.IgnoreBelow(pn)
}

func (h *receivedPacketHandler) DropPackets(encLevel protocol.EncryptionLevel) {
	switch encLevel {
	case protocol.EncryptionInitial:
		h.initialPackets = nil
	case protocol.EncryptionHandshake:
		h.handshakePackets = nil
	default:
		panic(fmt.Sprintf("Cannot drop keys for encryption level %s", encLevel))
	}
}

func (h *receivedPacketHandler) GetAlarmTimeout() time.Time {
	var initialAlarm, handshakeAlarm time.Time
	if h.initialPackets != nil {
		initialAlarm = h.initialPackets.GetAlarmTimeout()
	}
	if h.handshakePackets != nil {
		handshakeAlarm = h.handshakePackets.GetAlarmTimeout()
	}
	oneRTTAlarm := h.oneRTTPackets.GetAlarmTimeout()
	return utils.MinNonZeroTime(utils.MinNonZeroTime(initialAlarm, handshakeAlarm), oneRTTAlarm)
}

func (h *receivedPacketHandler) GetAckFrame(encLevel protocol.EncryptionLevel) *wire.AckFrame {
	var ack *wire.AckFrame
	switch encLevel {
	case protocol.EncryptionInitial:
		if h.initialPackets != nil {
			ack = h.initialPackets.GetAckFrame()
		}
	case protocol.EncryptionHandshake:
		if h.handshakePackets != nil {
			ack = h.handshakePackets.GetAckFrame()
		}
	case protocol.Encryption1RTT:
		return h.oneRTTPackets.GetAckFrame()
	default:
		return nil
	}
	// For Initial and Handshake ACKs, the delay time is ignored by the receiver.
	// Set it to 0 in order to save bytes.
	if ack != nil {
		ack.DelayTime = 0
	}
	return ack
}
