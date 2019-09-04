// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newWatcherDeleteWatchFunc(t Transport) WatcherDeleteWatch {
	return func(id string, o ...func(*WatcherDeleteWatchRequest)) (*Response, error) {
		var r = WatcherDeleteWatchRequest{WatchID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherDeleteWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-delete-watch.html
//
type WatcherDeleteWatch func(id string, o ...func(*WatcherDeleteWatchRequest)) (*Response, error)

// WatcherDeleteWatchRequest configures the Watcher Delete Watch API request.
//
type WatcherDeleteWatchRequest struct {
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
func (r WatcherDeleteWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

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
func (f WatcherDeleteWatch) WithContext(v context.Context) func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherDeleteWatch) WithPretty() func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherDeleteWatch) WithHuman() func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherDeleteWatch) WithErrorTrace() func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherDeleteWatch) WithFilterPath(v ...string) func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherDeleteWatch) WithHeader(h map[string]string) func(*WatcherDeleteWatchRequest) {
	return func(r *WatcherDeleteWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
