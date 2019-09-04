// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetCategoriesFunc(t Transport) MLGetCategories {
	return func(job_id string, o ...func(*MLGetCategoriesRequest)) (*Response, error) {
		var r = MLGetCategoriesRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetCategories - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-category.html
//
type MLGetCategories func(job_id string, o ...func(*MLGetCategoriesRequest)) (*Response, error)

// MLGetCategoriesRequest configures the ML Get Categories API request.
//
type MLGetCategoriesRequest struct {
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
func (r MLGetCategoriesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("results") + 1 + len("categories"))
	path.WriteString("/")
	path.WriteString("_ml")
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
func (f MLGetCategories) WithContext(v context.Context) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.ctx = v
	}
}

// WithBody - Category selection details if not provided in URI.
//
func (f MLGetCategories) WithBody(v io.Reader) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.Body = v
	}
}

// WithCategoryID - the identifier of the category definition of interest.
//
func (f MLGetCategories) WithCategoryID(v int) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.CategoryID = &v
	}
}

// WithFrom - skips a number of categories.
//
func (f MLGetCategories) WithFrom(v int) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of categories to get.
//
func (f MLGetCategories) WithSize(v int) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetCategories) WithPretty() func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetCategories) WithHuman() func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetCategories) WithErrorTrace() func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetCategories) WithFilterPath(v ...string) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetCategories) WithHeader(h map[string]string) func(*MLGetCategoriesRequest) {
	return func(r *MLGetCategoriesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
