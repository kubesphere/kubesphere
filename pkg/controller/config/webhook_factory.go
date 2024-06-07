/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package config

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	v1 "k8s.io/api/core/v1"
)

type ValidatorInterface interface {
	ValidateCreate(ctx context.Context, secret *v1.Secret) (admission.Warnings, error)
	ValidateUpdate(ctx context.Context, old, new *v1.Secret) (admission.Warnings, error)
	ValidateDelete(ctx context.Context, secret *v1.Secret) (admission.Warnings, error)
	ConfigType() v1.SecretType
}

type DefaulterInterface interface {
	Default(ctx context.Context, secret *v1.Secret) error
	ConfigType() v1.SecretType
}

func (w *WebhookFactory) RegisterValidator(validator ValidatorInterface) {
	w.validators[validator.ConfigType()] = validator
}

func (w *WebhookFactory) RegisterDefaulter(defaulter DefaulterInterface) {
	w.defaulters[defaulter.ConfigType()] = defaulter
}

func (w *WebhookFactory) GetValidator(secretType v1.SecretType) ValidatorInterface {
	return w.validators[secretType]
}

func (w *WebhookFactory) GetDefaulter(secretType v1.SecretType) DefaulterInterface {
	return w.defaulters[secretType]
}

type WebhookFactory struct {
	validators map[v1.SecretType]ValidatorInterface
	defaulters map[v1.SecretType]DefaulterInterface
}

func NewWebhookFactory() *WebhookFactory {
	return &WebhookFactory{
		validators: make(map[v1.SecretType]ValidatorInterface),
		defaulters: make(map[v1.SecretType]DefaulterInterface),
	}
}
