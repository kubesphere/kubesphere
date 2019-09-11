package mysql

import (
	"github.com/spf13/pflag"
	reflectutils "kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"time"
)

type MySQLOptions struct {
	Host                  string        `json:"host,omitempty" yaml:"host,omitempty" description:"MySQL service host address"`
	Username              string        `json:"username,omitempty" yaml:"username,omitempty"`
	Password              string        `json:"-" yaml:"password,omitempty"`
	MaxIdleConnections    int           `json:"maxIdleConnections,omitempty" yaml:"maxIdleConnections,omitempty"`
	MaxOpenConnections    int           `json:"maxOpenConnections,omitempty" yaml:"maxOpenConnections,omitempty"`
	MaxConnectionLifeTime time.Duration `json:"maxConnectionLifeTime,omitempty" yaml:"maxConnectionLifeTime,omitempty"`
}

// NewMySQLOptions create a `zero` value instance
func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{}
}

func (m *MySQLOptions) Validate() []error {
	errors := []error{}

	return errors
}

func (m *MySQLOptions) ApplyTo(options *MySQLOptions) {
	reflectutils.Override(options, m)
}

func (m *MySQLOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&m.Host, "mysql-host", m.Host, ""+
		"MySQL service host address. If left blank, following options will be ignored.")

	fs.StringVar(&m.Username, "mysql-username", m.Username, ""+
		"Username for access to mysql service.")

	fs.StringVar(&m.Password, "mysql-password", m.Password, ""+
		"Password for access to mysql, should be used pair with password.")
}
