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
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
)

type IdentityManagementInterface interface {
	CreateUser(user *iam.User) (*iam.User, error)
	DeleteUser(username string) error
	ModifyUser(user *iam.User) (*iam.User, error)
	DescribeUser(username string) (*iam.User, error)
	Authenticate(username, password string) (*iam.User, error)
}

type imOperator struct {
	ldapClient ldap.Interface
}

var (
	AuthRateLimitExceeded = errors.New("user auth rate limit exceeded")
	UserAlreadyExists     = errors.New("user already exists")
	UserNotExists         = errors.New("user not exists")
)

func NewLDAPOperator(ldapClient ldap.Interface) IdentityManagementInterface {
	return &imOperator{
		ldapClient: ldapClient,
	}

}

func (im *imOperator) ModifyUser(user *iam.User) (*iam.User, error) {

	err := im.ldapClient.Update(user)

	if err != nil {
		return nil, err
	}

	return im.ldapClient.Get(user.Name)
}

func (im *imOperator) Authenticate(username, password string) (*iam.User, error) {

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

func (im *imOperator) DescribeUser(username string) (*iam.User, error) {
	return im.ldapClient.Get(username)
}

func (im *imOperator) DeleteUser(username string) error {
	return im.ldapClient.Delete(username)
}

func (im *imOperator) CreateUser(user *iam.User) (*iam.User, error) {
	err := im.ldapClient.Create(user)

	if err != nil {
		return nil, err
	}

	return user, nil
}
