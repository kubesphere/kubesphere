package quic

import (
	"sync"

	"github.com/lucas-clemente/quic-go/internal/protocol"
)

type packetBuffer struct {
	Slice []byte

	// refCount counts how many packets the Slice is used in.
	// It doesn't support concurrent use.
	// It is > 1 when used for coalesced packet.
	refCount int
}

// Split increases the refCount.
// It must be called when a packet buffer is used for more than one packet,
// e.g. when splitting coalesced packets.
func (b *packetBuffer) Split() {
	b.refCount++
}

// Decrement decrements the reference counter.
// It doesn't put the buffer back into the pool.
func (b *packetBuffer) Decrement() {
	b.refCount--
	if b.refCount < 0 {
		panic("negative packetBuffer refCount")
	}
}

// MaybeRelease puts the packet buffer back into the pool,
// if the reference counter already reached 0.
func (b *packetBuffer) MaybeRelease() {
	// only put the packetBuffer back if it's not used any more
	if b.refCount == 0 {
		b.putBack()
	}
}

// Release puts back the packet buffer into the pool.
// It should be called when processing is definitely finished.
func (b *packetBuffer) Release() {
	b.Decrement()
	if b.refCount != 0 {
		panic("packetBuffer refCount not zero")
	}
	b.putBack()
}

func (b *packetBuffer) putBack() {
	if cap(b.Slice) != int(protocol.MaxReceivePacketSize) {
		panic("putPacketBuffer called with packet of wrong size!")
	}
	bufferPool.Put(b)
}

var bufferPool sync.Pool

func getPacketBuffer() *packetBuffer {
	buf := bufferPool.Get().(*packetBuffer)
	buf.refCount = 1
	buf.Slice = buf.Slice[:protocol.MaxReceivePacketSize]
	return buf
}

func init() {
	bufferPool.New = func() interface{} {
		return &packetBuffer{
			Slice: make([]byte, 0, protocol.MaxReceivePacketSize),
		}
	}
}
