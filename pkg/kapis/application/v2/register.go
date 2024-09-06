/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"

	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/simple/client/application"

	"kubesphere.io/utils/s3"

	"github.com/emicklei/go-restful/v3"
	appv2 "kubesphere.io/api/application/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

type appHandler struct {
	client        runtimeclient.Client
	clusterClient clusterclient.Interface
	s3opts        *s3.Options
	ossStore      s3.Interface
	cmStore       s3.Interface
}

func NewHandler(cacheClient runtimeclient.Client, clusterClient clusterclient.Interface, s3opts *s3.Options) rest.Handler {
	handler := &appHandler{
		client:        cacheClient,
		clusterClient: clusterClient,
		s3opts:        s3opts,
	}
	return handler
}

func NewFakeHandler() rest.Handler {
	return &appHandler{}
}

type funcInfo struct {
	Route      string
	Func       func(req *restful.Request, resp *restful.Response)
	Method     func(subPath string) *restful.RouteBuilder
	Doc        string
	NeedGlobal bool
	Workspace  bool
	Namespace  bool
	Params     []*restful.Parameter
}

func (h *appHandler) AddToContainer(c *restful.Container) (err error) {
	ws := runtime.NewWebService(appv2.SchemeGroupVersion)
	if h.cmStore, h.ossStore, err = application.InitStore(h.s3opts, h.client); err != nil {
		klog.Errorf("failed to init store: %v", err)
		return err
	}
	funcInfoList := []funcInfo{
		{Route: "/repos", Func: h.ListRepos, Method: ws.GET, Workspace: true},
		{Route: "/repos", Func: h.CreateOrUpdateRepo, Method: ws.POST, Workspace: true},
		{Route: "/repos/{repo}", Func: h.CreateOrUpdateRepo, Method: ws.PATCH, Workspace: true},
		{Route: "/repos/{repo}", Func: h.DeleteRepo, Method: ws.DELETE, Workspace: true},
		{Route: "/repos/{repo}", Func: h.DescribeRepo, Method: ws.GET, Workspace: true},
		{Route: "/repos/{repo}/events", Func: h.ListRepoEvents, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/action", Func: h.DoAppAction, Method: ws.POST},
		{Route: "/apps", Func: h.CreateOrUpdateApp, Method: ws.POST, Workspace: true},
		{Route: "/apps", Func: h.ListApps, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}", Func: h.CreateOrUpdateApp, Method: ws.POST, Workspace: true},
		{Route: "/apps/{app}", Func: h.PatchApp, Method: ws.PATCH, Workspace: true},
		{Route: "/apps/{app}", Func: h.DescribeApp, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}", Func: h.DeleteApp, Method: ws.DELETE, Workspace: true},
		{Route: "/apps/{app}/versions", Func: h.CreateOrUpdateAppVersion, Method: ws.POST, Workspace: true},
		{Route: "/apps/{app}/versions", Func: h.ListAppVersions, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/versions/{version}", Func: h.DeleteAppVersion, Method: ws.DELETE, Workspace: true},
		{Route: "/apps/{app}/versions/{version}", Func: h.DescribeAppVersion, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/versions/{version}", Func: h.CreateOrUpdateAppVersion, Method: ws.POST, Workspace: true},
		{Route: "/apps/{app}/versions/{version}/package", Func: h.GetAppVersionPackage, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/versions/{version}/files", Func: h.GetAppVersionFiles, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/versions/{version}/action", Func: h.AppVersionAction, Method: ws.POST, Workspace: true},
		{Route: "/applications", Func: h.ListAppRls, Method: ws.GET, Workspace: true, Namespace: true},
		{Route: "/applications", Func: h.CreateOrUpdateAppRls, Method: ws.POST, Workspace: true, Namespace: true},
		{Route: "/applications/{application}", Func: h.CreateOrUpdateAppRls, Method: ws.POST, Namespace: true},
		{Route: "/applications/{application}", Func: h.DescribeAppRls, Method: ws.GET, Namespace: true},
		{Route: "/applications/{application}", Func: h.DeleteAppRls, Method: ws.DELETE, Namespace: true},
		{Route: "/categories", Func: h.CreateOrUpdateCategory, Method: ws.POST},
		{Route: "/categories", Func: h.ListCategories, Method: ws.GET},
		{Route: "/categories/{category}", Func: h.DeleteCategory, Method: ws.DELETE},
		{Route: "/categories/{category}", Func: h.CreateOrUpdateCategory, Method: ws.POST},
		{Route: "/categories/{category}", Func: h.DescribeCategory, Method: ws.GET},
		{Route: "/reviews", Func: h.ListReviews, Method: ws.GET},
		{Route: "/attachments", Func: h.CreateAttachment, Method: ws.POST, Workspace: true},
		{Route: "/attachments/{attachment}", Func: h.DescribeAttachment, Method: ws.GET, Workspace: true},
		{Route: "/attachments/{attachment}", Func: h.DeleteAttachments, Method: ws.DELETE, Workspace: true},
		{Route: "/apps/{app}/examplecr/{name}", Func: h.exampleCr, Method: ws.GET, Workspace: true},
		{Route: "/apps/{app}/cr", Func: h.AppCrList, Method: ws.GET, Workspace: true, Namespace: true},
		{Route: "/cr", Func: h.CreateOrUpdateCR, Method: ws.POST, Workspace: true, Namespace: true},
		{Route: "/cr/{crname}", Func: h.DescribeAppCr, Method: ws.GET, Workspace: true, Namespace: true},
		{Route: "/cr/{crname}", Func: h.DeleteAppCr, Method: ws.DELETE, Workspace: true, Namespace: true},
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

		if info.Namespace {
			namespaceRoute := fmt.Sprintf("/namespaces/{namespace}%s", info.Route)
			builder = info.Method(namespaceRoute).To(info.Func).Doc(info.Doc)
			for _, param := range info.Params {
				builder = builder.Param(param)
			}
			builder.Param(ws.PathParameter("namespace", "namespace"))
			ws.Route(builder)
		}
	}

	c.Add(ws)
	return nil
}
