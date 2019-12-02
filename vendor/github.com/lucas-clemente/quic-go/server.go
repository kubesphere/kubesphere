package quic

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lucas-clemente/quic-go/internal/handshake"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

// packetHandler handles packets
type packetHandler interface {
	handlePacket(*receivedPacket)
	io.Closer
	destroy(error)
	getPerspective() protocol.Perspective
}

type unknownPacketHandler interface {
	handlePacket(*receivedPacket)
	setCloseError(error)
}

type packetHandlerManager interface {
	io.Closer
	SetServer(unknownPacketHandler)
	CloseServer()
	sessionRunner
}

type quicSession interface {
	EarlySession
	earlySessionReady() <-chan struct{}
	handlePacket(*receivedPacket)
	GetVersion() protocol.VersionNumber
	getPerspective() protocol.Perspective
	run() error
	destroy(error)
	closeForRecreating() protocol.PacketNumber
	closeRemote(error)
}

// A Listener of QUIC
type baseServer struct {
	mutex sync.Mutex

	acceptEarlySessions bool

	tlsConf *tls.Config
	config  *Config

	conn net.PacketConn
	// If the server is started with ListenAddr, we create a packet conn.
	// If it is started with Listen, we take a packet conn as a parameter.
	createdPacketConn bool

	tokenGenerator *handshake.TokenGenerator

	sessionHandler packetHandlerManager

	// set as a member, so they can be set in the tests
	newSession func(connection, sessionRunner, protocol.ConnectionID /* original connection ID */, protocol.ConnectionID /* client dest connection ID */, protocol.ConnectionID /* destination connection ID */, protocol.ConnectionID /* source connection ID */, *Config, *tls.Config, *handshake.TokenGenerator, utils.Logger, protocol.VersionNumber) quicSession

	serverError error
	errorChan   chan struct{}
	closed      bool

	sessionQueue    chan quicSession
	sessionQueueLen int32 // to be used as an atomic

	logger utils.Logger
}

var _ Listener = &baseServer{}
var _ unknownPacketHandler = &baseServer{}

type earlyServer struct{ *baseServer }

var _ EarlyListener = &earlyServer{}

func (s *earlyServer) Accept(ctx context.Context) (EarlySession, error) {
	return s.baseServer.accept(ctx)
}

// ListenAddr creates a QUIC server listening on a given address.
// The tls.Config must not be nil and must contain a certificate configuration.
// The quic.Config may be nil, in that case the default values will be used.
func ListenAddr(addr string, tlsConf *tls.Config, config *Config) (Listener, error) {
	return listenAddr(addr, tlsConf, config, false)
}

// ListenAddrEarly works like ListenAddr, but it returns sessions before the handshake completes.
func ListenAddrEarly(addr string, tlsConf *tls.Config, config *Config) (EarlyListener, error) {
	s, err := listenAddr(addr, tlsConf, config, true)
	if err != nil {
		return nil, err
	}
	return &earlyServer{s}, nil
}

func listenAddr(addr string, tlsConf *tls.Config, config *Config, acceptEarly bool) (*baseServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	serv, err := listen(conn, tlsConf, config, acceptEarly)
	if err != nil {
		return nil, err
	}
	serv.createdPacketConn = true
	return serv, nil
}

// Listen listens for QUIC connections on a given net.PacketConn.
// A single net.PacketConn only be used for a single call to Listen.
// The PacketConn can be used for simultaneous calls to Dial.
// QUIC connection IDs are used for demultiplexing the different connections.
// The tls.Config must not be nil and must contain a certificate configuration.
// Furthermore, it must define an application control (using NextProtos).
// The quic.Config may be nil, in that case the default values will be used.
func Listen(conn net.PacketConn, tlsConf *tls.Config, config *Config) (Listener, error) {
	return listen(conn, tlsConf, config, false)
}

// ListenEarly works like Listen, but it returns sessions before the handshake completes.
func ListenEarly(conn net.PacketConn, tlsConf *tls.Config, config *Config) (EarlyListener, error) {
	s, err := listen(conn, tlsConf, config, true)
	if err != nil {
		return nil, err
	}
	s.acceptEarlySessions = true
	return &earlyServer{s}, nil
}

