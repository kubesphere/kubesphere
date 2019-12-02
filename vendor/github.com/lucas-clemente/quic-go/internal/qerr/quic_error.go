package qerr

import (
	"fmt"
	"net"
)

// A QuicError consists of an error code plus a error reason
type QuicError struct {
	ErrorCode          ErrorCode
	FrameType          uint64 // only valid if this not an application error
	ErrorMessage       string
	isTimeout          bool
	isApplicationError bool
}

var _ net.Error = &QuicError{}

// Error creates a new QuicError instance
func Error(errorCode ErrorCode, errorMessage string) *QuicError {
	return &QuicError{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

// ErrorWithFrameType creates a new QuicError instance for a specific frame type
func ErrorWithFrameType(errorCode ErrorCode, frameType uint64, errorMessage string) *QuicError {
	return &QuicError{
		ErrorCode:    errorCode,
		FrameType:    frameType,
		ErrorMessage: errorMessage,
	}
}

// TimeoutError creates a new QuicError instance for a timeout error
func TimeoutError(errorMessage string) *QuicError {
	return &QuicError{
		ErrorMessage: errorMessage,
		isTimeout:    true,
	}
}

// CryptoError create a new QuicError instance for a crypto error
func CryptoError(tlsAlert uint8, errorMessage string) *QuicError {
	return &QuicError{
		ErrorCode:    0x100 + ErrorCode(tlsAlert),
		ErrorMessage: errorMessage,
	}
}

// ApplicationError creates a new QuicError instance for an application error
func ApplicationError(errorCode ErrorCode, errorMessage string) *QuicError {
	return &QuicError{
		ErrorCode:          errorCode,
		ErrorMessage:       errorMessage,
		isApplicationError: true,
	}
}

func (e *QuicError) Error() string {
	if e.isApplicationError {
		if len(e.ErrorMessage) == 0 {
			return fmt.Sprintf("Application error %#x", uint64(e.ErrorCode))
		}
		return fmt.Sprintf("Application error %#x: %s", uint64(e.ErrorCode), e.ErrorMessage)
	}
	str := e.ErrorCode.String()
	if e.FrameType != 0 {
		str += fmt.Sprintf(" (frame type: %#x)", e.FrameType)
	}
	msg := e.ErrorMessage
	if len(msg) == 0 {
		msg = e.ErrorCode.Message()
	}
	if len(msg) == 0 {
		return str
	}
	return str + ": " + msg
}

// IsCryptoError says if this error is a crypto error
func (e *QuicError) IsCryptoError() bool {
	return e.ErrorCode.isCryptoError()
}

// IsApplicationError says if this error is an application error
func (e *QuicError) IsApplicationError() bool {
	return e.isApplicationError
}

// Temporary says if the error is temporary.
func (e *QuicError) Temporary() bool {
	return false
}

// Timeout says if this error is a timeout.
func (e *QuicError) Timeout() bool {
	return e.isTimeout
}

// ToQuicError converts an arbitrary error to a QuicError. It leaves QuicErrors
// unchanged, and properly handles `ErrorCode`s.
func ToQuicError(err error) *QuicError {
	switch e := err.(type) {
	case *QuicError:
		return e
	case ErrorCode:
		return Error(e, "")
	}
	return Error(InternalError, err.Error())
}
