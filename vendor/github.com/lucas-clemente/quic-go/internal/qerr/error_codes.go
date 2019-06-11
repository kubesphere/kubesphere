package qerr

import (
	"fmt"

	"github.com/marten-seemann/qtls"
)

// ErrorCode can be used as a normal error without reason.
type ErrorCode uint16

// The error codes defined by QUIC
const (
	NoError                 ErrorCode = 0x0
	InternalError           ErrorCode = 0x1
	ServerBusy              ErrorCode = 0x2
	FlowControlError        ErrorCode = 0x3
	StreamLimitError        ErrorCode = 0x4
	StreamStateError        ErrorCode = 0x5
	FinalSizeError          ErrorCode = 0x6
	FrameEncodingError      ErrorCode = 0x7
	TransportParameterError ErrorCode = 0x8
	VersionNegotiationError ErrorCode = 0x9
	ProtocolViolation       ErrorCode = 0xa
	InvalidMigration        ErrorCode = 0xc
)

func (e ErrorCode) isCryptoError() bool {
	return e >= 0x100 && e < 0x200
}

func (e ErrorCode) Error() string {
	if e.isCryptoError() {
		return fmt.Sprintf("%s: %s", e.String(), qtls.Alert(e-0x100).Error())
	}
	return e.String()
}

func (e ErrorCode) String() string {
	switch e {
	case NoError:
		return "NO_ERROR"
	case InternalError:
		return "INTERNAL_ERROR"
	case ServerBusy:
		return "SERVER_BUSY"
	case FlowControlError:
		return "FLOW_CONTROL_ERROR"
	case StreamLimitError:
		return "STREAM_LIMIT_ERROR"
	case StreamStateError:
		return "STREAM_STATE_ERROR"
	case FinalSizeError:
		return "FINAL_SIZE_ERROR"
	case FrameEncodingError:
		return "FRAME_ENCODING_ERROR"
	case TransportParameterError:
		return "TRANSPORT_PARAMETER_ERROR"
	case VersionNegotiationError:
		return "VERSION_NEGOTIATION_ERROR"
	case ProtocolViolation:
		return "PROTOCOL_VIOLATION"
	case InvalidMigration:
		return "INVALID_MIGRATION"
	default:
		if e.isCryptoError() {
			return "CRYPTO_ERROR"
		}
		return fmt.Sprintf("unknown error code: %#x", uint16(e))
	}
}