func listen(conn net.PacketConn, tlsConf *tls.Config, config *Config, acceptEarly bool) (*baseServer, error) {
	if tlsConf == nil {
		return nil, errors.New("quic: tls.Config not set")
	}
	config = populateServerConfig(config)
	for _, v := range config.Versions {
		if !protocol.IsValidVersion(v) {
			return nil, fmt.Errorf("%s is not a valid QUIC version", v)
		}
	}

	sessionHandler, err := getMultiplexer().AddConn(conn, config.ConnectionIDLength, config.StatelessResetKey)
	if err != nil {
		return nil, err
	}
	tokenGenerator, err := handshake.NewTokenGenerator()
	if err != nil {
		return nil, err
	}
	s := &baseServer{
		conn:                conn,
		tlsConf:             tlsConf,
		config:              config,
		tokenGenerator:      tokenGenerator,
		sessionHandler:      sessionHandler,
		sessionQueue:        make(chan quicSession),
		errorChan:           make(chan struct{}),
		newSession:          newSession,
		logger:              utils.DefaultLogger.WithPrefix("server"),
		acceptEarlySessions: acceptEarly,
	}
	sessionHandler.SetServer(s)
	s.logger.Debugf("Listening for %s connections on %s", conn.LocalAddr().Network(), conn.LocalAddr().String())
	return s, nil
}

var defaultAcceptToken = func(clientAddr net.Addr, token *Token) bool {
	if token == nil {
		return false
	}
	validity := protocol.TokenValidity
	if token.IsRetryToken {
		validity = protocol.RetryTokenValidity
	}
	if time.Now().After(token.SentTime.Add(validity)) {
		return false
	}
	var sourceAddr string
	if udpAddr, ok := clientAddr.(*net.UDPAddr); ok {
		sourceAddr = udpAddr.IP.String()
	} else {
		sourceAddr = clientAddr.String()
	}
	return sourceAddr == token.RemoteAddr
}

// populateServerConfig populates fields in the quic.Config with their default values, if none are set
// it may be called with nil
func populateServerConfig(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}
	versions := config.Versions
	if len(versions) == 0 {
		versions = protocol.SupportedVersions
	}

	verifyToken := defaultAcceptToken
	if config.AcceptToken != nil {
		verifyToken = config.AcceptToken
	}

	handshakeTimeout := protocol.DefaultHandshakeTimeout
	if config.HandshakeTimeout != 0 {
		handshakeTimeout = config.HandshakeTimeout
	}
	idleTimeout := protocol.DefaultIdleTimeout
	if config.IdleTimeout != 0 {
		idleTimeout = config.IdleTimeout
	}

	maxReceiveStreamFlowControlWindow := config.MaxReceiveStreamFlowControlWindow
	if maxReceiveStreamFlowControlWindow == 0 {
		maxReceiveStreamFlowControlWindow = protocol.DefaultMaxReceiveStreamFlowControlWindow
	}
	maxReceiveConnectionFlowControlWindow := config.MaxReceiveConnectionFlowControlWindow
	if maxReceiveConnectionFlowControlWindow == 0 {
		maxReceiveConnectionFlowControlWindow = protocol.DefaultMaxReceiveConnectionFlowControlWindow
	}
	maxIncomingStreams := config.MaxIncomingStreams
	if maxIncomingStreams == 0 {
		maxIncomingStreams = protocol.DefaultMaxIncomingStreams
	} else if maxIncomingStreams < 0 {
		maxIncomingStreams = 0
	}
	maxIncomingUniStreams := config.MaxIncomingUniStreams
	if maxIncomingUniStreams == 0 {
		maxIncomingUniStreams = protocol.DefaultMaxIncomingUniStreams
	} else if maxIncomingUniStreams < 0 {
		maxIncomingUniStreams = 0
	}
	connIDLen := config.ConnectionIDLength
	if connIDLen == 0 {
		connIDLen = protocol.DefaultConnectionIDLength
	}

	return &Config{
		Versions:                              versions,
		HandshakeTimeout:                      handshakeTimeout,
		IdleTimeout:                           idleTimeout,
		AcceptToken:                           verifyToken,
		KeepAlive:                             config.KeepAlive,
		MaxReceiveStreamFlowControlWindow:     maxReceiveStreamFlowControlWindow,
		MaxReceiveConnectionFlowControlWindow: maxReceiveConnectionFlowControlWindow,
		MaxIncomingStreams:                    maxIncomingStreams,
		MaxIncomingUniStreams:                 maxIncomingUniStreams,
		ConnectionIDLength:                    connIDLen,
		StatelessResetKey:                     config.StatelessResetKey,
		QuicTracer:                            config.QuicTracer,
	}
}

// Accept returns sessions that already completed the handshake.
// It is only valid if acceptEarlySessions is false.
func (s *baseServer) Accept(ctx context.Context) (Session, error) {
	return s.accept(ctx)
}

func (s *baseServer) accept(ctx context.Context) (quicSession, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case sess := <-s.sessionQueue:
		atomic.AddInt32(&s.sessionQueueLen, -1)
		return sess, nil
	case <-s.errorChan:
		return nil, s.serverError
	}
}

