// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newCatHelpFunc(t Transport) CatHelp {
	return func(o ...func(*CatHelpRequest)) (*Response, error) {
		var r = CatHelpRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatHelp returns help for the Cat APIs.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat.html.
//
type CatHelp func(o ...func(*CatHelpRequest)) (*Response, error)

// CatHelpRequest configures the Cat Help API request.
//
type CatHelpRequest struct {
	Help *bool
	S    []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatHelpRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cat"))
	path.WriteString("/_cat")

	params = make(map[string]string)

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
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
func (f CatHelp) WithContext(v context.Context) func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.ctx = v
	}
}

// WithHelp - return help information.
//
func (f CatHelp) WithHelp(v bool) func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.Help = &v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatHelp) WithS(v ...string) func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.S = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatHelp) WithPretty() func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatHelp) WithHuman() func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatHelp) WithErrorTrace() func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatHelp) WithFilterPath(v ...string) func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatHelp) WithHeader(h map[string]string) func(*CatHelpRequest) {
	return func(r *CatHelpRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
