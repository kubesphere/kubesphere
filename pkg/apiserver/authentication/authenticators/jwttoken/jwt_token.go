package jwttoken

import (
	"context"
	"fmt"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/api/iam/token"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

var errTokenExpired = errors.New("expired token")

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	cacheClient    cache.Interface
	jwtTokenIssuer token.Issuer
}

func NewTokenAuthenticator(cacheClient cache.Interface, jwtSecret string) authenticator.Token {
	return &tokenAuthenticator{
		cacheClient:    cacheClient,
		jwtTokenIssuer: token.NewJwtTokenIssuer(token.DefaultIssuerName, []byte(jwtSecret)),
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	providedUser, err := t.jwtTokenIssuer.Verify(token)
	if err != nil {
		return nil, false, err
	}

	// TODO implement token cache
	//_, err = t.cacheClient.Get(tokenKeyForUsername(providedUser.Name(), token))
	//if err != nil {
	//	return nil, false, errTokenExpired
	//}

	// Should we need to refresh token?

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   providedUser.GetName(),
			UID:    providedUser.GetUID(),
			Groups: []string{user.AllAuthenticated},
		},
	}, true, nil

}

func tokenKeyForUsername(username, token string) string {
	return fmt.Sprintf("kubesphere:users:%s:token:%s", username, token)
}
