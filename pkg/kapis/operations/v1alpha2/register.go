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
