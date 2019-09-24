package ldap

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type LdapOptions struct {
	Host            string `json:"host,omitempty" yaml:"host"`
	ManagerDN       string `json:"managerDN,omitempty" yaml:"managerDN"`
	ManagerPassword string `json:"managerPassword,omitempty" yaml:"managerPassword"`
	UserSearchBase  string `json:"userSearchBase,omitempty" yaml:"userSearchBase"`
	GroupSearchBase string `json:"groupSearchBase,omitempty" yaml:"groupSearchBase"`
}

// NewLdapOptions return a default option
// which host field point to nowhere.
func NewLdapOptions() *LdapOptions {
	return &LdapOptions{
		Host:            "",
		ManagerDN:       "cn=admin,dc=example,dc=org",
		UserSearchBase:  "ou=Users,dc=example,dc=org",
		GroupSearchBase: "ou=Groups,dc=example,dc=org",
	}
}

func (l *LdapOptions) Validate() []error {
	errors := []error{}

	return errors
}

func (l *LdapOptions) ApplyTo(options *LdapOptions) {
	if l.Host != "" {
		reflectutils.Override(options, l)
	}
}

func (l *LdapOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&l.Host, "ldap-host", l.Host, ""+
		"Ldap service host, if left blank, all of the following ldap options will "+
		"be ignored and ldap will be disabled.")

	fs.StringVar(&l.ManagerDN, "ldap-manager-dn", l.ManagerDN, ""+
		"Ldap manager account domain name.")

	fs.StringVar(&l.ManagerPassword, "ldap-manager-password", l.ManagerPassword, ""+
		"Ldap manager account password.")

	fs.StringVar(&l.UserSearchBase, "ldap-user-search-base", l.UserSearchBase, ""+
		"Ldap user search base.")

	fs.StringVar(&l.GroupSearchBase, "ldap-group-search-base", l.GroupSearchBase, ""+
		"Ldap group search base.")
}
