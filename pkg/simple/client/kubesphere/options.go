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

package kubesphere

import "github.com/spf13/pflag"

type KubeSphereOptions struct {
	APIServer     string `json:"apiServer" yaml:"apiServer"`
	AccountServer string `json:"accountServer" yaml:"accountServer"`
}

// NewKubeSphereOptions create a default options
func NewKubeSphereOptions() *KubeSphereOptions {
	return &KubeSphereOptions{
		APIServer:     "http://ks-apiserver.kubesphere-system.svc",
		AccountServer: "http://ks-account.kubesphere-system.svc",
	}
}

func (s *KubeSphereOptions) ApplyTo(options *KubeSphereOptions) {
	if s.AccountServer != "" {
		options.AccountServer = s.AccountServer
	}

	if s.APIServer != "" {
		options.APIServer = s.APIServer
	}
}

func (s *KubeSphereOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (s *KubeSphereOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.APIServer, "kubesphere-apiserver-host", s.APIServer, ""+
		"KubeSphere apiserver host address.")

	fs.StringVar(&s.AccountServer, "kubesphere-account-host", s.AccountServer, ""+
		"KubeSphere account server host address.")
}
