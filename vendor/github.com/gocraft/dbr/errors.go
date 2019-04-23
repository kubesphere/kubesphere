package dbr

import "errors"

// package errors
var (
	ErrNotFound           = errors.New("dbr: not found")
	ErrNotSupported       = errors.New("dbr: not supported")
	ErrTableNotSpecified  = errors.New("dbr: table not specified")
	ErrColumnNotSpecified = errors.New("dbr: column not specified")
	ErrInvalidPointer     = errors.New("dbr: attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("dbr: wrong placeholder count")
	ErrInvalidSliceLength = errors.New("dbr: length of slice is 0. length must be >= 1")
	ErrCantConvertToTime  = errors.New("dbr: can't convert to time.Time")
	ErrInvalidTimestring  = errors.New("dbr: invalid time string")
)
