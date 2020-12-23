/*
Copyright 2019 The KubeSphere Authors.

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

package helmapplication

import (
	"context"
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

var _ reconcile.Reconciler = &ReconcileHelmApplication{}

// ReconcileHelmApplication reconciles a federated helm application object
type ReconcileHelmApplication struct {
	client.Client
	MultiClusterEnabled bool
}

func (r *ReconcileHelmApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.Infof("sync helm application: %s", request.String())
	defer func() {
		klog.Infof("sync helm application end: %s, elapsed: %v", request.String(), time.Now().Sub(start))
	}()

	instance := &v1beta1.FederatedHelmApplication{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp == nil &&
		!strings.HasSuffix(instance.Name, constants.HelmApplicationAppStoreSuffix) &&
		instance.Spec.Template.Spec.Status == constants.StateActive {
		return reconcile.Result{}, r.createHelmApplicationCopyInAppStore(instance)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHelmApplication) createHelmApplicationCopyInAppStore(from *v1beta1.FederatedHelmApplication) error {
	name := fmt.Sprintf("%s%s", from.Name, constants.HelmApplicationAppStoreSuffix)

	app := &v1beta1.FederatedHelmApplication{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: name}, app)
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	if app.Name == "" {
		app.Name = name
		labels := from.Labels
		if len(labels) == 0 {
			labels = make(map[string]string, 3)
		}
		labels[constants.ChartRepoIdLabelKey] = constants.AppStoreRepoId
		if labels[constants.CategoryIdLabelKey] == "" {
			labels[constants.CategoryIdLabelKey] = constants.UncategorizedId
		}
		delete(labels, constants.WorkspaceLabelKey)
		app.Labels = labels

		app.Spec = *from.Spec.DeepCopy()
		err = r.Client.Create(context.TODO(), app)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileHelmApplication) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	if !r.MultiClusterEnabled {
		_, err := mgr.GetRESTMapper().ResourceSingularizer(v1beta1.FederatedHelmApplicationVersionKind)
		if err != nil {
			klog.Info("federated helm application not exists, exit the controller")
			return nil
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.FederatedHelmApplication{}).Complete(r)
}

