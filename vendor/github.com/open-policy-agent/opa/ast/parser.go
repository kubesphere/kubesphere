// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"github.com/open-policy-agent/opa/ast/internal/scanner"
	"github.com/open-policy-agent/opa/ast/internal/tokens"
	"github.com/open-policy-agent/opa/ast/location"
)

// Note: This state is kept isolated from the parser so that we
// can do efficient shallow copies of these values when doing a
// save() and restore().
type state struct {
	s         *scanner.Scanner
	lastEnd   int
	skippedNL bool
	tok       tokens.Token
	tokEnd    int
	lit       string
	loc       Location
	errors    Errors
	comments  []*Comment
	wildcard  int
}

func (s *state) String() string {
	return fmt.Sprintf("<s: %v, tok: %v, lit: %q, loc: %v, errors: %d, comments: %d>", s.s, s.tok, s.lit, s.loc, len(s.errors), len(s.comments))
}

func (s *state) Loc() *location.Location {
	cpy := s.loc
	return &cpy
}

func (s *state) Text(offset, end int) []byte {
	bs := s.s.Bytes()
	if offset >= 0 && offset < len(bs) {
		if end >= offset && end <= len(bs) {
			return bs[offset:end]
		}
	}
	return nil
}

// Parser is used to parse Rego statements.
type Parser struct {
	r io.Reader
	s *state
}

// NewParser creates and initializes a Parser.
func NewParser() *Parser {
	p := &Parser{s: &state{}}
	return p
}

// WithFilename provides the filename for Location details
// on parsed statements.
func (p *Parser) WithFilename(filename string) *Parser {
	p.s.loc.File = filename
	return p
}

// WithReader provides the io.Reader that the parser will
// use as its source.
func (p *Parser) WithReader(r io.Reader) *Parser {
	p.r = r
	return p
}

// Parse will read the Rego source and parse statements and
// comments as they are found. Any errors encountered while
// parsing will be accumulated and returned as a list of Errors.
func (p *Parser) Parse() ([]Statement, []*Comment, Errors) {

	var err error
	p.s.s, err = scanner.New(p.r)
	if err != nil {
		return nil, nil, Errors{
			&Error{
				Code:     ParseErr,
				Message:  err.Error(),
				Location: nil,
			},
		}
	}

	// read the first token to initialize the parser
	p.scan()

	var stmts []Statement

	// Read from the scanner until the last token is reached or no statements
	// can be parsed. Attempt to parse package statements, import statements,
	// rule statements, and then body/query statements (in that order). If a
	// statement cannot be parsed, restore the parser state before trying the
	// next type of statement. If a statement can be parsed, continue from that
	// point trying to parse packages, imports, etc. in the same order.
	for p.s.tok != tokens.EOF {

		s := p.save()

		if pkg := p.parsePackage(); pkg != nil {
			stmts = append(stmts, pkg)
			continue
		} else if len(p.s.errors) > 0 {
			break
		}

		p.restore(s)
		s = p.save()

		if imp := p.parseImport(); imp != nil {
			stmts = append(stmts, imp)
			continue
		} else if len(p.s.errors) > 0 {
			break
		}

		p.restore(s)
		s = p.save()

		if rules := p.parseRules(); rules != nil {
			for i := range rules {
				stmts = append(stmts, rules[i])
			}
			continue
		} else if len(p.s.errors) > 0 {
			break
		}

		p.restore(s)
		s = p.save()

		if body := p.parseQuery(true, tokens.EOF); body != nil {
			stmts = append(stmts, body)
			continue
		}

		break
	}

	return stmts, p.s.comments, p.s.errors
}

func (p *Parser) parsePackage() *Package {

	var pkg Package
	pkg.SetLoc(p.s.Loc())

	if p.s.tok != tokens.Package {
		return nil
	}

	p.scan()
	if p.s.tok != tokens.Ident {
		p.illegalToken()
		return nil
	}

	term := p.parseTerm()

	if term != nil {
		switch v := term.Value.(type) {
		case Var:
			pkg.Path = Ref{
				DefaultRootDocument.Copy().SetLocation(term.Location),
				StringTerm(string(v)).SetLocation(term.Location),
			}
		case Ref:
			pkg.Path = make(Ref, len(v)+1)
			pkg.Path[0] = DefaultRootDocument.Copy().SetLocation(v[0].Location)
			first, ok := v[0].Value.(Var)
			if !ok {
				p.errorf(v[0].Location, "unexpected %v token: expecting var", TypeName(v[0].Value))
				return nil
			}
			pkg.Path[1] = StringTerm(string(first)).SetLocation(v[0].Location)
			for i := 2; i < len(pkg.Path); i++ {
				switch v[i-1].Value.(type) {
				case String:
					pkg.Path[i] = v[i-1]
				default:
					p.errorf(v[i-1].Location, "unexpected %v token: expecting string", TypeName(v[i-1].Value))
					return nil
				}
			}
		default:
			p.illegalToken()
			return nil
		}
	}

	if pkg.Path == nil {
		if len(p.s.errors) == 0 {
			p.error(p.s.Loc(), "expected path")
		}
		return nil
	}

	return &pkg
}

