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
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/applications"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/params"
	"net/http"
)

func ApplicationHandler(req *restful.Request, resp *restful.Response) {
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
			glog.Errorln("get application error", err)
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

func NamespacedApplicationHandler(req *restful.Request, resp *restful.Response) {
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
			glog.Errorln("get app failed", err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		resp.WriteEntity(app)
		return
	}

	namespace, err := resources.GetResource("", resources.Namespaces, namespaceName)

	if err != nil {
		glog.Errorln("get namespace failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var runtimeId string

	if ns, ok := namespace.(*v1.Namespace); ok {
		runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
	}

	if runtimeId == "" {
		glog.Errorln("runtime id not found")
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.New("openpitrix runtime not init"))
		return
	}

	result, err := applications.ListApplication(runtimeId, conditions, limit, offset)

	if err != nil {
		glog.Errorln("list applications failed", err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}
