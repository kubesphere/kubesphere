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

package sonarqube

import (
	"github.com/spf13/pflag"
)

type SonarQubeOptions struct {
	Host  string `json:",omitempty" yaml:"host" description:"SonarQube service host address"`
	Token string `json:",omitempty" yaml:"token" description:"SonarQube service token"`
}

func NewSonarQubeOptions() *SonarQubeOptions {
	return &SonarQubeOptions{
		Host:  "",
		Token: "",
	}
}

func NewDefaultSonarQubeOptions() *SonarQubeOptions {
	return NewSonarQubeOptions()
}

func (s *SonarQubeOptions) Validate() []error {
	errors := []error{}

	return errors
}

func (s *SonarQubeOptions) ApplyTo(options *SonarQubeOptions) {
	if s.Host != "" {
		options.Host = s.Host
		options.Token = s.Token
	}
}

func (s *SonarQubeOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Host, "sonarqube-host", s.Host, ""+
		"Sonarqube service address, if left empty, following sonarqube options will be ignored.")

	fs.StringVar(&s.Token, "sonarqube-token", s.Token, ""+
		"Sonarqube service access token.")
}