func (p *Parser) parseImport() *Import {

	var imp Import
	imp.SetLoc(p.s.Loc())

	if p.s.tok != tokens.Import {
		return nil
	}

	p.scan()
	if p.s.tok != tokens.Ident {
		p.error(p.s.Loc(), "expected ident")
		return nil
	}

	term := p.parseTerm()
	if term != nil {
		switch v := term.Value.(type) {
		case Var:
			imp.Path = RefTerm(term).SetLocation(term.Location)
		case Ref:
			for i := 1; i < len(v); i++ {
				if _, ok := v[i].Value.(String); !ok {
					p.errorf(v[i].Location, "unexpected %v token: expecting string", TypeName(v[i].Value))
					return nil
				}
			}
			imp.Path = term
		}
	}

	if imp.Path == nil {
		p.error(p.s.Loc(), "expected path")
		return nil
	}

	path := imp.Path.Value.(Ref)

	if !RootDocumentNames.Contains(path[0]) {
		p.errorf(imp.Path.Location, "unexpected import path, must begin with one of: %v, got: %v", RootDocumentNames, path[0])
		return nil
	}

	if p.s.tok == tokens.As {
		p.scan()

		if p.s.tok != tokens.Ident {
			p.illegal("expected var")
			return nil
		}

		alias := p.parseTerm()

		v, ok := alias.Value.(Var)
		if !ok {
			p.illegal("expected var")
			return nil
		}
		imp.Alias = v
	}

	return &imp
}

func (p *Parser) parseRules() []*Rule {

	var rule Rule
	rule.SetLoc(p.s.Loc())

	if p.s.tok == tokens.Default {
		p.scan()
		rule.Default = true
	}

	if p.s.tok != tokens.Ident {
		return nil
	}

	if rule.Head = p.parseHead(rule.Default); rule.Head == nil {
		return nil
	}

	if rule.Default {
		if !p.validateDefaultRuleValue(&rule) {
			return nil
		}

		rule.Body = NewBody(NewExpr(BooleanTerm(true).SetLocation(rule.Location)).SetLocation(rule.Location))
		return []*Rule{&rule}
	}

	if p.s.tok == tokens.LBrace {
		p.scan()
		if rule.Body = p.parseBody(tokens.RBrace); rule.Body == nil {
			return nil
		}
		p.scan()
	} else {
		return nil
	}

	if p.s.tok == tokens.Else {

		if rule.Head.Assign {
			p.error(p.s.Loc(), "else keyword cannot be used on rule declared with := operator")
			return nil
		}

		if rule.Head.Key != nil {
			p.error(p.s.Loc(), "else keyword cannot be used on partial rules")
			return nil
		}

		if rule.Else = p.parseElse(rule.Head); rule.Else == nil {
			return nil
		}
	}

	rule.Location.Text = p.s.Text(rule.Location.Offset, p.s.lastEnd)

	var rules []*Rule

	rules = append(rules, &rule)

	for p.s.tok == tokens.LBrace {

		if rule.Else != nil {
			p.error(p.s.Loc(), "expected else keyword")
			return nil
		}

		loc := p.s.Loc()

		p.scan()
		var next Rule

		if next.Body = p.parseBody(tokens.RBrace); next.Body == nil {
			return nil
		}
		p.scan()

		loc.Text = p.s.Text(loc.Offset, p.s.lastEnd)
		next.SetLoc(loc)

		// Chained rule head's keep the original
		// rule's head AST but have their location
		// set to the rule body.
		next.Head = rule.Head.Copy()
		setLocRecursive(next.Head, loc)

		rules = append(rules, &next)
	}

	return rules
}

func (p *Parser) parseElse(head *Head) *Rule {

	var rule Rule
	rule.SetLoc(p.s.Loc())

	rule.Head = head.Copy()
	rule.Head.SetLoc(p.s.Loc())

	defer func() {
		rule.Location.Text = p.s.Text(rule.Location.Offset, p.s.lastEnd)
	}()

	p.scan()

	switch p.s.tok {
	case tokens.LBrace:
		rule.Head.Value = BooleanTerm(true)
	case tokens.Unify:
		p.scan()
		rule.Head.Value = p.parseTermRelation()
		if rule.Head.Value == nil {
			return nil
		}
		rule.Head.Location.Text = p.s.Text(rule.Head.Location.Offset, p.s.lastEnd)
	default:
		p.illegal("expected else value term or rule body")
		return nil
	}

	if p.s.tok != tokens.LBrace {
		rule.Body = NewBody(NewExpr(BooleanTerm(true)))
		setLocRecursive(rule.Body, rule.Location)
		return &rule
	}

	p.scan()

	if rule.Body = p.parseBody(tokens.RBrace); rule.Body == nil {
		return nil
	}

	p.scan()

	if p.s.tok == tokens.Else {
		if rule.Else = p.parseElse(head); rule.Else == nil {
			return nil
		}
	}
	return &rule
}

