// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newLicensePostStartTrialFunc(t Transport) LicensePostStartTrial {
	return func(o ...func(*LicensePostStartTrialRequest)) (*Response, error) {
		var r = LicensePostStartTrialRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicensePostStartTrial - https://www.elastic.co/guide/en/elasticsearch/reference/master/start-trial.html
//
type LicensePostStartTrial func(o ...func(*LicensePostStartTrialRequest)) (*Response, error)

// LicensePostStartTrialRequest configures the License Post Start Trial API request.
//
type LicensePostStartTrialRequest struct {
	Acknowledge  *bool
	DocumentType string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r LicensePostStartTrialRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_license/start_trial"))
	path.WriteString("/_license/start_trial")

	params = make(map[string]string)

	if r.Acknowledge != nil {
		params["acknowledge"] = strconv.FormatBool(*r.Acknowledge)
	}

	if r.DocumentType != "" {
		params["type"] = r.DocumentType
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
func (f LicensePostStartTrial) WithContext(v context.Context) func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.ctx = v
	}
}

// WithAcknowledge - whether the user has acknowledged acknowledge messages (default: false).
//
func (f LicensePostStartTrial) WithAcknowledge(v bool) func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.Acknowledge = &v
	}
}

// WithDocumentType - the type of trial license to generate (default: "trial").
//
func (f LicensePostStartTrial) WithDocumentType(v string) func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.DocumentType = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicensePostStartTrial) WithPretty() func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicensePostStartTrial) WithHuman() func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicensePostStartTrial) WithErrorTrace() func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicensePostStartTrial) WithFilterPath(v ...string) func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicensePostStartTrial) WithHeader(h map[string]string) func(*LicensePostStartTrialRequest) {
	return func(r *LicensePostStartTrialRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
