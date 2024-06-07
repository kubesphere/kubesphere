/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

import (
	"context"
	"errors"
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

var once sync.Once

type WebhookHandler struct {
	client.Client
	getter identityprovider.ConfigurationGetter
}

func (w *WebhookHandler) Default(_ context.Context, secret *corev1.Secret) error {
	configuration, err := identityprovider.UnmarshalFrom(secret)
	if err != nil {
		return err
	}
	if configuration.Name != "" {
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		secret.Labels[identityprovider.SecretTypeIdentityProvider] = configuration.Name
	}
	return nil
}

func (w *WebhookHandler) ValidateCreate(ctx context.Context, secret *corev1.Secret) (admission.Warnings, error) {
	idp, err := identityprovider.UnmarshalFrom(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity provider: %v", err)
	}

	if idp.Name == "" {
		return nil, errors.New("invalid Identity Provider, please ensure that the provider name is not empty")
	}

	exists, err := w.isClientExist(ctx, idp.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check identity provider: %v", err)
	}

	if exists {
		return nil, fmt.Errorf("invalid provider, provider name '%s' already exists", idp.Name)
	}

	return nil, nil
}

func (w *WebhookHandler) ValidateUpdate(_ context.Context, old, new *corev1.Secret) (admission.Warnings, error) {
	oldIdp, err := identityprovider.UnmarshalFrom(old)
	if err != nil {
		return nil, err
	}

	newIdp, err := identityprovider.UnmarshalFrom(new)
	if err != nil {
		return nil, err
	}

	if newIdp.Name != oldIdp.Name {
		return nil, fmt.Errorf("the provider name is immutable, old: %s, new: %s", oldIdp.Name, newIdp.Name)
	}
	return nil, nil
}

func (w *WebhookHandler) ValidateDelete(_ context.Context, _ *corev1.Secret) (admission.Warnings, error) {
	return nil, nil
}

func (w *WebhookHandler) ConfigType() corev1.SecretType {
	return identityprovider.SecretTypeIdentityProvider
}

func (w *WebhookHandler) isClientExist(ctx context.Context, clientName string) (bool, error) {
	once.Do(func() {
		w.getter = identityprovider.NewConfigurationGetter(w.Client)
	})

	_, err := w.getter.GetConfiguration(ctx, clientName)
	if err != nil {
		if errors.Is(identityprovider.ErrorIdentityProviderNotFound, err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get identity provider: %v", err)
	}
	return true, nil
}
