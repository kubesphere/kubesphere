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

package deployments

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

	tags := []string{"deployments"}

	ws.Route(ws.GET(subPath).To(getDeployRevision).Consumes("*/*").Metadata(restfulspec.KeyOpenAPITags, tags).Doc("Handle deployment" +
		" operation").Param(ws.PathParameter("deployment", "deployment's name").DataType("string")).Param(ws.PathParameter("namespace",
		"deployment's namespace").DataType("string")).Param(ws.PathParameter("deployment", "deployment's name")).Writes(v1.ReplicaSet{}))
}

func getDeployRevision(req *restful.Request, resp *restful.Response) {
	deploy := req.PathParameter("deployment")
	namespace := req.PathParameter("namespace")
	revision := req.PathParameter("revision")

	res, err := models.GetDeployRevision(namespace, deploy, revision)

	if err != nil {
		if errors.IsNotFound(err) {
			resp.WriteHeaderAndEntity(http.StatusNotFound, constants.MessageResponse{Message: err.Error()})
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		}
	}

	resp.WriteEntity(res)
}
