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

func newMtermvectorsFunc(t Transport) Mtermvectors {
	return func(o ...func(*MtermvectorsRequest)) (*Response, error) {
		var r = MtermvectorsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Mtermvectors returns multiple termvectors in one request.
//
//
type Mtermvectors func(o ...func(*MtermvectorsRequest)) (*Response, error)

// MtermvectorsRequest configures the Mtermvectors API request.
//
type MtermvectorsRequest struct {
	Index        string
	DocumentType string

	Body io.Reader

	Fields          []string
	FieldStatistics *bool
	Ids             []string
	Offsets         *bool
	Payloads        *bool
	Positions       *bool
	Preference      string
	Realtime        *bool
	Routing         string
	TermStatistics  *bool
	Version         *int
	VersionType     string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MtermvectorsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len("_mtermvectors"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString("_mtermvectors")

	params = make(map[string]string)

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if r.FieldStatistics != nil {
		params["field_statistics"] = strconv.FormatBool(*r.FieldStatistics)
	}

	if len(r.Ids) > 0 {
		params["ids"] = strings.Join(r.Ids, ",")
	}

	if r.Offsets != nil {
		params["offsets"] = strconv.FormatBool(*r.Offsets)
	}

	if r.Payloads != nil {
		params["payloads"] = strconv.FormatBool(*r.Payloads)
	}

	if r.Positions != nil {
		params["positions"] = strconv.FormatBool(*r.Positions)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Realtime != nil {
		params["realtime"] = strconv.FormatBool(*r.Realtime)
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
	}

	if r.TermStatistics != nil {
		params["term_statistics"] = strconv.FormatBool(*r.TermStatistics)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
	}

	if r.VersionType != "" {
		params["version_type"] = r.VersionType
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
func (f Mtermvectors) WithContext(v context.Context) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.ctx = v
	}
}

// WithBody - Define ids, documents, parameters or a list of parameters per document here. You must at least provide a list of document ids. See documentation..
//
func (f Mtermvectors) WithBody(v io.Reader) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Body = v
	}
}

// WithIndex - the index in which the document resides..
//
func (f Mtermvectors) WithIndex(v string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Index = v
	}
}

// WithDocumentType - the type of the document..
//
func (f Mtermvectors) WithDocumentType(v string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.DocumentType = v
	}
}

// WithFields - a list of fields to return. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithFields(v ...string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Fields = v
	}
}

// WithFieldStatistics - specifies if document count, sum of document frequencies and sum of total term frequencies should be returned. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithFieldStatistics(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.FieldStatistics = &v
	}
}

// WithIds - a list of documents ids. you must define ids as parameter or set "ids" or "docs" in the request body.
//
func (f Mtermvectors) WithIds(v ...string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Ids = v
	}
}

// WithOffsets - specifies if term offsets should be returned. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithOffsets(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Offsets = &v
	}
}

// WithPayloads - specifies if term payloads should be returned. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithPayloads(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Payloads = &v
	}
}

// WithPositions - specifies if term positions should be returned. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithPositions(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Positions = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random) .applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithPreference(v string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Preference = v
	}
}

// WithRealtime - specifies if requests are real-time as opposed to near-real-time (default: true)..
//
func (f Mtermvectors) WithRealtime(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Realtime = &v
	}
}

// WithRouting - specific routing value. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithRouting(v string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Routing = v
	}
}

// WithTermStatistics - specifies if total term frequency and document frequency should be returned. applies to all returned documents unless otherwise specified in body "params" or "docs"..
//
func (f Mtermvectors) WithTermStatistics(v bool) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.TermStatistics = &v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f Mtermvectors) WithVersion(v int) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Version = &v
	}
}

// WithVersionType - specific version type.
//
func (f Mtermvectors) WithVersionType(v string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.VersionType = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Mtermvectors) WithPretty() func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Mtermvectors) WithHuman() func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Mtermvectors) WithErrorTrace() func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Mtermvectors) WithFilterPath(v ...string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Mtermvectors) WithHeader(h map[string]string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
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
func (f Mtermvectors) WithOpaqueID(s string) func(*MtermvectorsRequest) {
	return func(r *MtermvectorsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
