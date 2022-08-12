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
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesAddBlockFunc(t Transport) IndicesAddBlock {
	return func(index []string, block string, o ...func(*IndicesAddBlockRequest)) (*Response, error) {
		var r = IndicesAddBlockRequest{Index: index, Block: block}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesAddBlock adds a block to an index.
//
//
type IndicesAddBlock func(index []string, block string, o ...func(*IndicesAddBlockRequest)) (*Response, error)

// IndicesAddBlockRequest configures the Indices Add Block API request.
//
type IndicesAddBlockRequest struct {
	Index []string

	Block string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
	MasterTimeout     time.Duration
	Timeout           time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesAddBlockRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_block") + 1 + len(r.Block))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_block")
	path.WriteString("/")
	path.WriteString(r.Block)

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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

	req, err := newRequest(method, path.String(), nil)
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
func (f IndicesAddBlock) WithContext(v context.Context) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesAddBlock) WithAllowNoIndices(v bool) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesAddBlock) WithExpandWildcards(v string) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesAddBlock) WithIgnoreUnavailable(v bool) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesAddBlock) WithMasterTimeout(v time.Duration) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesAddBlock) WithTimeout(v time.Duration) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesAddBlock) WithPretty() func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesAddBlock) WithHuman() func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesAddBlock) WithErrorTrace() func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesAddBlock) WithFilterPath(v ...string) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesAddBlock) WithHeader(h map[string]string) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
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
func (f IndicesAddBlock) WithOpaqueID(s string) func(*IndicesAddBlockRequest) {
	return func(r *IndicesAddBlockRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
