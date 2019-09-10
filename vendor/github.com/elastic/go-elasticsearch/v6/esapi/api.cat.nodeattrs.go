// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatNodeattrsFunc(t Transport) CatNodeattrs {
	return func(o ...func(*CatNodeattrsRequest)) (*Response, error) {
		var r = CatNodeattrsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatNodeattrs returns information about custom node attributes.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-nodeattrs.html.
//
type CatNodeattrs func(o ...func(*CatNodeattrsRequest)) (*Response, error)

// CatNodeattrsRequest configures the Cat Nodeattrs API request.
//
type CatNodeattrsRequest struct {
	Format        string
	H             []string
	Help          *bool
	Local         *bool
	MasterTimeout time.Duration
	S             []string
	V             *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatNodeattrsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cat/nodeattrs"))
	path.WriteString("/_cat/nodeattrs")

	params = make(map[string]string)

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f CatNodeattrs) WithContext(v context.Context) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.ctx = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatNodeattrs) WithFormat(v string) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatNodeattrs) WithH(v ...string) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatNodeattrs) WithHelp(v bool) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatNodeattrs) WithLocal(v bool) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatNodeattrs) WithMasterTimeout(v time.Duration) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatNodeattrs) WithS(v ...string) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatNodeattrs) WithV(v bool) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatNodeattrs) WithPretty() func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatNodeattrs) WithHuman() func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatNodeattrs) WithErrorTrace() func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatNodeattrs) WithFilterPath(v ...string) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatNodeattrs) WithHeader(h map[string]string) func(*CatNodeattrsRequest) {
	return func(r *CatNodeattrsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
