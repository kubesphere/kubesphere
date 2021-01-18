/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	"k8s.io/component-base/metrics"
)

var (
	// MetricNameLabel is label under which model.Sample stores metric name
	MetricNameLabel model.LabelName = model.MetricNameLabel
	// QuantileLabel is label under which model.Sample stores latency quantile value
	QuantileLabel model.LabelName = model.QuantileLabel
)

// Metrics is generic metrics for other specific metrics
type Metrics map[string]model.Samples

// Equal returns true if all metrics are the same as the arguments.
func (m *Metrics) Equal(o Metrics) bool {
	leftKeySet := []string{}
	rightKeySet := []string{}
	for k := range *m {
		leftKeySet = append(leftKeySet, k)
	}
	for k := range o {
		rightKeySet = append(rightKeySet, k)
	}
	if !reflect.DeepEqual(leftKeySet, rightKeySet) {
		return false
	}
	for _, k := range leftKeySet {
		if !(*m)[k].Equal(o[k]) {
			return false
		}
	}
	return true
}

// NewMetrics returns new metrics which are initialized.
func NewMetrics() Metrics {
	result := make(Metrics)
	return result
}

// ParseMetrics parses Metrics from data returned from prometheus endpoint
func ParseMetrics(data string, output *Metrics) error {
	dec := expfmt.NewDecoder(strings.NewReader(data), expfmt.FmtText)
	decoder := expfmt.SampleDecoder{
		Dec:  dec,
		Opts: &expfmt.DecodeOptions{},
	}

	for {
		var v model.Vector
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				// Expected loop termination condition.
				return nil
			}
			continue
		}
		for _, metric := range v {
			name := string(metric.Metric[model.MetricNameLabel])
			(*output)[name] = append((*output)[name], metric)
		}
	}
}

// ExtractMetricSamples parses the prometheus metric samples from the input string.
func ExtractMetricSamples(metricsBlob string) ([]*model.Sample, error) {
	dec := expfmt.NewDecoder(strings.NewReader(metricsBlob), expfmt.FmtText)
	decoder := expfmt.SampleDecoder{
		Dec:  dec,
		Opts: &expfmt.DecodeOptions{},
	}

	var samples []*model.Sample
	for {
		var v model.Vector
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				// Expected loop termination condition.
				return samples, nil
			}
			return nil, err
		}
		samples = append(samples, v...)
	}
}

// PrintSample returns formated representation of metric Sample
func PrintSample(sample *model.Sample) string {
	buf := make([]string, 0)
	// Id is a VERY special label. For 'normal' container it's useless, but it's necessary
	// for 'system' containers (e.g. /docker-daemon, /kubelet, etc.). We know if that's the
	// case by checking if there's a label "kubernetes_container_name" present. It's hacky
	// but it works...
	_, normalContainer := sample.Metric["kubernetes_container_name"]
	for k, v := range sample.Metric {
		if strings.HasPrefix(string(k), "__") {
			continue
		}

		if string(k) == "id" && normalContainer {
			continue
		}
		buf = append(buf, fmt.Sprintf("%v=%v", string(k), v))
	}
	return fmt.Sprintf("[%v] = %v", strings.Join(buf, ","), sample.Value)
}

// ComputeHistogramDelta computes the change in histogram metric for a selected label.
// Results are stored in after samples
func ComputeHistogramDelta(before, after model.Samples, label model.LabelName) {
	beforeSamplesMap := make(map[string]*model.Sample)
	for _, bSample := range before {
		beforeSamplesMap[makeKey(bSample.Metric[label], bSample.Metric["le"])] = bSample
	}
	for _, aSample := range after {
		if bSample, found := beforeSamplesMap[makeKey(aSample.Metric[label], aSample.Metric["le"])]; found {
			aSample.Value = aSample.Value - bSample.Value
		}
	}
}

func makeKey(a, b model.LabelValue) string {
	return string(a) + "___" + string(b)
}

// GetMetricValuesForLabel returns value of metric for a given dimension
func GetMetricValuesForLabel(ms Metrics, metricName, label string) map[string]int64 {
	samples, found := ms[metricName]
	result := make(map[string]int64, len(samples))
	if !found {
		return result
	}
	for _, sample := range samples {
		count := int64(sample.Value)
		dimensionName := string(sample.Metric[model.LabelName(label)])
		result[dimensionName] = count
	}
	return result
}

// ValidateMetrics verifies if every sample of metric has all expected labels
func ValidateMetrics(metrics Metrics, metricName string, expectedLabels ...string) error {
	samples, ok := metrics[metricName]
	if !ok {
		return fmt.Errorf("metric %q was not found in metrics", metricName)
	}
	for _, sample := range samples {
		for _, l := range expectedLabels {
			if _, ok := sample.Metric[model.LabelName(l)]; !ok {
				return fmt.Errorf("metric %q is missing label %q, sample: %q", metricName, l, sample.String())
			}
		}
	}
	return nil
}

// Histogram wraps prometheus histogram DTO (data transfer object)
type Histogram struct {
	*dto.Histogram
}

