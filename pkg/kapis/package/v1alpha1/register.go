/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "package.kubesphere.io"
	Version   = "v1alpha1"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func NewHandler(cache runtimeclient.Reader) rest.Handler {
	return &handler{cache: cache}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)
	ws.Route(ws.GET("/extensionversions/{version}/files").
		To(h.ListFiles).
		Doc("List all files").
		Operation("list-extension-version-files").
		Param(ws.PathParameter("version", "The specified extension version name.")).
		Returns(http.StatusOK, api.StatusOK, []loader.BufferedFile{}))
	container.Add(ws)
	return nil
}
