/*
Copyright 2020 The KubeSphere Authors.

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
	"fmt"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/dispatch"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
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

		if info.Cluster == "" {
			handler.ServeHTTP(w, req)
		} else {
			dispatch.Dispatch(w, req, handler)
		}
	})
}
