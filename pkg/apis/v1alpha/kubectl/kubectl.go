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

package kubectl

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath).Consumes("*/*").Produces(restful.MIME_JSON).To(handleKubectl).Doc("use to " +
		"get a kubectl pod in specified namespaces").Param(ws.PathParameter("namespace",
		"namespace").DataType("string")).Do(returns200,returns500))

}

func handleKubectl(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("namespace")

	kubectlPod, err := models.GetKubectlPod(ns)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}

	resp.WriteEntity(kubectlPod)
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", nil)
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "fail", nil)
}
