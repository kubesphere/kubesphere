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

package identityprovider

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

const LdapIdentityProvider = "LDAPIdentityProvider"

type LdapProvider interface {
	Authenticate(username string, password string) (*iamv1alpha2.User, error)
}

type ldapOptions struct {
	Host            string `json:"host" yaml:"host"`
	ManagerDN       string `json:"managerDN"  yaml:"managerDN"`
	ManagerPassword string `json:"-"  yaml:"managerPassword"`
	UserSearchBase  string `json:"userSearchBase"  yaml:"userSearchBase"`
	//This is typically uid
	LoginAttribute       string `json:"loginAttribute" yaml:"loginAttribute"`
	MailAttribute        string `json:"mailAttribute" yaml:"mailAttribute"`
	DisplayNameAttribute string `json:"displayNameAttribute" yaml:"displayNameAttribute"`
}

type ldapProvider struct {
	options ldapOptions
}

func NewLdapProvider(options *oauth.DynamicOptions) (LdapProvider, error) {
	data, err := yaml.Marshal(options)
	if err != nil {
		return nil, err
	}
	var ldapOptions ldapOptions
	err = yaml.Unmarshal(data, &ldapOptions)
	if err != nil {
		return nil, err
	}
	return &ldapProvider{options: ldapOptions}, nil
}

func (l ldapProvider) Authenticate(username string, password string) (*iamv1alpha2.User, error) {
	conn, err := ldap.Dial("tcp", l.options.Host)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer conn.Close()
	err = conn.Bind(l.options.ManagerDN, l.options.ManagerPassword)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	filter := fmt.Sprintf("(&(%s=%s))", l.options.LoginAttribute, username)

	result, err := conn.Search(&ldap.SearchRequest{
		BaseDN:       l.options.UserSearchBase,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    1,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       filter,
		Attributes:   []string{l.options.LoginAttribute, l.options.MailAttribute, l.options.DisplayNameAttribute},
	})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(result.Entries) == 1 {
		entry := result.Entries[0]
		err = conn.Bind(entry.DN, password)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		return &iamv1alpha2.User{
			ObjectMeta: metav1.ObjectMeta{
				Name: username,
			},
			Spec: iamv1alpha2.UserSpec{
				Email:       entry.GetAttributeValue(l.options.MailAttribute),
				DisplayName: entry.GetAttributeValue(l.options.DisplayNameAttribute),
			},
		}, nil
	}

	return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf(" could not find user %s in LDAP directory", username))
}
