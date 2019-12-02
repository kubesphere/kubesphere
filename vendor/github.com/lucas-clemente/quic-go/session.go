package quic

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go/internal/ackhandler"
	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/flowcontrol"
	"github.com/lucas-clemente/quic-go/internal/handshake"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
	"github.com/lucas-clemente/quic-go/quictrace"
)

type unpacker interface {
	Unpack(hdr *wire.Header, rcvTime time.Time, data []byte) (*unpackedPacket, error)
}

type streamGetter interface {
	GetOrOpenReceiveStream(protocol.StreamID) (receiveStreamI, error)
	GetOrOpenSendStream(protocol.StreamID) (sendStreamI, error)
}

type streamManager interface {
	GetOrOpenSendStream(protocol.StreamID) (sendStreamI, error)
	GetOrOpenReceiveStream(protocol.StreamID) (receiveStreamI, error)
	OpenStream() (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenStreamSync(context.Context) (Stream, error)
	OpenUniStreamSync(context.Context) (SendStream, error)
	AcceptStream(context.Context) (Stream, error)
	AcceptUniStream(context.Context) (ReceiveStream, error)
	DeleteStream(protocol.StreamID) error
	UpdateLimits(*handshake.TransportParameters) error
	HandleMaxStreamsFrame(*wire.MaxStreamsFrame) error
	CloseWithError(error)
}

type cryptoStreamHandler interface {
	RunHandshake()
	ChangeConnectionID(protocol.ConnectionID)
	SetLargest1RTTAcked(protocol.PacketNumber)
	io.Closer
	ConnectionState() tls.ConnectionState
}

type receivedPacket struct {
	remoteAddr net.Addr
	rcvTime    time.Time
	data       []byte

	buffer *packetBuffer
}

func (p *receivedPacket) Clone() *receivedPacket {
	return &receivedPacket{
		remoteAddr: p.remoteAddr,
		rcvTime:    p.rcvTime,
		data:       p.data,
		buffer:     p.buffer,
	}
}

type sessionRunner interface {
	Add(protocol.ConnectionID, packetHandler) [16]byte
	Retire(protocol.ConnectionID)
	Remove(protocol.ConnectionID)
	ReplaceWithClosed(protocol.ConnectionID, packetHandler)
	AddResetToken([16]byte, packetHandler)
	RemoveResetToken([16]byte)
	RetireResetToken([16]byte)
}

type handshakeRunner struct {
	onReceivedParams    func([]byte)
	onError             func(error)
	dropKeys            func(protocol.EncryptionLevel)
	onHandshakeComplete func()
}

func (r *handshakeRunner) OnReceivedParams(b []byte)            { r.onReceivedParams(b) }
func (r *handshakeRunner) OnError(e error)                      { r.onError(e) }
func (r *handshakeRunner) DropKeys(el protocol.EncryptionLevel) { r.dropKeys(el) }
func (r *handshakeRunner) OnHandshakeComplete()                 { r.onHandshakeComplete() }

type closeError struct {
	err       error
	remote    bool
	immediate bool
}

var errCloseForRecreating = errors.New("closing session in order to recreate it")

// A Session is a QUIC session
type session struct {
	origDestConnID protocol.ConnectionID // if the server sends a Retry, this is the connection ID we used initially
	srcConnIDLen   int

	perspective    protocol.Perspective
	initialVersion protocol.VersionNumber // if version negotiation is performed, this is the version we initially tried
	version        protocol.VersionNumber
	config         *Config

	conn      connection
	sendQueue *sendQueue

	streamsMap      streamManager
	connIDManager   *connIDManager
	connIDGenerator *connIDGenerator

	rttStats *congestion.RTTStats

	cryptoStreamManager   *cryptoStreamManager
	sentPacketHandler     ackhandler.SentPacketHandler
	receivedPacketHandler ackhandler.ReceivedPacketHandler
	retransmissionQueue   *retransmissionQueue
	framer                framer
	windowUpdateQueue     *windowUpdateQueue
	connFlowController    flowcontrol.ConnectionFlowController
	tokenStoreKey         string                    // only set for the client
	tokenGenerator        *handshake.TokenGenerator // only set for the server

	unpacker    unpacker
	frameParser wire.FrameParser
	packer      packer

	cryptoStreamHandler cryptoStreamHandler

	receivedPackets  chan *receivedPacket
	sendingScheduled chan struct{}

	closeOnce sync.Once
	// closeChan is used to notify the run loop that it should terminate
	closeChan chan closeError

	ctx                context.Context
	ctxCancel          context.CancelFunc
	handshakeCtx       context.Context
	handshakeCtxCancel context.CancelFunc

	undecryptablePackets []*receivedPacket

	clientHelloWritten    <-chan struct{}
	earlySessionReadyChan chan struct{}
	handshakeCompleteChan chan struct{} // is closed when the handshake completes
	handshakeComplete     bool

	receivedRetry       bool
	receivedFirstPacket bool

	sessionCreationTime time.Time
	// The idle timeout is set based on the max of the time we received the last packet...
	lastPacketReceivedTime time.Time
	// ... and the time we sent a new ack-eliciting packet after receiving a packet.
	firstAckElicitingPacketAfterIdleSentTime time.Time
	// pacingDeadline is the time when the next packet should be sent
	pacingDeadline time.Time

	peerParams *handshake.TransportParameters

	timer *utils.Timer
	// keepAlivePingSent stores whether a Ping frame was sent to the peer or not
	// it is reset as soon as we receive a packet from the peer
	keepAlivePingSent bool

	traceCallback func(quictrace.Event)

	logID  string
	logger utils.Logger
}

var _ Session = &session{}
var _ EarlySession = &session{}
var _ streamSender = &session{}

var newSession = func(
	conn connection,
	runner sessionRunner,
	origDestConnID protocol.ConnectionID,
	clientDestConnID protocol.ConnectionID,
	destConnID protocol.ConnectionID,
	srcConnID protocol.ConnectionID,
	conf *Config,
	tlsConf *tls.Config,
	tokenGenerator *handshake.TokenGenerator,
	logger utils.Logger,
	v protocol.VersionNumber,
) quicSession {
	s := &session{
		conn:                  conn,
		config:                conf,
		srcConnIDLen:          srcConnID.Len(),
		tokenGenerator:        tokenGenerator,
		perspective:           protocol.PerspectiveServer,
		handshakeCompleteChan: make(chan struct{}),
		logger:                logger,
		version:               v,
	}
	if origDestConnID != nil {
		s.logID = origDestConnID.String()
	} else {
		s.logID = destConnID.String()
	}
	s.connIDManager = newConnIDManager(
		destConnID,
		func(token [16]byte) { runner.AddResetToken(token, s) },
		runner.RemoveResetToken,
		runner.RetireResetToken,
		s.queueControlFrame,
	)
	s.connIDGenerator = newConnIDGenerator(
		srcConnID,
		func(connID protocol.ConnectionID) [16]byte { return runner.Add(connID, s) },
		runner.Remove,
		runner.Retire,
		runner.ReplaceWithClosed,
		s.queueControlFrame,
	)
	s.preSetup()
	s.sentPacketHandler = ackhandler.NewSentPacketHandler(0, s.rttStats, s.traceCallback, s.logger)
	initialStream := newCryptoStream()
	handshakeStream := newCryptoStream()
	oneRTTStream := newPostHandshakeCryptoStream(s.framer)
	token := runner.Add(srcConnID, s)
	params := &handshake.TransportParameters{
		InitialMaxStreamDataBidiLocal:  protocol.InitialMaxStreamData,
		InitialMaxStreamDataBidiRemote: protocol.InitialMaxStreamData,
		InitialMaxStreamDataUni:        protocol.InitialMaxStreamData,
		InitialMaxData:                 protocol.InitialMaxData,
		IdleTimeout:                    s.config.IdleTimeout,
		MaxBidiStreamNum:               protocol.StreamNum(s.config.MaxIncomingStreams),
		MaxUniStreamNum:                protocol.StreamNum(s.config.MaxIncomingUniStreams),
		MaxAckDelay:                    protocol.MaxAckDelayInclGranularity,
		AckDelayExponent:               protocol.AckDelayExponent,
		DisableMigration:               true,
		StatelessResetToken:            &token,
		OriginalConnectionID:           origDestConnID,
		ActiveConnectionIDLimit:        protocol.MaxActiveConnectionIDs,
	}
	cs := handshake.NewCryptoSetupServer(
		initialStream,
		handshakeStream,
		oneRTTStream,
		clientDestConnID,
		conn.RemoteAddr(),
		params,
		&handshakeRunner{
			onReceivedParams:    s.processTransportParameters,
			onError:             s.closeLocal,
			dropKeys:            s.dropEncryptionLevel,
			onHandshakeComplete: func() { close(s.handshakeCompleteChan) },
		},
		tlsConf,
		s.rttStats,
		logger,
	)
	s.cryptoStreamHandler = cs
	s.packer = newPacketPacker(
		srcConnID,
		s.connIDManager.Get,
		initialStream,
		handshakeStream,
		s.sentPacketHandler,
		s.retransmissionQueue,
		s.RemoteAddr(),
		cs,
		s.framer,
		s.receivedPacketHandler,
		s.perspective,
		s.version,
	)
	s.unpacker = newPacketUnpacker(cs, s.version)
	s.cryptoStreamManager = newCryptoStreamManager(cs, initialStream, handshakeStream, oneRTTStream)
	return s
}

// declare this as a variable, such that we can it mock it in the tests
var newClientSession = func(
	conn connection,
	runner sessionRunner,
	destConnID protocol.ConnectionID,
	srcConnID protocol.ConnectionID,
	conf *Config,
	tlsConf *tls.Config,
	initialPacketNumber protocol.PacketNumber,
	initialVersion protocol.VersionNumber,
	logger utils.Logger,
	v protocol.VersionNumber,
) quicSession {
	s := &session{
		conn:                  conn,
		config:                conf,
		srcConnIDLen:          srcConnID.Len(),
		perspective:           protocol.PerspectiveClient,
		handshakeCompleteChan: make(chan struct{}),
		logID:                 destConnID.String(),
		logger:                logger,
		initialVersion:        initialVersion,
		version:               v,
	}
	s.connIDManager = newConnIDManager(
		destConnID,
		func(token [16]byte) { runner.AddResetToken(token, s) },
		runner.RemoveResetToken,
		runner.RetireResetToken,
		s.queueControlFrame,
	)
	s.connIDGenerator = newConnIDGenerator(
		srcConnID,
		func(connID protocol.ConnectionID) [16]byte { return runner.Add(connID, s) },
		runner.Remove,
		runner.Retire,
		runner.ReplaceWithClosed,
		s.queueControlFrame,
	)
	s.preSetup()
	s.sentPacketHandler = ackhandler.NewSentPacketHandler(initialPacketNumber, s.rttStats, s.traceCallback, s.logger)
	initialStream := newCryptoStream()
	handshakeStream := newCryptoStream()
	oneRTTStream := newPostHandshakeCryptoStream(s.framer)
	params := &handshake.TransportParameters{
		InitialMaxStreamDataBidiRemote: protocol.InitialMaxStreamData,
		InitialMaxStreamDataBidiLocal:  protocol.InitialMaxStreamData,
		InitialMaxStreamDataUni:        protocol.InitialMaxStreamData,
		InitialMaxData:                 protocol.InitialMaxData,
		IdleTimeout:                    s.config.IdleTimeout,
		MaxBidiStreamNum:               protocol.StreamNum(s.config.MaxIncomingStreams),
		MaxUniStreamNum:                protocol.StreamNum(s.config.MaxIncomingUniStreams),
		MaxAckDelay:                    protocol.MaxAckDelayInclGranularity,
		AckDelayExponent:               protocol.AckDelayExponent,
		DisableMigration:               true,
		ActiveConnectionIDLimit:        protocol.MaxActiveConnectionIDs,
	}
	cs, clientHelloWritten := handshake.NewCryptoSetupClient(
		initialStream,
		handshakeStream,
		oneRTTStream,
		destConnID,
		conn.RemoteAddr(),
		params,
		&handshakeRunner{
			onReceivedParams:    s.processTransportParameters,
			onError:             s.closeLocal,
			dropKeys:            s.dropEncryptionLevel,
			onHandshakeComplete: func() { close(s.handshakeCompleteChan) },
		},
		tlsConf,
		s.rttStats,
		logger,
	)
	s.clientHelloWritten = clientHelloWritten
	s.cryptoStreamHandler = cs
	s.cryptoStreamManager = newCryptoStreamManager(cs, initialStream, handshakeStream, oneRTTStream)
	s.unpacker = newPacketUnpacker(cs, s.version)
	s.packer = newPacketPacker(
		srcConnID,
		s.connIDManager.Get,
		initialStream,
		handshakeStream,
		s.sentPacketHandler,
		s.retransmissionQueue,
		s.RemoteAddr(),
		cs,
		s.framer,
		s.receivedPacketHandler,
		s.perspective,
		s.version,
	)
	if len(tlsConf.ServerName) > 0 {
		s.tokenStoreKey = tlsConf.ServerName
	} else {
		s.tokenStoreKey = conn.RemoteAddr().String()
	}
	if s.config.TokenStore != nil {
		if token := s.config.TokenStore.Pop(s.tokenStoreKey); token != nil {
			s.packer.SetToken(token.data)
		}
	}
	return s
}

func (s *session) preSetup() {
	s.sendQueue = newSendQueue(s.conn)
	s.retransmissionQueue = newRetransmissionQueue(s.version)
	s.frameParser = wire.NewFrameParser(s.version)
	s.rttStats = &congestion.RTTStats{}
	s.receivedPacketHandler = ackhandler.NewReceivedPacketHandler(s.rttStats, s.logger, s.version)
	s.connFlowController = flowcontrol.NewConnectionFlowController(
		protocol.InitialMaxData,
		protocol.ByteCount(s.config.MaxReceiveConnectionFlowControlWindow),
		s.onHasConnectionWindowUpdate,
		s.rttStats,
		s.logger,
	)
	s.earlySessionReadyChan = make(chan struct{})
	s.streamsMap = newStreamsMap(
		s,
		s.newFlowController,
		uint64(s.config.MaxIncomingStreams),
		uint64(s.config.MaxIncomingUniStreams),
		s.perspective,
		s.version,
	)
	s.framer = newFramer(s.streamsMap, s.version)
	s.receivedPackets = make(chan *receivedPacket, protocol.MaxSessionUnprocessedPackets)
	s.closeChan = make(chan closeError, 1)
	s.sendingScheduled = make(chan struct{}, 1)
	s.undecryptablePackets = make([]*receivedPacket, 0, protocol.MaxUndecryptablePackets)
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.handshakeCtx, s.handshakeCtxCancel = context.WithCancel(context.Background())

	s.timer = utils.NewTimer()
	now := time.Now()
	s.lastPacketReceivedTime = now
	s.sessionCreationTime = now

	s.windowUpdateQueue = newWindowUpdateQueue(s.streamsMap, s.connFlowController, s.framer.QueueControlFrame)

	if s.config.QuicTracer != nil {
		s.traceCallback = func(ev quictrace.Event) {
			s.config.QuicTracer.Trace(s.origDestConnID, ev)
		}
	}
}

// run the session main loop
func (s *session) run() error {
	defer s.ctxCancel()

	go s.cryptoStreamHandler.RunHandshake()
	go func() {
		if err := s.sendQueue.Run(); err != nil {
			s.closeLocal(err)
		}
	}()

	if s.perspective == protocol.PerspectiveClient {
		select {
		case <-s.clientHelloWritten:
			s.scheduleSending()
		case closeErr := <-s.closeChan:
			// put the close error back into the channel, so that the run loop can receive it
			s.closeChan <- closeErr
		}
	}

	var closeErr closeError

runLoop:
	for {
		// Close immediately if requested
		select {
		case closeErr = <-s.closeChan:
			break runLoop
		case <-s.handshakeCompleteChan:
			s.handleHandshakeComplete()
		default:
		}

		s.maybeResetTimer()

		select {
		case closeErr = <-s.closeChan:
			break runLoop
		case <-s.timer.Chan():
			s.timer.SetRead()
			// We do all the interesting stuff after the switch statement, so
			// nothing to see here.
		case <-s.sendingScheduled:
			// We do all the interesting stuff after the switch statement, so
			// nothing to see here.
		case p := <-s.receivedPackets:
			// Only reset the timers if this packet was actually processed.
			// This avoids modifying any state when handling undecryptable packets,
			// which could be injected by an attacker.
			if wasProcessed := s.handlePacketImpl(p); !wasProcessed {
				continue
			}
		case <-s.handshakeCompleteChan:
			s.handleHandshakeComplete()
		}

		now := time.Now()
		if timeout := s.sentPacketHandler.GetLossDetectionTimeout(); !timeout.IsZero() && timeout.Before(now) {
			// This could cause packets to be retransmitted.
			// Check it before trying to send packets.
			if err := s.sentPacketHandler.OnLossDetectionTimeout(); err != nil {
				s.closeLocal(err)
			}
		}

		var pacingDeadline time.Time
		if s.pacingDeadline.IsZero() { // the timer didn't have a pacing deadline set
			pacingDeadline = s.sentPacketHandler.TimeUntilSend()
		}
		if s.config.KeepAlive && !s.keepAlivePingSent && s.handshakeComplete && s.firstAckElicitingPacketAfterIdleSentTime.IsZero() && time.Since(s.lastPacketReceivedTime) >= s.peerParams.IdleTimeout/2 {
			// send a PING frame since there is no activity in the session
			s.logger.Debugf("Sending a keep-alive ping to keep the connection alive.")
			s.framer.QueueControlFrame(&wire.PingFrame{})
			s.keepAlivePingSent = true
		} else if !pacingDeadline.IsZero() && now.Before(pacingDeadline) {
			// If we get to this point before the pacing deadline, we should wait until that deadline.
			// This can happen when scheduleSending is called, or a packet is received.
			// Set the timer and restart the run loop.
			s.pacingDeadline = pacingDeadline
			continue
		}

		if !s.handshakeComplete && now.Sub(s.sessionCreationTime) >= s.config.HandshakeTimeout {
			s.destroyImpl(qerr.TimeoutError("Handshake did not complete in time"))
			continue
		}
		if s.handshakeComplete && now.Sub(s.idleTimeoutStartTime()) >= s.config.IdleTimeout {
			s.destroyImpl(qerr.TimeoutError("No recent network activity"))
			continue
		}

		if err := s.sendPackets(); err != nil {
			s.closeLocal(err)
		}
	}

	s.handleCloseError(closeErr)
	s.logger.Infof("Connection %s closed.", s.logID)
	s.cryptoStreamHandler.Close()
	s.sendQueue.Close()
	return closeErr.err
}

// blocks until the early session can be used
func (s *session) earlySessionReady() <-chan struct{} {
	return s.earlySessionReadyChan
}

func (s *session) HandshakeComplete() context.Context {
	return s.handshakeCtx
}

func (s *session) Context() context.Context {
	return s.ctx
}

func (s *session) ConnectionState() tls.ConnectionState {
	return s.cryptoStreamHandler.ConnectionState()
}

func (s *session) maybeResetTimer() {
	var deadline time.Time
	if s.config.KeepAlive && s.handshakeComplete && !s.keepAlivePingSent {
		deadline = s.idleTimeoutStartTime().Add(s.peerParams.IdleTimeout / 2)
	} else {
		deadline = s.idleTimeoutStartTime().Add(s.config.IdleTimeout)
	}

	if ackAlarm := s.receivedPacketHandler.GetAlarmTimeout(); !ackAlarm.IsZero() {
		deadline = utils.MinTime(deadline, ackAlarm)
	}
	if lossTime := s.sentPacketHandler.GetLossDetectionTimeout(); !lossTime.IsZero() {
		deadline = utils.MinTime(deadline, lossTime)
	}
	if !s.handshakeComplete {
		handshakeDeadline := s.sessionCreationTime.Add(s.config.HandshakeTimeout)
		deadline = utils.MinTime(deadline, handshakeDeadline)
	}
	if !s.pacingDeadline.IsZero() {
		deadline = utils.MinTime(deadline, s.pacingDeadline)
	}

	s.timer.Reset(deadline)
}

func (s *session) idleTimeoutStartTime() time.Time {
	return utils.MaxTime(s.lastPacketReceivedTime, s.firstAckElicitingPacketAfterIdleSentTime)
}

func (s *session) handleHandshakeComplete() {
	s.handshakeComplete = true
	s.handshakeCompleteChan = nil // prevent this case from ever being selected again
	s.handshakeCtxCancel()

	// The client completes the handshake first (after sending the CFIN).
	// We need to make sure it learns about the server completing the handshake,
	// in order to stop retransmitting handshake packets.
	// They will stop retransmitting handshake packets when receiving the first 1-RTT packet.
	if s.perspective == protocol.PerspectiveServer {
		token, err := s.tokenGenerator.NewToken(s.conn.RemoteAddr())
		if err != nil {
			s.closeLocal(err)
		}
		s.queueControlFrame(&wire.NewTokenFrame{Token: token})
	}
}

func (s *session) handlePacketImpl(rp *receivedPacket) bool {
	var counter uint8
	var lastConnID protocol.ConnectionID
	var processed bool
	data := rp.data
	p := rp
	for len(data) > 0 {
		if counter > 0 {
			p = p.Clone()
			p.data = data
		}

		hdr, packetData, rest, err := wire.ParsePacket(p.data, s.srcConnIDLen)
		if err != nil {
			s.logger.Debugf("error parsing packet: %s", err)
			break
		}

		if counter > 0 && !hdr.DestConnectionID.Equal(lastConnID) {
			s.logger.Debugf("coalesced packet has different destination connection ID: %s, expected %s", hdr.DestConnectionID, lastConnID)
			break
		}
		lastConnID = hdr.DestConnectionID

		if counter > 0 {
			p.buffer.Split()
		}
		counter++

		// only log if this actually a coalesced packet
		if s.logger.Debug() && (counter > 1 || len(rest) > 0) {
			s.logger.Debugf("Parsed a coalesced packet. Part %d: %d bytes. Remaining: %d bytes.", counter, len(packetData), len(rest))
		}
		p.data = packetData
		if wasProcessed := s.handleSinglePacket(p, hdr); wasProcessed {
			processed = true
		}
		data = rest
	}
	p.buffer.MaybeRelease()
	return processed
}

func (s *session) handleSinglePacket(p *receivedPacket, hdr *wire.Header) bool /* was the packet successfully processed */ {
	var wasQueued bool

	defer func() {
		// Put back the packet buffer if the packet wasn't queued for later decryption.
		if !wasQueued {
			p.buffer.Decrement()
		}
	}()

	if hdr.Type == protocol.PacketTypeRetry {
		return s.handleRetryPacket(hdr)
	}

	// The server can change the source connection ID with the first Handshake packet.
	// After this, all packets with a different source connection have to be ignored.
	destConnID := s.connIDManager.Get()
	if s.receivedFirstPacket && hdr.IsLongHeader && !hdr.SrcConnectionID.Equal(destConnID) {
		s.logger.Debugf("Dropping packet with unexpected source connection ID: %s (expected %s)", hdr.SrcConnectionID, destConnID)
		return false
	}
	// drop 0-RTT packets
	if hdr.Type == protocol.PacketType0RTT {
		return false
	}

	packet, err := s.unpacker.Unpack(hdr, p.rcvTime, p.data)
	if err != nil {
		switch err {
		case handshake.ErrKeysDropped:
			s.logger.Debugf("Dropping packet because we already dropped the keys.")
		case handshake.ErrOpenerNotYetAvailable:
			// Sealer for this encryption level not yet available.
			// Try again later.
			wasQueued = true
			s.tryQueueingUndecryptablePacket(p)
		case wire.ErrInvalidReservedBits:
			s.closeLocal(qerr.Error(qerr.ProtocolViolation, err.Error()))
		default:
			// This might be a packet injected by an attacker.
			// Drop it.
			s.logger.Debugf("Dropping packet that could not be unpacked. Error: %s", err)
		}
		return false
	}

	if s.logger.Debug() {
		s.logger.Debugf("<- Reading packet %#x (%d bytes) for connection %s, %s", packet.packetNumber, len(p.data), hdr.DestConnectionID, packet.encryptionLevel)
		packet.hdr.Log(s.logger)
	}

	if err := s.handleUnpackedPacket(packet, p.rcvTime); err != nil {
		s.closeLocal(err)
		return false
	}
	return true
}

func (s *session) handleRetryPacket(hdr *wire.Header) bool /* was this a valid Retry */ {
	if s.perspective == protocol.PerspectiveServer {
		s.logger.Debugf("Ignoring Retry.")
		return false
	}
	if s.receivedFirstPacket {
		s.logger.Debugf("Ignoring Retry, since we already received a packet.")
		return false
	}
	(&wire.ExtendedHeader{Header: *hdr}).Log(s.logger)
	destConnID := s.connIDManager.Get()
	if !hdr.OrigDestConnectionID.Equal(destConnID) {
		s.logger.Debugf("Ignoring spoofed Retry. Original Destination Connection ID: %s, expected: %s", hdr.OrigDestConnectionID, destConnID)
		return false
	}
	if hdr.SrcConnectionID.Equal(destConnID) {
		s.logger.Debugf("Ignoring Retry, since the server didn't change the Source Connection ID.")
		return false
	}
	// If a token is already set, this means that we already received a Retry from the server.
	// Ignore this Retry packet.
	if s.receivedRetry {
		s.logger.Debugf("Ignoring Retry, since a Retry was already received.")
		return false
	}
	s.logger.Debugf("<- Received Retry")
	s.logger.Debugf("Switching destination connection ID to: %s", hdr.SrcConnectionID)
	s.origDestConnID = destConnID
	newDestConnID := hdr.SrcConnectionID
	s.receivedRetry = true
	if err := s.sentPacketHandler.ResetForRetry(); err != nil {
		s.closeLocal(err)
		return false
	}
	s.cryptoStreamHandler.ChangeConnectionID(newDestConnID)
	s.packer.SetToken(hdr.Token)
	s.connIDManager.ChangeInitialConnID(newDestConnID)
	s.scheduleSending()
	return true
}

func (s *session) handleUnpackedPacket(packet *unpackedPacket, rcvTime time.Time) error {
	if len(packet.data) == 0 {
		return qerr.Error(qerr.ProtocolViolation, "empty packet")
	}

	// The server can change the source connection ID with the first Handshake packet.
	if s.perspective == protocol.PerspectiveClient && !s.receivedFirstPacket && packet.hdr.IsLongHeader && !packet.hdr.SrcConnectionID.Equal(s.connIDManager.Get()) {
		s.logger.Debugf("Received first packet. Switching destination connection ID to: %s", packet.hdr.SrcConnectionID)
		s.connIDManager.ChangeInitialConnID(packet.hdr.SrcConnectionID)
	}

	s.receivedFirstPacket = true
	s.lastPacketReceivedTime = rcvTime
	s.firstAckElicitingPacketAfterIdleSentTime = time.Time{}
	s.keepAlivePingSent = false

	// Only used for tracing.
	// If we're not tracing, this slice will always remain empty.
	var frames []wire.Frame
	var transportState *quictrace.TransportState

	r := bytes.NewReader(packet.data)
	var isAckEliciting bool
	for {
		frame, err := s.frameParser.ParseNext(r, packet.encryptionLevel)
		if err != nil {
			return err
		}
		if frame == nil {
			break
		}
		if ackhandler.IsFrameAckEliciting(frame) {
			isAckEliciting = true
		}
		if s.traceCallback != nil {
			frames = append(frames, frame)
		}
		if err := s.handleFrame(frame, packet.packetNumber, packet.encryptionLevel); err != nil {
			return err
		}
	}

	if s.traceCallback != nil {
		transportState = s.sentPacketHandler.GetStats()
		s.traceCallback(quictrace.Event{
			Time:            time.Now(),
			EventType:       quictrace.PacketReceived,
			TransportState:  transportState,
			EncryptionLevel: packet.encryptionLevel,
			PacketNumber:    packet.packetNumber,
			PacketSize:      protocol.ByteCount(len(packet.data)),
			Frames:          frames,
		})
	}

	s.receivedPacketHandler.ReceivedPacket(packet.packetNumber, packet.encryptionLevel, rcvTime, isAckEliciting)
	return nil
}

func (s *session) handleFrame(f wire.Frame, pn protocol.PacketNumber, encLevel protocol.EncryptionLevel) error {
	var err error
	wire.LogFrame(s.logger, f, false)
	switch frame := f.(type) {
	case *wire.CryptoFrame:
		err = s.handleCryptoFrame(frame, encLevel)
	case *wire.StreamFrame:
		err = s.handleStreamFrame(frame)
	case *wire.AckFrame:
		err = s.handleAckFrame(frame, pn, encLevel)
	case *wire.ConnectionCloseFrame:
		s.handleConnectionCloseFrame(frame)
	case *wire.ResetStreamFrame:
		err = s.handleResetStreamFrame(frame)
	case *wire.MaxDataFrame:
		s.handleMaxDataFrame(frame)
	case *wire.MaxStreamDataFrame:
		err = s.handleMaxStreamDataFrame(frame)
	case *wire.MaxStreamsFrame:
		err = s.handleMaxStreamsFrame(frame)
	case *wire.DataBlockedFrame:
	case *wire.StreamDataBlockedFrame:
	case *wire.StreamsBlockedFrame:
	case *wire.StopSendingFrame:
		err = s.handleStopSendingFrame(frame)
	case *wire.PingFrame:
	case *wire.PathChallengeFrame:
		s.handlePathChallengeFrame(frame)
	case *wire.PathResponseFrame:
		// since we don't send PATH_CHALLENGEs, we don't expect PATH_RESPONSEs
		err = errors.New("unexpected PATH_RESPONSE frame")
	case *wire.NewTokenFrame:
		err = s.handleNewTokenFrame(frame)
	case *wire.NewConnectionIDFrame:
		err = s.handleNewConnectionIDFrame(frame)
	case *wire.RetireConnectionIDFrame:
		err = s.handleRetireConnectionIDFrame(frame)
	default:
		err = fmt.Errorf("unexpected frame type: %s", reflect.ValueOf(&frame).Elem().Type().Name())
	}
	return err
}

// handlePacket is called by the server with a new packet
func (s *session) handlePacket(p *receivedPacket) {
	// Discard packets once the amount of queued packets is larger than
	// the channel size, protocol.MaxSessionUnprocessedPackets
	select {
	case s.receivedPackets <- p:
	default:
	}
}

func (s *session) handleConnectionCloseFrame(frame *wire.ConnectionCloseFrame) {
	var e error
	if frame.IsApplicationError {
		e = qerr.ApplicationError(frame.ErrorCode, frame.ReasonPhrase)
	} else {
		e = qerr.Error(frame.ErrorCode, frame.ReasonPhrase)
	}
	s.closeRemote(e)
}

func (s *session) handleCryptoFrame(frame *wire.CryptoFrame, encLevel protocol.EncryptionLevel) error {
	encLevelChanged, err := s.cryptoStreamManager.HandleCryptoFrame(frame, encLevel)
	if err != nil {
		return err
	}
	s.logger.Debugf("Handled crypto frame at level %s. encLevelChanged: %t", encLevel, encLevelChanged)
	if encLevelChanged {
		s.tryDecryptingQueuedPackets()
	}
	return nil
}

func (s *session) handleStreamFrame(frame *wire.StreamFrame) error {
	str, err := s.streamsMap.GetOrOpenReceiveStream(frame.StreamID)
	if err != nil {
		return err
	}
	if str == nil {
		// Stream is closed and already garbage collected
		// ignore this StreamFrame
		return nil
	}
	return str.handleStreamFrame(frame)
}

func (s *session) handleMaxDataFrame(frame *wire.MaxDataFrame) {
	s.connFlowController.UpdateSendWindow(frame.ByteOffset)
}

func (s *session) handleMaxStreamDataFrame(frame *wire.MaxStreamDataFrame) error {
	str, err := s.streamsMap.GetOrOpenSendStream(frame.StreamID)
	if err != nil {
		return err
	}
	if str == nil {
		// stream is closed and already garbage collected
		return nil
	}
	str.handleMaxStreamDataFrame(frame)
	return nil
}

func (s *session) handleMaxStreamsFrame(frame *wire.MaxStreamsFrame) error {
	return s.streamsMap.HandleMaxStreamsFrame(frame)
}

func (s *session) handleResetStreamFrame(frame *wire.ResetStreamFrame) error {
	str, err := s.streamsMap.GetOrOpenReceiveStream(frame.StreamID)
	if err != nil {
		return err
	}
	if str == nil {
		// stream is closed and already garbage collected
		return nil
	}
	return str.handleResetStreamFrame(frame)
}

func (s *session) handleStopSendingFrame(frame *wire.StopSendingFrame) error {
	str, err := s.streamsMap.GetOrOpenSendStream(frame.StreamID)
	if err != nil {
		return err
	}
	if str == nil {
		// stream is closed and already garbage collected
		return nil
	}
	str.handleStopSendingFrame(frame)
	return nil
}

func (s *session) handlePathChallengeFrame(frame *wire.PathChallengeFrame) {
	s.queueControlFrame(&wire.PathResponseFrame{Data: frame.Data})
}

func (s *session) handleNewTokenFrame(frame *wire.NewTokenFrame) error {
	if s.perspective == protocol.PerspectiveServer {
		return qerr.Error(qerr.ProtocolViolation, "Received NEW_TOKEN frame from the client.")
	}
	if s.config.TokenStore != nil {
		s.config.TokenStore.Put(s.tokenStoreKey, &ClientToken{data: frame.Token})
	}
	return nil
}

func (s *session) handleNewConnectionIDFrame(f *wire.NewConnectionIDFrame) error {
	return s.connIDManager.Add(f)
}

func (s *session) handleRetireConnectionIDFrame(f *wire.RetireConnectionIDFrame) error {
	return s.connIDGenerator.Retire(f.SequenceNumber)
}

func (s *session) handleAckFrame(frame *wire.AckFrame, pn protocol.PacketNumber, encLevel protocol.EncryptionLevel) error {
	if err := s.sentPacketHandler.ReceivedAck(frame, pn, encLevel, s.lastPacketReceivedTime); err != nil {
		return err
	}
	if encLevel == protocol.Encryption1RTT {
		s.receivedPacketHandler.IgnoreBelow(s.sentPacketHandler.GetLowestPacketNotConfirmedAcked())
		s.cryptoStreamHandler.SetLargest1RTTAcked(frame.LargestAcked())
	}
	return nil
}

// closeLocal closes the session and send a CONNECTION_CLOSE containing the error
func (s *session) closeLocal(e error) {
	s.closeOnce.Do(func() {
		if e == nil {
			s.logger.Infof("Closing session.")
		} else {
			s.logger.Errorf("Closing session with error: %s", e)
		}
		s.closeChan <- closeError{err: e, immediate: false, remote: false}
	})
}

// destroy closes the session without sending the error on the wire
func (s *session) destroy(e error) {
	s.destroyImpl(e)
	<-s.ctx.Done()
}

func (s *session) destroyImpl(e error) {
	s.closeOnce.Do(func() {
		if nerr, ok := e.(net.Error); ok && nerr.Timeout() {
			s.logger.Errorf("Destroying session %s: %s", s.connIDManager.Get(), e)
		} else {
			s.logger.Errorf("Destroying session %s with error: %s", s.connIDManager.Get(), e)
		}
		s.closeChan <- closeError{err: e, immediate: true, remote: false}
	})
}

// closeForRecreating closes the session in order to recreate it immediately afterwards
// It returns the first packet number that should be used in the new session.
func (s *session) closeForRecreating() protocol.PacketNumber {
	s.destroy(errCloseForRecreating)
	nextPN, _ := s.sentPacketHandler.PeekPacketNumber(protocol.EncryptionInitial)
	return nextPN
}

func (s *session) closeRemote(e error) {
	s.closeOnce.Do(func() {
		s.logger.Errorf("Peer closed session with error: %s", e)
		s.closeChan <- closeError{err: e, immediate: true, remote: true}
	})
}

// Close the connection. It sends a qerr.NoError.
// It waits until the run loop has stopped before returning
func (s *session) Close() error {
	s.closeLocal(nil)
	<-s.ctx.Done()
	return nil
}

func (s *session) CloseWithError(code protocol.ApplicationErrorCode, desc string) error {
	s.closeLocal(qerr.ApplicationError(qerr.ErrorCode(code), desc))
	<-s.ctx.Done()
	return nil
}

func (s *session) handleCloseError(closeErr closeError) {
	if closeErr.err == nil {
		closeErr.err = qerr.ApplicationError(0, "")
	}

	var quicErr *qerr.QuicError
	var ok bool
	if quicErr, ok = closeErr.err.(*qerr.QuicError); !ok {
		quicErr = qerr.ToQuicError(closeErr.err)
	}

	s.streamsMap.CloseWithError(quicErr)
	s.connIDManager.Close()

	// If this is a remote close we're done here
	if closeErr.remote {
		s.connIDGenerator.ReplaceWithClosed(newClosedRemoteSession(s.perspective))
		return
	}
	if closeErr.immediate {
		s.connIDGenerator.RemoveAll()
		return
	}
	connClosePacket, err := s.sendConnectionClose(quicErr)
	if err != nil {
		s.logger.Debugf("Error sending CONNECTION_CLOSE: %s", err)
	}
	cs := newClosedLocalSession(s.conn, connClosePacket, s.perspective, s.logger)
	s.connIDGenerator.ReplaceWithClosed(cs)
}

func (s *session) dropEncryptionLevel(encLevel protocol.EncryptionLevel) {
	s.sentPacketHandler.DropPackets(encLevel)
	s.receivedPacketHandler.DropPackets(encLevel)
}

func (s *session) processTransportParameters(data []byte) {
	var params *handshake.TransportParameters
	var err error
	switch s.perspective {
	case protocol.PerspectiveClient:
		params, err = s.processTransportParametersForClient(data)
	case protocol.PerspectiveServer:
		params, err = s.processTransportParametersForServer(data)
	}
	if err != nil {
		s.closeLocal(err)
		return
	}
	s.logger.Debugf("Received Transport Parameters: %s", params)
	s.peerParams = params
	if err := s.streamsMap.UpdateLimits(params); err != nil {
		s.closeLocal(err)
		return
	}
	s.packer.HandleTransportParameters(params)
	s.frameParser.SetAckDelayExponent(params.AckDelayExponent)
	s.connFlowController.UpdateSendWindow(params.InitialMaxData)
	s.rttStats.SetMaxAckDelay(params.MaxAckDelay)
	s.connIDGenerator.SetMaxActiveConnIDs(params.ActiveConnectionIDLimit)
	if params.StatelessResetToken != nil {
		s.connIDManager.SetStatelessResetToken(*params.StatelessResetToken)
	}
	// On the server side, the early session is ready as soon as we processed
	// the client's transport parameters.
	close(s.earlySessionReadyChan)
}

func (s *session) processTransportParametersForClient(data []byte) (*handshake.TransportParameters, error) {
	params := &handshake.TransportParameters{}
	if err := params.Unmarshal(data, s.perspective.Opposite()); err != nil {
		return nil, err
	}

	// check the Retry token
	if !params.OriginalConnectionID.Equal(s.origDestConnID) {
		return nil, qerr.Error(qerr.TransportParameterError, fmt.Sprintf("expected original_connection_id to equal %s, is %s", s.origDestConnID, params.OriginalConnectionID))
	}

	return params, nil
}

func (s *session) processTransportParametersForServer(data []byte) (*handshake.TransportParameters, error) {
	params := &handshake.TransportParameters{}
	if err := params.Unmarshal(data, s.perspective.Opposite()); err != nil {
		return nil, err
	}
	return params, nil
}

func (s *session) sendPackets() error {
	s.pacingDeadline = time.Time{}

	sendMode := s.sentPacketHandler.SendMode()
	if sendMode == ackhandler.SendNone { // shortcut: return immediately if there's nothing to send
		return nil
	}

	numPackets := s.sentPacketHandler.ShouldSendNumPackets()
	var numPacketsSent int
sendLoop:
	for {
		switch sendMode {
		case ackhandler.SendNone:
			break sendLoop
		case ackhandler.SendAck:
			// If we already sent packets, and the send mode switches to SendAck,
			// we've just become congestion limited.
			// There's no need to try to send an ACK at this moment.
			if numPacketsSent > 0 {
				return nil
			}
			// We can at most send a single ACK only packet.
			// There will only be a new ACK after receiving new packets.
			// SendAck is only returned when we're congestion limited, so we don't need to set the pacingt timer.
			return s.maybeSendAckOnlyPacket()
		case ackhandler.SendPTO:
			if err := s.sendProbePacket(); err != nil {
				return err
			}
			numPacketsSent++
		case ackhandler.SendAny:
			sentPacket, err := s.sendPacket()
			if err != nil {
				return err
			}
			if !sentPacket {
				break sendLoop
			}
			numPacketsSent++
		default:
			return fmt.Errorf("BUG: invalid send mode %d", sendMode)
		}
		if numPacketsSent >= numPackets {
			break
		}
		sendMode = s.sentPacketHandler.SendMode()
	}
	// Only start the pacing timer if we sent as many packets as we were allowed.
	// There will probably be more to send when calling sendPacket again.
	if numPacketsSent == numPackets {
		s.pacingDeadline = s.sentPacketHandler.TimeUntilSend()
	}
	return nil
}

func (s *session) maybeSendAckOnlyPacket() error {
	packet, err := s.packer.MaybePackAckPacket()
	if err != nil {
		return err
	}
	if packet == nil {
		return nil
	}
	s.sentPacketHandler.SentPacket(packet.ToAckHandlerPacket(s.retransmissionQueue))
	s.sendPackedPacket(packet)
	return nil
}

func (s *session) sendProbePacket() error {
	// Queue probe packets until we actually send out a packet.
	for {
		if wasQueued := s.sentPacketHandler.QueueProbePacket(); !wasQueued {
			break
		}
		sent, err := s.sendPacket()
		if err != nil {
			return err
		}
		if sent {
			return nil
		}
	}
	// If there is nothing else to queue, make sure we send out something.
	s.framer.QueueControlFrame(&wire.PingFrame{})
	_, err := s.sendPacket()
	return err
}

func (s *session) sendPacket() (bool, error) {
	if isBlocked, offset := s.connFlowController.IsNewlyBlocked(); isBlocked {
		s.framer.QueueControlFrame(&wire.DataBlockedFrame{DataLimit: offset})
	}
	s.windowUpdateQueue.QueueAll()

	packet, err := s.packer.PackPacket()
	if err != nil || packet == nil {
		return false, err
	}
	s.sentPacketHandler.SentPacket(packet.ToAckHandlerPacket(s.retransmissionQueue))
	s.sendPackedPacket(packet)
	return true, nil
}

func (s *session) sendPackedPacket(packet *packedPacket) {
	if s.firstAckElicitingPacketAfterIdleSentTime.IsZero() && packet.IsAckEliciting() {
		s.firstAckElicitingPacketAfterIdleSentTime = time.Now()
	}
	if s.traceCallback != nil {
		s.traceCallback(quictrace.Event{
			Time:            time.Now(),
			EventType:       quictrace.PacketSent,
			TransportState:  s.sentPacketHandler.GetStats(),
			EncryptionLevel: packet.EncryptionLevel(),
			PacketNumber:    packet.header.PacketNumber,
			PacketSize:      protocol.ByteCount(len(packet.raw)),
			// TODO: trace frames
			// Frames:          packet.frames,
		})
	}
	s.logPacket(packet)
	s.connIDManager.SentPacket()
	s.sendQueue.Send(packet)
}

func (s *session) sendConnectionClose(quicErr *qerr.QuicError) ([]byte, error) {
	var reason string
	// don't send details of crypto errors
	if !quicErr.IsCryptoError() {
		reason = quicErr.ErrorMessage
	}
	packet, err := s.packer.PackConnectionClose(&wire.ConnectionCloseFrame{
		IsApplicationError: quicErr.IsApplicationError(),
		ErrorCode:          quicErr.ErrorCode,
		FrameType:          quicErr.FrameType,
		ReasonPhrase:       reason,
	})
	if err != nil {
		return nil, err
	}
	s.logPacket(packet)
	return packet.raw, s.conn.Write(packet.raw)
}

func (s *session) logPacket(packet *packedPacket) {
	if !s.logger.Debug() {
		// We don't need to allocate the slices for calling the format functions
		return
	}
	s.logger.Debugf("-> Sending packet 0x%x (%d bytes) for connection %s, %s", packet.header.PacketNumber, len(packet.raw), s.logID, packet.EncryptionLevel())
	packet.header.Log(s.logger)
	if packet.ack != nil {
		wire.LogFrame(s.logger, packet.ack, true)
	}
	for _, frame := range packet.frames {
		wire.LogFrame(s.logger, frame.Frame, true)
	}
}

// AcceptStream returns the next stream openend by the peer
func (s *session) AcceptStream(ctx context.Context) (Stream, error) {
	return s.streamsMap.AcceptStream(ctx)
}

func (s *session) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	return s.streamsMap.AcceptUniStream(ctx)
}

