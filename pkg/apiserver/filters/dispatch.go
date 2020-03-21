package filters

import (
	"fmt"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/dispatch"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
	"strings"
)

// Multiple cluster dispatcher forward request to desired cluster based on request cluster name
// which included in request path clusters/{cluster}
func WithMultipleClusterDispatcher(handler http.Handler, dispatch dispatch.Dispatcher) http.Handler {
	if dispatch == nil {
		klog.V(4).Infof("Multiple cluster dispatcher is disabled")
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			responsewriters.InternalError(w, req, fmt.Errorf(""))
			return
		}

		if info.Cluster == "host-cluster" || info.Cluster == "" {
			handler.ServeHTTP(w, req)
		} else {
			// remove cluster path
			req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)
			dispatch.Dispatch(w, req)
		}
	})
}
