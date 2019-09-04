// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newTasksCancelFunc(t Transport) TasksCancel {
	return func(o ...func(*TasksCancelRequest)) (*Response, error) {
		var r = TasksCancelRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// TasksCancel cancels a task, if it can be cancelled through an API.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/tasks.html.
//
type TasksCancel func(o ...func(*TasksCancelRequest)) (*Response, error)

// TasksCancelRequest configures the Tasks Cancel API request.
//
type TasksCancelRequest struct {
	TaskID string

	Actions      []string
	Nodes        []string
	ParentTaskID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r TasksCancelRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_tasks") + 1 + len(r.TaskID) + 1 + len("_cancel"))
	path.WriteString("/")
	path.WriteString("_tasks")
	if r.TaskID != "" {
		path.WriteString("/")
		path.WriteString(r.TaskID)
	}
	path.WriteString("/")
	path.WriteString("_cancel")

	params = make(map[string]string)

	if len(r.Actions) > 0 {
		params["actions"] = strings.Join(r.Actions, ",")
	}

	if len(r.Nodes) > 0 {
		params["nodes"] = strings.Join(r.Nodes, ",")
	}

	if r.ParentTaskID != "" {
		params["parent_task_id"] = r.ParentTaskID
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
func (f TasksCancel) WithContext(v context.Context) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.ctx = v
	}
}

// WithTaskID - cancel the task with specified task ID (node_id:task_number).
//
func (f TasksCancel) WithTaskID(v string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.TaskID = v
	}
}

// WithActions - a list of actions that should be cancelled. leave empty to cancel all..
//
func (f TasksCancel) WithActions(v ...string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.Actions = v
	}
}

// WithNodes - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f TasksCancel) WithNodes(v ...string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.Nodes = v
	}
}

// WithParentTaskID - cancel tasks with specified parent task ID (node_id:task_number). set to -1 to cancel all..
//
func (f TasksCancel) WithParentTaskID(v string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.ParentTaskID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f TasksCancel) WithPretty() func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f TasksCancel) WithHuman() func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f TasksCancel) WithErrorTrace() func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f TasksCancel) WithFilterPath(v ...string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f TasksCancel) WithHeader(h map[string]string) func(*TasksCancelRequest) {
	return func(r *TasksCancelRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
