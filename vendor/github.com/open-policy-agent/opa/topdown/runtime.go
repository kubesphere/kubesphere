// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import "github.com/open-policy-agent/opa/ast"

func builtinOPARuntime(bctx BuiltinContext, _ []*ast.Term, iter func(*ast.Term) error) error {

	if bctx.Runtime == nil {
		return iter(ast.ObjectTerm())
	}

	return iter(bctx.Runtime)
}

func init() {
	RegisterBuiltinFunc(ast.OPARuntime.Name, builtinOPARuntime)
}
