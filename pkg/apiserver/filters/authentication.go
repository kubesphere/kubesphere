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
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

// WithAuthentication installs authentication handler to handler chain.
func WithAuthentication(handler http.Handler, auth authenticator.Request, loginRecorder im.LoginRecorder) http.Handler {
	if auth == nil {
		klog.Warningf("Authentication is disabled")
		return handler
	}
	s := serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp, ok, err := auth.AuthenticateRequest(req)
		if err != nil || !ok {
			if err != nil {
				klog.Errorf("Unable to authenticate the request due to error: %v", err)
				if err == im.AuthFailedIncorrectPassword { // log failed login attempts
					go func() {
						if loginRecorder != nil && resp != nil {
							err = loginRecorder.RecordLogin(resp.User.GetName(), err, req)
							klog.Errorf("Failed to record unsuccessful login attempt for user %s", resp.User.GetName())
						}
					}()
				}
			}

			ctx := req.Context()
			requestInfo, found := request.RequestInfoFrom(ctx)
			if !found {
				responsewriters.InternalError(w, req, errors.New("no RequestInfo found in the context"))
				return
			}

			gv := schema.GroupVersion{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion}
			responsewriters.ErrorNegotiated(apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)), s, gv, w, req)
			return
		}

		req = req.WithContext(request.WithUser(req.Context(), resp.User))
		handler.ServeHTTP(w, req)
	})
}
