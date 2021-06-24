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
	"net/http/httputil"
	"net/url"

	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"

	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

// WithDevOpsAPIServer proxy request to DevOps service if requests path has the APIGroup with devops.kubesphere.io
func WithDevOpsAPIServer(handler http.Handler, config *jenkins.Options, failed proxy.ErrorResponder) http.Handler {
	if config.DevOpsServiceAddress == "" || config.K8sBearerToken == "" {
		klog.V(6).Info("The DevOps service address or k8s token is empty, the proxy of DevOps server was not enabled")
		// this filter rely on a separate DevOps address
		// do not pass the proxy if there's no service address
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(w, req)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			err := errors.New("Unable to retrieve request info from request")
			klog.Error(err)
			responsewriters.InternalError(w, req, err)
		}

		if info.APIGroup == "devops.kubesphere.io" {
			serverAsProxy(w, req, config.DevOpsServiceAddress, config.K8sBearerToken)
		} else if info.APIGroup == "tenant.kubesphere.io" && info.Resource == "devops" ||
			info.APIGroup == "resources.kubesphere.io" && info.Resource == "s2ibuilders" ||
			info.APIGroup == "resources.kubesphere.io" && info.Resource == "s2iruns" {
			serverAsProxy(w, req, config.DevOpsPluginServiceAddress, config.K8sBearerToken)
		} else {
			handler.ServeHTTP(w, req)
		}
	})
}

func serverAsProxy(w http.ResponseWriter, req *http.Request, server, token string) {
	devopsURL := url.URL{
		Scheme: req.URL.Scheme,
		Host:   server,
	}
	if devopsURL.Scheme == "" {
		devopsURL.Scheme = "http"
	}

	req.Header.Set("X-Authorization", fmt.Sprintf("bearer %s", token))
	devopsProxy := httputil.NewSingleHostReverseProxy(&devopsURL)
	devopsProxy.ServeHTTP(w, req)
}
