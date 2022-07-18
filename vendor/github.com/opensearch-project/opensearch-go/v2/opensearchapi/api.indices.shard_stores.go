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

func newIndicesShardStoresFunc(t Transport) IndicesShardStores {
	return func(o ...func(*IndicesShardStoresRequest)) (*Response, error) {
		var r = IndicesShardStoresRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesShardStores provides store information for shard copies of indices.
//
//
type IndicesShardStores func(o ...func(*IndicesShardStoresRequest)) (*Response, error)

// IndicesShardStoresRequest configures the Indices Shard Stores API request.
//
type IndicesShardStoresRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
	Status            []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesShardStoresRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_shard_stores"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_shard_stores")

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

	if len(r.Status) > 0 {
		params["status"] = strings.Join(r.Status, ",")
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
func (f IndicesShardStores) WithContext(v context.Context) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesShardStores) WithIndex(v ...string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesShardStores) WithAllowNoIndices(v bool) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesShardStores) WithExpandWildcards(v string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesShardStores) WithIgnoreUnavailable(v bool) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithStatus - a list of statuses used to filter on shards to get store information for.
//
func (f IndicesShardStores) WithStatus(v ...string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.Status = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesShardStores) WithPretty() func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesShardStores) WithHuman() func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesShardStores) WithErrorTrace() func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesShardStores) WithFilterPath(v ...string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesShardStores) WithHeader(h map[string]string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
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
func (f IndicesShardStores) WithOpaqueID(s string) func(*IndicesShardStoresRequest) {
	return func(r *IndicesShardStoresRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
