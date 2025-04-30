/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package imagesearch

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HostDockerIo = "https://docker.io"
)

var (
	searchProviderFactories = make(map[string]SearchProviderFactory)
)

type SearchProvider interface {
	Search(imageName string, config SearchConfig) (*Results, error)
}

type Results struct {
	Total   int64    `json:"total"`
	Entries []string `json:"entries"`
}

type SearchConfig struct {
	Host         string
	ProviderType string
	Username     string
	Password     string
}

type SearchProviderFactory interface {
	Type() string
	Create(options map[string]interface{}) (SearchProvider, error)
}

func RegistrySearchProvider(factory SearchProviderFactory) {
	searchProviderFactories[factory.Type()] = factory
}

type SecretGetter interface {
	GetSecretConfig(ctx context.Context, name, namespace string) (*SearchConfig, error)
}

func NewSecretGetter(reader client.Reader) SecretGetter {
	return &secretGetter{reader}
}

type secretGetter struct {
	client.Reader
}

// {"auths":{"https://harbor.172.31.19.17.nip.io":{"username":"admin","password":"Harbor12345","email":"","auth":"YWRtaW46SGFyYm9yMTIzNDU="}}}

func (s *secretGetter) GetSecretConfig(ctx context.Context, name, namespace string) (*SearchConfig, error) {
	secret := &corev1.Secret{}
	err := s.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret)
	if err != nil {
		return nil, err
	}
	provider := secret.Annotations[SecretTypeImageSearchProvider]

	data := secret.Data[".dockerconfigjson"]
	auths := jsoniter.Get(data, "auths")

	host := auths.Keys()[0]
	username := auths.Get(host, "username").ToString()
	password := auths.Get(host, "password").ToString()
	return &SearchConfig{
		Host:         host,
		ProviderType: provider,
		Username:     username,
		Password:     password,
	}, nil
}
