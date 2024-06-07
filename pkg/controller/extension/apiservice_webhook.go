/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package extension

import (
	"context"
	"fmt"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

var _ admission.CustomValidator = &APIServiceWebhook{}
var _ kscontroller.Controller = &APIServiceWebhook{}

func (r *APIServiceWebhook) Name() string {
	return "apiservice-webhook"
}

type APIServiceWebhook struct {
	client.Client
}

func (r *APIServiceWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateAPIService(ctx, obj.(*extensionsv1alpha1.APIService))
}

func (r *APIServiceWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return r.validateAPIService(ctx, newObj.(*extensionsv1alpha1.APIService))
}

func (r *APIServiceWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *APIServiceWebhook) validateAPIService(ctx context.Context, service *extensionsv1alpha1.APIService) (admission.Warnings, error) {
	apiServices := &extensionsv1alpha1.APIServiceList{}
	if err := r.Client.List(ctx, apiServices, &client.ListOptions{}); err != nil {
		return nil, err
	}
	for _, apiService := range apiServices.Items {
		if apiService.Name != service.Name &&
			apiService.Spec.Group == service.Spec.Group &&
			apiService.Spec.Version == service.Spec.Version {
			return nil, fmt.Errorf("APIService %s/%s is already exists", service.Spec.Group, service.Spec.Version)
		}
	}
	return nil, nil
}

func (r *APIServiceWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		For(&extensionsv1alpha1.APIService{}).
		Complete()
}
