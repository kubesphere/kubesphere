// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newCCRFollowStatsFunc(t Transport) CCRFollowStats {
	return func(index []string, o ...func(*CCRFollowStatsRequest)) (*Response, error) {
		var r = CCRFollowStatsRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRFollowStats - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-get-follow-stats.html
//
type CCRFollowStats func(index []string, o ...func(*CCRFollowStatsRequest)) (*Response, error)

// CCRFollowStatsRequest configures the CCR Follow Stats API request.
//
type CCRFollowStatsRequest struct {
	Index []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRFollowStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_ccr") + 1 + len("stats"))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("stats")

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
func (f CCRFollowStats) WithContext(v context.Context) func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRFollowStats) WithPretty() func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRFollowStats) WithHuman() func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRFollowStats) WithErrorTrace() func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRFollowStats) WithFilterPath(v ...string) func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRFollowStats) WithHeader(h map[string]string) func(*CCRFollowStatsRequest) {
	return func(r *CCRFollowStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
