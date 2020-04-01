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
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

func NewOperator(ksClient kubesphereclient.Interface, informer informers.SharedInformerFactory) IdentityManagementInterface {

	return &defaultIMOperator{
		ksClient: ksClient,
		informer: informer,
	}

}

type defaultIMOperator struct {
	ksClient kubesphereclient.Interface
	informer informers.SharedInformerFactory
}

func (im *defaultIMOperator) ModifyUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	return im.ksClient.IamV1alpha2().Users().Update(user)
}

func (im *defaultIMOperator) Authenticate(username, password string) (*iamv1alpha2.User, error) {

	user, err := im.ksClient.IamV1alpha2().Users().Get(username, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	if checkPasswordHash(password, user.Spec.EncryptedPassword) {
		return user, nil
	}
	return nil, AuthFailedIncorrectPassword
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (im *defaultIMOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	user, err := im.ksClient.IamV1alpha2().Users().Get(username, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (im *defaultIMOperator) DeleteUser(username string) error {
	return im.ksClient.IamV1alpha2().Users().Delete(username, metav1.NewDeleteOptions(0))
}

func (im *defaultIMOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.ksClient.IamV1alpha2().Users().Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
