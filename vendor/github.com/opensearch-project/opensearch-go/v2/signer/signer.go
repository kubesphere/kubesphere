// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

package signer

import "net/http"

//Signer an interface that will sign http.Request
type Signer interface {
	SignRequest(request *http.Request) error
}
