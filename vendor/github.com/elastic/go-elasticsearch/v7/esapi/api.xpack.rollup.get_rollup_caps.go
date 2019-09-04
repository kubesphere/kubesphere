// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newRollupGetRollupCapsFunc(t Transport) RollupGetRollupCaps {
	return func(o ...func(*RollupGetRollupCapsRequest)) (*Response, error) {
		var r = RollupGetRollupCapsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// RollupGetRollupCaps -
//
type RollupGetRollupCaps func(o ...func(*RollupGetRollupCapsRequest)) (*Response, error)

// RollupGetRollupCapsRequest configures the Rollup Get Rollup Caps API request.
//
type RollupGetRollupCapsRequest struct {
	Index string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r RollupGetRollupCapsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_rollup") + 1 + len("data") + 1 + len(r.Index))
	path.WriteString("/")
	path.WriteString("_rollup")
	path.WriteString("/")
	path.WriteString("data")
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
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
func (f RollupGetRollupCaps) WithContext(v context.Context) func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.ctx = v
	}
}

// WithIndex - the ID of the index to check rollup capabilities on, or left blank for all jobs.
//
func (f RollupGetRollupCaps) WithIndex(v string) func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f RollupGetRollupCaps) WithPretty() func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f RollupGetRollupCaps) WithHuman() func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f RollupGetRollupCaps) WithErrorTrace() func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f RollupGetRollupCaps) WithFilterPath(v ...string) func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f RollupGetRollupCaps) WithHeader(h map[string]string) func(*RollupGetRollupCapsRequest) {
	return func(r *RollupGetRollupCapsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
