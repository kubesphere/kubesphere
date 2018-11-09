/*
Copyright 2018 The KubeSphere Authors.

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

package metrics

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"unicode"

	"github.com/golang/glog"
)

const (
	DefaultPageLimit = 5
	DefaultPage      = 1
)

type FormatedMetricDataWrapper struct {
	fmtMetricData FormatedMetricData
	by            func(p, q *map[string]interface{}) bool
}

func (wrapper FormatedMetricDataWrapper) Len() int {
	return len(wrapper.fmtMetricData.Result)
}

func (wrapper FormatedMetricDataWrapper) Less(i, j int) bool {
	return wrapper.by(&wrapper.fmtMetricData.Result[i], &wrapper.fmtMetricData.Result[j])
}

func (wrapper FormatedMetricDataWrapper) Swap(i, j int) {
	wrapper.fmtMetricData.Result[i], wrapper.fmtMetricData.Result[j] = wrapper.fmtMetricData.Result[j], wrapper.fmtMetricData.Result[i]
}

// sorted metric by ascending or descending order
func Sort(sortMetricName string, sortType string, fmtLevelMetric *FormatedLevelMetric, resourceType string) (*FormatedLevelMetric, int) {
	var maxLength = 0
	for _, metricItem := range fmtLevelMetric.Results {
		if metricItem.Data.ResultType == ResultTypeVector && metricItem.Status == MetricStatusSuccess {
			if maxLength < len(metricItem.Data.Result) {
				maxLength = len(metricItem.Data.Result)
			}
		}
	}

	if sortMetricName == "" {
		return fmtLevelMetric, maxLength
	}

	// default sort type is descending order
	if sortType == "" {
		sortType = ResultSortTypeDesc
	}

	var currentResourceMap = make(map[string]int)

	// indexMap store sorted index for each node/namespace/pod
	var indexMap = make(map[string]int)
	for _, metricItem := range fmtLevelMetric.Results {
		if metricItem.Data.ResultType == ResultTypeVector && metricItem.Status == MetricStatusSuccess {
			if metricItem.MetricName == sortMetricName {
				if sortType == ResultSortTypeAsce {
					// desc
					sort.Sort(FormatedMetricDataWrapper{metricItem.Data, func(p, q *map[string]interface{}) bool {
						value1 := (*p)[ResultItemValue].([]interface{})
						value2 := (*q)[ResultItemValue].([]interface{})
						v1, _ := strconv.ParseFloat(value1[len(value1)-1].(string), 64)
						v2, _ := strconv.ParseFloat(value2[len(value2)-1].(string), 64)
						return v1 < v2
					}})
				} else {
					// desc
					sort.Sort(FormatedMetricDataWrapper{metricItem.Data, func(p, q *map[string]interface{}) bool {
						value1 := (*p)[ResultItemValue].([]interface{})
						value2 := (*q)[ResultItemValue].([]interface{})
						v1, _ := strconv.ParseFloat(value1[len(value1)-1].(string), 64)
						v2, _ := strconv.ParseFloat(value2[len(value2)-1].(string), 64)
						return v1 > v2
					}})
				}

				i := 0
				for _, r := range metricItem.Data.Result {
					// for some reasons, 'metric' may not contain `resourceType` field
					// example: {"metric":{},"value":[1541142931.731,"3"]}
					k, exist := r[ResultItemMetric].(map[string]interface{})[resourceType]
					if exist {
						indexMap[k.(string)] = i
						i = i + 1
					}
				}
			}

			// iterator all metric to find max metricItems length
			for _, r := range metricItem.Data.Result {
				k, ok := r[ResultItemMetric].(map[string]interface{})[resourceType]
				if ok {
					currentResourceMap[k.(string)] = 1
				}
			}

		}
	}

	// sort other metric
	for i := 0; i < len(fmtLevelMetric.Results); i++ {
		re := fmtLevelMetric.Results[i]
		if re.MetricName != sortMetricName && re.Data.ResultType == ResultTypeVector && re.Status == MetricStatusSuccess {
			sortedMetric := make([]map[string]interface{}, len(indexMap))
			noneSortedMetric := make([]map[string]interface{}, 0)
			for j := 0; j < len(re.Data.Result); j++ {
				r := re.Data.Result[j]
				k, ok := r[ResultItemMetric].(map[string]interface{})[resourceType]
				if ok {
					index, exist := indexMap[k.(string)]
					if exist {
						sortedMetric[index] = r
					} else {
						noneSortedMetric = append(noneSortedMetric, r)
					}
				}
			}

			sortedMetric = append(sortedMetric, noneSortedMetric...)
			fmtLevelMetric.Results[i].Data.Result = sortedMetric
		}
	}

	return fmtLevelMetric, maxLength
}

func Page(pageNum string, limitNum string, fmtLevelMetric *FormatedLevelMetric, maxLength int) interface{} {
	// matrix type can not be sorted
	for _, metricItem := range fmtLevelMetric.Results {
		if metricItem.Data.ResultType != ResultTypeVector {
			return fmtLevelMetric
		}
	}

	var page = DefaultPage

	if pageNum != "" {
		p, err := strconv.Atoi(pageNum)
		if err != nil {
			glog.Errorln(err)
		} else {
			if p > 0 {
				page = p
			}
		}
	} else {
		// the default mode is none paging
		return fmtLevelMetric
	}

	var limit = DefaultPageLimit

	if limitNum != "" {
		l, err := strconv.Atoi(limitNum)
		if err != nil {
			glog.Errorln(err)
		} else {
			if l > 0 {
				limit = l
			}
		}
	}

	// the i page: [(page-1) * limit, (page) * limit - 1]
	start := (page - 1) * limit
	end := (page)*limit - 1

	for i := 0; i < len(fmtLevelMetric.Results); i++ {
		// only pageing when result type is `vector` and result status is `success`
		if fmtLevelMetric.Results[i].Data.ResultType != ResultTypeVector || fmtLevelMetric.Results[i].Status != MetricStatusSuccess {
			continue
		}
		resultLen := len(fmtLevelMetric.Results[i].Data.Result)
		if start >= resultLen {
			fmtLevelMetric.Results[i].Data.Result = nil
			continue
		}
		if end >= resultLen {
			end = resultLen - 1
		}
		slice := fmtLevelMetric.Results[i].Data.Result[start : end+1]
		fmtLevelMetric.Results[i].Data.Result = slice
	}

	allPage := int(math.Ceil(float64(maxLength) / float64(limit)))

	return &PagedFormatedLevelMetric{
		Message:     "paged",
		TotalPage:   allPage,
		CurrentPage: page,
		Metric:      *fmtLevelMetric,
	}
}

// maybe this function is time consuming
func ReformatJson(metric string, metricsName string, needDelParams ...string) *FormatedMetric {
	var formatMetric FormatedMetric
	err := json.Unmarshal([]byte(metric), &formatMetric)
	if err != nil {
		glog.Errorln("Unmarshal metric json failed", err)
	}
	if formatMetric.MetricName == "" {
		if metricsName != "" {
			formatMetric.MetricName = metricsName
		}
	}
	// retrive metrics success
	if formatMetric.Status == MetricStatusSuccess {
		result := formatMetric.Data.Result
		for _, res := range result {
			metric, exist := res[ResultItemMetric]
			metricMap := metric.(map[string]interface{})
			if exist {
				delete(metricMap, "__name__")
			}
			if len(needDelParams) > 0 {
				for _, p := range needDelParams {
					delete(metricMap, p)
				}
			}
		}
	}

	return &formatMetric
}

func ReformatNodeStatusField(nodeMetric *FormatedMetric) *FormatedMetric {
	metricCount := len(nodeMetric.Data.Result)
	for i := 0; i < metricCount; i++ {
		metric, exist := nodeMetric.Data.Result[i][ResultItemMetric]
		if exist {
			status, exist := metric.(map[string]interface{})[MetricStatus]
			if exist {
				status = UpperFirstLetter(status.(string))
				metric.(map[string]interface{})[MetricStatus] = status
			}
		}
	}
	return nodeMetric
}

func UpperFirstLetter(str string) string {
	for i, ch := range str {
		return string(unicode.ToUpper(ch)) + str[i+1:]
	}
	return ""
}
