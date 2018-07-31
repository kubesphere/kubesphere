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

package resources

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"resources"}

	ws.Route(ws.GET(subPath+"/{resource}").To(handleResource).Produces(restful.MIME_JSON).Metadata(restfulspec.KeyOpenAPITags, tags).Doc("Get resource" +
		" list").Param(ws.PathParameter("resource", "resource name").DataType(
		"string")).Param(ws.QueryParameter("conditions", "search "+
		"conditions").DataType("string")).Param(ws.QueryParameter("paging",
		"support paging function").DataType("string")).Writes(models.ResourceList{}))

}

func handleResource(req *restful.Request, resp *restful.Response) {

	resource := req.PathParameter("resource")
	if resource == "applications" {
		handleApplication(req, resp)
		return
	}
	conditions := req.QueryParameter("conditions")
	paging := req.QueryParameter("paging")
	res, err := models.ListResource(resource, conditions, paging)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(res)
}

func handleApplication(req *restful.Request, resp *restful.Response) {
	//searchWord := req.QueryParameter("search-word")
	paging := req.QueryParameter("paging")
	clusterId := req.QueryParameter("cluster_id")
	runtimeId := req.QueryParameter("runtime_id")
	conditions := req.QueryParameter("conditions")
	if len(clusterId) > 0 {
		app, err := models.GetApplication(clusterId)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
			return
		}
		resp.WriteEntity(app)
		return
	}

	res, err := models.ListApplication(runtimeId, conditions, paging)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(res)

}
