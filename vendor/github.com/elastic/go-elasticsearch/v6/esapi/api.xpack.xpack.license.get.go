// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackLicenseGetFunc(t Transport) XPackLicenseGet {
	return func(o ...func(*XPackLicenseGetRequest)) (*Response, error) {
		var r = XPackLicenseGetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackLicenseGet - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/get-license.html
//
type XPackLicenseGet func(o ...func(*XPackLicenseGetRequest)) (*Response, error)

// XPackLicenseGetRequest configures the X Pack License Get API request.
//
type XPackLicenseGetRequest struct {
	Local *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackLicenseGetRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack/license"))
	path.WriteString("/_xpack/license")

	params = make(map[string]string)

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
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
func (f XPackLicenseGet) WithContext(v context.Context) func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.ctx = v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f XPackLicenseGet) WithLocal(v bool) func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.Local = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackLicenseGet) WithPretty() func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackLicenseGet) WithHuman() func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackLicenseGet) WithErrorTrace() func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackLicenseGet) WithFilterPath(v ...string) func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackLicenseGet) WithHeader(h map[string]string) func(*XPackLicenseGetRequest) {
	return func(r *XPackLicenseGetRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
