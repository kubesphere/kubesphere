// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// This file contains extra functions for parsing Rego.
// Most of the parsing is handled by the code in parser.go,
// however, there are additional utilities that are
// helpful for dealing with Rego source inputs (e.g., REPL
// statements, source files, etc.)

package ast

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// MustParseBody returns a parsed body.
// If an error occurs during parsing, panic.
func MustParseBody(input string) Body {
	return MustParseBodyWithOpts(input, ParserOptions{})
}

// MustParseBodyWithOpts returns a parsed body.
// If an error occurs during parsing, panic.
func MustParseBodyWithOpts(input string, opts ParserOptions) Body {
	parsed, err := ParseBodyWithOpts(input, opts)
	if err != nil {
		panic(err)
	}
	return parsed
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
	parsed, err := ParseModuleWithOpts("", input, opts)
	if err != nil {
		panic(err)
	}
	return parsed
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

	if len(body) != 1 {
		return nil, fmt.Errorf("multiple expressions cannot be used for rule head")
	}

	return ParseRuleFromExpr(module, body[0])
}

// ParseRuleFromExpr returns a rule if the expression can be interpreted as a
// rule definition.
func ParseRuleFromExpr(module *Module, expr *Expr) (*Rule, error) {

	if len(expr.With) > 0 {
		return nil, fmt.Errorf("expressions using with keyword cannot be used for rule head")
	}

	if expr.Negated {
		return nil, fmt.Errorf("negated expressions cannot be used for rule head")
	}

	if _, ok := expr.Terms.(*SomeDecl); ok {
		return nil, errors.New("'some' declarations cannot be used for rule head")
	}

	if term, ok := expr.Terms.(*Term); ok {
		switch v := term.Value.(type) {
		case Ref:
			if len(v) > 2 { // 2+ dots
				return ParseCompleteDocRuleWithDotsFromTerm(module, term)
			}
			return ParsePartialSetDocRuleFromTerm(module, term)
		default:
			return nil, fmt.Errorf("%v cannot be used for rule name", TypeName(v))
		}
	}

	if _, ok := expr.Terms.([]*Term); !ok {
		// This is a defensive check in case other kinds of expression terms are
		// introduced in the future.
		return nil, errors.New("expression cannot be used for rule head")
	}

	if expr.IsEquality() {
		return parseCompleteRuleFromEq(module, expr)
	} else if expr.IsAssignment() {
		rule, err := parseCompleteRuleFromEq(module, expr)
		if err != nil {
			return nil, err
		}
		rule.Head.Assign = true
		return rule, nil
	}

	if _, ok := BuiltinMap[expr.Operator().String()]; ok {
		return nil, fmt.Errorf("rule name conflicts with built-in function")
	}

	return ParseRuleFromCallExpr(module, expr.Terms.([]*Term))
}

func parseCompleteRuleFromEq(module *Module, expr *Expr) (rule *Rule, err error) {

	// ensure the rule location is set to the expr location
	// the helper functions called below try to set the location based
	// on the terms they've been provided but that is not as accurate.
	defer func() {
		if rule != nil {
			rule.Location = expr.Location
			rule.Head.Location = expr.Location
		}
	}()

	lhs, rhs := expr.Operand(0), expr.Operand(1)
	if lhs == nil || rhs == nil {
		return nil, errors.New("assignment requires two operands")
	}

	rule, err = ParseRuleFromCallEqExpr(module, lhs, rhs)
	if err == nil {
		return rule, nil
	}

	rule, err = ParsePartialObjectDocRuleFromEqExpr(module, lhs, rhs)
	if err == nil {
		return rule, nil
	}

	return ParseCompleteDocRuleFromEqExpr(module, lhs, rhs)
}

// ParseCompleteDocRuleFromAssignmentExpr returns a rule if the expression can
// be interpreted as a complete document definition declared with the assignment
// operator.
func ParseCompleteDocRuleFromAssignmentExpr(module *Module, lhs, rhs *Term) (*Rule, error) {

	rule, err := ParseCompleteDocRuleFromEqExpr(module, lhs, rhs)
	if err != nil {
		return nil, err
	}

	rule.Head.Assign = true

	return rule, nil
}

