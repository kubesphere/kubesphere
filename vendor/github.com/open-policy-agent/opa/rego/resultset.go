package rego

import (
	v1 "github.com/open-policy-agent/opa/v1/rego"
)

// ResultSet represents a collection of output from Rego evaluation. An empty
// result set represents an undefined query.
type ResultSet = v1.ResultSet

// Vars represents a collection of variable bindings. The keys are the variable
// names and the values are the binding values.
type Vars = v1.Vars

// Result defines the output of Rego evaluation.
type Result = v1.Result

// Location defines a position in a Rego query or module.
type Location = v1.Location

// ExpressionValue defines the value of an expression in a Rego query.
type ExpressionValue = v1.ExpressionValue