// OpenStream opens a stream
func (s *session) OpenStream() (Stream, error) {
	return s.streamsMap.OpenStream()
}

func (s *session) OpenStreamSync(ctx context.Context) (Stream, error) {
	return s.streamsMap.OpenStreamSync(ctx)
}

func (s *session) OpenUniStream() (SendStream, error) {
	return s.streamsMap.OpenUniStream()
}

func (s *session) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	return s.streamsMap.OpenUniStreamSync(ctx)
}

func (s *session) newFlowController(id protocol.StreamID) flowcontrol.StreamFlowController {
	var initialSendWindow protocol.ByteCount
	if s.peerParams != nil {
		if id.Type() == protocol.StreamTypeUni {
			initialSendWindow = s.peerParams.InitialMaxStreamDataUni
		} else {
			if id.InitiatedBy() == s.perspective {
				initialSendWindow = s.peerParams.InitialMaxStreamDataBidiRemote
			} else {
				initialSendWindow = s.peerParams.InitialMaxStreamDataBidiLocal
			}
		}
	}
	return flowcontrol.NewStreamFlowController(
		id,
		s.connFlowController,
		protocol.InitialMaxStreamData,
		protocol.ByteCount(s.config.MaxReceiveStreamFlowControlWindow),
		initialSendWindow,
		s.onHasStreamWindowUpdate,
		s.rttStats,
		s.logger,
	)
}

