package http3

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/marten-seemann/qpack"
	"github.com/onsi/ginkgo"
)

// allows mocking of quic.Listen and quic.ListenAddr
var (
	quicListen     = quic.Listen
	quicListenAddr = quic.ListenAddr
)

const nextProtoH3 = "h3-23"

type requestError struct {
	err       error
	streamErr errorCode
	connErr   errorCode
}

func newStreamError(code errorCode, err error) requestError {
	return requestError{err: err, streamErr: code}
}

func newConnError(code errorCode, err error) requestError {
	return requestError{err: err, connErr: code}
}

// Server is a HTTP2 server listening for QUIC connections.
type Server struct {
	*http.Server

	// By providing a quic.Config, it is possible to set parameters of the QUIC connection.
	// If nil, it uses reasonable default values.
	QuicConfig *quic.Config

	port uint32 // used atomically

	mutex     sync.Mutex
	listeners map[*quic.Listener]struct{}
	closed    utils.AtomicBool

	logger utils.Logger
}

// ListenAndServe listens on the UDP address s.Addr and calls s.Handler to handle HTTP/3 requests on incoming connections.
func (s *Server) ListenAndServe() error {
	if s.Server == nil {
		return errors.New("use of http3.Server without http.Server")
	}
	return s.serveImpl(s.TLSConfig, nil)
}

// ListenAndServeTLS listens on the UDP address s.Addr and calls s.Handler to handle HTTP/3 requests on incoming connections.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// We currently only use the cert-related stuff from tls.Config,
	// so we don't need to make a full copy.
	config := &tls.Config{
		Certificates: certs,
	}
	return s.serveImpl(config, nil)
}

// Serve an existing UDP connection.
// It is possible to reuse the same connection for outgoing connections.
// Closing the server does not close the packet conn.
func (s *Server) Serve(conn net.PacketConn) error {
	return s.serveImpl(s.TLSConfig, conn)
}

func (s *Server) serveImpl(tlsConf *tls.Config, conn net.PacketConn) error {
	if s.closed.Get() {
		return http.ErrServerClosed
	}
	if s.Server == nil {
		return errors.New("use of http3.Server without http.Server")
	}
	s.logger = utils.DefaultLogger.WithPrefix("server")

	if tlsConf == nil {
		tlsConf = &tls.Config{}
	} else {
		tlsConf = tlsConf.Clone()
	}
	// Replace existing ALPNs by H3
	tlsConf.NextProtos = []string{nextProtoH3}
	if tlsConf.GetConfigForClient != nil {
		getConfigForClient := tlsConf.GetConfigForClient
		tlsConf.GetConfigForClient = func(ch *tls.ClientHelloInfo) (*tls.Config, error) {
			conf, err := getConfigForClient(ch)
			if err != nil || conf == nil {
				return conf, err
			}
			conf = conf.Clone()
			conf.NextProtos = []string{nextProtoH3}
			return conf, nil
		}
	}

	var ln quic.Listener
	var err error
	if conn == nil {
		ln, err = quicListenAddr(s.Addr, tlsConf, s.QuicConfig)
	} else {
		ln, err = quicListen(conn, tlsConf, s.QuicConfig)
	}
	if err != nil {
		return err
	}
	s.addListener(&ln)
	defer s.removeListener(&ln)

	for {
		sess, err := ln.Accept(context.Background())
		if err != nil {
			return err
		}
		go s.handleConn(sess)
	}
}

// We store a pointer to interface in the map set. This is safe because we only
// call trackListener via Serve and can track+defer untrack the same pointer to
// local variable there. We never need to compare a Listener from another caller.
func (s *Server) addListener(l *quic.Listener) {
	s.mutex.Lock()
	if s.listeners == nil {
		s.listeners = make(map[*quic.Listener]struct{})
	}
	s.listeners[l] = struct{}{}
	s.mutex.Unlock()
}

func (s *Server) removeListener(l *quic.Listener) {
	s.mutex.Lock()
	delete(s.listeners, l)
	s.mutex.Unlock()
}

func (s *Server) handleConn(sess quic.Session) {
	// TODO: accept control streams
	decoder := qpack.NewDecoder(nil)

	// send a SETTINGS frame
	str, err := sess.OpenUniStream()
	if err != nil {
		s.logger.Debugf("Opening the control stream failed.")
		return
	}
	buf := bytes.NewBuffer([]byte{0})
	(&settingsFrame{}).Write(buf)
	str.Write(buf.Bytes())

	for {
		str, err := sess.AcceptStream(context.Background())
		if err != nil {
			s.logger.Debugf("Accepting stream failed: %s", err)
			return
		}
		go func() {
			defer ginkgo.GinkgoRecover()
			rerr := s.handleRequest(str, decoder, func() {
				sess.CloseWithError(quic.ErrorCode(errorFrameUnexpected), "")
			})
			if rerr.err != nil || rerr.streamErr != 0 || rerr.connErr != 0 {
				s.logger.Debugf("Handling request failed: %s", err)
				if rerr.streamErr != 0 {
					str.CancelWrite(quic.ErrorCode(rerr.streamErr))
				}
				if rerr.connErr != 0 {
					var reason string
					if rerr.err != nil {
						reason = rerr.err.Error()
					}
					sess.CloseWithError(quic.ErrorCode(rerr.connErr), reason)
				}
				return
			}
			str.Close()
		}()
	}
}

func (s *Server) maxHeaderBytes() uint64 {
	if s.Server.MaxHeaderBytes <= 0 {
		return http.DefaultMaxHeaderBytes
	}
	return uint64(s.Server.MaxHeaderBytes)
}

