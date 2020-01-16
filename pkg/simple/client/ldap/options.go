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
