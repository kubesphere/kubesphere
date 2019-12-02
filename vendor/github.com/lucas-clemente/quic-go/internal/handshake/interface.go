package handshake

import (
	"crypto/tls"
	"errors"
	"io"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/marten-seemann/qtls"
)

var (
	// ErrOpenerNotYetAvailable is returned when an opener is requested for an encryption level,
	// but the corresponding opener has not yet been initialized
	// This can happen when packets arrive out of order.
	ErrOpenerNotYetAvailable = errors.New("CryptoSetup: opener at this encryption level not yet available")
	// ErrKeysDropped is returned when an opener or a sealer is requested for an encryption level,
	// but the corresponding keys have already been dropped.
	ErrKeysDropped = errors.New("CryptoSetup: keys were already dropped")
	// ErrDecryptionFailed is returned when the AEAD fails to open the packet.
	ErrDecryptionFailed = errors.New("decryption failed")
)

type headerDecryptor interface {
	DecryptHeader(sample []byte, firstByte *byte, pnBytes []byte)
}

// LongHeaderOpener opens a long header packet
type LongHeaderOpener interface {
	headerDecryptor
	Open(dst, src []byte, pn protocol.PacketNumber, associatedData []byte) ([]byte, error)
}

// ShortHeaderOpener opens a short header packet
type ShortHeaderOpener interface {
	headerDecryptor
	Open(dst, src []byte, rcvTime time.Time, pn protocol.PacketNumber, kp protocol.KeyPhaseBit, associatedData []byte) ([]byte, error)
}

// LongHeaderSealer seals a long header packet
type LongHeaderSealer interface {
	Seal(dst, src []byte, packetNumber protocol.PacketNumber, associatedData []byte) []byte
	EncryptHeader(sample []byte, firstByte *byte, pnBytes []byte)
	Overhead() int
}

// ShortHeaderSealer seals a short header packet
type ShortHeaderSealer interface {
	LongHeaderSealer
	KeyPhase() protocol.KeyPhaseBit
}

// A tlsExtensionHandler sends and received the QUIC TLS extension.
type tlsExtensionHandler interface {
	GetExtensions(msgType uint8) []qtls.Extension
	ReceivedExtensions(msgType uint8, exts []qtls.Extension)
	TransportParameters() <-chan []byte
}

type handshakeRunner interface {
	OnReceivedParams([]byte)
	OnHandshakeComplete()
	OnError(error)
	DropKeys(protocol.EncryptionLevel)
}

// CryptoSetup handles the handshake and protecting / unprotecting packets
type CryptoSetup interface {
	RunHandshake()
	io.Closer
	ChangeConnectionID(protocol.ConnectionID)

	HandleMessage([]byte, protocol.EncryptionLevel) bool
	SetLargest1RTTAcked(protocol.PacketNumber)
	ConnectionState() tls.ConnectionState

	GetInitialOpener() (LongHeaderOpener, error)
	GetHandshakeOpener() (LongHeaderOpener, error)
	Get1RTTOpener() (ShortHeaderOpener, error)

	GetInitialSealer() (LongHeaderSealer, error)
	GetHandshakeSealer() (LongHeaderSealer, error)
	Get1RTTSealer() (ShortHeaderSealer, error)
}
