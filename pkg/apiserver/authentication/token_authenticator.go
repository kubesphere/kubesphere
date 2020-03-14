package authentication

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	cacheClient cache.Interface
}

func NewTokenAuthenticator(cacheClient cache.Interface) authenticator.Token {
	return &tokenAuthenticator{
		cacheClient: cacheClient,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	//if len(token) == 0 {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   user.Anonymous,
				UID:    "",
				Groups: []string{user.AllUnauthenticated},
				Extra:  nil,
			},
		}, true, nil
	//}
}
