package protocol

import (
	"fmt"
	"time"
)

// The PacketType is the Long Header Type
type PacketType uint8

const (
	// PacketTypeInitial is the packet type of an Initial packet
	PacketTypeInitial PacketType = 1 + iota
	// PacketTypeRetry is the packet type of a Retry packet
	PacketTypeRetry
	// PacketTypeHandshake is the packet type of a Handshake packet
	PacketTypeHandshake
	// PacketType0RTT is the packet type of a 0-RTT packet
	PacketType0RTT
)

func (t PacketType) String() string {
	switch t {
	case PacketTypeInitial:
		return "Initial"
	case PacketTypeRetry:
		return "Retry"
	case PacketTypeHandshake:
		return "Handshake"
	case PacketType0RTT:
		return "0-RTT Protected"
	default:
		return fmt.Sprintf("unknown packet type: %d", t)
	}
}

// A ByteCount in QUIC
type ByteCount uint64

// MaxByteCount is the maximum value of a ByteCount
const MaxByteCount = ByteCount(1<<62 - 1)

// An ApplicationErrorCode is an application-defined error code.
type ApplicationErrorCode uint64

// MaxReceivePacketSize maximum packet size of any QUIC packet, based on
// ethernet's max size, minus the IP and UDP headers. IPv6 has a 40 byte header,
// UDP adds an additional 8 bytes.  This is a total overhead of 48 bytes.
// Ethernet's max packet size is 1500 bytes,  1500 - 48 = 1452.
const MaxReceivePacketSize ByteCount = 1452

// DefaultTCPMSS is the default maximum packet size used in the Linux TCP implementation.
// Used in QUIC for congestion window computations in bytes.
const DefaultTCPMSS ByteCount = 1460

// MinInitialPacketSize is the minimum size an Initial packet is required to have.
const MinInitialPacketSize = 1200

// MinStatelessResetSize is the minimum size of a stateless reset packet that we send
const MinStatelessResetSize = 1 /* first byte */ + 20 /* max. conn ID length */ + 4 /* max. packet number length */ + 1 /* min. payload length */ + 16 /* token */

// MinConnectionIDLenInitial is the minimum length of the destination connection ID on an Initial packet.
const MinConnectionIDLenInitial = 8

// DefaultAckDelayExponent is the default ack delay exponent
const DefaultAckDelayExponent = 3

// MaxAckDelayExponent is the maximum ack delay exponent
const MaxAckDelayExponent = 20

// DefaultMaxAckDelay is the default max_ack_delay
const DefaultMaxAckDelay = 25 * time.Millisecond

// MaxMaxAckDelay is the maximum max_ack_delay
const MaxMaxAckDelay = 1 << 14 * time.Millisecond

// MaxConnIDLen is the maximum length of the connection ID
const MaxConnIDLen = 20
