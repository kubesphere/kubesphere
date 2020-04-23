package errors

import "errors"

var (
	ErrClientNotEnabled = errors.New("Client not enabled")
)

func IsClientNotEnabledError(err error) bool {
	if err == ErrClientNotEnabled {
		return true
	} else {
		return false
	}
}
