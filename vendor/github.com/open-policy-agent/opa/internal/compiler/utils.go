// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package compiler

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/schemas"
	"github.com/open-policy-agent/opa/util"
)

type SchemaFile string

const (
	AuthorizationPolicySchema SchemaFile = "authorizationPolicy.json"
)

var schemaDefinitions = map[SchemaFile]interface{}{}

// VerifyAuthorizationPolicySchema performs type checking on rules against the schema for the Authorization Policy
// Input document.
// NOTE: The provided compiler should have already run the compilation process on the input modules
func VerifyAuthorizationPolicySchema(compiler *ast.Compiler, ref ast.Ref) error {

	rules := getRulesWithDependencies(compiler, ref)

	if len(rules) == 0 {
		return nil
	}

	schemaSet := ast.NewSchemaSet()
	schemaSet.Put(ast.SchemaRootRef, schemaDefinitions[AuthorizationPolicySchema])

	errs := ast.NewCompiler().WithSchemas(schemaSet).PassesTypeCheckRules(rules)

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// getRulesWithDependencies returns a slice of rules that are referred to by ref along with their dependencies
func getRulesWithDependencies(compiler *ast.Compiler, ref ast.Ref) []*ast.Rule {
	allRules := compiler.GetRules(ref)

	deps := map[*ast.Rule]struct{}{}
	for _, rule := range allRules {
		transitiveDependencies(compiler, rule, deps)
	}

	for dep := range deps {
		allRules = append(allRules, dep)
	}

	return allRules
}

func transitiveDependencies(compiler *ast.Compiler, rule *ast.Rule, deps map[*ast.Rule]struct{}) {
	for x := range compiler.Graph.Dependencies(rule) {
		other := x.(*ast.Rule)
		deps[other] = struct{}{}
		transitiveDependencies(compiler, other, deps)
	}
}

func loadAuthorizationPolicySchema() {

	cont, err := schemas.FS.ReadFile(string(AuthorizationPolicySchema))
	if err != nil {
		panic(err)
	}

	if len(cont) == 0 {
		panic("expected authorization policy schema file to be present")
	}

	var schema interface{}
	if err := util.Unmarshal(cont, &schema); err != nil {
		panic(err)
	}

	schemaDefinitions[AuthorizationPolicySchema] = schema
}

func init() {
	loadAuthorizationPolicySchema()
}
