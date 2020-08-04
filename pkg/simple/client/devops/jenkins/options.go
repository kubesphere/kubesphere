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

package jenkins

import (
	"fmt"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Host           string `json:",omitempty" yaml:"host" description:"Jenkins service host address"`
	Username       string `json:",omitempty" yaml:"username" description:"Jenkins admin username"`
	Password       string `json:",omitempty" yaml:"password" description:"Jenkins admin password"`
	MaxConnections int    `json:"maxConnections,omitempty" yaml:"maxConnections" description:"Maximum connections allowed to connect to Jenkins"`
}

// NewDevopsOptions returns a `zero` instance
func NewDevopsOptions() *Options {
	return &Options{
		Host:           "",
		Username:       "",
		Password:       "",
		MaxConnections: 100,
	}
}

// ApplyTo apply configuration to another options
func (s *Options) ApplyTo(options *Options) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

// Validate check if there is misconfiguration in options
func (s *Options) Validate() []error {
	var errors []error

	// devops is not needed, ignore rest options
	if s.Host == "" {
		return errors
	}

	if s.Username == "" || s.Password == "" {
		errors = append(errors, fmt.Errorf("jenkins's username or password is empty"))
	}

	if s.MaxConnections <= 0 {
		errors = append(errors, fmt.Errorf("jenkins's maximum connections should be greater than 0"))
	}

	return errors
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Host, "jenkins-host", c.Host, ""+
		"Jenkins service host address. If left blank, means Jenkins "+
		"is unnecessary.")

	fs.StringVar(&s.Username, "jenkins-username", c.Username, ""+
		"Username for access to Jenkins service. Leave it blank if there isn't any.")

	fs.StringVar(&s.Password, "jenkins-password", c.Password, ""+
		"Password for access to Jenkins service, used pair with username.")

	fs.IntVar(&s.MaxConnections, "jenkins-max-connections", c.MaxConnections, ""+
		"Maximum allowed connections to Jenkins. ")

}
