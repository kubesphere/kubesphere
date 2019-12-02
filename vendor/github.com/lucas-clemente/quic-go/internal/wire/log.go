package wire

import (
	"fmt"
	"strings"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

// LogFrame logs a frame, either sent or received
func LogFrame(logger utils.Logger, frame Frame, sent bool) {
	if !logger.Debug() {
		return
	}
	dir := "<-"
	if sent {
		dir = "->"
	}
	switch f := frame.(type) {
	case *CryptoFrame:
		dataLen := protocol.ByteCount(len(f.Data))
		logger.Debugf("\t%s &wire.CryptoFrame{Offset: 0x%x, Data length: 0x%x, Offset + Data length: 0x%x}", dir, f.Offset, dataLen, f.Offset+dataLen)
	case *StreamFrame:
		logger.Debugf("\t%s &wire.StreamFrame{StreamID: %d, FinBit: %t, Offset: 0x%x, Data length: 0x%x, Offset + Data length: 0x%x}", dir, f.StreamID, f.FinBit, f.Offset, f.DataLen(), f.Offset+f.DataLen())
	case *AckFrame:
		if len(f.AckRanges) > 1 {
			ackRanges := make([]string, len(f.AckRanges))
			for i, r := range f.AckRanges {
				ackRanges[i] = fmt.Sprintf("{Largest: %#x, Smallest: %#x}", r.Largest, r.Smallest)
			}
			logger.Debugf("\t%s &wire.AckFrame{LargestAcked: %#x, LowestAcked: %#x, AckRanges: {%s}, DelayTime: %s}", dir, f.LargestAcked(), f.LowestAcked(), strings.Join(ackRanges, ", "), f.DelayTime.String())
		} else {
			logger.Debugf("\t%s &wire.AckFrame{LargestAcked: %#x, LowestAcked: %#x, DelayTime: %s}", dir, f.LargestAcked(), f.LowestAcked(), f.DelayTime.String())
		}
	case *MaxStreamsFrame:
		switch f.Type {
		case protocol.StreamTypeUni:
			logger.Debugf("\t%s &wire.MaxStreamsFrame{Type: uni, MaxStreamNum: %d}", dir, f.MaxStreamNum)
		case protocol.StreamTypeBidi:
			logger.Debugf("\t%s &wire.MaxStreamsFrame{Type: bidi, MaxStreamNum: %d}", dir, f.MaxStreamNum)
		}
	case *StreamsBlockedFrame:
		switch f.Type {
		case protocol.StreamTypeUni:
			logger.Debugf("\t%s &wire.StreamsBlockedFrame{Type: uni, MaxStreams: %d}", dir, f.StreamLimit)
		case protocol.StreamTypeBidi:
			logger.Debugf("\t%s &wire.StreamsBlockedFrame{Type: bidi, MaxStreams: %d}", dir, f.StreamLimit)
		}
	case *NewConnectionIDFrame:
		logger.Debugf("\t%s &wire.NewConnectionIDFrame{SequenceNumber: %d, ConnectionID: %s, StatelessResetToken: %#x}", dir, f.SequenceNumber, f.ConnectionID, f.StatelessResetToken)
	case *NewTokenFrame:
		logger.Debugf("\t%s &wire.NewTokenFrame{Token: %#x}", dir, f.Token)
	default:
		logger.Debugf("\t%s %#v", dir, frame)
	}
}
