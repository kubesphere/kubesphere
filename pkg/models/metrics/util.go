/*

 Copyright 2019 The KubeSphere Authors.

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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"kubesphere.io/kubesphere/pkg/informers"
	"math"
	"sort"
	"strconv"

	"runtime/debug"
)

const (
	DefaultPageLimit = 5
	DefaultPage      = 1

	ResultTypeVector             = "vector"
	ResultTypeMatrix             = "matrix"
	MetricStatusSuccess          = "success"
	ResultItemMetricResourceName = "resource_name"
	ResultSortTypeDesc           = "desc"
	ResultSortTypeAsc            = "asc"
)

type FormatedMetricDataWrapper struct {
	fmtMetricData v1alpha2.QueryResult
	by            func(p, q *v1alpha2.QueryValue) bool
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
func (rawMetrics *Response) SortBy(sortMetricName string, sortType string) (*Response, int) {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorln(err)
			debug.PrintStack()
		}
	}()

	if sortMetricName == "" || rawMetrics == nil {
		return rawMetrics, -1
	}

	// default sort type is descending order
	if sortType == "" {
		sortType = ResultSortTypeDesc
	}

	var currentResourceMap = make(map[string]int)

	// {<Resource Name>: <Ordering>}
	var indexMap = make(map[string]int)
	i := 0

	// each metricItem is the result for a specific metric name
	// so we find the metricItem with sortMetricName, and sort it
	for _, metricItem := range rawMetrics.Results {
		// only vector type result can be sorted
		if metricItem.Data.ResultType == ResultTypeVector && metricItem.Status == MetricStatusSuccess {
			if metricItem.MetricName == sortMetricName {
				if sortType == ResultSortTypeAsc {
					// asc
					sort.Sort(FormatedMetricDataWrapper{metricItem.Data, func(p, q *v1alpha2.QueryValue) bool {
						value1 := p.Value
						value2 := q.Value
						v1, _ := strconv.ParseFloat(value1[len(value1)-1].(string), 64)
						v2, _ := strconv.ParseFloat(value2[len(value2)-1].(string), 64)
						if v1 == v2 {
							resourceName1 := p.Metric[ResultItemMetricResourceName]
							resourceName2 := q.Metric[ResultItemMetricResourceName]
							return resourceName1 < resourceName2
						}

						return v1 < v2
					}})
				} else {
					// desc
					sort.Sort(FormatedMetricDataWrapper{metricItem.Data, func(p, q *v1alpha2.QueryValue) bool {
						value1 := p.Value
						value2 := q.Value
						v1, _ := strconv.ParseFloat(value1[len(value1)-1].(string), 64)
						v2, _ := strconv.ParseFloat(value2[len(value2)-1].(string), 64)

						if v1 == v2 {
							resourceName1 := p.Metric[ResultItemMetricResourceName]
							resourceName2 := q.Metric[ResultItemMetricResourceName]
							return resourceName1 > resourceName2
						}

						return v1 > v2
					}})
				}

				for _, r := range metricItem.Data.Result {
					// record the ordering of resource_name to indexMap
					// example: {"metric":{ResultItemMetricResourceName: "Deployment:xxx"},"value":[1541142931.731,"3"]}
					resourceName, exist := r.Metric[ResultItemMetricResourceName]
					if exist {
						if _, exist := indexMap[resourceName]; !exist {
							indexMap[resourceName] = i
							i = i + 1
						}
					}
				}
			}

			// iterator all metric to find max metricItems length
			for _, r := range metricItem.Data.Result {
				k, ok := r.Metric[ResultItemMetricResourceName]
				if ok {
					currentResourceMap[k] = 1
				}
			}

		}
	}

	var keys []string
	for k := range currentResourceMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, resource := range keys {
		if _, exist := indexMap[resource]; !exist {
			indexMap[resource] = i
			i = i + 1
		}
	}

	// sort other metric
	for i := 0; i < len(rawMetrics.Results); i++ {
		re := rawMetrics.Results[i]
		if re.Data.ResultType == ResultTypeVector && re.Status == MetricStatusSuccess {
			sortedMetric := make([]v1alpha2.QueryValue, len(indexMap))
			for j := 0; j < len(re.Data.Result); j++ {
				r := re.Data.Result[j]
				k, exist := r.Metric[ResultItemMetricResourceName]
				if exist {
					index, exist := indexMap[k]
					if exist {
						sortedMetric[index] = r
					}
				}
			}

			rawMetrics.Results[i].Data.Result = sortedMetric
		}
	}

	return rawMetrics, len(indexMap)
}

func (fmtLevelMetric *Response) Page(pageNum string, limitNum string, maxLength int) *Response {
	if maxLength <= 0 {
		return fmtLevelMetric
	}
	// matrix type can not be sorted
	for _, metricItem := range fmtLevelMetric.Results {
		// if metric reterieved field, resultType: ""
		if metricItem.Data.ResultType == ResultTypeMatrix {
			return fmtLevelMetric
		}
	}

	var page = DefaultPage

	if pageNum != "" {
		p, err := strconv.Atoi(pageNum)
		if err != nil {
			klog.Errorln(err)
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
			klog.Errorln(err)
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

	// add page fields
	fmtLevelMetric.CurrentPage = page
	fmtLevelMetric.TotalItem = maxLength
	fmtLevelMetric.TotalPage = allPage

	return fmtLevelMetric
}

func getNodeAddressAndRole(nodeName string) (string, string) {
	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	node, err := nodeLister.Get(nodeName)
	if err != nil {
		return "", ""
	}

	var addr string
	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			addr = address.Address
			break
		}
	}

	role := "node"
	_, exists := node.Labels["node-role.kubernetes.io/master"]
	if exists {
		role = "master"
	}
	return addr, role
}

func getNodeName(nodeIp string) string {
	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	nodes, _ := nodeLister.List(labels.Everything())

	for _, node := range nodes {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" && address.Address == nodeIp {
				return node.Name
			}
		}
	}

	return ""
}
