// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"io"

	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// VersonIndex contains an index from built-in function name, language feature,
// and future rego keyword to version number. During the build, this is used to
// create an index of the minimum version required for the built-in/feature/kw.
type VersionIndex = v1.VersionIndex

// In the compiler, we used this to check that we're OK working with ref heads.
// If this isn't present, we'll fail. This is to ensure that older versions of
// OPA can work with policies that we're compiling -- if they don't know ref
// heads, they wouldn't be able to parse them.
const FeatureRefHeadStringPrefixes = v1.FeatureRefHeadStringPrefixes
const FeatureRefHeads = v1.FeatureRefHeads
const FeatureRegoV1 = v1.FeatureRegoV1
const FeatureRegoV1Import = v1.FeatureRegoV1Import

// Capabilities defines a structure containing data that describes the capabilities
// or features supported by a particular version of OPA.
type Capabilities = v1.Capabilities

// WasmABIVersion captures the Wasm ABI version. Its `Minor` version is indicating
// backwards-compatible changes.
type WasmABIVersion = v1.WasmABIVersion

// CapabilitiesForThisVersion returns the capabilities of this version of OPA.
func CapabilitiesForThisVersion() *Capabilities {
	return v1.CapabilitiesForThisVersion(v1.CapabilitiesRegoVersion(DefaultRegoVersion))
}

// LoadCapabilitiesJSON loads a JSON serialized capabilities structure from the reader r.
func LoadCapabilitiesJSON(r io.Reader) (*Capabilities, error) {
	return v1.LoadCapabilitiesJSON(r)
}

// LoadCapabilitiesVersion loads a JSON serialized capabilities structure from the specific version.
func LoadCapabilitiesVersion(version string) (*Capabilities, error) {
	return v1.LoadCapabilitiesVersion(version)
}

// LoadCapabilitiesFile loads a JSON serialized capabilities structure from a file.
func LoadCapabilitiesFile(file string) (*Capabilities, error) {
	return v1.LoadCapabilitiesFile(file)
}

// LoadCapabilitiesVersions loads all capabilities versions
func LoadCapabilitiesVersions() ([]string, error) {
	return v1.LoadCapabilitiesVersions()
}
