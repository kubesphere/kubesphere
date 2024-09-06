package ast

import (
	astJSON "github.com/open-policy-agent/opa/ast/json"
)

// customJSON is an interface that can be implemented by AST nodes that
// allows the parser to set options for JSON operations on that node.
type customJSON interface {
	setJSONOptions(astJSON.Options)
}
