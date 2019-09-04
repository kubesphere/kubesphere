// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetFiltersFunc(t Transport) MLGetFilters {
	return func(o ...func(*MLGetFiltersRequest)) (*Response, error) {
		var r = MLGetFiltersRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetFilters -
//
type MLGetFilters func(o ...func(*MLGetFiltersRequest)) (*Response, error)

// MLGetFiltersRequest configures the ML Get Filters API request.
//
type MLGetFiltersRequest struct {
	FilterID string

	From *int
	Size *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLGetFiltersRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("filters") + 1 + len(r.FilterID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("filters")
	if r.FilterID != "" {
		path.WriteString("/")
		path.WriteString(r.FilterID)
	}

	params = make(map[string]string)

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
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
func (f MLGetFilters) WithContext(v context.Context) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.ctx = v
	}
}

// WithFilterID - the ID of the filter to fetch.
//
func (f MLGetFilters) WithFilterID(v string) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.FilterID = v
	}
}

// WithFrom - skips a number of filters.
//
func (f MLGetFilters) WithFrom(v int) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of filters to get.
//
func (f MLGetFilters) WithSize(v int) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetFilters) WithPretty() func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetFilters) WithHuman() func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetFilters) WithErrorTrace() func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetFilters) WithFilterPath(v ...string) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetFilters) WithHeader(h map[string]string) func(*MLGetFiltersRequest) {
	return func(r *MLGetFiltersRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
