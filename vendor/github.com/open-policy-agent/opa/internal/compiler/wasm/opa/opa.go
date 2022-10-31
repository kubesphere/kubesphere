// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package opa contains bytecode for the OPA-WASM library.
package opa

import (
	_ "embed"
)

//go:embed opa.wasm
var wasmBase []byte

//go:embed callgraph.csv
var callGraphCSV []byte

// Bytes returns the OPA-WASM bytecode.
func Bytes() []byte {
	return wasmBase
}

// CallGraphCSV returns a CSV representation of the
// OPA-WASM bytecode's call graph: 'caller,callee'
func CallGraphCSV() []byte {
	return callGraphCSV
}