// Close the server
func (s *baseServer) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.sessionHandler.CloseServer()
	if s.serverError == nil {
		s.serverError = errors.New("server closed")
	}
	var err error
	// If the server was started with ListenAddr, we created the packet conn.
	// We need to close it in order to make the go routine reading from that conn return.
	if s.createdPacketConn {
		err = s.sessionHandler.Close()
	}
	s.closed = true
	close(s.errorChan)
	return err
}

func (s *baseServer) setCloseError(e error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.serverError = e
	close(s.errorChan)
}

// Addr returns the server's network address
func (s *baseServer) Addr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *baseServer) handlePacket(p *receivedPacket) {
	go func() {
		if shouldReleaseBuffer := s.handlePacketImpl(p); !shouldReleaseBuffer {
			p.buffer.Release()
		}
	}()
}

func (s *baseServer) handlePacketImpl(p *receivedPacket) bool /* was the packet passed on to a session */ {
	if len(p.data) < protocol.MinInitialPacketSize {
		s.logger.Debugf("Dropping a packet that is too small to be a valid Initial (%d bytes)", len(p.data))
		return false
	}
	// If we're creating a new session, the packet will be passed to the session.
	// The header will then be parsed again.
	hdr, _, _, err := wire.ParsePacket(p.data, s.config.ConnectionIDLength)
	if err != nil {
		s.logger.Debugf("Error parsing packet: %s", err)
		return false
	}
	// Short header packets should never end up here in the first place
	if !hdr.IsLongHeader {
		return false
	}
	// send a Version Negotiation Packet if the client is speaking a different protocol version
	if !protocol.IsSupportedVersion(s.config.Versions, hdr.Version) {
		s.sendVersionNegotiationPacket(p, hdr)
		return false
	}
	if hdr.IsLongHeader && hdr.Type != protocol.PacketTypeInitial {
		// Drop long header packets.
		// There's litte point in sending a Stateless Reset, since the client
		// might not have received the token yet.
		return false
	}

	s.logger.Debugf("<- Received Initial packet.")

	sess, err := s.handleInitialImpl(p, hdr)
	if err != nil {
		s.logger.Errorf("Error occurred handling initial packet: %s", err)
		return false
	}
	if sess == nil { // a retry was done, or the connection attempt was rejected
		return false
	}
	// Don't put the packet buffer back if a new session was created.
	// The session will handle the packet and take of that.
	return true
}

func (s *baseServer) handleInitialImpl(p *receivedPacket, hdr *wire.Header) (quicSession, error) {
	if len(hdr.Token) == 0 && hdr.DestConnectionID.Len() < protocol.MinConnectionIDLenInitial {
		return nil, errors.New("too short connection ID")
	}

	var token *Token
	var origDestConnectionID protocol.ConnectionID
	if len(hdr.Token) > 0 {
		c, err := s.tokenGenerator.DecodeToken(hdr.Token)
		if err == nil {
			token = &Token{
				IsRetryToken: c.IsRetryToken,
				RemoteAddr:   c.RemoteAddr,
				SentTime:     c.SentTime,
			}
			origDestConnectionID = c.OriginalDestConnectionID
		}
	}
	if !s.config.AcceptToken(p.remoteAddr, token) {
		// Log the Initial packet now.
		// If no Retry is sent, the packet will be logged by the session.
		(&wire.ExtendedHeader{Header: *hdr}).Log(s.logger)
		return nil, s.sendRetry(p.remoteAddr, hdr)
	}

	if queueLen := atomic.LoadInt32(&s.sessionQueueLen); queueLen >= protocol.MaxAcceptQueueSize {
		s.logger.Debugf("Rejecting new connection. Server currently busy. Accept queue length: %d (max %d)", queueLen, protocol.MaxAcceptQueueSize)
		return nil, s.sendServerBusy(p.remoteAddr, hdr)
	}

	connID, err := protocol.GenerateConnectionID(s.config.ConnectionIDLength)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("Changing connection ID to %s.", connID)
	sess := s.createNewSession(
		p.remoteAddr,
		origDestConnectionID,
		hdr.DestConnectionID,
		hdr.SrcConnectionID,
		connID,
		hdr.Version,
	)
	sess.handlePacket(p)
	return sess, nil
}

func (s *baseServer) createNewSession(
	remoteAddr net.Addr,
	origDestConnID protocol.ConnectionID,
	clientDestConnID protocol.ConnectionID,
	destConnID protocol.ConnectionID,
	srcConnID protocol.ConnectionID,
	version protocol.VersionNumber,
) quicSession {
	sess := s.newSession(
		&conn{pconn: s.conn, currentAddr: remoteAddr},
		s.sessionHandler,
		origDestConnID,
		clientDestConnID,
		destConnID,
		srcConnID,
		s.config,
		s.tlsConf,
		s.tokenGenerator,
		s.logger,
		version,
	)
	go sess.run()
	go s.handleNewSession(sess)
	return sess
}

