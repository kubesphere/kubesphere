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

func newIndicesExistsFunc(t Transport) IndicesExists {
	return func(index []string, o ...func(*IndicesExistsRequest)) (*Response, error) {
		var r = IndicesExistsRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesExists returns information about whether a particular index exists.
//
//
type IndicesExists func(index []string, o ...func(*IndicesExistsRequest)) (*Response, error)

// IndicesExistsRequest configures the Indices Exists API request.
//
type IndicesExistsRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	FlatSettings      *bool
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	Local             *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesExistsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "HEAD"

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

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.IncludeDefaults != nil {
		params["include_defaults"] = strconv.FormatBool(*r.IncludeDefaults)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
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
func (f IndicesExists) WithContext(v context.Context) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - ignore if a wildcard expression resolves to no concrete indices (default: false).
//
func (f IndicesExists) WithAllowNoIndices(v bool) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether wildcard expressions should get expanded to open or closed indices (default: open).
//
func (f IndicesExists) WithExpandWildcards(v string) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.ExpandWildcards = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesExists) WithFlatSettings(v bool) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.FlatSettings = &v
	}
}

// WithIgnoreUnavailable - ignore unavailable indexes (default: false).
//
func (f IndicesExists) WithIgnoreUnavailable(v bool) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithIncludeDefaults - whether to return all default setting for each of the indices..
//
func (f IndicesExists) WithIncludeDefaults(v bool) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.IncludeDefaults = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesExists) WithLocal(v bool) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.Local = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesExists) WithPretty() func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesExists) WithHuman() func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesExists) WithErrorTrace() func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesExists) WithFilterPath(v ...string) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesExists) WithHeader(h map[string]string) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
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
func (f IndicesExists) WithOpaqueID(s string) func(*IndicesExistsRequest) {
	return func(r *IndicesExistsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
