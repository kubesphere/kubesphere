// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package injectproxy

import (
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

type Enforcer struct {
	labelMatchers  map[string]*labels.Matcher
	errorOnReplace bool
}

func NewEnforcer(errorOnReplace bool, ms ...*labels.Matcher) *Enforcer {
	entries := make(map[string]*labels.Matcher)

	for _, matcher := range ms {
		entries[matcher.Name] = matcher
	}

	return &Enforcer{
		labelMatchers:  entries,
		errorOnReplace: errorOnReplace,
	}
}

type IllegalLabelMatcherError struct {
	msg string
}

func (e IllegalLabelMatcherError) Error() string { return e.msg }

func newIllegalLabelMatcherError(existing string, replacement string) IllegalLabelMatcherError {
	return IllegalLabelMatcherError{
		msg: fmt.Sprintf("label matcher value (%s) conflicts with injected value (%s)", existing, replacement),
	}
}

// EnforceNode walks the given node recursively
// and enforces the given label enforcer on it.
//
// Whenever a parser.MatrixSelector or parser.VectorSelector AST node is found,
// their label enforcer is being potentially modified.
// If a node's label matcher has the same name as a label matcher
// of the given enforcer, then it will be replaced.
func (ms Enforcer) EnforceNode(node parser.Node) error {
	switch n := node.(type) {
	case *parser.EvalStmt:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case parser.Expressions:
		for _, e := range n {
			if err := ms.EnforceNode(e); err != nil {
				return err
			}
		}

	case *parser.AggregateExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.BinaryExpr:
		if err := ms.EnforceNode(n.LHS); err != nil {
			return err
		}

		if err := ms.EnforceNode(n.RHS); err != nil {
			return err
		}

	case *parser.Call:
		if err := ms.EnforceNode(n.Args); err != nil {
			return err
		}

	case *parser.SubqueryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.ParenExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.UnaryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *parser.NumberLiteral, *parser.StringLiteral:
	// nothing to do

	case *parser.MatrixSelector:
		// inject labelselector
		if vs, ok := n.VectorSelector.(*parser.VectorSelector); ok {
			var err error
			vs.LabelMatchers, err = ms.EnforceMatchers(vs.LabelMatchers)
			if err != nil {
				return err
			}
		}

	case *parser.VectorSelector:
		// inject labelselector
		var err error
		n.LabelMatchers, err = ms.EnforceMatchers(n.LabelMatchers)
		if err != nil {
			return err
		}

	default:
		panic(fmt.Errorf("parser.Walk: unhandled node type %T", n))
	}

	return nil
}

// EnforceMatchers appends the configured label matcher if not present.
// If the label matcher that is to be injected is present (by labelname) but
// different (either by match type or value) the behavior depends on the
// errorOnReplace variable. If errorOnReplace is true an error is returned,
// otherwise the label matcher is silently replaced.
func (ms Enforcer) EnforceMatchers(targets []*labels.Matcher) ([]*labels.Matcher, error) {
	var res []*labels.Matcher

	for _, target := range targets {
		if matcher, ok := ms.labelMatchers[target.Name]; ok {
			// matcher.String() returns something like "labelfoo=value"
			if ms.errorOnReplace && matcher.String() != target.String() {
				return res, newIllegalLabelMatcherError(matcher.String(), target.String())
			}
			continue
		}

		res = append(res, target)
	}

	for _, enforcedMatcher := range ms.labelMatchers {
		res = append(res, enforcedMatcher)
	}

	return res, nil
}
