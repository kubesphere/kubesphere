/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package openpitrix

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
)

func ListApplications(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	namespaceName := req.PathParameter("namespace")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = "create_time"
		reverse = true
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if namespaceName != "" {
		namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

		if err != nil {
			klog.Errorln(err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		var runtimeId string

		if ns, ok := namespace.(*v1.Namespace); ok {
			runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
		}

		if runtimeId == "" {
			resp.WriteAsJson(models.PageableResponse{Items: []interface{}{}, TotalCount: 0})
			return
		} else {
			conditions.Match["runtime_id"] = runtimeId
		}
	}

	result, err := openpitrix.ListApplications(conditions, limit, offset, orderBy, reverse)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func DescribeApplication(req *restful.Request, resp *restful.Response) {
	clusterId := req.PathParameter("application")
	namespaceName := req.PathParameter("namespace")
	app, err := openpitrix.DescribeApplication(namespaceName, clusterId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}
	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId != app.Cluster.RuntimeId {
		err = fmt.Errorf("rumtime not match %s,%s", app.Cluster.RuntimeId, runtimeId)
		klog.V(4).Info(err)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		return
	}

	resp.WriteEntity(app)
	return
}

func CreateApplication(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	var createClusterRequest openpitrix.CreateClusterRequest
	err := req.ReadEntity(&createClusterRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	createClusterRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	err = openpitrix.CreateApplication(namespace, createClusterRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	resp.WriteEntity(errors.None)
}

func ModifyApplication(req *restful.Request, resp *restful.Response) {
	var modifyClusterAttributesRequest openpitrix.ModifyClusterAttributesRequest
	clusterId := req.PathParameter("application")
	namespaceName := req.PathParameter("namespace")
	err := req.ReadEntity(&modifyClusterAttributesRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	app, err := openpitrix.DescribeApplication(namespaceName, clusterId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId != app.Cluster.RuntimeId {
		err = fmt.Errorf("rumtime not match %s,%s", app.Cluster.RuntimeId, runtimeId)
		klog.V(4).Info(err)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		return
	}

	err = openpitrix.PatchApplication(&modifyClusterAttributesRequest)

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DeleteApplication(req *restful.Request, resp *restful.Response) {
	clusterId := req.PathParameter("application")
	namespaceName := req.PathParameter("namespace")
	app, err := openpitrix.DescribeApplication(namespaceName, clusterId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId != app.Cluster.RuntimeId {
		err = fmt.Errorf("rumtime not match %s,%s", app.Cluster.RuntimeId, runtimeId)
		klog.V(4).Info(err)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		return
	}

	err = openpitrix.DeleteApplication(clusterId)

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}
