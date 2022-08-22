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

func newIndicesExistsIndexTemplateFunc(t Transport) IndicesExistsIndexTemplate {
	return func(name string, o ...func(*IndicesExistsIndexTemplateRequest)) (*Response, error) {
		var r = IndicesExistsIndexTemplateRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesExistsIndexTemplate returns information about whether a particular index template exists.
//
//
type IndicesExistsIndexTemplate func(name string, o ...func(*IndicesExistsIndexTemplateRequest)) (*Response, error)

// IndicesExistsIndexTemplateRequest configures the Indices Exists Index Template API request.
//
type IndicesExistsIndexTemplateRequest struct {
	Name string

	FlatSettings  *bool
	Local         *bool
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
func (r IndicesExistsIndexTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "HEAD"

	path.Grow(1 + len("_index_template") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_index_template")
	path.WriteString("/")
	path.WriteString(r.Name)

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
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
func (f IndicesExistsIndexTemplate) WithContext(v context.Context) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.ctx = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesExistsIndexTemplate) WithFlatSettings(v bool) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.FlatSettings = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesExistsIndexTemplate) WithLocal(v bool) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IndicesExistsIndexTemplate) WithMasterTimeout(v time.Duration) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesExistsIndexTemplate) WithPretty() func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesExistsIndexTemplate) WithHuman() func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesExistsIndexTemplate) WithErrorTrace() func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesExistsIndexTemplate) WithFilterPath(v ...string) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesExistsIndexTemplate) WithHeader(h map[string]string) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
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
func (f IndicesExistsIndexTemplate) WithOpaqueID(s string) func(*IndicesExistsIndexTemplateRequest) {
	return func(r *IndicesExistsIndexTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
