package quic

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/lucas-clemente/quic-go/internal/utils"
)

var (
	connMuxerOnce sync.Once
	connMuxer     multiplexer
)

type multiplexer interface {
	AddConn(c net.PacketConn, connIDLen int, statelessResetKey []byte) (packetHandlerManager, error)
	RemoveConn(net.PacketConn) error
}

type connManager struct {
	connIDLen         int
	statelessResetKey []byte
	manager           packetHandlerManager
}

// The connMultiplexer listens on multiple net.PacketConns and dispatches
// incoming packets to the session handler.
type connMultiplexer struct {
	mutex sync.Mutex

	conns                   map[string] /* LocalAddr().String() */ connManager
	newPacketHandlerManager func(net.PacketConn, int, []byte, utils.Logger) packetHandlerManager // so it can be replaced in the tests

	logger utils.Logger
}

var _ multiplexer = &connMultiplexer{}

func getMultiplexer() multiplexer {
	connMuxerOnce.Do(func() {
		connMuxer = &connMultiplexer{
			conns:                   make(map[string]connManager),
			logger:                  utils.DefaultLogger.WithPrefix("muxer"),
			newPacketHandlerManager: newPacketHandlerMap,
		}
	})
	return connMuxer
}

func (m *connMultiplexer) AddConn(
	c net.PacketConn,
	connIDLen int,
	statelessResetKey []byte,
) (packetHandlerManager, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	connIndex := c.LocalAddr().Network() + " " + c.LocalAddr().String()
	p, ok := m.conns[connIndex]
	if !ok {
		manager := m.newPacketHandlerManager(c, connIDLen, statelessResetKey, m.logger)
		p = connManager{
			connIDLen:         connIDLen,
			statelessResetKey: statelessResetKey,
			manager:           manager,
		}
		m.conns[connIndex] = p
	}
	if p.connIDLen != connIDLen {
		return nil, fmt.Errorf("cannot use %d byte connection IDs on a connection that is already using %d byte connction IDs", connIDLen, p.connIDLen)
	}
	if statelessResetKey != nil && !bytes.Equal(p.statelessResetKey, statelessResetKey) {
		return nil, fmt.Errorf("cannot use different stateless reset keys on the same packet conn")
	}
	return p.manager, nil
}

func (m *connMultiplexer) RemoveConn(c net.PacketConn) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	connIndex := c.LocalAddr().Network() + " " + c.LocalAddr().String()
	if _, ok := m.conns[connIndex]; !ok {
		return fmt.Errorf("cannote remove connection, connection is unknown")
	}

	delete(m.conns, connIndex)
	return nil
}
