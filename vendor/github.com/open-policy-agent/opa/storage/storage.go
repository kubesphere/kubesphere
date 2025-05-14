// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package storage

import (
	"context"

	v1 "github.com/open-policy-agent/opa/v1/storage"
)

// NewTransactionOrDie is a helper function to create a new transaction. If the
// storage layer cannot create a new transaction, this function will panic. This
// function should only be used for tests.
func NewTransactionOrDie(ctx context.Context, store Store, params ...TransactionParams) Transaction {
	return v1.NewTransactionOrDie(ctx, store, params...)
}

// ReadOne is a convenience function to read a single value from the provided Store. It
// will create a new Transaction to perform the read with, and clean up after itself
// should an error occur.
func ReadOne(ctx context.Context, store Store, path Path) (interface{}, error) {
	return v1.ReadOne(ctx, store, path)
}

// WriteOne is a convenience function to write a single value to the provided Store. It
// will create a new Transaction to perform the write with, and clean up after itself
// should an error occur.
func WriteOne(ctx context.Context, store Store, op PatchOp, path Path, value interface{}) error {
	return v1.WriteOne(ctx, store, op, path, value)
}

// MakeDir inserts an empty object at path. If the parent path does not exist,
// MakeDir will create it recursively.
func MakeDir(ctx context.Context, store Store, txn Transaction, path Path) error {
	return v1.MakeDir(ctx, store, txn, path)
}

// Txn is a convenience function that executes f inside a new transaction
// opened on the store. If the function returns an error, the transaction is
// aborted and the error is returned. Otherwise, the transaction is committed
// and the result of the commit is returned.
func Txn(ctx context.Context, store Store, params TransactionParams, f func(Transaction) error) error {
	return v1.Txn(ctx, store, params, f)
}

// NonEmpty returns a function that tests if a path is non-empty. A
// path is non-empty if a Read on the path returns a value or a Read
// on any of the path prefixes returns a non-object value.
func NonEmpty(ctx context.Context, store Store, txn Transaction) func([]string) (bool, error) {
	return v1.NonEmpty(ctx, store, txn)
}
