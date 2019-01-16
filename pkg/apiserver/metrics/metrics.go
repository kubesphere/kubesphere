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
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/storage"
)

func V1Alpha2(ws *restful.WebService) {
	tags := []string{"Metrics"}
	ws.Route(ws.GET("/storageclasses/{storageclass}").
		To(getScMetrics).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Param(ws.PathParameter("storageclass", "storageclass's name")).
		Writes(ScMetricsItem{}))

	ws.Route(ws.GET("/storageclasses").
		To(getScMetricsList).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Writes(ScMetricsItem{}))
}

type ScMetricsItem struct {
	Name    string             `json:"name"`
	Metrics *storage.ScMetrics `json:"metrics"`
}

// Get StorageClass item
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/{storageclass}/metrics"
func getScMetrics(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")

	metrics, err := storage.GetScMetrics(scName)

	if errors.HandleError(err, response) {
		return
	}

	result := ScMetricsItem{
		Name: scName, Metrics: metrics,
	}

	response.WriteAsJson(result)
}

// Get StorageClass item list
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/metrics"
func getScMetricsList(request *restful.Request, response *restful.Response) {
	scList, err := storage.GetScList()

	if errors.HandleError(err, response) {
		return
	}

	// Set return value
	items := make([]ScMetricsItem, 0)

	for _, v := range scList {
		metrics, err := storage.GetScMetrics(v.GetName())

		if errors.HandleError(err, response) {
			return
		}

		item := ScMetricsItem{
			Name: v.GetName(), Metrics: metrics,
		}

		items = append(items, item)
	}

	response.WriteAsJson(items)
}
