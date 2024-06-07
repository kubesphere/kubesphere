/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/transport"
	"k8s.io/klog/v2"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/utils/directives"
)

type reverseProxy struct {
	next               http.Handler
	cache              cache.Cache
	proxyRoundTrippers *sync.Map
}

func WithReverseProxy(next http.Handler, cache cache.Cache) http.Handler {
	return &reverseProxy{next: next, cache: cache, proxyRoundTrippers: &sync.Map{}}
}

func (s *reverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		s.next.ServeHTTP(w, req)
		return
	}
	if requestInfo.IsResourceRequest {
		s.next.ServeHTTP(w, req)
		return
	}

	if !strings.HasPrefix(requestInfo.Path, extensionsv1alpha1.ProxyPrefix) {
		s.next.ServeHTTP(w, req)
		return
	}

	var reverseProxies extensionsv1alpha1.ReverseProxyList
	// If the target label is not set, it is also handled by ks-apiserver (backward compatibility)
	selector, _ := labels.Parse(fmt.Sprintf("%s!=%s", extensionsv1alpha1.ReverseProxyTargetLabel, extensionsv1alpha1.ReverseProxyTargetConsole))
	if err := s.cache.List(req.Context(), &reverseProxies, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		reason := "failed to list reverse proxies"
		klog.Errorf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	for _, reverseProxy := range reverseProxies.Items {
		if !s.match(reverseProxy.Spec.Matcher, req) {
			continue
		}
		if reverseProxy.Status.State != extensionsv1alpha1.StateAvailable {
			responsewriters.WriteRawJSON(http.StatusServiceUnavailable, fmt.Errorf("upstream %s is not available", reverseProxy.Name), w)
			return
		}

		s.handleProxyRequest(reverseProxy, w, req)
		return

	}
	s.next.ServeHTTP(w, req)
}

func (s *reverseProxy) match(matcher extensionsv1alpha1.Matcher, req *http.Request) bool {
	if matcher.Method != req.Method && matcher.Method != "*" {
		return false
	}
	if matcher.Path == req.URL.Path {
		return true
	}
	if strings.HasSuffix(matcher.Path, "*") &&
		strings.HasPrefix(req.URL.Path, strings.TrimRight(matcher.Path, "*")) {
		return true
	}
	return false
}

