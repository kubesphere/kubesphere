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
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/storage"
	"net/http"
)

type ScMetricsItem struct {
	Name    string             `json:"name"`
	Metrics *storage.ScMetrics `json:"metrics"`
}

// Get StorageClass item
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/{storageclass}/metrics"
func GetScMetrics(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")

	metrics, err := storage.GetScMetrics(scName)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	result := ScMetricsItem{
		Name: scName, Metrics: metrics,
	}

	response.WriteAsJson(result)
}

// Get StorageClass item list
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/metrics"
func GetScMetricsList(request *restful.Request, response *restful.Response) {
	scList, err := storage.GetScList()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	// Set return value
	items := make([]ScMetricsItem, 0)

	for _, v := range scList {
		metrics, err := storage.GetScMetrics(v.GetName())

		if err != nil {
			response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}

		item := ScMetricsItem{
			Name: v.GetName(), Metrics: metrics,
		}

		items = append(items, item)
	}

	response.WriteAsJson(items)
}
