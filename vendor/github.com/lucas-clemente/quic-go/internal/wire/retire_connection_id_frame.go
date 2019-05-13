package wire

import (
	"bytes"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

// A RetireConnectionIDFrame is a RETIRE_CONNECTION_ID frame
type RetireConnectionIDFrame struct {
	SequenceNumber uint64
}

func parseRetireConnectionIDFrame(r *bytes.Reader, _ protocol.VersionNumber) (*RetireConnectionIDFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	seq, err := utils.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	return &RetireConnectionIDFrame{SequenceNumber: seq}, nil
}

func (f *RetireConnectionIDFrame) Write(b *bytes.Buffer, _ protocol.VersionNumber) error {
	b.WriteByte(0x19)
	utils.WriteVarInt(b, f.SequenceNumber)
	return nil
}

// Length of a written frame
func (f *RetireConnectionIDFrame) Length(protocol.VersionNumber) protocol.ByteCount {
	return 1 + utils.VarIntLen(f.SequenceNumber)
}
