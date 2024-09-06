/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gatewayv1alpha2 "kubesphere.io/api/gateway/v1alpha2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "gateway.kubesphere.io"
	Version   = "v1alpha2"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func NewHandler(cache runtimeclient.Reader) rest.Handler {
	return &handler{
		cache: cache,
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/namespaces/{namespace}/availableingressclassscopes").
		To(h.ListIngressClassScopes).
		Doc("List ingressClassScope available for the namespace").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Returns(http.StatusOK, api.StatusOK, []gatewayv1alpha2.IngressClassScope{}))

	container.Add(ws)
	return nil
}
