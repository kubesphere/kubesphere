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

func newIndicesPutSettingsFunc(t Transport) IndicesPutSettings {
	return func(body io.Reader, o ...func(*IndicesPutSettingsRequest)) (*Response, error) {
		var r = IndicesPutSettingsRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesPutSettings updates the index settings.
//
//
type IndicesPutSettings func(body io.Reader, o ...func(*IndicesPutSettingsRequest)) (*Response, error)

// IndicesPutSettingsRequest configures the Indices Put Settings API request.
//
type IndicesPutSettingsRequest struct {
	Index []string

	Body io.Reader

	AllowNoIndices    *bool
	ExpandWildcards   string
	FlatSettings      *bool
	IgnoreUnavailable *bool
	MasterTimeout     time.Duration
	PreserveExisting  *bool
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
func (r IndicesPutSettingsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_settings"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_settings")

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

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.PreserveExisting != nil {
		params["preserve_existing"] = strconv.FormatBool(*r.PreserveExisting)
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
func (f IndicesPutSettings) WithContext(v context.Context) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesPutSettings) WithIndex(v ...string) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesPutSettings) WithAllowNoIndices(v bool) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesPutSettings) WithExpandWildcards(v string) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.ExpandWildcards = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesPutSettings) WithFlatSettings(v bool) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.FlatSettings = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesPutSettings) WithIgnoreUnavailable(v bool) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesPutSettings) WithMasterTimeout(v time.Duration) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.MasterTimeout = v
	}
}

// WithPreserveExisting - whether to update existing settings. if set to `true` existing settings on an index remain unchanged, the default is `false`.
//
func (f IndicesPutSettings) WithPreserveExisting(v bool) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.PreserveExisting = &v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesPutSettings) WithTimeout(v time.Duration) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesPutSettings) WithPretty() func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesPutSettings) WithHuman() func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesPutSettings) WithErrorTrace() func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesPutSettings) WithFilterPath(v ...string) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesPutSettings) WithHeader(h map[string]string) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
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
func (f IndicesPutSettings) WithOpaqueID(s string) func(*IndicesPutSettingsRequest) {
	return func(r *IndicesPutSettingsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
