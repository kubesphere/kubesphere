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
	identifier, order string
}

func (w wrapper) Len() int {
	return len(w.MetricValues)
}

func (w wrapper) Less(i, j int) bool {
	p := w.MetricValues[i]
	q := w.MetricValues[j]

	if p.Sample.Value() == q.Sample.Value() {
		return p.Metadata[w.identifier] < q.Metadata[w.identifier]
	}

	switch w.order {
	case OrderAscending:
		return p.Sample.Value() < q.Sample.Value()
	default:
		return p.Sample.Value() > q.Sample.Value()
	}
}

func (id wrapper) Swap(i, j int) {
	id.MetricValues[i], id.MetricValues[j] = id.MetricValues[j], id.MetricValues[i]
}

// SortMetrics sorts a group of resources by a given metric. Range query doesn't support ranking.
// Example:
//
// Before sorting:
// |  ID | Metric 1  |  Metric 2 |  Metric 3 |
// |  a  |     1     |     XL    |           |
// |  b  |     1     |     S     |           |
// |  c  |     3     |     M     |           |
//
// After sorting: target=metric_2, order=asc, identifier=id
// |  ID | Metric 1  |  Metric 2 (asc) |  Metric 3 |
// |  a  |     1     |        XL       |           |
// |  c  |     3     |        M        |           |
// |  b  |     1     |        S        |           |
func (raw *Metrics) Sort(target, order, identifier string) *Metrics {
	if target == "" || identifier == "" || len(raw.Results) == 0 {
		return raw
	}

	resourceSet := make(map[string]bool)    // resource set records possible values of the identifier
	resourceOrdinal := make(map[string]int) // resource-ordinal map

	ordinal := 0
	for _, item := range raw.Results {
		if item.MetricType != monitoring.MetricTypeVector || item.Error != "" {
			continue
		}

		if item.MetricName == target {
			sort.Sort(wrapper{
				MetricData: item.MetricData,
				identifier: identifier,
				order:      order,
			})

			for _, mv := range item.MetricValues {
				// Record ordinals in the final result
				v, ok := mv.Metadata[identifier]
				if ok && v != "" {
					resourceOrdinal[v] = ordinal
					ordinal++
				}
			}
		}

		// Add every unique identifier value to the set
		for _, mv := range item.MetricValues {
			v, ok := mv.Metadata[identifier]
			if ok && v != "" {
				resourceSet[v] = true
			}
		}
	}

	var resourceList []string
	for k := range resourceSet {
		resourceList = append(resourceList, k)
	}
	sort.Strings(resourceList)

	// Fill resource-ordinal map with resources never present in the target, and give them ordinals.
	for _, r := range resourceList {
		if _, ok := resourceOrdinal[r]; !ok {
			resourceOrdinal[r] = ordinal
			ordinal++
		}
	}

	// Sort metrics
	for i, item := range raw.Results {
		if item.MetricType != monitoring.MetricTypeVector || item.Error != "" {
			continue
		}

		sorted := make([]monitoring.MetricValue, len(resourceList))
		for _, mv := range item.MetricValues {
			v, ok := mv.Metadata[identifier]
			if ok && v != "" {
				ordinal, _ := resourceOrdinal[v]
				sorted[ordinal] = mv
			}
		}
		raw.Results[i].MetricValues = sorted
	}

	raw.CurrentPage = 1
	raw.TotalPages = 1
	raw.TotalItems = len(resourceList)
	return raw
}

func (raw *Metrics) Page(page, limit int) *Metrics {
	if page < 1 || limit < 1 || len(raw.Results) == 0 {
		return raw
	}

	start := (page - 1) * limit
	end := page * limit

	for i, item := range raw.Results {
		if item.MetricType != monitoring.MetricTypeVector || item.Error != "" {
			continue
		}

		total := len(item.MetricValues)
		if start >= total {
			raw.Results[i].MetricValues = nil
			continue
		}
		if end >= total {
			end = total
		}

		raw.Results[i].MetricValues = item.MetricValues[start:end]
	}

	raw.CurrentPage = page
	raw.TotalPages = int(math.Ceil(float64(raw.TotalItems) / float64(limit)))
	return raw
}
