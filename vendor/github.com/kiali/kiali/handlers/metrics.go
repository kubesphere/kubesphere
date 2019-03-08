package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
	"github.com/kiali/kiali/util"
)

func extractIstioMetricsQueryParams(r *http.Request, q *prometheus.IstioMetricsQuery, namespaceInfo *models.Namespace) error {
	q.FillDefaults()
	queryParams := r.URL.Query()
	if filters, ok := queryParams["filters[]"]; ok && len(filters) > 0 {
		q.Filters = filters
	}
	dir := queryParams.Get("direction")
	if dir != "" {
		if dir != "inbound" && dir != "outbound" {
			return errors.New("Bad request, query parameter 'direction' must be either 'inbound' or 'outbound'")
		}
		q.Direction = dir
	}
	requestProtocol := queryParams.Get("requestProtocol")
	if requestProtocol != "" {
		q.RequestProtocol = requestProtocol
	}
	reporter := queryParams.Get("reporter")
	if reporter != "" {
		if reporter != "source" && reporter != "destination" {
			return errors.New("Bad request, query parameter 'reporter' must be either 'source' or 'destination'")
		}
		q.Reporter = reporter
	}
	return extractBaseMetricsQueryParams(queryParams, &q.BaseMetricsQuery, namespaceInfo)
}

func extractCustomMetricsQueryParams(r *http.Request, q *prometheus.CustomMetricsQuery, namespaceInfo *models.Namespace) error {
	q.FillDefaults()
	queryParams := r.URL.Query()
	q.Version = queryParams.Get("version")
	op := queryParams.Get("rawDataAggregator")
	// Explicit white-listing operators to prevent any kind of injection
	// For a list of operators, see https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators
	if op == "sum" || op == "min" || op == "max" || op == "avg" || op == "stddev" || op == "stdvar" {
		q.RawDataAggregator = op
	}
	return extractBaseMetricsQueryParams(queryParams, &q.BaseMetricsQuery, namespaceInfo)
}

func extractBaseMetricsQueryParams(queryParams url.Values, q *prometheus.BaseMetricsQuery, namespaceInfo *models.Namespace) error {
	if rateIntervals, ok := queryParams["rateInterval"]; ok && len(rateIntervals) > 0 {
		// Only first is taken into consideration
		q.RateInterval = rateIntervals[0]
	}
	if rateFuncs, ok := queryParams["rateFunc"]; ok && len(rateFuncs) > 0 {
		// Only first is taken into consideration
		if rateFuncs[0] != "rate" && rateFuncs[0] != "irate" {
			// Bad request
			return errors.New("Bad request, query parameter 'rateFunc' must be either 'rate' or 'irate'")
		}
		q.RateFunc = rateFuncs[0]
	}
	if queryTimes, ok := queryParams["queryTime"]; ok && len(queryTimes) > 0 {
		if num, err := strconv.ParseInt(queryTimes[0], 10, 64); err == nil {
			q.End = time.Unix(num, 0)
		} else {
			// Bad request
			return errors.New("Bad request, cannot parse query parameter 'queryTime'")
		}
	}
	if durations, ok := queryParams["duration"]; ok && len(durations) > 0 {
		if num, err := strconv.ParseInt(durations[0], 10, 64); err == nil {
			duration := time.Duration(num) * time.Second
			q.Start = q.End.Add(-duration)
		} else {
			// Bad request
			return errors.New("Bad request, cannot parse query parameter 'duration'")
		}
	}
	if steps, ok := queryParams["step"]; ok && len(steps) > 0 {
		if num, err := strconv.Atoi(steps[0]); err == nil {
			q.Step = time.Duration(num) * time.Second
		} else {
			// Bad request
			return errors.New("Bad request, cannot parse query parameter 'step'")
		}
	}
	if quantiles, ok := queryParams["quantiles[]"]; ok && len(quantiles) > 0 {
		for _, quantile := range quantiles {
			f, err := strconv.ParseFloat(quantile, 64)
			if err != nil {
				// Non parseable quantile
				return errors.New("Bad request, cannot parse query parameter 'quantiles', float expected")
			}
			if f < 0 || f > 1 {
				return errors.New("Bad request, invalid quantile(s): should be between 0 and 1")
			}
		}
		q.Quantiles = quantiles
	}
	if avgFlags, ok := queryParams["avg"]; ok && len(avgFlags) > 0 {
		if avgFlag, err := strconv.ParseBool(avgFlags[0]); err == nil {
			q.Avg = avgFlag
		} else {
			// Bad request
			return errors.New("Bad request, cannot parse query parameter 'avg'")
		}
	}
	if lbls, ok := queryParams["byLabels[]"]; ok && len(lbls) > 0 {
		q.ByLabels = lbls
	}

	// If needed, adjust interval -- Make sure query won't fetch data before the namespace creation
	intervalStartTime, err := util.GetStartTimeForRateInterval(q.End, q.RateInterval)
	if err != nil {
		return err
	}
	if intervalStartTime.Before(namespaceInfo.CreationTimestamp) {
		q.RateInterval = fmt.Sprintf("%ds", int(q.End.Sub(namespaceInfo.CreationTimestamp).Seconds()))
		intervalStartTime = namespaceInfo.CreationTimestamp
		log.Debugf("[extractMetricsQueryParams] Interval set to: %v", q.RateInterval)
	}
	// If needed, adjust query start time (bound to namespace creation time)
	log.Debugf("[extractMetricsQueryParams] Requested query start time: %v", q.Start)
	intervalDuration := q.End.Sub(intervalStartTime)
	allowedStart := namespaceInfo.CreationTimestamp.Add(intervalDuration)
	if q.Start.Before(allowedStart) {
		q.Start = allowedStart
		log.Debugf("[extractMetricsQueryParams] Query start time set to: %v", q.Start)

		if q.Start.After(q.End) {
			// This means that the query range does not fall in the range
			// of life of the namespace. So, there are no metrics to query.
			log.Debugf("[extractMetricsQueryParams] Query end time = %v; not querying metrics.", q.End)
			return errors.New("After checks, query start time is after end time")
		}
	}

	// Adjust start & end times to be a multiple of step
	stepInSecs := int64(q.Step.Seconds())
	q.Start = time.Unix((q.Start.Unix()/stepInSecs)*stepInSecs, 0)
	return nil
}
