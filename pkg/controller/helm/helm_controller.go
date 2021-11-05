/*
Copyright 2021 The KubeSphere Authors.

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

package helm

import (
	"runtime"
	"time"

	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"

	"kubesphere.io/kubesphere/pkg/simple/client/gateway"

	"github.com/operator-framework/helm-operator-plugins/pkg/annotation"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler"
	"github.com/operator-framework/helm-operator-plugins/pkg/watches"
)

type Reconciler struct {
	GatewayOptions *gateway.Options
}

// SetupWithManager creates reconilers for each helm package that defined in the WatchFiles.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	var watchKinds []watches.Watch

	ws, err := watches.Load(r.GatewayOptions.WatchesPath)
	if err != nil {
		return err
	}
	watchKinds = append(watchKinds, ws...)

	for _, w := range watchKinds {
		// Register controller with the factory
		reconcilePeriod := time.Minute
		if w.ReconcilePeriod != nil {
			reconcilePeriod = w.ReconcilePeriod.Duration
		}

		maxConcurrentReconciles := runtime.NumCPU()
		if w.MaxConcurrentReconciles != nil {
			maxConcurrentReconciles = *w.MaxConcurrentReconciles
		}

		r, err := reconciler.New(
			reconciler.WithChart(*w.Chart),
			reconciler.WithGroupVersionKind(w.GroupVersionKind),
			reconciler.WithOverrideValues(r.defaultConfiguration()),
			reconciler.SkipDependentWatches(w.WatchDependentResources != nil && !*w.WatchDependentResources),
			reconciler.WithMaxConcurrentReconciles(maxConcurrentReconciles),
			reconciler.WithReconcilePeriod(reconcilePeriod),
			reconciler.WithInstallAnnotations(annotation.DefaultInstallAnnotations...),
			reconciler.WithUpgradeAnnotations(annotation.DefaultUpgradeAnnotations...),
			reconciler.WithUninstallAnnotations(annotation.DefaultUninstallAnnotations...),
		)
		if err != nil {
			return err
		}
		if err := r.SetupWithManager(mgr); err != nil {
			return err
		}
		klog.Info("configured watch", "gvk", w.GroupVersionKind, "chartPath", w.ChartPath, "maxConcurrentReconciles", maxConcurrentReconciles, "reconcilePeriod", reconcilePeriod)
	}
	return nil
}

func (r *Reconciler) defaultConfiguration() map[string]string {
	var overrideValues = make(map[string]string)
	if r.GatewayOptions.Repository != "" {
		overrideValues["controller.image.repository"] = r.GatewayOptions.Repository
	}
	if r.GatewayOptions.Tag != "" {
		overrideValues["controller.image.tag"] = r.GatewayOptions.Tag
	}
	return overrideValues
}
