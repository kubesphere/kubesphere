package csvutil

// Reader provides the interface for reading a single CSV record.
//
// If there is no data left to be read, Read returns (nil, io.EOF).
//
// It is implemented by csv.Reader.
type Reader interface {
	Read() ([]string, error)
}

// Writer provides the interface for writing a single CSV record.
//
// It is implemented by csv.Writer.
type Writer interface {
	Write([]string) error
}

// Unmarshaler is the interface implemented by types that can unmarshal
// a single record's field description of themselves.
type Unmarshaler interface {
	UnmarshalCSV([]byte) error
}

// Marshaler is the interface implemented by types that can marshal themselves
// into valid string.
type Marshaler interface {
	MarshalCSV() ([]byte, error)
}
