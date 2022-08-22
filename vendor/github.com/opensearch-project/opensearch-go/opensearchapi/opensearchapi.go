// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package opensearchapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/opensearch-project/opensearch-go/internal/version"
)

// Version returns the package version as a string.
//
const Version = version.Client

// Transport defines the interface for an API client.
//
type Transport interface {
	Perform(*http.Request) (*http.Response, error)
}

// BoolPtr returns a pointer to v.
//
// It is used as a convenience function for converting a bool value
// into a pointer when passing the value to a function or struct field
// which expects a pointer.
//
func BoolPtr(v bool) *bool { return &v }

// IntPtr returns a pointer to v.
//
// It is used as a convenience function for converting an int value
// into a pointer when passing the value to a function or struct field
// which expects a pointer.
//
func IntPtr(v int) *int { return &v }

// formatDuration converts duration to a string in the format
// accepted by Elasticsearch.
//
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return strconv.FormatInt(int64(d), 10) + "nanos"
	}
	return strconv.FormatInt(int64(d)/int64(time.Millisecond), 10) + "ms"
}