// ParseCompleteDocRuleFromEqExpr returns a rule if the expression can be
// interpreted as a complete document definition.
func ParseCompleteDocRuleFromEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	var head *Head

	if v, ok := lhs.Value.(Var); ok {
		head = NewHead(v)
	} else if r, ok := lhs.Value.(Ref); ok { // groundness ?
		if _, ok := r[0].Value.(Var); !ok {
			return nil, fmt.Errorf("invalid rule head: %v", r)
		}
		head = RefHead(r)
		if len(r) > 1 && !r[len(r)-1].IsGround() {
			return nil, fmt.Errorf("ref not ground")
		}
	} else {
		return nil, fmt.Errorf("%v cannot be used for rule name", TypeName(lhs.Value))
	}
	head.Value = rhs
	head.Location = lhs.Location

	return &Rule{
		Location: lhs.Location,
		Head:     head,
		Body: NewBody(
			NewExpr(BooleanTerm(true).SetLocation(rhs.Location)).SetLocation(rhs.Location),
		),
		Module: module,
	}, nil
}

func ParseCompleteDocRuleWithDotsFromTerm(module *Module, term *Term) (*Rule, error) {
	ref, ok := term.Value.(Ref)
	if !ok {
		return nil, fmt.Errorf("%v cannot be used for rule name", TypeName(term.Value))
	}

	if _, ok := ref[0].Value.(Var); !ok {
		return nil, fmt.Errorf("invalid rule head: %v", ref)
	}
	head := RefHead(ref, BooleanTerm(true).SetLocation(term.Location))
	head.Location = term.Location

	return &Rule{
		Location: term.Location,
		Head:     head,
		Body: NewBody(
			NewExpr(BooleanTerm(true).SetLocation(term.Location)).SetLocation(term.Location),
		),
		Module: module,
	}, nil
}

// ParsePartialObjectDocRuleFromEqExpr returns a rule if the expression can be
// interpreted as a partial object document definition.
func ParsePartialObjectDocRuleFromEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {
	ref, ok := lhs.Value.(Ref)
	if !ok {
		return nil, fmt.Errorf("%v cannot be used as rule name", TypeName(lhs.Value))
	}

	if _, ok := ref[0].Value.(Var); !ok {
		return nil, fmt.Errorf("invalid rule head: %v", ref)
	}

	head := RefHead(ref, rhs)
	if len(ref) == 2 { // backcompat for naked `foo.bar = "baz"` statements
		head.Name = ref[0].Value.(Var)
		head.Key = ref[1]
	}
	head.Location = rhs.Location

	rule := &Rule{
		Location: rhs.Location,
		Head:     head,
		Body: NewBody(
			NewExpr(BooleanTerm(true).SetLocation(rhs.Location)).SetLocation(rhs.Location),
		),
		Module: module,
	}

	return rule, nil
}

// ParsePartialSetDocRuleFromTerm returns a rule if the term can be interpreted
// as a partial set document definition.
func ParsePartialSetDocRuleFromTerm(module *Module, term *Term) (*Rule, error) {

	ref, ok := term.Value.(Ref)
	if !ok || len(ref) == 1 {
		return nil, fmt.Errorf("%vs cannot be used for rule head", TypeName(term.Value))
	}
	if _, ok := ref[0].Value.(Var); !ok {
		return nil, fmt.Errorf("invalid rule head: %v", ref)
	}

	head := RefHead(ref)
	if len(ref) == 2 {
		v, ok := ref[0].Value.(Var)
		if !ok {
			return nil, fmt.Errorf("%vs cannot be used for rule head", TypeName(term.Value))
		}
		head = NewHead(v)
		head.Key = ref[1]
	}
	head.Location = term.Location

	rule := &Rule{
		Location: term.Location,
		Head:     head,
		Body: NewBody(
			NewExpr(BooleanTerm(true).SetLocation(term.Location)).SetLocation(term.Location),
		),
		Module: module,
	}

	return rule, nil
}

// ParseRuleFromCallEqExpr returns a rule if the term can be interpreted as a
// function definition (e.g., f(x) = y => f(x) = y { true }).
func ParseRuleFromCallEqExpr(module *Module, lhs, rhs *Term) (*Rule, error) {

	call, ok := lhs.Value.(Call)
	if !ok {
		return nil, fmt.Errorf("must be call")
	}

	ref, ok := call[0].Value.(Ref)
	if !ok {
		return nil, fmt.Errorf("%vs cannot be used in function signature", TypeName(call[0].Value))
	}
	if _, ok := ref[0].Value.(Var); !ok {
		return nil, fmt.Errorf("invalid rule head: %v", ref)
	}

	head := RefHead(ref, rhs)
	head.Location = lhs.Location
	head.Args = Args(call[1:])

	rule := &Rule{
		Location: lhs.Location,
		Head:     head,
		Body:     NewBody(NewExpr(BooleanTerm(true).SetLocation(rhs.Location)).SetLocation(rhs.Location)),
		Module:   module,
	}

	return rule, nil
}

