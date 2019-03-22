// Package internalmetrics provides functionality to collect Prometheus metrics.
package internalmetrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	// Because this package is used all throughout the codebase, be VERY careful adding new
	// kiali imports here. Most likely you will encounter an import cycle error that will
	// cause a compilation failure.
)

// These constants define the different label names for the different metric timeseries
const (
	labelGraphKind        = "graph_kind"
	labelGraphType        = "graph_type"
	labelWithServiceNodes = "with_service_nodes"
	labelAppender         = "appender"
	labelRoute            = "route"
	labelQueryGroup       = "query_group"
	labelPackage          = "package"
	labelType             = "type"
	labelFunction         = "function"
)

// MetricsType defines all of Kiali's own internal metrics.
type MetricsType struct {
	GraphNodes               *prometheus.GaugeVec
	GraphGenerationTime      *prometheus.HistogramVec
	GraphAppenderTime        *prometheus.HistogramVec
	GraphMarshalTime         *prometheus.HistogramVec
	APIProcessingTime        *prometheus.HistogramVec
	PrometheusProcessingTime *prometheus.HistogramVec
	GoFunctionProcessingTime *prometheus.HistogramVec
	GoFunctionFailures       *prometheus.CounterVec
}

// Metrics contains all of Kiali's own internal metrics.
// These metrics can be accessed directly to update their values, or
// you can use available utility functions defined below.
var Metrics = MetricsType{
	GraphNodes: prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kiali_graph_nodes",
			Help: "The number of nodes in a generated graph.",
		},
		[]string{labelGraphKind, labelGraphType, labelWithServiceNodes},
	),
	GraphGenerationTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_graph_generation_duration_seconds",
			Help: "The time required to generate a graph.",
		},
		[]string{labelGraphKind, labelGraphType, labelWithServiceNodes},
	),
	GraphAppenderTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_graph_appender_duration_seconds",
			Help: "The time required to execute an appender while generating a graph.",
		},
		[]string{labelAppender},
	),
	GraphMarshalTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_graph_marshal_duration_seconds",
			Help: "The time required to marshal and return the JSON for a graph.",
		},
		[]string{labelGraphKind, labelGraphType, labelWithServiceNodes},
	),
	APIProcessingTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_api_processing_duration_seconds",
			Help: "The time required to execute a particular REST API route request.",
		},
		[]string{labelRoute},
	),
	PrometheusProcessingTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_prometheus_processing_duration_seconds",
			Help: "The time required to execute a Prometheus query.",
		},
		[]string{labelQueryGroup},
	),
	GoFunctionProcessingTime: prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "kiali_go_function_processing_duration_seconds",
			Help: "The time required to execute a particular Go function.",
		},
		[]string{labelPackage, labelType, labelFunction},
	),
	GoFunctionFailures: prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kiali_go_function_failures_total",
			Help: "Counts the total number of failures encountered by a particular Go function.",
		},
		[]string{labelPackage, labelType, labelFunction},
	),
}

// SuccessOrFailureMetricType let's you capture metrics for both successes and failures,
// where successes are tracked using a duration histogram and failures are tracked with a counter.
// Typical usage is:
// func SomeFunction(...) (..., err error) {
//     sof := GetSuccessOrFailureMetricTypeObject()
//     defer sof.ObserveNow(&err)
//     ... do the work of SomeFunction here...
// }
//
// If a function doesn't support returning an error, then call ObserveDuration directly:
//
// func SomeFunction(...) (...) {
//     sof := GetSuccessOrFailureMetricTypeObject()
//     defer sof.ObserveDuration()
//     ... do the work of SomeFunction here...
//
// If a function doesn't support returning an error, but you still need to report a failure,
// call Inc() directly to increment the failure counter:
//
// func SomeFunction(...) (...) {
//     sof := GetSuccessOrFailureMetricTypeObject()
//     defer func() { if (somethingBadHappened) { sof.Inc() } else { sof.ObserveDuration() }}()
//     ... do the work of SomeFunction here...
// }
type SuccessOrFailureMetricType struct {
	*prometheus.Timer
	prometheus.Counter
}

// ObserveNow will observe a duration unless *err is not nil
// in which case the error counter will be incremented instead.
// We use a pointer to err because this function is normally
// invoked via 'defer' and so the actual value of the error
// won't be set until the actual invocation of this function.
// (read the docs on 'defer' if you don't get it).
func (sof *SuccessOrFailureMetricType) ObserveNow(err *error) {
	if *err == nil {
		sof.ObserveDuration()
	} else {
		sof.Inc()
	}
}

// RegisterInternalMetrics must be called at startup to prepare the Prometheus scrape endpoint.
func RegisterInternalMetrics() {
	prometheus.MustRegister(
		Metrics.GraphNodes,
		Metrics.GraphGenerationTime,
		Metrics.GraphAppenderTime,
		Metrics.GraphMarshalTime,
		Metrics.APIProcessingTime,
		Metrics.PrometheusProcessingTime,
		Metrics.GoFunctionProcessingTime,
		Metrics.GoFunctionFailures,
	)
}

//
// The following are utility functions that can be used to update the internal metrics.
//

