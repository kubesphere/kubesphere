// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package bundle

import (
	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// HashingAlgorithm represents a subset of hashing algorithms implemented in Go
type HashingAlgorithm = v1.HashingAlgorithm

// Supported values for HashingAlgorithm
const (
	MD5       = v1.MD5
	SHA1      = v1.SHA1
	SHA224    = v1.SHA224
	SHA256    = v1.SHA256
	SHA384    = v1.SHA384
	SHA512    = v1.SHA512
	SHA512224 = v1.SHA512224
	SHA512256 = v1.SHA512256
)

// SignatureHasher computes a signature digest for a file with (structured or unstructured) data and policy
type SignatureHasher = v1.SignatureHasher

// NewSignatureHasher returns a signature hasher suitable for a particular hashing algorithm
func NewSignatureHasher(alg HashingAlgorithm) (SignatureHasher, error) {
	return v1.NewSignatureHasher(alg)
}
