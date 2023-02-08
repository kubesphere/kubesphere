// Copyright 2022 The KubeSphere Authors.
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
package v2

import (
	corev1 "k8s.io/api/core/v1"
)

type RegistryHelper interface {
	// check if secret has correct credential to authenticate with remote registry
	Auth(secret *corev1.Secret) (bool, error)

	// fetch OCI Image Manifest, specification described as in https://github.com/opencontainers/image-spec/blob/main/manifest.md
	Config(secret *corev1.Secret, image string) (*ImageConfig, error)

	// list all tags of given repository, experimental
	ListRepositoryTags(secret *corev1.Secret, repository string) (RepositoryTags, error)
}

type registryHelper struct{}

func NewRegistryHelper() RegistryHelper {
	return &registryHelper{}
}

func (r *registryHelper) Auth(secret *corev1.Secret) (bool, error) {
	secretAuth, err := NewSecretAuthenticator(secret)
	if err != nil {
		return false, err
	}

	return secretAuth.Auth()
}

func (r *registryHelper) Config(secret *corev1.Secret, image string) (*ImageConfig, error) {
	secretAuth, err := NewSecretAuthenticator(secret)
	if err != nil {
		return nil, err
	}

	registryer := NewRegistryer(secretAuth.Options()...)
	config, err := registryer.Config(image)
	return &ImageConfig{ConfigFile: config}, err
}

func (r *registryHelper) ListRepositoryTags(secret *corev1.Secret, image string) (RepositoryTags, error) {
	secretAuth, err := NewSecretAuthenticator(secret)
	if err != nil {
		return RepositoryTags{}, err
	}

	registryer := NewRegistryer(secretAuth.Options()...)
	return registryer.ListRepositoryTags(image)
}
