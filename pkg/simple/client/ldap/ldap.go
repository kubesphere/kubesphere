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
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/google/uuid"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"sync"
	"time"
)

const (
	ldapAttributeObjectClass       = "objectClass"
	ldapAttributeCommonName        = "cn"
	ldapAttributeSerialNumber      = "sn"
	ldapAttributeGlobalIDNumber    = "gidNumber"
	ldapAttributeHomeDirectory     = "homeDirectory"
	ldapAttributeUserID            = "uid"
	ldapAttributeUserIDNumber      = "uidNumber"
	ldapAttributeMail              = "mail"
	ldapAttributeUserPassword      = "userPassword"
	ldapAttributePreferredLanguage = "preferredLanguage"
	ldapAttributeDescription       = "description"
	ldapAttributeCreateTimestamp   = "createTimestamp"
	ldapAttributeOrganizationUnit  = "ou"

	// ldap create timestamp attribute layout
	ldapAttributeCreateTimestampLayout = "20060102150405Z"
)

var ErrUserAlreadyExisted = errors.New("user already existed")
var ErrUserNotExists = errors.New("user not exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type ldapInterfaceImpl struct {
	pool            Pool
	userSearchBase  string
	groupSearchBase string
	managerDN       string
	managerPassword string

	once sync.Once
}

var _ Interface = &ldapInterfaceImpl{}

