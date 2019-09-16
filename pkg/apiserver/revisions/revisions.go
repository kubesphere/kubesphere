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
	"kubesphere.io/kubesphere/pkg/models/revisions"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func GetDaemonSetRevision(req *restful.Request, resp *restful.Response) {
	daemonset := req.PathParameter("daemonset")
	namespace := req.PathParameter("namespace")
	revision, err := strconv.Atoi(req.PathParameter("revision"))

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := revisions.GetDaemonSetRevision(namespace, daemonset, revision)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func GetDeployRevision(req *restful.Request, resp *restful.Response) {
	deploy := req.PathParameter("deployment")
	namespace := req.PathParameter("namespace")
	revision := req.PathParameter("revision")

	result, err := revisions.GetDeployRevision(namespace, deploy, revision)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func GetStatefulSetRevision(req *restful.Request, resp *restful.Response) {
	statefulset := req.PathParameter("statefulset")
	namespace := req.PathParameter("namespace")
	revision, err := strconv.Atoi(req.PathParameter("revision"))

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := revisions.GetStatefulSetRevision(namespace, statefulset, revision)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}
