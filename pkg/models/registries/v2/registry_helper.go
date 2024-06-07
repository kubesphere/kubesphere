/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

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
