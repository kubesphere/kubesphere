// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newDeleteScriptFunc(t Transport) DeleteScript {
	return func(id string, o ...func(*DeleteScriptRequest)) (*Response, error) {
		var r = DeleteScriptRequest{ScriptID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DeleteScript deletes a script.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-scripting.html.
//
type DeleteScript func(id string, o ...func(*DeleteScriptRequest)) (*Response, error)

// DeleteScriptRequest configures the Delete Script API request.
//
type DeleteScriptRequest struct {
	ScriptID string

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
func (r DeleteScriptRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_scripts") + 1 + len(r.ScriptID))
	path.WriteString("/")
	path.WriteString("_scripts")
	path.WriteString("/")
	path.WriteString(r.ScriptID)

	params = make(map[string]string)

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
func (f DeleteScript) WithContext(v context.Context) func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f DeleteScript) WithMasterTimeout(v time.Duration) func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f DeleteScript) WithTimeout(v time.Duration) func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DeleteScript) WithPretty() func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DeleteScript) WithHuman() func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DeleteScript) WithErrorTrace() func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DeleteScript) WithFilterPath(v ...string) func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DeleteScript) WithHeader(h map[string]string) func(*DeleteScriptRequest) {
	return func(r *DeleteScriptRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
