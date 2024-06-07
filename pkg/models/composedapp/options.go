// Copyright 2024 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package composedapp

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/labels"
)

type Options struct {
	// KubeSphere is using sigs.k8s.io/application as fundamental object to implement Application Management.
	// There are other projects also built on sigs.k8s.io/application, when KubeSphere installed along side
	// them, conflicts happen. So we leave an option to only reconcile applications  matched with the given
	// selector. Default will reconcile all applications.
	//    For example
	//      "kubesphere.io/creator=" means reconcile applications with this label key
	//      "!kubesphere.io/creator" means exclude applications with this key
	AppSelector string `json:"appSelector,omitempty" yaml:"appSelector,omitempty" mapstructure:"appSelector,omitempty"`
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate() []error {
	var err []error
	if _, validateErr := labels.Parse(o.AppSelector); validateErr != nil {
		err = append(err, validateErr)
	}
	return err
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&o.AppSelector, "app-selector", s.AppSelector, "Selector to filter k8s applications to reconcile")
}
