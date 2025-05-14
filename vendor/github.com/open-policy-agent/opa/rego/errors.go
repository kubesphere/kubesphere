package rego

import v1 "github.com/open-policy-agent/opa/v1/rego"

// HaltError is an error type to return from a custom function implementation
// that will abort the evaluation process (analogous to topdown.Halt).
type HaltError = v1.HaltError

// NewHaltError wraps an error such that the evaluation process will stop
// when it occurs.
func NewHaltError(err error) error {
	return v1.NewHaltError(err)
}

// ErrorDetails interface is satisfied by an error that provides further
// details.
type ErrorDetails = v1.ErrorDetails
