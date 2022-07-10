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
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newSnapshotRestoreFunc(t Transport) SnapshotRestore {
	return func(repository string, snapshot string, o ...func(*SnapshotRestoreRequest)) (*Response, error) {
		var r = SnapshotRestoreRequest{Repository: repository, Snapshot: snapshot}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotRestore restores a snapshot.
//
//
type SnapshotRestore func(repository string, snapshot string, o ...func(*SnapshotRestoreRequest)) (*Response, error)

// SnapshotRestoreRequest configures the Snapshot Restore API request.
//
type SnapshotRestoreRequest struct {
	Body io.Reader

	Repository string
	Snapshot   string

	MasterTimeout     time.Duration
	WaitForCompletion *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotRestoreRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_snapshot") + 1 + len(r.Repository) + 1 + len(r.Snapshot) + 1 + len("_restore"))
	path.WriteString("/")
	path.WriteString("_snapshot")
	path.WriteString("/")
	path.WriteString(r.Repository)
	path.WriteString("/")
	path.WriteString(r.Snapshot)
	path.WriteString("/")
	path.WriteString("_restore")

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
	}

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, err := newRequest(method, path.String(), r.Body)
	if err != nil {
		return nil, err
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if r.Body != nil {
		req.Header[headerContentType] = headerContentTypeJSON
	}

	if len(r.Header) > 0 {
		if len(req.Header) == 0 {
			req.Header = r.Header
		} else {
			for k, vv := range r.Header {
				for _, v := range vv {
					req.Header.Add(k, v)
				}
			}
		}
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f SnapshotRestore) WithContext(v context.Context) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.ctx = v
	}
}

// WithBody - Details of what to restore.
//
func (f SnapshotRestore) WithBody(v io.Reader) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.Body = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotRestore) WithMasterTimeout(v time.Duration) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.MasterTimeout = v
	}
}

// WithWaitForCompletion - should this request wait until the operation has completed before returning.
//
func (f SnapshotRestore) WithWaitForCompletion(v bool) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotRestore) WithPretty() func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotRestore) WithHuman() func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotRestore) WithErrorTrace() func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotRestore) WithFilterPath(v ...string) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotRestore) WithHeader(h map[string]string) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}

// WithOpaqueID adds the X-Opaque-Id header to the HTTP request.
//
func (f SnapshotRestore) WithOpaqueID(s string) func(*SnapshotRestoreRequest) {
	return func(r *SnapshotRestoreRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
