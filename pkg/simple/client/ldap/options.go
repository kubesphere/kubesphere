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
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Host            string `json:"host,omitempty" yaml:"host"`
	ManagerDN       string `json:"managerDN,omitempty" yaml:"managerDN"`
	ManagerPassword string `json:"managerPassword,omitempty" yaml:"managerPassword"`
	UserSearchBase  string `json:"userSearchBase,omitempty" yaml:"userSearchBase"`
	GroupSearchBase string `json:"groupSearchBase,omitempty" yaml:"groupSearchBase"`
	InitialCap      int    `json:"initialCap,omitempty" yaml:"initialCap"`
	MaxCap          int    `json:"maxCap,omitempty" yaml:"maxCap"`
	PoolName        string `json:"poolName,omitempty" yaml:"poolName"`
}

// NewOptions return a default option
// which host field point to nowhere.
func NewOptions() *Options {
	return &Options{
		Host:            "",
		ManagerDN:       "cn=admin,dc=example,dc=org",
		UserSearchBase:  "ou=Users,dc=example,dc=org",
		GroupSearchBase: "ou=Groups,dc=example,dc=org",
		InitialCap:      10,
		MaxCap:          100,
		PoolName:        "ldap",
	}
}

func (l *Options) Validate() []error {
	var errors []error

	return errors
}

func (l *Options) ApplyTo(options *Options) {
	if l.Host != "" {
		reflectutils.Override(options, l)
	}
}

func (l *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&l.Host, "ldap-host", s.Host, ""+
		"Ldap service host, if left blank, all of the following ldap options will "+
		"be ignored and ldap will be disabled.")

	fs.StringVar(&l.ManagerDN, "ldap-manager-dn", s.ManagerDN, ""+
		"Ldap manager account domain name.")

	fs.StringVar(&l.ManagerPassword, "ldap-manager-password", s.ManagerPassword, ""+
		"Ldap manager account password.")

	fs.StringVar(&l.UserSearchBase, "ldap-user-search-base", s.UserSearchBase, ""+
		"Ldap user search base.")

	fs.StringVar(&l.GroupSearchBase, "ldap-group-search-base", s.GroupSearchBase, ""+
		"Ldap group search base.")
}
