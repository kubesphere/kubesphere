// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newClearScrollFunc(t Transport) ClearScroll {
	return func(o ...func(*ClearScrollRequest)) (*Response, error) {
		var r = ClearScrollRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClearScroll explicitly clears the search context for a scroll.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/search-request-scroll.html.
//
type ClearScroll func(o ...func(*ClearScrollRequest)) (*Response, error)

// ClearScrollRequest configures the Clear Scroll API request.
//
type ClearScrollRequest struct {
	Body io.Reader

	ScrollID []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClearScrollRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_search") + 1 + len("scroll") + 1 + len(strings.Join(r.ScrollID, ",")))
	path.WriteString("/")
	path.WriteString("_search")
	path.WriteString("/")
	path.WriteString("scroll")
	if len(r.ScrollID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.ScrollID, ","))
	}

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
func (f ClearScroll) WithContext(v context.Context) func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.ctx = v
	}
}

// WithBody - A comma-separated list of scroll IDs to clear if none was specified via the scroll_id parameter.
//
func (f ClearScroll) WithBody(v io.Reader) func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.Body = v
	}
}

// WithScrollID - a list of scroll ids to clear.
//
func (f ClearScroll) WithScrollID(v ...string) func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.ScrollID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClearScroll) WithPretty() func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClearScroll) WithHuman() func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClearScroll) WithErrorTrace() func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClearScroll) WithFilterPath(v ...string) func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClearScroll) WithHeader(h map[string]string) func(*ClearScrollRequest) {
	return func(r *ClearScrollRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
