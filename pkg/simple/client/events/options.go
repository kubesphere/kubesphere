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

package events

import (
	"github.com/spf13/pflag"

	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Host        string `json:"host" yaml:"host"`
	BasicAuth   bool   `json:"basicAuth" yaml:"basicAuth"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	IndexPrefix string `json:"indexPrefix,omitempty" yaml:"indexPrefix,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

func NewEventsOptions() *Options {
	return &Options{
		Host:        "",
		IndexPrefix: "ks-logstash-events",
		Version:     "",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

func (s *Options) Validate() []error {
	errs := make([]error, 0)

	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Host, "events-elasticsearch-host", c.Host, ""+
		"Elasticsearch service host. KubeSphere is using elastic as event store, "+
		"if this filed left blank, KubeSphere will use kubernetes builtin event API instead, and"+
		" the following elastic search options will be ignored.")

	fs.BoolVar(&s.BasicAuth, "events-elasticsearch-basicAuth", c.BasicAuth, ""+
		"Elasticsearch events service basic auth enabled. KubeSphere is using elastic as events store, "+
		"if it is set to true, KubeSphere will connect to ElasticSearch using provided username and password by "+
		"events-elasticsearch-username and events-elasticsearch-username. Otherwise, KubeSphere will "+
		"anonymously access the Elasticsearch.")

	fs.StringVar(&s.Username, "events-elasticsearch-username", c.Username, ""+
		"ElasticSearch authentication username, only needed when events-elasticsearch-basicAuth is"+
		"set to true. ")

	fs.StringVar(&s.Password, "events-elasticsearch-password", c.Password, ""+
		"ElasticSearch authentication password, only needed when events-elasticsearch-basicAuth is"+
		"set to true. ")

	fs.StringVar(&s.IndexPrefix, "events-index-prefix", c.IndexPrefix, ""+
		"Index name prefix. KubeSphere will retrieve events against indices matching the prefix.")

	fs.StringVar(&s.Version, "events-elasticsearch-version", c.Version, ""+
		"Elasticsearch major version, e.g. 5/6/7, if left blank, will detect automatically."+
		"Currently, minimum supported version is 5.x")
}
