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

package esclient

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type ElasticSearchOptions struct {
	Host        string `json:"host" yaml:"host"`
	IndexPrefix string `json:"indexPrefix,omitempty" yaml:"indexPrefix"`
	Version     string `json:"version" yaml:"version"`
}

func NewElasticSearchOptions() *ElasticSearchOptions {
	return &ElasticSearchOptions{
		Host:        "",
		IndexPrefix: "fluentbit",
		Version:     "",
	}
}

func (s *ElasticSearchOptions) ApplyTo(options *ElasticSearchOptions) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

func (s *ElasticSearchOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (s *ElasticSearchOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Host, "elasticsearch-host", s.Host, ""+
		"ElasticSearch logging service host. KubeSphere is using elastic as log store, "+
		"if this filed left blank, KubeSphere will use kubernetes builtin log API instead, and"+
		" the following elastic search options will be ignored.")

	fs.StringVar(&s.IndexPrefix, "index-prefix", s.IndexPrefix, ""+
		"Index name prefix. KubeSphere will retrieve logs against indices matching the prefix.")

	fs.StringVar(&s.Version, "elasticsearch-version", s.Version, ""+
		"ElasticSearch major version, e.g. 5/6/7, if left blank, will detect automatically."+
		"Currently, minimum supported version is 5.x")
}
