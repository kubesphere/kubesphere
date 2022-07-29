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
	"reflect"
	"sort"

	"k8s.io/klog"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/extension/util"
)

const (
	PluginFinalizer = "plugin.extensions.kubesphere.io"
)

var _ reconcile.Reconciler = &PluginReconciler{}

type PluginReconciler struct {
	client.Client
	K8sVersion string
}

// reconcileDelete delete the plugin.
func (r *PluginReconciler) reconcileDelete(ctx context.Context, plugin *extensionsv1alpha1.Plugin) (ctrl.Result, error) {
	klog.V(4).Infof("remove the finalizer from plugin %s", plugin.Name)
	// Remove the finalizer from the plugin
	controllerutil.RemoveFinalizer(plugin, PluginFinalizer)
	if err := r.Update(ctx, plugin); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcilePluginStatus(ctx context.Context, c client.Client, plugin *extensionsv1alpha1.Plugin, k8sVersion string) (*extensionsv1alpha1.Plugin, error) {
	versionList := extensionsv1alpha1.PluginVersionList{}

	if err := c.List(ctx, &versionList, client.MatchingLabels{
		constants.ExtensionPluginLabel: plugin.Name,
	}); err != nil {
		return plugin, err
	}
	versionInfo := make([]extensionsv1alpha1.PluginVersionInfo, 0, len(versionList.Items))
	for i := range versionList.Items {
		if versionList.Items[i].DeletionTimestamp == nil {
			versionInfo = append(versionInfo, extensionsv1alpha1.PluginVersionInfo{
				Version:           versionList.Items[i].Spec.Version,
				CreationTimestamp: versionList.Items[i].CreationTimestamp,
			})
		}
	}
	sort.Sort(util.PluginVersionList(versionInfo))

	pluginCopy := plugin.DeepCopy()

	if recommended := util.GetRecommendedPluginVersion(versionList.Items, k8sVersion); recommended != nil {
		pluginCopy.Status.RecommendVersion = recommended.Spec.Version
	}
	pluginCopy.Status.Versions = versionInfo

	if !reflect.DeepEqual(pluginCopy, plugin) {
		if err := c.Update(ctx, pluginCopy); err != nil {
			return pluginCopy, err
		}
	}

	return plugin, nil
}

func (r *PluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync plugin: %s ", req.String())

	plugin := &extensionsv1alpha1.Plugin{}
	if err := r.Client.Get(ctx, req.NamespacedName, plugin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(plugin, PluginFinalizer) {
		patch := client.MergeFrom(plugin.DeepCopy())
		controllerutil.AddFinalizer(plugin, PluginFinalizer)
		if err := r.Patch(ctx, plugin, patch); err != nil {
			klog.Errorf("unable to register finalizer for plugin %s, error: %s", plugin.Name, err)
			return ctrl.Result{}, err
		}
	}

	if plugin.ObjectMeta.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, plugin)
	}

	if _, err := reconcilePluginStatus(ctx, r.Client, plugin, r.K8sVersion); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *PluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		Named("plugin-controller").
		For(&extensionsv1alpha1.Plugin{}).Complete(r)
}
