package dbr

import "bytes"

type Buffer interface {
	WriteString(s string) (n int, err error)
	String() string

	WriteValue(v ...interface{}) (err error)
	Value() []interface{}
}

type buffer struct {
	bytes.Buffer
	v []interface{}
}

func NewBuffer() Buffer {
	return &buffer{}
}

func (b *buffer) WriteValue(v ...interface{}) error {
	b.v = append(b.v, v...)
	return nil
}

func (b *buffer) Value() []interface{} {
	return b.v
}
