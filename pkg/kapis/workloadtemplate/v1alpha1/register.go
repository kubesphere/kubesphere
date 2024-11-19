package v1alpha1

import (
	"fmt"

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

type funcInfo struct {
	Route     string
	Func      func(req *restful.Request, resp *restful.Response)
	Method    func(subPath string) *restful.RouteBuilder
	Doc       string
	Workspace bool
	Params    []*restful.Parameter
}

func (h *templateHandler) AddToContainer(c *restful.Container) (err error) {
	ws := runtime.NewWebService(SchemeGroupVersion)

	funcInfoList := []funcInfo{
		{Route: "/workloadtemplates", Func: h.list, Method: ws.GET, Workspace: true},
		{Route: "/workloadtemplates", Func: h.apply, Method: ws.POST, Workspace: true},
		{Route: "/workloadtemplates/{workloadtemplate}", Func: h.delete, Method: ws.DELETE, Workspace: true},
		{Route: "/workloadtemplates/{workloadtemplate}", Func: h.get, Method: ws.GET, Workspace: true},
	}
	for _, info := range funcInfoList {
		builder := info.Method(info.Route).To(info.Func).Doc(info.Doc)
		for _, param := range info.Params {
			builder = builder.Param(param)
		}
		ws.Route(builder)
		if info.Workspace {
			workspaceRoute := fmt.Sprintf("/workspaces/{workspace}%s", info.Route)
			builder = info.Method(workspaceRoute).To(info.Func).Doc(info.Doc)
			for _, param := range info.Params {
				builder = builder.Param(param)
			}
			builder.Param(ws.PathParameter("workspace", "workspace"))
			ws.Route(builder)
		}
	}
	c.Add(ws)
	return nil
}
