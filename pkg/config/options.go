/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package config

type ExperimentalOptions struct {
	ValidationDirective string `json:"validationDirective" yaml:"validationDirective" mapstructure:"validationDirective"`
}

func NewExperimentalOptions() *ExperimentalOptions {
	return &ExperimentalOptions{
		ValidationDirective: "",
	}
}