func (p *Parser) parseHead(defaultRule bool) *Head {

	var head Head
	head.SetLoc(p.s.Loc())

	defer func() {
		head.Location.Text = p.s.Text(head.Location.Offset, p.s.lastEnd)
	}()

	if term := p.parseVar(); term != nil {
		if v, ok := term.Value.(Var); ok {
			head.Name = v
		}
	}
	if head.Name == "" {
		p.illegal("expected rule head name")
	}

	p.scan()

	if p.s.tok == tokens.LParen {
		p.scan()
		if p.s.tok != tokens.RParen {
			head.Args = p.parseTermList(tokens.RParen, nil)
			if head.Args == nil {
				return nil
			}
		}
		p.scan()

		if p.s.tok == tokens.LBrack {
			return nil
		}
	}

	if p.s.tok == tokens.LBrack {
		p.scan()
		head.Key = p.parseTermRelation()
		if head.Key == nil {
			p.illegal("expected rule key term (e.g., %s[<VALUE>] { ... })", head.Name)
		}
		if p.s.tok != tokens.RBrack {
			p.illegal("non-terminated rule key")
		}
		p.scan()
	}

	if p.s.tok == tokens.Unify {
		p.scan()
		head.Value = p.parseTermRelation()
		if head.Value == nil {
			p.illegal("expected rule value term (e.g., %s[<VALUE>] { ... })", head.Name)
		}
	} else if p.s.tok == tokens.Assign {

		if defaultRule {
			p.error(p.s.Loc(), "default rules must use = operator (not := operator)")
			return nil
		} else if head.Key != nil {
			p.error(p.s.Loc(), "partial rules must use = operator (not := operator)")
			return nil
		} else if len(head.Args) > 0 {
			p.error(p.s.Loc(), "functions must use = operator (not := operator)")
			return nil
		}

		p.scan()
		head.Assign = true
		head.Value = p.parseTermRelation()
		if head.Value == nil {
			p.illegal("expected rule value term (e.g., %s := <VALUE> { ... })", head.Name)
		}
	}

	if head.Value == nil && head.Key == nil {
		head.Value = BooleanTerm(true).SetLocation(head.Location)
	}

	return &head
}

func (p *Parser) parseBody(end tokens.Token) Body {
	return p.parseQuery(false, end)
}

func (p *Parser) parseQuery(requireSemi bool, end tokens.Token) Body {
	body := Body{}

	if p.s.tok == end {
		p.error(p.s.Loc(), "found empty body")
		return nil
	}

	for {

		expr := p.parseLiteral()
		if expr == nil {
			return nil
		}

		body.Append(expr)

		if p.s.tok == tokens.Semicolon {
			p.scan()
			continue
		}

		if p.s.tok == end || requireSemi {
			return body
		}

		if !p.s.skippedNL {
			// If there was already an error then don't pile this one on
			if len(p.s.errors) == 0 {
				p.illegal(`expected \n or %s or %s`, tokens.Semicolon, end)
			}
			return nil
		}
	}
}

func (p *Parser) parseLiteral() (expr *Expr) {

	offset := p.s.loc.Offset
	loc := p.s.Loc()

	defer func() {
		if expr != nil {
			loc.Text = p.s.Text(offset, p.s.lastEnd)
			expr.SetLoc(loc)
		}
	}()

	var negated bool
	switch p.s.tok {
	case tokens.Some:
		return p.parseSome()
	case tokens.Not:
		p.scan()
		negated = true
		fallthrough
	default:
		expr := p.parseExpr()
		if expr != nil {
			expr.Negated = negated
			if p.s.tok == tokens.With {
				if expr.With = p.parseWith(); expr.With == nil {
					return nil
				}
			}
			return expr
		}
		return nil
	}
}

func (p *Parser) parseWith() []*With {

	withs := []*With{}

	for {

		with := With{
			Location: p.s.Loc(),
		}
		p.scan()

		if p.s.tok != tokens.Ident {
			p.illegal("expected ident")
			return nil
		}

		if with.Target = p.parseTerm(); with.Target == nil {
			return nil
		}

		switch with.Target.Value.(type) {
		case Ref, Var:
			break
		default:
			p.illegal("expected with target path")
		}

		if p.s.tok != tokens.As {
			p.illegal("expected as keyword")
			return nil
		}

		p.scan()

		if with.Value = p.parseTermRelation(); with.Value == nil {
			return nil
		}

		with.Location.Text = p.s.Text(with.Location.Offset, p.s.lastEnd)

		withs = append(withs, &with)

		if p.s.tok != tokens.With {
			break
		}
	}

	return withs
}

