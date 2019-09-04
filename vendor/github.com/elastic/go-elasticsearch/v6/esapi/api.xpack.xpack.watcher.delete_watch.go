// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newXPackWatcherDeleteWatchFunc(t Transport) XPackWatcherDeleteWatch {
	return func(id string, o ...func(*XPackWatcherDeleteWatchRequest)) (*Response, error) {
		var r = XPackWatcherDeleteWatchRequest{DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherDeleteWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-delete-watch.html
//
type XPackWatcherDeleteWatch func(id string, o ...func(*XPackWatcherDeleteWatchRequest)) (*Response, error)

// XPackWatcherDeleteWatchRequest configures the X Pack Watcher Delete Watch API request.
//
type XPackWatcherDeleteWatchRequest struct {
	DocumentID string

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
func (r XPackWatcherDeleteWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("watcher") + 1 + len("watch") + 1 + len(r.DocumentID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.DocumentID)

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
func (f XPackWatcherDeleteWatch) WithContext(v context.Context) func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f XPackWatcherDeleteWatch) WithMasterTimeout(v time.Duration) func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherDeleteWatch) WithPretty() func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherDeleteWatch) WithHuman() func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherDeleteWatch) WithErrorTrace() func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherDeleteWatch) WithFilterPath(v ...string) func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherDeleteWatch) WithHeader(h map[string]string) func(*XPackWatcherDeleteWatchRequest) {
	return func(r *XPackWatcherDeleteWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
