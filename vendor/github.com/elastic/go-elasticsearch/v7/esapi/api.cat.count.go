// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatCountFunc(t Transport) CatCount {
	return func(o ...func(*CatCountRequest)) (*Response, error) {
		var r = CatCountRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatCount provides quick access to the document count of the entire cluster, or individual indices.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-count.html.
//
type CatCount func(o ...func(*CatCountRequest)) (*Response, error)

// CatCountRequest configures the Cat Count API request.
//
type CatCountRequest struct {
	Index []string

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
func (r CatCountRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("count") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("count")
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

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
func (f CatCount) WithContext(v context.Context) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to limit the returned information.
//
func (f CatCount) WithIndex(v ...string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Index = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatCount) WithFormat(v string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatCount) WithH(v ...string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatCount) WithHelp(v bool) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatCount) WithLocal(v bool) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatCount) WithMasterTimeout(v time.Duration) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatCount) WithS(v ...string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatCount) WithV(v bool) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatCount) WithPretty() func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatCount) WithHuman() func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatCount) WithErrorTrace() func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatCount) WithFilterPath(v ...string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatCount) WithHeader(h map[string]string) func(*CatCountRequest) {
	return func(r *CatCountRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
