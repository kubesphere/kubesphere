package jwttoken

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	token2 "kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	jwtTokenIssuer token2.Issuer
}

func NewTokenAuthenticator(issuer token2.Issuer) authenticator.Token {
	return &tokenAuthenticator{
		jwtTokenIssuer: issuer,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	providedUser, err := t.jwtTokenIssuer.Verify(token)
	if err != nil {
		return nil, false, err
	}

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   providedUser.GetName(),
			UID:    providedUser.GetUID(),
			Groups: []string{user.AllAuthenticated},
		},
	}, true, nil
}
