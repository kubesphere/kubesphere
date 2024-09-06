/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package version

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sversion "k8s.io/apimachinery/pkg/version"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/version"
)

func NewHandler(k8sVersionInfo *k8sversion.Info) rest.Handler {
	return &handler{k8sVersionInfo: k8sVersionInfo}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

type handler struct {
	k8sVersionInfo *k8sversion.Info
}

func (h *handler) AddToContainer(container *restful.Container) error {
	legacy := runtime.NewWebService(schema.GroupVersion{})
	ws := &restful.WebService{}
	ws.Path("/version").Produces(restful.MIME_JSON)
	versionFunc := func(request *restful.Request, response *restful.Response) {
		ksVersion := version.Get()
		ksVersion.Kubernetes = h.k8sVersionInfo
		response.WriteAsJson(ksVersion)
	}
	legacy.Route(legacy.GET("/version").
		To(versionFunc).
		Deprecate().
		Doc("KubeSphere version info").
		Notes("Deprecated, please use `/version` instead.").
		Operation("version-legacy").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNonResourceAPI}).
		Returns(http.StatusOK, api.StatusOK, version.Info{}))

	ws.Route(ws.GET("").
		To(versionFunc).
		Doc("KubeSphere version info").
		Operation("version").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNonResourceAPI}).
		Returns(http.StatusOK, api.StatusOK, version.Info{}))

	container.Add(legacy)
	container.Add(ws)
	return nil
}
