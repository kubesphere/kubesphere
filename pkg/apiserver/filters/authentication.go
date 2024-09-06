/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"errors"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type authnFilter struct {
	next http.Handler
	authenticator.Request
	serializer runtime.NegotiatedSerializer
}

// WithAuthentication installs authentication handler to handler chain.
// The following part is a little bit ugly, WithAuthentication also logs user failed login attempt
// if using basic auth. But only treats request with requestURI `/oauth/authorize` as login attempt
func WithAuthentication(next http.Handler, authenticator authenticator.Request) http.Handler {
	if authenticator == nil {
		klog.Warningf("Authentication is disabled")
		return next
	}
	return &authnFilter{
		next:       next,
		Request:    authenticator,
		serializer: serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion(),
	}
}

func (a *authnFilter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp, ok, err := a.AuthenticateRequest(req)
	_, _, usingBasicAuth := req.BasicAuth()

	defer func() {
		// if we authenticated successfully, go ahead and remove the bearer token so that no one
		// is ever tempted to use it inside the API server
		if usingBasicAuth && ok {
			req.Header.Del("Authorization")
		}
	}()

	if err != nil || !ok {
		ctx := req.Context()
		requestInfo, found := request.RequestInfoFrom(ctx)
		if !found {
			responsewriters.InternalError(w, req, errors.New("no RequestInfo found in the context"))
			return
		}
		if err != nil {
			klog.Errorf("Request authentication failed: %v", err)
		}
		gv := schema.GroupVersion{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion}
		if err != nil {
			err = apierrors.NewUnauthorized(err.Error())
		} else {
			err = apierrors.NewUnauthorized("The request cannot be authenticated.")
		}
		responsewriters.ErrorNegotiated(err, a.serializer, gv, w, req)
		return
	}

	req = req.WithContext(request.WithUser(req.Context(), resp.User))
	a.next.ServeHTTP(w, req)
}
