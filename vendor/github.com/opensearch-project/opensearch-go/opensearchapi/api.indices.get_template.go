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

func newIndicesGetTemplateFunc(t Transport) IndicesGetTemplate {
	return func(o ...func(*IndicesGetTemplateRequest)) (*Response, error) {
		var r = IndicesGetTemplateRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesGetTemplate returns an index template.
//
//
type IndicesGetTemplate func(o ...func(*IndicesGetTemplateRequest)) (*Response, error)

// IndicesGetTemplateRequest configures the Indices Get Template API request.
//
type IndicesGetTemplateRequest struct {
	Name []string

	FlatSettings    *bool
	IncludeTypeName *bool
	Local           *bool
	MasterTimeout   time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesGetTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_template") + 1 + len(strings.Join(r.Name, ",")))
	path.WriteString("/")
	path.WriteString("_template")
	if len(r.Name) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Name, ","))
	}

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
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
func (f IndicesGetTemplate) WithContext(v context.Context) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.ctx = v
	}
}

// WithName - the comma separated names of the index templates.
//
func (f IndicesGetTemplate) WithName(v ...string) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.Name = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesGetTemplate) WithFlatSettings(v bool) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.FlatSettings = &v
	}
}

// WithIncludeTypeName - whether a type should be returned in the body of the mappings..
//
func (f IndicesGetTemplate) WithIncludeTypeName(v bool) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.IncludeTypeName = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesGetTemplate) WithLocal(v bool) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IndicesGetTemplate) WithMasterTimeout(v time.Duration) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesGetTemplate) WithPretty() func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesGetTemplate) WithHuman() func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesGetTemplate) WithErrorTrace() func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesGetTemplate) WithFilterPath(v ...string) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesGetTemplate) WithHeader(h map[string]string) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
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
func (f IndicesGetTemplate) WithOpaqueID(s string) func(*IndicesGetTemplateRequest) {
	return func(r *IndicesGetTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
