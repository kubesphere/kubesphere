// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackWatcherExecuteWatchFunc(t Transport) XPackWatcherExecuteWatch {
	return func(o ...func(*XPackWatcherExecuteWatchRequest)) (*Response, error) {
		var r = XPackWatcherExecuteWatchRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherExecuteWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-execute-watch.html
//
type XPackWatcherExecuteWatch func(o ...func(*XPackWatcherExecuteWatchRequest)) (*Response, error)

// XPackWatcherExecuteWatchRequest configures the X Pack Watcher Execute Watch API request.
//
type XPackWatcherExecuteWatchRequest struct {
	DocumentID string

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
func (r XPackWatcherExecuteWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("watcher") + 1 + len("watch") + 1 + len(r.DocumentID) + 1 + len("_execute"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("watcher")
	path.WriteString("/")
	path.WriteString("watch")
	if r.DocumentID != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentID)
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
func (f XPackWatcherExecuteWatch) WithContext(v context.Context) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.ctx = v
	}
}

// WithBody - Execution control.
//
func (f XPackWatcherExecuteWatch) WithBody(v io.Reader) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.Body = v
	}
}

// WithDocumentID - watch ID.
//
func (f XPackWatcherExecuteWatch) WithDocumentID(v string) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.DocumentID = v
	}
}

// WithDebug - indicates whether the watch should execute in debug mode.
//
func (f XPackWatcherExecuteWatch) WithDebug(v bool) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.Debug = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherExecuteWatch) WithPretty() func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherExecuteWatch) WithHuman() func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherExecuteWatch) WithErrorTrace() func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherExecuteWatch) WithFilterPath(v ...string) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherExecuteWatch) WithHeader(h map[string]string) func(*XPackWatcherExecuteWatchRequest) {
	return func(r *XPackWatcherExecuteWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
