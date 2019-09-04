// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newLicenseGetTrialStatusFunc(t Transport) LicenseGetTrialStatus {
	return func(o ...func(*LicenseGetTrialStatusRequest)) (*Response, error) {
		var r = LicenseGetTrialStatusRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicenseGetTrialStatus - https://www.elastic.co/guide/en/elasticsearch/reference/master/get-trial-status.html
//
type LicenseGetTrialStatus func(o ...func(*LicenseGetTrialStatusRequest)) (*Response, error)

// LicenseGetTrialStatusRequest configures the License Get Trial Status API request.
//
type LicenseGetTrialStatusRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r LicenseGetTrialStatusRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_license/trial_status"))
	path.WriteString("/_license/trial_status")

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
func (f LicenseGetTrialStatus) WithContext(v context.Context) func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicenseGetTrialStatus) WithPretty() func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicenseGetTrialStatus) WithHuman() func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicenseGetTrialStatus) WithErrorTrace() func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicenseGetTrialStatus) WithFilterPath(v ...string) func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicenseGetTrialStatus) WithHeader(h map[string]string) func(*LicenseGetTrialStatusRequest) {
	return func(r *LicenseGetTrialStatusRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