func (p *Parser) parseSome() *Expr {

	decl := &SomeDecl{}
	decl.SetLoc(p.s.Loc())

	for {

		p.scan()

		switch p.s.tok {
		case tokens.Ident:
		}

		if p.s.tok != tokens.Ident {
			p.illegal("expected var")
			return nil
		}

		decl.Symbols = append(decl.Symbols, p.parseVar())

		p.scan()

		if p.s.tok != tokens.Comma {
			break
		}
	}

	return NewExpr(decl).SetLocation(decl.Location)
}

func (p *Parser) parseExpr() *Expr {

	lhs := p.parseTermRelation()

	if lhs == nil {
		return nil
	}

	if op := p.parseTermOp(tokens.Assign, tokens.Unify); op != nil {
		if rhs := p.parseTermRelation(); rhs != nil {
			return NewExpr([]*Term{op, lhs, rhs})
		}
		return nil
	}

	// NOTE(tsandall): the top-level call term is converted to an expr because
	// the evaluator does not support the call term type (nested calls are
	// rewritten by the compiler.)
	if call, ok := lhs.Value.(Call); ok {
		return NewExpr([]*Term(call))
	}

	return NewExpr(lhs)
}

// parseTermRelation consumes the next term from the input and returns it. If a
// term cannot be parsed the return value is nil and error will be recorded. The
// scanner will be advanced to the next token before returning.
func (p *Parser) parseTermRelation() *Term {
	return p.parseTermRelationRec(nil, p.s.loc.Offset)
}

func (p *Parser) parseTermRelationRec(lhs *Term, offset int) *Term {
	if lhs == nil {
		lhs = p.parseTermOr(nil, offset)
	}
	if lhs != nil {
		if op := p.parseTermOp(tokens.Equal, tokens.Neq, tokens.Lt, tokens.Gt, tokens.Lte, tokens.Gte); op != nil {
			if rhs := p.parseTermOr(nil, p.s.loc.Offset); rhs != nil {
				call := p.setLoc(CallTerm(op, lhs, rhs), lhs.Location, offset, p.s.lastEnd)
				switch p.s.tok {
				case tokens.Equal, tokens.Neq, tokens.Lt, tokens.Gt, tokens.Lte, tokens.Gte:
					return p.parseTermRelationRec(call, offset)
				default:
					return call
				}
			}
		}
	}
	return lhs
}

func (p *Parser) parseTermOr(lhs *Term, offset int) *Term {
	if lhs == nil {
		lhs = p.parseTermAnd(nil, offset)
	}
	if lhs != nil {
		if op := p.parseTermOp(tokens.Or); op != nil {
			if rhs := p.parseTermAnd(nil, p.s.loc.Offset); rhs != nil {
				call := p.setLoc(CallTerm(op, lhs, rhs), lhs.Location, offset, p.s.lastEnd)
				switch p.s.tok {
				case tokens.Or:
					return p.parseTermOr(call, offset)
				default:
					return call
				}
			}
		}
		return lhs
	}
	return nil
}

func (p *Parser) parseTermAnd(lhs *Term, offset int) *Term {
	if lhs == nil {
		lhs = p.parseTermArith(nil, offset)
	}
	if lhs != nil {
		if op := p.parseTermOp(tokens.And); op != nil {
			if rhs := p.parseTermArith(nil, p.s.loc.Offset); rhs != nil {
				call := p.setLoc(CallTerm(op, lhs, rhs), lhs.Location, offset, p.s.lastEnd)
				switch p.s.tok {
				case tokens.And:
					return p.parseTermAnd(call, offset)
				default:
					return call
				}
			}
		}
		return lhs
	}
	return nil
}

func (p *Parser) parseTermArith(lhs *Term, offset int) *Term {
	if lhs == nil {
		lhs = p.parseTermFactor(nil, offset)
	}
	if lhs != nil {
		if op := p.parseTermOp(tokens.Add, tokens.Sub); op != nil {
			if rhs := p.parseTermFactor(nil, p.s.loc.Offset); rhs != nil {
				call := p.setLoc(CallTerm(op, lhs, rhs), lhs.Location, offset, p.s.lastEnd)
				switch p.s.tok {
				case tokens.Add, tokens.Sub:
					return p.parseTermArith(call, offset)
				default:
					return call
				}
			}
		}
	}
	return lhs
}

