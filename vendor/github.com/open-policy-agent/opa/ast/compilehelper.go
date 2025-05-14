// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import v1 "github.com/open-policy-agent/opa/v1/ast"

// CompileModules takes a set of Rego modules represented as strings and
// compiles them for evaluation. The keys of the map are used as filenames.
func CompileModules(modules map[string]string) (*Compiler, error) {
	return CompileModulesWithOpt(modules, CompileOpts{
		ParserOptions: ParserOptions{
			RegoVersion: DefaultRegoVersion,
		},
	})
}

// CompileOpts defines a set of options for the compiler.
type CompileOpts = v1.CompileOpts

// CompileModulesWithOpt takes a set of Rego modules represented as strings and
// compiles them for evaluation. The keys of the map are used as filenames.
func CompileModulesWithOpt(modules map[string]string, opts CompileOpts) (*Compiler, error) {
	if opts.ParserOptions.RegoVersion == RegoUndefined {
		opts.ParserOptions.RegoVersion = DefaultRegoVersion
	}

	return v1.CompileModulesWithOpt(modules, opts)
}

// MustCompileModules compiles a set of Rego modules represented as strings. If
// the compilation process fails, this function panics.
func MustCompileModules(modules map[string]string) *Compiler {
	return MustCompileModulesWithOpts(modules, CompileOpts{})
}

// MustCompileModulesWithOpts compiles a set of Rego modules represented as strings. If
// the compilation process fails, this function panics.
func MustCompileModulesWithOpts(modules map[string]string, opts CompileOpts) *Compiler {

	compiler, err := CompileModulesWithOpt(modules, opts)
	if err != nil {
		panic(err)
	}

	return compiler
}
