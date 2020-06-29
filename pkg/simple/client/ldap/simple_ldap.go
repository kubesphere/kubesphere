/*
Copyright 2019 The KubeSphere Authors.

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

package ldap

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

const FAKE_HOST string = "FAKE"

// simpleLdap is a implementation of ldap.Interface, you should never use this in production env!
type simpleLdap struct {
	store map[string]*iamv1alpha2.User
}

func NewSimpleLdap() Interface {
	sl := &simpleLdap{
		store: map[string]*iamv1alpha2.User{},
	}

	// initialize with a admin user
	admin := &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admin",
		},
		Spec: iamv1alpha2.UserSpec{
			Email:             "admin@kubesphere.io",
			Lang:              "eng",
			Description:       "administrator",
			Groups:            nil,
			EncryptedPassword: "P@88w0rd",
		},
	}
	sl.store[admin.Name] = admin
	return sl
}

func (s simpleLdap) Create(user *iamv1alpha2.User) error {
	s.store[user.Name] = user
	return nil
}

func (s simpleLdap) Update(user *iamv1alpha2.User) error {
	_, err := s.Get(user.Name)
	if err != nil {
		return err
	}
	s.store[user.Name] = user
	return nil
}

func (s simpleLdap) Delete(name string) error {
	_, err := s.Get(name)
	if err != nil {
		return err
	}
	delete(s.store, name)
	return nil
}

func (s simpleLdap) Get(name string) (*iamv1alpha2.User, error) {
	if user, ok := s.store[name]; !ok {
		return nil, ErrUserNotExists
	} else {
		return user, nil
	}
}

func (s simpleLdap) Authenticate(name string, password string) error {
	if user, err := s.Get(name); err != nil {
		return err
	} else {
		if user.Spec.EncryptedPassword != password {
			return ErrInvalidCredentials
		}
	}

	return nil
}

func (l *simpleLdap) List(query *query.Query) (*api.ListResult, error) {
	items := make([]interface{}, 0)

	for _, user := range l.store {
		items = append(items, user)
	}

	return &api.ListResult{
		Items:      items,
		TotalItems: len(items),
	}, nil
}
