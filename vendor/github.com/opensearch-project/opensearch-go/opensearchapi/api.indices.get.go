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

func newIndicesGetFunc(t Transport) IndicesGet {
	return func(index []string, o ...func(*IndicesGetRequest)) (*Response, error) {
		var r = IndicesGetRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesGet returns information about one or more indices.
//
//
type IndicesGet func(index []string, o ...func(*IndicesGetRequest)) (*Response, error)

// IndicesGetRequest configures the Indices Get API request.
//
type IndicesGetRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	FlatSettings      *bool
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	IncludeTypeName   *bool
	Local             *bool
	MasterTimeout     time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesGetRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

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

	if r.IncludeTypeName != nil {
		params["include_type_name"] = strconv.FormatBool(*r.IncludeTypeName)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f IndicesGet) WithContext(v context.Context) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - ignore if a wildcard expression resolves to no concrete indices (default: false).
//
func (f IndicesGet) WithAllowNoIndices(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether wildcard expressions should get expanded to open or closed indices (default: open).
//
func (f IndicesGet) WithExpandWildcards(v string) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.ExpandWildcards = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesGet) WithFlatSettings(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.FlatSettings = &v
	}
}

// WithIgnoreUnavailable - ignore unavailable indexes (default: false).
//
func (f IndicesGet) WithIgnoreUnavailable(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithIncludeDefaults - whether to return all default setting for each of the indices..
//
func (f IndicesGet) WithIncludeDefaults(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.IncludeDefaults = &v
	}
}

// WithIncludeTypeName - whether to add the type name to the response (default: false).
//
func (f IndicesGet) WithIncludeTypeName(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.IncludeTypeName = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesGet) WithLocal(v bool) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesGet) WithMasterTimeout(v time.Duration) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesGet) WithPretty() func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesGet) WithHuman() func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesGet) WithErrorTrace() func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesGet) WithFilterPath(v ...string) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesGet) WithHeader(h map[string]string) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
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
func (f IndicesGet) WithOpaqueID(s string) func(*IndicesGetRequest) {
	return func(r *IndicesGetRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
