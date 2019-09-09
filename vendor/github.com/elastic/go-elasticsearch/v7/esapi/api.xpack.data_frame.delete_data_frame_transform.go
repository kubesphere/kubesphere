// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newDataFrameDeleteDataFrameTransformFunc(t Transport) DataFrameDeleteDataFrameTransform {
	return func(transform_id string, o ...func(*DataFrameDeleteDataFrameTransformRequest)) (*Response, error) {
		var r = DataFrameDeleteDataFrameTransformRequest{TransformID: transform_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DataFrameDeleteDataFrameTransform - https://www.elastic.co/guide/en/elasticsearch/reference/current/delete-data-frame-transform.html
//
type DataFrameDeleteDataFrameTransform func(transform_id string, o ...func(*DataFrameDeleteDataFrameTransformRequest)) (*Response, error)

// DataFrameDeleteDataFrameTransformRequest configures the Data Frame Delete Data Frame Transform API request.
//
type DataFrameDeleteDataFrameTransformRequest struct {
	TransformID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r DataFrameDeleteDataFrameTransformRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_data_frame") + 1 + len("transforms") + 1 + len(r.TransformID))
	path.WriteString("/")
	path.WriteString("_data_frame")
	path.WriteString("/")
	path.WriteString("transforms")
	path.WriteString("/")
	path.WriteString(r.TransformID)

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
func (f DataFrameDeleteDataFrameTransform) WithContext(v context.Context) func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DataFrameDeleteDataFrameTransform) WithPretty() func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DataFrameDeleteDataFrameTransform) WithHuman() func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DataFrameDeleteDataFrameTransform) WithErrorTrace() func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DataFrameDeleteDataFrameTransform) WithFilterPath(v ...string) func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DataFrameDeleteDataFrameTransform) WithHeader(h map[string]string) func(*DataFrameDeleteDataFrameTransformRequest) {
	return func(r *DataFrameDeleteDataFrameTransformRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
