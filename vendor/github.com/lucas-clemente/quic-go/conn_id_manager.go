package quic

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type connIDManager struct {
	queue utils.NewConnectionIDList

	activeSequenceNumber      uint64
	retiredPriorTo            uint64
	activeConnectionID        protocol.ConnectionID
	activeStatelessResetToken *[16]byte

	// We change the connection ID after sending on average
	// protocol.PacketsPerConnectionID packets. The actual value is randomized
	// hide the packet loss rate from on-path observers.
	packetsSinceLastChange uint64
	rand                   *mrand.Rand
	packetsPerConnectionID uint64

	addStatelessResetToken    func([16]byte)
	removeStatelessResetToken func([16]byte)
	retireStatelessResetToken func([16]byte)
	queueControlFrame         func(wire.Frame)
}

func newConnIDManager(
	initialDestConnID protocol.ConnectionID,
	addStatelessResetToken func([16]byte),
	removeStatelessResetToken func([16]byte),
	retireStatelessResetToken func([16]byte),
	queueControlFrame func(wire.Frame),
) *connIDManager {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // ignore the error here. Nothing bad will happen if the seed is not perfectly random.
	seed := int64(binary.BigEndian.Uint64(b))
	return &connIDManager{
		activeConnectionID:        initialDestConnID,
		addStatelessResetToken:    addStatelessResetToken,
		removeStatelessResetToken: removeStatelessResetToken,
		retireStatelessResetToken: retireStatelessResetToken,
		queueControlFrame:         queueControlFrame,
		rand:                      mrand.New(mrand.NewSource(seed)),
	}
}

func (h *connIDManager) Add(f *wire.NewConnectionIDFrame) error {
	if err := h.add(f); err != nil {
		return err
	}
	if h.queue.Len() >= protocol.MaxActiveConnectionIDs {
		h.updateConnectionID()
	}
	return nil
}

func (h *connIDManager) add(f *wire.NewConnectionIDFrame) error {
	// If the NEW_CONNECTION_ID frame is reordered, such that its sequenece number
	// was already retired, send the RETIRE_CONNECTION_ID frame immediately.
	if f.SequenceNumber < h.retiredPriorTo {
		h.queueControlFrame(&wire.RetireConnectionIDFrame{
			SequenceNumber: f.SequenceNumber,
		})
		return nil
	}

	// Retire elements in the queue.
	// Doesn't retire the active connection ID.
	if f.RetirePriorTo > h.retiredPriorTo {
		var next *utils.NewConnectionIDElement
		for el := h.queue.Front(); el != nil; el = next {
			if el.Value.SequenceNumber >= f.RetirePriorTo {
				break
			}
			next = el.Next()
			h.queueControlFrame(&wire.RetireConnectionIDFrame{
				SequenceNumber: el.Value.SequenceNumber,
			})
			h.queue.Remove(el)
		}
		h.retiredPriorTo = f.RetirePriorTo
	}

	// insert a new element at the end
	if h.queue.Len() == 0 || h.queue.Back().Value.SequenceNumber < f.SequenceNumber {
		h.queue.PushBack(utils.NewConnectionID{
			SequenceNumber:      f.SequenceNumber,
			ConnectionID:        f.ConnectionID,
			StatelessResetToken: &f.StatelessResetToken,
		})
	} else {
		// insert a new element somewhere in the middle
		for el := h.queue.Front(); el != nil; el = el.Next() {
			if el.Value.SequenceNumber == f.SequenceNumber {
				if !el.Value.ConnectionID.Equal(f.ConnectionID) {
					return fmt.Errorf("received conflicting connection IDs for sequence number %d", f.SequenceNumber)
				}
				if *el.Value.StatelessResetToken != f.StatelessResetToken {
					return fmt.Errorf("received conflicting stateless reset tokens for sequence number %d", f.SequenceNumber)
				}
				break
			}
			if el.Value.SequenceNumber > f.SequenceNumber {
				h.queue.InsertBefore(utils.NewConnectionID{
					SequenceNumber:      f.SequenceNumber,
					ConnectionID:        f.ConnectionID,
					StatelessResetToken: &f.StatelessResetToken,
				}, el)
				break
			}
		}
	}

	// Retire the active connection ID, if necessary.
	if h.activeSequenceNumber < f.RetirePriorTo {
		// The queue is guaranteed to have at least one element at this point.
		h.updateConnectionID()
	}
	return nil
}

func (h *connIDManager) updateConnectionID() {
	h.queueControlFrame(&wire.RetireConnectionIDFrame{
		SequenceNumber: h.activeSequenceNumber,
	})
	if h.activeStatelessResetToken != nil {
		h.retireStatelessResetToken(*h.activeStatelessResetToken)
	}
	front := h.queue.Remove(h.queue.Front())
	h.activeSequenceNumber = front.SequenceNumber
	h.activeConnectionID = front.ConnectionID
	h.activeStatelessResetToken = front.StatelessResetToken
	h.packetsSinceLastChange = 0
	h.packetsPerConnectionID = protocol.PacketsPerConnectionID/2 + uint64(h.rand.Int63n(protocol.PacketsPerConnectionID))
	h.addStatelessResetToken(*h.activeStatelessResetToken)
}

func (h *connIDManager) Close() {
	if h.activeStatelessResetToken != nil {
		h.removeStatelessResetToken(*h.activeStatelessResetToken)
	}
}

// is called when the server performs a Retry
// and when the server changes the connection ID in the first Initial sent
func (h *connIDManager) ChangeInitialConnID(newConnID protocol.ConnectionID) {
	if h.activeSequenceNumber != 0 {
		panic("expected first connection ID to have sequence number 0")
	}
	h.activeConnectionID = newConnID
}

// is called when the server provides a stateless reset token in the transport parameters
func (h *connIDManager) SetStatelessResetToken(token [16]byte) {
	if h.activeSequenceNumber != 0 {
		panic("expected first connection ID to have sequence number 0")
	}
	h.activeStatelessResetToken = &token
	h.addStatelessResetToken(token)
}

func (h *connIDManager) SentPacket() {
	h.packetsSinceLastChange++
}

func (h *connIDManager) shouldUpdateConnID() bool {
	// iniate the first change as early as possible
	if h.queue.Len() > 0 && h.activeSequenceNumber == 0 {
		return true
	}
	// For later changes, only change if
	// 1. The queue of connection IDs is filled more than 50%.
	// 2. We sent at least PacketsPerConnectionID packets
	return 2*h.queue.Len() >= protocol.MaxActiveConnectionIDs &&
		h.packetsSinceLastChange >= h.packetsPerConnectionID
}

func (h *connIDManager) Get() protocol.ConnectionID {
	if h.shouldUpdateConnID() {
		h.updateConnectionID()
	}
	return h.activeConnectionID
}
