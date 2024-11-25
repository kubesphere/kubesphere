package v1alpha1

import (
	"net/http"

	"kubesphere.io/kubesphere/pkg/api"

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
	client runtimeclient.Client
}

func NewHandler(cacheClient runtimeclient.Client) rest.Handler {
	handler := &templateHandler{
		client: cacheClient,
	}
	return handler
}

func (h *templateHandler) AddToContainer(c *restful.Container) (err error) {
	ws := runtime.NewWebService(SchemeGroupVersion)

	ws.Route(ws.GET("/workloadtemplates").
		To(h.list).
		Doc("List workload templates"))

	ws.Route(ws.POST("/workloadtemplates").
		To(h.apply).
		Doc("Apply a workload template").
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.DELETE("/workloadtemplates/{workloadtemplate}").
		To(h.delete).
		Doc("Delete a workload template").
		Param(ws.PathParameter("workloadtemplate", "The specified workload template").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/workloadtemplates/{workloadtemplate}").
		To(h.get).
		Doc("Get a specific workload template").
		Param(ws.PathParameter("workloadtemplate", "The specified workload template").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/workspaces/{workspace}/workloadtemplates").
		To(h.list).
		Doc("List workload templates in a workspace").
		Param(ws.PathParameter("workspace", "The specified workspace").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.POST("/workspaces/{workspace}/workloadtemplates").
		To(h.apply).
		Doc("Apply a workload template in a workspace").
		Param(ws.PathParameter("workspace", "The specified workspace").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.DELETE("/workspaces/{workspace}/workloadtemplates/{workloadtemplate}").
		To(h.delete).
		Doc("Delete a workload template in a workspace").
		Param(ws.PathParameter("workspace", "The specified workspace").Required(true)).
		Param(ws.PathParameter("workloadtemplate", "The specified workload template").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/workspaces/{workspace}/workloadtemplates/{workloadtemplate}").
		To(h.get).
		Doc("Get a specific workload template in a workspace").
		Param(ws.PathParameter("workspace", "The specified workspace").Required(true)).
		Param(ws.PathParameter("workloadtemplate", "The specified workload template").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	c.Add(ws)
	return nil
}
