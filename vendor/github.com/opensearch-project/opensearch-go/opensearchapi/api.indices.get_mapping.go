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

func newIndicesGetMappingFunc(t Transport) IndicesGetMapping {
	return func(o ...func(*IndicesGetMappingRequest)) (*Response, error) {
		var r = IndicesGetMappingRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesGetMapping returns mappings for one or more indices.
//
//
type IndicesGetMapping func(o ...func(*IndicesGetMappingRequest)) (*Response, error)

// IndicesGetMappingRequest configures the Indices Get Mapping API request.
//
type IndicesGetMappingRequest struct {
	Index        []string
	DocumentType []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
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
func (r IndicesGetMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_mapping") + 1 + len(strings.Join(r.DocumentType, ",")))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_mapping")
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
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
func (f IndicesGetMapping) WithContext(v context.Context) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names.
//
func (f IndicesGetMapping) WithIndex(v ...string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types.
//
func (f IndicesGetMapping) WithDocumentType(v ...string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesGetMapping) WithAllowNoIndices(v bool) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesGetMapping) WithExpandWildcards(v string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesGetMapping) WithIgnoreUnavailable(v bool) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithIncludeTypeName - whether to add the type name to the response (default: false).
//
func (f IndicesGetMapping) WithIncludeTypeName(v bool) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.IncludeTypeName = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesGetMapping) WithLocal(v bool) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesGetMapping) WithMasterTimeout(v time.Duration) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesGetMapping) WithPretty() func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesGetMapping) WithHuman() func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesGetMapping) WithErrorTrace() func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesGetMapping) WithFilterPath(v ...string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesGetMapping) WithHeader(h map[string]string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
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
func (f IndicesGetMapping) WithOpaqueID(s string) func(*IndicesGetMappingRequest) {
	return func(r *IndicesGetMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
