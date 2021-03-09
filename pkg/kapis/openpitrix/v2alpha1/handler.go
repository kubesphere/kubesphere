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

package v2alpha1

import (
	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	openpitrix "kubesphere.io/kubesphere/pkg/models/openpitrix/v2alpha1"
	openpitrixoptions "kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
)

type openpitrixHandler struct {
	openpitrix openpitrix.Interface
}

func newOpenpitrixHandler(ksInformers informers.InformerFactory, ksClient versioned.Interface, options *openpitrixoptions.Options) *openpitrixHandler {
	return &openpitrixHandler{
		openpitrix.NewOpenPitrixOperator(ksInformers),
	}
}

func (h *openpitrixHandler) DescribeRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	result, err := h.openpitrix.DescribeRepo(repoId)

	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		klog.Errorln(err)
		api.HandleInternalError(resp, req, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListRepos(req *restful.Request, resp *restful.Response) {
	q := query.ParseQueryParameter(req)
	workspace := req.PathParameter("workspace")

	result, err := h.openpitrix.ListRepos(workspace, q)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, req, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) DescribeApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	workspace := req.PathParameter("workspace")
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")

	app, err := h.openpitrix.DescribeApplication(workspace, clusterName, namespace, applicationId)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(app)
	return
}

func (h *openpitrixHandler) ListApplications(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("workspace")
	q := query.ParseQueryParameter(req)

	result, err := h.openpitrix.ListApplications(workspace, clusterName, namespace, q)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) ListApps(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	q := query.ParseQueryParameter(req)

	result, err := h.openpitrix.ListApps(workspace, q)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) ListAppVersion(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	app := req.PathParameter("app")
	q := query.ParseQueryParameter(req)

	result, err := h.openpitrix.ListAppVersions(workspace, app, q)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) ListCategories(req *restful.Request, resp *restful.Response) {
	q := query.ParseQueryParameter(req)

	result, err := h.openpitrix.ListCategories(q)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) DescribeCategory(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("category")

	result, err := h.openpitrix.DescribeCategory(id)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) DescribeApp(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter("app")

	result, err := h.openpitrix.DescribeApp(app)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) DescribeAppVersion(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("version")

	result, err := h.openpitrix.DescribeAppVersion(id)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}
