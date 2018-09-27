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

package hpa

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"horizontalpodautoscalers"}

	ws.Route(ws.GET(subPath).To(getHpa).Consumes("*/*").Metadata(restfulspec.KeyOpenAPITags, tags).Doc(
		"get horizontalpodautoscalers").Param(ws.PathParameter("namespace",
		"horizontalpodautoscalers's namespace").DataType("string")).Param(ws.PathParameter(
		"horizontalpodautoscaler", "horizontalpodautoscaler's name")).Writes(v1.HorizontalPodAutoscaler{}))
}

func getHpa(req *restful.Request, resp *restful.Response) {
	hpa := req.PathParameter("horizontalpodautoscaler")
	namespace := req.PathParameter("namespace")
	client := client.NewK8sClient()

	res, err := client.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(hpa, metaV1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			resp.WriteHeaderAndEntity(http.StatusOK, nil)
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, nil)
		}
	}

	resp.WriteEntity(res)
}
