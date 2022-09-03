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
	"strings"

	"k8s.io/client-go/transport"

	"k8s.io/apimachinery/pkg/util/httpstream"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type reverseProxyDispatcher struct {
	cache cache.Cache
}

func NewReverseProxyDispatcher(cache cache.Cache) Dispatcher {
	return &reverseProxyDispatcher{cache: cache}
}

func (s *reverseProxyDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	info, _ := request.RequestInfoFrom(req.Context())
	if info.IsKubernetesRequest {
		return false
	}
	var reverseProxies extensionsv1alpha1.ReverseProxyList
	if err := s.cache.List(req.Context(), &reverseProxies, &client.ListOptions{}); err != nil {
		responsewriters.InternalError(w, req, err)
		return true
	}

	for _, reverseProxy := range reverseProxies.Items {
		if reverseProxy.Status.State != extensionsv1alpha1.StateAvailable {
			continue
		}
		if !s.match(reverseProxy.Spec.Matcher, req) {
			continue
		}
		s.handleProxyRequest(reverseProxy, w, req)
		return true

	}
	return false
}

func (s *reverseProxyDispatcher) match(matcher extensionsv1alpha1.Matcher, req *http.Request) bool {
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

func (s *reverseProxyDispatcher) handleProxyRequest(reverseProxy extensionsv1alpha1.ReverseProxy, w http.ResponseWriter, req *http.Request) {
	endpoint, err := url.Parse(reverseProxy.Spec.Upstream.RawURL())
	if err != nil {
		api.HandleServiceUnavailable(w, nil, err)
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
				addHeader(newReq.Header, strings.TrimPrefix(header, "+"), false)
			} else {
				addHeader(newReq.Header, header, true)
			}
		}
	}

	proxyRoundTripper := http.DefaultTransport
	if reverseProxy.Spec.Directives.AuthProxy {
		user, _ := request.UserFrom(req.Context())
		proxyRoundTripper = transport.NewAuthProxyRoundTripper(user.GetName(), user.GetGroups(), user.GetExtra(), proxyRoundTripper)
	}

	upgrade := httpstream.IsUpgradeRequest(req)
	handler := proxy.NewUpgradeAwareHandler(location, proxyRoundTripper, false, upgrade, s)
	if reverseProxy.Spec.Directives.WrapTransport {
		handler.WrapTransport = true
	}
	if reverseProxy.Spec.Directives.ChangeOrigin {
		handler.UseLocationHost = true
	}
	if reverseProxy.Spec.Directives.InterceptRedirects {
		handler.InterceptRedirects = true
		handler.RequireSameHostRedirects = true
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

func addHeader(header http.Header, keyValues string, replace bool) {
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
			addHeader(rww.Header(), strings.TrimPrefix(header, "+"), false)
		} else {
			addHeader(rww.Header(), header, true)
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

func (s *reverseProxyDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
