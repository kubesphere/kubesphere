// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newWatcherDeactivateWatchFunc(t Transport) WatcherDeactivateWatch {
	return func(watch_id string, o ...func(*WatcherDeactivateWatchRequest)) (*Response, error) {
		var r = WatcherDeactivateWatchRequest{WatchID: watch_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherDeactivateWatch - https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-deactivate-watch.html
//
type WatcherDeactivateWatch func(watch_id string, o ...func(*WatcherDeactivateWatchRequest)) (*Response, error)

// WatcherDeactivateWatchRequest configures the Watcher Deactivate Watch API request.
//
type WatcherDeactivateWatchRequest struct {
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
func (r WatcherDeactivateWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_watcher") + 1 + len("watch") + 1 + len(r.WatchID) + 1 + len("_deactivate"))
	path.WriteString("/")
	path.WriteString("_watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)
	path.WriteString("/")
	path.WriteString("_deactivate")

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
func (f WatcherDeactivateWatch) WithContext(v context.Context) func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherDeactivateWatch) WithPretty() func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherDeactivateWatch) WithHuman() func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherDeactivateWatch) WithErrorTrace() func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherDeactivateWatch) WithFilterPath(v ...string) func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherDeactivateWatch) WithHeader(h map[string]string) func(*WatcherDeactivateWatchRequest) {
	return func(r *WatcherDeactivateWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
