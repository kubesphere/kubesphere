// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newMLForecastFunc(t Transport) MLForecast {
	return func(job_id string, o ...func(*MLForecastRequest)) (*Response, error) {
		var r = MLForecastRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLForecast -
//
type MLForecast func(job_id string, o ...func(*MLForecastRequest)) (*Response, error)

// MLForecastRequest configures the ML Forecast API request.
//
type MLForecastRequest struct {
	JobID string

	Duration  time.Duration
	ExpiresIn time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLForecastRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_forecast"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_forecast")

	params = make(map[string]string)

	if r.Duration != 0 {
		params["duration"] = formatDuration(r.Duration)
	}

	if r.ExpiresIn != 0 {
		params["expires_in"] = formatDuration(r.ExpiresIn)
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
func (f MLForecast) WithContext(v context.Context) func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.ctx = v
	}
}

// WithDuration - the duration of the forecast.
//
func (f MLForecast) WithDuration(v time.Duration) func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.Duration = v
	}
}

// WithExpiresIn - the time interval after which the forecast expires. expired forecasts will be deleted at the first opportunity..
//
func (f MLForecast) WithExpiresIn(v time.Duration) func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.ExpiresIn = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLForecast) WithPretty() func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLForecast) WithHuman() func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLForecast) WithErrorTrace() func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLForecast) WithFilterPath(v ...string) func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLForecast) WithHeader(h map[string]string) func(*MLForecastRequest) {
	return func(r *MLForecastRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
