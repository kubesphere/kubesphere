package filters

import (
	"context"
	"errors"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
)

// WithAuthorization passes all authorized requests on to handler, and returns forbidden error otherwise.
func WithAuthorization(handler http.Handler, a authorizer.Authorizer) http.Handler {
	if a == nil {
		klog.Warningf("Authorization is disabled")
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		attributes, err := GetAuthorizerAttributes(ctx)
		if err != nil {
			responsewriters.InternalError(w, req, err)
		}

		authorized, reason, err := a.Authorize(attributes)
		if authorized == authorizer.DecisionAllow {
			handler.ServeHTTP(w, req)
			return
		}

		if err != nil {
			responsewriters.InternalError(w, req, err)
			return
		}

		klog.V(4).Infof("Forbidden: %#v, Reason: %q", req.RequestURI, reason)
		w.WriteHeader(http.StatusForbidden)
	})
}

func GetAuthorizerAttributes(ctx context.Context) (authorizer.Attributes, error) {
	attribs := authorizer.AttributesRecord{}

	user, ok := k8srequest.UserFrom(ctx)
	if ok {
		attribs.User = user
	}

	requestInfo, found := request.RequestInfoFrom(ctx)
	if !found {
		return nil, errors.New("no RequestInfo found in the context")
	}

	// Start with common attributes that apply to resource and non-resource requests
	attribs.ResourceRequest = requestInfo.IsResourceRequest
	attribs.Path = requestInfo.Path
	attribs.Verb = requestInfo.Verb
	attribs.Cluster = requestInfo.Cluster
	attribs.Workspace = requestInfo.Workspace
	attribs.KubernetesRequest = requestInfo.IsKubernetesRequest

	attribs.APIGroup = requestInfo.APIGroup
	attribs.APIVersion = requestInfo.APIVersion
	attribs.Resource = requestInfo.Resource
	attribs.Subresource = requestInfo.Subresource
	attribs.Namespace = requestInfo.Namespace
	attribs.Name = requestInfo.Name

	return &attribs, nil
}
