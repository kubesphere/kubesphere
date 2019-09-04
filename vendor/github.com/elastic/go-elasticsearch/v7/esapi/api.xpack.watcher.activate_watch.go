// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newWatcherActivateWatchFunc(t Transport) WatcherActivateWatch {
	return func(watch_id string, o ...func(*WatcherActivateWatchRequest)) (*Response, error) {
		var r = WatcherActivateWatchRequest{WatchID: watch_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherActivateWatch - https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-activate-watch.html
//
type WatcherActivateWatch func(watch_id string, o ...func(*WatcherActivateWatchRequest)) (*Response, error)

// WatcherActivateWatchRequest configures the Watcher Activate Watch API request.
//
type WatcherActivateWatchRequest struct {
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
func (r WatcherActivateWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_watcher") + 1 + len("watch") + 1 + len(r.WatchID) + 1 + len("_activate"))
	path.WriteString("/")
	path.WriteString("_watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)
	path.WriteString("/")
	path.WriteString("_activate")

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
func (f WatcherActivateWatch) WithContext(v context.Context) func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherActivateWatch) WithPretty() func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherActivateWatch) WithHuman() func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherActivateWatch) WithErrorTrace() func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherActivateWatch) WithFilterPath(v ...string) func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherActivateWatch) WithHeader(h map[string]string) func(*WatcherActivateWatchRequest) {
	return func(r *WatcherActivateWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
