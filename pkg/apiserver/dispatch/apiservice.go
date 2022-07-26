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

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	extensionsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/extensions/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type apiServiceDispatcher struct {
	apiServiceInformer extensionsinformers.APIServiceInformer
}

func (s *apiServiceDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		return false
	}
	apiServices, err := s.apiServiceInformer.Lister().List(labels.Everything())
	if err != nil {
		responsewriters.InternalError(w, req, err)
		return true
	}
	for _, apiService := range apiServices {
		if apiService.Status.State == extensionsv1alpha1.StateEnabled && (sliceutil.HasString(apiService.Spec.NonResourceURLs, requestInfo.Path) ||
			(apiService.Spec.Group == requestInfo.APIGroup && apiService.Spec.Version == requestInfo.APIVersion)) {
			endpoint, err := url.Parse(apiService.Spec.Endpoint.RawURL())
			if err != nil {
				responsewriters.InternalError(w, req, err)
				return true
			}
			location := req.URL
			location.Host = endpoint.Host
			location.Scheme = endpoint.Scheme
			location.Path = endpoint.Path + location.Path
			// TODO support TLS transport
			httpProxy := proxy.NewUpgradeAwareHandler(req.URL, http.DefaultTransport, false, false, s)
			httpProxy.ServeHTTP(w, req)
			return true
		}
	}
	return false
}

func (s *apiServiceDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}

func NewAPIServiceDispatcher(apiServiceInformer extensionsinformers.APIServiceInformer) Dispatcher {
	return &apiServiceDispatcher{apiServiceInformer: apiServiceInformer}
}
