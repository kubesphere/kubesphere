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

package monitoring

import (
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"math"
	"sort"
)

// TODO(huanggze): the id value is dependent of Prometheus label-value pair (i.e. label_kubesphere_io_workspace). We should regulate the naming convention.
const (
	IdentifierNode      = "node"
	IdentifierWorkspace = "label_kubesphere_io_workspace"
	IdentifierNamespace = "namespace"
	IdentifierWorkload  = "workload"
	IdentifierPod       = "pod"
	IdentifierContainer = "container"
	IdentifierPVC       = "persistentvolumeclaim"

	OrderAscending  = "asc"
	OrderDescending = "desc"
)

type wrapper struct {
	monitoring.MetricData
	by func(p, q *monitoring.MetricValue) bool
}

func (w wrapper) Len() int {
	return len(w.MetricValues)
}

func (w wrapper) Less(i, j int) bool {
	return w.by(&w.MetricValues[i], &w.MetricValues[j])
}

func (w wrapper) Swap(i, j int) {
	w.MetricValues[i], w.MetricValues[j] = w.MetricValues[j], w.MetricValues[i]
}

// The sortMetrics sorts a group of resources by a given metric
// Example:
//
// before sorting
// |------| Metric 1 | Metric 2 | Metric 3 |
// | ID a |     1     |     XL    |           |
// | ID b |     1     |     S     |           |
// | ID c |     3     |     M     |           |
//
// sort by metrics_2
// |------| Metric 1 | Metric 2 (asc) | Metric 3 |
// | ID a |     1     |        XL       |           |
// | ID c |     3     |        M        |           |
// | ID b |     1     |        S        |           |
//
// ranking can only be applied to instant query results, not range query
func (mo monitoringOperator) SortMetrics(raw v1alpha2.APIResponse, target, order, identifier string) (v1alpha2.APIResponse, int) {
	if target == "" || len(raw.Results) == 0 {
		return raw, -1
	}

	if order == "" {
		order = OrderDescending
	}

	var currentResourceMap = make(map[string]int)

	// resource-ordinal map
	var indexMap = make(map[string]int)
	i := 0

	for _, item := range raw.Results {
		if item.MetricType == monitoring.MetricTypeVector && item.Status == monitoring.StatusSuccess {
			if item.MetricName == target {
				if order == OrderAscending {
					sort.Sort(wrapper{item.MetricData, func(p, q *monitoring.MetricValue) bool {
						if p.Sample[1] == q.Sample[1] {
							return p.Metadata[identifier] < q.Metadata[identifier]
						}
						return p.Sample[1] < q.Sample[1]
					}})
				} else {
					sort.Sort(wrapper{item.MetricData, func(p, q *monitoring.MetricValue) bool {
						if p.Sample[1] == q.Sample[1] {
							return p.Metadata[identifier] > q.Metadata[identifier]
						}
						return p.Sample[1] > q.Sample[1]
					}})
				}

				for _, r := range item.MetricValues {
					// record the ordinal of resource to indexMap
					resourceName, exist := r.Metadata[identifier]
					if exist {
						if _, exist := indexMap[resourceName]; !exist {
							indexMap[resourceName] = i
							i = i + 1
						}
					}
				}
			}

			// get total number of rows
			for _, r := range item.MetricValues {
				k, ok := r.Metadata[identifier]
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

	// sort other metrics
	for i := 0; i < len(raw.Results); i++ {
		item := raw.Results[i]
		if item.MetricType == monitoring.MetricTypeVector && item.Status == monitoring.StatusSuccess {
			sortedMetric := make([]monitoring.MetricValue, len(indexMap))
			for j := 0; j < len(item.MetricValues); j++ {
				r := item.MetricValues[j]
				k, exist := r.Metadata[identifier]
				if exist {
					index, exist := indexMap[k]
					if exist {
						sortedMetric[index] = r
					}
				}
			}

			raw.Results[i].MetricValues = sortedMetric
		}
	}

	return raw, len(indexMap)
}

func (mo monitoringOperator) PageMetrics(raw v1alpha2.APIResponse, page, limit, rows int) v1alpha2.APIResponse {
	if page <= 0 || limit <= 0 || rows <= 0 || len(raw.Results) == 0 {
		return raw
	}

	// matrix type can not be sorted
	for _, item := range raw.Results {
		if item.MetricType != monitoring.MetricTypeVector {
			return raw
		}
	}

	// the i page: [(page-1) * limit, (page) * limit - 1]
	start := (page - 1) * limit
	end := (page)*limit - 1

	for i := 0; i < len(raw.Results); i++ {
		if raw.Results[i].MetricType != monitoring.MetricTypeVector || raw.Results[i].Status != monitoring.StatusSuccess {
			continue
		}
		resultLen := len(raw.Results[i].MetricValues)
		if start >= resultLen {
			raw.Results[i].MetricValues = nil
			continue
		}
		if end >= resultLen {
			end = resultLen - 1
		}
		slice := raw.Results[i].MetricValues[start : end+1]
		raw.Results[i].MetricValues = slice
	}

	raw.CurrentPage = page
	raw.TotalPage = int(math.Ceil(float64(rows) / float64(limit)))
	raw.TotalItem = rows
	return raw
}
