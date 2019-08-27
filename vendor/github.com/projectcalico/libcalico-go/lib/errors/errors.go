// Copyright (c) 2016 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"
)

// Error indicating a problem connecting to the backend.
type ErrorDatastoreError struct {
	Err        error
	Identifier interface{}
}

func (e ErrorDatastoreError) Error() string {
	return e.Err.Error()
}

// Error indicating a resource does not exist.  Used when attempting to delete or
// udpate a non-existent resource.
type ErrorResourceDoesNotExist struct {
	Err        error
	Identifier interface{}
}

func (e ErrorResourceDoesNotExist) Error() string {
	return fmt.Sprintf("resource does not exist: %v with error: %v", e.Identifier, e.Err)
}

// Error indicating an operation is not supported.
type ErrorOperationNotSupported struct {
	Operation  string
	Identifier interface{}
	Reason     string
}

func (e ErrorOperationNotSupported) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("operation %s is not supported on %v", e.Operation, e.Identifier)
	} else {
		return fmt.Sprintf("operation %s is not supported on %v: %s", e.Operation, e.Identifier, e.Reason)
	}
}

// Error indicating a resource already exists.  Used when attempting to create a
// resource that already exists.
type ErrorResourceAlreadyExists struct {
	Err        error
	Identifier interface{}
}

func (e ErrorResourceAlreadyExists) Error() string {
	return fmt.Sprintf("resource already exists: %v", e.Identifier)
}

// Error indicating a problem connecting to the backend.
type ErrorConnectionUnauthorized struct {
	Err error
}

func (e ErrorConnectionUnauthorized) Error() string {
	return fmt.Sprintf("connection is unauthorized: %v", e.Err)
}

// Validation error containing the fields that are failed validation.
type ErrorValidation struct {
	ErroredFields []ErroredField
}

type ErroredField struct {
	Name   string
	Value  interface{}
	Reason string
}

func (e ErroredField) String() string {
	var fieldString string
	if e.Value == nil {
		fieldString = e.Name
	} else {
		fieldString = fmt.Sprintf("%s = '%v'", e.Name, e.Value)
	}
	if e.Reason != "" {
		fieldString = fmt.Sprintf("%s (%s)", fieldString, e.Reason)
	}
	return fieldString
}

func (e ErrorValidation) Error() string {
	if len(e.ErroredFields) == 0 {
		return fmt.Sprintf("unknown validation error: %v", e)
	} else if len(e.ErroredFields) == 1 {
		f := e.ErroredFields[0]
		return fmt.Sprintf("error with field %s", f)
	} else {
		s := "error with the following fields:\n"
		for _, f := range e.ErroredFields {
			s = s + fmt.Sprintf("-  %s\n", f)
		}
		return s
	}
}

// Error indicating insufficient identifiers have been supplied on a resource
// management request (create, apply, update, get, delete).
type ErrorInsufficientIdentifiers struct {
	Name string
}

func (e ErrorInsufficientIdentifiers) Error() string {
	return fmt.Sprintf("insufficient identifiers, missing '%s'", e.Name)
}

// Error indicating an atomic update attempt that failed due to a update conflict.
type ErrorResourceUpdateConflict struct {
	Err        error
	Identifier interface{}
}

func (e ErrorResourceUpdateConflict) Error() string {
	return fmt.Sprintf("update conflict: %v", e.Identifier)
}

// Error indicating that the operation may have partially succeeded, then
// failed, without rolling back. A common example is when a function failed
// in an acceptable way after it succesfully wrote some data to the datastore.
type ErrorPartialFailure struct {
	Err error
}

func (e ErrorPartialFailure) Error() string {
	return fmt.Sprintf("operation partially failed: %v", e.Err)
}

// UpdateErrorIdentifier modifies the supplied error to use the new resource
// identifier.
func UpdateErrorIdentifier(err error, id interface{}) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case ErrorDatastoreError:
		e.Identifier = id
		err = e
	case ErrorResourceDoesNotExist:
		e.Identifier = id
		err = e
	case ErrorOperationNotSupported:
		e.Identifier = id
		err = e
	case ErrorResourceAlreadyExists:
		e.Identifier = id
		err = e
	case ErrorResourceUpdateConflict:
		e.Identifier = id
		err = e
	}
	return err
}

// Error indicating the watcher has been terminated.
type ErrorWatchTerminated struct {
	Err            error
	ClosedByRemote bool
}

func (e ErrorWatchTerminated) Error() string {
	return fmt.Sprintf("watch terminated (closedByRemote:%v): %v", e.ClosedByRemote, e.Err)
}

// Error indicating the datastore has failed to parse an entry.
type ErrorParsingDatastoreEntry struct {
	RawKey   string
	RawValue string
	Err      error
}

func (e ErrorParsingDatastoreEntry) Error() string {
	return fmt.Sprintf("failed to parse datastore entry key=%s; value=%s: %v", e.RawKey, e.RawValue, e.Err)
}
