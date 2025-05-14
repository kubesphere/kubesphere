// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package storage

import (
	v1 "github.com/open-policy-agent/opa/v1/storage"
)

// Transaction defines the interface that identifies a consistent snapshot over
// the policy engine's storage layer.
type Transaction = v1.Transaction

// Store defines the interface for the storage layer's backend.
type Store = v1.Store

// MakeDirer defines the interface a Store could realize to override the
// generic MakeDir functionality in storage.MakeDir
type MakeDirer = v1.MakeDirer

// TransactionParams describes a new transaction.
type TransactionParams = v1.TransactionParams

// Context is a simple container for key/value pairs.
type Context = v1.Context

// NewContext returns a new context object.
func NewContext() *Context {
	return v1.NewContext()
}

// WriteParams specifies the TransactionParams for a write transaction.
var WriteParams = v1.WriteParams

// PatchOp is the enumeration of supposed modifications.
type PatchOp = v1.PatchOp

// Patch supports add, remove, and replace operations.
const (
	AddOp     = v1.AddOp
	RemoveOp  = v1.RemoveOp
	ReplaceOp = v1.ReplaceOp
)

// WritesNotSupported provides a default implementation of the write
// interface which may be used if the backend does not support writes.
type WritesNotSupported = v1.WritesNotSupported

// Policy defines the interface for policy module storage.
type Policy = v1.Policy

// PolicyNotSupported provides a default implementation of the policy interface
// which may be used if the backend does not support policy storage.
type PolicyNotSupported = v1.PolicyNotSupported

// PolicyEvent describes a change to a policy.
type PolicyEvent = v1.PolicyEvent

// DataEvent describes a change to a base data document.
type DataEvent = v1.DataEvent

// TriggerEvent describes the changes that caused the trigger to be invoked.
type TriggerEvent = v1.TriggerEvent

// TriggerConfig contains the trigger registration configuration.
type TriggerConfig = v1.TriggerConfig

// Trigger defines the interface that stores implement to register for change
// notifications when the store is changed.
type Trigger = v1.Trigger

// TriggersNotSupported provides default implementations of the Trigger
// interface which may be used if the backend does not support triggers.
type TriggersNotSupported = v1.TriggersNotSupported

// TriggerHandle defines the interface that can be used to unregister triggers that have
// been registered on a Store.
type TriggerHandle = v1.TriggerHandle

// Iterator defines the interface that can be used to read files from a directory starting with
// files at the base of the directory, then sub-directories etc.
type Iterator = v1.Iterator

// Update contains information about a file
type Update = v1.Update
