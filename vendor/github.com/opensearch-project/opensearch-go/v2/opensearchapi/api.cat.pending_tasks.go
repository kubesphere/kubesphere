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

func newCatPendingTasksFunc(t Transport) CatPendingTasks {
	return func(o ...func(*CatPendingTasksRequest)) (*Response, error) {
		var r = CatPendingTasksRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatPendingTasks returns a concise representation of the cluster pending tasks.
//
//
type CatPendingTasks func(o ...func(*CatPendingTasksRequest)) (*Response, error)

// CatPendingTasksRequest configures the Cat Pending Tasks API request.
//
type CatPendingTasksRequest struct {
	Format        string
	H             []string
	Help          *bool
	Local         *bool
	MasterTimeout time.Duration
	S             []string
	Time          string
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
func (r CatPendingTasksRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cat/pending_tasks"))
	path.WriteString("/_cat/pending_tasks")

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

	if r.Time != "" {
		params["time"] = r.Time
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
func (f CatPendingTasks) WithContext(v context.Context) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.ctx = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatPendingTasks) WithFormat(v string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatPendingTasks) WithH(v ...string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatPendingTasks) WithHelp(v bool) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatPendingTasks) WithLocal(v bool) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatPendingTasks) WithMasterTimeout(v time.Duration) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatPendingTasks) WithS(v ...string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.S = v
	}
}

// WithTime - the unit in which to display time values.
//
func (f CatPendingTasks) WithTime(v string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Time = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatPendingTasks) WithV(v bool) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatPendingTasks) WithPretty() func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatPendingTasks) WithHuman() func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatPendingTasks) WithErrorTrace() func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatPendingTasks) WithFilterPath(v ...string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatPendingTasks) WithHeader(h map[string]string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
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
func (f CatPendingTasks) WithOpaqueID(s string) func(*CatPendingTasksRequest) {
	return func(r *CatPendingTasksRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
