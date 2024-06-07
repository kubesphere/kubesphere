/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package jwt

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/utils/serviceaccount"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	tokenOperator auth.TokenManagementInterface
	cache         runtimecache.Cache
	clusterRole   string
}

func NewTokenAuthenticator(cache runtimecache.Cache, tokenOperator auth.TokenManagementInterface, clusterRole string) authenticator.Token {
	return &tokenAuthenticator{
		tokenOperator: tokenOperator,
		cache:         cache,
		clusterRole:   clusterRole,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	verified, err := t.tokenOperator.Verify(token)
	if err != nil {
		klog.Warning(err)
		return nil, false, err
	}

	if serviceaccount.IsServiceAccountToken(verified.Subject) {
		if t.clusterRole == string(clusterv1alpha1.ClusterRoleHost) {
			_, err = t.validateServiceAccount(ctx, verified)
			if err != nil {
				return nil, false, err
			}
		}
		return &authenticator.Response{
			User: verified.User,
		}, true, nil
	}

	if verified.User.GetName() == iamv1beta1.PreRegistrationUser {
		return &authenticator.Response{
			User: verified.User,
		}, true, nil
	}

	if t.clusterRole == string(clusterv1alpha1.ClusterRoleHost) {
		userInfo := &iamv1beta1.User{}
		if err := t.cache.Get(ctx, types.NamespacedName{Name: verified.User.GetName()}, userInfo); err != nil {
			return nil, false, err
		}

		// AuthLimitExceeded state should be ignored
		if userInfo.Status.State == iamv1beta1.UserDisabled {
			return nil, false, auth.AccountIsNotActiveError
		}
	}

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name: verified.User.GetName(),
			// TODO(wenhaozhou) Add user`s groups(can be searched by GroupBinding)
			Groups: []string{user.AllAuthenticated},
		},
	}, true, nil
}

func (t *tokenAuthenticator) validateServiceAccount(ctx context.Context, verify *token.VerifiedResponse) (*corev1alpha1.ServiceAccount, error) {
	// Ensure the relative service account exist
	name, namespace := serviceaccount.SplitUsername(verify.Username)
	sa := &corev1alpha1.ServiceAccount{}
	if err := t.cache.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, sa); err != nil {
		return nil, err
	}
	return sa, nil
}
