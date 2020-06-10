package filters

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
)

func WithAuditing(handler http.Handler, a auditing.Auditing) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// When auditing level is LevelNone, request should not be auditing.
		if !a.Enable() {
			handler.ServeHTTP(w, req)
			return
		}

		e := a.LogRequestObject(req)
		resp := auditing.NewResponseCapture(w)

		// Create a new goroutine to finish the request, and wait for the response body.
		// The advantage of using goroutine is that recording the return value of the
		// request will not affect the processing of the request, even if the auditing fails.
		go handler.ServeHTTP(resp, req)

		select {
		case <-req.Context().Done():
			klog.Error("Server timeout")
			return
		case <-resp.StopCh:
			info, ok := request.RequestInfoFrom(req.Context())
			if !ok {
				klog.Error("Unable to retrieve request info from request")
				return
			}
			a.LogResponseObject(e, resp, info)
			return
		}
	})
}
