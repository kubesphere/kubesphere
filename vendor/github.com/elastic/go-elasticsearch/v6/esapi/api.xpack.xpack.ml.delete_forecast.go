// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackMLDeleteForecastFunc(t Transport) XPackMLDeleteForecast {
	return func(job_id string, o ...func(*XPackMLDeleteForecastRequest)) (*Response, error) {
		var r = XPackMLDeleteForecastRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteForecast - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-forecast.html
//
type XPackMLDeleteForecast func(job_id string, o ...func(*XPackMLDeleteForecastRequest)) (*Response, error)

// XPackMLDeleteForecastRequest configures the X PackML Delete Forecast API request.
//
type XPackMLDeleteForecastRequest struct {
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
func (r XPackMLDeleteForecastRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_forecast") + 1 + len(r.ForecastID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLDeleteForecast) WithContext(v context.Context) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.ctx = v
	}
}

// WithForecastID - the ID of the forecast to delete, can be comma delimited list. leaving blank implies `_all`.
//
func (f XPackMLDeleteForecast) WithForecastID(v string) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.ForecastID = v
	}
}

// WithAllowNoForecasts - whether to ignore if `_all` matches no forecasts.
//
func (f XPackMLDeleteForecast) WithAllowNoForecasts(v bool) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.AllowNoForecasts = &v
	}
}

// WithTimeout - controls the time to wait until the forecast(s) are deleted. default to 30 seconds.
//
func (f XPackMLDeleteForecast) WithTimeout(v time.Duration) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteForecast) WithPretty() func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteForecast) WithHuman() func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteForecast) WithErrorTrace() func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteForecast) WithFilterPath(v ...string) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteForecast) WithHeader(h map[string]string) func(*XPackMLDeleteForecastRequest) {
	return func(r *XPackMLDeleteForecastRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
