// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package tokens

// Token represents a single Rego source code token
// for use by the Parser.
type Token int

func (t Token) String() string {
	if t < 0 || int(t) >= len(strings) {
		return "unknown"
	}
	return strings[t]
}

// All tokens must be defined here
const (
	Illegal Token = iota
	EOF
	Whitespace
	Ident
	Comment

	Package
	Import
	As
	Default
	Else
	Not
	Some
	With
	Null
	True
	False

	Number
	String

	LBrack
	RBrack
	LBrace
	RBrace
	LParen
	RParen
	Comma
	Colon

	Add
	Sub
	Mul
	Quo
	Rem
	And
	Or
	Unify
	Equal
	Assign
	Neq
	Gt
	Lt
	Gte
	Lte
	Dot
	Semicolon
)

var strings = [...]string{
	Illegal:    "illegal",
	EOF:        "eof",
	Whitespace: "whitespace",
	Comment:    "comment",
	Ident:      "ident",
	Package:    "package",
	Import:     "import",
	As:         "as",
	Default:    "default",
	Else:       "else",
	Not:        "not",
	Some:       "some",
	With:       "with",
	Null:       "null",
	True:       "true",
	False:      "false",
	Number:     "number",
	String:     "string",
	LBrack:     "[",
	RBrack:     "]",
	LBrace:     "{",
	RBrace:     "}",
	LParen:     "(",
	RParen:     ")",
	Comma:      ",",
	Colon:      ":",
	Add:        "plus",
	Sub:        "minus",
	Mul:        "mul",
	Quo:        "div",
	Rem:        "rem",
	And:        "and",
	Or:         "or",
	Unify:      "eq",
	Equal:      "equal",
	Assign:     "assign",
	Neq:        "neq",
	Gt:         "gt",
	Lt:         "lt",
	Gte:        "gte",
	Lte:        "lte",
	Dot:        ".",
	Semicolon:  ";",
}

var keywords = map[string]Token{
	"package": Package,
	"import":  Import,
	"as":      As,
	"default": Default,
	"else":    Else,
	"not":     Not,
	"some":    Some,
	"with":    With,
	"null":    Null,
	"true":    True,
	"false":   False,
}

// Keyword will return a token for the passed in
// literal value. If the value is a Rego keyword
// then the appropriate token is returned. Everything
// else is an Ident.
func Keyword(lit string) Token {
	if tok, ok := keywords[lit]; ok {
		return tok
	}
	return Ident
}
