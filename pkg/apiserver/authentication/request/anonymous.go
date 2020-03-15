package request

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"net/http"
	"strings"
)

type AnonymousAuthenticator struct{}

func (a *AnonymousAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth == "" {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   user.Anonymous,
				UID:    "",
				Groups: []string{user.AllUnauthenticated},
			},
		}, true, nil
	}
	return nil, false, nil
}
