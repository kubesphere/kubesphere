/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"net/http"

	"k8s.io/apiserver/pkg/endpoints/responsewriter"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type auditingFilter struct {
	next http.Handler
	auditing.Auditing
}

func WithAuditing(next http.Handler, auditing auditing.Auditing) http.Handler {
	return &auditingFilter{
		next:     next,
		Auditing: auditing,
	}
}

func (a *auditingFilter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// When auditing level is LevelNone, request should not be auditing.
	// Auditing level can be modified with cr kube-auditing-webhook,
	// so it need to judge every time.
	if !a.Enabled() {
		a.next.ServeHTTP(w, req)
		return
	}

	info, ok := request.RequestInfoFrom(req.Context())
	if !ok {
		klog.Error("Unable to retrieve request info from request")
		a.next.ServeHTTP(w, req)
		return
	}

	if event := a.LogRequestObject(req, info); event != nil {
		resp := auditing.NewResponseCapture(w)
		a.next.ServeHTTP(responsewriter.WrapForHTTP1Or2(resp), req)
		go a.LogResponseObject(event, resp)
	} else {
		a.next.ServeHTTP(w, req)
	}
}
