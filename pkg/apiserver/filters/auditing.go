/*
Copyright 2020 KubeSphere Authors

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

package filters

import (
	"net/http"

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

	// Auditing should ignore k8s request when k8s auditing is enabled.
	if info.IsKubernetesRequest && a.K8sAuditingEnabled() {
		a.next.ServeHTTP(w, req)
		return
	}

	if event := a.LogRequestObject(req, info); event != nil {
		resp := auditing.NewResponseCapture(w)
		a.next.ServeHTTP(resp, req)
		go a.LogResponseObject(event, resp)
	} else {
		a.next.ServeHTTP(w, req)
	}
}
