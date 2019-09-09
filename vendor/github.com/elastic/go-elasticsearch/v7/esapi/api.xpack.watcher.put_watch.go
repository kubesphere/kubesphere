// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newWatcherPutWatchFunc(t Transport) WatcherPutWatch {
	return func(id string, o ...func(*WatcherPutWatchRequest)) (*Response, error) {
		var r = WatcherPutWatchRequest{WatchID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// WatcherPutWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html
//
type WatcherPutWatch func(id string, o ...func(*WatcherPutWatchRequest)) (*Response, error)

// WatcherPutWatchRequest configures the Watcher Put Watch API request.
//
type WatcherPutWatchRequest struct {
	WatchID string

	Body io.Reader

	Active        *bool
	IfPrimaryTerm *int
	IfSeqNo       *int
	Version       *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r WatcherPutWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_watcher") + 1 + len("watch") + 1 + len(r.WatchID))
	path.WriteString("/")
	path.WriteString("_watcher")
	path.WriteString("/")
	path.WriteString("watch")
	path.WriteString("/")
	path.WriteString(r.WatchID)

	params = make(map[string]string)

	if r.Active != nil {
		params["active"] = strconv.FormatBool(*r.Active)
	}

	if r.IfPrimaryTerm != nil {
		params["if_primary_term"] = strconv.FormatInt(int64(*r.IfPrimaryTerm), 10)
	}

	if r.IfSeqNo != nil {
		params["if_seq_no"] = strconv.FormatInt(int64(*r.IfSeqNo), 10)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
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
func (f WatcherPutWatch) WithContext(v context.Context) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.ctx = v
	}
}

// WithBody - The watch.
//
func (f WatcherPutWatch) WithBody(v io.Reader) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.Body = v
	}
}

// WithActive - specify whether the watch is in/active by default.
//
func (f WatcherPutWatch) WithActive(v bool) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.Active = &v
	}
}

// WithIfPrimaryTerm - only update the watch if the last operation that has changed the watch has the specified primary term.
//
func (f WatcherPutWatch) WithIfPrimaryTerm(v int) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.IfPrimaryTerm = &v
	}
}

// WithIfSeqNo - only update the watch if the last operation that has changed the watch has the specified sequence number.
//
func (f WatcherPutWatch) WithIfSeqNo(v int) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.IfSeqNo = &v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f WatcherPutWatch) WithVersion(v int) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.Version = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f WatcherPutWatch) WithPretty() func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f WatcherPutWatch) WithHuman() func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f WatcherPutWatch) WithErrorTrace() func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f WatcherPutWatch) WithFilterPath(v ...string) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f WatcherPutWatch) WithHeader(h map[string]string) func(*WatcherPutWatchRequest) {
	return func(r *WatcherPutWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
