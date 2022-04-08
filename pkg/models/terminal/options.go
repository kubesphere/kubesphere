// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package terminal

import "github.com/spf13/pflag"

type Options struct {
	Image   string `json:"image,omitempty" yaml:"image,omitempty"`
	Timeout int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func NewTerminalOptions() *Options {
	return &Options{
		Image:   "alpine:3.15",
		Timeout: 600,
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {

}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {

}
