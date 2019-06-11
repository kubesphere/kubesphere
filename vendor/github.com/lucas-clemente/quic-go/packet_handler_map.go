package quic

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"hash"
	"net"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

// The packetHandlerMap stores packetHandlers, identified by connection ID.
// It is used:
// * by the server to store sessions
// * when multiplexing outgoing connections to store clients
type packetHandlerMap struct {
	mutex sync.RWMutex

	conn      net.PacketConn
	connIDLen int

	handlers    map[string] /* string(ConnectionID)*/ packetHandler
	resetTokens map[[16]byte] /* stateless reset token */ packetHandler
	server      unknownPacketHandler

	listening chan struct{} // is closed when listen returns
	closed    bool

	deleteRetiredSessionsAfter time.Duration

	statelessResetEnabled bool
	statelessResetHasher  hash.Hash

	logger utils.Logger
}

var _ packetHandlerManager = &packetHandlerMap{}

func newPacketHandlerMap(
	conn net.PacketConn,
	connIDLen int,
	statelessResetKey []byte,
	logger utils.Logger,
) packetHandlerManager {
	m := &packetHandlerMap{
		conn:                       conn,
		connIDLen:                  connIDLen,
		listening:                  make(chan struct{}),
		handlers:                   make(map[string]packetHandler),
		resetTokens:                make(map[[16]byte]packetHandler),
		deleteRetiredSessionsAfter: protocol.RetiredConnectionIDDeleteTimeout,
		statelessResetEnabled:      len(statelessResetKey) > 0,
		statelessResetHasher:       hmac.New(sha256.New, statelessResetKey),
		logger:                     logger,
	}
	go m.listen()
	return m
}

func (h *packetHandlerMap) Add(id protocol.ConnectionID, handler packetHandler) {
	h.mutex.Lock()
	h.handlers[string(id)] = handler
	h.mutex.Unlock()
}

func (h *packetHandlerMap) Remove(id protocol.ConnectionID) {
	h.removeByConnectionIDAsString(string(id))
}

func (h *packetHandlerMap) removeByConnectionIDAsString(id string) {
	h.mutex.Lock()
	delete(h.handlers, id)
	h.mutex.Unlock()
}

func (h *packetHandlerMap) Retire(id protocol.ConnectionID) {
	h.retireByConnectionIDAsString(string(id))
}

func (h *packetHandlerMap) retireByConnectionIDAsString(id string) {
	time.AfterFunc(h.deleteRetiredSessionsAfter, func() {
		h.removeByConnectionIDAsString(id)
	})
}

func (h *packetHandlerMap) AddResetToken(token [16]byte, handler packetHandler) {
	h.mutex.Lock()
	h.resetTokens[token] = handler
	h.mutex.Unlock()
}

func (h *packetHandlerMap) RemoveResetToken(token [16]byte) {
	h.mutex.Lock()
	delete(h.resetTokens, token)
	h.mutex.Unlock()
}

func (h *packetHandlerMap) SetServer(s unknownPacketHandler) {
	h.mutex.Lock()
	h.server = s
	h.mutex.Unlock()
}

func (h *packetHandlerMap) CloseServer() {
	h.mutex.Lock()
	h.server = nil
	var wg sync.WaitGroup
	for id, handler := range h.handlers {
		if handler.getPerspective() == protocol.PerspectiveServer {
			wg.Add(1)
			go func(id string, handler packetHandler) {
				// session.Close() blocks until the CONNECTION_CLOSE has been sent and the run-loop has stopped
				_ = handler.Close()
				h.retireByConnectionIDAsString(id)
				wg.Done()
			}(id, handler)
		}
	}
	h.mutex.Unlock()
	wg.Wait()
}

// Close the underlying connection and wait until listen() has returned.
func (h *packetHandlerMap) Close() error {
	if err := h.conn.Close(); err != nil {
		return err
	}
	<-h.listening // wait until listening returns
	return nil
}

