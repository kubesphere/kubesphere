// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

// CompileModules takes a set of Rego modules represented as strings and
// compiles them for evaluation. The keys of the map are used as filenames.
func CompileModules(modules map[string]string) (*Compiler, error) {

	parsed := make(map[string]*Module, len(modules))

	for f, module := range modules {
		var pm *Module
		var err error
		if pm, err = ParseModule(f, module); err != nil {
			return nil, err
		}
		parsed[f] = pm
	}

	compiler := NewCompiler()
	compiler.Compile(parsed)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

// MustCompileModules compiles a set of Rego modules represented as strings. If
// the compilation process fails, this function panics.
func MustCompileModules(modules map[string]string) *Compiler {

	compiler, err := CompileModules(modules)
	if err != nil {
		panic(err)
	}

	return compiler
}
