/*
Copyright 2016 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
// Printing of syntax trees.

package build

import (
	"bytes"
	"fmt"
	"strings"
)

// Format returns the formatted form of the given BUILD file.
func Format(f *File) []byte {
	pr := &printer{}
	pr.file(f)
	return pr.Bytes()
}

// FormatString returns the string form of the given expression.
func FormatString(x Expr) string {
	pr := &printer{}
	switch x := x.(type) {
	case *File:
		pr.file(x)
	default:
		pr.expr(x, precLow)
	}
	return pr.String()
}

// A printer collects the state during printing of a file or expression.
type printer struct {
	bytes.Buffer           // output buffer
	comment      []Comment // pending end-of-line comments
	margin       int       // left margin (indent), a number of spaces
	depth        int       // nesting depth inside ( ) [ ] { }
}

// printf prints to the buffer.
func (p *printer) printf(format string, args ...interface{}) {
	fmt.Fprintf(p, format, args...)
}

// indent returns the position on the current line, in bytes, 0-indexed.
func (p *printer) indent() int {
	b := p.Bytes()
	n := 0
	for n < len(b) && b[len(b)-1-n] != '\n' {
		n++
	}
	return n
}

// newline ends the current line, flushing end-of-line comments.
// It must only be called when printing a newline is known to be safe:
// when not inside an expression or when p.depth > 0.
// To break a line inside an expression that might not be enclosed
// in brackets of some kind, use breakline instead.
func (p *printer) newline() {
	if len(p.comment) > 0 {
		p.printf("  ")
		for i, com := range p.comment {
			if i > 0 {
				p.trim()
				p.printf("\n%*s", p.margin, "")
			}
			p.printf("%s", strings.TrimSpace(com.Token))
		}
		p.comment = p.comment[:0]
	}

	p.trim()
	p.printf("\n%*s", p.margin, "")
}

// breakline breaks the current line, inserting a continuation \ if needed.
// If no continuation \ is needed, breakline flushes end-of-line comments.
func (p *printer) breakline() {
	if p.depth == 0 {
		// Cannot have both final \ and comments.
		p.printf(" \\\n%*s", p.margin, "")
		return
	}

	// Safe to use newline.
	p.newline()
}

// trim removes trailing spaces from the current line.
func (p *printer) trim() {
	// Remove trailing space from line we're about to end.
	b := p.Bytes()
	n := len(b)
	for n > 0 && b[n-1] == ' ' {
		n--
	}
	p.Truncate(n)
}

// file formats the given file into the print buffer.
func (p *printer) file(f *File) {
	for _, com := range f.Before {
		p.printf("%s", strings.TrimSpace(com.Token))
		p.newline()
	}

	for i, stmt := range f.Stmt {
		switch stmt := stmt.(type) {
		case *CommentBlock:
			// comments already handled

		case *PythonBlock:
			for _, com := range stmt.Before {
				p.printf("%s", strings.TrimSpace(com.Token))
				p.newline()
			}
			p.printf("%s", stmt.Token) // includes trailing newline

		default:
			p.expr(stmt, precLow)
			p.newline()
		}

		for _, com := range stmt.Comment().After {
			p.printf("%s", strings.TrimSpace(com.Token))
			p.newline()
		}

		if i+1 < len(f.Stmt) && !compactStmt(stmt, f.Stmt[i+1]) {
			p.newline()
		}
	}

	for _, com := range f.After {
		p.printf("%s", strings.TrimSpace(com.Token))
		p.newline()
	}
}

// compactStmt reports whether the pair of statements s1, s2
// should be printed without an intervening blank line.
// We omit the blank line when both are subinclude statements
// and the second one has no leading comments.
func compactStmt(s1, s2 Expr) bool {
	if len(s2.Comment().Before) > 0 {
		return false
	}

	return (isCall(s1, "subinclude") || isCall(s1, "load")) &&
		(isCall(s2, "subinclude") || isCall(s2, "load"))
}

// isCall reports whether x is a call to a function with the given name.
func isCall(x Expr, name string) bool {
	c, ok := x.(*CallExpr)
	if !ok {
		return false
	}
	nam, ok := c.X.(*LiteralExpr)
	if !ok {
		return false
	}
	return nam.Token == name
}

// Expression formatting.

// The expression formatter must introduce parentheses to force the
// meaning described by the parse tree. We preserve parentheses in the
// input, so extra parentheses are only needed if we have edited the tree.
//
// For example consider these expressions:
//	(1) "x" "y" % foo
//	(2) "x" + "y" % foo
//	(3) "x" + ("y" % foo)
//	(4) ("x" + "y") % foo
// When we parse (1), we represent the concatenation as an addition.
// However, if we print the addition back out without additional parens,
// as in (2), it has the same meaning as (3), which is not the original
// meaning. To preserve the original meaning we must add parens as in (4).
//
// To allow arbitrary rewrites to be formatted properly, we track full
// operator precedence while printing instead of just handling this one
// case of string concatenation.
//
// The precedences are assigned values low to high. A larger number
// binds tighter than a smaller number. All binary operators bind
// left-to-right.
const (
	precLow = iota
	precAssign
	precComma
	precColon
	precIn
	precOr
	precAnd
	precCmp
	precAdd
	precMultiply
	precSuffix
	precUnary
	precConcat
)

// opPrec gives the precedence for operators found in a BinaryExpr.
var opPrec = map[string]int{
	"=":   precAssign,
	"+=":  precAssign,
	"or":  precOr,
	"and": precAnd,
	"<":   precCmp,
	">":   precCmp,
	"==":  precCmp,
	"!=":  precCmp,
	"<=":  precCmp,
	">=":  precCmp,
	"+":   precAdd,
	"-":   precAdd,
	"*":   precMultiply,
	"/":   precMultiply,
	"%":   precMultiply,
}

// expr prints the expression v to the print buffer.
// The value outerPrec gives the precedence of the operator
// outside expr. If that operator binds tighter than v's operator,
// expr must introduce parentheses to preserve the meaning
// of the parse tree (see above).
func (p *printer) expr(v Expr, outerPrec int) {
	// Emit line-comments preceding this expression.
	// If we are in the middle of an expression but not inside ( ) [ ] { }
	// then we cannot just break the line: we'd have to end it with a \.
	// However, even then we can't emit line comments since that would
	// end the expression. This is only a concern if we have rewritten
	// the parse tree. If comments were okay before this expression in
	// the original input they're still okay now, in the absense of rewrites.
	//
	// TODO(bazel-team): Check whether it is valid to emit comments right now,
	// and if not, insert them earlier in the output instead, at the most
	// recent \n not following a \ line.
	if before := v.Comment().Before; len(before) > 0 {
		// Want to print a line comment.
		// Line comments must be at the current margin.
		p.trim()
		if p.indent() > 0 {
			// There's other text on the line. Start a new line.
			p.printf("\n")
		}
		// Re-indent to margin.
		p.printf("%*s", p.margin, "")
		for _, com := range before {
			p.printf("%s", strings.TrimSpace(com.Token))
			p.newline()
		}
	}

	// Do we introduce parentheses?
	// The result depends on the kind of expression.
	// Each expression type that might need parentheses
	// calls addParen with its own precedence.
	// If parentheses are necessary, addParen prints the
	// opening parenthesis and sets parenthesized so that
	// the code after the switch can print the closing one.
	parenthesized := false
	addParen := func(prec int) {
		if prec < outerPrec {
			p.printf("(")
			p.depth++
			parenthesized = true
		}
	}

	switch v := v.(type) {
	default:
		panic(fmt.Errorf("printer: unexpected type %T", v))

	case *LiteralExpr:
		p.printf("%s", v.Token)

	case *StringExpr:
		// If the Token is a correct quoting of Value, use it.
		// This preserves the specific escaping choices that
		// BUILD authors have made, and it also works around
		// b/7272572.
		if strings.HasPrefix(v.Token, `"`) {
			s, triple, err := unquote(v.Token)
			if s == v.Value && triple == v.TripleQuote && err == nil {
				p.printf("%s", v.Token)
				break
			}
		}

		p.printf("%s", quote(v.Value, v.TripleQuote))

	case *DotExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf(".%s", v.Name)

	case *IndexExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf("[")
		p.expr(v.Y, precLow)
		p.printf("]")

	case *KeyValueExpr:
		p.expr(v.Key, precLow)
		p.printf(": ")
		p.expr(v.Value, precLow)

	case *SliceExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf("[")
		if v.Y != nil {
			p.expr(v.Y, precLow)
		}
		p.printf(":")
		if v.Z != nil {
			p.expr(v.Z, precLow)
		}
		p.printf("]")

	case *UnaryExpr:
		addParen(precUnary)
		if v.Op == "not" {
			p.printf("not ") // Requires a space after it.
		} else {
			p.printf("%s", v.Op)
		}
		p.expr(v.X, precUnary)

	case *LambdaExpr:
		addParen(precColon)
		p.printf("lambda ")
		for i, name := range v.Var {
			if i > 0 {
				p.printf(", ")
			}
			p.expr(name, precLow)
		}
		p.printf(": ")
		p.expr(v.Expr, precColon)

	case *BinaryExpr:
		// Precedence: use the precedence of the operator.
		// Since all binary expressions format left-to-right,
		// it is okay for the left side to reuse the same operator
		// without parentheses, so we use prec for v.X.
		// For the same reason, the right side cannot reuse the same
		// operator, or else a parse tree for a + (b + c), where the ( ) are
		// not present in the source, will format as a + b + c, which
		// means (a + b) + c. Treat the right expression as appearing
		// in a context one precedence level higher: use prec+1 for v.Y.
		//
		// Line breaks: if we are to break the line immediately after
		// the operator, introduce a margin at the current column,
		// so that the second operand lines up with the first one and
		// also so that neither operand can use space to the left.
		// If the operator is an =, indent the right side another 4 spaces.
		prec := opPrec[v.Op]
		addParen(prec)
		m := p.margin
		if v.LineBreak {
			p.margin = p.indent()
			if v.Op == "=" {
				p.margin += 4
			}
		}

		p.expr(v.X, prec)
		p.printf(" %s", v.Op)
		if v.LineBreak {
			p.breakline()
		} else {
			p.printf(" ")
		}
		p.expr(v.Y, prec+1)
		p.margin = m

	case *ParenExpr:
		p.seq("()", []Expr{v.X}, &v.End, modeParen, false, v.ForceMultiLine)

	case *CallExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.seq("()", v.List, &v.End, modeCall, v.ForceCompact, v.ForceMultiLine)

	case *ListExpr:
		p.seq("[]", v.List, &v.End, modeList, false, v.ForceMultiLine)

	case *SetExpr:
		p.seq("{}", v.List, &v.End, modeList, false, v.ForceMultiLine)

	case *TupleExpr:
		p.seq("()", v.List, &v.End, modeTuple, v.ForceCompact, v.ForceMultiLine)

	case *DictExpr:
		var list []Expr
		for _, x := range v.List {
			list = append(list, x)
		}
		p.seq("{}", list, &v.End, modeDict, false, v.ForceMultiLine)

	case *ListForExpr:
		p.listFor(v)

	case *ConditionalExpr:
		addParen(precSuffix)
		p.expr(v.Then, precSuffix)
		p.printf(" if ")
		p.expr(v.Test, precSuffix)
		p.printf(" else ")
		p.expr(v.Else, precSuffix)
	}

	// Add closing parenthesis if needed.
	if parenthesized {
		p.depth--
		p.printf(")")
	}

	// Queue end-of-line comments for printing when we
	// reach the end of the line.
	p.comment = append(p.comment, v.Comment().Suffix...)
}

// A seqMode describes a formatting mode for a sequence of values,
// like a list or call arguments.
type seqMode int

const (
	_ seqMode = iota

	modeCall  // f(x)
	modeList  // [x]
	modeTuple // (x,)
	modeParen // (x)
	modeDict  // {x:y}
)

// seq formats a list of values inside a given bracket pair (brack = "()", "[]", "{}").
// The end node holds any trailing comments to be printed just before the
// closing bracket.
// The mode parameter specifies the sequence mode (see above).
// If multiLine is true, seq avoids the compact form even
// for 0- and 1-element sequences.
func (p *printer) seq(brack string, list []Expr, end *End, mode seqMode, forceCompact, forceMultiLine bool) {
	p.printf("%s", brack[:1])
	p.depth++

	// If there are line comments, force multiline
	// so we can print the comments before the closing bracket.
	for _, x := range list {
		if len(x.Comment().Before) > 0 {
			forceMultiLine = true
		}
	}
	if len(end.Before) > 0 {
		forceMultiLine = true
	}

	// Resolve possibly ambiguous call arguments explicitly
	// instead of depending on implicit resolution in logic below.
	if forceMultiLine {
		forceCompact = false
	}

	switch {
	case len(list) == 0 && !forceMultiLine:
		// Compact form: print nothing.

	case len(list) == 1 && !forceMultiLine:
		// Compact form.
		p.expr(list[0], precLow)
		// Tuple must end with comma, to mark it as a tuple.
		if mode == modeTuple {
			p.printf(",")
		}

	case forceCompact:
		// Compact form but multiple elements.
		for i, x := range list {
			if i > 0 {
				p.printf(", ")
			}
			p.expr(x, precLow)
		}

	default:
		// Multi-line form.
		p.margin += 4
		for i, x := range list {
			// If we are about to break the line before the first
			// element and there are trailing end-of-line comments
			// waiting to be printed, delay them and print them as
			// whole-line comments preceding that element.
			// Do this by printing a newline ourselves and positioning
			// so that the end-of-line comment, with the two spaces added,
			// will line up with the current margin.
			if i == 0 && len(p.comment) > 0 {
				p.printf("\n%*s", p.margin-2, "")
			}

			p.newline()
			p.expr(x, precLow)
			if mode != modeParen || i+1 < len(list) {
				p.printf(",")
			}
		}
		// Final comments.
		for _, com := range end.Before {
			p.newline()
			p.printf("%s", strings.TrimSpace(com.Token))
		}
		p.margin -= 4
		p.newline()
	}
	p.depth--
	p.printf("%s", brack[1:])
}

// listFor formats a ListForExpr (list comprehension).
// The single-line form is:
//	[x for y in z if c]
//
// and the multi-line form is:
//	[
//	    x
//	    for y in z
//	    if c
//	]
//
func (p *printer) listFor(v *ListForExpr) {
	multiLine := v.ForceMultiLine || len(v.End.Before) > 0

	// space breaks the line in multiline mode
	// or else prints a space.
	space := func() {
		if multiLine {
			p.breakline()
		} else {
			p.printf(" ")
		}
	}

	if v.Brack != "" {
		p.depth++
		p.printf("%s", v.Brack[:1])
	}

	if multiLine {
		if v.Brack != "" {
			p.margin += 4
		}
		p.newline()
	}

	p.expr(v.X, precLow)

	for _, c := range v.For {
		space()
		p.printf("for ")
		for i, name := range c.For.Var {
			if i > 0 {
				p.printf(", ")
			}
			p.expr(name, precLow)
		}
		p.printf(" in ")
		p.expr(c.For.Expr, precLow)
		p.comment = append(p.comment, c.For.Comment().Suffix...)

		for _, i := range c.Ifs {
			space()
			p.printf("if ")
			p.expr(i.Cond, precLow)
			p.comment = append(p.comment, i.Comment().Suffix...)
		}
		p.comment = append(p.comment, c.Comment().Suffix...)

	}

	if multiLine {
		for _, com := range v.End.Before {
			p.newline()
			p.printf("%s", strings.TrimSpace(com.Token))
		}
		if v.Brack != "" {
			p.margin -= 4
		}
		p.newline()
	}

	if v.Brack != "" {
		p.printf("%s", v.Brack[1:])
		p.depth--
	}
}