// ParseRuleFromCallExpr returns a rule if the terms can be interpreted as a
// function returning true or some value (e.g., f(x) => f(x) = true { true }).
func ParseRuleFromCallExpr(module *Module, terms []*Term) (*Rule, error) {

	if len(terms) <= 1 {
		return nil, fmt.Errorf("rule argument list must take at least one argument")
	}

	loc := terms[0].Location
	ref := terms[0].Value.(Ref)
	if _, ok := ref[0].Value.(Var); !ok {
		return nil, fmt.Errorf("invalid rule head: %v", ref)
	}
	head := RefHead(ref, BooleanTerm(true).SetLocation(loc))
	head.Location = loc
	head.Args = terms[1:]

	rule := &Rule{
		Location: loc,
		Head:     head,
		Module:   module,
		Body:     NewBody(NewExpr(BooleanTerm(true).SetLocation(loc)).SetLocation(loc)),
	}
	return rule, nil
}

// ParseImports returns a slice of Import objects.
func ParseImports(input string) ([]*Import, error) {
	stmts, _, err := ParseStatements("", input)
	if err != nil {
		return nil, err
	}
	result := []*Import{}
	for _, stmt := range stmts {
		if imp, ok := stmt.(*Import); ok {
			result = append(result, imp)
		} else {
			return nil, fmt.Errorf("expected import but got %T", stmt)
		}
	}
	return result, nil
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
	stmts, comments, err := ParseStatementsWithOpts(filename, input, popts)
	if err != nil {
		return nil, err
	}
	return parseModule(filename, stmts, comments)
}

// ParseBody returns exactly one body.
// If multiple bodies are parsed, an error is returned.
func ParseBody(input string) (Body, error) {
	return ParseBodyWithOpts(input, ParserOptions{SkipRules: true})
}

// ParseBodyWithOpts returns exactly one body. It does _not_ set SkipRules: true on its own,
// but respects whatever ParserOptions it's been given.
func ParseBodyWithOpts(input string, popts ParserOptions) (Body, error) {

	stmts, _, err := ParseStatementsWithOpts("", input, popts)
	if err != nil {
		return nil, err
	}

	result := Body{}

	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case Body:
			for i := range stmt {
				result.Append(stmt[i])
			}
		case *Comment:
			// skip
		default:
			return nil, fmt.Errorf("expected body but got %T", stmt)
		}
	}

	return result, nil
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
	stmt, err := ParseStatement(input)
	if err != nil {
		return nil, err
	}
	pkg, ok := stmt.(*Package)
	if !ok {
		return nil, fmt.Errorf("expected package but got %T", stmt)
	}
	return pkg, nil
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
	stmts, _, err := ParseStatementsWithOpts("", input, opts)
	if err != nil {
		return nil, err
	}
	if len(stmts) != 1 {
		return nil, fmt.Errorf("expected exactly one statement (rule), got %v = %T, %T", stmts, stmts[0], stmts[1])
	}
	rule, ok := stmts[0].(*Rule)
	if !ok {
		return nil, fmt.Errorf("expected rule but got %T", stmts[0])
	}
	return rule, nil
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
		return nil, fmt.Errorf("expected exactly one statement")
	}
	return stmts[0], nil
}

// ParseStatements is deprecated. Use ParseStatementWithOpts instead.
func ParseStatements(filename, input string) ([]Statement, []*Comment, error) {
	return ParseStatementsWithOpts(filename, input, ParserOptions{})
}

