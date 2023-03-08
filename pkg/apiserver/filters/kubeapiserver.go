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
		s := *req.URL
		s.Host = k.kubeAPIServer.Host
		s.Scheme = k.kubeAPIServer.Scheme

		// make sure we don't override kubernetes's authorization
		req.Header.Del("Authorization")
		httpProxy := proxy.NewUpgradeAwareHandler(&s, k.transport, true, false, &responder{})
		httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(k.transport, k.transport)
		httpProxy.ServeHTTP(w, req)
		return
	}

	k.next.ServeHTTP(w, req)
}
