// Copyright 2024 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

var RegoV1CompatibleRef = v1.RegoV1CompatibleRef

// RegoVersion defines the Rego syntax requirements for a module.
type RegoVersion = v1.RegoVersion

const DefaultRegoVersion = RegoV0

const (
	RegoUndefined = v1.RegoUndefined
	// RegoV0 is the default, original Rego syntax.
	RegoV0 = v1.RegoV0
	// RegoV0CompatV1 requires modules to comply with both the RegoV0 and RegoV1 syntax (as when 'rego.v1' is imported in a module).
	// Shortly, RegoV1 compatibility is required, but 'rego.v1' or 'future.keywords' must also be imported.
	RegoV0CompatV1 = v1.RegoV0CompatV1
	// RegoV1 is the Rego syntax enforced by OPA 1.0; e.g.:
	// future.keywords part of default keyword set, and don't require imports;
	// 'if' and 'contains' required in rule heads;
	// (some) strict checks on by default.
	RegoV1 = v1.RegoV1
)

func RegoVersionFromInt(i int) RegoVersion {
	return v1.RegoVersionFromInt(i)
}

// Parser is used to parse Rego statements.
type Parser = v1.Parser

// ParserOptions defines the options for parsing Rego statements.
type ParserOptions = v1.ParserOptions

// NewParser creates and initializes a Parser.
func NewParser() *Parser {
	return v1.NewParser().WithRegoVersion(DefaultRegoVersion)
}

func IsFutureKeyword(s string) bool {
	return v1.IsFutureKeywordForRegoVersion(s, RegoV0)
}
