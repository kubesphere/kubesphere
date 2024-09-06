/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	k8serr "k8s.io/apimachinery/pkg/api/errors"

	"kubesphere.io/kubesphere/pkg/models/workloads"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

type handler struct {
	jobRunner workloads.JobRunner
}

func (h *handler) JobReRun(request *restful.Request, response *restful.Response) {
	var err error

	job := request.PathParameter("job")
	namespace := request.PathParameter("namespace")
	action := request.QueryParameter("action")
	resourceVersion := request.QueryParameter("resourceVersion")

	switch action {
	case "rerun":
		err = h.jobRunner.JobReRun(namespace, job, resourceVersion)
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
