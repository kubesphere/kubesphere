package qpack

import (
	"io"
)

// An Encoder performs QPACK encoding.
type Encoder struct {
	wrotePrefix bool

	w   io.Writer
	buf []byte
}

// NewEncoder returns a new Encoder which performs QPACK encoding. An
// encoded data is written to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// WriteField encodes f into a single Write to e's underlying Writer.
// This function may also produce bytes for the Header Block Prefix
// if necessary. If produced, it is done before encoding f.
func (e *Encoder) WriteField(f HeaderField) error {
	// write the Header Block Prefix
	if !e.wrotePrefix {
		e.buf = appendVarInt(e.buf, 8, 0)
		e.buf = appendVarInt(e.buf, 7, 0)
		e.wrotePrefix = true
	}

	e.writeLiteralFieldWithoutNameReference(f)
	e.w.Write(e.buf)
	e.buf = e.buf[:0]
	return nil
}

// Close declares that the encoding is complete and resets the Encoder
// to be reused again for a new header block.
func (e *Encoder) Close() error {
	e.wrotePrefix = false
	return nil
}

func (e *Encoder) writeLiteralFieldWithoutNameReference(f HeaderField) {
	offset := len(e.buf)
	e.buf = appendVarInt(e.buf, 3, uint64(len(f.Name)))
	e.buf[offset] ^= 0x20
	e.buf = append(e.buf, []byte(f.Name)...)
	e.buf = appendVarInt(e.buf, 7, uint64(len(f.Value)))
	e.buf = append(e.buf, []byte(f.Value)...)
}
