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
package im

import (
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	ListUsers(query *query.Query) (*api.ListResult, error)
	DeleteUser(username string) error
	UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	DescribeUser(username string) (*iamv1alpha2.User, error)
	ModifyPassword(username string, password string) error
	ListLoginRecords(query *query.Query) (*api.ListResult, error)
	PasswordVerify(username string, password string) error
}

func NewOperator(ksClient kubesphere.Interface, factory informers.InformerFactory, options *authoptions.AuthenticationOptions) IdentityManagementInterface {
	im := &defaultIMOperator{
		ksClient:       ksClient,
		resourceGetter: resourcev1alpha3.NewResourceGetter(factory),
		options:        options,
	}
	return im
}

type defaultIMOperator struct {
	ksClient       kubesphere.Interface
	resourceGetter *resourcev1alpha3.ResourceGetter
	options        *authoptions.AuthenticationOptions
}

func (im *defaultIMOperator) UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", user.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	old := obj.(*iamv1alpha2.User).DeepCopy()
	if user.Annotations == nil {
		user.Annotations = make(map[string]string, 0)
	}
	user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] = old.Annotations[iamv1alpha2.PasswordEncryptedAnnotation]
	user.Spec.EncryptedPassword = old.Spec.EncryptedPassword
	user.Status = old.Status

	updated, err := im.ksClient.IamV1alpha2().Users().Update(user)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return ensurePasswordNotOutput(updated), nil
}

func (im *defaultIMOperator) ModifyPassword(username string, password string) error {
	obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", username)

	if err != nil {
		klog.Error(err)
		return err
	}

	user := obj.(*iamv1alpha2.User).DeepCopy()
	delete(user.Annotations, iamv1alpha2.PasswordEncryptedAnnotation)
	user.Spec.EncryptedPassword = password

	_, err = im.ksClient.IamV1alpha2().Users().Update(user)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (im *defaultIMOperator) ListUsers(query *query.Query) (result *api.ListResult, err error) {
	result, err = im.resourceGetter.List(iamv1alpha2.ResourcesPluralUser, "", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)
	for _, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		out := ensurePasswordNotOutput(user)
		items = append(items, out)
	}
	result.Items = items
	return result, nil
}

func (im *defaultIMOperator) PasswordVerify(username string, password string) error {
	obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", username)
	if err != nil {
		klog.Error(err)
		return err
	}
	user := obj.(*iamv1alpha2.User)
	if checkPasswordHash(password, user.Spec.EncryptedPassword) {
		return nil
	}
	return AuthFailedIncorrectPassword
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (im *defaultIMOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	user := obj.(*iamv1alpha2.User)
	return ensurePasswordNotOutput(user), nil
}

func (im *defaultIMOperator) DeleteUser(username string) error {
	return im.ksClient.IamV1alpha2().Users().Delete(username, metav1.NewDeleteOptions(0))
}

func (im *defaultIMOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.ksClient.IamV1alpha2().Users().Create(user)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return user, nil
}

func (im *defaultIMOperator) ListLoginRecords(query *query.Query) (*api.ListResult, error) {
	result, err := im.resourceGetter.List(iamv1alpha2.ResourcesPluralLoginRecord, "", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

func ensurePasswordNotOutput(user *iamv1alpha2.User) *iamv1alpha2.User {
	out := user.DeepCopy()
	// ensure encrypted password will not be output
	out.Spec.EncryptedPassword = ""
	return out
}
