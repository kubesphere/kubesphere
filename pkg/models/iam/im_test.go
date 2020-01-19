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

package iam

import (
	"github.com/golang/mock/gomock"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"testing"
)

func TestIMOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ldappool, err := ldap.NewMockClient(&ldap.Options{
		Host:            "192.168.0.7:30389",
		ManagerDN:       "cn=admin,dc=kubesphere,dc=io",
		ManagerPassword: "admin",
		UserSearchBase:  "ou=Users,dc=kubesphere,dc=io",
		GroupSearchBase: "ou=Groups,dc=kubesphere,dc=io",
		InitialCap:      8,
		MaxCap:          64,
	}, ctrl, func(client *ldap.MockClient) {
		client.EXPECT().Search(gomock.Any()).AnyTimes()
	})

	if err != nil {
		t.Fatal(err)
	}

	defer ldappool.Close()

	im := NewIMOperator(ldappool, Config{})

	err = im.Init()

	if err != nil {
		t.Fatal(err)
	}
}
