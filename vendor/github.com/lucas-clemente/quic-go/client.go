package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type client struct {
	mutex sync.Mutex

	conn connection
	// If the client is created with DialAddr, we create a packet conn.
	// If it is started with Dial, we take a packet conn as a parameter.
	createdPacketConn bool

	packetHandlers packetHandlerManager

	versionNegotiated                utils.AtomicBool // has the server accepted our version
	receivedVersionNegotiationPacket bool
	negotiatedVersions               []protocol.VersionNumber // the list of versions from the version negotiation packet

	tlsConf *tls.Config
	config  *Config

	srcConnID  protocol.ConnectionID
	destConnID protocol.ConnectionID

	initialPacketNumber protocol.PacketNumber

	initialVersion protocol.VersionNumber
	version        protocol.VersionNumber

	handshakeChan chan struct{}

	session quicSession

	logger utils.Logger
}

var _ packetHandler = &client{}

var (
	// make it possible to mock connection ID generation in the tests
	generateConnectionID           = protocol.GenerateConnectionID
	generateConnectionIDForInitial = protocol.GenerateConnectionIDForInitial
)

// DialAddr establishes a new QUIC connection to a server.
// It uses a new UDP connection and closes this connection when the QUIC session is closed.
// The hostname for SNI is taken from the given address.
func DialAddr(
	addr string,
	tlsConf *tls.Config,
	config *Config,
) (Session, error) {
	return DialAddrContext(context.Background(), addr, tlsConf, config)
}

// DialAddrContext establishes a new QUIC connection to a server using the provided context.
// See DialAddr for details.
func DialAddrContext(
	ctx context.Context,
	addr string,
	tlsConf *tls.Config,
	config *Config,
) (Session, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	return dialContext(ctx, udpConn, udpAddr, addr, tlsConf, config, true)
}

// Dial establishes a new QUIC connection to a server using a net.PacketConn.
// The same PacketConn can be used for multiple calls to Dial and Listen,
// QUIC connection IDs are used for demultiplexing the different connections.
// The host parameter is used for SNI.
// The tls.Config must define an application protocol (using NextProtos).
func Dial(
	pconn net.PacketConn,
	remoteAddr net.Addr,
	host string,
	tlsConf *tls.Config,
	config *Config,
) (Session, error) {
	return DialContext(context.Background(), pconn, remoteAddr, host, tlsConf, config)
}

// DialContext establishes a new QUIC connection to a server using a net.PacketConn using the provided context.
// See Dial for details.
func DialContext(
	ctx context.Context,
	pconn net.PacketConn,
	remoteAddr net.Addr,
	host string,
	tlsConf *tls.Config,
	config *Config,
) (Session, error) {
	return dialContext(ctx, pconn, remoteAddr, host, tlsConf, config, false)
}

func dialContext(
	ctx context.Context,
	pconn net.PacketConn,
	remoteAddr net.Addr,
	host string,
	tlsConf *tls.Config,
	config *Config,
	createdPacketConn bool,
) (Session, error) {
	if tlsConf == nil {
		return nil, errors.New("quic: tls.Config not set")
	}
	config = populateClientConfig(config, createdPacketConn)
	packetHandlers, err := getMultiplexer().AddConn(pconn, config.ConnectionIDLength, config.StatelessResetKey)
	if err != nil {
		return nil, err
	}
	c, err := newClient(pconn, remoteAddr, config, tlsConf, host, createdPacketConn)
	if err != nil {
		return nil, err
	}
	c.packetHandlers = packetHandlers
	if err := c.dial(ctx); err != nil {
		return nil, err
	}
	return c.session, nil
}

func newClient(
	pconn net.PacketConn,
	remoteAddr net.Addr,
	config *Config,
	tlsConf *tls.Config,
	host string,
	createdPacketConn bool,
) (*client, error) {
	if tlsConf == nil {
		tlsConf = &tls.Config{}
	}
	if tlsConf.ServerName == "" {
		sni := host
		if strings.IndexByte(sni, ':') != -1 {
			var err error
			sni, _, err = net.SplitHostPort(sni)
			if err != nil {
				return nil, err
			}
		}

		tlsConf.ServerName = sni
	}

	// check that all versions are actually supported
	if config != nil {
		for _, v := range config.Versions {
			if !protocol.IsValidVersion(v) {
				return nil, fmt.Errorf("%s is not a valid QUIC version", v)
			}
		}
	}

	srcConnID, err := generateConnectionID(config.ConnectionIDLength)
	if err != nil {
		return nil, err
	}
	destConnID, err := generateConnectionIDForInitial()
	if err != nil {
		return nil, err
	}
	c := &client{
		srcConnID:         srcConnID,
		destConnID:        destConnID,
		conn:              &conn{pconn: pconn, currentAddr: remoteAddr},
		createdPacketConn: createdPacketConn,
		tlsConf:           tlsConf,
		config:            config,
		version:           config.Versions[0],
		handshakeChan:     make(chan struct{}),
		logger:            utils.DefaultLogger.WithPrefix("client"),
	}
	return c, nil
}

