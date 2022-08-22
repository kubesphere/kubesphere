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
)

func newIngestSimulateFunc(t Transport) IngestSimulate {
	return func(body io.Reader, o ...func(*IngestSimulateRequest)) (*Response, error) {
		var r = IngestSimulateRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IngestSimulate allows to simulate a pipeline with example documents.
//
//
type IngestSimulate func(body io.Reader, o ...func(*IngestSimulateRequest)) (*Response, error)

// IngestSimulateRequest configures the Ingest Simulate API request.
//
type IngestSimulateRequest struct {
	PipelineID string

	Body io.Reader

	Verbose *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IngestSimulateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ingest") + 1 + len("pipeline") + 1 + len(r.PipelineID) + 1 + len("_simulate"))
	path.WriteString("/")
	path.WriteString("_ingest")
	path.WriteString("/")
	path.WriteString("pipeline")
	if r.PipelineID != "" {
		path.WriteString("/")
		path.WriteString(r.PipelineID)
	}
	path.WriteString("/")
	path.WriteString("_simulate")

	params = make(map[string]string)

	if r.Verbose != nil {
		params["verbose"] = strconv.FormatBool(*r.Verbose)
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
func (f IngestSimulate) WithContext(v context.Context) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.ctx = v
	}
}

// WithPipelineID - pipeline ID.
//
func (f IngestSimulate) WithPipelineID(v string) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.PipelineID = v
	}
}

// WithVerbose - verbose mode. display data output for each processor in executed pipeline.
//
func (f IngestSimulate) WithVerbose(v bool) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.Verbose = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IngestSimulate) WithPretty() func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IngestSimulate) WithHuman() func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IngestSimulate) WithErrorTrace() func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IngestSimulate) WithFilterPath(v ...string) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IngestSimulate) WithHeader(h map[string]string) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
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
func (f IngestSimulate) WithOpaqueID(s string) func(*IngestSimulateRequest) {
	return func(r *IngestSimulateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
