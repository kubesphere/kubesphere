// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newTasksGetFunc(t Transport) TasksGet {
	return func(task_id string, o ...func(*TasksGetRequest)) (*Response, error) {
		var r = TasksGetRequest{TaskID: task_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// TasksGet returns information about a task.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/tasks.html.
//
type TasksGet func(task_id string, o ...func(*TasksGetRequest)) (*Response, error)

// TasksGetRequest configures the Tasks Get API request.
//
type TasksGetRequest struct {
	TaskID string

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
func (r TasksGetRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_tasks") + 1 + len(r.TaskID))
	path.WriteString("/")
	path.WriteString("_tasks")
	path.WriteString("/")
	path.WriteString(r.TaskID)

	params = make(map[string]string)

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

	req, _ := newRequest(method, path.String(), nil)

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
func (f TasksGet) WithContext(v context.Context) func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.ctx = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f TasksGet) WithTimeout(v time.Duration) func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.Timeout = v
	}
}

// WithWaitForCompletion - wait for the matching tasks to complete (default: false).
//
func (f TasksGet) WithWaitForCompletion(v bool) func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f TasksGet) WithPretty() func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f TasksGet) WithHuman() func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f TasksGet) WithErrorTrace() func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f TasksGet) WithFilterPath(v ...string) func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f TasksGet) WithHeader(h map[string]string) func(*TasksGetRequest) {
	return func(r *TasksGetRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
