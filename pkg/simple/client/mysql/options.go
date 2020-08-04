/*
Copyright 2020 KubeSphere Authors

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

package mysql

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"time"
)

type Options struct {
	Host                  string        `json:"host,omitempty" yaml:"host" description:"MySQL service host address"`
	Username              string        `json:"username,omitempty" yaml:"username"`
	Password              string        `json:"-" yaml:"password"`
	MaxIdleConnections    int           `json:"maxIdleConnections,omitempty" yaml:"maxIdleConnections"`
	MaxOpenConnections    int           `json:"maxOpenConnections,omitempty" yaml:"maxOpenConnections"`
	MaxConnectionLifeTime time.Duration `json:"maxConnectionLifeTime,omitempty" yaml:"maxConnectionLifeTime"`
}

// NewMySQLOptions create a `zero` value instance
func NewMySQLOptions() *Options {
	return &Options{
		Host:                  "",
		Username:              "",
		Password:              "",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Duration(10) * time.Second,
	}
}

func (m *Options) Validate() []error {
	var errors []error

	return errors
}

func (m *Options) ApplyTo(options *Options) {
	reflectutils.Override(options, m)
}

func (m *Options) AddFlags(fs *pflag.FlagSet, c *Options) {

	fs.StringVar(&m.Host, "mysql-host", c.Host, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")

	fs.StringVar(&m.Username, "mysql-username", c.Username, ""+
		"Username for access to mysql service.")

	fs.StringVar(&m.Password, "mysql-password", c.Password, ""+
		"Password for access to mysql, should be used pair with password.")

	fs.IntVar(&m.MaxIdleConnections, "mysql-max-idle-connections", c.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")

	fs.IntVar(&m.MaxOpenConnections, "mysql-max-open-connections", c.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")

	fs.DurationVar(&m.MaxConnectionLifeTime, "mysql-max-connection-life-time", c.MaxConnectionLifeTime, ""+
		"Maximum connection life time allowed to connecto to mysql.")
}
