package graph

import (
	"net/http"
)

type Response struct {
	Message string
	Code    int
}

// Error panics with InternalServerError and the provided message
func Error(message string) {
	Panic(message, http.StatusInternalServerError)
}

// BadRequest panics with BadRequest and the provided message
func BadRequest(message string) {
	Panic(message, http.StatusBadRequest)
}

// Forbidden panics with Forbidden and the provided message
func Forbidden(message string) {
	Panic(message, http.StatusForbidden)
}

// Panic panics with the provided HTTP response code and message
func Panic(message string, code int) Response {
	panic(Response{
		Message: message,
		Code:    code,
	})
}

// CheckError panics with the supplied error if it is non-nil
func CheckError(err error) {
	if err != nil {
		panic(err.Error)
	}
}

// IsOK just validates that a telemetry label value is not empty or unknown
func IsOK(telemetryVal string) bool {
	return telemetryVal != "" && telemetryVal != Unknown
}
