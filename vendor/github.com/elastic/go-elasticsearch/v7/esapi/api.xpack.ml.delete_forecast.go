// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newMLDeleteForecastFunc(t Transport) MLDeleteForecast {
	return func(job_id string, o ...func(*MLDeleteForecastRequest)) (*Response, error) {
		var r = MLDeleteForecastRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLDeleteForecast - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-forecast.html
//
type MLDeleteForecast func(job_id string, o ...func(*MLDeleteForecastRequest)) (*Response, error)

// MLDeleteForecastRequest configures the ML Delete Forecast API request.
//
type MLDeleteForecastRequest struct {
	ForecastID string
	JobID      string

	AllowNoForecasts *bool
	Timeout          time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLDeleteForecastRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_forecast") + 1 + len(r.ForecastID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_forecast")
	if r.ForecastID != "" {
		path.WriteString("/")
		path.WriteString(r.ForecastID)
	}

	params = make(map[string]string)

	if r.AllowNoForecasts != nil {
		params["allow_no_forecasts"] = strconv.FormatBool(*r.AllowNoForecasts)
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
func (f MLDeleteForecast) WithContext(v context.Context) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.ctx = v
	}
}

// WithForecastID - the ID of the forecast to delete, can be comma delimited list. leaving blank implies `_all`.
//
func (f MLDeleteForecast) WithForecastID(v string) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.ForecastID = v
	}
}

// WithAllowNoForecasts - whether to ignore if `_all` matches no forecasts.
//
func (f MLDeleteForecast) WithAllowNoForecasts(v bool) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.AllowNoForecasts = &v
	}
}

// WithTimeout - controls the time to wait until the forecast(s) are deleted. default to 30 seconds.
//
func (f MLDeleteForecast) WithTimeout(v time.Duration) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLDeleteForecast) WithPretty() func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLDeleteForecast) WithHuman() func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLDeleteForecast) WithErrorTrace() func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLDeleteForecast) WithFilterPath(v ...string) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLDeleteForecast) WithHeader(h map[string]string) func(*MLDeleteForecastRequest) {
	return func(r *MLDeleteForecastRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