func (h *packetHandlerMap) close(e error) error {
	h.mutex.Lock()
	if h.closed {
		h.mutex.Unlock()
		return nil
	}
	h.closed = true

	var wg sync.WaitGroup
	for _, handler := range h.handlers {
		wg.Add(1)
		go func(handler packetHandler) {
			handler.destroy(e)
			wg.Done()
		}(handler)
	}

	if h.server != nil {
		h.server.closeWithError(e)
	}
	h.mutex.Unlock()
	wg.Wait()
	return getMultiplexer().RemoveConn(h.conn)
}

func (h *packetHandlerMap) listen() {
	defer close(h.listening)
	for {
		buffer := getPacketBuffer()
		data := buffer.Slice
		// The packet size should not exceed protocol.MaxReceivePacketSize bytes
		// If it does, we only read a truncated packet, which will then end up undecryptable
		n, addr, err := h.conn.ReadFrom(data)
		if err != nil {
			h.close(err)
			return
		}
		h.handlePacket(addr, buffer, data[:n])
	}
}

func (h *packetHandlerMap) handlePacket(
	addr net.Addr,
	buffer *packetBuffer,
	data []byte,
) {
	connID, err := wire.ParseConnectionID(data, h.connIDLen)
	if err != nil {
		h.logger.Debugf("error parsing connection ID on packet from %s: %s", addr, err)
		return
	}
	rcvTime := time.Now()

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if isStatelessReset := h.maybeHandleStatelessReset(data); isStatelessReset {
		return
	}

	handler, handlerFound := h.handlers[string(connID)]

	p := &receivedPacket{
		remoteAddr: addr,
		rcvTime:    rcvTime,
		buffer:     buffer,
		data:       data,
	}
	if handlerFound { // existing session
		handler.handlePacket(p)
		return
	}
	if data[0]&0x80 == 0 {
		go h.maybeSendStatelessReset(p, connID)
		return
	}
	if h.server == nil { // no server set
		h.logger.Debugf("received a packet with an unexpected connection ID %s", connID)
		return
	}
	h.server.handlePacket(p)
}

func (h *packetHandlerMap) maybeHandleStatelessReset(data []byte) bool {
	// stateless resets are always short header packets
	if data[0]&0x80 != 0 {
		return false
	}
	if len(data) < protocol.MinStatelessResetSize {
		return false
	}

	var token [16]byte
	copy(token[:], data[len(data)-16:])
	if sess, ok := h.resetTokens[token]; ok {
		h.logger.Debugf("Received a stateless retry with token %#x. Closing session.", token)
		go sess.destroy(errors.New("received a stateless reset"))
		return true
	}
	return false
}

func (h *packetHandlerMap) GetStatelessResetToken(connID protocol.ConnectionID) [16]byte {
	var token [16]byte
	if !h.statelessResetEnabled {
		// Return a random stateless reset token.
		// This token will be sent in the server's transport parameters.
		// By using a random token, an off-path attacker won't be able to disrupt the connection.
		rand.Read(token[:])
		return token
	}
	h.statelessResetHasher.Write(connID.Bytes())
	copy(token[:], h.statelessResetHasher.Sum(nil))
	h.statelessResetHasher.Reset()
	return token
}

func (h *packetHandlerMap) maybeSendStatelessReset(p *receivedPacket, connID protocol.ConnectionID) {
	defer p.buffer.Release()
	if !h.statelessResetEnabled {
		return
	}
	// Don't send a stateless reset in response to very small packets.
	// This includes packets that could be stateless resets.
	if len(p.data) <= protocol.MinStatelessResetSize {
		return
	}
	token := h.GetStatelessResetToken(connID)
	h.logger.Debugf("Sending stateless reset to %s (connection ID: %s). Token: %#x", p.remoteAddr, connID, token)
	data := make([]byte, 23)
	rand.Read(data)
	data[0] = (data[0] & 0x7f) | 0x40
	data = append(data, token[:]...)
	if _, err := h.conn.WriteTo(data, p.remoteAddr); err != nil {
		h.logger.Debugf("Error sending Stateless Reset: %s", err)
	}
}
