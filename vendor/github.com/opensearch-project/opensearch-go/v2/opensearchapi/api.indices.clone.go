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
	"strings"
	"time"
)

func newIndicesCloneFunc(t Transport) IndicesClone {
	return func(index string, target string, o ...func(*IndicesCloneRequest)) (*Response, error) {
		var r = IndicesCloneRequest{Index: index, Target: target}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesClone clones an index
//
//
type IndicesClone func(index string, target string, o ...func(*IndicesCloneRequest)) (*Response, error)

// IndicesCloneRequest configures the Indices Clone API request.
//
type IndicesCloneRequest struct {
	Index string

	Body io.Reader

	Target string

	MasterTimeout       time.Duration
	Timeout             time.Duration
	WaitForActiveShards string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesCloneRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(r.Index) + 1 + len("_clone") + 1 + len(r.Target))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_clone")
	path.WriteString("/")
	path.WriteString(r.Target)

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.WaitForActiveShards != "" {
		params["wait_for_active_shards"] = r.WaitForActiveShards
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
func (f IndicesClone) WithContext(v context.Context) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.ctx = v
	}
}

// WithBody - The configuration for the target index (`settings` and `aliases`).
//
func (f IndicesClone) WithBody(v io.Reader) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.Body = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesClone) WithMasterTimeout(v time.Duration) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesClone) WithTimeout(v time.Duration) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - set the number of active shards to wait for on the cloned index before the operation returns..
//
func (f IndicesClone) WithWaitForActiveShards(v string) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesClone) WithPretty() func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesClone) WithHuman() func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesClone) WithErrorTrace() func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesClone) WithFilterPath(v ...string) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesClone) WithHeader(h map[string]string) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
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
func (f IndicesClone) WithOpaqueID(s string) func(*IndicesCloneRequest) {
	return func(r *IndicesCloneRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
