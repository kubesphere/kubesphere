/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package imagesearch

import (
	"fmt"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

const (
	SecretDataKey = "configuration.yaml"
)

type Configuration struct {
	// The provider name.
	Name string `json:"name" yaml:"name"`

	// The type of image search provider
	Type string `json:"type" yaml:"type"`

	// The options of image search provider
	ProviderOptions map[string]interface{} `json:"provider" yaml:"provider"`
}

func UnmarshalFrom(secret *v1.Secret) (*Configuration, error) {
	config := &Configuration{}
	if err := yaml.Unmarshal(secret.Data[SecretDataKey], config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret data: %s", err)
	}
	return config, nil
}
