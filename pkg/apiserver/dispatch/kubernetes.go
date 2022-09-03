package dispatch

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

func NewKubernetesAPIDispatcher(config *rest.Config) (Dispatcher, error) {
	serverURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config.Host: %v", err)
	}
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create transport from rest.Config:: %v", err)
	}
	return &kubernetesAPIDispatcher{
		serverURL: serverURL,
		transport: transport,
	}, nil
}

type kubernetesAPIDispatcher struct {
	serverURL *url.URL
	transport http.RoundTripper
}

func (s *kubernetesAPIDispatcher) Dispatch(w http.ResponseWriter, req *http.Request) bool {
	requestInfo, _ := request.RequestInfoFrom(req.Context())
	if requestInfo.IsKubernetesRequest {
		location := &url.URL{}
		location.Scheme = s.serverURL.Scheme
		location.Host = s.serverURL.Host
		location.Path = req.URL.Path
		location.RawQuery = req.URL.Query().Encode()

		newReq := req.WithContext(req.Context())
		newReq.Header = utilnet.CloneHeader(req.Header)
		// make sure we don't override kubernetes's authorization
		newReq.Header.Del("Authorization")
		newReq.URL = location

		upgrade := httpstream.IsUpgradeRequest(req)
		handler := proxy.NewUpgradeAwareHandler(location, s.transport, false, upgrade, s)
		handler.UseLocationHost = true
		handler.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(s.transport, s.transport)
		handler.ServeHTTP(w, req)
		return true
	}
	return false
}

func (s *kubernetesAPIDispatcher) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
