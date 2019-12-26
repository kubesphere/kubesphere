package errors

import "fmt"

// NewFieldRequired returns a *ValidationError indicating "value required"
func NewFieldRequired(field string) Error {
	return Error{Type: ErrorTypeRequired, Field: field}
}

// NewFieldInvalidValue returns a ValidationError indicating "invalid value"
func NewFieldInvalidValue(field string) Error {
	return Error{Type: ErrorInvalidValue, Field: field}
}

// NewFieldInvalidValueWithReason returns a ValidationError indicating "invalid value" and a reason for the error
func NewFieldInvalidValueWithReason(field, reason string) Error {
	return Error{Type: ErrorInvalidValue, Field: field, Reason: reason}
}

// ErrorType is a machine readable value providing more detail about why a field
// is invalid.
type ErrorType string

const (
	// ErrorTypeRequired is used to report required values that are not provided
	// (e.g. empty strings, null values, or empty arrays).
	ErrorTypeRequired ErrorType = "FieldValueRequired"

	// ErrorInvalidValue is used to report values that do not conform to the
	// expected schema.
	ErrorInvalidValue ErrorType = "InvalidValue"
)

// Error is an implementation of the 'error' interface, which represents an
// error of validation.
type Error struct {
	Type   ErrorType
	Field  string
	Reason string
}

func (v Error) Error() string {
	var msg string
	switch v.Type {
	case ErrorInvalidValue:
		msg = fmt.Sprintf("Invalid value specified for %q", v.Field)
	case ErrorTypeRequired:
		msg = fmt.Sprintf("Required value not specified for %q", v.Field)
	default:
		msg = fmt.Sprintf("%s: %s", v.Type, v.Field)
	}
	if len(v.Reason) > 0 {
		msg = fmt.Sprintf("%s: %s", msg, v.Reason)
	}
	return msg
}
