/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package generic

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

// genericProxy is a simple proxy for external service.
type genericProxy struct {
	// proxy service endpoint
	Endpoint *url.URL

	// api group name exposed to clients
	GroupName string

	// api version
	Version string

	// mark as desprecated
	desprecated bool
}

func NewGenericProxy(endpoint string, groupName string, version string) (*genericProxy, error) {
	parse, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// trim path suffix slash
	parse.Path = strings.Trim(parse.Path, "/")

	return &genericProxy{
		Endpoint:  parse,
		GroupName: groupName,
		Version:   version,
	}, nil
}

func (g *genericProxy) SetProxyDesprecated() {
	g.desprecated = true
}

// currently, we only support proxy GET/PUT/POST/DELETE/PATCH.
// Maybe we can try another way to implement proxy.
func (g *genericProxy) AddToContainer(container *restful.Container) error {
	webservice := runtime.NewWebService(schema.GroupVersion{
		Group:   g.GroupName,
		Version: g.Version,
	})

	webservice.Route(webservice.GET("/{path:*}").
		To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.PUT("/{path:*}").
		To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.POST("/{path:*}").
		To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.DELETE("/{path:*}").
		To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.PATCH("/{path:*}").
		To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	container.Add(webservice)
	return nil

}

func (g *genericProxy) handler(request *restful.Request, response *restful.Response) {
	if g.desprecated {
		klog.Warning(fmt.Sprintf("This proxy group %s has deprecated", g.GroupName))
	}

	u := g.makeURL(request)

	httpProxy := proxy.NewUpgradeAwareHandler(u, http.DefaultTransport, false, false, &errorResponder{})
	httpProxy.ServeHTTP(response, request.Request)
}

func (g *genericProxy) makeURL(request *restful.Request) *url.URL {
	u := *(request.Request.URL)
	u.Host = g.Endpoint.Host
	u.Scheme = g.Endpoint.Scheme
	u.Path = strings.Replace(request.Request.URL.Path, fmt.Sprintf("/kapis/%s", g.GroupName), "", 1)

	// prepend path from endpoint
	if len(g.Endpoint.Path) != 0 {
		u.Path = fmt.Sprintf("/%s%s", g.Endpoint.Path, u.Path)
	}

	return &u
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
}
