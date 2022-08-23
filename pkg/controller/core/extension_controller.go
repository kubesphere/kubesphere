/*
Copyright 2022 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"context"
	"reflect"
	"sort"

	"k8s.io/klog"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ExtensionFinalizer = "extensions.kubesphere.io"
)

var _ reconcile.Reconciler = &ExtensionReconciler{}

type ExtensionReconciler struct {
	client.Client
	K8sVersion string
}

// reconcileDelete delete the extension.
func (r *ExtensionReconciler) reconcileDelete(ctx context.Context, extension *corev1alpha1.Extension) (ctrl.Result, error) {
	klog.V(4).Infof("remove the finalizer from extension %s", extension.Name)
	// Remove the finalizer from the extension
	controllerutil.RemoveFinalizer(extension, ExtensionFinalizer)
	if err := r.Update(ctx, extension); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileExtensionStatus(ctx context.Context, c client.Client, extension *corev1alpha1.Extension, k8sVersion string) (*corev1alpha1.Extension, error) {
	versionList := corev1alpha1.ExtensionVersionList{}

	if err := c.List(ctx, &versionList, client.MatchingLabels{
		corev1alpha1.ExtensionLabel: extension.Name,
	}); err != nil {
		return extension, err
	}
	versionInfo := make([]corev1alpha1.ExtensionVersionInfo, 0, len(versionList.Items))
	for i := range versionList.Items {
		if versionList.Items[i].DeletionTimestamp == nil {
			versionInfo = append(versionInfo, corev1alpha1.ExtensionVersionInfo{
				Version:           versionList.Items[i].Spec.Version,
				CreationTimestamp: versionList.Items[i].CreationTimestamp,
			})
		}
	}
	sort.Sort(VersionList(versionInfo))

	extensionCopy := extension.DeepCopy()

	if recommended := getRecommendedExtensionVersion(versionList.Items, k8sVersion); recommended != nil {
		extensionCopy.Status.RecommendVersion = recommended.Spec.Version
	}
	extensionCopy.Status.Versions = versionInfo

	if !reflect.DeepEqual(extensionCopy, extension) {
		if err := c.Update(ctx, extensionCopy); err != nil {
			return extensionCopy, err
		}
	}

	return extension, nil
}

func (r *ExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync extension: %s ", req.String())

	extension := &corev1alpha1.Extension{}
	if err := r.Client.Get(ctx, req.NamespacedName, extension); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(extension, ExtensionFinalizer) {
		patch := client.MergeFrom(extension.DeepCopy())
		controllerutil.AddFinalizer(extension, ExtensionFinalizer)
		if err := r.Patch(ctx, extension, patch); err != nil {
			klog.Errorf("unable to register finalizer for extension %s, error: %s", extension.Name, err)
			return ctrl.Result{}, err
		}
	}

	if extension.ObjectMeta.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, extension)
	}

	if _, err := reconcileExtensionStatus(ctx, r.Client, extension, r.K8sVersion); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ExtensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		Named("extension-controller").
		For(&corev1alpha1.Extension{}).Complete(r)
}
