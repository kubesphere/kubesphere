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
	"strings"
	"time"
)

func newIndicesDeleteTemplateFunc(t Transport) IndicesDeleteTemplate {
	return func(name string, o ...func(*IndicesDeleteTemplateRequest)) (*Response, error) {
		var r = IndicesDeleteTemplateRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesDeleteTemplate deletes an index template.
//
//
type IndicesDeleteTemplate func(name string, o ...func(*IndicesDeleteTemplateRequest)) (*Response, error)

// IndicesDeleteTemplateRequest configures the Indices Delete Template API request.
//
type IndicesDeleteTemplateRequest struct {
	Name string

	MasterTimeout time.Duration
	Timeout       time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesDeleteTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_template") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_template")
	path.WriteString("/")
	path.WriteString(r.Name)

	params = make(map[string]string)

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
func (f IndicesDeleteTemplate) WithContext(v context.Context) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesDeleteTemplate) WithMasterTimeout(v time.Duration) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesDeleteTemplate) WithTimeout(v time.Duration) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesDeleteTemplate) WithPretty() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesDeleteTemplate) WithHuman() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesDeleteTemplate) WithErrorTrace() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesDeleteTemplate) WithFilterPath(v ...string) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesDeleteTemplate) WithHeader(h map[string]string) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
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
func (f IndicesDeleteTemplate) WithOpaqueID(s string) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
