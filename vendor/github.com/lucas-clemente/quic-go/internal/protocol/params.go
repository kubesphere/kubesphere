package protocol

import "time"

// MaxPacketSizeIPv4 is the maximum packet size that we use for sending IPv4 packets.
const MaxPacketSizeIPv4 = 1252

// MaxPacketSizeIPv6 is the maximum packet size that we use for sending IPv6 packets.
const MaxPacketSizeIPv6 = 1232

const defaultMaxCongestionWindowPackets = 1000

// DefaultMaxCongestionWindow is the default for the max congestion window
const DefaultMaxCongestionWindow ByteCount = defaultMaxCongestionWindowPackets * DefaultTCPMSS

// InitialCongestionWindow is the initial congestion window in QUIC packets
const InitialCongestionWindow ByteCount = 32 * DefaultTCPMSS

// MaxUndecryptablePackets limits the number of undecryptable packets that are queued in the session.
const MaxUndecryptablePackets = 10

// ConnectionFlowControlMultiplier determines how much larger the connection flow control windows needs to be relative to any stream's flow control window
// This is the value that Chromium is using
const ConnectionFlowControlMultiplier = 1.5

// InitialMaxStreamData is the stream-level flow control window for receiving data
const InitialMaxStreamData = (1 << 10) * 512 // 512 kb

// InitialMaxData is the connection-level flow control window for receiving data
const InitialMaxData = ConnectionFlowControlMultiplier * InitialMaxStreamData

// DefaultMaxReceiveStreamFlowControlWindow is the default maximum stream-level flow control window for receiving data, for the server
const DefaultMaxReceiveStreamFlowControlWindow = 6 * (1 << 20) // 6 MB

// DefaultMaxReceiveConnectionFlowControlWindow is the default connection-level flow control window for receiving data, for the server
const DefaultMaxReceiveConnectionFlowControlWindow = 15 * (1 << 20) // 12 MB

// WindowUpdateThreshold is the fraction of the receive window that has to be consumed before an higher offset is advertised to the client
const WindowUpdateThreshold = 0.25

// DefaultMaxIncomingStreams is the maximum number of streams that a peer may open
const DefaultMaxIncomingStreams = 100

// DefaultMaxIncomingUniStreams is the maximum number of unidirectional streams that a peer may open
const DefaultMaxIncomingUniStreams = 100

// MaxSessionUnprocessedPackets is the max number of packets stored in each session that are not yet processed.
const MaxSessionUnprocessedPackets = defaultMaxCongestionWindowPackets

// SkipPacketAveragePeriodLength is the average period length in which one packet number is skipped to prevent an Optimistic ACK attack
const SkipPacketAveragePeriodLength PacketNumber = 500

// MaxTrackedSkippedPackets is the maximum number of skipped packet numbers the SentPacketHandler keep track of for Optimistic ACK attack mitigation
const MaxTrackedSkippedPackets = 10

// MaxAcceptQueueSize is the maximum number of sessions that the server queues for accepting.
// If the queue is full, new connection attempts will be rejected.
const MaxAcceptQueueSize = 32

// TokenValidity is the duration that a (non-retry) token is considered valid
const TokenValidity = 24 * time.Hour

// RetryTokenValidity is the duration that a retry token is considered valid
const RetryTokenValidity = 10 * time.Second

// MaxOutstandingSentPackets is maximum number of packets saved for retransmission.
// When reached, it imposes a soft limit on sending new packets:
// Sending ACKs and retransmission is still allowed, but now new regular packets can be sent.
const MaxOutstandingSentPackets = 2 * defaultMaxCongestionWindowPackets

// MaxTrackedSentPackets is maximum number of sent packets saved for retransmission.
// When reached, no more packets will be sent.
// This value *must* be larger than MaxOutstandingSentPackets.
const MaxTrackedSentPackets = MaxOutstandingSentPackets * 5 / 4

// MaxNonAckElicitingAcks is the maximum number of packets containing an ACK,
// but no ack-eliciting frames, that we send in a row
const MaxNonAckElicitingAcks = 19

