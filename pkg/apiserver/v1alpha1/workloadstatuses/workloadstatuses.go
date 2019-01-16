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

package workloadstatuses

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/status"
)

func Route(ws *restful.WebService) {
	tags := []string{"workloadStatus"}

	ws.Route(ws.GET("/workloadstatuses").
		Doc("get abnormal workloads' count of whole cluster").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(getClusterResourceStatus))
	ws.Route(ws.GET("/namespaces/{namespace}/workloadstatuses").
		Doc("get abnormal workloads' count of specified namespace").
		Param(ws.PathParameter("namespace", "the name of namespace").
			DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(getNamespacesResourceStatus))

}

func getClusterResourceStatus(req *restful.Request, resp *restful.Response) {
	res, err := status.GetClusterResourceStatus()
	if errors.HandlerError(err, resp) {
		return
	}
	resp.WriteAsJson(res)
}

func getNamespacesResourceStatus(req *restful.Request, resp *restful.Response) {
	res, err := status.GetNamespacesResourceStatus(req.PathParameter("namespace"))
	if errors.HandlerError(err, resp) {
		return
	}
	resp.WriteAsJson(res)
}
