package client

import (
	"fmt"
	"strings"
)

// Errors is a list of error.
type Errors []error

// Errors implements error
var _ error = Errors{}

// Error implements error.
func (errs Errors) Error() string {
	s := make([]string, len(errs))
	for _, e := range errs {
		s = append(s, e.Error())
	}
	return strings.Join(s, "\n")
}

type ErrorMap map[string]error

func (e ErrorMap) Error() string {
	b := &strings.Builder{}
	for k, v := range e {
		fmt.Fprintf(b, "%s: %s\n", k, v)
	}
	return b.String()
}

type UnrecognizedConstraintError struct {
	s string
}

func (e *UnrecognizedConstraintError) Error() string {
	return fmt.Sprintf("Constraint kind %s is not recognized", e.s)
}

func NewUnrecognizedConstraintError(text string) error {
	return &UnrecognizedConstraintError{text}
}

func IsMissingConstraintError(e error) bool {
	_, ok := e.(*MissingConstraintError)
	return ok
}

type MissingConstraintError struct {
	s string
}

func (e *MissingConstraintError) Error() string {
	return fmt.Sprintf("Constraint kind %s is not recognized", e.s)
}

func NewMissingConstraintError(subPath string) error {
	return &MissingConstraintError{subPath}
}

func IsMissingTemplateError(e error) bool {
	_, ok := e.(*MissingTemplateError)
	return ok
}

type MissingTemplateError struct {
	s string
}

func (e *MissingTemplateError) Error() string {
	return fmt.Sprintf("Constraint kind %s is not recognized", e.s)
}

func NewMissingTemplateError(mapKey string) error {
	return &MissingTemplateError{mapKey}
}
