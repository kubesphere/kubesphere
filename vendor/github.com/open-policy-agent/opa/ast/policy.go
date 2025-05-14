// Copyright 2024 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	astJSON "github.com/open-policy-agent/opa/ast/json"
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// DefaultRootDocument is the default root document.
//
// All package directives inside source files are implicitly prefixed with the
// DefaultRootDocument value.
var DefaultRootDocument = v1.DefaultRootDocument

// InputRootDocument names the document containing query arguments.
var InputRootDocument = v1.InputRootDocument

// SchemaRootDocument names the document containing external data schemas.
var SchemaRootDocument = v1.SchemaRootDocument

// FunctionArgRootDocument names the document containing function arguments.
// It's only for internal usage, for referencing function arguments between
// the index and topdown.
var FunctionArgRootDocument = v1.FunctionArgRootDocument

// FutureRootDocument names the document containing new, to-become-default,
// features.
var FutureRootDocument = v1.FutureRootDocument

// RegoRootDocument names the document containing new, to-become-default,
// features in a future versioned release.
var RegoRootDocument = v1.RegoRootDocument

// RootDocumentNames contains the names of top-level documents that can be
// referred to in modules and queries.
//
// Note, the schema document is not currently implemented in the evaluator so it
// is not registered as a root document name (yet).
var RootDocumentNames = v1.RootDocumentNames

// DefaultRootRef is a reference to the root of the default document.
//
// All refs to data in the policy engine's storage layer are prefixed with this ref.
var DefaultRootRef = v1.DefaultRootRef

// InputRootRef is a reference to the root of the input document.
//
// All refs to query arguments are prefixed with this ref.
var InputRootRef = v1.InputRootRef

// SchemaRootRef is a reference to the root of the schema document.
//
// All refs to schema documents are prefixed with this ref. Note, the schema
// document is not currently implemented in the evaluator so it is not
// registered as a root document ref (yet).
var SchemaRootRef = v1.SchemaRootRef

// RootDocumentRefs contains the prefixes of top-level documents that all
// non-local references start with.
var RootDocumentRefs = v1.RootDocumentRefs

// SystemDocumentKey is the name of the top-level key that identifies the system
// document.
const SystemDocumentKey = v1.SystemDocumentKey

// ReservedVars is the set of names that refer to implicitly ground vars.
var ReservedVars = v1.ReservedVars

// Wildcard represents the wildcard variable as defined in the language.
var Wildcard = v1.Wildcard

// WildcardPrefix is the special character that all wildcard variables are
// prefixed with when the statement they are contained in is parsed.
const WildcardPrefix = v1.WildcardPrefix

// Keywords contains strings that map to language keywords.
var Keywords = v1.Keywords

var KeywordsV0 = v1.KeywordsV0

var KeywordsV1 = v1.KeywordsV1

func KeywordsForRegoVersion(v RegoVersion) []string {
	return v1.KeywordsForRegoVersion(v)
}

// IsKeyword returns true if s is a language keyword.
func IsKeyword(s string) bool {
	return v1.IsKeyword(s)
}

func IsInKeywords(s string, keywords []string) bool {
	return v1.IsInKeywords(s, keywords)
}

// IsKeywordInRegoVersion returns true if s is a language keyword.
func IsKeywordInRegoVersion(s string, regoVersion RegoVersion) bool {
	return v1.IsKeywordInRegoVersion(s, regoVersion)
}

type (
	// Node represents a node in an AST. Nodes may be statements in a policy module
	// or elements of an ad-hoc query, expression, etc.
	Node = v1.Node

	// Statement represents a single statement in a policy module.
	Statement = v1.Statement
)

type (

	// Module represents a collection of policies (defined by rules)
	// within a namespace (defined by the package) and optional
	// dependencies on external documents (defined by imports).
	Module = v1.Module

	// Comment contains the raw text from the comment in the definition.
	Comment = v1.Comment

	// Package represents the namespace of the documents produced
	// by rules inside the module.
	Package = v1.Package

	// Import represents a dependency on a document outside of the policy
	// namespace. Imports are optional.
	Import = v1.Import

	// Rule represents a rule as defined in the language. Rules define the
	// content of documents that represent policy decisions.
	Rule = v1.Rule

	// Head represents the head of a rule.
	Head = v1.Head

	// Args represents zero or more arguments to a rule.
	Args = v1.Args

	// Body represents one or more expressions contained inside a rule or user
	// function.
	Body = v1.Body

	// Expr represents a single expression contained inside the body of a rule.
	Expr = v1.Expr

	// SomeDecl represents a variable declaration statement. The symbols are variables.
	SomeDecl = v1.SomeDecl

	Every = v1.Every

	// With represents a modifier on an expression.
	With = v1.With
)

// NewComment returns a new Comment object.
func NewComment(text []byte) *Comment {
	return v1.NewComment(text)
}

// IsValidImportPath returns an error indicating if the import path is invalid.
// If the import path is valid, err is nil.
func IsValidImportPath(v Value) (err error) {
	return v1.IsValidImportPath(v)
}

// NewHead returns a new Head object. If args are provided, the first will be
// used for the key and the second will be used for the value.
func NewHead(name Var, args ...*Term) *Head {
	return v1.NewHead(name, args...)
}

// VarHead creates a head object, initializes its Name, Location, and Options,
// and returns the new head.
func VarHead(name Var, location *Location, jsonOpts *astJSON.Options) *Head {
	return v1.VarHead(name, location, jsonOpts)
}

// RefHead returns a new Head object with the passed Ref. If args are provided,
// the first will be used for the value.
func RefHead(ref Ref, args ...*Term) *Head {
	return v1.RefHead(ref, args...)
}

// DocKind represents the collection of document types that can be produced by rules.
type DocKind = v1.DocKind

const (
	// CompleteDoc represents a document that is completely defined by the rule.
	CompleteDoc = v1.CompleteDoc

	// PartialSetDoc represents a set document that is partially defined by the rule.
	PartialSetDoc = v1.PartialSetDoc

	// PartialObjectDoc represents an object document that is partially defined by the rule.
	PartialObjectDoc = v1.PartialObjectDoc
)

type RuleKind = v1.RuleKind

const (
	SingleValue = v1.SingleValue
	MultiValue  = v1.MultiValue
)

// NewBody returns a new Body containing the given expressions. The indices of
// the immediate expressions will be reset.
func NewBody(exprs ...*Expr) Body {
	return v1.NewBody(exprs...)
}

// NewExpr returns a new Expr object.
func NewExpr(terms interface{}) *Expr {
	return v1.NewExpr(terms)
}

// NewBuiltinExpr creates a new Expr object with the supplied terms.
// The builtin operator must be the first term.
func NewBuiltinExpr(terms ...*Term) *Expr {
	return v1.NewBuiltinExpr(terms...)
}

// Copy returns a deep copy of the AST node x. If x is not an AST node, x is returned unmodified.
func Copy(x interface{}) interface{} {
	return v1.Copy(x)
}

// RuleSet represents a collection of rules that produce a virtual document.
type RuleSet = v1.RuleSet

// NewRuleSet returns a new RuleSet containing the given rules.
func NewRuleSet(rules ...*Rule) RuleSet {
	return v1.NewRuleSet(rules...)
}
