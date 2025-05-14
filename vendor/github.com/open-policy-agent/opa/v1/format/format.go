// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package format implements formatting of Rego source files.
package format

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/open-policy-agent/opa/internal/future"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/types"
)

// Opts lets you control the code formatting via `AstWithOpts()`.
type Opts struct {
	// IgnoreLocations instructs the formatter not to use the AST nodes' locations
	// into account when laying out the code: notably, when the input is the result
	// of partial evaluation, arguments maybe have been shuffled around, but still
	// carry along their original source locations.
	IgnoreLocations bool

	// RegoVersion is the version of Rego to format code for.
	RegoVersion ast.RegoVersion

	// ParserOptions is the parser options used when parsing the module to be formatted.
	ParserOptions *ast.ParserOptions

	// DropV0Imports instructs the formatter to drop all v0 imports from the module; i.e. 'rego.v1' and 'future.keywords' imports.
	// Imports are only removed if [Opts.RegoVersion] makes them redundant.
	DropV0Imports bool
}

func (o Opts) effectiveRegoVersion() ast.RegoVersion {
	if o.RegoVersion == ast.RegoUndefined {
		return ast.DefaultRegoVersion
	}
	return o.RegoVersion
}

// defaultLocationFile is the file name used in `Ast()` for terms
// without a location, as could happen when pretty-printing the
// results of partial eval.
const defaultLocationFile = "__format_default__"

// Source formats a Rego source file. The bytes provided must describe a complete
// Rego module. If they don't, Source will return an error resulting from the attempt
// to parse the bytes.
func Source(filename string, src []byte) ([]byte, error) {
	return SourceWithOpts(filename, src, Opts{})
}

func SourceWithOpts(filename string, src []byte, opts Opts) ([]byte, error) {
	regoVersion := opts.effectiveRegoVersion()

	var parserOpts ast.ParserOptions
	if opts.ParserOptions != nil {
		parserOpts = *opts.ParserOptions
	} else if regoVersion == ast.RegoV1 {
		// If the rego version is V1, we need to parse it as such, to allow for future keywords not being imported.
		// Otherwise, we'll default to the default rego-version.
		parserOpts.RegoVersion = ast.RegoV1
	}

	if parserOpts.RegoVersion == ast.RegoUndefined {
		parserOpts.RegoVersion = ast.DefaultRegoVersion
	}

	module, err := ast.ParseModuleWithOpts(filename, string(src), parserOpts)
	if err != nil {
		return nil, err
	}

	if regoVersion == ast.RegoV0CompatV1 || regoVersion == ast.RegoV1 {
		checkOpts := ast.NewRegoCheckOptions()
		// The module is parsed as v0, so we need to disable checks that will be automatically amended by the AstWithOpts call anyways.
		checkOpts.RequireIfKeyword = false
		checkOpts.RequireContainsKeyword = false
		checkOpts.RequireRuleBodyOrValue = false
		errs := ast.CheckRegoV1WithOptions(module, checkOpts)
		if len(errs) > 0 {
			return nil, errs
		}
	}

	formatted, err := AstWithOpts(module, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", filename, err)
	}

	return formatted, nil
}

// MustAst is a helper function to format a Rego AST element. If any errors
// occur this function will panic. This is mostly used for test
func MustAst(x interface{}) []byte {
	bs, err := Ast(x)
	if err != nil {
		panic(err)
	}
	return bs
}

// MustAstWithOpts is a helper function to format a Rego AST element. If any errors
// occur this function will panic. This is mostly used for test
func MustAstWithOpts(x interface{}, opts Opts) []byte {
	bs, err := AstWithOpts(x, opts)
	if err != nil {
		panic(err)
	}
	return bs
}

// Ast formats a Rego AST element. If the passed value is not a valid AST
// element, Ast returns nil and an error. If AST nodes are missing locations
// an arbitrary location will be used.
func Ast(x interface{}) ([]byte, error) {
	return AstWithOpts(x, Opts{})
}

type fmtOpts struct {
	// When the future keyword "contains" is imported, all the pretty-printed
	// modules will use that format for partial sets.
	// NOTE(sr): For ref-head rules, this will be the default behaviour, since
	// we need "contains" to disambiguate complete rules from partial sets.
	contains bool

	// Same logic applies as for "contains": if `future.keywords.if` (or all
	// future keywords) is imported, we'll render rules that can use `if` with
	// `if`.
	ifs bool

	// We check all rule ref heads to see if any of them _requires_ support
	// for ref heads -- if they do, we'll print all of them in a different way
	// than if they don't.
	refHeads bool

	regoV1         bool
	regoV1Imported bool
	futureKeywords []string
}

func (o fmtOpts) keywords() []string {
	if o.regoV1 {
		return ast.KeywordsV1[:]
	}
	kws := ast.KeywordsV0[:]
	return append(kws, o.futureKeywords...)
}

