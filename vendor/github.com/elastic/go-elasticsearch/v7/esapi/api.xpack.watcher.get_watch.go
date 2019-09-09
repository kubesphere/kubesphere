// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newWatcherGetWatchFunc(t Transport) WatcherGetWatch {
	return func(id string, o ...func(*WatcherGetWatchRequest)) (*Response, error) {
		var r = WatcherGetWatchRequest{WatchID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherGetWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-get-watch.html
//
type WatcherGetWatch func(id string, o ...func(*WatcherGetWatchRequest)) (*Response, error)

// WatcherGetWatchRequest configures the Watcher Get Watch API request.
//
type WatcherGetWatchRequest struct {
	WatchID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r WatcherGetWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_watcher") + 1 + len("watch") + 1 + len(r.WatchID))
	path.WriteString("/")
	path.WriteString("_watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)

	params = make(map[string]string)

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
func (f WatcherGetWatch) WithContext(v context.Context) func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherGetWatch) WithPretty() func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherGetWatch) WithHuman() func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherGetWatch) WithErrorTrace() func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherGetWatch) WithFilterPath(v ...string) func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherGetWatch) WithHeader(h map[string]string) func(*WatcherGetWatchRequest) {
	return func(r *WatcherGetWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
