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

	"kubesphere.io/kubesphere/pkg/api"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type jsBundleDispatcher struct {
	cache cache.Cache
}

func NewJSBundleDispatcher(cache cache.Cache) Dispatcher {
	return &jsBundleDispatcher{cache: cache}
}

func (s *jsBundleDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		return false
	}
	var jsBundles extensionsv1alpha1.JSBundleList
	if err := s.cache.List(req.Context(), &jsBundles, &client.ListOptions{}); err != nil {
		responsewriters.InternalError(w, req, err)
		return true
	}

	for _, jsBundle := range jsBundles.Items {
		if jsBundle.Status.State == extensionsv1alpha1.StateAvailable && jsBundle.Status.Link == requestInfo.Path {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			if jsBundle.Spec.Raw != nil {
				w.Write(jsBundle.Spec.Raw)
				return true
			}
			if jsBundle.Spec.RawFrom.ConfigMapKeyRef != nil {
				var cm v1.ConfigMap
				if err := s.cache.Get(req.Context(), types.NamespacedName{
					Namespace: jsBundle.Spec.RawFrom.ConfigMapKeyRef.Namespace,
					Name:      jsBundle.Spec.RawFrom.ConfigMapKeyRef.Name,
				}, &cm); err != nil {
					responsewriters.InternalError(w, req, err)
					return true
				}
				w.Write([]byte(cm.Data[jsBundle.Spec.RawFrom.ConfigMapKeyRef.Key]))
				return true
			}
			if jsBundle.Spec.RawFrom.SecretKeyRef != nil {
				var secret v1.Secret
				if err := s.cache.Get(req.Context(), types.NamespacedName{
					Namespace: jsBundle.Spec.RawFrom.SecretKeyRef.Namespace,
					Name:      jsBundle.Spec.RawFrom.SecretKeyRef.Name,
				}, &secret); err != nil {
					responsewriters.InternalError(w, req, err)
					return true
				}
				w.Write(secret.Data[jsBundle.Spec.RawFrom.SecretKeyRef.Key])
				return true
			}
			var rawURL = jsBundle.Spec.RawFrom.RawURL()
			if rawURL != "" {
				location, err := url.Parse(rawURL)
				if err != nil {
					api.HandleServiceUnavailable(w, nil, err)
					return true
				}
				handler := proxy.NewUpgradeAwareHandler(location, http.DefaultTransport, false, false, s)
				handler.ServeHTTP(w, req)
				return true
			}
		}
	}
	return false
}

func (s *jsBundleDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
