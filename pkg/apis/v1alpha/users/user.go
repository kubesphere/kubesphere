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

package user

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.POST(subPath).To(createUser).Consumes("*/*").Produces(restful.MIME_JSON))
	ws.Route(ws.DELETE(subPath).To(delUser).Produces(restful.MIME_JSON))
	ws.Route(ws.GET(subPath + "/kubectl").To(getKubectl).Produces(restful.MIME_JSON))
	ws.Route(ws.GET(subPath + "/kubeconfig").To(getKubeconfig).Produces(restful.MIME_JSON))

}

func createUser(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	err := models.CreateKubeConfig(user)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = models.CreateKubectlPod(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(constants.MessageResponse{Message: "successfully created"})
}

func delUser(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	err := models.DelKubectlPod(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = models.DelKubeConfig(user)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = models.DeleteRoleBindings(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(constants.MessageResponse{Message: "successfully deleted"})
}

func getKubectl(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	kubectlPod, err := models.GetKubectlPod(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(kubectlPod)
}

func getKubeconfig(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	kubectlConfig, err := models.GetKubeConfig(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(kubectlConfig)
}
