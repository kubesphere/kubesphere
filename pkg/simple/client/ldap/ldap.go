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
	"github.com/golang/mock/gomock"
	"k8s.io/klog"
)

type Client interface {
	NewConn() (ldap.Client, error)
	Close()
	GroupSearchBase() string
	UserSearchBase() string
}

type poolClient struct {
	pool    Pool
	options *Options
}

// panic if cannot connect to ldap service
func NewClient(options *Options, stopCh <-chan struct{}) (Client, error) {
	pool, err := newChannelPool(options.InitialCap, options.MaxCap, options.PoolName, func(s string) (ldap.Client, error) {
		conn, err := ldap.Dial("tcp", options.Host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}, []uint16{ldap.LDAPResultAdminLimitExceeded, ldap.ErrorNetwork})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	client := &poolClient{
		pool:    pool,
		options: options,
	}

	go func() {
		<-stopCh
		client.Close()
	}()

	return client, nil
}
func (l *poolClient) Close() {
	if l.pool != nil {
		l.pool.Close()
	}
}

func (l *poolClient) NewConn() (ldap.Client, error) {
	if l.pool == nil {
		err := fmt.Errorf("ldap connection pool is not initialized")
		klog.Errorln(err)
		return nil, err
	}

	conn, err := l.pool.Get()
	// cannot connect to ldap server or pool is closed
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	err = conn.Bind(l.options.ManagerDN, l.options.ManagerPassword)
	if err != nil {
		conn.Close()
		klog.Errorln(err)
		return nil, err
	}
	return conn, nil
}

func (l *poolClient) GroupSearchBase() string {
	return l.options.GroupSearchBase
}

func (l *poolClient) UserSearchBase() string {
	return l.options.UserSearchBase
}

func NewMockClient(options *Options, ctrl *gomock.Controller, setup func(client *MockClient)) (Client, error) {
	pool, err := newChannelPool(options.InitialCap, options.MaxCap, options.PoolName, func(s string) (ldap.Client, error) {
		conn := newMockClient(ctrl)
		conn.EXPECT().Bind(gomock.Any(), gomock.Any()).AnyTimes()
		conn.EXPECT().Close().AnyTimes()
		setup(conn)
		return conn, nil
	}, []uint16{ldap.LDAPResultAdminLimitExceeded, ldap.ErrorNetwork})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	client := &poolClient{
		pool:    pool,
		options: options,
	}

	return client, nil
}
