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

package basic

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type basicAuthenticator struct {
	authenticator im.PasswordAuthenticator
	loginRecorder im.LoginRecordInterface
}

func NewBasicAuthenticator(loginRecorder im.LoginRecordInterface, authenticator im.PasswordAuthenticator) authenticator.Password {
	return &basicAuthenticator{
		authenticator: authenticator,
		loginRecorder: loginRecorder,
	}
}

func (t *basicAuthenticator) AuthenticatePassword(ctx context.Context, username, password string) (*authenticator.Response, bool, error) {

	providedUser, err := t.authenticator.Authenticate(username, password)

	if err != nil {
		if err == im.AuthFailedIncorrectPassword {
			klog.V(4).Info(err)
			if err := t.loginRecorder.RecordLoginLogs(username, iamv1alpha2.LoginRecordSpec{
				Type:   iamv1alpha2.LoginFailure,
				Reason: err.Error(),
			}); err != nil {
				klog.Error(err)
				err := apierrors.NewInternalError(err)
				return nil, false, err
			}
		}
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
