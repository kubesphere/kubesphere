// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newDataFrameStartDataFrameTransformFunc(t Transport) DataFrameStartDataFrameTransform {
	return func(transform_id string, o ...func(*DataFrameStartDataFrameTransformRequest)) (*Response, error) {
		var r = DataFrameStartDataFrameTransformRequest{TransformID: transform_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DataFrameStartDataFrameTransform - https://www.elastic.co/guide/en/elasticsearch/reference/current/start-data-frame-transform.html
//
type DataFrameStartDataFrameTransform func(transform_id string, o ...func(*DataFrameStartDataFrameTransformRequest)) (*Response, error)

// DataFrameStartDataFrameTransformRequest configures the Data Frame Start Data Frame Transform API request.
//
type DataFrameStartDataFrameTransformRequest struct {
	TransformID string

	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r DataFrameStartDataFrameTransformRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_data_frame") + 1 + len("transforms") + 1 + len(r.TransformID) + 1 + len("_start"))
	path.WriteString("/")
	path.WriteString("_data_frame")
	path.WriteString("/")
	path.WriteString("transforms")
	path.WriteString("/")
	path.WriteString(r.TransformID)
	path.WriteString("/")
	path.WriteString("_start")

	params = make(map[string]string)

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f DataFrameStartDataFrameTransform) WithContext(v context.Context) func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.ctx = v
	}
}

// WithTimeout - controls the time to wait for the transform to start.
//
func (f DataFrameStartDataFrameTransform) WithTimeout(v time.Duration) func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DataFrameStartDataFrameTransform) WithPretty() func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DataFrameStartDataFrameTransform) WithHuman() func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DataFrameStartDataFrameTransform) WithErrorTrace() func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DataFrameStartDataFrameTransform) WithFilterPath(v ...string) func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DataFrameStartDataFrameTransform) WithHeader(h map[string]string) func(*DataFrameStartDataFrameTransformRequest) {
	return func(r *DataFrameStartDataFrameTransformRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
