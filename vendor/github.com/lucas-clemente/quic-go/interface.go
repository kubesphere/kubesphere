package quic

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/quictrace"
)

// The StreamID is the ID of a QUIC stream.
type StreamID = protocol.StreamID

// A VersionNumber is a QUIC version number.
type VersionNumber = protocol.VersionNumber

// A Token can be used to verify the ownership of the client address.
type Token struct {
	// IsRetryToken encodes how the client received the token. There are two ways:
	// * In a Retry packet sent when trying to establish a new connection.
	// * In a NEW_TOKEN frame on a previous connection.
	IsRetryToken bool
	RemoteAddr   string
	SentTime     time.Time
}

// A ClientToken is a token received by the client.
// It can be used to skip address validation on future connection attempts.
type ClientToken struct {
	data []byte
}

type TokenStore interface {
	// Pop searches for a ClientToken associated with the given key.
	// Since tokens are not supposed to be reused, it must remove the token from the cache.
	// It returns nil when no token is found.
	Pop(key string) (token *ClientToken)

	// Put adds a token to the cache with the given key. It might get called
	// multiple times in a connection.
	Put(key string, token *ClientToken)
}

// An ErrorCode is an application-defined error code.
// Valid values range between 0 and MAX_UINT62.
type ErrorCode = protocol.ApplicationErrorCode

// Stream is the interface implemented by QUIC streams
type Stream interface {
	// StreamID returns the stream ID.
	StreamID() StreamID
	// Read reads data from the stream.
	// Read can be made to time out and return a net.Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	// If the stream was canceled by the peer, the error implements the StreamError
	// interface, and Canceled() == true.
	// If the session was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	io.Reader
	// Write writes data to the stream.
	// Write can be made to time out and return a net.Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	// If the stream was canceled by the peer, the error implements the StreamError
	// interface, and Canceled() == true.
	// If the session was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// CancelWrite aborts sending on this stream.
	// Data already written, but not yet delivered to the peer is not guaranteed to be delivered reliably.
	// Write will unblock immediately, and future calls to Write will fail.
	// When called multiple times or after closing the stream it is a no-op.
	CancelWrite(ErrorCode)
	// CancelRead aborts receiving on this stream.
	// It will ask the peer to stop transmitting stream data.
	// Read will unblock immediately, and future Read calls will fail.
	// When called multiple times or after reading the io.EOF it is a no-op.
	CancelRead(ErrorCode)
	// The context is canceled as soon as the write-side of the stream is closed.
	// This happens when Close() or CancelWrite() is called, or when the peer
	// cancels the read-side of their stream.
	// Warning: This API should not be considered stable and might change soon.
	Context() context.Context
	// SetReadDeadline sets the deadline for future Read calls and
	// any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error
	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	SetDeadline(t time.Time) error
}

// A ReceiveStream is a unidirectional Receive Stream.
type ReceiveStream interface {
	// see Stream.StreamID
	StreamID() StreamID
	// see Stream.Read
	io.Reader
	// see Stream.CancelRead
	CancelRead(ErrorCode)
	// see Stream.SetReadDealine
	SetReadDeadline(t time.Time) error
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// see Stream.StreamID
	StreamID() StreamID
	// see Stream.Write
	io.Writer
	// see Stream.Close
	io.Closer
	// see Stream.CancelWrite
	CancelWrite(ErrorCode)
	// see Stream.Context
	Context() context.Context
	// see Stream.SetWriteDeadline
	SetWriteDeadline(t time.Time) error
}

// StreamError is returned by Read and Write when the peer cancels the stream.
type StreamError interface {
	error
	Canceled() bool
	ErrorCode() ErrorCode
}

// A Session is a QUIC connection between two peers.
type Session interface {
	// AcceptStream returns the next stream opened by the peer, blocking until one is available.
	// If the session was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	AcceptStream(context.Context) (Stream, error)
	// AcceptUniStream returns the next unidirectional stream opened by the peer, blocking until one is available.
	// If the session was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	AcceptUniStream(context.Context) (ReceiveStream, error)
	// OpenStream opens a new bidirectional QUIC stream.
	// There is no signaling to the peer about new streams:
	// The peer can only accept the stream after data has been sent on the stream.
	// If the error is non-nil, it satisfies the net.Error interface.
	// When reaching the peer's stream limit, err.Temporary() will be true.
	// If the session was closed due to a timeout, Timeout() will be true.
	OpenStream() (Stream, error)
	// OpenStreamSync opens a new bidirectional QUIC stream.
	// It blocks until a new stream can be opened.
	// If the error is non-nil, it satisfies the net.Error interface.
	// If the session was closed due to a timeout, Timeout() will be true.
	OpenStreamSync(context.Context) (Stream, error)
	// OpenUniStream opens a new outgoing unidirectional QUIC stream.
	// If the error is non-nil, it satisfies the net.Error interface.
	// When reaching the peer's stream limit, Temporary() will be true.
	// If the session was closed due to a timeout, Timeout() will be true.
	OpenUniStream() (SendStream, error)
	// OpenUniStreamSync opens a new outgoing unidirectional QUIC stream.
	// It blocks until a new stream can be opened.
	// If the error is non-nil, it satisfies the net.Error interface.
	// If the session was closed due to a timeout, Timeout() will be true.
	OpenUniStreamSync(context.Context) (SendStream, error)
	// LocalAddr returns the local address.
	LocalAddr() net.Addr
	// RemoteAddr returns the address of the peer.
	RemoteAddr() net.Addr
	// Close the connection.
	io.Closer
	// Close the connection with an error.
	// The error string will be sent to the peer.
	CloseWithError(ErrorCode, string) error
	// The context is cancelled when the session is closed.
	// Warning: This API should not be considered stable and might change soon.
	Context() context.Context
	// ConnectionState returns basic details about the QUIC connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnectionState() tls.ConnectionState
}

