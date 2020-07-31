/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha3

import (
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/devops"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
)

type devopsHandler struct {
	devops devops.DevopsOperator
}

func newDevOpsHandler(devopsClient devopsClient.Interface, k8sclient kubernetes.Interface, ksclient kubesphere.Interface,
	ksInformers externalversions.SharedInformerFactory,
	k8sInformers informers.SharedInformerFactory) *devopsHandler {

	return &devopsHandler{
		devops: devops.NewDevopsOperator(devopsClient, k8sclient, ksclient, ksInformers, k8sInformers),
	}
}

// devopsproject handler about get/list/post/put/delete
func (h *devopsHandler) GetDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devops := request.PathParameter("devops")

	project, err := h.devops.GetDevOpsProject(workspace, devops)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(project)
}

func (h *devopsHandler) ListDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	limit, offset := params.ParsePaging(request)

	projectList, err := h.devops.ListDevOpsProject(workspace, limit, offset)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(projectList)
}

func (h *devopsHandler) CreateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.devops.CreateDevOpsProject(workspace, &devOpsProject)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else if errors.IsConflict(err) {
			api.HandleConflict(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *devopsHandler) UpdateDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var devOpsProject v1alpha3.DevOpsProject
	err := request.ReadEntity(&devOpsProject)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	project, err := h.devops.UpdateDevOpsProject(workspace, &devOpsProject)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(project)
}

func (h *devopsHandler) DeleteDevOpsProject(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	devops := request.PathParameter("devops")

	err := h.devops.DeleteDevOpsProject(workspace, devops)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

// pipeline handler about get/list/post/put/delete
func (h *devopsHandler) GetPipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	obj, err := h.devops.GetPipelineObj(devops, pipeline)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(obj)
}

func (h *devopsHandler) ListPipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	limit, offset := params.ParsePaging(request)

	objs, err := h.devops.ListPipelineObj(devops, limit, offset)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(objs)
}

func (h *devopsHandler) CreatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.devops.CreatePipelineObj(devops, &pipeline)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *devopsHandler) UpdatePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")

	var pipeline v1alpha3.Pipeline
	err := request.ReadEntity(&pipeline)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	obj, err := h.devops.UpdatePipelineObj(devops, &pipeline)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(obj)
}

func (h *devopsHandler) DeletePipeline(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	pipeline := request.PathParameter("pipeline")

	err := h.devops.DeletePipelineObj(devops, pipeline)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

//credential handler about get/list/post/put/delete
func (h *devopsHandler) GetCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	obj, err := h.devops.GetCredentialObj(devops, credential)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(obj)
}

func (h *devopsHandler) ListCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	query := query.ParseQueryParameter(request)
	objs, err := h.devops.ListCredentialObj(devops, query)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(objs)
}

func (h *devopsHandler) CreateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.devops.CreateCredentialObj(devops, &obj)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *devopsHandler) UpdateCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	var obj v1.Secret
	err := request.ReadEntity(&obj)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.devops.UpdateCredentialObj(devops, &obj)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *devopsHandler) DeleteCredential(request *restful.Request, response *restful.Response) {
	devops := request.PathParameter("devops")
	credential := request.PathParameter("credential")

	err := h.devops.DeleteCredentialObj(devops, credential)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}
