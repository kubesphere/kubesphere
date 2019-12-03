package wire

import (
	"bytes"
	"fmt"
	"io"

	"github.com/lucas-clemente/quic-go/internal/utils"

	"github.com/lucas-clemente/quic-go/internal/protocol"
)

// A NewConnectionIDFrame is a NEW_CONNECTION_ID frame
type NewConnectionIDFrame struct {
	SequenceNumber      uint64
	RetirePriorTo       uint64
	ConnectionID        protocol.ConnectionID
	StatelessResetToken [16]byte
}

func parseNewConnectionIDFrame(r *bytes.Reader, _ protocol.VersionNumber) (*NewConnectionIDFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	seq, err := utils.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	ret, err := utils.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	if ret > seq {
		//nolint:stylecheck
		return nil, fmt.Errorf("Retire Prior To value (%d) larger than Sequence Number (%d)", ret, seq)
	}
	connIDLen, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if connIDLen > protocol.MaxConnIDLen {
		return nil, fmt.Errorf("invalid connection ID length: %d", connIDLen)
	}
	connID, err := protocol.ReadConnectionID(r, int(connIDLen))
	if err != nil {
		return nil, err
	}
	frame := &NewConnectionIDFrame{
		SequenceNumber: seq,
		RetirePriorTo:  ret,
		ConnectionID:   connID,
	}
	if _, err := io.ReadFull(r, frame.StatelessResetToken[:]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}

	return frame, nil
}

func (f *NewConnectionIDFrame) Write(b *bytes.Buffer, _ protocol.VersionNumber) error {
	b.WriteByte(0x18)
	utils.WriteVarInt(b, f.SequenceNumber)
	utils.WriteVarInt(b, f.RetirePriorTo)
	connIDLen := f.ConnectionID.Len()
	if connIDLen > protocol.MaxConnIDLen {
		return fmt.Errorf("invalid connection ID length: %d", connIDLen)
	}
	b.WriteByte(uint8(connIDLen))
	b.Write(f.ConnectionID.Bytes())
	b.Write(f.StatelessResetToken[:])
	return nil
}

// Length of a written frame
func (f *NewConnectionIDFrame) Length(protocol.VersionNumber) protocol.ByteCount {
	return 1 + utils.VarIntLen(f.SequenceNumber) + utils.VarIntLen(f.RetirePriorTo) + 1 /* connection ID length */ + protocol.ByteCount(f.ConnectionID.Len()) + 16
}
