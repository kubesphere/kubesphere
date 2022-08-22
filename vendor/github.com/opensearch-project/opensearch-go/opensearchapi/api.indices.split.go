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

func newIndicesSplitFunc(t Transport) IndicesSplit {
	return func(index string, target string, o ...func(*IndicesSplitRequest)) (*Response, error) {
		var r = IndicesSplitRequest{Index: index, Target: target}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesSplit allows you to split an existing index into a new index with more primary shards.
//
//
type IndicesSplit func(index string, target string, o ...func(*IndicesSplitRequest)) (*Response, error)

// IndicesSplitRequest configures the Indices Split API request.
//
type IndicesSplitRequest struct {
	Index string

	Body io.Reader

	Target string

	CopySettings        *bool
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
func (r IndicesSplitRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(r.Index) + 1 + len("_split") + 1 + len(r.Target))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_split")
	path.WriteString("/")
	path.WriteString(r.Target)

	params = make(map[string]string)

	if r.CopySettings != nil {
		params["copy_settings"] = strconv.FormatBool(*r.CopySettings)
	}

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
func (f IndicesSplit) WithContext(v context.Context) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.ctx = v
	}
}

// WithBody - The configuration for the target index (`settings` and `aliases`).
//
func (f IndicesSplit) WithBody(v io.Reader) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.Body = v
	}
}

// WithCopySettings - whether or not to copy settings from the source index (defaults to false).
//
func (f IndicesSplit) WithCopySettings(v bool) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.CopySettings = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesSplit) WithMasterTimeout(v time.Duration) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesSplit) WithTimeout(v time.Duration) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - set the number of active shards to wait for on the shrunken index before the operation returns..
//
func (f IndicesSplit) WithWaitForActiveShards(v string) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesSplit) WithPretty() func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesSplit) WithHuman() func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesSplit) WithErrorTrace() func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesSplit) WithFilterPath(v ...string) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesSplit) WithHeader(h map[string]string) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
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
func (f IndicesSplit) WithOpaqueID(s string) func(*IndicesSplitRequest) {
	return func(r *IndicesSplitRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
