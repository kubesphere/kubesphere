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

	"fmt"
	"strings"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"resources"}

	ws.Route(ws.GET(subPath+"/{resource}").To(listResource).
		Produces(restful.MIME_JSON).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get resource list").
		Param(ws.PathParameter("resource", "resource name").DataType("string")).
		Param(ws.QueryParameter("conditions", "search conditions").DataType("string")).
		Param(ws.QueryParameter("reverse", "support reverse ordering").DataType("bool").DefaultValue("false")).
		Param(ws.QueryParameter("order", "the field for sorting").DataType("string")).
		Param(ws.QueryParameter("paging", "support paging function").DataType("string")).
		Writes(models.ResourceList{}))
}

func isInvalid(str string) bool {
	invalidList := []string{"exec", "insert", "select", "delete", "update", "count", "*", "%", "truncate", "drop"}
	str = strings.Replace(str, "=", " ", -1)
	str = strings.Replace(str, ",", " ", -1)
	str = strings.Replace(str, "~", " ", -1)
	items := strings.Split(str, " ")

	for _, invalid := range invalidList {
		for _, item := range items {
			if item == invalid || strings.ToLower(item) == invalid {
				return true
			}
		}
	}

	return false
}

func listResource(req *restful.Request, resp *restful.Response) {

	resource := req.PathParameter("resource")
	if resource == "applications" {
		handleApplication(req, resp)
		return
	}
	conditions := req.QueryParameter("conditions")
	paging := req.QueryParameter("paging")
	orderField := req.QueryParameter("order")
	reverse := req.QueryParameter("reverse")

	if len(orderField) > 0 {
		if reverse == "true" {
			orderField = fmt.Sprintf("%s %s", orderField, "desc")
		} else {
			orderField = fmt.Sprintf("%s %s", orderField, "asc")
		}
	}

	if isInvalid(conditions) || isInvalid(paging) || isInvalid(orderField) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: "invalid input"})
		return
	}

	res, err := models.ListResource(resource, conditions, paging, orderField)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(res)
}

func handleApplication(req *restful.Request, resp *restful.Response) {
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
