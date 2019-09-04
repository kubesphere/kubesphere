// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetCategoriesFunc(t Transport) XPackMLGetCategories {
	return func(job_id string, o ...func(*XPackMLGetCategoriesRequest)) (*Response, error) {
		var r = XPackMLGetCategoriesRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetCategories - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-category.html
//
type XPackMLGetCategories func(job_id string, o ...func(*XPackMLGetCategoriesRequest)) (*Response, error)

// XPackMLGetCategoriesRequest configures the X PackML Get Categories API request.
//
type XPackMLGetCategoriesRequest struct {
	Body io.Reader

	CategoryID *int
	JobID      string

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
func (r XPackMLGetCategoriesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("results") + 1 + len("categories"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("results")
	path.WriteString("/")
	path.WriteString("categories")
	if r.CategoryID != nil {
		value := strconv.FormatInt(int64(*r.CategoryID), 10)
		path.Grow(1 + len(value))
		path.WriteString("/")
		path.WriteString(value)
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
func (f XPackMLGetCategories) WithContext(v context.Context) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.ctx = v
	}
}

// WithBody - Category selection details if not provided in URI.
//
func (f XPackMLGetCategories) WithBody(v io.Reader) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.Body = v
	}
}

// WithCategoryID - the identifier of the category definition of interest.
//
func (f XPackMLGetCategories) WithCategoryID(v int) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.CategoryID = &v
	}
}

// WithFrom - skips a number of categories.
//
func (f XPackMLGetCategories) WithFrom(v int) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of categories to get.
//
func (f XPackMLGetCategories) WithSize(v int) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetCategories) WithPretty() func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetCategories) WithHuman() func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetCategories) WithErrorTrace() func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetCategories) WithFilterPath(v ...string) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetCategories) WithHeader(h map[string]string) func(*XPackMLGetCategoriesRequest) {
	return func(r *XPackMLGetCategoriesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
