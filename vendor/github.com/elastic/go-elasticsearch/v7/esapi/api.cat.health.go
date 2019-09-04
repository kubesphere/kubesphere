// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatHealthFunc(t Transport) CatHealth {
	return func(o ...func(*CatHealthRequest)) (*Response, error) {
		var r = CatHealthRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatHealth returns a concise representation of the cluster health.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-health.html.
//
type CatHealth func(o ...func(*CatHealthRequest)) (*Response, error)

// CatHealthRequest configures the Cat Health API request.
//
type CatHealthRequest struct {
	Format        string
	H             []string
	Help          *bool
	Local         *bool
	MasterTimeout time.Duration
	S             []string
	Ts            *bool
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
func (r CatHealthRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cat/health"))
	path.WriteString("/_cat/health")

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

	if r.Ts != nil {
		params["ts"] = strconv.FormatBool(*r.Ts)
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
func (f CatHealth) WithContext(v context.Context) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.ctx = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatHealth) WithFormat(v string) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatHealth) WithH(v ...string) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatHealth) WithHelp(v bool) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatHealth) WithLocal(v bool) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatHealth) WithMasterTimeout(v time.Duration) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatHealth) WithS(v ...string) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.S = v
	}
}

// WithTs - set to false to disable timestamping.
//
func (f CatHealth) WithTs(v bool) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Ts = &v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatHealth) WithV(v bool) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatHealth) WithPretty() func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatHealth) WithHuman() func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatHealth) WithErrorTrace() func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatHealth) WithFilterPath(v ...string) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatHealth) WithHeader(h map[string]string) func(*CatHealthRequest) {
	return func(r *CatHealthRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
