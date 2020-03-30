package filters

import (
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
)

// WithAuthentication installs authentication handler to handler chain.
func WithAuthentication(handler http.Handler, auth authenticator.Request) http.Handler {
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
			}

			ctx := req.Context()
			requestInfo, found := request.RequestInfoFrom(ctx)
			if !found {
				responsewriters.InternalError(w, req, errors.New("no RequestInfo found in the context"))
				return
			}

			gv := schema.GroupVersion{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion}
			responsewriters.ErrorNegotiated(apierrors.NewUnauthorized("Unauthorized"), s, gv, w, req)
			return
		}

		// authorization header is not required anymore in case of a successful authentication.
		req.Header.Del("Authorization")

		req = req.WithContext(request.WithUser(req.Context(), resp.User))
		handler.ServeHTTP(w, req)
	})
}
