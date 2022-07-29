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
	corev1 "k8s.io/client-go/informers/core/v1"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	extensionsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/extensions/v1alpha1"
)

type jsBundleDispatcher struct {
	jsBundleInformer  extensionsinformers.JSBundleInformer
	configMapInformer corev1.ConfigMapInformer
	secretInformer    corev1.SecretInformer
}

func NewJSBundleDispatcher(jsBundleInformer extensionsinformers.JSBundleInformer,
	configMapInformer corev1.ConfigMapInformer,
	secretInformer corev1.SecretInformer) Dispatcher {
	return &jsBundleDispatcher{jsBundleInformer: jsBundleInformer,
		configMapInformer: configMapInformer,
		secretInformer:    secretInformer}
}

func (s *jsBundleDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		return false
	}
	jsBundles, err := s.jsBundleInformer.Lister().List(labels.Everything())
	if err != nil {
		responsewriters.InternalError(w, req, err)
		return true
	}

	for _, jsBundle := range jsBundles {
		if jsBundle.Status.State == extensionsv1alpha1.StateEnabled && jsBundle.Status.Link == requestInfo.Path {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			if jsBundle.Spec.Raw != nil {
				w.Write(jsBundle.Spec.Raw)
				return true
			}
			if jsBundle.Spec.RawFrom.ConfigMapKeyRef != nil {
				cm, err := s.configMapInformer.Lister().ConfigMaps(jsBundle.Spec.RawFrom.ConfigMapKeyRef.Namespace).Get(jsBundle.Spec.RawFrom.ConfigMapKeyRef.Name)
				if err != nil {
					responsewriters.InternalError(w, req, err)
					return true
				}
				w.Write([]byte(cm.Data[jsBundle.Spec.RawFrom.ConfigMapKeyRef.Key]))
				return true
			}
			if jsBundle.Spec.RawFrom.SecretKeyRef != nil {
				secret, err := s.secretInformer.Lister().Secrets(jsBundle.Spec.RawFrom.SecretKeyRef.Namespace).Get(jsBundle.Spec.RawFrom.SecretKeyRef.Name)
				if err != nil {
					responsewriters.InternalError(w, req, err)
					return true
				}
				w.Write(secret.Data[jsBundle.Spec.RawFrom.ConfigMapKeyRef.Key])
				return true
			}
			var rawURL = jsBundle.Spec.RawFrom.RawURL()
			if rawURL != "" {
				location, err := url.Parse(rawURL)
				if err != nil {
					responsewriters.InternalError(w, req, err)
					return true
				}
				httpProxy := proxy.NewUpgradeAwareHandler(location, http.DefaultTransport, false, false, s)
				httpProxy.ServeHTTP(w, req)
				return true
			}
		}
	}
	return false
}

func (s *jsBundleDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
