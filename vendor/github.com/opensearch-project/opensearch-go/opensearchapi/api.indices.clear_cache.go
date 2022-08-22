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

func newIndicesClearCacheFunc(t Transport) IndicesClearCache {
	return func(o ...func(*IndicesClearCacheRequest)) (*Response, error) {
		var r = IndicesClearCacheRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesClearCache clears all or specific caches for one or more indices.
//
//
type IndicesClearCache func(o ...func(*IndicesClearCacheRequest)) (*Response, error)

// IndicesClearCacheRequest configures the Indices Clear Cache API request.
//
type IndicesClearCacheRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	Fielddata         *bool
	Fields            []string
	IgnoreUnavailable *bool
	Query             *bool
	Request           *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesClearCacheRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_cache") + 1 + len("clear"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_cache")
	path.WriteString("/")
	path.WriteString("clear")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Fielddata != nil {
		params["fielddata"] = strconv.FormatBool(*r.Fielddata)
	}

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if len(r.Index) > 0 {
		params["index"] = strings.Join(r.Index, ",")
	}

	if r.Query != nil {
		params["query"] = strconv.FormatBool(*r.Query)
	}

	if r.Request != nil {
		params["request"] = strconv.FormatBool(*r.Request)
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
func (f IndicesClearCache) WithContext(v context.Context) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index name to limit the operation.
//
func (f IndicesClearCache) WithIndex(v ...string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesClearCache) WithAllowNoIndices(v bool) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesClearCache) WithExpandWildcards(v string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.ExpandWildcards = v
	}
}

// WithFielddata - clear field data.
//
func (f IndicesClearCache) WithFielddata(v bool) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Fielddata = &v
	}
}

// WithFields - a list of fields to clear when using the `fielddata` parameter (default: all).
//
func (f IndicesClearCache) WithFields(v ...string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Fields = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesClearCache) WithIgnoreUnavailable(v bool) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithQuery - clear query caches.
//
func (f IndicesClearCache) WithQuery(v bool) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Query = &v
	}
}

// WithRequest - clear request cache.
//
func (f IndicesClearCache) WithRequest(v bool) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Request = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesClearCache) WithPretty() func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesClearCache) WithHuman() func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesClearCache) WithErrorTrace() func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesClearCache) WithFilterPath(v ...string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesClearCache) WithHeader(h map[string]string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
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
func (f IndicesClearCache) WithOpaqueID(s string) func(*IndicesClearCacheRequest) {
	return func(r *IndicesClearCacheRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
