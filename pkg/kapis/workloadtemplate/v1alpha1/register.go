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
		To(h.list).
		Doc("List workload templates"))
	ws.Route(ws.GET("/workspaces/{workspace}/workloadtemplates").
		To(h.list).
		Doc("List workload templates in a workspace").
		Param(ws.PathParameter("workspace", "workspace")))
	ws.Route(ws.GET("/namespaces/{namespace}/workloadtemplates").
		To(h.list).
		Doc("List workload templates in a namespace").
		Param(ws.PathParameter("namespace", "namespace")))

	ws.Route(ws.POST("/namespaces/{namespace}/workloadtemplates").
		To(h.apply).
		Doc("Apply a workload template").
		Param(ws.PathParameter("namespace", "namespace")))

	ws.Route(ws.PUT("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.apply).
		Doc("Update a workload template").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	ws.Route(ws.DELETE("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.delete).
		Doc("Delete a workload template").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	ws.Route(ws.GET("/namespaces/{namespace}/workloadtemplates/{workloadtemplate}").
		To(h.get).
		Doc("Get a workload template").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("workloadtemplate", "workloadtemplate")))

	c.Add(ws)
	return nil
}
