// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newCatSegmentsFunc(t Transport) CatSegments {
	return func(o ...func(*CatSegmentsRequest)) (*Response, error) {
		var r = CatSegmentsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatSegments provides low-level information about the segments in the shards of an index.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-segments.html.
//
type CatSegments func(o ...func(*CatSegmentsRequest)) (*Response, error)

// CatSegmentsRequest configures the Cat Segments API request.
//
type CatSegmentsRequest struct {
	Index []string

	Bytes  string
	Format string
	H      []string
	Help   *bool
	S      []string
	V      *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatSegmentsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("segments") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("segments")
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

	params = make(map[string]string)

	if r.Bytes != "" {
		params["bytes"] = r.Bytes
	}

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
	}

	if r.V != nil {
		params["v"] = strconv.FormatBool(*r.V)
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
func (f CatSegments) WithContext(v context.Context) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to limit the returned information.
//
func (f CatSegments) WithIndex(v ...string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Index = v
	}
}

// WithBytes - the unit in which to display byte values.
//
func (f CatSegments) WithBytes(v string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Bytes = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatSegments) WithFormat(v string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatSegments) WithH(v ...string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatSegments) WithHelp(v bool) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Help = &v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatSegments) WithS(v ...string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatSegments) WithV(v bool) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatSegments) WithPretty() func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatSegments) WithHuman() func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatSegments) WithErrorTrace() func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatSegments) WithFilterPath(v ...string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatSegments) WithHeader(h map[string]string) func(*CatSegmentsRequest) {
	return func(r *CatSegmentsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
