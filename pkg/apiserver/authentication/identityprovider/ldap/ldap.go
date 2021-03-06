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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"time"
)

const (
	ldapIdentityProvider = "LDAPIdentityProvider"
	defaultReadTimeout   = 15000
)

func init() {
	identityprovider.RegisterGenericProvider(&ldapProviderFactory{})
}

type ldapProvider struct {
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
	// The following three fields are direct mappings of attributes on the user entry.
	// login attribute used for comparing user entries.
	LoginAttribute string `json:"loginAttribute" yaml:"loginAttribute"`
	MailAttribute  string `json:"mailAttribute" yaml:"mailAttribute"`
}

type ldapProviderFactory struct {
}

func (l *ldapProviderFactory) Type() string {
	return ldapIdentityProvider
}

func (l *ldapProviderFactory) Create(options oauth.DynamicOptions) (identityprovider.GenericProvider, error) {
	var ldapProvider ldapProvider
	if err := mapstructure.Decode(options, &ldapProvider); err != nil {
		return nil, err
	}
	if ldapProvider.ReadTimeout <= 0 {
		ldapProvider.ReadTimeout = defaultReadTimeout
	}
	return &ldapProvider, nil
}

type ldapIdentity struct {
	Username string
	Email    string
}

func (l *ldapIdentity) GetUserID() string {
	return l.Username
}

func (l *ldapIdentity) GetUsername() string {
	return l.Username
}

func (l *ldapIdentity) GetEmail() string {
	return l.Email
}

func (l ldapProvider) Authenticate(username string, password string) (identityprovider.Identity, error) {
	conn, err := l.newConn()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	conn.SetTimeout(time.Duration(l.ReadTimeout) * time.Millisecond)
	defer conn.Close()
	err = conn.Bind(l.ManagerDN, l.ManagerPassword)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	filter := fmt.Sprintf("(%s=%s)", l.LoginAttribute, ldap.EscapeFilter(username))
	if l.UserSearchFilter != "" {
		filter = fmt.Sprintf("(&%s%s)", filter, l.UserSearchFilter)
	}
	result, err := conn.Search(&ldap.SearchRequest{
		BaseDN:       l.UserSearchBase,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    1,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       filter,
		Attributes:   []string{l.LoginAttribute, l.MailAttribute},
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, errors.NewUnauthorized("incorrect password")
	}

	entry := result.Entries[0]
	if err = conn.Bind(entry.DN, password); err != nil {
		klog.Error(err)
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, errors.NewUnauthorized("incorrect password")
		}
		return nil, err
	}
	email := entry.GetAttributeValue(l.MailAttribute)
	return &ldapIdentity{
		Username: username,
		Email:    email,
	}, nil
}

func (l *ldapProvider) newConn() (*ldap.Conn, error) {
	if !l.StartTLS {
		return ldap.Dial("tcp", l.Host)
	}
	tlsConfig := tls.Config{}
	if l.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	tlsConfig.RootCAs = x509.NewCertPool()
	var caCert []byte
	var err error
	// Load CA cert
	if l.RootCA != "" {
		if caCert, err = ioutil.ReadFile(l.RootCA); err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	if l.RootCAData != "" {
		if caCert, err = base64.StdEncoding.DecodeString(l.RootCAData); err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	if caCert != nil {
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	}
	return ldap.DialTLS("tcp", l.Host, &tlsConfig)
}
