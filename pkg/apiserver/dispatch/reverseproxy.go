/*

 Copyright 2021 The KubeSphere Authors.

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

package dispatch

import (
	"net/http"
	"net/url"
	"regexp"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	extensionsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/extensions/v1alpha1"
)

type reverseProxyDispatcher struct {
	reverseProxyInformer extensionsinformers.ReverseProxyInformer
}

func (s *reverseProxyDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	info, _ := request.RequestInfoFrom(req.Context())
	if info.IsKubernetesRequest {
		return false
	}
	reverseProxies, err := s.reverseProxyInformer.Lister().List(labels.Everything())
	if err != nil {
		responsewriters.InternalError(w, req, err)
		return true
	}

	for _, reverseProxy := range reverseProxies {
		if reverseProxy.Status.State != extensionsv1alpha1.StateEnabled {
			continue
		}

		if reverseProxy.Spec.Matcher.Method != req.Method && reverseProxy.Spec.Matcher.Method != "*" {
			continue
		}

		matched, err := regexp.MatchString(reverseProxy.Spec.Matcher.Path, info.Path)
		if err != nil {
			klog.Warningf("reverse proxy dispatcher patch match failed: %v", err)
			continue
		}

		if matched {
			endpoint, err := url.Parse(reverseProxy.Spec.Upstream.RawURL())
			if err != nil {
				responsewriters.InternalError(w, req, err)
				return true
			}
			location := req.URL
			location.Host = endpoint.Host
			location.Scheme = endpoint.Scheme
			location.Path = endpoint.Path + location.Path
			httpProxy := proxy.NewUpgradeAwareHandler(location, http.DefaultTransport, false, false, s)
			httpProxy.ServeHTTP(w, req)
			return true
		}
	}
	return false
}

func (s *reverseProxyDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}

func NewReverseProxyDispatcher(reverseProxyInformer extensionsinformers.ReverseProxyInformer) Dispatcher {
	return &reverseProxyDispatcher{reverseProxyInformer: reverseProxyInformer}
}
