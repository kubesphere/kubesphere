package wire

import (
	"bytes"
	"fmt"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
)

type frameParser struct {
	ackDelayExponent uint8

	version protocol.VersionNumber
}

// NewFrameParser creates a new frame parser.
func NewFrameParser(v protocol.VersionNumber) FrameParser {
	return &frameParser{version: v}
}

// ParseNextFrame parses the next frame
// It skips PADDING frames.
func (p *frameParser) ParseNext(r *bytes.Reader, encLevel protocol.EncryptionLevel) (Frame, error) {
	for r.Len() != 0 {
		typeByte, _ := r.ReadByte()
		if typeByte == 0x0 { // PADDING frame
			continue
		}
		r.UnreadByte()

		return p.parseFrame(r, typeByte, encLevel)
	}
	return nil, nil
}

func (p *frameParser) parseFrame(r *bytes.Reader, typeByte byte, encLevel protocol.EncryptionLevel) (Frame, error) {
	var frame Frame
	var err error
	if typeByte&0xf8 == 0x8 {
		frame, err = parseStreamFrame(r, p.version)
		if err != nil {
			return nil, qerr.Error(qerr.FrameEncodingError, err.Error())
		}
		return frame, nil
	}
	switch typeByte {
	case 0x1:
		frame, err = parsePingFrame(r, p.version)
	case 0x2, 0x3:
		ackDelayExponent := p.ackDelayExponent
		if encLevel != protocol.Encryption1RTT {
			ackDelayExponent = protocol.DefaultAckDelayExponent
		}
		frame, err = parseAckFrame(r, ackDelayExponent, p.version)
	case 0x4:
		frame, err = parseResetStreamFrame(r, p.version)
	case 0x5:
		frame, err = parseStopSendingFrame(r, p.version)
	case 0x6:
		frame, err = parseCryptoFrame(r, p.version)
	case 0x7:
		frame, err = parseNewTokenFrame(r, p.version)
	case 0x10:
		frame, err = parseMaxDataFrame(r, p.version)
	case 0x11:
		frame, err = parseMaxStreamDataFrame(r, p.version)
	case 0x12, 0x13:
		frame, err = parseMaxStreamsFrame(r, p.version)
	case 0x14:
		frame, err = parseDataBlockedFrame(r, p.version)
	case 0x15:
		frame, err = parseStreamDataBlockedFrame(r, p.version)
	case 0x16, 0x17:
		frame, err = parseStreamsBlockedFrame(r, p.version)
	case 0x18:
		frame, err = parseNewConnectionIDFrame(r, p.version)
	case 0x19:
		frame, err = parseRetireConnectionIDFrame(r, p.version)
	case 0x1a:
		frame, err = parsePathChallengeFrame(r, p.version)
	case 0x1b:
		frame, err = parsePathResponseFrame(r, p.version)
	case 0x1c, 0x1d:
		frame, err = parseConnectionCloseFrame(r, p.version)
	default:
		err = fmt.Errorf("unknown type byte 0x%x", typeByte)
	}
	if err != nil {
		return nil, qerr.Error(qerr.FrameEncodingError, err.Error())
	}
	return frame, nil
}

func (p *frameParser) SetAckDelayExponent(exp uint8) {
	p.ackDelayExponent = exp
}
