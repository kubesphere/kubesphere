/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package im

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type LoginRecordInterface interface {
	RecordLoginLogs(username string, record iamv1alpha2.LoginRecordSpec) error
}

type loginRecoder struct {
	ksClient kubesphere.Interface
}

func NewLoginRecoder(ksClient kubesphere.Interface) LoginRecordInterface {
	return &loginRecoder{
		ksClient: ksClient,
	}
}

func (l loginRecoder) RecordLoginLogs(username string, record iamv1alpha2.LoginRecordSpec) error {
	loginRecord := &iamv1alpha2.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", username),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel: username,
			},
		},
		Spec: record,
	}
	_, err := l.ksClient.IamV1alpha2().LoginRecords().Create(loginRecord)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