// scheduleSending signals that we have data for sending
func (s *session) scheduleSending() {
	select {
	case s.sendingScheduled <- struct{}{}:
	default:
	}
}

func (s *session) tryQueueingUndecryptablePacket(p *receivedPacket) {
	if s.handshakeComplete {
		s.logger.Debugf("Received undecryptable packet from %s after the handshake (%d bytes)", p.remoteAddr.String(), len(p.data))
		return
	}
	if len(s.undecryptablePackets)+1 > protocol.MaxUndecryptablePackets {
		s.logger.Infof("Dropping undecrytable packet (%d bytes). Undecryptable packet queue full.", len(p.data))
		return
	}
	s.logger.Infof("Queueing packet (%d bytes) for later decryption", len(p.data))
	s.undecryptablePackets = append(s.undecryptablePackets, p)
}

func (s *session) tryDecryptingQueuedPackets() {
	for _, p := range s.undecryptablePackets {
		s.handlePacket(p)
	}
	s.undecryptablePackets = s.undecryptablePackets[:0]
}

func (s *session) queueControlFrame(f wire.Frame) {
	s.framer.QueueControlFrame(f)
	s.scheduleSending()
}

func (s *session) onHasStreamWindowUpdate(id protocol.StreamID) {
	s.windowUpdateQueue.AddStream(id)
	s.scheduleSending()
}

func (s *session) onHasConnectionWindowUpdate() {
	s.windowUpdateQueue.AddConnection()
	s.scheduleSending()
}

func (s *session) onHasStreamData(id protocol.StreamID) {
	s.framer.AddActiveStream(id)
	s.scheduleSending()
}

func (s *session) onStreamCompleted(id protocol.StreamID) {
	if err := s.streamsMap.DeleteStream(id); err != nil {
		s.closeLocal(err)
	}
}

func (s *session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *session) getPerspective() protocol.Perspective {
	return s.perspective
}

func (s *session) GetVersion() protocol.VersionNumber {
	return s.version
}
