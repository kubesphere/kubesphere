// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMigrationDeprecationsFunc(t Transport) XPackMigrationDeprecations {
	return func(o ...func(*XPackMigrationDeprecationsRequest)) (*Response, error) {
		var r = XPackMigrationDeprecationsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMigrationDeprecations - http://www.elastic.co/guide/en/elasticsearch/reference/6.7/migration-api-deprecation.html
//
type XPackMigrationDeprecations func(o ...func(*XPackMigrationDeprecationsRequest)) (*Response, error)

// XPackMigrationDeprecationsRequest configures the X Pack Migration Deprecations API request.
//
type XPackMigrationDeprecationsRequest struct {
	Index string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMigrationDeprecationsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len("_xpack") + 1 + len("migration") + 1 + len("deprecations"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("migration")
	path.WriteString("/")
	path.WriteString("deprecations")

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
func (f XPackMigrationDeprecations) WithContext(v context.Context) func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.ctx = v
	}
}

// WithIndex - index pattern.
//
func (f XPackMigrationDeprecations) WithIndex(v string) func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMigrationDeprecations) WithPretty() func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMigrationDeprecations) WithHuman() func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMigrationDeprecations) WithErrorTrace() func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMigrationDeprecations) WithFilterPath(v ...string) func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMigrationDeprecations) WithHeader(h map[string]string) func(*XPackMigrationDeprecationsRequest) {
	return func(r *XPackMigrationDeprecationsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
