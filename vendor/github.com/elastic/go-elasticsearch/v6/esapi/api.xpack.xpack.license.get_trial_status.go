// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackLicenseGetTrialStatusFunc(t Transport) XPackLicenseGetTrialStatus {
	return func(o ...func(*XPackLicenseGetTrialStatusRequest)) (*Response, error) {
		var r = XPackLicenseGetTrialStatusRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackLicenseGetTrialStatus - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/get-basic-status.html
//
type XPackLicenseGetTrialStatus func(o ...func(*XPackLicenseGetTrialStatusRequest)) (*Response, error)

// XPackLicenseGetTrialStatusRequest configures the X Pack License Get Trial Status API request.
//
type XPackLicenseGetTrialStatusRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackLicenseGetTrialStatusRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack/license/trial_status"))
	path.WriteString("/_xpack/license/trial_status")

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
func (f XPackLicenseGetTrialStatus) WithContext(v context.Context) func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackLicenseGetTrialStatus) WithPretty() func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackLicenseGetTrialStatus) WithHuman() func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackLicenseGetTrialStatus) WithErrorTrace() func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackLicenseGetTrialStatus) WithFilterPath(v ...string) func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackLicenseGetTrialStatus) WithHeader(h map[string]string) func(*XPackLicenseGetTrialStatusRequest) {
	return func(r *XPackLicenseGetTrialStatusRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
