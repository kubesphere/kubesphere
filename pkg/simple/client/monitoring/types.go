/*
Copyright 2020 KubeSphere Authors

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

package monitoring

import (
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/jszwec/csvutil"
	"strconv"
	"strings"
	"time"
)

const (
	MetricTypeMatrix = "matrix"
	MetricTypeVector = "vector"
)

type Metadata struct {
	Metric string `json:"metric,omitempty" description:"metric name"`
	Type   string `json:"type,omitempty" description:"metric type"`
	Help   string `json:"help,omitempty" description:"metric description"`
}

type Metric struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum" csv:"metric_name"`
	MetricData `json:"data,omitempty" description:"actual metric result"`
	Error      string `json:"error,omitempty" csv:"-"`
}

type MetricValues []MetricValue

func (m MetricValues) MarshalCSV() ([]byte, error) {

	var ret []string
	for _, v := range m {
		tmp, err := v.MarshalCSV()
		if err != nil {
			return nil, err
		}

		ret = append(ret, string(tmp))
	}

	return []byte(strings.Join(ret, "||")), nil
}

type MetricData struct {
	MetricType   string `json:"resultType,omitempty" description:"result type, one of matrix, vector" csv:"metric_type"`
	MetricValues `json:"result,omitempty" description:"metric data including labels, time series and values" csv:"metric_values"`
}

func (m MetricData) MarshalCSV() ([]byte, error) {
	var ret []byte

	for _, v := range m.MetricValues {
		tmp, err := csvutil.Marshal(&v)
		if err != nil {
			return nil, err
		}

		ret = append(ret, tmp...)
	}

	return ret, nil
}

// The first element is the timestamp, the second is the metric value.
// eg, [1585658599.195, 0.528]
type Point [2]float64

type MetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"time series labels"`
	// The type of Point is a float64 array with fixed length of 2.
	// So Point will always be initialized as [0, 0], rather than nil.
	// To allow empty Sample, we should declare Sample to type *Point
	Sample         *Point        `json:"value,omitempty" description:"time series, values of vector type"`
	Series         []Point       `json:"values,omitempty" description:"time series, values of matrix type"`
	ExportSample   *ExportPoint  `json:"exported_value,omitempty" description:"exported time series, values of vector type"`
	ExportedSeries []ExportPoint `json:"exported_values,omitempty" description:"exported time series, values of matrix type"`

	MinValue     float64 `json:"min_value" description:"minimum value from monitor points"`
	MaxValue     float64 `json:"max_value" description:"maximum value from monitor points"`
	AvgValue     float64 `json:"avg_value" description:"average value from monitor points"`
	SumValue     float64 `json:"sum_value" description:"sum value from monitor points"`
	Fee          float64 `json:"fee" description:"resource fee"`
	ResourceUnit string  `json:"resource_unit"`
	CurrencyUnit string  `json:"currency_unit"`
}

func (mv MetricValue) MarshalCSV() ([]byte, error) {
	// metric value format:
	// 	target,stats value(include fees),exported_value,exported_values
	// for example:
	// 	{workspace:demo-ws},,2021-02-23 01:00:00 AM 0|2021-02-23 02:00:00 AM 0|...
	var metricValueCSVTemplate = "{%s},unit:%s|min:%.3f|max:%.3f|avg:%.3f|sum:%.3f|fee:%.2f %s,%s,%s"

	var targetList []string
	for k, v := range mv.Metadata {
		targetList = append(targetList, fmt.Sprintf("%s=%s", k, v))
	}

	exportedSampleStr := ""
	if mv.ExportSample != nil {
		exportedSampleStr = mv.ExportSample.Format()
	}

	exportedSeriesStrList := []string{}
	for _, v := range mv.ExportedSeries {
		exportedSeriesStrList = append(exportedSeriesStrList, v.Format())
	}

	return []byte(fmt.Sprintf(metricValueCSVTemplate,
		strings.Join(targetList, "|"),
		mv.ResourceUnit,
		mv.MinValue,
		mv.MaxValue,
		mv.AvgValue,
		mv.SumValue,
		mv.Fee,
		mv.CurrencyUnit,
		exportedSampleStr,
		exportedSeriesStrList)), nil
}

func (mv *MetricValue) TransferToExportedMetricValue() {

	if mv.Sample != nil {
		sample := mv.Sample.transferToExported()
		mv.ExportSample = &sample
		mv.Sample = nil
	}

	for _, item := range mv.Series {
		mv.ExportedSeries = append(mv.ExportedSeries, item.transferToExported())
	}
	mv.Series = nil

	return
}

func (p Point) Timestamp() float64 {
	return p[0]
}

func (p Point) Value() float64 {
	return p[1]
}

func (p Point) transferToExported() ExportPoint {
	return ExportPoint{p[0], p[1]}
}

// MarshalJSON implements json.Marshaler. It will be called when writing JSON to HTTP response
// Inspired by prometheus/client_golang
func (p Point) MarshalJSON() ([]byte, error) {
	t, err := jsoniter.Marshal(p.Timestamp())
	if err != nil {
		return nil, err
	}
	v, err := jsoniter.Marshal(strconv.FormatFloat(p.Value(), 'f', -1, 64))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[%s,%s]", t, v)), nil
}

// UnmarshalJSON implements json.Unmarshaler. This is for unmarshaling test data.
func (p *Point) UnmarshalJSON(b []byte) error {
	var v []interface{}
	if err := jsoniter.Unmarshal(b, &v); err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	if len(v) != 2 {
		return errors.New("unsupported array length")
	}

	ts, ok := v[0].(float64)
	if !ok {
		return errors.New("failed to unmarshal [timestamp]")
	}
	valstr, ok := v[1].(string)
	if !ok {
		return errors.New("failed to unmarshal [value]")
	}
	valf, err := strconv.ParseFloat(valstr, 64)
	if err != nil {
		return err
	}

	p[0] = ts
	p[1] = valf
	return nil
}

type ExportPoint [2]float64

func (p ExportPoint) Timestamp() string {
	return time.Unix(int64(p[0]), 0).Format("2006-01-02 03:04:05 PM")
}

func (p ExportPoint) Value() float64 {
	return p[1]
}

func (p ExportPoint) Format() string {
	return p.Timestamp() + " " + strconv.FormatFloat(p.Value(), 'f', -1, 64)
}