// An EarlySession is a session that is handshaking.
// Data sent during the handshake is encrypted using the forward secure keys.
// When using client certificates, the client's identity is only verified
// after completion of the handshake.
type EarlySession interface {
	Session

	// Blocks until the handshake completes (or fails).
	// Data sent before completion of the handshake is encrypted with 1-RTT keys.
	// Note that the client's identity hasn't been verified yet.
	HandshakeComplete() context.Context
}

// Config contains all configuration data needed for a QUIC server or client.
type Config struct {
	// The QUIC versions that can be negotiated.
	// If not set, it uses all versions available.
	// Warning: This API should not be considered stable and will change soon.
	Versions []VersionNumber
	// The length of the connection ID in bytes.
	// It can be 0, or any value between 4 and 18.
	// If not set, the interpretation depends on where the Config is used:
	// If used for dialing an address, a 0 byte connection ID will be used.
	// If used for a server, or dialing on a packet conn, a 4 byte connection ID will be used.
	// When dialing on a packet conn, the ConnectionIDLength value must be the same for every Dial call.
	ConnectionIDLength int
	// HandshakeTimeout is the maximum duration that the cryptographic handshake may take.
	// If the timeout is exceeded, the connection is closed.
	// If this value is zero, the timeout is set to 10 seconds.
	HandshakeTimeout time.Duration
	// IdleTimeout is the maximum duration that may pass without any incoming network activity.
	// This value only applies after the handshake has completed.
	// If the timeout is exceeded, the connection is closed.
	// If this value is zero, the timeout is set to 30 seconds.
	IdleTimeout time.Duration
	// AcceptToken determines if a Token is accepted.
	// It is called with token = nil if the client didn't send a token.
	// If not set, a default verification function is used:
	// * it verifies that the address matches, and
	//   * if the token is a retry token, that it was issued within the last 5 seconds
	//   * else, that it was issued within the last 24 hours.
	// This option is only valid for the server.
	AcceptToken func(clientAddr net.Addr, token *Token) bool
	// The TokenStore stores tokens received from the server.
	// Tokens are used to skip address validation on future connection attempts.
	// The key used to store tokens is the ServerName from the tls.Config, if set
	// otherwise the token is associated with the server's IP address.
	TokenStore TokenStore
	// MaxReceiveStreamFlowControlWindow is the maximum stream-level flow control window for receiving data.
	// If this value is zero, it will default to 1 MB for the server and 6 MB for the client.
	MaxReceiveStreamFlowControlWindow uint64
	// MaxReceiveConnectionFlowControlWindow is the connection-level flow control window for receiving data.
	// If this value is zero, it will default to 1.5 MB for the server and 15 MB for the client.
	MaxReceiveConnectionFlowControlWindow uint64
	// MaxIncomingStreams is the maximum number of concurrent bidirectional streams that a peer is allowed to open.
	// If not set, it will default to 100.
	// If set to a negative value, it doesn't allow any bidirectional streams.
	MaxIncomingStreams int
	// MaxIncomingUniStreams is the maximum number of concurrent unidirectional streams that a peer is allowed to open.
	// If not set, it will default to 100.
	// If set to a negative value, it doesn't allow any unidirectional streams.
	MaxIncomingUniStreams int
	// The StatelessResetKey is used to generate stateless reset tokens.
	// If no key is configured, sending of stateless resets is disabled.
	StatelessResetKey []byte
	// KeepAlive defines whether this peer will periodically send a packet to keep the connection alive.
	KeepAlive bool
	// QUIC Event Tracer.
	// Warning: Experimental. This API should not be considered stable and will change soon.
	QuicTracer quictrace.Tracer
}

// A Listener for incoming QUIC connections
type Listener interface {
	// Close the server. All active sessions will be closed.
	Close() error
	// Addr returns the local network addr that the server is listening on.
	Addr() net.Addr
	// Accept returns new sessions. It should be called in a loop.
	Accept(context.Context) (Session, error)
}

// An EarlyListener listens for incoming QUIC connections,
// and returns them before the handshake completes.
type EarlyListener interface {
	// Close the server. All active sessions will be closed.
	Close() error
	// Addr returns the local network addr that the server is listening on.
	Addr() net.Addr
	// Accept returns new early sessions. It should be called in a loop.
	Accept(context.Context) (EarlySession, error)
}