func (s *reverseProxy) handleProxyRequest(reverseProxy extensionsv1alpha1.ReverseProxy, w http.ResponseWriter, req *http.Request) {
	endpoint, err := url.Parse(reverseProxy.Spec.Upstream.RawURL())
	if err != nil {
		reason := fmt.Sprintf("endpoint %s is not available", endpoint)
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

	if reverseProxy.Spec.Directives.Method != "" {
		newReq.Method = reverseProxy.Spec.Directives.Method
	}
	if reverseProxy.Spec.Directives.StripPathPrefix != "" {
		location.Path = strings.TrimPrefix(location.Path, reverseProxy.Spec.Directives.StripPathPrefix)
	}
	if reverseProxy.Spec.Directives.StripPathSuffix != "" {
		location.Path = strings.TrimSuffix(location.Path, reverseProxy.Spec.Directives.StripPathSuffix)
	}
	if len(reverseProxy.Spec.Directives.HeaderUp) > 0 {
		for _, header := range reverseProxy.Spec.Directives.HeaderUp {
			if strings.HasPrefix(header, "-") {
				removeHeader(newReq.Header, strings.TrimPrefix(header, "-"))
			} else if strings.HasPrefix(header, "+") {
				addOrReplaceHeader(newReq.Header, strings.TrimPrefix(header, "+"), false)
			} else {
				addOrReplaceHeader(newReq.Header, header, true)
			}
		}
	}

	if err = directives.HandlerRequest(newReq, reverseProxy.Spec.Directives.Rewrite, directives.WithRewriteFilter); err != nil {
		reason := "failed to create handler directives Directives.Rewrite"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	if err = directives.HandlerRequest(newReq, reverseProxy.Spec.Directives.Replace, directives.WithReplaceFilter); err != nil {
		reason := "failed to create handler directives Directives.Replace"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	if err = directives.HandlerRequest(newReq, reverseProxy.Spec.Directives.PathRegexp, directives.WithPathRegexpFilter); err != nil {
		reason := "failed to create handler directives Directives.PathRegexp"
		klog.Warningf("%v: %v\n", reason, err)
		responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
		return
	}

	var proxyRoundTripper http.RoundTripper
	if newProxyRoundTripper, ok := s.proxyRoundTrippers.Load(reverseProxy.Name); !ok {
		tlsConfig := transport.TLSConfig{
			Insecure: reverseProxy.Spec.Upstream.InsecureSkipVerify,
		}
		if !reverseProxy.Spec.Upstream.InsecureSkipVerify && len(reverseProxy.Spec.Upstream.CABundle) > 0 {
			caData, err := base64.StdEncoding.DecodeString(string(reverseProxy.Spec.Upstream.CABundle))
			if err != nil {
				reason := fmt.Sprintf("failed to decode CA bundle from upstream %s", reverseProxy.Name)
				klog.Warningf("%v: %v\n", reason, err)
				responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
				return
			}
			tlsConfig.CAData = caData
		}

		newProxyRoundTripper, err := transport.New(&transport.Config{
			TLS: tlsConfig,
			WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
				return &http.Transport{
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          0,
					MaxConnsPerHost:       0,
					MaxIdleConnsPerHost:   100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
					TLSClientConfig:       rt.(*http.Transport).TLSClientConfig,
				}
			},
		})
		if err != nil {
			reason := "failed to create transport.TLSConfig"
			klog.Warningf("%v: %v\n", reason, err)
			responsewriters.WriteRawJSON(http.StatusServiceUnavailable, errors.NewServiceUnavailable(reason), w)
			return
		}
		proxyRoundTripper = newProxyRoundTripper
		s.proxyRoundTrippers.Store(reverseProxy.Name, newProxyRoundTripper)
	} else {
		proxyRoundTripper = newProxyRoundTripper.(http.RoundTripper)
	}

	if reverseProxy.Spec.Directives.AuthProxy {
		user, _ := request.UserFrom(req.Context())
		proxyRoundTripper = transport.NewAuthProxyRoundTripper(user.GetName(), user.GetGroups(), user.GetExtra(), proxyRoundTripper)
	}

	upgrade := httpstream.IsUpgradeRequest(req)
	handler := proxy.NewUpgradeAwareHandler(location, proxyRoundTripper, false, upgrade, &responder{})
	if reverseProxy.Spec.Directives.WrapTransport {
		handler.WrapTransport = true
	}

	if len(reverseProxy.Spec.Directives.HeaderDown) > 0 {
		w = &responseWriterWrapper{
			ResponseWriter: w,
			HeaderDown:     reverseProxy.Spec.Directives.HeaderDown,
		}
	}

	handler.ServeHTTP(w, newReq)
}

func removeHeader(header http.Header, key string) {
	if strings.HasSuffix(key, "*") {
		prefix := strings.TrimSuffix(key, "*")
		for key := range header {
			if strings.HasSuffix(key, prefix) {
				header.Del(key)
			}
		}
	} else {
		header.Del(key)
	}
}

func addOrReplaceHeader(header http.Header, keyValues string, replace bool) {
	values := strings.SplitN(keyValues, " ", 2)
	if len(values) != 2 {
		return
	}
	key := values[0]
	value := values[1]
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "\""), "\"")
	}
	if replace {
		header.Set(key, value)
	} else {
		header.Add(key, value)
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	wroteHeader bool
	HeaderDown  []string
}

func (rww *responseWriterWrapper) WriteHeader(status int) {
	if rww.wroteHeader {
		return
	}
	rww.wroteHeader = true

	for _, header := range rww.HeaderDown {
		if strings.HasPrefix(header, "-") {
			removeHeader(rww.Header(), strings.TrimPrefix(header, "-"))
		} else if strings.HasPrefix(header, "+") {
			addOrReplaceHeader(rww.Header(), strings.TrimPrefix(header, "+"), false)
		} else {
			addOrReplaceHeader(rww.Header(), header, true)
		}
	}

	rww.ResponseWriter.WriteHeader(status)
}

func (rww *responseWriterWrapper) Write(d []byte) (int, error) {
	if !rww.wroteHeader {
		rww.WriteHeader(http.StatusOK)
	}
	return rww.ResponseWriter.Write(d)
}
