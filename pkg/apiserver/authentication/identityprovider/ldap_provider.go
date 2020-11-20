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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/constants"
	"time"
)

const (
	LdapIdentityProvider = "LDAPIdentityProvider"
	defaultReadTimeout   = 15000
)

type LdapProvider interface {
	Authenticate(username string, password string) (*iamv1alpha2.User, error)
}

type ldapOptions struct {
	// Host and optional port of the LDAP server in the form "host:port".
	// If the port is not supplied, 389 for insecure or StartTLS connections, 636
	Host string `json:"host,omitempty" yaml:"managerDN"`
	// Timeout duration when reading data from remote server. Default to 15s.
	ReadTimeout int `json:"readTimeout" yaml:"readTimeout"`
	// If specified, connections will use the ldaps:// protocol
	StartTLS bool `json:"startTLS,omitempty" yaml:"startTLS"`
	// Used to turn off TLS certificate checks
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	// Path to a trusted root certificate file. Default: use the host's root CA.
	RootCA string `json:"rootCA,omitempty" yaml:"rootCA"`
	// A raw certificate file can also be provided inline. Base64 encoded PEM file
	RootCAData string `json:"rootCAData,omitempty" yaml:"rootCAData"`
	// Username (DN) of the "manager" user identity.
	ManagerDN string `json:"managerDN,omitempty" yaml:"managerDN"`
	// The password for the manager DN.
	ManagerPassword string `json:"-,omitempty" yaml:"managerPassword"`
	// User search scope.
	UserSearchBase string `json:"userSearchBase,omitempty" yaml:"userSearchBase"`
	// LDAP filter used to identify objects of type user. e.g. (objectClass=person)
	UserSearchFilter string `json:"userSearchFilter,omitempty" yaml:"userSearchFilter"`
	// Group search scope.
	GroupSearchBase string `json:"groupSearchBase,omitempty" yaml:"groupSearchBase"`
	// LDAP filter used to identify objects of type group. e.g. (objectclass=group)
	GroupSearchFilter string `json:"groupSearchFilter,omitempty" yaml:"groupSearchFilter"`
	// Attribute on a user object storing the groups the user is a member of.
	UserMemberAttribute string `json:"userMemberAttribute,omitempty" yaml:"userMemberAttribute"`
	// Attribute on a group object storing the information for primary group membership.
	GroupMemberAttribute string `json:"groupMemberAttribute,omitempty" yaml:"groupMemberAttribute"`
	// login attribute used for comparing user entries.
	// The following three fields are direct mappings of attributes on the user entry.
	LoginAttribute       string `json:"loginAttribute" yaml:"loginAttribute"`
	MailAttribute        string `json:"mailAttribute" yaml:"mailAttribute"`
	DisplayNameAttribute string `json:"displayNameAttribute" yaml:"displayNameAttribute"`
}

type ldapProvider struct {
	options ldapOptions
}

func NewLdapProvider(options *oauth.DynamicOptions) (LdapProvider, error) {
	var ldapOptions ldapOptions
	if err := mapstructure.Decode(options, &ldapOptions); err != nil {
		return nil, err
	}
	if ldapOptions.ReadTimeout <= 0 {
		ldapOptions.ReadTimeout = defaultReadTimeout
	}
	return &ldapProvider{options: ldapOptions}, nil
}

func (l ldapProvider) Authenticate(username string, password string) (*iamv1alpha2.User, error) {
	conn, err := l.newConn()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	conn.SetTimeout(time.Duration(l.options.ReadTimeout) * time.Millisecond)
	defer conn.Close()
	err = conn.Bind(l.options.ManagerDN, l.options.ManagerPassword)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	filter := fmt.Sprintf("(&(%s=%s)%s)", l.options.LoginAttribute, username, l.options.UserSearchFilter)
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
		email := entry.GetAttributeValue(l.options.MailAttribute)
		displayName := entry.GetAttributeValue(l.options.DisplayNameAttribute)
		return &iamv1alpha2.User{
			ObjectMeta: metav1.ObjectMeta{
				Name: username,
				Annotations: map[string]string{
					constants.DisplayNameAnnotationKey: displayName,
				},
			},
			Spec: iamv1alpha2.UserSpec{
				Email:       email,
				DisplayName: displayName,
			},
		}, nil
	}

	return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("could not find user %s in LDAP directory", username))
}

func (l *ldapProvider) newConn() (*ldap.Conn, error) {
	if !l.options.StartTLS {
		return ldap.Dial("tcp", l.options.Host)
	}
	tlsConfig := tls.Config{}
	if l.options.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	tlsConfig.RootCAs = x509.NewCertPool()
	var caCert []byte
	var err error
	// Load CA cert
	if l.options.RootCA != "" {
		if caCert, err = ioutil.ReadFile(l.options.RootCA); err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	if l.options.RootCAData != "" {
		if caCert, err = base64.StdEncoding.DecodeString(l.options.RootCAData); err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	if caCert != nil {
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	}
	return ldap.DialTLS("tcp", l.options.Host, &tlsConfig)
}
