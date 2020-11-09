package prometheus

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/prometheus/internalmetrics"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func getMetrics(api v1.API, q *IstioMetricsQuery) Metrics {
	labels, labelsError := buildLabelStrings(q)
	grouping := strings.Join(q.ByLabels, ",")
	metrics := fetchAllMetrics(api, q, labels, labelsError, grouping)
	return metrics
}

func buildLabelStrings(q *IstioMetricsQuery) (string, string) {
	labels := []string{fmt.Sprintf(`reporter="%s"`, q.Reporter)}
	ref := "destination"
	if q.Direction == "outbound" {
		ref = "source"
	}

	if q.Service != "" {
		// inbound only
		labels = append(labels, fmt.Sprintf(`destination_service_name="%s"`, q.Service))
		if q.Namespace != "" {
			labels = append(labels, fmt.Sprintf(`destination_service_namespace="%s"`, q.Namespace))
		}
	} else if q.Namespace != "" {
		labels = append(labels, fmt.Sprintf(`%s_workload_namespace="%s"`, ref, q.Namespace))
	}
	if q.Workload != "" {
		labels = append(labels, fmt.Sprintf(`%s_workload="%s"`, ref, q.Workload))
	}
	if q.App != "" {
		labels = append(labels, fmt.Sprintf(`%s_app="%s"`, ref, q.App))
	}
	if q.RequestProtocol != "" {
		labels = append(labels, fmt.Sprintf(`request_protocol="%s"`, q.RequestProtocol))
	}

	full := "{" + strings.Join(labels, ",") + "}"

	labels = append(labels, `response_code=~"[5|4].*"`)
	errors := "{" + strings.Join(labels, ",") + "}"

	return full, errors
}

func fetchAllMetrics(api v1.API, q *IstioMetricsQuery, labels, labelsError, grouping string) Metrics {
	var wg sync.WaitGroup
	fetchRate := func(p8sFamilyName string, metric **Metric, lbl string) {
		defer wg.Done()
		m := fetchRateRange(api, p8sFamilyName, lbl, grouping, &q.BaseMetricsQuery)
		*metric = m
	}

	fetchHisto := func(p8sFamilyName string, histo *Histogram) {
		defer wg.Done()
		h := fetchHistogramRange(api, p8sFamilyName, labels, grouping, &q.BaseMetricsQuery)
		*histo = h
	}

	type resultHolder struct {
		metric     *Metric
		histo      Histogram
		definition istioMetric
	}
	maxResults := len(istioMetrics)
	results := make([]*resultHolder, maxResults, maxResults)

	for i, istioMetric := range istioMetrics {
		// if filters is empty, fetch all anyway
		doFetch := len(q.Filters) == 0
		if !doFetch {
			for _, filter := range q.Filters {
				if filter == istioMetric.kialiName {
					doFetch = true
					break
				}
			}
		}
		if doFetch {
			wg.Add(1)
			result := resultHolder{definition: istioMetric}
			results[i] = &result
			if istioMetric.isHisto {
				go fetchHisto(istioMetric.istioName, &result.histo)
			} else {
				labelsToUse := istioMetric.labelsToUse(labels, labelsError)
				go fetchRate(istioMetric.istioName, &result.metric, labelsToUse)
			}
		}
	}
	wg.Wait()

	// Return results as two maps per reporter
	metrics := make(map[string]*Metric)
	histograms := make(map[string]Histogram)
	for _, result := range results {
		if result != nil {
			if result.definition.isHisto {
				histograms[result.definition.kialiName] = result.histo
			} else {
				metrics[result.definition.kialiName] = result.metric
			}
		}
	}
	return Metrics{
		Metrics:    metrics,
		Histograms: histograms,
	}
}

func fetchRateRange(api v1.API, metricName, labels, grouping string, q *BaseMetricsQuery) *Metric {
	var query string
	// Example: round(sum(rate(my_counter{foo=bar}[5m])) by (baz), 0.001)
	if grouping == "" {
		query = fmt.Sprintf("sum(%s(%s%s[%s]))", q.RateFunc, metricName, labels, q.RateInterval)
	} else {
		query = fmt.Sprintf("sum(%s(%s%s[%s])) by (%s)", q.RateFunc, metricName, labels, q.RateInterval, grouping)
	}
	query = roundSignificant(query, 0.001)
	return fetchRange(api, query, q.Range)
}

func fetchHistogramRange(api v1.API, metricName, labels, grouping string, q *BaseMetricsQuery) Histogram {
	histogram := make(Histogram)

	// Note: the p8s queries are not run in parallel here, but they are at the caller's place.
	//	This is because we may not want to create too many threads in the lowest layer
	if q.Avg {
		groupingAvg := ""
		if grouping != "" {
			groupingAvg = fmt.Sprintf(" by (%s)", grouping)
		}
		// Average
		// Example: sum(rate(my_histogram_sum{foo=bar}[5m])) by (baz) / sum(rate(my_histogram_count{foo=bar}[5m])) by (baz)
		query := fmt.Sprintf("sum(rate(%s_sum%s[%s]))%s / sum(rate(%s_count%s[%s]))%s",
			metricName, labels, q.RateInterval, groupingAvg, metricName, labels, q.RateInterval, groupingAvg)
		query = roundSignificant(query, 0.001)
		log.Infof("Query: %s\n", query)
		histogram["avg"] = fetchRange(api, query, q.Range)
	}

	groupingQuantile := ""
	if grouping != "" {
		groupingQuantile = fmt.Sprintf(",%s", grouping)
	}
	for _, quantile := range q.Quantiles {
		// Example: round(histogram_quantile(0.5, sum(rate(my_histogram_bucket{foo=bar}[5m])) by (le,baz)), 0.001)
		query := fmt.Sprintf("histogram_quantile(%s, sum(rate(%s_bucket%s[%s])) by (le%s))",
			quantile, metricName, labels, q.RateInterval, groupingQuantile)
		query = roundSignificant(query, 0.001)
		histogram[quantile] = fetchRange(api, query, q.Range)
		log.Infof("Query: %s\n", query)
	}

	return histogram
}

