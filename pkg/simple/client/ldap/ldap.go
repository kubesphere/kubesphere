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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"sort"
	"strings"
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
	defer conn.Close()

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
	defer conn.Close()

	err = conn.Bind(l.managerDN, l.managerPassword)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *ldapInterfaceImpl) dnForUsername(username string) string {
	return fmt.Sprintf("uid=%s,%s", username, l.userSearchBase)
}

func (l *ldapInterfaceImpl) filterForUsername(username string) string {
	return fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(%s=%s)(%s=%s)))", ldapAttributeUserID, username, ldapAttributeMail, username)
}

func (l *ldapInterfaceImpl) Get(name string) (*iamv1alpha2.User, error) {
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
			ldapAttributeCreateTimestamp,
		},
	}

	searchResults, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(searchResults.Entries) == 0 {
		return nil, ErrUserNotExists
	}

	userEntry := searchResults.Entries[0]

	user := &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userEntry.GetAttributeValue(ldapAttributeUserID),
		},
		Spec: iamv1alpha2.UserSpec{
			Email:       userEntry.GetAttributeValue(ldapAttributeMail),
			Description: userEntry.GetAttributeValue(ldapAttributeDescription),
		},
	}

	createTimestamp, _ := time.Parse(ldapAttributeCreateTimestampLayout, userEntry.GetAttributeValue(ldapAttributeCreateTimestamp))
	user.ObjectMeta.CreationTimestamp.Time = createTimestamp
	return user, nil
}

func (l *ldapInterfaceImpl) Create(user *iamv1alpha2.User) error {
	createRequest := &ldap.AddRequest{
		DN: l.dnForUsername(user.Name),
		Attributes: []ldap.Attribute{
			{
				Type: ldapAttributeObjectClass,
				Vals: []string{"inetOrgPerson", "top"},
			},
			{
				Type: ldapAttributeCommonName,
				Vals: []string{user.Name},
			},
			{
				Type: ldapAttributeSerialNumber,
				Vals: []string{user.Name},
			},
			{
				Type: ldapAttributeUserPassword,
				Vals: []string{user.Spec.EncryptedPassword},
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

func (l *ldapInterfaceImpl) Update(newUser *iamv1alpha2.User) error {
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

	if newUser.Spec.EncryptedPassword != "" {
		modifyRequest.Replace(ldapAttributeUserPassword, []string{newUser.Spec.EncryptedPassword})
	}

	return conn.Modify(modifyRequest)

}

func (l *ldapInterfaceImpl) Authenticate(username, password string) error {
	conn, err := l.newConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	dn := l.dnForUsername(username)
	err = conn.Bind(dn, password)
	if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		return ErrInvalidCredentials
	}
	return err
}

func (l *ldapInterfaceImpl) List(query *query.Query) (*api.ListResult, error) {
	conn, err := l.newConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	pageControl := ldap.NewControlPaging(1000)

	users := make([]iamv1alpha2.User, 0)

	filter := "(&(objectClass=inetOrgPerson))"

	for {
		userSearchRequest := ldap.NewSearchRequest(
			l.userSearchBase,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter,
			[]string{ldapAttributeUserID, ldapAttributeMail, ldapAttributeDescription, ldapAttributeCreateTimestamp},
			[]ldap.Control{pageControl},
		)

		response, err := conn.Search(userSearchRequest)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		for _, entry := range response.Entries {

			uid := entry.GetAttributeValue(ldapAttributeUserID)
			email := entry.GetAttributeValue(ldapAttributeMail)
			description := entry.GetAttributeValue(ldapAttributeDescription)
			createTimestamp, _ := time.Parse(ldapAttributeCreateTimestampLayout, entry.GetAttributeValue(ldapAttributeCreateTimestamp))

			user := iamv1alpha2.User{
				ObjectMeta: metav1.ObjectMeta{
					Name:              uid,
					CreationTimestamp: metav1.Time{Time: createTimestamp}},
				Spec: iamv1alpha2.UserSpec{
					Email:       email,
					Description: description,
				}}

			users = append(users, user)
		}

		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pageControl.SetCookie(ctrl.Cookie)
			continue
		}

		break
	}

	sort.Slice(users, func(i, j int) bool {
		if !query.Ascending {
			i, j = j, i
		}
		switch query.SortBy {
		case "username":
			return strings.Compare(users[i].Name, users[j].Name) <= 0
		case "createTime":
			fallthrough
		default:
			return users[i].CreationTimestamp.Before(&users[j].CreationTimestamp)
		}
	})

	items := make([]interface{}, 0)

	for i, user := range users {
		if i >= query.Pagination.Offset && len(items) < query.Pagination.Limit {
			items = append(items, user)
		}
	}

	return &api.ListResult{
		Items:      items,
		TotalItems: len(users),
	}, nil
}
