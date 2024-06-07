/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package ldap

import (
	"os"
	"testing"

	"kubesphere.io/kubesphere/pkg/server/options"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestNewLdapProvider(t *testing.T) {
	opts := `
host: test.sn.mynetname.net:389
managerDN: uid=root,cn=users,dc=test,dc=sn,dc=mynetname,dc=net
managerPassword: test
startTLS: false
userSearchBase: dc=test,dc=sn,dc=mynetname,dc=net
loginAttribute: uid
mailAttribute: mail
`
	var dynamicOptions options.DynamicOptions
	err := yaml.Unmarshal([]byte(opts), &dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	got, err := new(ldapProviderFactory).Create(dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	expected := &ldapProvider{
		Host:                 "test.sn.mynetname.net:389",
		StartTLS:             false,
		InsecureSkipVerify:   false,
		ReadTimeout:          15000,
		RootCA:               "",
		RootCAData:           "",
		ManagerDN:            "uid=root,cn=users,dc=test,dc=sn,dc=mynetname,dc=net",
		ManagerPassword:      "test",
		UserSearchBase:       "dc=test,dc=sn,dc=mynetname,dc=net",
		UserSearchFilter:     "",
		GroupSearchBase:      "",
		GroupSearchFilter:    "",
		UserMemberAttribute:  "",
		GroupMemberAttribute: "",
		LoginAttribute:       "uid",
		MailAttribute:        "mail",
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("%T differ (-got, +want): %s", expected, diff)
	}
}

func TestLdapProvider_Authenticate(t *testing.T) {
	configFile := os.Getenv("LDAP_TEST_FILE")
	if configFile == "" {
		t.Skip("Skipped")
	}
	opts, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	var dynamicOptions options.DynamicOptions
	if err = yaml.Unmarshal(opts, &dynamicOptions); err != nil {
		t.Fatal(err)
	}
	ldapProvider, err := new(ldapProviderFactory).Create(dynamicOptions)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ldapProvider.Authenticate("test", "test"); err != nil {
		t.Fatal(err)
	}
}
