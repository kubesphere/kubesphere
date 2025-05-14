// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package rego

import (
	"context"
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/ir"
)

var targetPlugins = map[string]TargetPlugin{}
var pluginMtx sync.Mutex

type TargetPlugin interface {
	IsTarget(string) bool
	PrepareForEval(context.Context, *ir.Policy, ...PrepareOption) (TargetPluginEval, error)
}

type TargetPluginEval interface {
	Eval(context.Context, *EvalContext, ast.Value) (ast.Value, error)
}

func (*Rego) targetPlugin(tgt string) TargetPlugin {
	for _, p := range targetPlugins {
		if p.IsTarget(tgt) {
			return p
		}
	}
	return nil
}

func RegisterPlugin(name string, p TargetPlugin) {
	pluginMtx.Lock()
	defer pluginMtx.Unlock()
	if _, ok := targetPlugins[name]; ok {
		panic("plugin already registered " + name)
	}
	targetPlugins[name] = p
}
