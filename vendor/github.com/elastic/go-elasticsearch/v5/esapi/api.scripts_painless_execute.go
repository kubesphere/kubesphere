// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"strings"
)

func newScriptsPainlessExecuteFunc(t Transport) ScriptsPainlessExecute {
	return func(o ...func(*ScriptsPainlessExecuteRequest)) (*Response, error) {
		var r = ScriptsPainlessExecuteRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ScriptsPainlessExecute allows an arbitrary script to be executed and a result to be returned
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/painless/master/painless-execute-api.html.
//
type ScriptsPainlessExecute func(o ...func(*ScriptsPainlessExecuteRequest)) (*Response, error)

// ScriptsPainlessExecuteRequest configures the Scripts Painless Execute API request.
//
type ScriptsPainlessExecuteRequest struct {
	Body io.Reader

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ScriptsPainlessExecuteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_scripts/painless/_execute"))
	path.WriteString("/_scripts/painless/_execute")

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
func (f ScriptsPainlessExecute) WithContext(v context.Context) func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.ctx = v
	}
}

// WithBody - The script to execute.
//
func (f ScriptsPainlessExecute) WithBody(v io.Reader) func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.Body = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ScriptsPainlessExecute) WithPretty() func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ScriptsPainlessExecute) WithHuman() func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ScriptsPainlessExecute) WithErrorTrace() func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ScriptsPainlessExecute) WithFilterPath(v ...string) func(*ScriptsPainlessExecuteRequest) {
	return func(r *ScriptsPainlessExecuteRequest) {
		r.FilterPath = v
	}
}
