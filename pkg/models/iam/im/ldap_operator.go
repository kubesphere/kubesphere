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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
)

type ldapOperator struct {
	ldapClient ldap.Interface
}

func NewLDAPOperator(ldapClient ldap.Interface) IdentityManagementInterface {
	return &ldapOperator{
		ldapClient: ldapClient,
	}
}

func (im *ldapOperator) UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {

	err := im.ldapClient.Update(user)

	if err != nil {
		return nil, err
	}

	return im.ldapClient.Get(user.Name)
}

func (im *ldapOperator) Authenticate(username, password string) (*iamv1alpha2.User, error) {

	user, err := im.ldapClient.Get(username)

	if err != nil {
		return nil, err
	}

	err = im.ldapClient.Authenticate(user.Name, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (im *ldapOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	return im.ldapClient.Get(username)
}

func (im *ldapOperator) DeleteUser(username string) error {
	return im.ldapClient.Delete(username)
}

func (im *ldapOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	err := im.ldapClient.Create(user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (im *ldapOperator) ListUsers(query *query.Query) (*api.ListResult, error) {
	result, err := im.ldapClient.List(query)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return result, nil
}
