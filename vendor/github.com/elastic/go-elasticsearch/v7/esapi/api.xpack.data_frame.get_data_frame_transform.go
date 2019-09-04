// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newDataFrameGetDataFrameTransformFunc(t Transport) DataFrameGetDataFrameTransform {
	return func(o ...func(*DataFrameGetDataFrameTransformRequest)) (*Response, error) {
		var r = DataFrameGetDataFrameTransformRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DataFrameGetDataFrameTransform - https://www.elastic.co/guide/en/elasticsearch/reference/current/get-data-frame-transform.html
//
type DataFrameGetDataFrameTransform func(o ...func(*DataFrameGetDataFrameTransformRequest)) (*Response, error)

// DataFrameGetDataFrameTransformRequest configures the Data Frame Get Data Frame Transform API request.
//
type DataFrameGetDataFrameTransformRequest struct {
	TransformID string

	AllowNoMatch *bool
	From         *int
	Size         *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r DataFrameGetDataFrameTransformRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_data_frame") + 1 + len("transforms") + 1 + len(r.TransformID))
	path.WriteString("/")
	path.WriteString("_data_frame")
	path.WriteString("/")
	path.WriteString("transforms")
	if r.TransformID != "" {
		path.WriteString("/")
		path.WriteString(r.TransformID)
	}

	params = make(map[string]string)

	if r.AllowNoMatch != nil {
		params["allow_no_match"] = strconv.FormatBool(*r.AllowNoMatch)
	}

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
func (f DataFrameGetDataFrameTransform) WithContext(v context.Context) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.ctx = v
	}
}

// WithTransformID - the ID or comma delimited list of ID expressions of the transforms to get, '_all' or '*' implies get all transforms.
//
func (f DataFrameGetDataFrameTransform) WithTransformID(v string) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.TransformID = v
	}
}

// WithAllowNoMatch - whether to ignore if a wildcard expression matches no data frame transforms. (this includes `_all` string or when no data frame transforms have been specified).
//
func (f DataFrameGetDataFrameTransform) WithAllowNoMatch(v bool) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.AllowNoMatch = &v
	}
}

// WithFrom - skips a number of transform configs, defaults to 0.
//
func (f DataFrameGetDataFrameTransform) WithFrom(v int) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of transforms to get, defaults to 100.
//
func (f DataFrameGetDataFrameTransform) WithSize(v int) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DataFrameGetDataFrameTransform) WithPretty() func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DataFrameGetDataFrameTransform) WithHuman() func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DataFrameGetDataFrameTransform) WithErrorTrace() func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DataFrameGetDataFrameTransform) WithFilterPath(v ...string) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DataFrameGetDataFrameTransform) WithHeader(h map[string]string) func(*DataFrameGetDataFrameTransformRequest) {
	return func(r *DataFrameGetDataFrameTransformRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
