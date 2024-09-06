/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package openapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

type certKeyFunc func() ([]byte, []byte)

const (
	aggregatedDiscoveryTimeout = 5 * time.Second
)

type ApiService interface {
	Name() string
	ResolveEndpoint() (*url.URL, error)
	UpdateAPIService() error
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type defaultApiService struct {
	apiService                 *extensionsv1alpha1.APIService
	proxyTransport             *http.Transport
	restConfig                 *restclient.Config
	proxyRoundTripper          http.RoundTripper
	proxyCurrentCertKeyContent certKeyFunc
}

func NewApiService(apiService *extensionsv1alpha1.APIService) ApiService {
	return &defaultApiService{
		apiService:                 apiService,
		proxyCurrentCertKeyContent: func() (bytes []byte, bytes2 []byte) { return nil, nil },
	}
}

func (d *defaultApiService) Name() string {
	return d.apiService.Name
}

func (d *defaultApiService) ResolveEndpoint() (*url.URL, error) {

	if d.apiService.Spec.Service != nil &&
		d.apiService.Spec.Service.Name != "" &&
		d.apiService.Spec.Service.Namespace != "" &&
		*d.apiService.Spec.Service.Port != 0 {
		return &url.URL{Scheme: "https", Host: fmt.Sprintf("%s.%s.svc:%d",
			d.apiService.Spec.Service.Name, d.apiService.Spec.Service.Namespace, d.apiService.Spec.Service.Port)}, nil
	}
	if d.apiService.Spec.URL != nil && *d.apiService.Spec.URL != "" {
		u, err := url.Parse(*d.apiService.Spec.URL)
		if err != nil {
			return nil, err
		}
		if d.apiService.Spec.InsecureSkipVerify {
			u.Scheme = "http"
		}
		return u, nil
	}
	return nil, fmt.Errorf("cannot resolve an apiservice %s", d.Name())
}

func (d *defaultApiService) UpdateAPIService() error {
	proxyClientCert, proxyClientKey := d.proxyCurrentCertKeyContent()

	tlsConfig := restclient.TLSClientConfig{
		Insecure: d.apiService.Spec.InsecureSkipVerify,
	}

	if !d.apiService.Spec.InsecureSkipVerify && len(d.apiService.Spec.CABundle) > 0 {
		caData, err := base64.StdEncoding.DecodeString(string(d.apiService.Spec.CABundle))
		if err != nil {
			klog.Warning(err.Error())
			return err
		}
		tlsConfig.ServerName = d.apiService.Spec.Service.Name + "." + d.apiService.Spec.Service.Namespace + ".svc"
		tlsConfig.CertData = proxyClientCert
		tlsConfig.KeyData = proxyClientKey
		tlsConfig.CAData = caData
	}
	d.restConfig = &restclient.Config{
		TLSClientConfig: tlsConfig,
	}

	if d.proxyTransport != nil && d.proxyTransport.DialContext != nil {
		d.restConfig.Dial = d.proxyTransport.DialContext
	}
	proxyRoundTripper, err := restclient.TransportFor(d.restConfig)
	if err != nil {
		klog.Warning(err.Error())
		return err
	}
	d.proxyRoundTripper = proxyRoundTripper
	return nil
}

func (d *defaultApiService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//user, ok := genericapirequest.UserFrom(req.Context())
	//if !ok {
	//	proxyError(w, req, "missing user", http.StatusInternalServerError)
	//	return
	//}

	// write a new location based on the existing request pointed at the target service
	location, err := d.ResolveEndpoint()
	if err != nil {
		klog.Errorf("error resolving %s: %v", d.Name(), err)
		proxyError(w, req, "service unavailable", http.StatusServiceUnavailable)
	}
	location.Path = req.URL.Path
	location.RawQuery = req.URL.Query().Encode()

	newReq, cancelFn := newRequestForProxy(location, req)
	defer cancelFn()

	if d.proxyRoundTripper == nil {
		proxyError(w, req, "", http.StatusNotFound)
		return
	}

	proxyRoundTripper := d.proxyRoundTripper
	upgrade := httpstream.IsUpgradeRequest(req)

	//proxyRoundTripper = transport.NewAuthProxyRoundTripper(user.GetName(), user.GetGroups(), user.GetExtra(), proxyRoundTripper)

	//if upgrade {
	//transport.SetAuthProxyHeaders(newReq, user.GetName(), user.GetGroups(), user.GetExtra())
	//}

	handler := proxy.NewUpgradeAwareHandler(location, proxyRoundTripper, true, upgrade, &responder{w: w})
	handler.ServeHTTP(w, newReq)
}

type responder struct {
	w http.ResponseWriter
}

func (r *responder) Object(statusCode int, obj runtime.Object) {
	responsewriters.WriteRawJSON(statusCode, obj, r.w)
}

func (r *responder) Error(_ http.ResponseWriter, _ *http.Request, err error) {
	http.Error(r.w, err.Error(), http.StatusServiceUnavailable)
}

func proxyError(w http.ResponseWriter, req *http.Request, error string, code int) {
	http.Error(w, error, code)
}

// newRequestForProxy returns a shallow copy of the original request with a context that may include a timeout for discovery requests
func newRequestForProxy(location *url.URL, req *http.Request) (*http.Request, context.CancelFunc) {
	newCtx := req.Context()
	cancelFn := func() {}

	if requestInfo, ok := genericapirequest.RequestInfoFrom(req.Context()); ok {
		// trim leading and trailing slashes. Then "/apis/group/version" requests are for discovery, so if we have exactly three
		// segments that we are going to proxy, we have a discovery request.
		if !requestInfo.IsResourceRequest && len(strings.Split(strings.Trim(requestInfo.Path, "/"), "/")) == 3 {
			// discovery requests are used by kubectl and others to determine which resources a server has.  This is a cheap call that
			// should be fast for every aggregated apiserver.  Latency for aggregation is expected to be low (as for all extensions)
			// so forcing a short timeout here helps responsiveness of all clients.
			newCtx, cancelFn = context.WithTimeout(newCtx, aggregatedDiscoveryTimeout)
		}
	}

	// WithContext creates a shallow clone of the request with the same context.
	newReq := req.WithContext(newCtx)
	newReq.Header = utilnet.CloneHeader(req.Header)
	newReq.URL = location
	newReq.Host = location.Host

	return newReq, cancelFn
}
