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
	"context"
	"fmt"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/models/auth"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	ListUsers(query *query.Query) (*api.ListResult, error)
	DeleteUser(username string) error
	UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	DescribeUser(username string) (*iamv1alpha2.User, error)
	ModifyPassword(username string, password string) error
	ListLoginRecords(username string, query *query.Query) (*api.ListResult, error)
	PasswordVerify(username string, password string) error
}

func NewOperator(ksClient kubesphere.Interface, userGetter resources.Interface, loginRecordGetter resources.Interface, options *authentication.Options) IdentityManagementInterface {
	im := &imOperator{
		ksClient:          ksClient,
		userGetter:        userGetter,
		loginRecordGetter: loginRecordGetter,
		options:           options,
	}
	return im
}

type imOperator struct {
	ksClient          kubesphere.Interface
	userGetter        resources.Interface
	loginRecordGetter resources.Interface
	options           *authentication.Options
}

// UpdateUser returns user information after update.
func (im *imOperator) UpdateUser(new *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	old, err := im.fetch(new.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// keep encrypted password and user status
	new.Spec.EncryptedPassword = old.Spec.EncryptedPassword
	new.Status = old.Status
	updated, err := im.ksClient.IamV1alpha2().Users().Update(context.Background(), new, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return ensurePasswordNotOutput(updated), nil
}

func (im *imOperator) fetch(username string) (*iamv1alpha2.User, error) {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	user := obj.(*iamv1alpha2.User).DeepCopy()
	return user, nil
}

func (im *imOperator) ModifyPassword(username string, password string) error {
	user, err := im.fetch(username)
	if err != nil {
		klog.Error(err)
		return err
	}
	user.Spec.EncryptedPassword = password
	_, err = im.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (im *imOperator) ListUsers(query *query.Query) (result *api.ListResult, err error) {
	result, err = im.userGetter.List("", query)
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

func (im *imOperator) PasswordVerify(username string, password string) error {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return err
	}
	user := obj.(*iamv1alpha2.User)
	if err = auth.PasswordVerify(user.Spec.EncryptedPassword, password); err != nil {
		return err
	}
	return nil
}

func (im *imOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	user := obj.(*iamv1alpha2.User)
	return ensurePasswordNotOutput(user), nil
}

func (im *imOperator) DeleteUser(username string) error {
	return im.ksClient.IamV1alpha2().Users().Delete(context.Background(), username, *metav1.NewDeleteOptions(0))
}

func (im *imOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.ksClient.IamV1alpha2().Users().Create(context.Background(), user, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return user, nil
}

func (im *imOperator) ListLoginRecords(username string, q *query.Query) (*api.ListResult, error) {
	q.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", iamv1alpha2.UserReferenceLabel, username))
	result, err := im.loginRecordGetter.List("", q)
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
