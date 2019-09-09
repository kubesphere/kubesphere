// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMigrationUpgradeFunc(t Transport) XPackMigrationUpgrade {
	return func(index string, o ...func(*XPackMigrationUpgradeRequest)) (*Response, error) {
		var r = XPackMigrationUpgradeRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMigrationUpgrade - https://www.elastic.co/guide/en/elasticsearch/reference/current/migration-api-upgrade.html
//
type XPackMigrationUpgrade func(index string, o ...func(*XPackMigrationUpgradeRequest)) (*Response, error)

// XPackMigrationUpgradeRequest configures the X Pack Migration Upgrade API request.
//
type XPackMigrationUpgradeRequest struct {
	Index string

	WaitForCompletion *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMigrationUpgradeRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("migration") + 1 + len("upgrade") + 1 + len(r.Index))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("migration")
	path.WriteString("/")
	path.WriteString("upgrade")
	path.WriteString("/")
	path.WriteString(r.Index)

	params = make(map[string]string)

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
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
func (f XPackMigrationUpgrade) WithContext(v context.Context) func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.ctx = v
	}
}

// WithWaitForCompletion - should the request block until the upgrade operation is completed.
//
func (f XPackMigrationUpgrade) WithWaitForCompletion(v bool) func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMigrationUpgrade) WithPretty() func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMigrationUpgrade) WithHuman() func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMigrationUpgrade) WithErrorTrace() func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMigrationUpgrade) WithFilterPath(v ...string) func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMigrationUpgrade) WithHeader(h map[string]string) func(*XPackMigrationUpgradeRequest) {
	return func(r *XPackMigrationUpgradeRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
