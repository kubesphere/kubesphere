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
	"k8s.io/klog"
)

type LdapClient struct {
	pool    Pool
	options *LdapOptions
}

// panic if cannot connect to ldap service
func NewLdapClient(options *LdapOptions, stopCh <-chan struct{}) (*LdapClient, error) {
	pool, err := NewChannelPool(8, 64, "kubesphere", func(s string) (ldap.Client, error) {
		conn, err := ldap.Dial("tcp", options.Host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}, []uint16{ldap.LDAPResultAdminLimitExceeded, ldap.ErrorNetwork})

	if err != nil {
		klog.Error(err)
		pool.Close()
		return nil, err
	}

	client := &LdapClient{
		pool:    pool,
		options: options,
	}

	go func() {
		<-stopCh
		if client.pool != nil {
			client.pool.Close()
		}
	}()

	return client, nil
}

func (l *LdapClient) NewConn() (ldap.Client, error) {
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
		klog.Error(err)
		return nil, err
	}
	return conn, nil
}

func (l *LdapClient) GroupSearchBase() string {
	return l.options.GroupSearchBase
}

func (l *LdapClient) UserSearchBase() string {
	return l.options.UserSearchBase
}
