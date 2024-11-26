/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"github.com/Masterminds/semver/v3"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "workloadtemplate.kubesphere.io", Version: "v1alpha1"}
)

type templateHandler struct {
	client         runtimeclient.Client
	authorizer     authorizer.Authorizer
	resourceGetter *resourcesv1alpha3.Getter
}

func NewHandler(cacheClient runtimeclient.Client, k8sVersion *semver.Version, authorizer authorizer.Authorizer) rest.Handler {
	handler := &templateHandler{
		client:         cacheClient,
		authorizer:     authorizer,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(cacheClient, k8sVersion),
	}
	return handler
}
func (h *templateHandler) AddToContainer(c *restful.Container) (err error) {
	ws := runtime.NewWebService(SchemeGroupVersion)

	ws.Route(ws.GET("/workloadtemplates").
		To(h.listWorkloadTemplate).
		Doc("List workload templates").
		Notes("List workload templates.").
		Operation("listWorkloadTemplate"))
	ws.Route(ws.GET("/workspaces/{workspace}/workloadtemplates").
		To(h.listWorkloadTemplate).
		Doc("List workload templates in a workspace").
		Notes("List workload templates in a workspace.").
		Operation("listWorkloadTemplate").
		Param(ws.PathParameter("workspace", "workspace")))
	ws.Route(ws.GET("/namespaces/{namespace}/workloadtemplates").
		To(h.listWorkloadTemplate).
		Doc("List workload templates in a namespace").
		Notes("List workload templates in a namespace.").
		Operation("listWorkloadTemplate").
		Param(ws.PathParameter("namespace", "namespace")))

	ws.Route(ws.POST("/namespaces/{namespace}/workloadtemplates").
		To(h.applyWorkloadTemplate).
		Doc("Apply a workload template in a namespace").
		Notes("Apply a workload template in a namespace.").
		Operation("applyWorkloadTemplate").
		Param(ws.PathParameter("namespace", "namespace")))

	ws.Route(ws.PUT("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.applyWorkloadTemplate).
		Doc("Update a workload template").
		Notes("Update a workload template in a namespace.").
		Operation("applyWorkloadTemplate").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	ws.Route(ws.DELETE("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.deleteWorkloadTemplate).
		Doc("Delete a workload template in a namespace").
		Notes("List workload templates in a namespace.").
		Operation("deleteWorkloadTemplate").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	ws.Route(ws.GET("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.getWorkloadTemplate).
		Doc("Get a workload template in a namespace").
		Notes("Get a workload template in a namespace.").
		Operation("getWorkloadTemplate").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	c.Add(ws)
	return nil
}
