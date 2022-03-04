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
package v2

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfig map[string]DockerConfigEntry

// DockerConfigEntry wraps a docker config as a entry
type DockerConfigEntry struct {
	Username string
	Password string
	Email    string
	Auth     string
}

// DockerConfigJSON represents ~/.docker/config.json file info
// see https://github.com/docker/docker/pull/12009
type DockerConfigJSON struct {
	Auths DockerConfig `json:"auths"`
	// +optional
	HTTPHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

type RepositoryTags struct {
	Registry   string   `json:"registry"`
	Repository string   `json:"repository"`
	Tags       []string `json:"tags"`
}

// ImageConfig wraps v1.ConfigFile to avoid direct dependency
type ImageConfig struct {
	*v1.ConfigFile `json:",inline"`
}
