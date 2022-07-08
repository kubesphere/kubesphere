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

func newTasksListFunc(t Transport) TasksList {
	return func(o ...func(*TasksListRequest)) (*Response, error) {
		var r = TasksListRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// TasksList returns a list of tasks.
//
// This API is experimental.
//
//
type TasksList func(o ...func(*TasksListRequest)) (*Response, error)

// TasksListRequest configures the Tasks List API request.
//
type TasksListRequest struct {
	Actions           []string
	Detailed          *bool
	GroupBy           string
	Nodes             []string
	ParentTaskID      string
	Timeout           time.Duration
	WaitForCompletion *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r TasksListRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_tasks"))
	path.WriteString("/_tasks")

	params = make(map[string]string)

	if len(r.Actions) > 0 {
		params["actions"] = strings.Join(r.Actions, ",")
	}

	if r.Detailed != nil {
		params["detailed"] = strconv.FormatBool(*r.Detailed)
	}

	if r.GroupBy != "" {
		params["group_by"] = r.GroupBy
	}

	if len(r.Nodes) > 0 {
		params["nodes"] = strings.Join(r.Nodes, ",")
	}

	if r.ParentTaskID != "" {
		params["parent_task_id"] = r.ParentTaskID
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
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
func (f TasksList) WithContext(v context.Context) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.ctx = v
	}
}

// WithActions - a list of actions that should be returned. leave empty to return all..
//
func (f TasksList) WithActions(v ...string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Actions = v
	}
}

// WithDetailed - return detailed task information (default: false).
//
func (f TasksList) WithDetailed(v bool) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Detailed = &v
	}
}

// WithGroupBy - group tasks by nodes or parent/child relationships.
//
func (f TasksList) WithGroupBy(v string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.GroupBy = v
	}
}

// WithNodes - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f TasksList) WithNodes(v ...string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Nodes = v
	}
}

// WithParentTaskID - return tasks with specified parent task ID (node_id:task_number). set to -1 to return all..
//
func (f TasksList) WithParentTaskID(v string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.ParentTaskID = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f TasksList) WithTimeout(v time.Duration) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Timeout = v
	}
}

// WithWaitForCompletion - wait for the matching tasks to complete (default: false).
//
func (f TasksList) WithWaitForCompletion(v bool) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f TasksList) WithPretty() func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f TasksList) WithHuman() func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f TasksList) WithErrorTrace() func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f TasksList) WithFilterPath(v ...string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f TasksList) WithHeader(h map[string]string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
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
func (f TasksList) WithOpaqueID(s string) func(*TasksListRequest) {
	return func(r *TasksListRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
