/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/transport"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type jsBundle struct {
	next  http.Handler
	cache cache.Cache
}

func WithJSBundle(next http.Handler, cache cache.Cache) http.Handler {
	return &jsBundle{next: next, cache: cache}
}

func (s *jsBundle) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsResourceRequest || requestInfo.IsKubernetesRequest {
		s.next.ServeHTTP(w, req)
		return
	}
	if !strings.HasPrefix(requestInfo.Path, extensionsv1alpha1.DistPrefix) {
		s.next.ServeHTTP(w, req)
		return
	}
	var jsBundles extensionsv1alpha1.JSBundleList
	if err := s.cache.List(req.Context(), &jsBundles, &client.ListOptions{}); err != nil {
		reason := "failed to list js bundles"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}
	for _, jsBundle := range jsBundles.Items {
		if jsBundle.Status.State == extensionsv1alpha1.StateAvailable && jsBundle.Status.Link == requestInfo.Path {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			if jsBundle.Spec.Raw != nil {
				s.rawContent(jsBundle.Spec.Raw, w, req)
				return
			}
			if jsBundle.Spec.RawFrom.ConfigMapKeyRef != nil {
				s.rawFromConfigMap(jsBundle.Spec.RawFrom.ConfigMapKeyRef, w, req)
				return
			}
			if jsBundle.Spec.RawFrom.SecretKeyRef != nil {
				s.rawFromSecret(jsBundle.Spec.RawFrom.SecretKeyRef, w, req)
				return
			}
			if jsBundle.Spec.RawFrom.URL != nil || jsBundle.Spec.RawFrom.Service != nil {
				s.rawFromRemote(jsBundle.Spec.RawFrom.Endpoint, w, req)
				return
			}
		}
	}
	s.next.ServeHTTP(w, req)
}

func (s *jsBundle) rawFromRemote(endpoint extensionsv1alpha1.Endpoint, w http.ResponseWriter, req *http.Request) {
	location, err := url.Parse(endpoint.RawURL())
	if err != nil {
		reason := "failed to fetch content"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	tr, err := transport.New(&transport.Config{
		TLS: transport.TLSConfig{
			CAData:   endpoint.CABundle,
			Insecure: endpoint.InsecureSkipVerify,
		},
	})

	if err != nil {
		reason := "failed to create transport.TLSConfig"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	handler := proxy.NewUpgradeAwareHandler(location, tr, false, false, &responder{})
	handler.UseLocationHost = true
	handler.ServeHTTP(w, req)
}

func (s *jsBundle) rawContent(base64EncodedData []byte, w http.ResponseWriter, _ *http.Request) {
	dec := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(base64EncodedData))
	if _, err := io.Copy(w, dec); err != nil {
		reason := "failed to decode raw content"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
	}
}

func (s *jsBundle) rawFromConfigMap(configMapRef *extensionsv1alpha1.ConfigMapKeyRef, w http.ResponseWriter, req *http.Request) {
	var cm v1.ConfigMap
	ref := types.NamespacedName{
		Namespace: configMapRef.Namespace,
		Name:      configMapRef.Name,
	}
	if err := s.cache.Get(req.Context(), ref, &cm); err != nil {
		reason := "failed to fetch content from configMap"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}
	if cm.Data != nil {
		_, _ = w.Write([]byte(cm.Data[configMapRef.Key]))
	} else if cm.BinaryData != nil {
		_, _ = w.Write(cm.BinaryData[configMapRef.Key])
	}
}

func (s *jsBundle) rawFromSecret(secretRef *extensionsv1alpha1.SecretKeyRef, w http.ResponseWriter, req *http.Request) {
	var secret v1.Secret
	ref := types.NamespacedName{
		Namespace: secretRef.Namespace,
		Name:      secretRef.Name,
	}
	if err := s.cache.Get(req.Context(), ref, &secret); err != nil {
		reason := "failed to fetch content from secret"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}
	_, _ = w.Write(secret.Data[secretRef.Key])
}