// MaxStreamFrameSorterGaps is the maximum number of gaps between received StreamFrames
// prevents DoS attacks against the streamFrameSorter
const MaxStreamFrameSorterGaps = 1000

// MinStreamFrameBufferSize is the minimum data length of a received STREAM frame
// that we use the buffer for. This protects against a DoS where an attacker would send us
// very small STREAM frames to consume a lot of memory.
const MinStreamFrameBufferSize = 128

// MaxCryptoStreamOffset is the maximum offset allowed on any of the crypto streams.
// This limits the size of the ClientHello and Certificates that can be received.
const MaxCryptoStreamOffset = 16 * (1 << 10)

// MinRemoteIdleTimeout is the minimum value that we accept for the remote idle timeout
const MinRemoteIdleTimeout = 5 * time.Second

// DefaultIdleTimeout is the default idle timeout
const DefaultIdleTimeout = 30 * time.Second

// DefaultHandshakeTimeout is the default timeout for a connection until the crypto handshake succeeds.
const DefaultHandshakeTimeout = 10 * time.Second

// RetiredConnectionIDDeleteTimeout is the time we keep closed sessions around in order to retransmit the CONNECTION_CLOSE.
// after this time all information about the old connection will be deleted
const RetiredConnectionIDDeleteTimeout = 5 * time.Second

// MinStreamFrameSize is the minimum size that has to be left in a packet, so that we add another STREAM frame.
// This avoids splitting up STREAM frames into small pieces, which has 2 advantages:
// 1. it reduces the framing overhead
// 2. it reduces the head-of-line blocking, when a packet is lost
const MinStreamFrameSize ByteCount = 128

// MaxPostHandshakeCryptoFrameSize is the maximum size of CRYPTO frames
// we send after the handshake completes.
const MaxPostHandshakeCryptoFrameSize ByteCount = 1000

// MaxAckFrameSize is the maximum size for an ACK frame that we write
// Due to the varint encoding, ACK frames can grow (almost) indefinitely large.
// The MaxAckFrameSize should be large enough to encode many ACK range,
// but must ensure that a maximum size ACK frame fits into one packet.
const MaxAckFrameSize ByteCount = 1000

// MaxNumAckRanges is the maximum number of ACK ranges that we send in an ACK frame.
// It also serves as a limit for the packet history.
// If at any point we keep track of more ranges, old ranges are discarded.
const MaxNumAckRanges = 500

// MinPacingDelay is the minimum duration that is used for packet pacing
// If the packet packing frequency is higher, multiple packets might be sent at once.
// Example: For a packet pacing delay of 20 microseconds, we would send 5 packets at once, wait for 100 microseconds, and so forth.
const MinPacingDelay time.Duration = 100 * time.Microsecond

// DefaultConnectionIDLength is the connection ID length that is used for multiplexed connections
// if no other value is configured.
const DefaultConnectionIDLength = 4

// MaxActiveConnectionIDs is the number of connection IDs that we're storing.
const MaxActiveConnectionIDs = 4

// MaxIssuedConnectionIDs is the maximum number of connection IDs that we're issuing at the same time.
const MaxIssuedConnectionIDs = 6

// PacketsPerConnectionID is the number of packets we send using one connection ID.
// If the peer provices us with enough new connection IDs, we switch to a new connection ID.
const PacketsPerConnectionID = 10000

// AckDelayExponent is the ack delay exponent used when sending ACKs.
const AckDelayExponent = 3

// Estimated timer granularity.
// The loss detection timer will not be set to a value smaller than granularity.
const TimerGranularity = time.Millisecond

// MaxAckDelay is the maximum time by which we delay sending ACKs.
const MaxAckDelay = 25 * time.Millisecond

// MaxAckDelayInclGranularity is the max_ack_delay including the timer granularity.
// This is the value that should be advertised to the peer.
const MaxAckDelayInclGranularity = MaxAckDelay + TimerGranularity

// KeyUpdateInterval is the maximum number of packets we send or receive before initiating a key udpate.
const KeyUpdateInterval = 100 * 1000
