package ackhandler

import (
	"fmt"

	"github.com/lucas-clemente/quic-go/internal/protocol"
)

type sentPacketHistory struct {
	packetList *PacketList
	packetMap  map[protocol.PacketNumber]*PacketElement
}

func newSentPacketHistory() *sentPacketHistory {
	return &sentPacketHistory{
		packetList: NewPacketList(),
		packetMap:  make(map[protocol.PacketNumber]*PacketElement),
	}
}

func (h *sentPacketHistory) SentPacket(p *Packet) {
	el := h.packetList.PushBack(*p)
	h.packetMap[p.PacketNumber] = el
}

func (h *sentPacketHistory) GetPacket(p protocol.PacketNumber) *Packet {
	if el, ok := h.packetMap[p]; ok {
		return &el.Value
	}
	return nil
}

// Iterate iterates through all packets.
// The callback must not modify the history.
func (h *sentPacketHistory) Iterate(cb func(*Packet) (cont bool, err error)) error {
	cont := true
	for el := h.packetList.Front(); cont && el != nil; el = el.Next() {
		var err error
		cont, err = cb(&el.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// FirstOutStanding returns the first outstanding packet.
// It must not be modified (e.g. retransmitted).
// Use DequeueFirstPacketForRetransmission() to retransmit it.
func (h *sentPacketHistory) FirstOutstanding() *Packet {
	if !h.HasOutstandingPackets() {
		return nil
	}
	return &h.packetList.Front().Value
}

func (h *sentPacketHistory) Len() int {
	return len(h.packetMap)
}

func (h *sentPacketHistory) Remove(p protocol.PacketNumber) error {
	el, ok := h.packetMap[p]
	if !ok {
		return fmt.Errorf("packet %d not found in sent packet history", p)
	}
	h.packetList.Remove(el)
	delete(h.packetMap, p)
	return nil
}

func (h *sentPacketHistory) HasOutstandingPackets() bool {
	return h.packetList.Len() > 0
}