func (p *Parser) parseTermFactor(lhs *Term, offset int) *Term {
	if lhs == nil {
		lhs = p.parseTerm()
	}
	if lhs != nil {
		if op := p.parseTermOp(tokens.Mul, tokens.Quo, tokens.Rem); op != nil {
			if rhs := p.parseTerm(); rhs != nil {
				call := p.setLoc(CallTerm(op, lhs, rhs), lhs.Location, offset, p.s.lastEnd)
				switch p.s.tok {
				case tokens.Mul, tokens.Quo, tokens.Rem:
					return p.parseTermFactor(call, offset)
				default:
					return call
				}
			}
		}
	}
	return lhs
}

func (p *Parser) parseTerm() *Term {
	var term *Term
	switch p.s.tok {
	case tokens.Null:
		term = NullTerm().SetLocation(p.s.Loc())
	case tokens.True:
		term = BooleanTerm(true).SetLocation(p.s.Loc())
	case tokens.False:
		term = BooleanTerm(false).SetLocation(p.s.Loc())
	case tokens.Sub, tokens.Dot, tokens.Number:
		term = p.parseNumber()
	case tokens.String:
		term = p.parseString()
	case tokens.Ident:
		term = p.parseVar()
	case tokens.LBrack:
		term = p.parseArray()
	case tokens.LBrace:
		term = p.parseSetOrObject()
	case tokens.LParen:
		offset := p.s.loc.Offset
		p.scan()
		if r := p.parseTermRelation(); r != nil {
			if p.s.tok == tokens.RParen {
				r.Location.Text = p.s.Text(offset, p.s.tokEnd)
				term = r
			} else {
				p.error(p.s.Loc(), "non-terminated expression")
			}
		}
	default:
		p.illegalToken()
		return nil
	}

	return p.parseTermFinish(term)
}

func (p *Parser) parseTermFinish(head *Term) *Term {
	if head == nil {
		return nil
	}
	offset := p.s.loc.Offset
	p.scanWS()
	switch p.s.tok {
	case tokens.LParen, tokens.Dot, tokens.LBrack:
		return p.parseRef(head, offset)
	case tokens.Whitespace:
		p.scan()
		fallthrough
	default:
		if _, ok := head.Value.(Var); ok && RootDocumentNames.Contains(head) {
			return RefTerm(head).SetLocation(head.Location)
		}
		return head
	}
}

func (p *Parser) parseNumber() *Term {
	var prefix string
	loc := p.s.Loc()
	if p.s.tok == tokens.Sub {
		prefix = "-"
		p.scan()
		switch p.s.tok {
		case tokens.Number, tokens.Dot:
			break
		default:
			p.illegal("expected number")
			return nil
		}
	}
	if p.s.tok == tokens.Dot {
		prefix += "."
		p.scan()
		if p.s.tok != tokens.Number {
			p.illegal("expected number")
			return nil
		}
	}

	// Ensure that the number is valid
	s := prefix + p.s.lit
	f, ok := new(big.Float).SetString(s)
	if !ok {
		p.illegal("expected number")
		return nil
	}

	// Put limit on size of exponent to prevent non-linear cost of String()
	// function on big.Float from causing denial of service: https://github.com/golang/go/issues/11068
	//
	// n == sign * mantissa * 2^exp
	// 0.5 <= mantissa < 1.0
	//
	// The limit is arbitrary.
	exp := f.MantExp(nil)
	if exp > 1e5 || exp < -1e5 {
		p.error(p.s.Loc(), "number too big")
		return nil
	}

	// Note: Use the original string, do *not* round trip from
	// the big.Float as it can cause precision loss.
	r := NumberTerm(json.Number(s)).SetLocation(loc)
	return r
}

func (p *Parser) parseString() *Term {
	if p.s.lit[0] == '"' {
		var s string
		err := json.Unmarshal([]byte(p.s.lit), &s)
		if err != nil {
			p.errorf(p.s.Loc(), "illegal string literal: %s", p.s.lit)
			return nil
		}
		term := StringTerm(s).SetLocation(p.s.Loc())
		return term
	}
	return p.parseRawString()
}

func (p *Parser) parseRawString() *Term {
	if len(p.s.lit) < 2 {
		return nil
	}
	term := StringTerm(p.s.lit[1 : len(p.s.lit)-1]).SetLocation(p.s.Loc())
	return term
}

// this is the name to use for instantiating an empty set, e.g., `set()`.
var setConstructor = RefTerm(VarTerm("set"))

func (p *Parser) parseCall(operator *Term, offset int) (term *Term) {

	loc := operator.Location
	var end int

	defer func() {
		p.setLoc(term, loc, offset, end)
	}()

	p.scan()

	if p.s.tok == tokens.RParen {
		end = p.s.tokEnd
		p.scanWS()
		if operator.Equal(setConstructor) {
			return SetTerm()
		}
		return CallTerm(operator)
	}

	if r := p.parseTermList(tokens.RParen, []*Term{operator}); r != nil {
		end = p.s.tokEnd
		p.scanWS()
		return CallTerm(r...)
	}

	return nil
}

