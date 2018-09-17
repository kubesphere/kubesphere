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

package statefulsets

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"statefulsets"}

	ws.Route(ws.GET(subPath).To(getDaemonSetRevision).Consumes("*/*").Metadata(restfulspec.KeyOpenAPITags, tags).Doc("Handle statefulset" +
		" operation").Param(ws.PathParameter("statefulset", "statefulset's name").DataType("string")).Param(ws.PathParameter("namespace",
		"statefulset's namespace").DataType("string")).Param(ws.PathParameter("revision", "statefulset's revision")).Writes(v1.StatefulSet{}))
}

func getDaemonSetRevision(req *restful.Request, resp *restful.Response) {
	statefulset := req.PathParameter("statefulset")
	namespace := req.PathParameter("namespace")
	revision := req.PathParameter("revision")

	res, err := models.GetStatefulSetRevision(namespace, statefulset, revision)

	if err != nil {
		if errors.IsNotFound(err) {
			resp.WriteHeaderAndEntity(http.StatusNotFound, constants.MessageResponse{Message: err.Error()})
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		}
	}

	resp.WriteEntity(res)
}
