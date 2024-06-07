/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package extension

import (
	"context"
	"fmt"
	"strings"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

var _ admission.CustomValidator = &JSBundleWebhook{}
var _ admission.CustomDefaulter = &JSBundleWebhook{}
var _ kscontroller.Controller = &JSBundleWebhook{}

type JSBundleWebhook struct {
	client.Client
}

func (r *JSBundleWebhook) Name() string {
	return "jsbundle-webhook"
}

func (r *JSBundleWebhook) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *JSBundleWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		WithDefaulter(r).
		For(&extensionsv1alpha1.JSBundle{}).
		Complete()
}

var _ admission.CustomDefaulter = &JSBundleWebhook{}

func (r *JSBundleWebhook) Default(_ context.Context, obj runtime.Object) error {
	jsBundle := obj.(*extensionsv1alpha1.JSBundle)
	extensionName := jsBundle.Labels[v1alpha1.ExtensionReferenceLabel]
	if jsBundle.Status.Link == "" && extensionName != "" {
		jsBundle.Status.Link = fmt.Sprintf("/dist/%s/index.js", extensionName)
	}
	return nil
}

func (r *JSBundleWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateJSBundle(ctx, obj.(*extensionsv1alpha1.JSBundle))
}

func (r *JSBundleWebhook) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	return r.validateJSBundle(ctx, newObj.(*extensionsv1alpha1.JSBundle))
}

func (r *JSBundleWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *JSBundleWebhook) validateJSBundle(ctx context.Context, jsBundle *extensionsv1alpha1.JSBundle) (admission.Warnings, error) {
	if jsBundle.Status.Link == "" {
		return nil, nil
	}
	extensionName := jsBundle.Labels[v1alpha1.ExtensionReferenceLabel]
	if extensionName != "" && !strings.HasPrefix(jsBundle.Status.Link, fmt.Sprintf("/dist/%s", extensionName)) {
		return nil, fmt.Errorf("the prefix of status.link must be in the format /dist/%s/", extensionName)
	}
	jsBundles := &extensionsv1alpha1.JSBundleList{}
	if err := r.Client.List(ctx, jsBundles, &client.ListOptions{}); err != nil {
		return nil, err
	}
	for _, item := range jsBundles.Items {
		if item.Name != jsBundle.Name &&
			item.Status.Link == jsBundle.Status.Link {
			return nil, fmt.Errorf("JSBundle %s is already exists", jsBundle.Status.Link)
		}
	}
	return nil, nil
}
