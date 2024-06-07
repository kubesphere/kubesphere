/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"net/mail"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type userMapper struct {
	cache runtimeclient.Reader
}

// Find returns the user associated with the username or email
func (u *userMapper) Find(ctx context.Context, username string) (*iamv1beta1.User, error) {
	user := &iamv1beta1.User{}
	if _, err := mail.ParseAddress(username); err != nil {
		return user, u.cache.Get(ctx, types.NamespacedName{Name: username}, user)
	}

	// TODO cache with index
	userList := &iamv1beta1.UserList{}
	if err := u.cache.List(ctx, userList); err != nil {
		return nil, err
	}

	for _, user := range userList.Items {
		if user.Spec.Email == username {
			return &user, nil
		}
	}

	return nil, errors.NewNotFound(iamv1beta1.Resource("user"), username)
}

// FindMappedUser returns the user which mapped to the identity
func (u *userMapper) FindMappedUser(ctx context.Context, idp, uid string) (*iamv1beta1.User, error) {
	userList := &iamv1beta1.UserList{}
	if err := u.cache.List(ctx, userList, runtimeclient.MatchingLabels{
		iamv1beta1.IdentifyProviderLabel: idp,
		iamv1beta1.OriginUIDLabel:        uid,
	}); err != nil {
		return nil, err
	}
	if len(userList.Items) != 1 {
		return nil, nil
	}
	return &userList.Items[0], nil
}
