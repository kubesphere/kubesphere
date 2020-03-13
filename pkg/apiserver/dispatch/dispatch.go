package dispatch

import (
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"net/http"

	"k8s.io/apimachinery/pkg/util/proxy"
)

// Dispatcher defines how to forward request to desired cluster apiserver
type Dispatcher interface {
	Dispatch(w http.ResponseWriter, req *http.Request)
}

var DefaultClusterDispatch = newClusterDispatch()

type clusterDispatch struct {
	transport *http.Transport
}

func newClusterDispatch() Dispatcher {
	return &clusterDispatch{}
}

func (c *clusterDispatch) Dispatch(w http.ResponseWriter, req *http.Request) {

	u := *req.URL
	// u.Host = someHost

	httpProxy := proxy.NewUpgradeAwareHandler(&u, c.transport, false, false, c)
	httpProxy.ServeHTTP(w, req)
}

func (c *clusterDispatch) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
