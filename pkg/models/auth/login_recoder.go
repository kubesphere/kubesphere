/*
Copyright 2020 KubeSphere Authors

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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
)

type LoginRecorder interface {
	RecordLogin(username string, loginType iamv1alpha2.LoginType, provider string, sourceIP string, userAgent string, authErr error) error
}

type loginRecorder struct {
	ksClient   kubesphere.Interface
	userGetter *userGetter
}

func NewLoginRecorder(ksClient kubesphere.Interface, userLister iamv1alpha2listers.UserLister) LoginRecorder {
	return &loginRecorder{
		ksClient:   ksClient,
		userGetter: &userGetter{userLister: userLister},
	}
}

// RecordLogin Create v1alpha2.LoginRecord for existing accounts
func (l *loginRecorder) RecordLogin(username string, loginType iamv1alpha2.LoginType, provider, sourceIP, userAgent string, authErr error) error {
	// only for existing accounts, solve the problem of huge entries
	user, err := l.userGetter.findUser(username)
	if err != nil {
		// ignore not found error
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}
	loginEntry := &iamv1alpha2.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", user.Name),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel: user.Name,
			},
		},
		Spec: iamv1alpha2.LoginRecordSpec{
			Type:      loginType,
			Provider:  provider,
			Success:   true,
			Reason:    iamv1alpha2.AuthenticatedSuccessfully,
			SourceIP:  sourceIP,
			UserAgent: userAgent,
		},
	}

	if authErr != nil {
		loginEntry.Spec.Success = false
		loginEntry.Spec.Reason = authErr.Error()
	}

	_, err = l.ksClient.IamV1alpha2().LoginRecords().Create(context.Background(), loginEntry, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
