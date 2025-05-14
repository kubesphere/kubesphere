// Copyright 2024 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"errors"
	"fmt"

	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// MustParseBody returns a parsed body.
// If an error occurs during parsing, panic.
func MustParseBody(input string) Body {
	return MustParseBodyWithOpts(input, ParserOptions{})
}

// MustParseBodyWithOpts returns a parsed body.
// If an error occurs during parsing, panic.
func MustParseBodyWithOpts(input string, opts ParserOptions) Body {
	return v1.MustParseBodyWithOpts(input, setDefaultRegoVersion(opts))
}

// MustParseExpr returns a parsed expression.
// If an error occurs during parsing, panic.
func MustParseExpr(input string) *Expr {
	parsed, err := ParseExpr(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseImports returns a slice of imports.
// If an error occurs during parsing, panic.
func MustParseImports(input string) []*Import {
	parsed, err := ParseImports(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseModule returns a parsed module.
// If an error occurs during parsing, panic.
func MustParseModule(input string) *Module {
	return MustParseModuleWithOpts(input, ParserOptions{})
}

// MustParseModuleWithOpts returns a parsed module.
// If an error occurs during parsing, panic.
func MustParseModuleWithOpts(input string, opts ParserOptions) *Module {
	return v1.MustParseModuleWithOpts(input, setDefaultRegoVersion(opts))
}

// MustParsePackage returns a Package.
// If an error occurs during parsing, panic.
func MustParsePackage(input string) *Package {
	parsed, err := ParsePackage(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseStatements returns a slice of parsed statements.
// If an error occurs during parsing, panic.
func MustParseStatements(input string) []Statement {
	parsed, _, err := ParseStatements("", input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseStatement returns exactly one statement.
// If an error occurs during parsing, panic.
func MustParseStatement(input string) Statement {
	parsed, err := ParseStatement(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

func MustParseStatementWithOpts(input string, popts ParserOptions) Statement {
	return v1.MustParseStatementWithOpts(input, setDefaultRegoVersion(popts))
}

// MustParseRef returns a parsed reference.
// If an error occurs during parsing, panic.
func MustParseRef(input string) Ref {
	parsed, err := ParseRef(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseRule returns a parsed rule.
// If an error occurs during parsing, panic.
func MustParseRule(input string) *Rule {
	parsed, err := ParseRule(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// MustParseRuleWithOpts returns a parsed rule.
// If an error occurs during parsing, panic.
func MustParseRuleWithOpts(input string, opts ParserOptions) *Rule {
	return v1.MustParseRuleWithOpts(input, setDefaultRegoVersion(opts))
}

// MustParseTerm returns a parsed term.
// If an error occurs during parsing, panic.
func MustParseTerm(input string) *Term {
	parsed, err := ParseTerm(input)
	if err != nil {
		panic(err)
	}
	return parsed
}

// ParseRuleFromBody returns a rule if the body can be interpreted as a rule
// definition. Otherwise, an error is returned.
func ParseRuleFromBody(module *Module, body Body) (*Rule, error) {
	return v1.ParseRuleFromBody(module, body)
}

// ParseRuleFromExpr returns a rule if the expression can be interpreted as a
// rule definition.
func ParseRuleFromExpr(module *Module, expr *Expr) (*Rule, error) {
	return v1.ParseRuleFromExpr(module, expr)
}

// ParseCompleteDocRuleFromAssignmentExpr returns a rule if the expression can
// be interpreted as a complete document definition declared with the assignment
// operator.
func ParseCompleteDocRuleFromAssignmentExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	return v1.ParseCompleteDocRuleFromAssignmentExpr(module, lhs, rhs)
}

// ParseCompleteDocRuleFromEqExpr returns a rule if the expression can be
// interpreted as a complete document definition.
func ParseCompleteDocRuleFromEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	return v1.ParseCompleteDocRuleFromEqExpr(module, lhs, rhs)
}

func ParseCompleteDocRuleWithDotsFromTerm(module *Module, term *Term) (*Rule, error) {
	return v1.ParseCompleteDocRuleWithDotsFromTerm(module, term)
}

// ParsePartialObjectDocRuleFromEqExpr returns a rule if the expression can be
// interpreted as a partial object document definition.
func ParsePartialObjectDocRuleFromEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	return v1.ParsePartialObjectDocRuleFromEqExpr(module, lhs, rhs)
}

// ParsePartialSetDocRuleFromTerm returns a rule if the term can be interpreted
// as a partial set document definition.
func ParsePartialSetDocRuleFromTerm(module *Module, term *Term) (*Rule, error) {
	return v1.ParsePartialSetDocRuleFromTerm(module, term)
}

// ParseRuleFromCallEqExpr returns a rule if the term can be interpreted as a
// function definition (e.g., f(x) = y => f(x) = y { true }).
func ParseRuleFromCallEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	return v1.ParseRuleFromCallEqExpr(module, lhs, rhs)
}

// ParseRuleFromCallExpr returns a rule if the terms can be interpreted as a
// function returning true or some value (e.g., f(x) => f(x) = true { true }).
func ParseRuleFromCallExpr(module *Module, terms []*Term) (*Rule, error) {
	return v1.ParseRuleFromCallExpr(module, terms)
}

// ParseImports returns a slice of Import objects.
func ParseImports(input string) ([]*Import, error) {
	return v1.ParseImports(input)
}

// ParseModule returns a parsed Module object.
// For details on Module objects and their fields, see policy.go.
// Empty input will return nil, nil.
func ParseModule(filename, input string) (*Module, error) {
	return ParseModuleWithOpts(filename, input, ParserOptions{})
}

// ParseModuleWithOpts returns a parsed Module object, and has an additional input ParserOptions
// For details on Module objects and their fields, see policy.go.
// Empty input will return nil, nil.
func ParseModuleWithOpts(filename, input string, popts ParserOptions) (*Module, error) {
	return v1.ParseModuleWithOpts(filename, input, setDefaultRegoVersion(popts))
}

// ParseBody returns exactly one body.
// If multiple bodies are parsed, an error is returned.
func ParseBody(input string) (Body, error) {
	return ParseBodyWithOpts(input, ParserOptions{SkipRules: true})
}

// ParseBodyWithOpts returns exactly one body. It does _not_ set SkipRules: true on its own,
// but respects whatever ParserOptions it's been given.
func ParseBodyWithOpts(input string, popts ParserOptions) (Body, error) {
	return v1.ParseBodyWithOpts(input, setDefaultRegoVersion(popts))
}

// ParseExpr returns exactly one expression.
// If multiple expressions are parsed, an error is returned.
func ParseExpr(input string) (*Expr, error) {
	body, err := ParseBody(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}
	if len(body) != 1 {
		return nil, fmt.Errorf("expected exactly one expression but got: %v", body)
	}
	return body[0], nil
}

// ParsePackage returns exactly one Package.
// If multiple statements are parsed, an error is returned.
func ParsePackage(input string) (*Package, error) {
	return v1.ParsePackage(input)
}

// ParseTerm returns exactly one term.
// If multiple terms are parsed, an error is returned.
func ParseTerm(input string) (*Term, error) {
	body, err := ParseBody(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse term: %w", err)
	}
	if len(body) != 1 {
		return nil, fmt.Errorf("expected exactly one term but got: %v", body)
	}
	term, ok := body[0].Terms.(*Term)
	if !ok {
		return nil, fmt.Errorf("expected term but got %v", body[0].Terms)
	}
	return term, nil
}

// ParseRef returns exactly one reference.
func ParseRef(input string) (Ref, error) {
	term, err := ParseTerm(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ref: %w", err)
	}
	ref, ok := term.Value.(Ref)
	if !ok {
		return nil, fmt.Errorf("expected ref but got %v", term)
	}
	return ref, nil
}

// ParseRuleWithOpts returns exactly one rule.
// If multiple rules are parsed, an error is returned.
func ParseRuleWithOpts(input string, opts ParserOptions) (*Rule, error) {
	return v1.ParseRuleWithOpts(input, setDefaultRegoVersion(opts))
}

// ParseRule returns exactly one rule.
// If multiple rules are parsed, an error is returned.
func ParseRule(input string) (*Rule, error) {
	return ParseRuleWithOpts(input, ParserOptions{})
}

// ParseStatement returns exactly one statement.
// A statement might be a term, expression, rule, etc. Regardless,
// this function expects *exactly* one statement. If multiple
// statements are parsed, an error is returned.
func ParseStatement(input string) (Statement, error) {
	stmts, _, err := ParseStatements("", input)
	if err != nil {
		return nil, err
	}
	if len(stmts) != 1 {
		return nil, errors.New("expected exactly one statement")
	}
	return stmts[0], nil
}

func ParseStatementWithOpts(input string, popts ParserOptions) (Statement, error) {
	return v1.ParseStatementWithOpts(input, setDefaultRegoVersion(popts))
}

// ParseStatements is deprecated. Use ParseStatementWithOpts instead.
func ParseStatements(filename, input string) ([]Statement, []*Comment, error) {
	return ParseStatementsWithOpts(filename, input, ParserOptions{})
}

// ParseStatementsWithOpts returns a slice of parsed statements. This is the
// default return value from the parser.
func ParseStatementsWithOpts(filename, input string, popts ParserOptions) ([]Statement, []*Comment, error) {
	return v1.ParseStatementsWithOpts(filename, input, setDefaultRegoVersion(popts))
}

// ParserErrorDetail holds additional details for parser errors.
type ParserErrorDetail = v1.ParserErrorDetail

func setDefaultRegoVersion(opts ParserOptions) ParserOptions {
	if opts.RegoVersion == RegoUndefined {
		opts.RegoVersion = DefaultRegoVersion
	}
	return opts
}
