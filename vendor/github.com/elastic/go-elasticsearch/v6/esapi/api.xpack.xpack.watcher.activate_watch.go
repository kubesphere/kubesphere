// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newXPackWatcherActivateWatchFunc(t Transport) XPackWatcherActivateWatch {
	return func(watch_id string, o ...func(*XPackWatcherActivateWatchRequest)) (*Response, error) {
		var r = XPackWatcherActivateWatchRequest{WatchID: watch_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherActivateWatch - https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-activate-watch.html
//
type XPackWatcherActivateWatch func(watch_id string, o ...func(*XPackWatcherActivateWatchRequest)) (*Response, error)

// XPackWatcherActivateWatchRequest configures the X Pack Watcher Activate Watch API request.
//
type XPackWatcherActivateWatchRequest struct {
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
func (r XPackWatcherActivateWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("watcher") + 1 + len("watch") + 1 + len(r.WatchID) + 1 + len("_activate"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)
	path.WriteString("/")
	path.WriteString("_activate")

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
func (f XPackWatcherActivateWatch) WithContext(v context.Context) func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f XPackWatcherActivateWatch) WithMasterTimeout(v time.Duration) func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherActivateWatch) WithPretty() func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherActivateWatch) WithHuman() func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherActivateWatch) WithErrorTrace() func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherActivateWatch) WithFilterPath(v ...string) func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherActivateWatch) WithHeader(h map[string]string) func(*XPackWatcherActivateWatchRequest) {
	return func(r *XPackWatcherActivateWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
