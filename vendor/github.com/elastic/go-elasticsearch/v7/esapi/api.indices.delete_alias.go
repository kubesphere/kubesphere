// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newIndicesDeleteAliasFunc(t Transport) IndicesDeleteAlias {
	return func(index []string, name []string, o ...func(*IndicesDeleteAliasRequest)) (*Response, error) {
		var r = IndicesDeleteAliasRequest{Index: index, Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesDeleteAlias deletes an alias.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-aliases.html.
//
type IndicesDeleteAlias func(index []string, name []string, o ...func(*IndicesDeleteAliasRequest)) (*Response, error)

// IndicesDeleteAliasRequest configures the Indices Delete Alias API request.
//
type IndicesDeleteAliasRequest struct {
	Index []string

	Name []string

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
func (r IndicesDeleteAliasRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_aliases") + 1 + len(strings.Join(r.Name, ",")))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_aliases")
	path.WriteString("/")
	path.WriteString(strings.Join(r.Name, ","))

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
func (f IndicesDeleteAlias) WithContext(v context.Context) func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesDeleteAlias) WithMasterTimeout(v time.Duration) func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit timestamp for the document.
//
func (f IndicesDeleteAlias) WithTimeout(v time.Duration) func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesDeleteAlias) WithPretty() func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesDeleteAlias) WithHuman() func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesDeleteAlias) WithErrorTrace() func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesDeleteAlias) WithFilterPath(v ...string) func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesDeleteAlias) WithHeader(h map[string]string) func(*IndicesDeleteAliasRequest) {
	return func(r *IndicesDeleteAliasRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
