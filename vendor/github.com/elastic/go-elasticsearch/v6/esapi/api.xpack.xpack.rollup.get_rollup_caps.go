// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackRollupGetRollupCapsFunc(t Transport) XPackRollupGetRollupCaps {
	return func(o ...func(*XPackRollupGetRollupCapsRequest)) (*Response, error) {
		var r = XPackRollupGetRollupCapsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackRollupGetRollupCaps -
//
type XPackRollupGetRollupCaps func(o ...func(*XPackRollupGetRollupCapsRequest)) (*Response, error)

// XPackRollupGetRollupCapsRequest configures the X Pack Rollup Get Rollup Caps API request.
//
type XPackRollupGetRollupCapsRequest struct {
	DocumentID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackRollupGetRollupCapsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("rollup") + 1 + len("data") + 1 + len(r.DocumentID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("rollup")
	path.WriteString("/")
	path.WriteString("data")
	if r.DocumentID != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentID)
	}

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
func (f XPackRollupGetRollupCaps) WithContext(v context.Context) func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.ctx = v
	}
}

// WithDocumentID - the ID of the index to check rollup capabilities on, or left blank for all jobs.
//
func (f XPackRollupGetRollupCaps) WithDocumentID(v string) func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.DocumentID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackRollupGetRollupCaps) WithPretty() func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackRollupGetRollupCaps) WithHuman() func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackRollupGetRollupCaps) WithErrorTrace() func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackRollupGetRollupCaps) WithFilterPath(v ...string) func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackRollupGetRollupCaps) WithHeader(h map[string]string) func(*XPackRollupGetRollupCapsRequest) {
	return func(r *XPackRollupGetRollupCapsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
