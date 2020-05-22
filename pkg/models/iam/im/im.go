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
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"net/mail"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	ListUsers(query *query.Query) (*api.ListResult, error)
	DeleteUser(username string) error
	UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	DescribeUser(username string) (*iamv1alpha2.User, error)
	Authenticate(username, password string) (*iamv1alpha2.User, error)
}

var (
	AuthRateLimitExceeded       = errors.New("user auth rate limit exceeded")
	AuthFailedIncorrectPassword = errors.New("incorrect password")
	UserAlreadyExists           = errors.New("user already exists")
	UserNotExists               = errors.New("user not exists")
)

func NewOperator(ksClient kubesphereclient.Interface, factory informers.InformerFactory) IdentityManagementInterface {

	return &defaultIMOperator{
		ksClient:       ksClient,
		resourceGetter: resourcev1alpha3.NewResourceGetter(factory),
	}

}

type defaultIMOperator struct {
	ksClient       kubesphereclient.Interface
	resourceGetter *resourcev1alpha3.ResourceGetter
}

func (im *defaultIMOperator) UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", user.Name)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	old := obj.(*iamv1alpha2.User).DeepCopy()
	user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] = old.Annotations[iamv1alpha2.PasswordEncryptedAnnotation]
	user.Spec.EncryptedPassword = old.Spec.EncryptedPassword

	return im.ksClient.IamV1alpha2().Users().Update(user)
}

func (im *defaultIMOperator) Authenticate(username, password string) (*iamv1alpha2.User, error) {

	var user *iamv1alpha2.User

	if _, err := mail.ParseAddress(username); err != nil {

		obj, err := im.resourceGetter.Get(iamv1alpha2.ResourcesPluralUser, "", username)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		user = obj.(*iamv1alpha2.User)
	} else {
		objs, err := im.resourceGetter.List(iamv1alpha2.ResourcesPluralUser, "", &query.Query{
			Pagination: query.NoPagination,
			Filters:    map[query.Field]query.Value{iamv1alpha2.FieldEmail: query.Value(username)},
		})

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if len(objs.Items) != 1 {
			if len(objs.Items) == 0 {
				klog.Warningf("username or email: %s not exist", username)
			} else {
				klog.Errorf("duplicate user entries: %+v", objs)
			}
			return nil, AuthFailedIncorrectPassword
		}

		user = objs.Items[0].(*iamv1alpha2.User)
	}

	if checkPasswordHash(password, user.Spec.EncryptedPassword) {
		return user, nil
	}

	return nil, AuthFailedIncorrectPassword
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
		out := user.DeepCopy()
		// ensure encrypted password will not be output
		out.Spec.EncryptedPassword = ""
		items = append(items, out)
	}

	result.Items = items

	return result, nil
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
	out := user.DeepCopy()
	// ensure encrypted password will not be output
	out.Spec.EncryptedPassword = ""

	return out, nil
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