// SetGraphNodes sets the node count metric
func SetGraphNodes(graphKind string, graphType string, withServiceNodes bool, nodeCount int) {
	Metrics.GraphNodes.With(prometheus.Labels{
		labelGraphKind:        graphKind,
		labelGraphType:        graphType,
		labelWithServiceNodes: strconv.FormatBool(withServiceNodes),
	}).Set(float64(nodeCount))
}

// GetGraphGenerationTimePrometheusTimer returns a timer that can be used to store
// a value for the graph generation time metric. The timer is ticking immediately
// when this function returns.
// Typical usage is as follows:
//    promtimer := GetGraphGenerationTimePrometheusTimer(...)
//    defer promtimer.ObserveDuration()
func GetGraphGenerationTimePrometheusTimer(graphKind string, graphType string, withServiceNodes bool) *prometheus.Timer {
	timer := prometheus.NewTimer(Metrics.GraphGenerationTime.With(prometheus.Labels{
		labelGraphKind:        graphKind,
		labelGraphType:        graphType,
		labelWithServiceNodes: strconv.FormatBool(withServiceNodes),
	}))
	return timer
}

// GetGraphAppenderTimePrometheusTimer returns a timer that can be used to store
// a value for the graph appender time metric. The timer is ticking immediately
// when this function returns.
// Typical usage is as follows:
//    promtimer := GetGraphAppenderTimePrometheusTimer(...)
//    ... run the appender ...
//    promtimer.ObserveDuration()
func GetGraphAppenderTimePrometheusTimer(appenderName string) *prometheus.Timer {
	timer := prometheus.NewTimer(Metrics.GraphAppenderTime.With(prometheus.Labels{
		labelAppender: appenderName,
	}))
	return timer
}

// GetGraphMarshalTimePrometheusTimer returns a timer that can be used to store
// a value for the graph marshal time metric. The timer is ticking immediately
// when this function returns.
// Typical usage is as follows:
//    promtimer := GetGraphMarshalTimePrometheusTimer(...)
//    defer promtimer.ObserveDuration()
func GetGraphMarshalTimePrometheusTimer(graphKind string, graphType string, withServiceNodes bool) *prometheus.Timer {
	timer := prometheus.NewTimer(Metrics.GraphMarshalTime.With(prometheus.Labels{
		labelGraphKind:        graphKind,
		labelGraphType:        graphType,
		labelWithServiceNodes: strconv.FormatBool(withServiceNodes),
	}))
	return timer
}

// GetAPIProcessingTimePrometheusTimer returns a timer that can be used to store
// a value for the API processing time metric. The timer is ticking immediately
// when this function returns.
// Typical usage is as follows:
//    promtimer := GetAPIProcessingTimePrometheusTimer(...)
//    defer promtimer.ObserveDuration()
func GetAPIProcessingTimePrometheusTimer(apiRouteName string) *prometheus.Timer {
	timer := prometheus.NewTimer(Metrics.APIProcessingTime.With(prometheus.Labels{
		labelRoute: apiRouteName,
	}))
	return timer
}

// GetPrometheusProcessingTimePrometheusTimer returns a timer that can be used to store
// a value for the Prometheus query processing time metric. The timer is ticking immediately
// when this function returns.
//
// Note that the queryGroup parameter is simply some string that can be used to
// identify a particular set of Prometheus queries. This queryGroup does not necessarily have to
// identify a unique query (indeed, if you do that, that might cause too many timeseries to
// be collected), but it only needs to identify a set of queries. For example, perhaps there
// is a group of similar Prometheus queries used to generate a graph - in this case,
// the processing time for all of those queries can be combined into a single metric timeseries
// by passing in a queryGroup of "Graph-Generation".
//
// Typical usage is as follows:
//    promtimer := GetPrometheusProcessingTimePrometheusTimer(...)
//    ... execute the query ...
//    promtimer.ObserveDuration()
func GetPrometheusProcessingTimePrometheusTimer(queryGroup string) *prometheus.Timer {
	timer := prometheus.NewTimer(Metrics.PrometheusProcessingTime.With(prometheus.Labels{
		labelQueryGroup: queryGroup,
	}))
	return timer
}

// GetGoFunctionMetric returns a SuccessOrFailureMetricType object that can be used to store
// a duration value for the Go Function processing time metric when the function is successful,
// or increments the failure counter if not successful.
// If the Go Function is not on a type (i.e. is a global function), pass in an empty string for goType.
// The timer is ticking immediately when this function returns.
// See the comments for SuccessOrFailureMetricType for documentation on how to use the returned object.
func GetGoFunctionMetric(goPkg string, goType string, goFunc string) SuccessOrFailureMetricType {
	return SuccessOrFailureMetricType{
		prometheus.NewTimer(Metrics.GoFunctionProcessingTime.With(prometheus.Labels{
			labelPackage:  goPkg,
			labelType:     goType,
			labelFunction: goFunc,
		})),
		Metrics.GoFunctionFailures.With(prometheus.Labels{
			labelPackage:  goPkg,
			labelType:     goType,
			labelFunction: goFunc,
		}),
	}
}