// ParseStatementsWithOpts returns a slice of parsed statements. This is the
// default return value from the parser.
func ParseStatementsWithOpts(filename, input string, popts ParserOptions) ([]Statement, []*Comment, error) {

	parser := NewParser().
		WithFilename(filename).
		WithReader(bytes.NewBufferString(input)).
		WithProcessAnnotation(popts.ProcessAnnotation).
		WithFutureKeywords(popts.FutureKeywords...).
		WithAllFutureKeywords(popts.AllFutureKeywords).
		WithCapabilities(popts.Capabilities).
		WithSkipRules(popts.SkipRules).
		withUnreleasedKeywords(popts.unreleasedKeywords)

	stmts, comments, errs := parser.Parse()

	if len(errs) > 0 {
		return nil, nil, errs
	}

	return stmts, comments, nil
}

func parseModule(filename string, stmts []Statement, comments []*Comment) (*Module, error) {

	if len(stmts) == 0 {
		return nil, NewError(ParseErr, &Location{File: filename}, "empty module")
	}

	var errs Errors

	pkg, ok := stmts[0].(*Package)
	if !ok {
		loc := stmts[0].Loc()
		errs = append(errs, NewError(ParseErr, loc, "package expected"))
	}

	mod := &Module{
		Package: pkg,
		stmts:   stmts,
	}

	// The comments slice only holds comments that were not their own statements.
	mod.Comments = append(mod.Comments, comments...)

	for i, stmt := range stmts[1:] {
		switch stmt := stmt.(type) {
		case *Import:
			mod.Imports = append(mod.Imports, stmt)
		case *Rule:
			setRuleModule(stmt, mod)
			mod.Rules = append(mod.Rules, stmt)
		case Body:
			rule, err := ParseRuleFromBody(mod, stmt)
			if err != nil {
				errs = append(errs, NewError(ParseErr, stmt[0].Location, err.Error()))
				continue
			}
			mod.Rules = append(mod.Rules, rule)

			// NOTE(tsandall): the statement should now be interpreted as a
			// rule so update the statement list. This is important for the
			// logic below that associates annotations with statements.
			stmts[i+1] = rule
		case *Package:
			errs = append(errs, NewError(ParseErr, stmt.Loc(), "unexpected package"))
		case *Annotations:
			mod.Annotations = append(mod.Annotations, stmt)
		case *Comment:
			// Ignore comments, they're handled above.
		default:
			panic("illegal value") // Indicates grammar is out-of-sync with code.
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	errs = append(errs, attachAnnotationsNodes(mod)...)

	if len(errs) > 0 {
		return nil, errs
	}

	return mod, nil
}

func newScopeAttachmentErr(a *Annotations, want string) *Error {
	var have string
	if a.node != nil {
		have = fmt.Sprintf(" (have %v)", TypeName(a.node))
	}
	return NewError(ParseErr, a.Loc(), "annotation scope '%v' must be applied to %v%v", a.Scope, want, have)
}

func setRuleModule(rule *Rule, module *Module) {
	rule.Module = module
	if rule.Else != nil {
		setRuleModule(rule.Else, module)
	}
}

// ParserErrorDetail holds additional details for parser errors.
type ParserErrorDetail struct {
	Line string `json:"line"`
	Idx  int    `json:"idx"`
}

func newParserErrorDetail(bs []byte, offset int) *ParserErrorDetail {

	// Find first non-space character at or before offset position.
	if offset >= len(bs) {
		offset = len(bs) - 1
	} else if offset < 0 {
		offset = 0
	}

	for offset > 0 && unicode.IsSpace(rune(bs[offset])) {
		offset--
	}

	// Find beginning of line containing offset.
	begin := offset

	for begin > 0 && !isNewLineChar(bs[begin]) {
		begin--
	}

	if isNewLineChar(bs[begin]) {
		begin++
	}

	// Find end of line containing offset.
	end := offset

	for end < len(bs) && !isNewLineChar(bs[end]) {
		end++
	}

	if begin > end {
		begin = end
	}

	// Extract line and compute index of offset byte in line.
	line := bs[begin:end]
	index := offset - begin

	return &ParserErrorDetail{
		Line: string(line),
		Idx:  index,
	}
}

// Lines returns the pretty formatted line output for the error details.
func (d ParserErrorDetail) Lines() []string {
	line := strings.TrimLeft(d.Line, "\t") // remove leading tabs
	tabCount := len(d.Line) - len(line)
	indent := d.Idx - tabCount
	if indent < 0 {
		indent = 0
	}
	return []string{line, strings.Repeat(" ", indent) + "^"}
}

func isNewLineChar(b byte) bool {
	return b == '\r' || b == '\n'
}
