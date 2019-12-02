package handshake

import (
	"crypto/tls"
	"net"
	"time"
	"unsafe"

	"github.com/marten-seemann/qtls"
)

type conn struct {
	remoteAddr net.Addr
}

func newConn(remote net.Addr) net.Conn {
	return &conn{remoteAddr: remote}
}

var _ net.Conn = &conn{}

func (c *conn) Read([]byte) (int, error)         { return 0, nil }
func (c *conn) Write([]byte) (int, error)        { return 0, nil }
func (c *conn) Close() error                     { return nil }
func (c *conn) RemoteAddr() net.Addr             { return c.remoteAddr }
func (c *conn) LocalAddr() net.Addr              { return nil }
func (c *conn) SetReadDeadline(time.Time) error  { return nil }
func (c *conn) SetWriteDeadline(time.Time) error { return nil }
func (c *conn) SetDeadline(time.Time) error      { return nil }

type clientSessionCache struct {
	tls.ClientSessionCache
}

var _ qtls.ClientSessionCache = &clientSessionCache{}

func (c *clientSessionCache) Get(sessionKey string) (*qtls.ClientSessionState, bool) {
	sess, ok := c.ClientSessionCache.Get(sessionKey)
	if sess == nil {
		return nil, ok
	}
	// qtls.ClientSessionState is identical to the tls.ClientSessionState.
	// In order to allow users of quic-go to use a tls.Config,
	// we need this workaround to use the ClientSessionCache.
	// In unsafe.go we check that the two structs are actually identical.
	usess := (*[unsafe.Sizeof(*sess)]byte)(unsafe.Pointer(sess))[:]
	var session qtls.ClientSessionState
	usession := (*[unsafe.Sizeof(session)]byte)(unsafe.Pointer(&session))[:]
	copy(usession, usess)
	return &session, ok
}

func (c *clientSessionCache) Put(sessionKey string, cs *qtls.ClientSessionState) {
	if cs == nil {
		c.ClientSessionCache.Put(sessionKey, nil)
		return
	}
	// qtls.ClientSessionState is identical to the tls.ClientSessionState.
	// In order to allow users of quic-go to use a tls.Config,
	// we need this workaround to use the ClientSessionCache.
	// In unsafe.go we check that the two structs are actually identical.
	usess := (*[unsafe.Sizeof(*cs)]byte)(unsafe.Pointer(cs))[:]
	var session tls.ClientSessionState
	usession := (*[unsafe.Sizeof(session)]byte)(unsafe.Pointer(&session))[:]
	copy(usession, usess)
	c.ClientSessionCache.Put(sessionKey, &session)
}

func tlsConfigToQtlsConfig(
	c *tls.Config,
	recordLayer qtls.RecordLayer,
	extHandler tlsExtensionHandler,
) *qtls.Config {
	if c == nil {
		c = &tls.Config{}
	}
	// Clone the config first. This executes the tls.Config.serverInit().
	// This sets the SessionTicketKey, if the user didn't supply one.
	c = c.Clone()
	// QUIC requires TLS 1.3 or newer
	minVersion := c.MinVersion
	if minVersion < qtls.VersionTLS13 {
		minVersion = qtls.VersionTLS13
	}
	maxVersion := c.MaxVersion
	if maxVersion < qtls.VersionTLS13 {
		maxVersion = qtls.VersionTLS13
	}
	var getConfigForClient func(ch *tls.ClientHelloInfo) (*qtls.Config, error)
	if c.GetConfigForClient != nil {
		getConfigForClient = func(ch *tls.ClientHelloInfo) (*qtls.Config, error) {
			tlsConf, err := c.GetConfigForClient(ch)
			if err != nil {
				return nil, err
			}
			if tlsConf == nil {
				return nil, nil
			}
			return tlsConfigToQtlsConfig(tlsConf, recordLayer, extHandler), nil
		}
	}
	var csc qtls.ClientSessionCache
	if c.ClientSessionCache != nil {
		csc = &clientSessionCache{c.ClientSessionCache}
	}
	return &qtls.Config{
		Rand:                        c.Rand,
		Time:                        c.Time,
		Certificates:                c.Certificates,
		NameToCertificate:           c.NameToCertificate,
		GetCertificate:              c.GetCertificate,
		GetClientCertificate:        c.GetClientCertificate,
		GetConfigForClient:          getConfigForClient,
		VerifyPeerCertificate:       c.VerifyPeerCertificate,
		RootCAs:                     c.RootCAs,
		NextProtos:                  c.NextProtos,
		EnforceNextProtoSelection:   true,
		ServerName:                  c.ServerName,
		ClientAuth:                  c.ClientAuth,
		ClientCAs:                   c.ClientCAs,
		InsecureSkipVerify:          c.InsecureSkipVerify,
		CipherSuites:                c.CipherSuites,
		PreferServerCipherSuites:    c.PreferServerCipherSuites,
		SessionTicketsDisabled:      c.SessionTicketsDisabled,
		SessionTicketKey:            c.SessionTicketKey,
		ClientSessionCache:          csc,
		MinVersion:                  minVersion,
		MaxVersion:                  maxVersion,
		CurvePreferences:            c.CurvePreferences,
		DynamicRecordSizingDisabled: c.DynamicRecordSizingDisabled,
		// no need to copy Renegotiation, it's not supported by TLS 1.3
		KeyLogWriter:           c.KeyLogWriter,
		AlternativeRecordLayer: recordLayer,
		GetExtensions:          extHandler.GetExtensions,
		ReceivedExtensions:     extHandler.ReceivedExtensions,
	}
}

func cipherSuiteName(id uint16) string {
	switch id {
	case qtls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case qtls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case qtls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	default:
		return "unknown cipher suite"
	}
}
