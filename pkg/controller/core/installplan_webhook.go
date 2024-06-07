/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode"

	"k8s.io/apimachinery/pkg/util/yaml"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.CustomValidator = &InstallPlanWebhook{}
var _ kscontroller.Controller = &InstallPlanWebhook{}

func (r *InstallPlanWebhook) Name() string {
	return "installplan-webhook"
}

type InstallPlanWebhook struct {
	client.Client
}

func trimSpace(data string) string {
	lines := strings.Split(data, "\n")
	var buf bytes.Buffer
	max := len(lines)
	for i, line := range lines {
		buf.Write([]byte(strings.TrimRightFunc(line, unicode.IsSpace)))
		if i < max-1 {
			buf.Write([]byte("\n"))
		}
	}
	return buf.String()
}

func (r *InstallPlanWebhook) Default(ctx context.Context, obj runtime.Object) error {
	installPlan := obj.(*corev1alpha1.InstallPlan)
	installPlan.Spec.Config = trimSpace(installPlan.Spec.Config)
	if installPlan.Spec.ClusterScheduling != nil {
		for k, v := range installPlan.Spec.ClusterScheduling.Overrides {
			installPlan.Spec.ClusterScheduling.Overrides[k] = trimSpace(v)
		}
	}
	return nil
}

func (r *InstallPlanWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateInstallPlan(ctx, obj.(*corev1alpha1.InstallPlan))
}

func (r *InstallPlanWebhook) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	return r.validateInstallPlan(ctx, newObj.(*corev1alpha1.InstallPlan))
}

func (r *InstallPlanWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *InstallPlanWebhook) validateInstallPlan(_ context.Context, installPlan *corev1alpha1.InstallPlan) (admission.Warnings, error) {
	var data interface{}

	if err := yaml.Unmarshal([]byte(installPlan.Spec.Config), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extension config: %v", err)
	}

	if installPlan.Spec.ClusterScheduling != nil {
		for cluster, config := range installPlan.Spec.ClusterScheduling.Overrides {
			if err := yaml.Unmarshal([]byte(config), &data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal cluster %s agent config: %v", cluster, err)
			}
		}
	}

	return nil, nil
}

func (r *InstallPlanWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		WithDefaulter(r).
		For(&corev1alpha1.InstallPlan{}).
		Complete()
}
