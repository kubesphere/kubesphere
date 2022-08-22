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

func newIndicesPutMappingFunc(t Transport) IndicesPutMapping {
	return func(body io.Reader, o ...func(*IndicesPutMappingRequest)) (*Response, error) {
		var r = IndicesPutMappingRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesPutMapping updates the index mappings.
//
//
type IndicesPutMapping func(body io.Reader, o ...func(*IndicesPutMappingRequest)) (*Response, error)

// IndicesPutMappingRequest configures the Indices Put Mapping API request.
//
type IndicesPutMappingRequest struct {
	Index        []string
	DocumentType string

	Body io.Reader

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
	IncludeTypeName   *bool
	MasterTimeout     time.Duration
	Timeout           time.Duration
	WriteIndexOnly    *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesPutMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len(strings.Join(r.Index, ",")) + len("/_mapping") + len(r.DocumentType) + 2)
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_mapping")
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
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

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.WriteIndexOnly != nil {
		params["write_index_only"] = strconv.FormatBool(*r.WriteIndexOnly)
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
func (f IndicesPutMapping) WithContext(v context.Context) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names the mapping should be added to (supports wildcards); use `_all` or omit to add the mapping on all indices..
//
func (f IndicesPutMapping) WithIndex(v ...string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.Index = v
	}
}

// WithDocumentType - the name of the document type.
//
func (f IndicesPutMapping) WithDocumentType(v string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesPutMapping) WithAllowNoIndices(v bool) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesPutMapping) WithExpandWildcards(v string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesPutMapping) WithIgnoreUnavailable(v bool) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithIncludeTypeName - whether a type should be expected in the body of the mappings..
//
func (f IndicesPutMapping) WithIncludeTypeName(v bool) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.IncludeTypeName = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesPutMapping) WithMasterTimeout(v time.Duration) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesPutMapping) WithTimeout(v time.Duration) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.Timeout = v
	}
}

// WithWriteIndexOnly - when true, applies mappings only to the write index of an alias or data stream.
//
func (f IndicesPutMapping) WithWriteIndexOnly(v bool) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.WriteIndexOnly = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesPutMapping) WithPretty() func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesPutMapping) WithHuman() func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesPutMapping) WithErrorTrace() func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesPutMapping) WithFilterPath(v ...string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesPutMapping) WithHeader(h map[string]string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
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
func (f IndicesPutMapping) WithOpaqueID(s string) func(*IndicesPutMappingRequest) {
	return func(r *IndicesPutMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
