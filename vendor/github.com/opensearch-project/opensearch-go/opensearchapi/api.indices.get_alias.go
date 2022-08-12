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

func newIndicesGetAliasFunc(t Transport) IndicesGetAlias {
	return func(o ...func(*IndicesGetAliasRequest)) (*Response, error) {
		var r = IndicesGetAliasRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesGetAlias returns an alias.
//
//
type IndicesGetAlias func(o ...func(*IndicesGetAliasRequest)) (*Response, error)

// IndicesGetAliasRequest configures the Indices Get Alias API request.
//
type IndicesGetAliasRequest struct {
	Index []string

	Name []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
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
func (r IndicesGetAliasRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_alias") + 1 + len(strings.Join(r.Name, ",")))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_alias")
	if len(r.Name) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Name, ","))
	}

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
func (f IndicesGetAlias) WithContext(v context.Context) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to filter aliases.
//
func (f IndicesGetAlias) WithIndex(v ...string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.Index = v
	}
}

// WithName - a list of alias names to return.
//
func (f IndicesGetAlias) WithName(v ...string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.Name = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesGetAlias) WithAllowNoIndices(v bool) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesGetAlias) WithExpandWildcards(v string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesGetAlias) WithIgnoreUnavailable(v bool) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesGetAlias) WithLocal(v bool) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.Local = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesGetAlias) WithPretty() func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesGetAlias) WithHuman() func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesGetAlias) WithErrorTrace() func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesGetAlias) WithFilterPath(v ...string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesGetAlias) WithHeader(h map[string]string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
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
func (f IndicesGetAlias) WithOpaqueID(s string) func(*IndicesGetAliasRequest) {
	return func(r *IndicesGetAliasRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
