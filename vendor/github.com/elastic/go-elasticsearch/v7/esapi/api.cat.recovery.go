// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatRecoveryFunc(t Transport) CatRecovery {
	return func(o ...func(*CatRecoveryRequest)) (*Response, error) {
		var r = CatRecoveryRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatRecovery returns information about index shard recoveries, both on-going completed.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-recovery.html.
//
type CatRecovery func(o ...func(*CatRecoveryRequest)) (*Response, error)

// CatRecoveryRequest configures the Cat Recovery API request.
//
type CatRecoveryRequest struct {
	Index []string

	Bytes         string
	Format        string
	H             []string
	Help          *bool
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
func (r CatRecoveryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("recovery") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("recovery")
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
func (f CatRecovery) WithContext(v context.Context) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to limit the returned information.
//
func (f CatRecovery) WithIndex(v ...string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Index = v
	}
}

// WithBytes - the unit in which to display byte values.
//
func (f CatRecovery) WithBytes(v string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Bytes = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatRecovery) WithFormat(v string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatRecovery) WithH(v ...string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatRecovery) WithHelp(v bool) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Help = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatRecovery) WithMasterTimeout(v time.Duration) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatRecovery) WithS(v ...string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatRecovery) WithV(v bool) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatRecovery) WithPretty() func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatRecovery) WithHuman() func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatRecovery) WithErrorTrace() func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatRecovery) WithFilterPath(v ...string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatRecovery) WithHeader(h map[string]string) func(*CatRecoveryRequest) {
	return func(r *CatRecoveryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
