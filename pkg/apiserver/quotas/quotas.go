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

package quotas

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/errors"

	"kubesphere.io/kubesphere/pkg/models/quotas"
)

func V1Alpha2(ws *restful.WebService) {

	tags := []string{"Quotas"}

	ws.Route(ws.GET("/quotas").
		To(getClusterQuotas).
		Doc("get whole cluster's resource usage").
		Writes(quotas.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/quotas").
		Doc("get specified namespace's resource quota and usage").
		Param(ws.PathParameter("namespace", "namespace's name").
			DataType("string")).
		Writes(quotas.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(getNamespaceQuotas))

}

func getNamespaceQuotas(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	quota, err := quotas.GetNamespaceQuotas(namespace)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(quota)
}

func getClusterQuotas(req *restful.Request, resp *restful.Response) {
	quota, err := quotas.GetClusterQuotas()

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(quota)
}
