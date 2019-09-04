// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newIndicesRecoveryFunc(t Transport) IndicesRecovery {
	return func(o ...func(*IndicesRecoveryRequest)) (*Response, error) {
		var r = IndicesRecoveryRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesRecovery returns information about ongoing index shard recoveries.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-recovery.html.
//
type IndicesRecovery func(o ...func(*IndicesRecoveryRequest)) (*Response, error)

// IndicesRecoveryRequest configures the Indices Recovery API request.
//
type IndicesRecoveryRequest struct {
	Index []string

	ActiveOnly *bool
	Detailed   *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesRecoveryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_recovery"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_recovery")

	params = make(map[string]string)

	if r.ActiveOnly != nil {
		params["active_only"] = strconv.FormatBool(*r.ActiveOnly)
	}

	if r.Detailed != nil {
		params["detailed"] = strconv.FormatBool(*r.Detailed)
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
func (f IndicesRecovery) WithContext(v context.Context) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesRecovery) WithIndex(v ...string) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.Index = v
	}
}

// WithActiveOnly - display only those recoveries that are currently on-going.
//
func (f IndicesRecovery) WithActiveOnly(v bool) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.ActiveOnly = &v
	}
}

// WithDetailed - whether to display detailed information about shard recovery.
//
func (f IndicesRecovery) WithDetailed(v bool) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.Detailed = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesRecovery) WithPretty() func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesRecovery) WithHuman() func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesRecovery) WithErrorTrace() func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesRecovery) WithFilterPath(v ...string) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesRecovery) WithHeader(h map[string]string) func(*IndicesRecoveryRequest) {
	return func(r *IndicesRecoveryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
