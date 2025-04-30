/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"fmt"
	"net/mail"

	"k8s.io/apimachinery/pkg/types"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type UserMapper interface {
	Find(ctx context.Context, username string) (iamv1beta1.User, error)
	FindMappedUser(ctx context.Context, idp, uid string) (iamv1beta1.User, error)
}

type userMapper struct {
	cache runtimeclient.Reader
}

func (u *userMapper) Find(ctx context.Context, username string) (iamv1beta1.User, error) {
	user, err := u.getUserByUsernameOrEmail(ctx, username)
	if err != nil {
		return iamv1beta1.User{}, err
	}
	return user, nil
}

func (u *userMapper) FindMappedUser(ctx context.Context, idp, uid string) (iamv1beta1.User, error) {
	user, err := u.getUserByAnnotation(ctx, fmt.Sprintf("%s.%s", iamv1beta1.IdentityProviderAnnotation, idp), uid)
	if err != nil {
		return iamv1beta1.User{}, err
	}
	return user, nil
}

func (u *userMapper) getUserByUsernameOrEmail(ctx context.Context, username string) (iamv1beta1.User, error) {
	if _, err := mail.ParseAddress(username); err == nil {
		return u.getUserByEmail(ctx, username)
	}
	return u.getUserByName(ctx, username)
}

func (u *userMapper) getUserByName(ctx context.Context, username string) (iamv1beta1.User, error) {
	user := iamv1beta1.User{}
	err := u.cache.Get(ctx, types.NamespacedName{Name: username}, &user)
	return user, runtimeclient.IgnoreNotFound(err)
}

func (u *userMapper) getUserByEmail(ctx context.Context, email string) (iamv1beta1.User, error) {
	users := &iamv1beta1.UserList{}
	if err := u.cache.List(ctx, users); err != nil {
		return iamv1beta1.User{}, err
	}
	for _, user := range users.Items {
		if user.Spec.Email == email {
			return user, nil
		}
	}
	return iamv1beta1.User{}, nil
}

func (u *userMapper) getUserByAnnotation(ctx context.Context, annotation, value string) (iamv1beta1.User, error) {
	users := &iamv1beta1.UserList{}
	if err := u.cache.List(ctx, users); err != nil {
		return iamv1beta1.User{}, err
	}
	for _, user := range users.Items {
		if user.Annotations[annotation] == value {
			return user, nil
		}
	}
	return iamv1beta1.User{}, nil
}