func (p *Parser) parseRef(head *Term, offset int) (term *Term) {

	loc := head.Location
	var end int

	defer func() {
		p.setLoc(term, loc, offset, end)
	}()

	switch h := head.Value.(type) {
	case Var, *Array, Object, Set, *ArrayComprehension, *ObjectComprehension, *SetComprehension, Call:
		// ok
	default:
		p.errorf(loc, "illegal ref (head cannot be %v)", TypeName(h))
	}

	ref := []*Term{head}

	for {
		switch p.s.tok {
		case tokens.Dot:
			p.scanWS()
			if p.s.tok != tokens.Ident {
				p.illegal("expected %v", tokens.Ident)
				return nil
			}
			ref = append(ref, StringTerm(p.s.lit).SetLocation(p.s.Loc()))
			p.scanWS()
		case tokens.LParen:
			term = p.parseCall(p.setLoc(RefTerm(ref...), loc, offset, p.s.loc.Offset), offset)
			if term != nil {
				switch p.s.tok {
				case tokens.Whitespace:
					p.scan()
					end = p.s.lastEnd
					return term
				case tokens.Dot, tokens.LBrack:
					term = p.parseRef(term, offset)
				}
			}
			end = p.s.tokEnd
			return term
		case tokens.LBrack:
			p.scan()
			if term := p.parseTermRelation(); term != nil {
				if p.s.tok != tokens.RBrack {
					p.illegal("expected %v", tokens.LBrack)
					return nil
				}
				ref = append(ref, term)
				p.scanWS()
			} else {
				return nil
			}
		case tokens.Whitespace:
			end = p.s.lastEnd
			p.scan()
			return RefTerm(ref...)
		default:
			end = p.s.lastEnd
			return RefTerm(ref...)
		}
	}
}

func (p *Parser) parseArray() (term *Term) {

	loc := p.s.Loc()
	offset := p.s.loc.Offset

	defer func() {
		p.setLoc(term, loc, offset, p.s.tokEnd)
	}()

	p.scan()

	if p.s.tok == tokens.RBrack {
		return ArrayTerm()
	}

	potentialComprehension := true

	// Skip leading commas, eg [, x, y]
	// Supported for backwards compatibility. In the future
	// we should make this a parse error.
	if p.s.tok == tokens.Comma {
		potentialComprehension = false
		p.scan()
	}

	s := p.save()

	// NOTE(tsandall): The parser cannot attempt a relational term here because
	// of ambiguity around comprehensions. For example, given:
	//
	//  {1 | 1}
	//
	// Does this represent a set comprehension or a set containing binary OR
	// call? We resolve the ambiguity by prioritizing comprehensions.
	head := p.parseTerm()

	if head == nil {
		return nil
	}

	switch p.s.tok {
	case tokens.RBrack:
		return ArrayTerm(head)
	case tokens.Comma:
		p.scan()
		if terms := p.parseTermList(tokens.RBrack, []*Term{head}); terms != nil {
			return NewTerm(NewArray(terms...))
		}
		return nil
	case tokens.Or:
		if potentialComprehension {
			// Try to parse as if it is an array comprehension
			p.scan()
			if body := p.parseBody(tokens.RBrack); body != nil {
				return ArrayComprehensionTerm(head, body)
			}
			if p.s.tok != tokens.Comma {
				return nil
			}
		}
		// fall back to parsing as a normal array definition
		fallthrough
	default:
		p.restore(s)
		if terms := p.parseTermList(tokens.RBrack, nil); terms != nil {
			return NewTerm(NewArray(terms...))
		}
		return nil
	}
}

func (p *Parser) parseSetOrObject() (term *Term) {

	loc := p.s.Loc()
	offset := p.s.loc.Offset

	defer func() {
		p.setLoc(term, loc, offset, p.s.tokEnd)
	}()

	p.scan()

	if p.s.tok == tokens.RBrace {
		return ObjectTerm()
	}

	potentialComprehension := true

	// Skip leading commas, eg {, x, y}
	// Supported for backwards compatibility. In the future
	// we should make this a parse error.
	if p.s.tok == tokens.Comma {
		potentialComprehension = false
		p.scan()
	}

	s := p.save()

	// Try parsing just a single term first to give comprehensions higher
	// priority to "or" calls in ambiguous situations. Eg: { a | b }
	// will be a set comprehension.
	//
	// Note: We don't know yet if it is a set or object being defined.
	head := p.parseTerm()

	if head == nil {
		return nil
	}

	switch p.s.tok {
	case tokens.Or:
		if potentialComprehension {
			return p.parseSet(s, head, potentialComprehension)
		}
	case tokens.RBrace, tokens.Comma:
		return p.parseSet(s, head, potentialComprehension)
	case tokens.Colon:
		return p.parseObject(head, potentialComprehension)
	}

	p.restore(s)

	if head = p.parseTermRelation(); head == nil {
		return nil
	}

	switch p.s.tok {
	case tokens.RBrace, tokens.Comma:
		return p.parseSet(s, head, false)
	case tokens.Colon:
		// It still might be an object comprehension, eg { a+1: b | ... }
		return p.parseObject(head, potentialComprehension)
	default:
		p.illegal("non-terminated set")
	}

	return nil
}

