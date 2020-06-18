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
		// Auditing level can be modified with cr kube-auditing-webhook,
		// so it need to judge every time.
		if !a.Enabled() {
			handler.ServeHTTP(w, req)
			return
		}

		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			klog.Error("Unable to retrieve request info from request")
			handler.ServeHTTP(w, req)
			return
		}

		// Auditing should igonre k8s request when k8s auditing is enabled.
		if info.IsKubernetesRequest && a.K8sAuditingEnabled() {
			handler.ServeHTTP(w, req)
			return
		}

		e := a.LogRequestObject(req, info)
		req = req.WithContext(request.WithAuditEvent(req.Context(), e))
		resp := auditing.NewResponseCapture(w)
		handler.ServeHTTP(resp, req)

		go a.LogResponseObject(e, resp, info)
	})
}
