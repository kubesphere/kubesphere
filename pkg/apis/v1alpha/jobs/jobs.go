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

package jobs

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful-openapi"

	"fmt"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/controllers"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"jobs"}

	ws.Route(ws.POST(subPath).To(handleJob).Consumes("*/*").Metadata(restfulspec.KeyOpenAPITags, tags).Doc("Handle job" +
		" operation").Param(ws.PathParameter("job", "job name").DataType("string")).Param(ws.PathParameter("namespace",
		"job's namespace").DataType("string")).Param(ws.QueryParameter("a",
		"action").DataType("string")).Writes(""))
}

func handleJob(req *restful.Request, resp *restful.Response) {
	var res interface{}
	var err error

	job := req.PathParameter("job")
	namespace := req.PathParameter("namespace")
	action := req.QueryParameter("a")

	switch action {
	case "rerun":
		res, err = controllers.JobReRun(namespace, job)
	default:
		resp.WriteHeaderAndEntity(http.StatusForbidden, constants.MessageResponse{Message: fmt.Sprintf("invalid operation %s", action)})
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(res)
}
