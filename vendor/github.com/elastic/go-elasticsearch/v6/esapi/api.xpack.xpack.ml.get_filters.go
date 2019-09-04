// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetFiltersFunc(t Transport) XPackMLGetFilters {
	return func(o ...func(*XPackMLGetFiltersRequest)) (*Response, error) {
		var r = XPackMLGetFiltersRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetFilters -
//
type XPackMLGetFilters func(o ...func(*XPackMLGetFiltersRequest)) (*Response, error)

// XPackMLGetFiltersRequest configures the X PackML Get Filters API request.
//
type XPackMLGetFiltersRequest struct {
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
func (r XPackMLGetFiltersRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("filters") + 1 + len(r.FilterID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLGetFilters) WithContext(v context.Context) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.ctx = v
	}
}

// WithFilterID - the ID of the filter to fetch.
//
func (f XPackMLGetFilters) WithFilterID(v string) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.FilterID = v
	}
}

// WithFrom - skips a number of filters.
//
func (f XPackMLGetFilters) WithFrom(v int) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of filters to get.
//
func (f XPackMLGetFilters) WithSize(v int) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetFilters) WithPretty() func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetFilters) WithHuman() func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetFilters) WithErrorTrace() func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetFilters) WithFilterPath(v ...string) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetFilters) WithHeader(h map[string]string) func(*XPackMLGetFiltersRequest) {
	return func(r *XPackMLGetFiltersRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
