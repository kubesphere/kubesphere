/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"context"
	"reflect"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	extensionVersionController = "extensionVersion"
)

var _ kscontroller.Controller = &ExtensionVersionReconciler{}
var _ reconcile.Reconciler = &ExtensionVersionReconciler{}

func (r *ExtensionVersionReconciler) Name() string {
	return extensionVersionController
}

func (r *ExtensionVersionReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type ExtensionVersionReconciler struct {
	client.Client
	logger logr.Logger
}

func (r *ExtensionVersionReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(extensionVersionController)
	return ctrl.NewControllerManagedBy(mgr).
		Named(extensionVersionController).
		For(&corev1alpha1.ExtensionVersion{}).
		Complete(r)
}

func (r *ExtensionVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("extensionVersion", req.String())
	logger.V(4).Info("reconciling extension version")
	ctx = klog.NewContext(ctx, logger)

	extensionVersion := &corev1alpha1.ExtensionVersion{}
	if err := r.Client.Get(ctx, req.NamespacedName, extensionVersion); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !extensionVersion.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if err := r.syncExtensionVersion(ctx, extensionVersion); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to sync extension version")
	}

	logger.V(4).Info("extension version successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *ExtensionVersionReconciler) syncExtensionVersion(ctx context.Context, extensionVersion *corev1alpha1.ExtensionVersion) error {
	logger := klog.FromContext(ctx)

	// automatically synchronized by the repository controller
	if extensionVersion.Spec.Repository != "" {
		return nil
	}

	if extensionVersion.Spec.Version != "" {
		return nil
	}

	extensionVersionSpec, err := fetchExtensionVersionSpec(ctx, r.Client, extensionVersion)
	if err != nil {
		return errors.Wrap(err, "failed to fetch extension version spec")
	}

	if len(isValidExtensionName(extensionVersionSpec.Name)) > 0 {
		logger.V(4).Info("invalid extension name found", "name", extensionVersionSpec.Name)
		return nil
	}

	expected := extensionVersion.DeepCopy()
	if expected.Labels == nil {
		expected.Labels = map[string]string{}
	}
	expected.Labels[corev1alpha1.ExtensionReferenceLabel] = extensionVersionSpec.Name
	expected.Labels[corev1alpha1.CategoryLabel] = extensionVersionSpec.Category
	expected.Spec = extensionVersionSpec

	if !reflect.DeepEqual(expected.Spec, extensionVersion.Spec) ||
		!reflect.DeepEqual(expected.Labels, extensionVersion.Labels) {
		if err := r.Update(ctx, expected); err != nil {
			return errors.Wrap(err, "failed to update extension version")
		}
		logger.V(4).Info("extension version updated")
	}

	extension := &corev1alpha1.Extension{
		ObjectMeta: metav1.ObjectMeta{Name: extensionVersionSpec.Name},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, extension, func() error {
		if !needUpdate(extensionVersionSpec.Version, extension.Status.Versions) {
			return nil
		}
		if extension.Labels == nil {
			extension.Labels = make(map[string]string)
		}
		if extensionVersion.Spec.Category != "" {
			extension.Labels[corev1alpha1.CategoryLabel] = extensionVersion.Spec.Category
		}
		extension.Spec.ExtensionInfo = expected.Spec.ExtensionInfo
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "failed to update extension: %v", err)
	}

	logger.V(4).Info("extension successfully updated", "operation", op, "name", extension.Name)
	return nil
}

func needUpdate(version string, versions []corev1alpha1.ExtensionVersionInfo) bool {
	v1, _ := semver.NewVersion(version)
	for _, v := range versions {
		v2, _ := semver.NewVersion(v.Version)
		if v2.GreaterThan(v1) {
			return false
		}
	}
	return true
}
