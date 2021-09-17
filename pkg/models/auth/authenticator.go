/*

 Copyright 2020 The KubeSphere Authors.

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

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
)

var (
	RateLimitExceededError  = fmt.Errorf("auth rate limit exceeded")
	IncorrectPasswordError  = fmt.Errorf("incorrect password")
	AccountIsNotActiveError = fmt.Errorf("account is not active")
)

// PasswordAuthenticator is an interface implemented by authenticator which take a
// username and password.
type PasswordAuthenticator interface {
	Authenticate(ctx context.Context, username, password string) (authuser.Info, string, error)
}

type OAuthAuthenticator interface {
	Authenticate(ctx context.Context, provider string, req *http.Request) (authuser.Info, string, error)
}

type userGetter struct {
	userLister iamv1alpha2listers.UserLister
}

func preRegistrationUser(idp string, identity identityprovider.Identity) authuser.Info {
	return &authuser.DefaultInfo{
		Name: iamv1alpha2.PreRegistrationUser,
		Extra: map[string][]string{
			iamv1alpha2.ExtraIdentityProvider: {idp},
			iamv1alpha2.ExtraUID:              {identity.GetUserID()},
			iamv1alpha2.ExtraUsername:         {identity.GetUsername()},
			iamv1alpha2.ExtraEmail:            {identity.GetEmail()},
		},
	}
}

func mappedUser(idp string, identity identityprovider.Identity) *iamv1alpha2.User {
	// username convert
	username := strings.ToLower(identity.GetUsername())
	return &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Labels: map[string]string{
				iamv1alpha2.IdentifyProviderLabel: idp,
				iamv1alpha2.OriginUIDLabel:        identity.GetUserID(),
			},
		},
		Spec: iamv1alpha2.UserSpec{Email: identity.GetEmail()},
	}
}

// findUser returns the user associated with the username or email
func (u *userGetter) findUser(username string) (*iamv1alpha2.User, error) {
	if _, err := mail.ParseAddress(username); err != nil {
		return u.userLister.Get(username)
	}

	users, err := u.userLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, user := range users {
		if user.Spec.Email == username {
			return user, nil
		}
	}

	return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), username)
}

// findMappedUser returns the user which mapped to the identity
func (u *userGetter) findMappedUser(idp, uid string) (*iamv1alpha2.User, error) {
	selector := labels.SelectorFromSet(labels.Set{
		iamv1alpha2.IdentifyProviderLabel: idp,
		iamv1alpha2.OriginUIDLabel:        uid,
	})

	users, err := u.userLister.List(selector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(users) != 1 {
		return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), uid)
	}

	return users[0], err
}
