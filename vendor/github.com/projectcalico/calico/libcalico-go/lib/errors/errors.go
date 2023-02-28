// Copyright (c) 2020 Tigera, Inc. All rights reserved.

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
	"net/http"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Error indicating a problem connecting to the backend.
type ErrorDatastoreError struct {
	Err        error
	Identifier interface{}
}

func (e ErrorDatastoreError) Error() string {
	return e.Err.Error()
}

func (e ErrorDatastoreError) Status() metav1.Status {
	if i, ok := e.Err.(apierrors.APIStatus); ok {
		return i.Status()
	}

	// Just wrap in a status error.
	return metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusBadRequest,
		Reason:  metav1.StatusReasonInvalid,
		Message: fmt.Sprintf(e.Error()),
		Details: &metav1.StatusDetails{
			Name: fmt.Sprintf("%v", e.Identifier),
		},
	}
}

// Error indicating a resource does not exist.  Used when attempting to delete or
// update a non-existent resource.
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
		return "unknown validation error"
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

// Error indicating that the caller has attempted to release an IP address using
// outdated information.
type ErrorBadHandle struct {
	Requested string
	Expected  string
}

func (e ErrorBadHandle) Error() string {
	f := "the given handle (%s) does not match (%s) when attempting to release IP"
	return fmt.Sprintf(f, e.Requested, e.Expected)
}

// Error indicating that the caller has attempted to release an IP address using
// outdated information.
type ErrorBadSequenceNumber struct {
	Requested uint64
	Expected  uint64
}

func (e ErrorBadSequenceNumber) Error() string {
	f := "the given sequence number (%d) does not match (%d) when attempting to release IP"
	return fmt.Sprintf(f, e.Requested, e.Expected)
}

// Error indicating that the operation may have partially succeeded, then
// failed, without rolling back. A common example is when a function failed
// in an acceptable way after it successfully wrote some data to the datastore.
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

// Error indicating the datastore has failed to parse an entry.
type ErrorParsingDatastoreEntry struct {
	RawKey   string
	RawValue string
	Err      error
}

func (e ErrorParsingDatastoreEntry) Error() string {
	return fmt.Sprintf("failed to parse datastore entry key=%s; value=%s: %v", e.RawKey, e.RawValue, e.Err)
}

type ErrorPolicyConversionRule struct {
	EgressRule  *networkingv1.NetworkPolicyEgressRule
	IngressRule *networkingv1.NetworkPolicyIngressRule
	Reason      string
}

func (e ErrorPolicyConversionRule) String() string {
	var fieldString string

	switch {
	case e.EgressRule != nil:
		fieldString = fmt.Sprintf("%+v", e.EgressRule)
	case e.IngressRule != nil:
		fieldString = fmt.Sprintf("%+v", e.IngressRule)
	default:
		fieldString = "unknown rule"
	}

	if e.Reason != "" {
		fieldString = fmt.Sprintf("%s (%s)", fieldString, e.Reason)
	}

	return fieldString
}

type ErrorPolicyConversion struct {
	PolicyName string
	Rules      []ErrorPolicyConversionRule
}

func (e *ErrorPolicyConversion) BadEgressRule(rule *networkingv1.NetworkPolicyEgressRule, reason string) {
	// Copy rule
	badRule := *rule

	e.Rules = append(e.Rules, ErrorPolicyConversionRule{
		EgressRule:  &badRule,
		IngressRule: nil,
		Reason:      reason,
	})
}

func (e *ErrorPolicyConversion) BadIngressRule(
	rule *networkingv1.NetworkPolicyIngressRule, reason string) {
	// Copy rule
	badRule := *rule

	e.Rules = append(e.Rules, ErrorPolicyConversionRule{
		EgressRule:  nil,
		IngressRule: &badRule,
		Reason:      reason,
	})
}

func (e ErrorPolicyConversion) Error() string {
	s := fmt.Sprintf("policy: %s", e.PolicyName)

	switch {
	case len(e.Rules) == 0:
		s += ": unknown policy conversion error"
	case len(e.Rules) == 1:
		f := e.Rules[0]

		s += fmt.Sprintf(": error with rule %s", f)
	default:
		s += ": error with the following rules:\n"
		for _, f := range e.Rules {
			s += fmt.Sprintf("-  %s\n", f)
		}
	}

	return s
}

func (e ErrorPolicyConversion) GetError() error {
	if len(e.Rules) == 0 {
		return nil
	}

	return e
}
