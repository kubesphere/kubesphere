// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package bundle provide helpers that assist in the bundle signature verification process
package bundle

import (
	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// Verifier is the interface expected for implementations that verify bundle signatures.
type Verifier v1.Verifier

// VerifyBundleSignature will retrieve the Verifier implementation based
// on the Plugin specified in SignaturesConfig, and call its implementation
// of VerifyBundleSignature. VerifyBundleSignature verifies the bundle signature
// using the given public keys or secret. If a signature is verified, it keeps
// track of the files specified in the JWT payload
func VerifyBundleSignature(sc SignaturesConfig, bvc *VerificationConfig) (map[string]FileInfo, error) {
	return v1.VerifyBundleSignature(sc, bvc)
}

// DefaultVerifier is the default bundle verification implementation. It verifies bundles by checking
// the JWT signature using a locally-accessible public key.
type DefaultVerifier = v1.DefaultVerifier

// GetVerifier returns the Verifier registered under the given id
func GetVerifier(id string) (Verifier, error) {
	return v1.GetVerifier(id)
}

// RegisterVerifier registers a Verifier under the given id
func RegisterVerifier(id string, v Verifier) error {
	return v1.RegisterVerifier(id, v)
}
