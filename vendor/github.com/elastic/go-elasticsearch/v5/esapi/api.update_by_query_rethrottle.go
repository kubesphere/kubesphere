// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"strconv"
	"strings"
)

func newUpdateByQueryRethrottleFunc(t Transport) UpdateByQueryRethrottle {
	return func(task_id string, requests_per_second *int, o ...func(*UpdateByQueryRethrottleRequest)) (*Response, error) {
		var r = UpdateByQueryRethrottleRequest{TaskID: task_id, RequestsPerSecond: requests_per_second}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// UpdateByQueryRethrottle changes the number of requests per second for a particular Update By Query operation.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update-by-query.html.
//
type UpdateByQueryRethrottle func(task_id string, requests_per_second *int, o ...func(*UpdateByQueryRethrottleRequest)) (*Response, error)

// UpdateByQueryRethrottleRequest configures the Update By Query Rethrottle API request.
//
type UpdateByQueryRethrottleRequest struct {
	TaskID            string
	RequestsPerSecond *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r UpdateByQueryRethrottleRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_update_by_query") + 1 + len(r.TaskID) + 1 + len("_rethrottle"))
	path.WriteString("/")
	path.WriteString("_update_by_query")
	path.WriteString("/")
	path.WriteString(r.TaskID)
	path.WriteString("/")
	path.WriteString("_rethrottle")

	params = make(map[string]string)

	if r.RequestsPerSecond != nil {
		params["requests_per_second"] = strconv.FormatInt(int64(*r.RequestsPerSecond), 10)
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

	req, _ := newRequest(method, path.String(), nil)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
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
func (f UpdateByQueryRethrottle) WithContext(v context.Context) func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.ctx = v
	}
}

// WithRequestsPerSecond - the throttle to set on this request in floating sub-requests per second. -1 means set no throttle..
//
func (f UpdateByQueryRethrottle) WithRequestsPerSecond(v int) func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.RequestsPerSecond = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f UpdateByQueryRethrottle) WithPretty() func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f UpdateByQueryRethrottle) WithHuman() func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f UpdateByQueryRethrottle) WithErrorTrace() func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f UpdateByQueryRethrottle) WithFilterPath(v ...string) func(*UpdateByQueryRethrottleRequest) {
	return func(r *UpdateByQueryRethrottleRequest) {
		r.FilterPath = v
	}
}
