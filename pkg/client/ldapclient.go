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
package client

import (
	"flag"
	"fmt"
	"github.com/go-ldap/ldap"
	ldapPool "kubesphere.io/kubesphere/pkg/client/ldap"
	"os"
	"sync"
)

var (
	once            sync.Once
	pool            ldapPool.Pool
	ldapHost        string
	ManagerDN       string
	ManagerPassword string
	UserSearchBase  string
	GroupSearchBase string
)

func init() {
	flag.StringVar(&ldapHost, "ldap-server", "localhost:389", "ldap server host")
	flag.StringVar(&ManagerDN, "ldap-manager-dn", "cn=admin,dc=example,dc=org", "ldap manager dn")
	flag.StringVar(&ManagerPassword, "ldap-manager-password", "admin", "ldap manager password")
	flag.StringVar(&UserSearchBase, "ldap-user-search-base", "ou=Users,dc=example,dc=org", "ldap user search base")
	flag.StringVar(&GroupSearchBase, "ldap-group-search-base", "ou=Groups,dc=example,dc=org", "ldap group search base")
}

func LdapClient() ldapPool.Pool {

	once.Do(func() {
		var err error
		pool, err = ldapPool.NewChannelPool(8, 96, "kubesphere", func(s string) (ldap.Client, error) {
			conn, err := ldap.Dial("tcp", ldapHost)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}, []uint16{ldap.LDAPResultTimeLimitExceeded, ldap.ErrorNetwork})

		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			panic(err)
		}
	})
	return pool
}
