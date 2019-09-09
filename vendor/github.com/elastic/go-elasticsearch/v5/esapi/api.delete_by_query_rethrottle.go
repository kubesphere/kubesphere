// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"strconv"
	"strings"
)

func newDeleteByQueryRethrottleFunc(t Transport) DeleteByQueryRethrottle {
	return func(task_id string, requests_per_second *int, o ...func(*DeleteByQueryRethrottleRequest)) (*Response, error) {
		var r = DeleteByQueryRethrottleRequest{TaskID: task_id, RequestsPerSecond: requests_per_second}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DeleteByQueryRethrottle changes the number of requests per second for a particular Delete By Query operation.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-delete-by-query.html.
//
type DeleteByQueryRethrottle func(task_id string, requests_per_second *int, o ...func(*DeleteByQueryRethrottleRequest)) (*Response, error)

// DeleteByQueryRethrottleRequest configures the Delete By Query Rethrottle API request.
//
type DeleteByQueryRethrottleRequest struct {
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
func (r DeleteByQueryRethrottleRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_delete_by_query") + 1 + len(r.TaskID) + 1 + len("_rethrottle"))
	path.WriteString("/")
	path.WriteString("_delete_by_query")
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
func (f DeleteByQueryRethrottle) WithContext(v context.Context) func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.ctx = v
	}
}

// WithRequestsPerSecond - the throttle to set on this request in floating sub-requests per second. -1 means set no throttle..
//
func (f DeleteByQueryRethrottle) WithRequestsPerSecond(v int) func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.RequestsPerSecond = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DeleteByQueryRethrottle) WithPretty() func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DeleteByQueryRethrottle) WithHuman() func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DeleteByQueryRethrottle) WithErrorTrace() func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DeleteByQueryRethrottle) WithFilterPath(v ...string) func(*DeleteByQueryRethrottleRequest) {
	return func(r *DeleteByQueryRethrottleRequest) {
		r.FilterPath = v
	}
}
