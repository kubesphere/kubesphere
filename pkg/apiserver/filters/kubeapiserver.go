/*
Copyright 2020 The KubeSphere Authors.

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

package filters

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type kubeAPIProxy struct {
	next          http.Handler
	kubeAPIServer *url.URL
	transport     http.RoundTripper
}

// WithKubeAPIServer proxy request to kubernetes service if requests path starts with /api
func WithKubeAPIServer(next http.Handler, config *rest.Config) http.Handler {
	kubeAPIServer, _ := url.Parse(config.Host)
	transport, err := rest.TransportFor(config)
	if err != nil {
		klog.Errorf("Unable to create transport from rest.Config: %v", err)
		return next
	}
	return &kubeAPIProxy{
		next:          next,
		kubeAPIServer: kubeAPIServer,
		transport:     transport,
	}
}

func (k kubeAPIProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info, ok := request.RequestInfoFrom(req.Context())
	if !ok {
		responsewriters.InternalError(w, req, fmt.Errorf("no RequestInfo found in the context"))
		return
	}

	if info.IsKubernetesRequest {
		location := &url.URL{}
		location.Scheme = k.kubeAPIServer.Scheme
		location.Host = k.kubeAPIServer.Host
		location.Path = req.URL.Path
		location.RawQuery = req.URL.Query().Encode()

		newReq := req.WithContext(req.Context())
		newReq.Header = utilnet.CloneHeader(req.Header)
		newReq.URL = location
		newReq.Host = location.Host

		// make sure we don't override kubernetes's authorization
		newReq.Header.Del("Authorization")
		upgrade := httpstream.IsUpgradeRequest(req)
		httpProxy := proxy.NewUpgradeAwareHandler(location, k.transport, false, upgrade, &responder{})
		httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(k.transport, k.transport)
		httpProxy.ServeHTTP(w, newReq)
		return
	}

	k.next.ServeHTTP(w, req)
}