func (s *Server) handleRequest(str quic.Stream, decoder *qpack.Decoder, onFrameError func()) requestError {
	frame, err := parseNextFrame(str)
	if err != nil {
		return newStreamError(errorRequestIncomplete, err)
	}
	hf, ok := frame.(*headersFrame)
	if !ok {
		return newConnError(errorFrameUnexpected, errors.New("expected first frame to be a HEADERS frame"))
	}
	if hf.Length > s.maxHeaderBytes() {
		return newStreamError(errorFrameError, fmt.Errorf("HEADERS frame too large: %d bytes (max: %d)", hf.Length, s.maxHeaderBytes()))
	}
	headerBlock := make([]byte, hf.Length)
	if _, err := io.ReadFull(str, headerBlock); err != nil {
		return newStreamError(errorRequestIncomplete, err)
	}
	hfs, err := decoder.DecodeFull(headerBlock)
	if err != nil {
		// TODO: use the right error code
		return newConnError(errorGeneralProtocolError, err)
	}
	req, err := requestFromHeaders(hfs)
	if err != nil {
		// TODO: use the right error code
		return newStreamError(errorGeneralProtocolError, err)
	}
	req.Body = newRequestBody(str, onFrameError)

	if s.logger.Debug() {
		s.logger.Infof("%s %s%s, on stream %d", req.Method, req.Host, req.RequestURI, str.StreamID())
	} else {
		s.logger.Infof("%s %s%s", req.Method, req.Host, req.RequestURI)
	}

	req = req.WithContext(str.Context())
	responseWriter := newResponseWriter(str, s.logger)
	handler := s.Handler
	if handler == nil {
		handler = http.DefaultServeMux
	}

	var panicked, readEOF bool
	func() {
		defer func() {
			if p := recover(); p != nil {
				// Copied from net/http/server.go
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				s.logger.Errorf("http: panic serving: %v\n%s", p, buf)
				panicked = true
			}
		}()
		handler.ServeHTTP(responseWriter, req)
		// read the eof
		if _, err = str.Read([]byte{0}); err == io.EOF {
			readEOF = true
		}
	}()

	if panicked {
		responseWriter.WriteHeader(500)
	} else {
		responseWriter.WriteHeader(200)
	}

	if !readEOF {
		str.CancelRead(quic.ErrorCode(errorEarlyResponse))
	}
	return requestError{}
}

// Close the server immediately, aborting requests and sending CONNECTION_CLOSE frames to connected clients.
// Close in combination with ListenAndServe() (instead of Serve()) may race if it is called before a UDP socket is established.
func (s *Server) Close() error {
	s.closed.Set(true)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var err error
	for ln := range s.listeners {
		if cerr := (*ln).Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// CloseGracefully shuts down the server gracefully. The server sends a GOAWAY frame first, then waits for either timeout to trigger, or for all running requests to complete.
// CloseGracefully in combination with ListenAndServe() (instead of Serve()) may race if it is called before a UDP socket is established.
func (s *Server) CloseGracefully(timeout time.Duration) error {
	// TODO: implement
	return nil
}

// SetQuicHeaders can be used to set the proper headers that announce that this server supports QUIC.
// The values that are set depend on the port information from s.Server.Addr, and currently look like this (if Addr has port 443):
//  Alt-Svc: quic=":443"; ma=2592000; v="33,32,31,30"
func (s *Server) SetQuicHeaders(hdr http.Header) error {
	port := atomic.LoadUint32(&s.port)

	if port == 0 {
		// Extract port from s.Server.Addr
		_, portStr, err := net.SplitHostPort(s.Server.Addr)
		if err != nil {
			return err
		}
		portInt, err := net.LookupPort("tcp", portStr)
		if err != nil {
			return err
		}
		port = uint32(portInt)
		atomic.StoreUint32(&s.port, port)
	}

	hdr.Add("Alt-Svc", fmt.Sprintf(`%s=":%d"; ma=2592000`, nextProtoH3, port))

	return nil
}

// ListenAndServeQUIC listens on the UDP network address addr and calls the
// handler for HTTP/3 requests on incoming connections. http.DefaultServeMux is
// used when handler is nil.
func ListenAndServeQUIC(addr, certFile, keyFile string, handler http.Handler) error {
	server := &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
	return server.ListenAndServeTLS(certFile, keyFile)
}

// ListenAndServe listens on the given network address for both, TLS and QUIC
// connetions in parallel. It returns if one of the two returns an error.
// http.DefaultServeMux is used when handler is nil.
// The correct Alt-Svc headers for QUIC are set.
func ListenAndServe(addr, certFile, keyFile string, handler http.Handler) error {
	// Load certs
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// We currently only use the cert-related stuff from tls.Config,
	// so we don't need to make a full copy.
	config := &tls.Config{
		Certificates: certs,
	}

	// Open the listeners
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	tcpConn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer tcpConn.Close()

	tlsConn := tls.NewListener(tcpConn, config)
	defer tlsConn.Close()

	// Start the servers
	httpServer := &http.Server{
		Addr:      addr,
		TLSConfig: config,
	}

	quicServer := &Server{
		Server: httpServer,
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}
	httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		quicServer.SetQuicHeaders(w.Header())
		handler.ServeHTTP(w, r)
	})

	hErr := make(chan error)
	qErr := make(chan error)
	go func() {
		hErr <- httpServer.Serve(tlsConn)
	}()
	go func() {
		qErr <- quicServer.Serve(udpConn)
	}()

	select {
	case err := <-hErr:
		quicServer.Close()
		return err
	case err := <-qErr:
		// Cannot close the HTTP server or wait for requests to complete properly :/
		return err
	}
}
