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

package workloadstatus

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"workloadStatus"}

	ws.Route(ws.GET(subPath).Doc("get abnormal workloads' count of whole cluster").Metadata(restfulspec.KeyOpenAPITags, tags).To(getClusterStatus).Produces(restful.MIME_JSON))
	ws.Route(ws.GET(subPath+"/namespaces/{namespace}").Doc("get abnormal workloads' count of specified namespace").Param(ws.PathParameter("namespace",
		"the name of namespace").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(getNamespaceStatus).Produces(restful.MIME_JSON))

}

func getClusterStatus(req *restful.Request, resp *restful.Response) {
	res, err := models.GetClusterResourceStatus()
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	}
	resp.WriteEntity(res)
}

func getNamespaceStatus(req *restful.Request, resp *restful.Response) {
	res, err := models.GetNamespacesResourceStatus(req.PathParameter("namespace"))
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	}
	resp.WriteEntity(res)
}
