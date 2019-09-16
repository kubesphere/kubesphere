/*

 Copyright 2019 The KubeSphere Authors.

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
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/applications"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"net/http"
)

func ListApplication(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	clusterId := req.QueryParameter("cluster_id")
	runtimeId := req.QueryParameter("runtime_id")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	if err != nil {
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
			return
		}
	}
	if len(clusterId) > 0 {
		app, err := applications.GetApp(clusterId)
		if err != nil {
			klog.Errorln("get application error", err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		resp.WriteEntity(app)
		return
	}

	result, err := applications.ListApplication(runtimeId, conditions, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)

}

func ListNamespacedApplication(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	namespaceName := req.PathParameter("namespace")
	clusterId := req.QueryParameter("cluster_id")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	if len(clusterId) > 0 {
		app, err := applications.GetApp(clusterId)
		if err != nil {
			klog.Errorln("get app failed", err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		resp.WriteEntity(app)
		return
	}

	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln("get namespace failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId == "" {
		klog.Errorln("runtime id not found")
		resp.WriteAsJson(models.PageableResponse{Items: []interface{}{}, TotalCount: 0})
		return
	}

	result, err := applications.ListApplication(runtimeId, conditions, limit, offset)

	if err != nil {
		klog.Errorln("list applications failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func DescribeApplication(req *restful.Request, resp *restful.Response) {
	clusterId := req.PathParameter("application")
	namespaceName := req.PathParameter("namespace")
	app, err := applications.GetApp(clusterId)
	if err != nil {
		klog.Errorln("get app failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln("get namespace failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId != app.RuntimeId {
		klog.Errorln("runtime not match", app.RuntimeId, runtimeId)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.New(fmt.Sprintf("rumtime not match %s,%s", app.RuntimeId, runtimeId)))
		return
	}

	resp.WriteEntity(app)
	return
}

func DeployApplication(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	var app openpitrix.CreateClusterRequest
	err := req.ReadEntity(&app)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	err = applications.DeployApplication(namespace, app)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	resp.WriteEntity(errors.None)
}

func DeleteApplication(req *restful.Request, resp *restful.Response) {
	clusterId := req.PathParameter("application")
	namespaceName := req.PathParameter("namespace")
	app, err := applications.GetApp(clusterId)
	if err != nil {
		klog.Errorln("get app failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln("get namespace failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId != app.RuntimeId {
		klog.Errorln("runtime not match", app.RuntimeId, runtimeId)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.New(fmt.Sprintf("rumtime not match %s,%s", app.RuntimeId, runtimeId)))
		return
	}

	err = applications.DeleteApplication(clusterId)

	if err != nil {
		klog.Errorln("delete application failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}
