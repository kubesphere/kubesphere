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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"strings"
)

type LoginRecorder interface {
	RecordLogin(username string, loginType iamv1alpha2.LoginType, provider string, sourceIP string, userAgent string, authErr error) error
}

type loginRecorder struct {
	ksClient kubesphere.Interface
}

func NewLoginRecorder(ksClient kubesphere.Interface) LoginRecorder {
	return &loginRecorder{
		ksClient: ksClient,
	}
}

func (l *loginRecorder) RecordLogin(username string, loginType iamv1alpha2.LoginType, provider string, sourceIP string, userAgent string, authErr error) error {
	// This is a temporary solution in case of user login with email,
	// '@' is not allowed in Kubernetes object name.
	username = strings.Replace(username, "@", "-", -1)

	loginEntry := &iamv1alpha2.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", username),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel: username,
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

	_, err := l.ksClient.IamV1alpha2().LoginRecords().Create(context.Background(), loginEntry, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