func (p *Parser) parseSet(s *state, head *Term, potentialComprehension bool) *Term {
	switch p.s.tok {
	case tokens.RBrace:
		return SetTerm(head)
	case tokens.Comma:
		p.scan()
		if terms := p.parseTermList(tokens.RBrace, []*Term{head}); terms != nil {
			return SetTerm(terms...)
		}
		return nil
	case tokens.Or:
		if potentialComprehension {
			// Try to parse as if it is a set comprehension
			p.scan()
			if body := p.parseBody(tokens.RBrace); body != nil {
				return SetComprehensionTerm(head, body)
			}
			if p.s.tok != tokens.Comma {
				return nil
			}
		}
		// Fall back to parsing as normal set definition
		p.restore(s)
		if terms := p.parseTermList(tokens.RBrace, nil); terms != nil {
			return SetTerm(terms...)
		}
		return nil
	}
	return nil
}

func (p *Parser) parseObject(k *Term, potentialComprehension bool) *Term {
	// NOTE(tsandall): Assumption: this function is called after parsing the key
	// of the head element and then receiving a colon token from the scanner.
	// Advance beyond the colon and attempt to parse an object.
	p.scan()

	s := p.save()
	v := p.parseTerm()

	if v == nil {
		return nil
	}

	switch p.s.tok {
	case tokens.RBrace, tokens.Comma, tokens.Or:
		if potentialComprehension {
			if term := p.parseObjectFinish(k, v, true); term != nil {
				return term
			}
		}
	}

	p.restore(s)

	if v = p.parseTermRelation(); v == nil {
		return nil
	}

	switch p.s.tok {
	case tokens.Comma, tokens.RBrace:
		return p.parseObjectFinish(k, v, false)
	default:
		p.illegal("non-terminated object")
	}

	return nil
}

func (p *Parser) parseObjectFinish(key, val *Term, potentialComprehension bool) *Term {
	switch p.s.tok {
	case tokens.RBrace:
		return ObjectTerm([2]*Term{key, val})
	case tokens.Or:
		if potentialComprehension {
			p.scan()
			if body := p.parseBody(tokens.RBrace); body != nil {
				return ObjectComprehensionTerm(key, val, body)
			}
		} else {
			p.illegal("non-terminated object")
		}
	case tokens.Comma:
		p.scan()
		if r := p.parseTermPairList(tokens.RBrace, [][2]*Term{{key, val}}); r != nil {
			return ObjectTerm(r...)
		}
	}
	return nil
}

func (p *Parser) parseTermList(end tokens.Token, r []*Term) []*Term {
	if p.s.tok == end {
		return r
	}
	for {
		term := p.parseTermRelation()
		if term != nil {
			r = append(r, term)
			switch p.s.tok {
			case end:
				return r
			case tokens.Comma:
				p.scan()
				if p.s.tok == end {
					return r
				}
				continue
			default:
				p.illegal(fmt.Sprintf("expected %q or %q", tokens.Comma, end))
				return nil
			}
		}
		return nil
	}
}

func (p *Parser) parseTermPairList(end tokens.Token, r [][2]*Term) [][2]*Term {
	if p.s.tok == end {
		return r
	}
	for {
		key := p.parseTermRelation()
		if key != nil {
			switch p.s.tok {
			case tokens.Colon:
				p.scan()
				if val := p.parseTermRelation(); val != nil {
					r = append(r, [2]*Term{key, val})
					switch p.s.tok {
					case end:
						return r
					case tokens.Comma:
						p.scan()
						if p.s.tok == end {
							return r
						}
						continue
					default:
						p.illegal(fmt.Sprintf("expected %q or %q", tokens.Comma, end))
						return nil
					}
				}
			default:
				p.illegal(fmt.Sprintf("expected %q", tokens.Colon))
				return nil
			}
		}
		return nil
	}
}

func (p *Parser) parseTermOp(values ...tokens.Token) *Term {
	for i := range values {
		if p.s.tok == values[i] {
			r := RefTerm(VarTerm(fmt.Sprint(p.s.tok)).SetLocation(p.s.Loc())).SetLocation(p.s.Loc())
			p.scan()
			return r
		}
	}
	return nil
}

