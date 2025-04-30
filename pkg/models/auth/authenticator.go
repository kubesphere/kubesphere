/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

var (
	RateLimitExceededError  = fmt.Errorf("auth rate limit exceeded")
	IncorrectPasswordError  = fmt.Errorf("incorrect password")
	AccountIsNotActiveError = fmt.Errorf("account is not active")
)

// PasswordAuthenticator is an interface implemented by authenticator which take a
// username ,password and provider. provider refers to the identity provider`s name,
// if the provider is empty, authenticate from kubesphere account. Note that implement this
// interface you should also obey the error specification errors.Error defined at package
// "k8s.io/apimachinery/pkg/api", and restful.ServerError defined at package
// "github.com/emicklei/go-restful/v3", or the server cannot handle error correctly.
type PasswordAuthenticator interface {
	Authenticate(ctx context.Context, provider, username, password string) (authuser.Info, error)
}

// OAuthAuthenticator authenticate users by OAuth 2.0 Authorization Framework. Note that implement this
// interface you should also obey the error specification errors.Error defined at package
// "k8s.io/apimachinery/pkg/api", and restful.ServerError defined at package
// "github.com/emicklei/go-restful/v3", or the server cannot handle error correctly.
type OAuthAuthenticator interface {
	Authenticate(ctx context.Context, provider string, req *http.Request) (authuser.Info, error)
}

func newRreRegistrationUser(idp string, identity identityprovider.Identity) authuser.Info {
	return &authuser.DefaultInfo{
		Name: iamv1beta1.PreRegistrationUser,
		Extra: map[string][]string{
			iamv1beta1.ExtraIdentityProvider: {idp},
			iamv1beta1.ExtraUID:              {identity.GetUserID()},
			iamv1beta1.ExtraUsername:         {identity.GetUsername()},
			iamv1beta1.ExtraEmail:            {identity.GetEmail()},
		},
	}
}

func authByIdentityProvider(ctx context.Context, client client.Client, mapper UserMapper, providerConfig *identityprovider.Configuration, identity identityprovider.Identity) (authuser.Info, error) {
	mappedUser, err := mapper.FindMappedUser(ctx, providerConfig.Name, identity.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("failed to find mapped user: %s", err)
	}

	if mappedUser.Name == "" {
		if providerConfig.MappingMethod == identityprovider.MappingMethodLookup {
			return nil, fmt.Errorf("failed to find mapped user: %s", identity.GetUserID())
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodManual {
			return newRreRegistrationUser(providerConfig.Name, identity), nil
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodAuto {
			mappedUser := iamv1beta1.User{ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(identity.GetUsername())}}

			op, err := ctrl.CreateOrUpdate(ctx, client, &mappedUser, func() error {

				if mappedUser.Status.State == iamv1beta1.UserDisabled {
					return AccountIsNotActiveError
				}

				if mappedUser.Annotations == nil {
					mappedUser.Annotations = make(map[string]string)
				}
				mappedUser.Annotations[fmt.Sprintf("%s.%s", iamv1beta1.IdentityProviderAnnotation, providerConfig.Name)] = identity.GetUserID()
				mappedUser.Status.State = iamv1beta1.UserActive
				if identity.GetEmail() != "" {
					mappedUser.Spec.Email = identity.GetEmail()
				}
				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("failed to create or update user %s, error: %v", mappedUser.Name, err)
			}

			klog.V(4).Infof("user %s has been updated successfully, operation: %s", mappedUser.Name, op)

			return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
		}

		return nil, fmt.Errorf("invalid mapping method found %s", providerConfig.MappingMethod)
	}

	if mappedUser.Status.State == iamv1beta1.UserDisabled {
		return nil, AccountIsNotActiveError
	}

	return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
}
