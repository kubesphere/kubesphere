// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package storage

import (
	v1 "github.com/open-policy-agent/opa/v1/storage"
)

const (
	// InternalErr indicates an unknown, internal error has occurred.
	InternalErr = v1.InternalErr

	// NotFoundErr indicates the path used in the storage operation does not
	// locate a document.
	NotFoundErr = v1.NotFoundErr

	// WriteConflictErr indicates a write on the path enocuntered a conflicting
	// value inside the transaction.
	WriteConflictErr = v1.WriteConflictErr

	// InvalidPatchErr indicates an invalid patch/write was issued. The patch
	// was rejected.
	InvalidPatchErr = v1.InvalidPatchErr

	// InvalidTransactionErr indicates an invalid operation was performed
	// inside of the transaction.
	InvalidTransactionErr = v1.InvalidTransactionErr

	// TriggersNotSupportedErr indicates the caller attempted to register a
	// trigger against a store that does not support them.
	TriggersNotSupportedErr = v1.TriggersNotSupportedErr

	// WritesNotSupportedErr indicate the caller attempted to perform a write
	// against a store that does not support them.
	WritesNotSupportedErr = v1.WritesNotSupportedErr

	// PolicyNotSupportedErr indicate the caller attempted to perform a policy
	// management operation against a store that does not support them.
	PolicyNotSupportedErr = v1.PolicyNotSupportedErr
)

// Error is the error type returned by the storage layer.
type Error = v1.Error

// IsNotFound returns true if this error is a NotFoundErr.
func IsNotFound(err error) bool {
	return v1.IsNotFound(err)
}

// IsWriteConflictError returns true if this error a WriteConflictErr.
func IsWriteConflictError(err error) bool {
	return v1.IsWriteConflictError(err)
}

// IsInvalidPatch returns true if this error is a InvalidPatchErr.
func IsInvalidPatch(err error) bool {
	return v1.IsInvalidPatch(err)
}

// IsInvalidTransaction returns true if this error is a InvalidTransactionErr.
func IsInvalidTransaction(err error) bool {
	return v1.IsInvalidTransaction(err)
}

// IsIndexingNotSupported is a stub for backwards-compatibility.
//
// Deprecated: We no longer return IndexingNotSupported errors, so it is
// unnecessary to check for them.
func IsIndexingNotSupported(err error) bool {
	return v1.IsIndexingNotSupported(err)
}
