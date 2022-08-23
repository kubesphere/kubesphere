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

func newScrollFunc(t Transport) Scroll {
	return func(o ...func(*ScrollRequest)) (*Response, error) {
		var r = ScrollRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Scroll allows to retrieve a large numbers of results from a single search request.
//
//
type Scroll func(o ...func(*ScrollRequest)) (*Response, error)

// ScrollRequest configures the Scroll API request.
//
type ScrollRequest struct {
	Body io.Reader

	ScrollID string

	RestTotalHitsAsInt *bool
	Scroll             time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ScrollRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_search/scroll"))
	path.WriteString("/_search/scroll")

	params = make(map[string]string)

	if r.RestTotalHitsAsInt != nil {
		params["rest_total_hits_as_int"] = strconv.FormatBool(*r.RestTotalHitsAsInt)
	}

	if r.Scroll != 0 {
		params["scroll"] = formatDuration(r.Scroll)
	}

	if r.ScrollID != "" {
		params["scroll_id"] = r.ScrollID
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
func (f Scroll) WithContext(v context.Context) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.ctx = v
	}
}

// WithBody - The scroll ID if not passed by URL or query parameter..
//
func (f Scroll) WithBody(v io.Reader) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.Body = v
	}
}

// WithScrollID - the scroll ID.
//
func (f Scroll) WithScrollID(v string) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.ScrollID = v
	}
}

// WithRestTotalHitsAsInt - indicates whether hits.total should be rendered as an integer or an object in the rest search response.
//
func (f Scroll) WithRestTotalHitsAsInt(v bool) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.RestTotalHitsAsInt = &v
	}
}

// WithScroll - specify how long a consistent view of the index should be maintained for scrolled search.
//
func (f Scroll) WithScroll(v time.Duration) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.Scroll = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Scroll) WithPretty() func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Scroll) WithHuman() func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Scroll) WithErrorTrace() func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Scroll) WithFilterPath(v ...string) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Scroll) WithHeader(h map[string]string) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
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
func (f Scroll) WithOpaqueID(s string) func(*ScrollRequest) {
	return func(r *ScrollRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
