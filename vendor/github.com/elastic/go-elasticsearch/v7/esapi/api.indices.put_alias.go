// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func newIndicesPutAliasFunc(t Transport) IndicesPutAlias {
	return func(index []string, name string, o ...func(*IndicesPutAliasRequest)) (*Response, error) {
		var r = IndicesPutAliasRequest{Index: index, Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesPutAlias creates or updates an alias.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-aliases.html.
//
type IndicesPutAlias func(index []string, name string, o ...func(*IndicesPutAliasRequest)) (*Response, error)

// IndicesPutAliasRequest configures the Indices Put Alias API request.
//
type IndicesPutAliasRequest struct {
	Index []string

	Body io.Reader

	Name string

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
func (r IndicesPutAliasRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_aliases") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_aliases")
	path.WriteString("/")
	path.WriteString(r.Name)

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
func (f IndicesPutAlias) WithContext(v context.Context) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.ctx = v
	}
}

// WithBody - The settings for the alias, such as `routing` or `filter`.
//
func (f IndicesPutAlias) WithBody(v io.Reader) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.Body = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesPutAlias) WithMasterTimeout(v time.Duration) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit timestamp for the document.
//
func (f IndicesPutAlias) WithTimeout(v time.Duration) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesPutAlias) WithPretty() func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesPutAlias) WithHuman() func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesPutAlias) WithErrorTrace() func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesPutAlias) WithFilterPath(v ...string) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesPutAlias) WithHeader(h map[string]string) func(*IndicesPutAliasRequest) {
	return func(r *IndicesPutAliasRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
