/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	appv2 "kubesphere.io/api/application/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.CustomValidator = &ReleaseWebhook{}
var _ kscontroller.ClusterSelector = &ReleaseWebhook{}
var _ kscontroller.Controller = &ReleaseWebhook{}

type ReleaseWebhook struct {
	cache.Cache
}

func (a *ReleaseWebhook) Name() string {
	return "applicationrelease-webhook"
}

func (a *ReleaseWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	a.Cache = mgr.GetCache()
	return ctrl.NewWebhookManagedBy(mgr).WithValidator(a).For(&appv2.ApplicationRelease{}).Complete()

}

func (a *ReleaseWebhook) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (a *ReleaseWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return a.validateAppVersionState(ctx, obj.(*appv2.ApplicationRelease))
}

func (a *ReleaseWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	return a.validateAppVersionState(ctx, newObj.(*appv2.ApplicationRelease))
}

func (a *ReleaseWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (a *ReleaseWebhook) validateAppVersionState(ctx context.Context, release *appv2.ApplicationRelease) (warnings admission.Warnings, err error) {
	versionID := release.Spec.AppVersionID
	appVersion := &appv2.ApplicationVersion{}
	err = a.Get(ctx, types.NamespacedName{Name: versionID}, appVersion)
	if err != nil {
		return nil, err
	}
	if appVersion.Status.State != appv2.ReviewStatusActive && release.Status.State != appv2.ReviewStatusPassed {

		return nil, fmt.Errorf("invalid application version: %s, state: %s, for release: %s",
			versionID, appVersion.Status.State, release.Name)
	}
	return nil, nil
}
