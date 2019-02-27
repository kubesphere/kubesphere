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
package revisions

import (
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/api/apps/v1"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/revisions"
)

func V1Alpha2(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/daemonsets/{daemonset}/revisions/{revision}").
		To(getDaemonSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{"daemonsets", "revision"}).
		Doc("Handle daemonset operation").
		Param(ws.PathParameter("daemonset", "daemonset's name").
			DataType("string")).
		Param(ws.PathParameter("namespace", "daemonset's namespace").
			DataType("string")).
		Param(ws.PathParameter("revision", "daemonset's revision")).
		Writes(v1.DaemonSet{}))
	ws.Route(ws.GET("/namespaces/{namespace}/deployments/{deployment}/revisions/{revision}").
		To(getDeployRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{"deployments", "revision"}).
		Doc("Handle deployment operation").
		Param(ws.PathParameter("deployment", "deployment's name").
			DataType("string")).
		Param(ws.PathParameter("namespace",
			"deployment's namespace").
			DataType("string")).
		Param(ws.PathParameter("deployment", "deployment's name")).
		Writes(v1.ReplicaSet{}))
	ws.Route(ws.GET("/namespaces/{namespace}/statefulsets/{statefulset}/revisions/{revision}").
		To(getStatefulSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{"statefulsets", "revisions"}).
		Doc("Handle statefulset operation").
		Param(ws.PathParameter("statefulset", "statefulset's name").
			DataType("string")).
		Param(ws.PathParameter("namespace", "statefulset's namespace").
			DataType("string")).
		Param(ws.PathParameter("revision", "statefulset's revision")).
		Writes(v1.StatefulSet{}))
}

func getDaemonSetRevision(req *restful.Request, resp *restful.Response) {
	daemonset := req.PathParameter("daemonset")
	namespace := req.PathParameter("namespace")
	revision, err := strconv.Atoi(req.PathParameter("revision"))

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	result, err := revisions.GetDaemonSetRevision(namespace, daemonset, revision)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(result)
}

func getDeployRevision(req *restful.Request, resp *restful.Response) {
	deploy := req.PathParameter("deployment")
	namespace := req.PathParameter("namespace")
	revision := req.PathParameter("revision")

	result, err := revisions.GetDeployRevision(namespace, deploy, revision)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(result)
}

func getStatefulSetRevision(req *restful.Request, resp *restful.Response) {
	statefulset := req.PathParameter("statefulset")
	namespace := req.PathParameter("namespace")
	revision, err := strconv.Atoi(req.PathParameter("revision"))

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	result, err := revisions.GetStatefulSetRevision(namespace, statefulset, revision)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(result)
}
