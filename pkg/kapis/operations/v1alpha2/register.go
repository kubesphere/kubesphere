/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/workloads"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "operations.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func NewHandler(cacheClient runtimeclient.Client) rest.Handler {
	return &handler{
		jobRunner: workloads.NewJobRunner(cacheClient),
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.POST("/namespaces/{namespace}/jobs/{job}").
		To(h.JobReRun).
		Deprecate().
		Doc("Job rerun").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAdvancedOperations}).
		Notes("Rerun job whether the job is complete or not.").
		Param(ws.PathParameter("job", "job name")).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.QueryParameter("action", "action must be \"rerun\"")).
		Param(ws.QueryParameter("resourceVersion", "version of job, rerun when the version matches").Required(true)).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))

	c.Add(ws)
	return nil
}
