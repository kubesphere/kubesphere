/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package basic

import (
	"context"
	"errors"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/basictoken"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/auth"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type basicAuthenticator struct {
	authenticator auth.PasswordAuthenticator
	loginRecorder auth.LoginRecorder
}

func NewBasicAuthenticator(authenticator auth.PasswordAuthenticator, loginRecorder auth.LoginRecorder) basictoken.Password {
	return &basicAuthenticator{
		authenticator: authenticator,
		loginRecorder: loginRecorder,
	}
}

func (t *basicAuthenticator) AuthenticatePassword(ctx context.Context, username, password string) (*authenticator.Response, bool, error) {
	authenticated, err := t.authenticator.Authenticate(ctx, "", username, password)
	if err != nil {
		if t.loginRecorder != nil && errors.Is(err, auth.IncorrectPasswordError) {
			var sourceIP, userAgent string
			if requestInfo, ok := request.RequestInfoFrom(ctx); ok {
				sourceIP = requestInfo.SourceIP
				userAgent = requestInfo.UserAgent
			}
			if err := t.loginRecorder.RecordLogin(ctx, username, iamv1beta1.Password, "", sourceIP, userAgent, err); err != nil {
				klog.Errorf("Failed to record unsuccessful login attempt for user %s, error: %v", username, err)
			}
		}
		return nil, false, err
	}
	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   authenticated.GetName(),
			Groups: append(authenticated.GetGroups(), user.AllAuthenticated),
		},
	}, true, nil
}
