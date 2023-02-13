// Copyright 2021 by the contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signalhttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type instrumentationMiddleware struct {
	requestCounter  *prometheus.CounterVec
	requestSize     *prometheus.SummaryVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
}

// HandlerInstrumenter can instrument handlers.
type HandlerInstrumenter interface {
	NewHandler(labels prometheus.Labels, handler http.Handler) http.HandlerFunc
}

// NewHandlerInstrumenter creates a new middleware that observes some metrics for HTTP handlers.
func NewHandlerInstrumenter(r prometheus.Registerer, extraLabels []string) HandlerInstrumenter {
	labels := append([]string{"code", "method"}, extraLabels...)

	ins := &instrumentationMiddleware{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Counter of HTTP requests.",
			},
			labels,
		),
		requestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "Size of HTTP requests.",
			},
			labels,
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of latencies for HTTP requests.",
				Buckets: []float64{.1, .2, .4, 1, 2.5, 5, 8, 20, 60, 120},
			},
			labels,
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Histogram of response size for HTTP requests.",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			labels,
		),
	}

	if r != nil {
		r.MustRegister(
			ins.requestCounter,
			ins.requestSize,
			ins.requestDuration,
			ins.responseSize,
		)
	}

	return ins
}

// NewHandler wraps a HTTP handler with some metrics for HTTP handlers.
func (ins *instrumentationMiddleware) NewHandler(labels prometheus.Labels, handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerCounter(ins.requestCounter.MustCurryWith(labels),
		promhttp.InstrumentHandlerRequestSize(ins.requestSize.MustCurryWith(labels),
			promhttp.InstrumentHandlerDuration(ins.requestDuration.MustCurryWith(labels),
				promhttp.InstrumentHandlerResponseSize(ins.responseSize.MustCurryWith(labels),
					handler,
				),
			),
		),
	)
}
