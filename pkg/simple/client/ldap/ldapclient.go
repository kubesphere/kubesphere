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
	"flag"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"log"
	"sync"
)

var (
	once            sync.Once
	pool            Pool
	ldapHost        string
	ManagerDN       string
	ManagerPassword string
	UserSearchBase  string
	GroupSearchBase string
	poolSize        int
)

func init() {
	flag.StringVar(&ldapHost, "ldap-server", "localhost:389", "ldap server host")
	flag.StringVar(&ManagerDN, "ldap-manager-dn", "cn=admin,dc=example,dc=org", "ldap manager dn")
	flag.StringVar(&ManagerPassword, "ldap-manager-password", "admin", "ldap manager password")
	flag.StringVar(&UserSearchBase, "ldap-user-search-base", "ou=Users,dc=example,dc=org", "ldap user search base")
	flag.StringVar(&GroupSearchBase, "ldap-group-search-base", "ou=Groups,dc=example,dc=org", "ldap group search base")
	flag.IntVar(&poolSize, "ldap-pool-size", 64, "ldap connection pool size")
}

func ldapClientPool() Pool {

	once.Do(func() {
		var err error
		pool, err = NewChannelPool(8, poolSize, "kubesphere", func(s string) (ldap.Client, error) {
			conn, err := ldap.Dial("tcp", ldapHost)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}, []uint16{ldap.LDAPResultTimeLimitExceeded, ldap.ErrorNetwork})

		if err != nil {
			log.Fatalln(err)
		}
	})
	return pool
}

func Client() (ldap.Client, error) {
	conn, err := ldapClientPool().Get()

	if err != nil {
		glog.Errorln("get ldap connection from pool", err)
		return nil, err
	}

	err = conn.Bind(ManagerDN, ManagerPassword)

	if err != nil {
		conn.Close()
		glog.Errorln("bind manager dn", err)
		return nil, err
	}

	return conn, nil
}
