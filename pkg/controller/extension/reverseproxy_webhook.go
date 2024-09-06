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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

var _ admission.CustomValidator = &ReverseProxyWebhook{}
var _ kscontroller.Controller = &ReverseProxyWebhook{}

func (r *ReverseProxyWebhook) Name() string {
	return "reverseproxy-webhook"
}

type ReverseProxyWebhook struct {
	client.Client
}

func (r *ReverseProxyWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		For(&extensionsv1alpha1.ReverseProxy{}).
		Complete()
}

func (r *ReverseProxyWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateReverseProxy(ctx, obj.(*extensionsv1alpha1.ReverseProxy))
}

func (r *ReverseProxyWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return r.validateReverseProxy(ctx, newObj.(*extensionsv1alpha1.ReverseProxy))
}

func (r *ReverseProxyWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *ReverseProxyWebhook) validateReverseProxy(ctx context.Context, proxy *extensionsv1alpha1.ReverseProxy) (admission.Warnings, error) {
	reverseProxies := &extensionsv1alpha1.ReverseProxyList{}
	if err := r.Client.List(ctx, reverseProxies, &client.ListOptions{}); err != nil {
		return nil, err
	}
	for _, reverseProxy := range reverseProxies.Items {
		if reverseProxy.Name == proxy.Name {
			continue
		}
		if reverseProxy.Spec.Matcher.Method != proxy.Spec.Matcher.Method &&
			reverseProxy.Spec.Matcher.Method != "*" {
			continue
		}
		if reverseProxy.Spec.Matcher.Path == proxy.Spec.Matcher.Path {
			return nil, fmt.Errorf("ReverseProxy %v is already exists", proxy.Spec.Matcher)
		}
		if strings.HasSuffix(reverseProxy.Spec.Matcher.Path, "*") &&
			strings.HasPrefix(proxy.Spec.Matcher.Path, strings.TrimRight(reverseProxy.Spec.Matcher.Path, "*")) {
			return nil, fmt.Errorf("ReverseProxy %v is already exists", proxy.Spec.Matcher)
		}
	}
	return nil, nil
}
