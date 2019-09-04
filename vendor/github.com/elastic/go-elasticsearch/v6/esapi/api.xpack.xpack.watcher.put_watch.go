// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackWatcherPutWatchFunc(t Transport) XPackWatcherPutWatch {
	return func(id string, o ...func(*XPackWatcherPutWatchRequest)) (*Response, error) {
		var r = XPackWatcherPutWatchRequest{DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherPutWatch - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html
//
type XPackWatcherPutWatch func(id string, o ...func(*XPackWatcherPutWatchRequest)) (*Response, error)

// XPackWatcherPutWatchRequest configures the X Pack Watcher Put Watch API request.
//
type XPackWatcherPutWatchRequest struct {
	DocumentID string

	Body io.Reader

	Active        *bool
	IfPrimaryTerm *int
	IfSeqNo       *int
	MasterTimeout time.Duration
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
func (r XPackWatcherPutWatchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

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

	if r.Active != nil {
		params["active"] = strconv.FormatBool(*r.Active)
	}

	if r.IfPrimaryTerm != nil {
		params["if_primary_term"] = strconv.FormatInt(int64(*r.IfPrimaryTerm), 10)
	}

	if r.IfSeqNo != nil {
		params["if_seq_no"] = strconv.FormatInt(int64(*r.IfSeqNo), 10)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f XPackWatcherPutWatch) WithContext(v context.Context) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.ctx = v
	}
}

// WithBody - The watch.
//
func (f XPackWatcherPutWatch) WithBody(v io.Reader) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.Body = v
	}
}

// WithActive - specify whether the watch is in/active by default.
//
func (f XPackWatcherPutWatch) WithActive(v bool) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.Active = &v
	}
}

// WithIfPrimaryTerm - only update the watch if the last operation that has changed the watch has the specified primary term.
//
func (f XPackWatcherPutWatch) WithIfPrimaryTerm(v int) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.IfPrimaryTerm = &v
	}
}

// WithIfSeqNo - only update the watch if the last operation that has changed the watch has the specified sequence number.
//
func (f XPackWatcherPutWatch) WithIfSeqNo(v int) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.IfSeqNo = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f XPackWatcherPutWatch) WithMasterTimeout(v time.Duration) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.MasterTimeout = v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f XPackWatcherPutWatch) WithVersion(v int) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.Version = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherPutWatch) WithPretty() func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherPutWatch) WithHuman() func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherPutWatch) WithErrorTrace() func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherPutWatch) WithFilterPath(v ...string) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherPutWatch) WithHeader(h map[string]string) func(*XPackWatcherPutWatchRequest) {
	return func(r *XPackWatcherPutWatchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