func NewLdapClient(options *Options, stopCh <-chan struct{}) (Interface, error) {

	poolFactory := func(s string) (ldap.Client, error) {
		conn, err := ldap.Dial("tcp", options.Host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	pool, err := newChannelPool(options.InitialCap,
		options.MaxCap,
		options.PoolName,
		poolFactory,
		[]uint16{ldap.LDAPResultAdminLimitExceeded, ldap.ErrorNetwork})

	if err != nil {
		return nil, err
	}

	client := &ldapInterfaceImpl{
		pool:            pool,
		userSearchBase:  options.UserSearchBase,
		groupSearchBase: options.GroupSearchBase,
		managerDN:       options.ManagerDN,
		managerPassword: options.ManagerPassword,
		once:            sync.Once{},
	}

	go func() {
		<-stopCh
		client.close()
	}()

	client.once.Do(func() {
		_ = client.createSearchBase()
	})

	return client, nil
}

func (l *ldapInterfaceImpl) createSearchBase() error {
	conn, err := l.newConn()
	if err != nil {
		return err
	}

	createIfNotExistsFunc := func(request *ldap.AddRequest) error {
		searchRequest := &ldap.SearchRequest{
			BaseDN:       request.DN,
			Scope:        ldap.ScopeWholeSubtree,
			DerefAliases: ldap.NeverDerefAliases,
			SizeLimit:    0,
			TimeLimit:    0,
			TypesOnly:    false,
			Filter:       "(objectClass=*)",
		}

		_, err = conn.Search(searchRequest)
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			return conn.Add(request)
		}
		return nil
	}

	userSearchBaseAddRequest := &ldap.AddRequest{
		DN: l.userSearchBase,
		Attributes: []ldap.Attribute{
			{
				Type: ldapAttributeObjectClass,
				Vals: []string{"organizationalUnit", "top"},
			},
			{
				Type: ldapAttributeOrganizationUnit,
				Vals: []string{"Users"},
			},
		},
	}

	err = createIfNotExistsFunc(userSearchBaseAddRequest)
	if err != nil {
		return err
	}

	groupSearchBaseAddRequest := *userSearchBaseAddRequest
	groupSearchBaseAddRequest.DN = l.groupSearchBase

	return createIfNotExistsFunc(&groupSearchBaseAddRequest)

}

func (l *ldapInterfaceImpl) close() {
	if l.pool != nil {
		l.pool.Close()
	}
}

func (l *ldapInterfaceImpl) newConn() (ldap.Client, error) {
	conn, err := l.pool.Get()
	if err != nil {
		return nil, err
	}

	err = conn.Bind(l.managerDN, l.managerPassword)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (l *ldapInterfaceImpl) dnForUsername(username string) string {
	return fmt.Sprintf("uid=%s,%s", username, l.userSearchBase)
}

func (l *ldapInterfaceImpl) filterForUsername(username string) string {
	return fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", username, username)
}

func (l *ldapInterfaceImpl) Get(name string) (*iam.User, error) {
	conn, err := l.newConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchRequest := &ldap.SearchRequest{
		BaseDN:       l.userSearchBase,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       l.filterForUsername(name),
		Attributes: []string{
			ldapAttributeMail,
			ldapAttributeDescription,
			ldapAttributePreferredLanguage,
			ldapAttributeCreateTimestamp,
		},
	}

	searchResults, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(searchResults.Entries) != 1 {
		return nil, ErrUserNotExists
	}

	userEntry := searchResults.Entries[0]

	user := &iam.User{
		Name:        userEntry.GetAttributeValue(ldapAttributeUserID),
		Email:       userEntry.GetAttributeValue(ldapAttributeMail),
		Lang:        userEntry.GetAttributeValue(ldapAttributePreferredLanguage),
		Description: userEntry.GetAttributeValue(ldapAttributeDescription),
	}

	createTimestamp, _ := time.Parse(ldapAttributeCreateTimestampLayout, userEntry.GetAttributeValue(ldapAttributeCreateTimestamp))
	user.CreateTime = createTimestamp

	return user, nil
}

func (l *ldapInterfaceImpl) Create(user *iam.User) error {
	if _, err := l.Get(user.Name); err != nil {
		return ErrUserAlreadyExisted
	}

	createRequest := &ldap.AddRequest{
		DN: l.dnForUsername(user.Name),
		Attributes: []ldap.Attribute{
			{
				Type: ldapAttributeObjectClass,
				Vals: []string{"inetOrgPerson", "posixAccount", "top"},
			},
			{
				Type: ldapAttributeCommonName,
				Vals: []string{user.Name},
			},
			{
				Type: ldapAttributeSerialNumber,
				Vals: []string{" "},
			},
			{
				Type: ldapAttributeGlobalIDNumber,
				Vals: []string{"500"},
			},
			{
				Type: ldapAttributeHomeDirectory,
				Vals: []string{"/home/" + user.Name},
			},
			{
				Type: ldapAttributeUserID,
				Vals: []string{user.Name},
			},
			{
				Type: ldapAttributeUserIDNumber,
				Vals: []string{uuid.New().String()},
			},
			{
				Type: ldapAttributeMail,
				Vals: []string{user.Email},
			},
			{
				Type: ldapAttributeUserPassword,
				Vals: []string{user.Password},
			},
			{
				Type: ldapAttributePreferredLanguage,
				Vals: []string{user.Lang},
			},
			{
				Type: ldapAttributeDescription,
				Vals: []string{user.Description},
			},
		},
	}

	conn, err := l.newConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Add(createRequest)
}

func (l *ldapInterfaceImpl) Delete(name string) error {
	conn, err := l.newConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = l.Get(name)
	if err != nil {
		if err == ErrUserNotExists {
			return nil
		}
		return err
	}

	deleteRequest := &ldap.DelRequest{
		DN: l.dnForUsername(name),
	}

	return conn.Del(deleteRequest)
}

func (l *ldapInterfaceImpl) Update(newUser *iam.User) error {
	conn, err := l.newConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	// check user existed
	_, err = l.Get(newUser.Name)
	if err != nil {
		return err
	}

	modifyRequest := &ldap.ModifyRequest{
		DN: l.dnForUsername(newUser.Name),
	}

	if newUser.Description != "" {
		modifyRequest.Replace(ldapAttributeDescription, []string{newUser.Description})
	}

	if newUser.Lang != "" {
		modifyRequest.Replace(ldapAttributePreferredLanguage, []string{newUser.Lang})
	}

	if newUser.Password != "" {
		modifyRequest.Replace(ldapAttributeUserPassword, []string{newUser.Password})
	}

	return conn.Modify(modifyRequest)

}

func (l *ldapInterfaceImpl) Verify(username, password string) error {
	conn, err := l.newConn()
	if err != nil {
		return err
	}

	dn := l.dnForUsername(username)
	err = conn.Bind(dn, password)
	if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		return ErrInvalidCredentials
	}
	return err
}