func (p *Parser) parseVar() *Term {

	s := p.s.lit

	term := VarTerm(s).SetLocation(p.s.Loc())

	// Update wildcard values with unique identifiers
	if term.Equal(Wildcard) {
		term.Value = Var(p.genwildcard())
	}

	return term
}

func (p *Parser) genwildcard() string {
	c := p.s.wildcard
	p.s.wildcard++
	return fmt.Sprintf("%v%d", WildcardPrefix, c)
}

func (p *Parser) error(loc *location.Location, reason string) {
	p.errorf(loc, reason)
}

func (p *Parser) errorf(loc *location.Location, f string, a ...interface{}) {
	p.s.errors = append(p.s.errors, &Error{
		Code:     ParseErr,
		Message:  fmt.Sprintf(f, a...),
		Location: loc,
		Details:  newParserErrorDetail(p.s.s.Bytes(), loc.Offset),
	})
}

func (p *Parser) illegal(note string, a ...interface{}) {

	tok := p.s.tok.String()

	if p.s.tok == tokens.Illegal {
		p.errorf(p.s.Loc(), "illegal token")
		return
	}

	tokType := "token"
	if p.s.tok >= tokens.Package && p.s.tok <= tokens.False {
		tokType = "keyword"
	}

	note = fmt.Sprintf(note, a...)
	if len(note) > 0 {
		p.errorf(p.s.Loc(), "unexpected %s %s: %v", tok, tokType, note)
	} else {
		p.errorf(p.s.Loc(), "unexpected %s %s", tok, tokType)
	}
}

func (p *Parser) illegalToken() {
	p.illegal("")
}

func (p *Parser) scan() {
	p.doScan(true)
}

func (p *Parser) scanWS() {
	p.doScan(false)
}

func (p *Parser) doScan(skipws bool) {

	// NOTE(tsandall): the last position is used to compute the "text" field for
	// complex AST nodes. Whitespace never affects the last position of an AST
	// node so do not update it when scanning.
	if p.s.tok != tokens.Whitespace {
		p.s.lastEnd = p.s.tokEnd
		p.s.skippedNL = false
	}

	var errs []scanner.Error
	for {
		var pos scanner.Position
		p.s.tok, pos, p.s.lit, errs = p.s.s.Scan()

		p.s.tokEnd = pos.End
		p.s.loc.Row = pos.Row
		p.s.loc.Col = pos.Col
		p.s.loc.Offset = pos.Offset
		p.s.loc.Text = p.s.Text(pos.Offset, pos.End)

		for _, err := range errs {
			p.error(p.s.Loc(), err.Message)
		}

		if len(errs) > 0 {
			p.s.tok = tokens.Illegal
		}

		if p.s.tok == tokens.Whitespace {
			if p.s.lit == "\n" {
				p.s.skippedNL = true
			}
			if skipws {
				continue
			}
		}

		if p.s.tok != tokens.Comment {
			break
		}

		// For backwards compatibility leave a nil
		// Text value if there is no text rather than
		// an empty string.
		var commentText []byte
		if len(p.s.lit) > 1 {
			commentText = []byte(p.s.lit[1:])
		}
		comment := NewComment(commentText)
		comment.SetLoc(p.s.Loc())
		p.s.comments = append(p.s.comments, comment)
	}
}

func (p *Parser) save() *state {
	cpy := *p.s
	s := *cpy.s
	cpy.s = &s
	return &cpy
}

func (p *Parser) restore(s *state) {
	p.s = s
}

func setLocRecursive(x interface{}, loc *location.Location) {
	NewGenericVisitor(func(x interface{}) bool {
		if node, ok := x.(Node); ok {
			node.SetLoc(loc)
		}
		return false
	}).Walk(x)
}

func (p *Parser) setLoc(term *Term, loc *location.Location, offset, end int) *Term {
	if term != nil {
		cpy := *loc
		term.Location = &cpy
		term.Location.Text = p.s.Text(offset, end)
	}
	return term
}

func (p *Parser) validateDefaultRuleValue(rule *Rule) bool {
	if rule.Head.Value == nil {
		p.error(rule.Loc(), fmt.Sprintf("illegal default rule (must have a value)"))
		return false
	}

	valid := true
	vis := NewGenericVisitor(func(x interface{}) bool {
		switch x.(type) {
		case *ArrayComprehension, *ObjectComprehension, *SetComprehension: // skip closures
			return true
		case Ref, Var, Call:
			p.error(rule.Loc(), fmt.Sprintf("illegal default rule (value cannot contain %v)", TypeName(x)))
			valid = false
			return true
		}
		return false
	})

	vis.Walk(rule.Head.Value.Value)
	return valid
}
