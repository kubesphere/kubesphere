// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newWatcherExecuteWatchFunc(t Transport) WatcherExecuteWatch {
	return func(o ...func(*WatcherExecuteWatchRequest)) (*Response, error) {
		var r = WatcherExecuteWatchRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherExecuteWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-execute-watch.html
//
type WatcherExecuteWatch func(o ...func(*WatcherExecuteWatchRequest)) (*Response, error)

// WatcherExecuteWatchRequest configures the Watcher Execute Watch API request.
//
type WatcherExecuteWatchRequest struct {
	WatchID string

	Body io.Reader

	Debug *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r WatcherExecuteWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_watcher") + 1 + len("watch") + 1 + len(r.WatchID) + 1 + len("_execute"))
	path.WriteString("/")
	path.WriteString("_watcher")
	path.WriteString("/")
	path.WriteString("watch")
	if r.WatchID != "" {
		path.WriteString("/")
		path.WriteString(r.WatchID)
	}
	path.WriteString("/")
	path.WriteString("_execute")

	params = make(map[string]string)

	if r.Debug != nil {
		params["debug"] = strconv.FormatBool(*r.Debug)
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

	req, _ := newRequest(method, path.String(), r.Body)

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
func (f WatcherExecuteWatch) WithContext(v context.Context) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.ctx = v
	}
}

// WithBody - Execution control.
//
func (f WatcherExecuteWatch) WithBody(v io.Reader) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.Body = v
	}
}

// WithWatchID - watch ID.
//
func (f WatcherExecuteWatch) WithWatchID(v string) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.WatchID = v
	}
}

// WithDebug - indicates whether the watch should execute in debug mode.
//
func (f WatcherExecuteWatch) WithDebug(v bool) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.Debug = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherExecuteWatch) WithPretty() func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherExecuteWatch) WithHuman() func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherExecuteWatch) WithErrorTrace() func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherExecuteWatch) WithFilterPath(v ...string) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherExecuteWatch) WithHeader(h map[string]string) func(*WatcherExecuteWatchRequest) {
	return func(r *WatcherExecuteWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