func fetchTimestamp(api v1.API, query string, t time.Time) (model.Vector, error) {
	result, _, err := api.Query(context.Background(), query, t)
	if err != nil {
		return nil, err
	}
	switch result.Type() {
	case model.ValVector:
		return result.(model.Vector), nil
	}
	return nil, fmt.Errorf("Invalid query, vector expected: %s", query)
}

func fetchRange(api v1.API, query string, bounds v1.Range) *Metric {
	result, _, err := api.QueryRange(context.Background(), query, bounds)
	if err != nil {
		return &Metric{err: err}
	}
	switch result.Type() {
	case model.ValMatrix:
		return &Metric{Matrix: result.(model.Matrix)}
	}
	return &Metric{err: fmt.Errorf("Invalid query, matrix expected: %s", query)}
}

func replaceInvalidCharacters(metricName string) string {
	// See https://github.com/prometheus/prometheus/blob/master/util/strutil/strconv.go#L43
	return invalidLabelCharRE.ReplaceAllString(metricName, "_")
}

// getAllRequestRates retrieves traffic rates for requests entering, internal to, or exiting the namespace.
// Uses source telemetry unless working on the Istio namespace.
func getAllRequestRates(api v1.API, namespace string, queryTime time.Time, ratesInterval string) (model.Vector, error) {
	// traffic originating outside the namespace to destinations inside the namespace
	lbl := fmt.Sprintf(`destination_service_namespace="%s",source_workload_namespace!="%s"`, namespace, namespace)
	fromOutside, err := getRequestRatesForLabel(api, queryTime, lbl, ratesInterval)
	if err != nil {
		return model.Vector{}, err
	}
	// traffic originating inside the namespace to destinations inside or outside the namespace
	lbl = fmt.Sprintf(`source_workload_namespace="%s"`, namespace)
	fromInside, err := getRequestRatesForLabel(api, queryTime, lbl, ratesInterval)
	if err != nil {
		return model.Vector{}, err
	}
	// Merge results
	all := append(fromOutside, fromInside...)
	return all, nil
}

// getNamespaceServicesRequestRates retrieves traffic rates for requests entering or internal to the namespace.
// Uses source telemetry unless working on the Istio namespace.
func getNamespaceServicesRequestRates(api v1.API, namespace string, queryTime time.Time, ratesInterval string) (model.Vector, error) {
	// traffic for the namespace services
	lblNs := fmt.Sprintf(`destination_service_namespace="%s"`, namespace)
	ns, err := getRequestRatesForLabel(api, queryTime, lblNs, ratesInterval)
	if err != nil {
		return model.Vector{}, err
	}
	return ns, nil
}

func getServiceRequestRates(api v1.API, namespace, service string, queryTime time.Time, ratesInterval string) (model.Vector, error) {
	lbl := fmt.Sprintf(`destination_service_name="%s",destination_service_namespace="%s"`, service, namespace)
	in, err := getRequestRatesForLabel(api, queryTime, lbl, ratesInterval)
	if err != nil {
		return model.Vector{}, err
	}
	return in, nil
}

func getItemRequestRates(api v1.API, namespace, item, itemLabelSuffix string, queryTime time.Time, ratesInterval string) (model.Vector, model.Vector, error) {
	lblIn := fmt.Sprintf(`destination_workload_namespace="%s",destination_%s="%s"`, namespace, itemLabelSuffix, item)
	lblOut := fmt.Sprintf(`source_workload_namespace="%s",source_%s="%s"`, namespace, itemLabelSuffix, item)
	in, err := getRequestRatesForLabel(api, queryTime, lblIn, ratesInterval)
	if err != nil {
		return model.Vector{}, model.Vector{}, err
	}
	out, err := getRequestRatesForLabel(api, queryTime, lblOut, ratesInterval)
	if err != nil {
		return model.Vector{}, model.Vector{}, err
	}
	return in, out, nil
}

func getRequestRatesForLabel(api v1.API, time time.Time, labels, ratesInterval string) (model.Vector, error) {
	query := fmt.Sprintf("rate(istio_requests_total{%s}[%s])", labels, ratesInterval)
	promtimer := internalmetrics.GetPrometheusProcessingTimePrometheusTimer("Metrics-GetRequestRates")
	result, _, err := api.Query(context.Background(), query, time)
	if err != nil {
		return model.Vector{}, err
	}
	promtimer.ObserveDuration() // notice we only collect metrics for successful prom queries
	return result.(model.Vector), nil
}

// roundSignificant will output promQL that performs rounding only if the resulting value is significant, that is, higher than the requested precision
func roundSignificant(innerQuery string, precision float64) string {
	return fmt.Sprintf("round(%s, %f) > %f or %s", innerQuery, precision, precision, innerQuery)
}
