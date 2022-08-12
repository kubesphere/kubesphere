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

func newCatTemplatesFunc(t Transport) CatTemplates {
	return func(o ...func(*CatTemplatesRequest)) (*Response, error) {
		var r = CatTemplatesRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatTemplates returns information about existing templates.
//
//
type CatTemplates func(o ...func(*CatTemplatesRequest)) (*Response, error)

// CatTemplatesRequest configures the Cat Templates API request.
//
type CatTemplatesRequest struct {
	Name string

	Format        string
	H             []string
	Help          *bool
	Local         *bool
	MasterTimeout time.Duration
	S             []string
	V             *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatTemplatesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("templates") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("templates")
	if r.Name != "" {
		path.WriteString("/")
		path.WriteString(r.Name)
	}

	params = make(map[string]string)

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
	}

	if r.V != nil {
		params["v"] = strconv.FormatBool(*r.V)
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
func (f CatTemplates) WithContext(v context.Context) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.ctx = v
	}
}

// WithName - a pattern that returned template names must match.
//
func (f CatTemplates) WithName(v string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Name = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatTemplates) WithFormat(v string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatTemplates) WithH(v ...string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatTemplates) WithHelp(v bool) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatTemplates) WithLocal(v bool) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatTemplates) WithMasterTimeout(v time.Duration) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatTemplates) WithS(v ...string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatTemplates) WithV(v bool) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatTemplates) WithPretty() func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatTemplates) WithHuman() func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatTemplates) WithErrorTrace() func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatTemplates) WithFilterPath(v ...string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatTemplates) WithHeader(h map[string]string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
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
func (f CatTemplates) WithOpaqueID(s string) func(*CatTemplatesRequest) {
	return func(r *CatTemplatesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
