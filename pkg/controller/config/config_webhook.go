/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package config

import (
	"context"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/config/identityprovider"
	"kubesphere.io/kubesphere/pkg/controller/config/oauthclient"
)

func (w *Webhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	secret := obj.(*v1.Secret)
	validator := w.factory.GetValidator(secret.Type)
	if validator != nil {
		return validator.ValidateCreate(ctx, secret)
	}
	return nil, nil
}

func (w *Webhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	newSecret := newObj.(*v1.Secret)
	oldSecret := oldObj.(*v1.Secret)
	if validator := w.factory.GetValidator(newSecret.Type); validator != nil {
		return validator.ValidateUpdate(ctx, oldSecret, newSecret)
	}
	return nil, nil
}

func (w *Webhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	secret := obj.(*v1.Secret)
	validator := w.factory.GetValidator(secret.Type)
	if validator != nil {
		return validator.ValidateDelete(ctx, secret)
	}
	return nil, nil
}

func (w *Webhook) Default(ctx context.Context, obj runtime.Object) error {
	secret := obj.(*v1.Secret)
	if secret.Namespace != constants.KubeSphereNamespace {
		return nil
	}

	defaulter := w.factory.GetDefaulter(secret.Type)
	if defaulter != nil {
		return defaulter.Default(ctx, secret)
	}

	return nil
}

var _ admission.CustomDefaulter = &Webhook{}
var _ admission.CustomValidator = &Webhook{}
var _ kscontroller.Controller = &Webhook{}

const webhookName = "kubesphere-config-webhook"

func (w *Webhook) Name() string {
	return webhookName
}

type Webhook struct {
	client.Client
	factory *WebhookFactory
}

func (w *Webhook) SetupWithManager(mgr *kscontroller.Manager) error {
	factory := NewWebhookFactory()
	oauthWebhookHandler := &oauthclient.WebhookHandler{Client: mgr.GetClient()}
	factory.RegisterValidator(oauthWebhookHandler)
	factory.RegisterDefaulter(oauthWebhookHandler)
	identityProviderWebhookHandler := &identityprovider.WebhookHandler{Client: mgr.GetClient()}
	factory.RegisterValidator(identityProviderWebhookHandler)
	factory.RegisterDefaulter(identityProviderWebhookHandler)

	w.Client = mgr.GetClient()
	w.factory = factory

	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(w).
		WithDefaulter(w).
		For(&v1.Secret{}).
		Complete()
}
