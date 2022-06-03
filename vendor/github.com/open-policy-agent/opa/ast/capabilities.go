// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"io"
	"sort"

	"github.com/open-policy-agent/opa/util"
)

// Capabilities defines a structure containing data that describes the capablilities
// or features supported by a particular version of OPA.
type Capabilities struct {
	Builtins []*Builtin `json:"builtins"` // builtins is a set of built-in functions that are supported.
}

// CapabilitiesForThisVersion returns the capabilities of this version of OPA.
func CapabilitiesForThisVersion() *Capabilities {

	f := &Capabilities{
		Builtins: []*Builtin{},
	}

	for _, bi := range Builtins {
		f.Builtins = append(f.Builtins, bi)
	}

	sort.Slice(f.Builtins, func(i, j int) bool {
		return f.Builtins[i].Name < f.Builtins[j].Name
	})

	return f
}

// LoadCapabilitiesJSON loads a JSON serialized capabilities structure from the reader r.
func LoadCapabilitiesJSON(r io.Reader) (*Capabilities, error) {
	d := util.NewJSONDecoder(r)
	var c Capabilities
	return &c, d.Decode(&c)
}
