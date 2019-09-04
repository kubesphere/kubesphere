// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newXPackWatcherDeactivateWatchFunc(t Transport) XPackWatcherDeactivateWatch {
	return func(watch_id string, o ...func(*XPackWatcherDeactivateWatchRequest)) (*Response, error) {
		var r = XPackWatcherDeactivateWatchRequest{WatchID: watch_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherDeactivateWatch - https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-deactivate-watch.html
//
type XPackWatcherDeactivateWatch func(watch_id string, o ...func(*XPackWatcherDeactivateWatchRequest)) (*Response, error)

// XPackWatcherDeactivateWatchRequest configures the X Pack Watcher Deactivate Watch API request.
//
type XPackWatcherDeactivateWatchRequest struct {
	WatchID string

	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackWatcherDeactivateWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("watcher") + 1 + len("watch") + 1 + len(r.WatchID) + 1 + len("_deactivate"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)
	path.WriteString("/")
	path.WriteString("_deactivate")

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f XPackWatcherDeactivateWatch) WithContext(v context.Context) func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f XPackWatcherDeactivateWatch) WithMasterTimeout(v time.Duration) func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherDeactivateWatch) WithPretty() func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherDeactivateWatch) WithHuman() func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherDeactivateWatch) WithErrorTrace() func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherDeactivateWatch) WithFilterPath(v ...string) func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherDeactivateWatch) WithHeader(h map[string]string) func(*XPackWatcherDeactivateWatchRequest) {
	return func(r *XPackWatcherDeactivateWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
