/*
Copyright 2020 KubeSphere Authors

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

package generic

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"net/http"
	"net/url"
	"strings"
)

// genericProxy is a simple proxy for external service.
type genericProxy struct {
	// proxy service endpoint
	Endpoint *url.URL

	// api group name exposed to clients
	GroupName string

	// api version
	Version string
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
