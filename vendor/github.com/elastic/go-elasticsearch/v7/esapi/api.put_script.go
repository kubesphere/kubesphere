// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func newPutScriptFunc(t Transport) PutScript {
	return func(id string, body io.Reader, o ...func(*PutScriptRequest)) (*Response, error) {
		var r = PutScriptRequest{ScriptID: id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// PutScript creates or updates a script.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-scripting.html.
//
type PutScript func(id string, body io.Reader, o ...func(*PutScriptRequest)) (*Response, error)

// PutScriptRequest configures the Put Script API request.
//
type PutScriptRequest struct {
	ScriptID string

	Body io.Reader

	ScriptContext string

	MasterTimeout time.Duration
	Timeout       time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r PutScriptRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_scripts") + 1 + len(r.ScriptID) + 1 + len(r.ScriptContext))
	path.WriteString("/")
	path.WriteString("_scripts")
	path.WriteString("/")
	path.WriteString(r.ScriptID)
	if r.ScriptContext != "" {
		path.WriteString("/")
		path.WriteString(r.ScriptContext)
	}

	params = make(map[string]string)

	if r.ScriptContext != "" {
		params["context"] = r.ScriptContext
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

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
func (f PutScript) WithContext(v context.Context) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.ctx = v
	}
}

// WithScriptContext - script context.
//
func (f PutScript) WithScriptContext(v string) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.ScriptContext = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f PutScript) WithMasterTimeout(v time.Duration) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f PutScript) WithTimeout(v time.Duration) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f PutScript) WithPretty() func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f PutScript) WithHuman() func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f PutScript) WithErrorTrace() func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f PutScript) WithFilterPath(v ...string) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f PutScript) WithHeader(h map[string]string) func(*PutScriptRequest) {
	return func(r *PutScriptRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
