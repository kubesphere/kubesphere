package quic

import (
	"fmt"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type connIDGenerator struct {
	connIDLen  int
	highestSeq uint64

	activeSrcConnIDs map[uint64]protocol.ConnectionID

	addConnectionID    func(protocol.ConnectionID) [16]byte
	removeConnectionID func(protocol.ConnectionID)
	retireConnectionID func(protocol.ConnectionID)
	replaceWithClosed  func(protocol.ConnectionID, packetHandler)
	queueControlFrame  func(wire.Frame)
}

func newConnIDGenerator(
	initialConnectionID protocol.ConnectionID,
	addConnectionID func(protocol.ConnectionID) [16]byte,
	removeConnectionID func(protocol.ConnectionID),
	retireConnectionID func(protocol.ConnectionID),
	replaceWithClosed func(protocol.ConnectionID, packetHandler),
	queueControlFrame func(wire.Frame),
) *connIDGenerator {
	m := &connIDGenerator{
		connIDLen:          initialConnectionID.Len(),
		activeSrcConnIDs:   make(map[uint64]protocol.ConnectionID),
		addConnectionID:    addConnectionID,
		removeConnectionID: removeConnectionID,
		retireConnectionID: retireConnectionID,
		replaceWithClosed:  replaceWithClosed,
		queueControlFrame:  queueControlFrame,
	}
	m.activeSrcConnIDs[0] = initialConnectionID
	return m
}

func (m *connIDGenerator) SetMaxActiveConnIDs(limit uint64) error {
	if m.connIDLen == 0 {
		return nil
	}
	// The active_connection_id_limit transport parameter is the number of
	// connection IDs issued in NEW_CONNECTION_IDs frame that the peer will store.
	for i := uint64(0); i < utils.MinUint64(limit, protocol.MaxIssuedConnectionIDs); i++ {
		if err := m.issueNewConnID(); err != nil {
			return err
		}
	}
	return nil
}

func (m *connIDGenerator) Retire(seq uint64) error {
	if seq > m.highestSeq {
		return qerr.Error(qerr.ProtocolViolation, fmt.Sprintf("tried to retire connection ID %d. Highest issued: %d", seq, m.highestSeq))
	}
	connID, ok := m.activeSrcConnIDs[seq]
	// We might already have deleted this connection ID, if this is a duplicate frame.
	if !ok {
		return nil
	}
	m.retireConnectionID(connID)
	delete(m.activeSrcConnIDs, seq)
	// Don't issue a replacement for the initial connection ID.
	if seq == 0 {
		return nil
	}
	return m.issueNewConnID()
}

func (m *connIDGenerator) issueNewConnID() error {
	connID, err := protocol.GenerateConnectionID(m.connIDLen)
	if err != nil {
		return err
	}
	m.activeSrcConnIDs[m.highestSeq+1] = connID
	token := m.addConnectionID(connID)
	m.queueControlFrame(&wire.NewConnectionIDFrame{
		SequenceNumber:      m.highestSeq + 1,
		ConnectionID:        connID,
		StatelessResetToken: token,
	})
	m.highestSeq++
	return nil
}

func (m *connIDGenerator) RemoveAll() {
	for _, connID := range m.activeSrcConnIDs {
		m.removeConnectionID(connID)
	}
}

func (m *connIDGenerator) ReplaceWithClosed(handler packetHandler) {
	for _, connID := range m.activeSrcConnIDs {
		m.replaceWithClosed(connID, handler)
	}
}
