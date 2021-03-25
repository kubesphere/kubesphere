/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jwttoken

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/auth"

	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	tokenOperator auth.TokenManagementInterface
	userLister    iamv1alpha2listers.UserLister
}

func NewTokenAuthenticator(tokenOperator auth.TokenManagementInterface, userLister iamv1alpha2listers.UserLister) authenticator.Token {
	return &tokenAuthenticator{
		tokenOperator: tokenOperator,
		userLister:    userLister,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	providedUser, err := t.tokenOperator.Verify(token)
	if err != nil {
		klog.Warning(err)
		return nil, false, err
	}

	if providedUser.GetName() == iamv1alpha2.PreRegistrationUser {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   providedUser.GetName(),
				Extra:  providedUser.GetExtra(),
				Groups: providedUser.GetGroups(),
			},
		}, true, nil
	}

	dbUser, err := t.userLister.Get(providedUser.GetName())
	if err != nil {
		return nil, false, err
	}

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   dbUser.GetName(),
			Groups: append(dbUser.Spec.Groups, user.AllAuthenticated),
		},
	}, true, nil
}
