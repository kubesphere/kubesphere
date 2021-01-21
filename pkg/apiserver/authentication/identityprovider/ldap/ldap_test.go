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

package ldap

import (
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"os"
	"testing"
)

func TestNewLdapProvider(t *testing.T) {
	options := `
host: test.sn.mynetname.net:389
managerDN: uid=root,cn=users,dc=test,dc=sn,dc=mynetname,dc=net
managerPassword: test
startTLS: false
userSearchBase: dc=test,dc=sn,dc=mynetname,dc=net
loginAttribute: uid
mailAttribute: mail
`
	var dynamicOptions oauth.DynamicOptions
	err := yaml.Unmarshal([]byte(options), &dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	got, err := new(ldapProviderFactory).Create(dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	expected := &ldapProvider{
		Host:                 "test.sn.mynetname.net:389",
		StartTLS:             false,
		InsecureSkipVerify:   false,
		ReadTimeout:          15000,
		RootCA:               "",
		RootCAData:           "",
		ManagerDN:            "uid=root,cn=users,dc=test,dc=sn,dc=mynetname,dc=net",
		ManagerPassword:      "test",
		UserSearchBase:       "dc=test,dc=sn,dc=mynetname,dc=net",
		UserSearchFilter:     "",
		GroupSearchBase:      "",
		GroupSearchFilter:    "",
		UserMemberAttribute:  "",
		GroupMemberAttribute: "",
		LoginAttribute:       "uid",
		MailAttribute:        "mail",
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("%T differ (-got, +want): %s", expected, diff)
	}
}

func TestLdapProvider_Authenticate(t *testing.T) {
	configFile := os.Getenv("LDAP_TEST_FILE")
	if configFile == "" {
		t.Skip("Skipped")
	}
	options, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	var dynamicOptions oauth.DynamicOptions
	if err = yaml.Unmarshal(options, &dynamicOptions); err != nil {
		t.Fatal(err)
	}
	ldapProvider, err := new(ldapProviderFactory).Create(dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ldapProvider.Authenticate("test", "test"); err != nil {
		t.Fatal(err)
	}
}
