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

package hpa

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/api/autoscaling/v1"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/hpa"
)

func V1Alpha2(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/horizontalpodautoscalers/{horizontalpodautoscaler}").
		To(getHpa).
		Metadata(restfulspec.KeyOpenAPITags, []string{"hpa"}).
		Doc("get horizontalpodautoscalers").
		Param(ws.PathParameter("namespace", "horizontalpodautoscalers's namespace").
			DataType("string")).
		Param(ws.PathParameter("horizontalpodautoscaler", "horizontalpodautoscaler's name")).
		Writes(v1.HorizontalPodAutoscaler{}))
}

func getHpa(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("horizontalpodautoscaler")
	namespace := req.PathParameter("namespace")

	result, err := hpa.GetHPA(namespace, name)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(result)
}