// populateClientConfig populates fields in the quic.Config with their default values, if none are set
// it may be called with nil
func populateClientConfig(config *Config, createdPacketConn bool) *Config {
	if config == nil {
		config = &Config{}
	}
	versions := config.Versions
	if len(versions) == 0 {
		versions = protocol.SupportedVersions
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
	if connIDLen == 0 && !createdPacketConn {
		connIDLen = protocol.DefaultConnectionIDLength
	}

	return &Config{
		Versions:                              versions,
		HandshakeTimeout:                      handshakeTimeout,
		IdleTimeout:                           idleTimeout,
		ConnectionIDLength:                    connIDLen,
		MaxReceiveStreamFlowControlWindow:     maxReceiveStreamFlowControlWindow,
		MaxReceiveConnectionFlowControlWindow: maxReceiveConnectionFlowControlWindow,
		MaxIncomingStreams:                    maxIncomingStreams,
		MaxIncomingUniStreams:                 maxIncomingUniStreams,
		KeepAlive:                             config.KeepAlive,
		StatelessResetKey:                     config.StatelessResetKey,
		QuicTracer:                            config.QuicTracer,
		TokenStore:                            config.TokenStore,
	}
}

func (c *client) dial(ctx context.Context) error {
	c.logger.Infof("Starting new connection to %s (%s -> %s), source connection ID %s, destination connection ID %s, version %s", c.tlsConf.ServerName, c.conn.LocalAddr(), c.conn.RemoteAddr(), c.srcConnID, c.destConnID, c.version)
	c.createNewTLSSession(c.version)
	err := c.establishSecureConnection(ctx)
	if err == errCloseForRecreating {
		return c.dial(ctx)
	}
	return err
}

// establishSecureConnection runs the session, and tries to establish a secure connection
// It returns:
// - errCloseForRecreating when the server sends a version negotiation packet
// - any other error that might occur
// - when the connection is forward-secure
func (c *client) establishSecureConnection(ctx context.Context) error {
	errorChan := make(chan error, 1)

	go func() {
		err := c.session.run() // returns as soon as the session is closed
		if err != errCloseForRecreating && c.createdPacketConn {
			c.packetHandlers.Close()
		}
		errorChan <- err
	}()

	select {
	case <-ctx.Done():
		// The session will send a PeerGoingAway error to the server.
		c.session.Close()
		return ctx.Err()
	case err := <-errorChan:
		return err
	case <-c.session.HandshakeComplete().Done():
		// handshake successfully completed
		return nil
	}
}

func (c *client) handlePacket(p *receivedPacket) {
	if wire.IsVersionNegotiationPacket(p.data) {
		go c.handleVersionNegotiationPacket(p)
		return
	}

	// this is the first packet we are receiving
	// since it is not a Version Negotiation Packet, this means the server supports the suggested version
	if !c.versionNegotiated.Get() {
		c.versionNegotiated.Set(true)
	}

	c.session.handlePacket(p)
}

func (c *client) handleVersionNegotiationPacket(p *receivedPacket) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	hdr, _, _, err := wire.ParsePacket(p.data, 0)
	if err != nil {
		c.logger.Debugf("Error parsing Version Negotiation packet: %s", err)
		return
	}

	// ignore delayed / duplicated version negotiation packets
	if c.receivedVersionNegotiationPacket || c.versionNegotiated.Get() {
		c.logger.Debugf("Received a delayed Version Negotiation packet.")
		return
	}

	for _, v := range hdr.SupportedVersions {
		if v == c.version {
			// The Version Negotiation packet contains the version that we offered.
			// This might be a packet sent by an attacker (or by a terribly broken server implementation).
			return
		}
	}

	c.logger.Infof("Received a Version Negotiation packet. Supported Versions: %s", hdr.SupportedVersions)
	newVersion, ok := protocol.ChooseSupportedVersion(c.config.Versions, hdr.SupportedVersions)
	if !ok {
		//nolint:stylecheck
		c.session.destroy(fmt.Errorf("No compatible QUIC version found. We support %s, server offered %s", c.config.Versions, hdr.SupportedVersions))
		c.logger.Debugf("No compatible QUIC version found.")
		return
	}
	c.receivedVersionNegotiationPacket = true
	c.negotiatedVersions = hdr.SupportedVersions

	// switch to negotiated version
	c.initialVersion = c.version
	c.version = newVersion

	c.logger.Infof("Switching to QUIC version %s. New connection ID: %s", newVersion, c.destConnID)
	c.initialPacketNumber = c.session.closeForRecreating()
}

func (c *client) createNewTLSSession(_ protocol.VersionNumber) {
	c.mutex.Lock()
	c.session = newClientSession(
		c.conn,
		c.packetHandlers,
		c.destConnID,
		c.srcConnID,
		c.config,
		c.tlsConf,
		c.initialPacketNumber,
		c.initialVersion,
		c.logger,
		c.version,
	)
	c.mutex.Unlock()
	// It's not possible to use the stateless reset token for the client's (first) connection ID,
	// since there's no way to securely communicate it to the server.
	c.packetHandlers.Add(c.srcConnID, c)
}

func (c *client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}

func (c *client) destroy(e error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.session == nil {
		return
	}
	c.session.destroy(e)
}

func (c *client) GetVersion() protocol.VersionNumber {
	c.mutex.Lock()
	v := c.version
	c.mutex.Unlock()
	return v
}

func (c *client) getPerspective() protocol.Perspective {
	return protocol.PerspectiveClient
}