func (s *baseServer) handleNewSession(sess quicSession) {
	sessCtx := sess.Context()
	if s.acceptEarlySessions {
		// wait until the early session is ready (or the handshake fails)
		select {
		case <-sess.earlySessionReady():
		case <-sessCtx.Done():
			return
		}
	} else {
		// wait until the handshake is complete (or fails)
		select {
		case <-sess.HandshakeComplete().Done():
		case <-sessCtx.Done():
			return
		}
	}

	atomic.AddInt32(&s.sessionQueueLen, 1)
	select {
	case s.sessionQueue <- sess:
		// blocks until the session is accepted
	case <-sessCtx.Done():
		atomic.AddInt32(&s.sessionQueueLen, -1)
		// don't pass sessions that were already closed to Accept()
	}
}

func (s *baseServer) sendRetry(remoteAddr net.Addr, hdr *wire.Header) error {
	token, err := s.tokenGenerator.NewRetryToken(remoteAddr, hdr.DestConnectionID)
	if err != nil {
		return err
	}
	connID, err := protocol.GenerateConnectionID(s.config.ConnectionIDLength)
	if err != nil {
		return err
	}
	replyHdr := &wire.ExtendedHeader{}
	replyHdr.IsLongHeader = true
	replyHdr.Type = protocol.PacketTypeRetry
	replyHdr.Version = hdr.Version
	replyHdr.SrcConnectionID = connID
	replyHdr.DestConnectionID = hdr.SrcConnectionID
	replyHdr.OrigDestConnectionID = hdr.DestConnectionID
	replyHdr.Token = token
	s.logger.Debugf("Changing connection ID to %s.", connID)
	s.logger.Debugf("-> Sending Retry")
	replyHdr.Log(s.logger)
	buf := &bytes.Buffer{}
	if err := replyHdr.Write(buf, hdr.Version); err != nil {
		return err
	}
	if _, err := s.conn.WriteTo(buf.Bytes(), remoteAddr); err != nil {
		s.logger.Debugf("Error sending Retry: %s", err)
	}
	return nil
}

func (s *baseServer) sendServerBusy(remoteAddr net.Addr, hdr *wire.Header) error {
	sealer, _ := handshake.NewInitialAEAD(hdr.DestConnectionID, protocol.PerspectiveServer)
	packetBuffer := getPacketBuffer()
	defer packetBuffer.Release()
	buf := bytes.NewBuffer(packetBuffer.Slice[:0])

	ccf := &wire.ConnectionCloseFrame{ErrorCode: qerr.ServerBusy}

	replyHdr := &wire.ExtendedHeader{}
	replyHdr.IsLongHeader = true
	replyHdr.Type = protocol.PacketTypeInitial
	replyHdr.Version = hdr.Version
	replyHdr.SrcConnectionID = hdr.DestConnectionID
	replyHdr.DestConnectionID = hdr.SrcConnectionID
	replyHdr.PacketNumberLen = protocol.PacketNumberLen4
	replyHdr.Length = 4 /* packet number len */ + ccf.Length(hdr.Version) + protocol.ByteCount(sealer.Overhead())
	if err := replyHdr.Write(buf, hdr.Version); err != nil {
		return err
	}
	payloadOffset := buf.Len()

	if err := ccf.Write(buf, hdr.Version); err != nil {
		return err
	}

	raw := buf.Bytes()
	_ = sealer.Seal(raw[payloadOffset:payloadOffset], raw[payloadOffset:], replyHdr.PacketNumber, raw[:payloadOffset])
	raw = raw[0 : buf.Len()+sealer.Overhead()]

	pnOffset := payloadOffset - int(replyHdr.PacketNumberLen)
	sealer.EncryptHeader(
		raw[pnOffset+4:pnOffset+4+16],
		&raw[0],
		raw[pnOffset:payloadOffset],
	)

	replyHdr.Log(s.logger)
	wire.LogFrame(s.logger, ccf, true)
	if _, err := s.conn.WriteTo(raw, remoteAddr); err != nil {
		s.logger.Debugf("Error rejecting connection: %s", err)
	}
	return nil
}

func (s *baseServer) sendVersionNegotiationPacket(p *receivedPacket, hdr *wire.Header) {
	s.logger.Debugf("Client offered version %s, sending Version Negotiation", hdr.Version)
	data, err := wire.ComposeVersionNegotiation(hdr.SrcConnectionID, hdr.DestConnectionID, s.config.Versions)
	if err != nil {
		s.logger.Debugf("Error composing Version Negotiation: %s", err)
		return
	}
	if _, err := s.conn.WriteTo(data, p.remoteAddr); err != nil {
		s.logger.Debugf("Error sending Version Negotiation: %s", err)
	}
}
