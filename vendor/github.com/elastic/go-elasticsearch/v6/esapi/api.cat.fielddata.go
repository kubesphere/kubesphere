// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatFielddataFunc(t Transport) CatFielddata {
	return func(o ...func(*CatFielddataRequest)) (*Response, error) {
		var r = CatFielddataRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatFielddata shows how much heap memory is currently being used by fielddata on every data node in the cluster.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-fielddata.html.
//
type CatFielddata func(o ...func(*CatFielddataRequest)) (*Response, error)

// CatFielddataRequest configures the Cat Fielddata API request.
//
type CatFielddataRequest struct {
	Fields []string

	Bytes         string
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
func (r CatFielddataRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("fielddata") + 1 + len(strings.Join(r.Fields, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("fielddata")
	if len(r.Fields) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Fields, ","))
	}

	params = make(map[string]string)

	if r.Bytes != "" {
		params["bytes"] = r.Bytes
	}

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
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
func (f CatFielddata) WithContext(v context.Context) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.ctx = v
	}
}

// WithFields - a list of fields to return the fielddata size.
//
func (f CatFielddata) WithFields(v ...string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Fields = v
	}
}

// WithBytes - the unit in which to display byte values.
//
func (f CatFielddata) WithBytes(v string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Bytes = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatFielddata) WithFormat(v string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatFielddata) WithH(v ...string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatFielddata) WithHelp(v bool) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatFielddata) WithLocal(v bool) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatFielddata) WithMasterTimeout(v time.Duration) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatFielddata) WithS(v ...string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatFielddata) WithV(v bool) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatFielddata) WithPretty() func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatFielddata) WithHuman() func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatFielddata) WithErrorTrace() func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatFielddata) WithFilterPath(v ...string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatFielddata) WithHeader(h map[string]string) func(*CatFielddataRequest) {
	return func(r *CatFielddataRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
