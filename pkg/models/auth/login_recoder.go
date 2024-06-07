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
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type LoginRecorder interface {
	RecordLogin(ctx context.Context, username string, loginType iamv1beta1.LoginType, provider string, sourceIP string, userAgent string, authErr error) error
}

type loginRecorder struct {
	client     runtimeclient.Client
	userMapper userMapper
}

func NewLoginRecorder(cacheClient runtimeclient.Client) LoginRecorder {
	return &loginRecorder{
		client:     cacheClient,
		userMapper: userMapper{cache: cacheClient},
	}
}

// RecordLogin Create v1alpha2.LoginRecord for existing accounts
func (l *loginRecorder) RecordLogin(ctx context.Context, username string, loginType iamv1beta1.LoginType, provider, sourceIP, userAgent string, authErr error) error {
	// only for existing accounts, solve the problem of huge entries
	user, err := l.userMapper.Find(ctx, username)
	if err != nil {
		// ignore not found error
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}
	record := &iamv1beta1.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", user.Name),
			Labels: map[string]string{
				iamv1beta1.UserReferenceLabel: user.Name,
			},
		},
		Spec: iamv1beta1.LoginRecordSpec{
			Type:      loginType,
			Provider:  provider,
			Success:   true,
			Reason:    iamv1beta1.AuthenticatedSuccessfully,
			SourceIP:  sourceIP,
			UserAgent: userAgent,
		},
	}

	if authErr != nil {
		record.Spec.Success = false
		record.Spec.Reason = authErr.Error()
	}

	if err = l.client.Create(context.Background(), record); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
