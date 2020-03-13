package authentication

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

type TokenAuthenticator struct {
	cacheClient cache.Interface
}

func NewTokenAuthenticator(cacheClient cache.Interface) authenticator.Token {
	return &TokenAuthenticator{
		cacheClient: cacheClient,
	}
}

func (t *TokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   "admin",
			UID:    "",
			Groups: nil,
			Extra:  nil,
		},
	}, true, nil
}
