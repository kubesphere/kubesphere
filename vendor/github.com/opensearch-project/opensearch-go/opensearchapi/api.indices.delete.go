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

func newIndicesDeleteFunc(t Transport) IndicesDelete {
	return func(index []string, o ...func(*IndicesDeleteRequest)) (*Response, error) {
		var r = IndicesDeleteRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesDelete deletes an index.
//
//
type IndicesDelete func(index []string, o ...func(*IndicesDeleteRequest)) (*Response, error)

// IndicesDeleteRequest configures the Indices Delete API request.
//
type IndicesDeleteRequest struct {
	Index []string

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
func (r IndicesDeleteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))

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
func (f IndicesDelete) WithContext(v context.Context) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - ignore if a wildcard expression resolves to no concrete indices (default: false).
//
func (f IndicesDelete) WithAllowNoIndices(v bool) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether wildcard expressions should get expanded to open or closed indices (default: open).
//
func (f IndicesDelete) WithExpandWildcards(v string) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - ignore unavailable indexes (default: false).
//
func (f IndicesDelete) WithIgnoreUnavailable(v bool) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesDelete) WithMasterTimeout(v time.Duration) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesDelete) WithTimeout(v time.Duration) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesDelete) WithPretty() func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesDelete) WithHuman() func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesDelete) WithErrorTrace() func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesDelete) WithFilterPath(v ...string) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesDelete) WithHeader(h map[string]string) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
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
func (f IndicesDelete) WithOpaqueID(s string) func(*IndicesDeleteRequest) {
	return func(r *IndicesDeleteRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