// GetHistogramFromGatherer collects a metric from a gatherer implementing k8s.io/component-base/metrics.Gatherer interface.
// Used only for testing purposes where we need to gather metrics directly from a running binary (without metrics endpoint).
func GetHistogramFromGatherer(gatherer metrics.Gatherer, metricName string) (Histogram, error) {
	var metricFamily *dto.MetricFamily
	m, err := gatherer.Gather()
	if err != nil {
		return Histogram{}, err
	}
	for _, mFamily := range m {
		if mFamily.Name != nil && *mFamily.Name == metricName {
			metricFamily = mFamily
			break
		}
	}

	if metricFamily == nil {
		return Histogram{}, fmt.Errorf("Metric %q not found", metricName)
	}

	if metricFamily.GetMetric() == nil {
		return Histogram{}, fmt.Errorf("Metric %q is empty", metricName)
	}

	if len(metricFamily.GetMetric()) == 0 {
		return Histogram{}, fmt.Errorf("Metric %q is empty", metricName)
	}

	return Histogram{
		// Histograms are stored under the first index (based on observation).
		// Given there's only one histogram registered per each metric name, accessing
		// the first index is sufficient.
		metricFamily.GetMetric()[0].GetHistogram(),
	}, nil
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}

// Bucket of a histogram
type bucket struct {
	upperBound float64
	count      float64
}

func bucketQuantile(q float64, buckets []bucket) float64 {
	if q < 0 {
		return math.Inf(-1)
	}
	if q > 1 {
		return math.Inf(+1)
	}

	if len(buckets) < 2 {
		return math.NaN()
	}

	rank := q * buckets[len(buckets)-1].count
	b := sort.Search(len(buckets)-1, func(i int) bool { return buckets[i].count >= rank })

	if b == 0 {
		return buckets[0].upperBound * (rank / buckets[0].count)
	}

	// linear approximation of b-th bucket
	brank := rank - buckets[b-1].count
	bSize := buckets[b].upperBound - buckets[b-1].upperBound
	bCount := buckets[b].count - buckets[b-1].count

	return buckets[b-1].upperBound + bSize*(brank/bCount)
}

// Quantile computes q-th quantile of a cumulative histogram.
// It's expected the histogram is valid (by calling Validate)
func (hist *Histogram) Quantile(q float64) float64 {
	buckets := []bucket{}

	for _, bckt := range hist.Bucket {
		buckets = append(buckets, bucket{
			count:      float64(*bckt.CumulativeCount),
			upperBound: *bckt.UpperBound,
		})
	}

	// bucketQuantile expects the upper bound of the last bucket to be +inf
	// buckets[len(buckets)-1].upperBound = math.Inf(+1)

	return bucketQuantile(q, buckets)
}

// Average computes histogram's average value
func (hist *Histogram) Average() float64 {
	return *hist.SampleSum / float64(*hist.SampleCount)
}

// Clear clears all fields of the wrapped histogram
func (hist *Histogram) Clear() {
	if hist.SampleCount != nil {
		*hist.SampleCount = 0
	}
	if hist.SampleSum != nil {
		*hist.SampleSum = 0
	}
	for _, b := range hist.Bucket {
		if b.CumulativeCount != nil {
			*b.CumulativeCount = 0
		}
	}
}

// Validate makes sure the wrapped histogram has all necessary fields set and with valid values.
func (hist *Histogram) Validate() error {
	if hist.SampleCount == nil || *hist.SampleCount == 0 {
		return fmt.Errorf("nil or empty histogram SampleCount")
	}

	if hist.SampleSum == nil || *hist.SampleSum == 0 {
		return fmt.Errorf("nil or empty histogram SampleSum")
	}

	for _, bckt := range hist.Bucket {
		if bckt == nil {
			return fmt.Errorf("empty histogram bucket")
		}
		if bckt.UpperBound == nil || *bckt.UpperBound < 0 {
			return fmt.Errorf("nil or negative histogram bucket UpperBound")
		}
	}

	return nil
}

// GetGaugeMetricValue extract metric value from GaugeMetric
func GetGaugeMetricValue(m metrics.GaugeMetric) (float64, error) {
	metricProto := &dto.Metric{}
	if err := m.Write(metricProto); err != nil {
		return 0, fmt.Errorf("Error writing m: %v", err)
	}
	return metricProto.Gauge.GetValue(), nil
}

// GetCounterMetricValue extract metric value from CounterMetric
func GetCounterMetricValue(m metrics.CounterMetric) (float64, error) {
	metricProto := &dto.Metric{}
	if err := m.(metrics.Metric).Write(metricProto); err != nil {
		return 0, fmt.Errorf("Error writing m: %v", err)
	}
	return metricProto.Counter.GetValue(), nil
}

// GetHistogramMetricValue extract sum of all samples from ObserverMetric
func GetHistogramMetricValue(m metrics.ObserverMetric) (float64, error) {
	metricProto := &dto.Metric{}
	if err := m.(metrics.Metric).Write(metricProto); err != nil {
		return 0, fmt.Errorf("Error writing m: %v", err)
	}
	return metricProto.Histogram.GetSampleSum(), nil
}