func AstWithOpts(x interface{}, opts Opts) ([]byte, error) {
	// The node has to be deep copied because it may be mutated below. Alternatively,
	// we could avoid the copy by checking if mutation will occur first. For now,
	// since format is not latency sensitive, just deep copy in all cases.
	x = ast.Copy(x)

	wildcards := map[ast.Var]*ast.Term{}

	// NOTE(sr): When the formatter encounters a call to internal.member_2
	// or internal.member_3, it will sugarize them into usage of the `in`
	// operator. It has to ensure that the proper future keyword import is
	// present.
	extraFutureKeywordImports := map[string]struct{}{}

	o := fmtOpts{}

	regoVersion := opts.effectiveRegoVersion()
	if regoVersion == ast.RegoV0CompatV1 || regoVersion == ast.RegoV1 {
		o.regoV1 = true
		o.ifs = true
		o.contains = true
	}

	memberRef := ast.Member.Ref()
	memberWithKeyRef := ast.MemberWithKey.Ref()

	// Preprocess the AST. Set any required defaults and calculate
	// values required for printing the formatted output.
	ast.WalkNodes(x, func(x ast.Node) bool {
		switch n := x.(type) {
		case ast.Body:
			if len(n) == 0 {
				return false
			}
		case *ast.Term:
			unmangleWildcardVar(wildcards, n)

		case *ast.Expr:
			switch {
			case n.IsCall() && memberRef.Equal(n.Operator()) || memberWithKeyRef.Equal(n.Operator()):
				extraFutureKeywordImports["in"] = struct{}{}
			case n.IsEvery():
				extraFutureKeywordImports["every"] = struct{}{}
			}

		case *ast.Import:
			if kw, ok := future.WhichFutureKeyword(n); ok {
				o.futureKeywords = append(o.futureKeywords, kw)
			}

			switch {
			case isRegoV1Compatible(n):
				o.regoV1Imported = true
				o.contains = true
				o.ifs = true
			case future.IsAllFutureKeywords(n):
				o.contains = true
				o.ifs = true
			case future.IsFutureKeyword(n, "contains"):
				o.contains = true
			case future.IsFutureKeyword(n, "if"):
				o.ifs = true
			}

		case *ast.Rule:
			if len(n.Head.Ref()) > 2 {
				o.refHeads = true
			}
			if len(n.Head.Ref()) == 2 && n.Head.Key != nil && n.Head.Value == nil { // p.q contains "x"
				o.refHeads = true
			}
		}

		if opts.IgnoreLocations || x.Loc() == nil {
			x.SetLoc(defaultLocation(x))
		}
		return false
	})

	w := &writer{
		indent:  "\t",
		errs:    make([]*ast.Error, 0),
		fmtOpts: o,
	}

	switch x := x.(type) {
	case *ast.Module:
		if regoVersion == ast.RegoV1 && opts.DropV0Imports {
			x.Imports = filterRegoV1Import(x.Imports)
		} else if regoVersion == ast.RegoV0CompatV1 {
			x.Imports = ensureRegoV1Import(x.Imports)
		}

		regoV1Imported := moduleIsRegoV1Compatible(x)
		if regoVersion == ast.RegoV0CompatV1 || regoVersion == ast.RegoV1 || regoV1Imported {
			if !opts.DropV0Imports && !regoV1Imported {
				for _, kw := range o.futureKeywords {
					x.Imports = ensureFutureKeywordImport(x.Imports, kw)
				}
			} else {
				x.Imports = future.FilterFutureImports(x.Imports)
			}
		} else {
			for kw := range extraFutureKeywordImports {
				x.Imports = ensureFutureKeywordImport(x.Imports, kw)
			}
		}
		err := w.writeModule(x)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Package:
		_, err := w.writePackage(x, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Import:
		_, err := w.writeImports([]*ast.Import{x}, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Rule:
		_, err := w.writeRule(x, false /* isElse */, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Head:
		_, err := w.writeHead(x,
			false, // isDefault
			false, // isExpandedConst
			nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case ast.Body:
		_, err := w.writeBody(x, nil)
		if err != nil {
			return nil, err
		}
	case *ast.Expr:
		_, err := w.writeExpr(x, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.With:
		_, err := w.writeWith(x, nil, false)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Term:
		_, err := w.writeTerm(x, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case ast.Value:
		_, err := w.writeTerm(&ast.Term{Value: x, Location: &ast.Location{}}, nil)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	case *ast.Comment:
		err := w.writeComments([]*ast.Comment{x})
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
	default:
		return nil, fmt.Errorf("not an ast element: %v", x)
	}

	if len(w.errs) > 0 {
		return nil, w.errs
	}
	return squashTrailingNewlines(w.buf.Bytes()), nil
}

func unmangleWildcardVar(wildcards map[ast.Var]*ast.Term, n *ast.Term) {

	v, ok := n.Value.(ast.Var)
	if !ok || !v.IsWildcard() {
		return
	}

	first, ok := wildcards[v]
	if !ok {
		wildcards[v] = n
		return
	}

	w := v[len(ast.WildcardPrefix):]

	// Prepend an underscore to ensure the variable will parse.
	if len(w) == 0 || w[0] != '_' {
		w = "_" + w
	}

	if first != nil {
		first.Value = w
		wildcards[v] = nil
	}

	n.Value = w
}

func squashTrailingNewlines(bs []byte) []byte {
	if bytes.HasSuffix(bs, []byte("\n")) {
		return append(bytes.TrimRight(bs, "\n"), '\n')
	}
	return bs
}

func defaultLocation(x ast.Node) *ast.Location {
	return ast.NewLocation([]byte(x.String()), defaultLocationFile, 1, 1)
}

type writer struct {
	buf bytes.Buffer

	indent                  string
	level                   int
	inline                  bool
	beforeEnd               *ast.Comment
	delay                   bool
	errs                    ast.Errors
	fmtOpts                 fmtOpts
	writeCommentOnFinalLine bool
}

func (w *writer) writeModule(module *ast.Module) error {
	var pkg *ast.Package
	var others []interface{}
	var comments []*ast.Comment
	visitor := ast.NewGenericVisitor(func(x interface{}) bool {
		switch x := x.(type) {
		case *ast.Comment:
			comments = append(comments, x)
			return true
		case *ast.Import, *ast.Rule:
			others = append(others, x)
			return true
		case *ast.Package:
			pkg = x
			return true
		default:
			return false
		}
	})
	visitor.Walk(module)

	sort.Slice(comments, func(i, j int) bool {
		l, err := locLess(comments[i], comments[j])
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
		return l
	})

	sort.Slice(others, func(i, j int) bool {
		l, err := locLess(others[i], others[j])
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
		return l
	})

	comments = trimTrailingWhitespaceInComments(comments)

	var err error
	comments, err = w.writePackage(pkg, comments)
	if err != nil {
		return err
	}
	var imports []*ast.Import
	var rules []*ast.Rule
	for len(others) > 0 {
		imports, others = gatherImports(others)
		comments, err = w.writeImports(imports, comments)
		if err != nil {
			return err
		}
		rules, others = gatherRules(others)
		comments, err = w.writeRules(rules, comments)
		if err != nil {
			return err
		}
	}

	for i, c := range comments {
		w.writeLine(c.String())
		if i == len(comments)-1 {
			w.write("\n")
		}
	}

	return nil
}

func trimTrailingWhitespaceInComments(comments []*ast.Comment) []*ast.Comment {
	for _, c := range comments {
		c.Text = bytes.TrimRightFunc(c.Text, unicode.IsSpace)
	}

	return comments
}

func (w *writer) writePackage(pkg *ast.Package, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, pkg.Location)
	if err != nil {
		return nil, err
	}

	w.startLine()

	// Omit head as all packages have the DefaultRootDocument prepended at parse time.
	path := make(ast.Ref, len(pkg.Path)-1)
	if len(path) == 0 {
		w.errs = append(w.errs, ast.NewError(ast.FormatErr, pkg.Location, "invalid package path: %s", pkg.Path))
		return comments, nil
	}

	path[0] = ast.VarTerm(string(pkg.Path[1].Value.(ast.String)))
	copy(path[1:], pkg.Path[2:])

	w.write("package ")
	_, err = w.writeRef(path, nil)
	if err != nil {
		return nil, err
	}

	w.blankLine()

	return comments, nil
}

func (w *writer) writeComments(comments []*ast.Comment) error {
	for i := range comments {
		if i > 0 {
			l, err := locCmp(comments[i], comments[i-1])
			if err != nil {
				return err
			}
			if l > 1 {
				w.blankLine()
			}
		}

		w.writeLine(comments[i].String())
	}

	return nil
}

func (w *writer) writeRules(rules []*ast.Rule, comments []*ast.Comment) ([]*ast.Comment, error) {
	for i, rule := range rules {
		var err error
		comments, err = w.insertComments(comments, rule.Location)
		if err != nil && !errors.As(err, &unexpectedCommentError{}) {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}

		comments, err = w.writeRule(rule, false, comments)
		if err != nil && !errors.As(err, &unexpectedCommentError{}) {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}

		if i < len(rules)-1 && w.groupableOneLiner(rule) {
			next := rules[i+1]
			if w.groupableOneLiner(next) && next.Location.Row == rule.Location.Row+1 {
				// Current rule and the next are both groupable one-liners, and
				// adjacent in the original policy (i.e. no extra newlines between them).
				continue
			}
		}
		w.blankLine()
	}
	return comments, nil
}

var expandedConst = ast.NewBody(ast.NewExpr(ast.InternedBooleanTerm(true)))

func (w *writer) groupableOneLiner(rule *ast.Rule) bool {
	// Location required to determine if two rules are adjacent in the policy.
	// If not, we respect line breaks between rules.
	if len(rule.Body) > 1 || rule.Default || rule.Location == nil {
		return false
	}

	partialSetException := w.fmtOpts.contains || rule.Head.Value != nil

	return (w.fmtOpts.regoV1 || w.fmtOpts.ifs) && partialSetException
}

func (w *writer) writeRule(rule *ast.Rule, isElse bool, comments []*ast.Comment) ([]*ast.Comment, error) {
	if rule == nil {
		return comments, nil
	}

	if !isElse {
		w.startLine()
	}

	if rule.Default {
		w.write("default ")
	}

	// OPA transforms lone bodies like `foo = {"a": "b"}` into rules of the form
	// `foo = {"a": "b"} { true }` in the AST. We want to preserve that notation
	// in the formatted code instead of expanding the bodies into rules, so we
	// pretend that the rule has no body in this case.
	isExpandedConst := rule.Body.Equal(expandedConst) && rule.Else == nil
	w.writeCommentOnFinalLine = isExpandedConst

	var err error
	var unexpectedComment bool
	comments, err = w.writeHead(rule.Head, rule.Default, isExpandedConst, comments)
	if err != nil {
		if errors.As(err, &unexpectedCommentError{}) {
			unexpectedComment = true
		} else {
			return nil, err
		}
	}

	if len(rule.Body) == 0 || isExpandedConst {
		w.endLine()
		return comments, nil
	}

	w.writeCommentOnFinalLine = true

	// this excludes partial sets UNLESS `contains` is used
	partialSetException := w.fmtOpts.contains || rule.Head.Value != nil

	if (w.fmtOpts.regoV1 || w.fmtOpts.ifs) && partialSetException {
		w.write(" if")
		if len(rule.Body) == 1 {
			if rule.Body[0].Location.Row == rule.Head.Location.Row {
				w.write(" ")
				var err error
				comments, err = w.writeExpr(rule.Body[0], comments)
				if err != nil {
					return nil, err
				}
				w.endLine()
				if rule.Else != nil {
					comments, err = w.writeElse(rule, comments)
					if err != nil {
						return nil, err
					}
				}
				return comments, nil
			}
		}
	}
	if unexpectedComment && len(comments) > 0 {
		w.write(" { ")
	} else {
		w.write(" {")
		w.endLine()
	}

	w.up()

	comments, err = w.writeBody(rule.Body, comments)
	if err != nil {
		// the unexpected comment error is passed up to be handled by writeHead
		if !errors.As(err, &unexpectedCommentError{}) {
			return nil, err
		}
	}

	var closeLoc *ast.Location

	if len(rule.Head.Args) > 0 {
		closeLoc = closingLoc('(', ')', '{', '}', rule.Location)
	} else if rule.Head.Key != nil {
		closeLoc = closingLoc('[', ']', '{', '}', rule.Location)
	} else {
		closeLoc = closingLoc(0, 0, '{', '}', rule.Location)
	}

	comments, err = w.insertComments(comments, closeLoc)
	if err != nil {
		return nil, err
	}

	if err := w.down(); err != nil {
		return nil, err
	}
	w.startLine()
	w.write("}")
	if rule.Else != nil {
		comments, err = w.writeElse(rule, comments)
		if err != nil {
			return nil, err
		}
	}
	return comments, nil
}

var elseVar ast.Value = ast.Var("else")

func (w *writer) writeElse(rule *ast.Rule, comments []*ast.Comment) ([]*ast.Comment, error) {
	// If there was nothing else on the line before the "else" starts
	// then preserve this style of else block, otherwise it will be
	// started as an "inline" else eg:
	//
	//     p {
	//     	...
	//     }
	//
	//     else {
	//     	...
	//     }
	//
	// versus
	//
	//     p {
	// 	    ...
	//     } else {
	//     	...
	//     }
	//
	// Note: This doesn't use the `close` as it currently isn't accurate for all
	// types of values. Checking the actual line text is the most consistent approach.
	wasInline := false
	ruleLines := bytes.Split(rule.Location.Text, []byte("\n"))
	relativeElseRow := rule.Else.Location.Row - rule.Location.Row
	if relativeElseRow > 0 && relativeElseRow < len(ruleLines) {
		elseLine := ruleLines[relativeElseRow]
		if !bytes.HasPrefix(bytes.TrimSpace(elseLine), []byte("else")) {
			wasInline = true
		}
	}

	// If there are any comments between the closing brace of the previous rule and the start
	// of the else block we will always insert a new blank line between them.
	hasCommentAbove := len(comments) > 0 && comments[0].Location.Row-rule.Else.Head.Location.Row < 0 || w.beforeEnd != nil

	if !hasCommentAbove && wasInline {
		w.write(" ")
	} else {
		w.blankLine()
		w.startLine()
	}

	rule.Else.Head.Name = "else" // NOTE(sr): whaaat

	elseHeadReference := ast.NewTerm(elseVar)            // construct a reference for the term
	elseHeadReference.Location = rule.Else.Head.Location // and set the location to match the rule location

	rule.Else.Head.Reference = ast.Ref{elseHeadReference}
	rule.Else.Head.Args = nil
	var err error
	comments, err = w.insertComments(comments, rule.Else.Head.Location)
	if err != nil {
		return nil, err
	}

	if hasCommentAbove && !wasInline {
		// The comments would have ended the line, be sure to start one again
		// before writing the rest of the "else" rule.
		w.startLine()
	}

	// For backwards compatibility adjust the rule head value location
	// TODO: Refactor the logic for inserting comments, or special
	// case comments in a rule head value so this can be removed
	if rule.Else.Head.Value != nil {
		rule.Else.Head.Value.Location = rule.Else.Head.Location
	}

	return w.writeRule(rule.Else, true, comments)
}

func (w *writer) writeHead(head *ast.Head, isDefault bool, isExpandedConst bool, comments []*ast.Comment) ([]*ast.Comment, error) {
	ref := head.Ref()
	if head.Key != nil && head.Value == nil && !head.HasDynamicRef() {
		ref = ref.GroundPrefix()
	}
	if w.fmtOpts.refHeads || len(ref) == 1 {
		var err error
		comments, err = w.writeRef(ref, comments)
		if err != nil {
			return nil, err
		}
	} else {
		// if there are comments within the object in the rule head, don't format it
		if len(comments) > 0 && ref[1].Location.Row == comments[0].Location.Row {
			comments, err := w.writeUnformatted(head.Location, comments)
			if err != nil {
				return nil, err
			}
			return comments, nil
		}

		w.write(ref[0].String())
		w.write("[")
		w.write(ref[1].String())
		w.write("]")
	}

	if len(head.Args) > 0 {
		w.write("(")
		var args []interface{}
		for _, arg := range head.Args {
			args = append(args, arg)
		}
		var err error
		comments, err = w.writeIterable(args, head.Location, closingLoc(0, 0, '(', ')', head.Location), comments, w.listWriter())
		w.write(")")
		if err != nil {
			return comments, err
		}
	}
	if head.Key != nil {
		if w.fmtOpts.contains && head.Value == nil {
			w.write(" contains ")
			var err error
			comments, err = w.writeTerm(head.Key, comments)
			if err != nil {
				return comments, err
			}
		} else if head.Value == nil { // no `if` for p[x] notation
			w.write("[")
			var err error
			comments, err = w.writeTerm(head.Key, comments)
			if err != nil {
				return comments, err
			}
			w.write("]")
		}
	}

	if head.Value != nil &&
		(head.Key != nil || !ast.InternedBooleanTerm(true).Equal(head.Value) || isExpandedConst || isDefault) {

		// in rego v1, explicitly print value for ref-head constants that aren't partial set assignments, e.g.:
		// * a -> parser error, won't reach here
		// * a.b -> a contains "b"
		// * a.b.c -> a.b.c := true
		// * a.b.c.d -> a.b.c.d := true
		isRegoV1RefConst := w.fmtOpts.regoV1 && isExpandedConst && head.Key == nil && len(head.Args) == 0

		if head.Location == head.Value.Location &&
			head.Name != "else" &&
			ast.InternedBooleanTerm(true).Equal(head.Value) &&
			!isRegoV1RefConst {
			// If the value location is the same as the location of the head,
			// we know that the value is generated, i.e. f(1)
			// Don't print the value (` = true`) as it is implied.
			return comments, nil
		}

		if head.Assign || w.fmtOpts.regoV1 {
			// preserve assignment operator, and enforce it if formatting for Rego v1
			w.write(" := ")
		} else {
			w.write(" = ")
		}
		var err error
		comments, err = w.writeTerm(head.Value, comments)
		if err != nil {
			return comments, err
		}
	}
	return comments, nil
}

func (w *writer) insertComments(comments []*ast.Comment, loc *ast.Location) ([]*ast.Comment, error) {
	before, at, comments := partitionComments(comments, loc)

	err := w.writeComments(before)
	if err != nil {
		return nil, err
	}
	if len(before) > 0 && loc.Row-before[len(before)-1].Location.Row > 1 {
		w.blankLine()
	}

	return comments, w.beforeLineEnd(at)
}

func (w *writer) writeBody(body ast.Body, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, body.Loc())
	if err != nil {
		return comments, err
	}
	for i, expr := range body {
		// Insert a blank line in before the expression if it was not right
		// after the previous expression.
		if i > 0 {
			lastRow := body[i-1].Location.Row
			for _, c := range body[i-1].Location.Text {
				if c == '\n' {
					lastRow++
				}
			}
			if expr.Location.Row > lastRow+1 {
				w.blankLine()
			}
		}
		w.startLine()

		comments, err = w.writeExpr(expr, comments)
		if err != nil && !errors.As(err, &unexpectedCommentError{}) {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
		w.endLine()
	}
	return comments, nil
}

func (w *writer) writeExpr(expr *ast.Expr, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, expr.Location)
	if err != nil {
		return comments, err
	}
	if !w.inline {
		w.startLine()
	}

	if expr.Negated {
		w.write("not ")
	}

	switch t := expr.Terms.(type) {
	case *ast.SomeDecl:
		comments, err = w.writeSomeDecl(t, comments)
		if err != nil {
			return nil, err
		}
	case *ast.Every:
		comments, err = w.writeEvery(t, comments)
		if err != nil {
			return nil, err
		}
	case []*ast.Term:
		comments, err = w.writeFunctionCall(expr, comments)
		if err != nil {
			return comments, err
		}
	case *ast.Term:
		comments, err = w.writeTerm(t, comments)
		if err != nil {
			return comments, err
		}
	}

	var indented, down bool
	for i, with := range expr.With {
		if i == 0 || with.Location.Row == expr.With[i-1].Location.Row { // we're on the same line
			comments, err = w.writeWith(with, comments, false)
			if err != nil {
				return nil, err
			}
		} else { // we're on a new line
			if !indented {
				indented = true

				w.up()
				down = true
			}
			w.endLine()
			w.startLine()
			comments, err = w.writeWith(with, comments, true)
			if err != nil {
				return nil, err
			}
		}
	}

	if down {
		if err := w.down(); err != nil {
			return nil, err
		}
	}

	return comments, nil
}

func (w *writer) writeSomeDecl(decl *ast.SomeDecl, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, decl.Location)
	if err != nil {
		return nil, err
	}
	w.write("some ")

	row := decl.Location.Row

	for i, term := range decl.Symbols {
		switch val := term.Value.(type) {
		case ast.Var:
			if term.Location.Row > row {
				w.endLine()
				w.startLine()
				w.write(w.indent)
				row = term.Location.Row
			} else if i > 0 {
				w.write(" ")
			}

			comments, err = w.writeTerm(term, comments)
			if err != nil {
				return nil, err
			}

			if i < len(decl.Symbols)-1 {
				w.write(",")
			}
		case ast.Call:
			comments, err = w.writeInOperator(false, val[1:], comments, decl.Location, ast.BuiltinMap[val[0].String()].Decl)
			if err != nil {
				return nil, err
			}
		}
	}

	return comments, nil
}

func (w *writer) writeEvery(every *ast.Every, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, every.Location)
	if err != nil {
		return nil, err
	}
	w.write("every ")
	if every.Key != nil {
		comments, err = w.writeTerm(every.Key, comments)
		if err != nil {
			return nil, err
		}
		w.write(", ")
	}
	comments, err = w.writeTerm(every.Value, comments)
	if err != nil {
		return nil, err
	}
	w.write(" in ")
	comments, err = w.writeTerm(every.Domain, comments)
	if err != nil {
		return nil, err
	}
	w.write(" {")
	comments, err = w.writeComprehensionBody('{', '}', every.Body, every.Loc(), every.Loc(), comments)
	if err != nil {
		// the unexpected comment error is passed up to be handled by writeHead
		if !errors.As(err, &unexpectedCommentError{}) {
			return nil, err
		}
	}

	if len(every.Body) == 1 &&
		every.Body[0].Location.Row == every.Location.Row {
		w.write(" ")
	}
	w.write("}")
	return comments, nil
}

func (w *writer) writeFunctionCall(expr *ast.Expr, comments []*ast.Comment) ([]*ast.Comment, error) {

	terms := expr.Terms.([]*ast.Term)
	operator := terms[0].Value.String()

	switch operator {
	case ast.Member.Name, ast.MemberWithKey.Name:
		return w.writeInOperator(false, terms[1:], comments, terms[0].Location, ast.BuiltinMap[terms[0].String()].Decl)
	}

	bi, ok := ast.BuiltinMap[operator]
	if !ok || bi.Infix == "" {
		return w.writeFunctionCallPlain(terms, comments)
	}

	numDeclArgs := bi.Decl.Arity()
	numCallArgs := len(terms) - 1

	var err error
	switch numCallArgs {
	case numDeclArgs: // Print infix where result is unassigned (e.g., x != y)
		comments, err = w.writeTerm(terms[1], comments)
		if err != nil {
			return nil, err
		}
		w.write(" " + bi.Infix + " ")
		return w.writeTerm(terms[2], comments)
	case numDeclArgs + 1: // Print infix where result is assigned (e.g., z = x + y)
		comments, err = w.writeTerm(terms[3], comments)
		if err != nil {
			return nil, err
		}
		w.write(" " + ast.Equality.Infix + " ")
		comments, err = w.writeTerm(terms[1], comments)
		if err != nil {
			return nil, err
		}
		w.write(" " + bi.Infix + " ")
		comments, err = w.writeTerm(terms[2], comments)
		if err != nil {
			return nil, err
		}
		return comments, nil
	}
	// NOTE(Trolloldem): in this point we are operating with a built-in function with the
	// wrong arity even when the assignment notation is used
	w.errs = append(w.errs, ArityFormatMismatchError(terms[1:], terms[0].String(), terms[0].Location, bi.Decl))
	return w.writeFunctionCallPlain(terms, comments)
}

func (w *writer) writeFunctionCallPlain(terms []*ast.Term, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write(terms[0].String() + "(")
	defer w.write(")")
	args := make([]interface{}, len(terms)-1)
	for i, t := range terms[1:] {
		args[i] = t
	}
	loc := terms[0].Location
	var err error
	comments, err = w.writeIterable(args, loc, closingLoc(0, 0, '(', ')', loc), comments, w.listWriter())
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (w *writer) writeWith(with *ast.With, comments []*ast.Comment, indented bool) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, with.Location)
	if err != nil {
		return nil, err
	}
	if !indented {
		w.write(" ")
	}
	w.write("with ")
	comments, err = w.writeTerm(with.Target, comments)
	if err != nil {
		return nil, err
	}
	w.write(" as ")
	comments, err = w.writeTerm(with.Value, comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (w *writer) writeTerm(term *ast.Term, comments []*ast.Comment) ([]*ast.Comment, error) {
	currentComments := make([]*ast.Comment, len(comments))
	copy(currentComments, comments)

	currentLen := w.buf.Len()

	comments, err := w.writeTermParens(false, term, comments)
	if err != nil {
		if errors.As(err, &unexpectedCommentError{}) {
			w.buf.Truncate(currentLen)

			comments, uErr := w.writeUnformatted(term.Location, currentComments)
			if uErr != nil {
				return nil, uErr
			}
			return comments, err
		}
		return nil, err
	}

	return comments, nil
}

// writeUnformatted writes the unformatted text instead and updates the comment state
func (w *writer) writeUnformatted(location *ast.Location, currentComments []*ast.Comment) ([]*ast.Comment, error) {
	if len(location.Text) == 0 {
		return nil, errors.New("original unformatted text is empty")
	}

	rawRule := string(location.Text)
	rowNum := len(strings.Split(rawRule, "\n"))

	w.write(string(location.Text))

	comments := make([]*ast.Comment, 0, len(currentComments))
	for _, c := range currentComments {
		// if there is a body then wait to write the last comment
		if w.writeCommentOnFinalLine && c.Location.Row == location.Row+rowNum-1 {
			w.write(" " + string(c.Location.Text))
			continue
		}

		// drop comments that occur within the rule raw text
		if c.Location.Row < location.Row+rowNum-1 {
			continue
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (w *writer) writeTermParens(parens bool, term *ast.Term, comments []*ast.Comment) ([]*ast.Comment, error) {
	var err error
	comments, err = w.insertComments(comments, term.Location)
	if err != nil {
		return nil, err
	}
	if !w.inline {
		w.startLine()
	}

	switch x := term.Value.(type) {
	case ast.Ref:
		comments, err = w.writeRef(x, comments)
		if err != nil {
			return nil, err
		}
	case ast.Object:
		comments, err = w.writeObject(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case *ast.Array:
		comments, err = w.writeArray(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case ast.Set:
		comments, err = w.writeSet(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case *ast.ArrayComprehension:
		comments, err = w.writeArrayComprehension(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case *ast.ObjectComprehension:
		comments, err = w.writeObjectComprehension(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case *ast.SetComprehension:
		comments, err = w.writeSetComprehension(x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case ast.String:
		if term.Location.Text[0] == '`' {
			// To preserve raw strings, we need to output the original text,
			w.write(string(term.Location.Text))
		} else {
			// x.String() cannot be used by default because it can change the input string "\u0000" to "\x00"
			var after, quote string
			var found bool
			// term.Location.Text could contain the prefix `else :=`, remove it
			switch term.Location.Text[len(term.Location.Text)-1] {
			case '"':
				quote = "\""
				_, after, found = strings.Cut(string(term.Location.Text), quote)
			case '`':
				quote = "`"
				_, after, found = strings.Cut(string(term.Location.Text), quote)
			}

			if !found {
				// If no quoted string was found, that means it is a key being formatted to a string
				// e.g. partial_set.y to partial_set["y"]
				w.write(x.String())
			} else {
				w.write(quote + after)
			}

		}
	case ast.Var:
		w.write(w.formatVar(x))
	case ast.Call:
		comments, err = w.writeCall(parens, x, term.Location, comments)
		if err != nil {
			return nil, err
		}
	case fmt.Stringer:
		w.write(x.String())
	}

	if !w.inline {
		w.startLine()
	}
	return comments, nil
}

func (w *writer) writeRef(x ast.Ref, comments []*ast.Comment) ([]*ast.Comment, error) {
	if len(x) > 0 {
		parens := false
		_, ok := x[0].Value.(ast.Call)
		if ok {
			parens = x[0].Location.Text[0] == 40 // Starts with "("
		}
		var err error
		comments, err = w.writeTermParens(parens, x[0], comments)
		if err != nil {
			return nil, err
		}
		path := x[1:]
		for _, t := range path {
			switch p := t.Value.(type) {
			case ast.String:
				w.writeRefStringPath(p)
			case ast.Var:
				w.writeBracketed(w.formatVar(p))
			default:
				w.write("[")
				comments, err = w.writeTerm(t, comments)
				if err != nil {
					if errors.As(err, &unexpectedCommentError{}) {
						// add a new line so that the closing bracket isn't part of the unexpected comment
						w.write("\n")
					} else {
						return nil, err
					}
				}
				w.write("]")
			}
		}
	}

	return comments, nil
}

func (w *writer) writeBracketed(str string) {
	w.write("[" + str + "]")
}

var varRegexp = regexp.MustCompile("^[[:alpha:]_][[:alpha:][:digit:]_]*$")

func (w *writer) writeRefStringPath(s ast.String) {
	str := string(s)
	if varRegexp.MatchString(str) && !ast.IsInKeywords(str, w.fmtOpts.keywords()) {
		w.write("." + str)
	} else {
		w.writeBracketed(s.String())
	}
}

func (*writer) formatVar(v ast.Var) string {
	if v.IsWildcard() {
		return ast.Wildcard.String()
	}
	return v.String()
}

func (w *writer) writeCall(parens bool, x ast.Call, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	bi, ok := ast.BuiltinMap[x[0].String()]
	if !ok || bi.Infix == "" {
		return w.writeFunctionCallPlain(x, comments)
	}

	if bi.Infix == "in" {
		// NOTE(sr): `in` requires special handling, mirroring what happens in the parser,
		// since there can be one or two lhs arguments.
		return w.writeInOperator(true, x[1:], comments, loc, bi.Decl)
	}

	// TODO(tsandall): improve to consider precedence?
	if parens {
		w.write("(")
	}

	// NOTE(Trolloldem): writeCall is only invoked when the function call is a term
	// of another function. The only valid arity is the one of the
	// built-in function
	if bi.Decl.Arity() != len(x)-1 {
		w.errs = append(w.errs, ArityFormatMismatchError(x[1:], x[0].String(), loc, bi.Decl))
		return comments, nil
	}

	var err error
	comments, err = w.writeTermParens(true, x[1], comments)
	if err != nil {
		return nil, err
	}
	w.write(" " + bi.Infix + " ")
	comments, err = w.writeTermParens(true, x[2], comments)
	if err != nil {
		return nil, err
	}
	if parens {
		w.write(")")
	}

	return comments, nil
}

func (w *writer) writeInOperator(parens bool, operands []*ast.Term, comments []*ast.Comment, loc *ast.Location, f *types.Function) ([]*ast.Comment, error) {

	if len(operands) != f.Arity() {
		// The number of operands does not math the arity of the `in` operator
		operator := ast.Member.Name
		if f.Arity() == 3 {
			operator = ast.MemberWithKey.Name
		}
		w.errs = append(w.errs, ArityFormatMismatchError(operands, operator, loc, f))
		return comments, nil
	}
	kw := "in"
	var err error
	switch len(operands) {
	case 2:
		comments, err = w.writeTermParens(true, operands[0], comments)
		if err != nil {
			return nil, err
		}
		w.write(" ")
		w.write(kw)
		w.write(" ")
		comments, err = w.writeTermParens(true, operands[1], comments)
		if err != nil {
			return nil, err
		}
	case 3:
		if parens {
			w.write("(")
			defer w.write(")")
		}
		comments, err = w.writeTermParens(true, operands[0], comments)
		if err != nil {
			return nil, err
		}
		w.write(", ")
		comments, err = w.writeTermParens(true, operands[1], comments)
		if err != nil {
			return nil, err
		}
		w.write(" ")
		w.write(kw)
		w.write(" ")
		comments, err = w.writeTermParens(true, operands[2], comments)
		if err != nil {
			return nil, err
		}
	}
	return comments, nil
}

func (w *writer) writeObject(obj ast.Object, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write("{")
	defer w.write("}")

	var s []interface{}
	obj.Foreach(func(k, v *ast.Term) {
		s = append(s, ast.Item(k, v))
	})
	return w.writeIterable(s, loc, closingLoc(0, 0, '{', '}', loc), comments, w.objectWriter())
}

func (w *writer) writeArray(arr *ast.Array, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write("[")
	defer w.write("]")

	var s []interface{}
	arr.Foreach(func(t *ast.Term) {
		s = append(s, t)
	})
	var err error
	comments, err = w.writeIterable(s, loc, closingLoc(0, 0, '[', ']', loc), comments, w.listWriter())
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (w *writer) writeSet(set ast.Set, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {

	if set.Len() == 0 {
		w.write("set()")
		var err error
		comments, err = w.insertComments(comments, closingLoc(0, 0, '(', ')', loc))
		if err != nil {
			return nil, err
		}
		return comments, nil
	}

	w.write("{")
	defer w.write("}")

	var s []interface{}
	set.Foreach(func(t *ast.Term) {
		s = append(s, t)
	})
	var err error
	comments, err = w.writeIterable(s, loc, closingLoc(0, 0, '{', '}', loc), comments, w.listWriter())
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (w *writer) writeArrayComprehension(arr *ast.ArrayComprehension, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write("[")
	defer w.write("]")

	return w.writeComprehension('[', ']', arr.Term, arr.Body, loc, comments)
}

func (w *writer) writeSetComprehension(set *ast.SetComprehension, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write("{")
	defer w.write("}")

	return w.writeComprehension('{', '}', set.Term, set.Body, loc, comments)
}

func (w *writer) writeObjectComprehension(object *ast.ObjectComprehension, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	w.write("{")
	defer w.write("}")

	object.Value.Location = object.Key.Location // Ensure the value is not written on the next line.
	if object.Key.Location.Row-loc.Row > 1 {
		w.endLine()
		w.startLine()
	}

	var err error
	comments, err = w.writeTerm(object.Key, comments)
	if err != nil {
		return nil, err
	}
	w.write(": ")
	return w.writeComprehension('{', '}', object.Value, object.Body, loc, comments)
}

func (w *writer) writeComprehension(openChar, closeChar byte, term *ast.Term, body ast.Body, loc *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	if term.Location.Row-loc.Row >= 1 {
		w.endLine()
		w.startLine()
	}

	parens := false
	_, ok := term.Value.(ast.Call)
	if ok {
		parens = term.Location.Text[0] == 40 // Starts with "("
	}
	var err error
	comments, err = w.writeTermParens(parens, term, comments)
	if err != nil {
		return nil, err
	}
	w.write(" |")

	return w.writeComprehensionBody(openChar, closeChar, body, term.Location, loc, comments)
}

func (w *writer) writeComprehensionBody(openChar, closeChar byte, body ast.Body, term, compr *ast.Location, comments []*ast.Comment) ([]*ast.Comment, error) {
	exprs := make([]interface{}, 0, len(body))
	for _, expr := range body {
		exprs = append(exprs, expr)
	}
	lines, err := w.groupIterable(exprs, term)
	if err != nil {
		return nil, err
	}

	if body.Loc().Row-term.Row > 0 || len(lines) > 1 {
		w.endLine()
		w.up()
		defer w.startLine()
		defer func() {
			if err := w.down(); err != nil {
				w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
			}
		}()

		var err error
		comments, err = w.writeBody(body, comments)
		if err != nil {
			return comments, err
		}
	} else {
		w.write(" ")
		i := 0
		for ; i < len(body)-1; i++ {
			comments, err = w.writeExpr(body[i], comments)
			if err != nil {
				return comments, err
			}
			w.write("; ")
		}
		comments, err = w.writeExpr(body[i], comments)
		if err != nil {
			return comments, err
		}
	}
	comments, err = w.insertComments(comments, closingLoc(0, 0, openChar, closeChar, compr))
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (w *writer) writeImports(imports []*ast.Import, comments []*ast.Comment) ([]*ast.Comment, error) {
	m, comments := mapImportsToComments(imports, comments)

	groups := groupImports(imports)
	for _, group := range groups {
		var err error
		comments, err = w.insertComments(comments, group[0].Loc())
		if err != nil {
			return nil, err
		}

		// Sort imports within a newline grouping.
		slices.SortFunc(group, (*ast.Import).Compare)
		for _, i := range group {
			w.startLine()
			err = w.writeImport(i)
			if err != nil {
				return nil, err
			}
			if c, ok := m[i]; ok {
				w.write(" " + c.String())
			}
			w.endLine()
		}
		w.blankLine()
	}

	return comments, nil
}

func (w *writer) writeImport(imp *ast.Import) error {
	path := imp.Path.Value.(ast.Ref)

	buf := []string{"import"}

	if _, ok := future.WhichFutureKeyword(imp); ok {
		// We don't want to wrap future.keywords imports in parens, so we create a new writer that doesn't
		w2 := writer{
			buf: bytes.Buffer{},
		}
		_, err := w2.writeRef(path, nil)
		if err != nil {
			return err
		}
		buf = append(buf, w2.buf.String())
	} else {
		buf = append(buf, path.String())
	}

	if len(imp.Alias) > 0 {
		buf = append(buf, "as "+imp.Alias.String())
	}
	w.write(strings.Join(buf, " "))

	return nil
}

type entryWriter func(interface{}, []*ast.Comment) ([]*ast.Comment, error)

func (w *writer) writeIterable(elements []interface{}, last *ast.Location, close *ast.Location, comments []*ast.Comment, fn entryWriter) ([]*ast.Comment, error) {
	lines, err := w.groupIterable(elements, last)
	if err != nil {
		return nil, err
	}
	if len(lines) > 1 {
		w.delayBeforeEnd()
		w.startMultilineSeq()
	}

	i := 0
	for ; i < len(lines)-1; i++ {
		comments, err = w.writeIterableLine(lines[i], comments, fn)
		if err != nil {
			return nil, err
		}
		w.write(",")

		w.endLine()
		w.startLine()
	}

	comments, err = w.writeIterableLine(lines[i], comments, fn)
	if err != nil {
		return nil, err
	}

	if len(lines) > 1 {
		w.write(",")
		w.endLine()
		comments, err = w.insertComments(comments, close)
		if err != nil {
			return nil, err
		}
		if err := w.down(); err != nil {
			return nil, err
		}
		w.startLine()
	}

	return comments, nil
}

func (w *writer) writeIterableLine(elements []interface{}, comments []*ast.Comment, fn entryWriter) ([]*ast.Comment, error) {
	if len(elements) == 0 {
		return comments, nil
	}

	i := 0
	for ; i < len(elements)-1; i++ {
		var err error
		comments, err = fn(elements[i], comments)
		if err != nil {
			return nil, err
		}
		w.write(", ")
	}

	return fn(elements[i], comments)
}

func (w *writer) objectWriter() entryWriter {
	return func(x interface{}, comments []*ast.Comment) ([]*ast.Comment, error) {
		entry := x.([2]*ast.Term)

		call, isCall := entry[0].Value.(ast.Call)

		paren := false
		if isCall && ast.Or.Ref().Equal(call[0].Value) && entry[0].Location.Text[0] == 40 { // Starts with "("
			paren = true
			w.write("(")
		}

		var err error
		comments, err = w.writeTerm(entry[0], comments)
		if err != nil {
			return nil, err
		}
		if paren {
			w.write(")")
		}

		w.write(": ")

		call, isCall = entry[1].Value.(ast.Call)
		if isCall && ast.Or.Ref().Equal(call[0].Value) && entry[1].Location.Text[0] == 40 { // Starts with "("
			w.write("(")
			defer w.write(")")
		}

		return w.writeTerm(entry[1], comments)
	}
}

func (w *writer) listWriter() entryWriter {
	return func(x interface{}, comments []*ast.Comment) ([]*ast.Comment, error) {
		t, ok := x.(*ast.Term)
		if ok {
			call, isCall := t.Value.(ast.Call)
			if isCall && ast.Or.Ref().Equal(call[0].Value) && t.Location.Text[0] == 40 { // Starts with "("
				w.write("(")
				defer w.write(")")
			}
		}

		return w.writeTerm(t, comments)
	}
}

// groupIterable will group the `elements` slice into slices according to their
// location: anything on the same line will be put into a slice.
func (w *writer) groupIterable(elements []interface{}, last *ast.Location) ([][]interface{}, error) {
	// Generated vars occur in the AST when we're rendering the result of
	// partial evaluation in a bundle build with optimization.
	// Those variables, and wildcard variables have the "default location",
	// set in `Ast()`). That is no proper file location, and the grouping
	// based on source location will yield a bad result.
	// Another case is generated variables: they do have proper file locations,
	// but their row/col information may no longer match their AST location.
	// So, for generated variables, we also don't trust the location, but
	// keep them ungrouped.
	def := false // default location found?
	for _, elem := range elements {
		ast.WalkTerms(elem, func(t *ast.Term) bool {
			if t.Location.File == defaultLocationFile {
				def = true
				return true
			}
			return false
		})
		ast.WalkVars(elem, func(v ast.Var) bool {
			if v.IsGenerated() {
				def = true
				return true
			}
			return false
		})
		if def { // return as-is
			return [][]interface{}{elements}, nil
		}
	}

	slices.SortFunc(elements, func(i, j any) int {
		l, err := locCmp(i, j)
		if err != nil {
			w.errs = append(w.errs, ast.NewError(ast.FormatErr, &ast.Location{}, err.Error()))
		}
		return l
	})

	var lines [][]interface{}
	cur := make([]interface{}, 0, len(elements))
	for i, t := range elements {
		elem := t
		loc, err := getLoc(elem)
		if err != nil {
			return nil, err
		}
		lineDiff := loc.Row - last.Row
		if lineDiff > 0 && i > 0 {
			lines = append(lines, cur)
			cur = nil
		}

		last = loc
		cur = append(cur, elem)
	}
	return append(lines, cur), nil
}

func mapImportsToComments(imports []*ast.Import, comments []*ast.Comment) (map[*ast.Import]*ast.Comment, []*ast.Comment) {
	var leftovers []*ast.Comment
	m := map[*ast.Import]*ast.Comment{}

	for _, c := range comments {
		matched := false
		for _, i := range imports {
			if c.Loc().Row == i.Loc().Row {
				m[i] = c
				matched = true
				break
			}
		}
		if !matched {
			leftovers = append(leftovers, c)
		}
	}

	return m, leftovers
}

func groupImports(imports []*ast.Import) [][]*ast.Import {
	switch len(imports) { // shortcuts
	case 0:
		return nil
	case 1:
		return [][]*ast.Import{imports}
	}
	// there are >=2 imports to group

	var groups [][]*ast.Import
	group := []*ast.Import{imports[0]}

	for _, i := range imports[1:] {
		last := group[len(group)-1]

		// nil-location imports have been sorted up to come first
		if i.Loc() != nil && last.Loc() != nil && // first import with a location, or
			i.Loc().Row-last.Loc().Row > 1 { // more than one row apart from previous import

			// start a new group
			groups = append(groups, group)
			group = []*ast.Import{}
		}
		group = append(group, i)
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}

	return groups
}

func partitionComments(comments []*ast.Comment, l *ast.Location) ([]*ast.Comment, *ast.Comment, []*ast.Comment) {
	if len(comments) == 0 {
		return nil, nil, nil
	}

	numBefore, numAfter := 0, 0
	for _, c := range comments {
		switch cmp := c.Location.Row - l.Row; {
		case cmp < 0:
			numBefore++
		case cmp > 0:
			numAfter++
		}
	}

	if numAfter == len(comments) {
		return nil, nil, comments
	}

	var at *ast.Comment

	before := make([]*ast.Comment, 0, numBefore)
	after := comments[0 : 0 : len(comments)-numBefore]

	for _, c := range comments {
		switch cmp := c.Location.Row - l.Row; {
		case cmp < 0:
			before = append(before, c)
		case cmp > 0:
			after = append(after, c)
		default:
			at = c
		}
	}

	return before, at, after
}

func gatherImports(others []interface{}) (imports []*ast.Import, rest []interface{}) {
	i := 0
loop:
	for ; i < len(others); i++ {
		switch x := others[i].(type) {
		case *ast.Import:
			imports = append(imports, x)
		case *ast.Rule:
			break loop
		}
	}
	return imports, others[i:]
}

func gatherRules(others []interface{}) (rules []*ast.Rule, rest []interface{}) {
	i := 0
loop:
	for ; i < len(others); i++ {
		switch x := others[i].(type) {
		case *ast.Rule:
			rules = append(rules, x)
		case *ast.Import:
			break loop
		}
	}
	return rules, others[i:]
}

func locLess(a, b interface{}) (bool, error) {
	c, err := locCmp(a, b)
	return c < 0, err
}

func locCmp(a, b interface{}) (int, error) {
	al, err := getLoc(a)
	if err != nil {
		return 0, err
	}
	bl, err := getLoc(b)
	if err != nil {
		return 0, err
	}
	switch {
	case al == nil && bl == nil:
		return 0, nil
	case al == nil:
		return -1, nil
	case bl == nil:
		return 1, nil
	}

	if cmp := al.Row - bl.Row; cmp != 0 {
		return cmp, nil

	}
	return al.Col - bl.Col, nil
}

func getLoc(x interface{}) (*ast.Location, error) {
	switch x := x.(type) {
	case ast.Node: // *ast.Head, *ast.Expr, *ast.With, *ast.Term
		return x.Loc(), nil
	case *ast.Location:
		return x, nil
	case [2]*ast.Term: // Special case to allow for easy printing of objects.
		return x[0].Location, nil
	default:
		return nil, fmt.Errorf("unable to get location for type %v", x)
	}
}

var negativeRow = &ast.Location{Row: -1}

func closingLoc(skipOpen, skipClose, openChar, closeChar byte, loc *ast.Location) *ast.Location {
	i, offset := 0, 0

	// Skip past parens/brackets/braces in rule heads.
	if skipOpen > 0 {
		i, offset = skipPast(skipOpen, skipClose, loc)
	}

	for ; i < len(loc.Text); i++ {
		if loc.Text[i] == openChar {
			break
		}
	}

	if i >= len(loc.Text) {
		return negativeRow
	}

	state := 1
	for state > 0 {
		i++
		if i >= len(loc.Text) {
			return negativeRow
		}

		switch loc.Text[i] {
		case openChar:
			state++
		case closeChar:
			state--
		case '\n':
			offset++
		}
	}

	return &ast.Location{Row: loc.Row + offset}
}

func skipPast(openChar, closeChar byte, loc *ast.Location) (int, int) {
	i := 0
	for ; i < len(loc.Text); i++ {
		if loc.Text[i] == openChar {
			break
		}
	}

	state := 1
	offset := 0
	for state > 0 {
		i++
		if i >= len(loc.Text) {
			return i, offset
		}

		switch loc.Text[i] {
		case openChar:
			state++
		case closeChar:
			state--
		case '\n':
			offset++
		}
	}

	return i, offset
}

// startLine begins a line with the current indentation level.
func (w *writer) startLine() {
	w.inline = true
	for range w.level {
		w.write(w.indent)
	}
}

// endLine ends a line with a newline.
func (w *writer) endLine() {
	w.inline = false
	if w.beforeEnd != nil && !w.delay {
		w.write(" " + w.beforeEnd.String())
		w.beforeEnd = nil
	}
	w.delay = false
	w.write("\n")
}

type unexpectedCommentError struct {
	newComment         string
	newCommentRow      int
	existingComment    string
	existingCommentRow int
}

func (u unexpectedCommentError) Error() string {
	return fmt.Sprintf("unexpected new comment (%s) on line %d because there is already a comment (%s) registered for line %d",
		u.newComment, u.newCommentRow, u.existingComment, u.existingCommentRow)
}

// beforeLineEnd registers a comment to be printed at the end of the current line.
func (w *writer) beforeLineEnd(c *ast.Comment) error {
	if w.beforeEnd != nil {
		if c == nil {
			return nil
		}

		existingComment := truncatedString(w.beforeEnd.String(), 100)
		existingCommentRow := w.beforeEnd.Location.Row
		newComment := truncatedString(c.String(), 100)
		w.beforeEnd = nil

		return unexpectedCommentError{
			newComment:         newComment,
			newCommentRow:      c.Location.Row,
			existingComment:    existingComment,
			existingCommentRow: existingCommentRow,
		}
	}
	w.beforeEnd = c
	return nil
}

func truncatedString(s string, max int) string {
	if len(s) > max {
		return s[:max-2] + "..."
	}
	return s
}

func (w *writer) delayBeforeEnd() {
	w.delay = true
}

// line prints a blank line. If the writer is currently in the middle of a line,
// line ends it and then prints a blank one.
func (w *writer) blankLine() {
	if w.inline {
		w.endLine()
	}
	w.write("\n")
}

// write the input string and writes it to the buffer.
func (w *writer) write(s string) {
	w.buf.WriteString(s)
}

// writeLine writes the string on a newly started line, then terminate the line.
func (w *writer) writeLine(s string) {
	if !w.inline {
		w.startLine()
	}
	w.write(s)
	w.endLine()
}

func (w *writer) startMultilineSeq() {
	w.endLine()
	w.up()
	w.startLine()
}

// up increases the indentation level
func (w *writer) up() {
	w.level++
}

// down decreases the indentation level
func (w *writer) down() error {
	if w.level == 0 {
		return errors.New("negative indentation level")
	}
	w.level--
	return nil
}

func ensureFutureKeywordImport(imps []*ast.Import, kw string) []*ast.Import {
	for _, imp := range imps {
		if future.IsAllFutureKeywords(imp) ||
			future.IsFutureKeyword(imp, kw) ||
			(future.IsFutureKeyword(imp, "every") && kw == "in") { // "every" implies "in", so we don't need to add both
			return imps
		}
	}
	imp := &ast.Import{
		// NOTE: This is a hack to not error on the ref containing a keyword already present in v1.
		// A cleaner solution would be to instead allow refs to contain keyword terms.
		// E.g. in v1, `import future.keywords["in"]` is valid, but `import future.keywords.in` is not
		// as it contains a reserved keyword.
		Path: ast.MustParseTerm("future.keywords[\"" + kw + "\"]"),
		//Path: ast.MustParseTerm("future.keywords." + kw),
	}
	imp.Location = defaultLocation(imp)
	return append(imps, imp)
}

func ensureRegoV1Import(imps []*ast.Import) []*ast.Import {
	return ensureImport(imps, ast.RegoV1CompatibleRef)
}

func filterRegoV1Import(imps []*ast.Import) []*ast.Import {
	var ret []*ast.Import
	for _, imp := range imps {
		path := imp.Path.Value.(ast.Ref)
		if !ast.RegoV1CompatibleRef.Equal(path) {
			ret = append(ret, imp)
		}
	}
	return ret
}

func ensureImport(imps []*ast.Import, path ast.Ref) []*ast.Import {
	for _, imp := range imps {
		p := imp.Path.Value.(ast.Ref)
		if p.Equal(path) {
			return imps
		}
	}
	imp := &ast.Import{
		Path: ast.NewTerm(path),
	}
	imp.Location = defaultLocation(imp)
	return append(imps, imp)
}

// ArityFormatErrDetail but for `fmt` checks since compiler has not run yet.
type ArityFormatErrDetail struct {
	Have []string `json:"have"`
	Want []string `json:"want"`
}

// ArityFormatMismatchError but for `fmt` checks since the compiler has not run yet.
func ArityFormatMismatchError(operands []*ast.Term, operator string, loc *ast.Location, f *types.Function) *ast.Error {
	want := make([]string, f.Arity())
	for i, arg := range f.FuncArgs().Args {
		want[i] = types.Sprint(arg)
	}

	have := make([]string, len(operands))
	for i := range operands {
		have[i] = ast.ValueName(operands[i].Value)
	}
	err := ast.NewError(ast.TypeErr, loc, "%s: %s", operator, "arity mismatch")
	err.Details = &ArityFormatErrDetail{
		Have: have,
		Want: want,
	}
	return err
}

// Lines returns the string representation of the detail.
func (d *ArityFormatErrDetail) Lines() []string {
	return []string{
		"have: (" + strings.Join(d.Have, ",") + ")",
		"want: (" + strings.Join(d.Want, ",") + ")",
	}
}

func moduleIsRegoV1Compatible(m *ast.Module) bool {
	for _, imp := range m.Imports {
		if isRegoV1Compatible(imp) {
			return true
		}
	}
	return false
}

var v1StringTerm = ast.StringTerm("v1")

// isRegoV1Compatible returns true if the passed *ast.Import is `rego.v1`
func isRegoV1Compatible(imp *ast.Import) bool {
	path := imp.Path.Value.(ast.Ref)
	return len(path) == 2 &&
		ast.RegoRootDocument.Equal(path[0]) &&
		path[1].Equal(v1StringTerm)
}
