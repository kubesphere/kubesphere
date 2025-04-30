/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

import (
	"context"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/options"
)

const (
	MappingMethodManual MappingMethod = "manual"

	MappingMethodAuto MappingMethod = "auto"

	// MappingMethodLookup Looks up an existing identity, user identity mapping, and user, but does not automatically
	// provision users or identities. Using this method requires you to manually provision users.
	MappingMethodLookup MappingMethod = "lookup"

	ConfigTypeIdentityProvider = "identityprovider"
	SecretTypeIdentityProvider = "config.kubesphere.io/" + ConfigTypeIdentityProvider

	SecretDataKey = "configuration.yaml"
)

var ErrorIdentityProviderNotFound = errors.New("the Identity provider was not found")

var IdentityProviderIsDisabled = errors.New("the Identity provider was Disabled")

type MappingMethod string

type Configuration struct {
	// The provider name.
	Name string `json:"name" yaml:"name"`

	// Defines how new identities are mapped to users when they login. Allowed values are:
	//  - manual: The user needs to confirm the mapped username on the onboarding page.
	//  - auto: Skip the onboarding screen, so the user cannot change its username.
	//            Fails if a user with that username is already mapped to another identity.
	//  - lookup: Looks up an existing identity, user identity mapping, and user, but does not automatically
	//            provision users or identities. Using this method requires you to manually provision users.
	MappingMethod MappingMethod `json:"mappingMethod" yaml:"mappingMethod"`

	// The type of identity provider
	Type string `json:"type" yaml:"type"`

	// The hidden of login button
	Hidden bool `json:"hidden" yaml:"hidden"`

	// The disabled of identify provider
	Disabled bool `json:"disabled" yaml:"disabled"`

	// The display name of identify provider
	DisplayName string `json:"displayName" yaml:"displayName"`

	// The options of identify provider
	ProviderOptions options.DynamicOptions `json:"provider" yaml:"provider"`
}

type ConfigurationGetter interface {
	GetConfiguration(ctx context.Context, name string) (*Configuration, error)
	ListConfigurations(ctx context.Context) ([]*Configuration, error)
}

func NewConfigurationGetter(client client.Client) ConfigurationGetter {
	return &configurationGetter{client}
}

type configurationGetter struct {
	client.Client
}

func (o *configurationGetter) ListConfigurations(ctx context.Context) ([]*Configuration, error) {
	configurations := make([]*Configuration, 0)
	secrets := &v1.SecretList{}
	if err := o.List(ctx, secrets, client.InNamespace(constants.KubeSphereNamespace),
		client.MatchingLabels{constants.GenericConfigTypeLabel: ConfigTypeIdentityProvider}); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %s", err)
	}
	for _, secret := range secrets.Items {
		if secret.Type != SecretTypeIdentityProvider {
			continue
		}
		if config, err := UnmarshalFrom(&secret); err != nil {
			klog.Errorf("failed to unmarshal secret data: %s", err)
			continue
		} else {
			configurations = append(configurations, config)
		}
	}
	return configurations, nil
}

func (o *configurationGetter) GetConfiguration(ctx context.Context, name string) (*Configuration, error) {
	configurations, err := o.ListConfigurations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %s", err)
	}
	for _, config := range configurations {
		if config.Name == name {
			if config.Disabled {
				return nil, IdentityProviderIsDisabled
			}
			return config, nil
		}
	}
	return nil, ErrorIdentityProviderNotFound
}

func UnmarshalFrom(secret *v1.Secret) (*Configuration, error) {
	config := &Configuration{}
	if err := yaml.Unmarshal(secret.Data[SecretDataKey], config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret data: %s", err)
	}
	return config, nil
}

func IsIdentityProviderConfiguration(secret *v1.Secret) bool {
	if secret.Namespace != constants.KubeSphereNamespace {
		return false
	}
	return secret.Type == SecretTypeIdentityProvider
}
