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

func newGetScriptFunc(t Transport) GetScript {
	return func(id string, o ...func(*GetScriptRequest)) (*Response, error) {
		var r = GetScriptRequest{ScriptID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// GetScript returns a script.
//
//
type GetScript func(id string, o ...func(*GetScriptRequest)) (*Response, error)

// GetScriptRequest configures the Get Script API request.
//
type GetScriptRequest struct {
	ScriptID string

	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r GetScriptRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_scripts") + 1 + len(r.ScriptID))
	path.WriteString("/")
	path.WriteString("_scripts")
	path.WriteString("/")
	path.WriteString(r.ScriptID)

	params = make(map[string]string)

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
func (f GetScript) WithContext(v context.Context) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f GetScript) WithMasterTimeout(v time.Duration) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f GetScript) WithPretty() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f GetScript) WithHuman() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f GetScript) WithErrorTrace() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f GetScript) WithFilterPath(v ...string) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f GetScript) WithHeader(h map[string]string) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
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
func (f GetScript) WithOpaqueID(s string) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
