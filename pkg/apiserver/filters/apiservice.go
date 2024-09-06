/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/transport"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type apiService struct {
	next  http.Handler
	cache cache.Cache
}

func WithAPIService(next http.Handler, cache cache.Cache) http.Handler {
	return &apiService{next: next, cache: cache}
}

func (s *apiService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		s.next.ServeHTTP(w, req)
		return
	}
	if !requestInfo.IsResourceRequest {
		s.next.ServeHTTP(w, req)
		return
	}
	var apiServices extensionsv1alpha1.APIServiceList
	if err := s.cache.List(req.Context(), &apiServices); err != nil {
		reason := fmt.Errorf("failed to list api services")
		klog.Errorf("%v: %v", reason, err)
		responsewriters.InternalError(w, req, errors.NewInternalError(reason))
		return
	}
	for _, apiService := range apiServices.Items {
		if apiService.Spec.Group != requestInfo.APIGroup || apiService.Spec.Version != requestInfo.APIVersion {
			continue
		}
		if apiService.Status.State != extensionsv1alpha1.StateAvailable {
			reason := fmt.Sprintf("apiService %s is not available", apiService.Name)
			responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
			return
		}
		s.handleProxyRequest(apiService, w, req)
		return
	}
	s.next.ServeHTTP(w, req)
}

func (s *apiService) handleProxyRequest(apiService extensionsv1alpha1.APIService, w http.ResponseWriter, req *http.Request) {
	endpoint, err := url.Parse(apiService.Spec.RawURL())
	if err != nil {
		reason := fmt.Sprintf("apiService %s is not available", apiService.Name)
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}
	location := &url.URL{}
	location.Scheme = endpoint.Scheme
	location.Host = endpoint.Host
	location.Path = req.URL.Path
	location.RawQuery = req.URL.Query().Encode()

	newReq := req.WithContext(req.Context())
	newReq.Header = utilnet.CloneHeader(req.Header)
	newReq.URL = location
	newReq.Host = location.Host

	tlsConfig := transport.TLSConfig{
		Insecure: apiService.Spec.InsecureSkipVerify,
	}
	if !apiService.Spec.InsecureSkipVerify && len(apiService.Spec.CABundle) > 0 {
		caData, err := base64.StdEncoding.DecodeString(string(apiService.Spec.CABundle))
		if err != nil {
			reason := "failed to base64 decode cabundle"
			klog.Warningf("%v: %v\n", reason, err)
			responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
			return
		}
		tlsConfig.CAData = caData
	}

	tr, err := transport.New(&transport.Config{
		TLS: tlsConfig,
	})

	if err != nil {
		reason := "failed to create transport.TLSConfig"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	user, _ := request.UserFrom(req.Context())
	proxyRoundTripper := transport.NewAuthProxyRoundTripper(user.GetName(), user.GetGroups(), user.GetExtra(), tr)

	upgrade := httpstream.IsUpgradeRequest(req)
	handler := proxy.NewUpgradeAwareHandler(location, proxyRoundTripper, false, upgrade, &responder{})
	handler.ServeHTTP(w, newReq)
}
