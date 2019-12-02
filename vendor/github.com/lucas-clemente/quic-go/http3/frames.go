package http3

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

type byteReader interface {
	io.ByteReader
	io.Reader
}

type byteReaderImpl struct{ io.Reader }

func (br *byteReaderImpl) ReadByte() (byte, error) {
	b := make([]byte, 1)
	if _, err := br.Reader.Read(b); err != nil {
		return 0, err
	}
	return b[0], nil
}

type frame interface{}

func parseNextFrame(b io.Reader) (frame, error) {
	br, ok := b.(byteReader)
	if !ok {
		br = &byteReaderImpl{b}
	}
	t, err := utils.ReadVarInt(br)
	if err != nil {
		return nil, err
	}
	l, err := utils.ReadVarInt(br)
	if err != nil {
		return nil, err
	}

	switch t {
	case 0x0:
		return &dataFrame{Length: l}, nil
	case 0x1:
		return &headersFrame{Length: l}, nil
	case 0x4:
		return parseSettingsFrame(br, l)
	case 0x3: // CANCEL_PUSH
		fallthrough
	case 0x5: // PUSH_PROMISE
		fallthrough
	case 0x7: // GOAWAY
		fallthrough
	case 0xd: // MAX_PUSH_ID
		fallthrough
	case 0xe: // DUPLICATE_PUSH
		fallthrough
	default:
		// skip over unknown frames
		if _, err := io.CopyN(ioutil.Discard, br, int64(l)); err != nil {
			return nil, err
		}
		return parseNextFrame(b)
	}
}

type dataFrame struct {
	Length uint64
}

func (f *dataFrame) Write(b *bytes.Buffer) {
	utils.WriteVarInt(b, 0x0)
	utils.WriteVarInt(b, f.Length)
}

type headersFrame struct {
	Length uint64
}

func (f *headersFrame) Write(b *bytes.Buffer) {
	utils.WriteVarInt(b, 0x1)
	utils.WriteVarInt(b, f.Length)
}

type settingsFrame struct {
	settings map[uint64]uint64
}

func parseSettingsFrame(r io.Reader, l uint64) (*settingsFrame, error) {
	if l > 8*(1<<10) {
		return nil, fmt.Errorf("unexpected size for SETTINGS frame: %d", l)
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}
	frame := &settingsFrame{settings: make(map[uint64]uint64)}
	b := bytes.NewReader(buf)
	for b.Len() > 0 {
		id, err := utils.ReadVarInt(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}
		val, err := utils.ReadVarInt(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}
		if _, ok := frame.settings[id]; ok {
			return nil, fmt.Errorf("duplicate setting: %d", id)
		}
		frame.settings[id] = val
	}
	return frame, nil
}

func (f *settingsFrame) Write(b *bytes.Buffer) {
	utils.WriteVarInt(b, 0x4)
	var l protocol.ByteCount
	for id, val := range f.settings {
		l += utils.VarIntLen(id) + utils.VarIntLen(val)
	}
	utils.WriteVarInt(b, uint64(l))
	for id, val := range f.settings {
		utils.WriteVarInt(b, id)
		utils.WriteVarInt(b, val)
	}
}
