/*
Copyright 2020 KubeSphere Authors

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

package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/models/workloads"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

type operationHandler struct {
	jobRunner workloads.JobRunner
}

func newOperationHandler(client kubernetes.Interface) *operationHandler {
	return &operationHandler{
		jobRunner: workloads.NewJobRunner(client),
	}
}

func (r *operationHandler) handleJobReRun(request *restful.Request, response *restful.Response) {
	var err error

	job := request.PathParameter("job")
	namespace := request.PathParameter("namespace")
	action := request.QueryParameter("action")
	resourceVersion := request.QueryParameter("resourceVersion")

	switch action {
	case "rerun":
		err = r.jobRunner.JobReRun(namespace, job, resourceVersion)
	default:
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid operation %s", action)))
		return
	}
	if err != nil {
		if k8serr.IsConflict(err) {
			response.WriteHeaderAndEntity(http.StatusConflict, errors.Wrap(err))
			return
		}
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(errors.None)
}
