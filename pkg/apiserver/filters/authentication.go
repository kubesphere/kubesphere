package filters

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"
	"net/http"
)

// WithAuthentication installs authentication handler to handler chain.
func WithAuthentication(handler http.Handler, auth authenticator.Request, failed http.Handler) http.Handler {
	if auth == nil {
		klog.Warningf("Authentication is disabled")
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//authenticationStart := time.Now()

		resp, ok, err := auth.AuthenticateRequest(req)
		if err != nil || !ok {
			if err != nil {
				klog.Errorf("Unable to authenticate the request due to error: %v", err)
			}
			failed.ServeHTTP(w, req)
			return
		}

		// authorization header is not required anymore in case of a successful authentication.
		req.Header.Del("Authorization")

		req = req.WithContext(request.WithUser(req.Context(), resp.User))
		handler.ServeHTTP(w, req)
	})
}
