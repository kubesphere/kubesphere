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
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

func Register(ws *restful.WebService, subPath string) {

	tags := []string{"users"}

	ws.Route(ws.POST(subPath).Doc("create user").Param(ws.PathParameter("user",
		"the username to be created").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).
		To(createUser).Consumes("*/*").Produces(restful.MIME_JSON))
	ws.Route(ws.DELETE(subPath).Doc("delete user").Param(ws.PathParameter("user",
		"the username to be deleted").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(delUser).Produces(restful.MIME_JSON))
	ws.Route(ws.GET(subPath+"/kubectl").Doc("get user's kubectl pod").Param(ws.PathParameter("user",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(getKubectl).Produces(restful.MIME_JSON))
	ws.Route(ws.GET(subPath+"/kubeconfig").Doc("get users' kubeconfig").Param(ws.PathParameter("user",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(getKubeconfig).Produces(restful.MIME_JSON))

}

func createUser(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	err := models.CreateKubeConfig(user)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = models.CreateKubectlDeploy(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(constants.MessageResponse{Message: "successfully created"})
}

func delUser(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	err := models.DelKubectlDeploy(user)

	if err != nil && !apierrors.IsNotFound(err) {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = models.DelKubeConfig(user)

	if err != nil && !apierrors.IsNotFound(err) {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = iam.DeleteRoleBindings(user)

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
