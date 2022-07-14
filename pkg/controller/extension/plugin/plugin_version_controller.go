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

package plugin

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	PluginVersionFinalizer = "plugin-version.extensions.kubesphere.io"
)

var _ reconcile.Reconciler = &PluginVersionReconciler{}

type PluginVersionReconciler struct {
	client.Client
	K8sVersion string
}

// reconcileDelete delete the plugin.
func (r *PluginVersionReconciler) reconcileDelete(ctx context.Context, pluginVersion *extensionsv1alpha1.PluginVersion) (ctrl.Result, error) {
	klog.V(4).Infof("remove the finalizer from plugin version %s", pluginVersion.Name)

	// Remove the finalizer from the plugin
	controllerutil.RemoveFinalizer(pluginVersion, PluginVersionFinalizer)
	if err := r.Update(ctx, pluginVersion); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PluginVersionReconciler) reconcile(ctx context.Context, pluginVersion *extensionsv1alpha1.PluginVersion) (ctrl.Result, error) {
	plugin := &extensionsv1alpha1.Plugin{}
	name := pluginVersion.Labels[constants.ExtensionPluginLabel]
	if err := r.Get(ctx, types.NamespacedName{Name: name}, plugin); err != nil {
		return ctrl.Result{}, err
	}

	if _, err := reconcilePluginStatus(ctx, r.Client, plugin, r.K8sVersion); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PluginVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync plugin version: %s ", req.String())

	pluginVersion := &extensionsv1alpha1.PluginVersion{}
	if err := r.Client.Get(ctx, req.NamespacedName, pluginVersion); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(pluginVersion, PluginVersionFinalizer) {
		patch := client.MergeFrom(pluginVersion.DeepCopy())
		controllerutil.AddFinalizer(pluginVersion, PluginVersionFinalizer)
		if err := r.Patch(ctx, pluginVersion, patch); err != nil {
			klog.Errorf("unable to register finalizer for plugin version %s, error: %s", pluginVersion.Name, err)
			return ctrl.Result{}, err
		}
	}

	if pluginVersion.ObjectMeta.DeletionTimestamp != nil {
		if result, err := r.reconcileDelete(ctx, pluginVersion); err != nil {
			return result, err
		}
	}

	if res, err := r.reconcile(ctx, pluginVersion); err != nil {
		return res, err
	}
	return ctrl.Result{}, nil
}

func (r *PluginVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		Named("plugin-version-controller").
		For(&extensionsv1alpha1.PluginVersion{}).Complete(r)
}
