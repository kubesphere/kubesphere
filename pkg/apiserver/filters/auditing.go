package filters

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
)

func WithAuditing(handler http.Handler, a auditing.Auditing) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if !a.Enable() {
			handler.ServeHTTP(w, req)
			return
		}

		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			klog.Error("Unable to retrieve request info from request")
			handler.ServeHTTP(w, req)
			return
		}

		if info.IsKubernetesRequest {
			handler.ServeHTTP(w, req)
			return
		}

		e := a.LogRequestObject(info, req)
		resp := auditing.NewResponseCapture(w)
		go handler.ServeHTTP(resp, req)

		select {
		case <-req.Context().Done():
			klog.Error("Server timeout")
			return
		case <-resp.StopCh:
			a.LogResponseObject(e, resp)
			return
		}
	})
}
