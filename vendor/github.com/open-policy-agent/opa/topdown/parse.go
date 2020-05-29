// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"encoding/json"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func builtinRegoParseModule(a, b ast.Value) (ast.Value, error) {

	filename, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	input, err := builtins.StringOperand(b, 1)
	if err != nil {
		return nil, err
	}

	module, err := ast.ParseModule(string(filename), string(input))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(module); err != nil {
		return nil, err
	}

	term, err := ast.ParseTerm(buf.String())
	if err != nil {
		return nil, err
	}

	return term.Value, nil
}

func init() {
	RegisterFunctionalBuiltin2(ast.RegoParseModule.Name, builtinRegoParseModule)
}
