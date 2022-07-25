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
)

func newIndicesFlushFunc(t Transport) IndicesFlush {
	return func(o ...func(*IndicesFlushRequest)) (*Response, error) {
		var r = IndicesFlushRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesFlush performs the flush operation on one or more indices.
//
//
type IndicesFlush func(o ...func(*IndicesFlushRequest)) (*Response, error)

// IndicesFlushRequest configures the Indices Flush API request.
//
type IndicesFlushRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	Force             *bool
	IgnoreUnavailable *bool
	WaitIfOngoing     *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesFlushRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_flush"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_flush")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Force != nil {
		params["force"] = strconv.FormatBool(*r.Force)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.WaitIfOngoing != nil {
		params["wait_if_ongoing"] = strconv.FormatBool(*r.WaitIfOngoing)
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
func (f IndicesFlush) WithContext(v context.Context) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all for all indices.
//
func (f IndicesFlush) WithIndex(v ...string) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesFlush) WithAllowNoIndices(v bool) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesFlush) WithExpandWildcards(v string) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.ExpandWildcards = v
	}
}

// WithForce - whether a flush should be forced even if it is not necessarily needed ie. if no changes will be committed to the index. this is useful if transaction log ids should be incremented even if no uncommitted changes are present. (this setting can be considered as internal).
//
func (f IndicesFlush) WithForce(v bool) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.Force = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesFlush) WithIgnoreUnavailable(v bool) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithWaitIfOngoing - if set to true the flush operation will block until the flush can be executed if another flush operation is already executing. the default is true. if set to false the flush will be skipped iff if another flush operation is already running..
//
func (f IndicesFlush) WithWaitIfOngoing(v bool) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.WaitIfOngoing = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesFlush) WithPretty() func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesFlush) WithHuman() func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesFlush) WithErrorTrace() func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesFlush) WithFilterPath(v ...string) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesFlush) WithHeader(h map[string]string) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
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
func (f IndicesFlush) WithOpaqueID(s string) func(*IndicesFlushRequest) {
	return func(r *IndicesFlushRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
