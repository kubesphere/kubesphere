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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"strings"
)

func init() {
	registerMetrics()
}

const (
	helmApplicationControllerName = "helm-application-controller"
)

var _ reconcile.Reconciler = &ReconcileHelmApplication{}

// ReconcileHelmApplication reconciles a federated helm application object
type ReconcileHelmApplication struct {
	client.Client
}

const (
	appFinalizer = "helmapplication.application.kubesphere.io"
)

func (r *ReconcileHelmApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	klog.V(4).Infof("sync helm application: %s ", request.String())

	rootCtx := context.Background()
	app := &v1alpha1.HelmApplication{}
	err := r.Client.Get(rootCtx, request.NamespacedName, app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if app.DeletionTimestamp == nil {
		// new app, update finalizer
		if !sliceutil.HasString(app.ObjectMeta.Finalizers, appFinalizer) {
			app.ObjectMeta.Finalizers = append(app.ObjectMeta.Finalizers, appFinalizer)
			if err := r.Update(rootCtx, app); err != nil {
				return reconcile.Result{}, err
			}
			// create app success
			appOperationTotal.WithLabelValues("creation", app.GetTrueName(), strconv.FormatBool(inAppStore(app))).Inc()
		}

		if !inAppStore(app) {
			if app.Status.State == v1alpha1.StateActive ||
				app.Status.State == v1alpha1.StateSuspended {
				if err := r.createAppCopyInAppStore(rootCtx, app); err != nil {
					klog.Errorf("create app copy failed, error: %s", err)
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			}
		}
	} else {
		// delete app copy in appStore
		if !inAppStore(app) {
			if err := r.deleteAppCopyInAppStore(rootCtx, app.Name); err != nil {
				return reconcile.Result{}, err
			}
		}

		app.ObjectMeta.Finalizers = sliceutil.RemoveString(app.ObjectMeta.Finalizers, func(item string) bool {
			return item == appFinalizer
		})
		klog.V(4).Info("update app")
		if err := r.Update(rootCtx, app); err != nil {
			klog.Errorf("update app failed, error: %s", err)
			return ctrl.Result{}, err
		} else {
			// delete app success
			appOperationTotal.WithLabelValues("deletion", app.GetTrueName(), strconv.FormatBool(inAppStore(app))).Inc()
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHelmApplication) deleteAppCopyInAppStore(ctx context.Context, name string) error {
	appInStore := &v1alpha1.HelmApplication{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s%s", name, v1alpha1.HelmApplicationAppStoreSuffix)}, appInStore)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		err = r.Delete(ctx, appInStore)
		return err
	}

	return nil
}

// createAppCopyInAppStore create a application copy in app store
func (r *ReconcileHelmApplication) createAppCopyInAppStore(ctx context.Context, originApp *v1alpha1.HelmApplication) error {
	name := fmt.Sprintf("%s%s", originApp.Name, v1alpha1.HelmApplicationAppStoreSuffix)

	app := &v1alpha1.HelmApplication{}
	err := r.Get(ctx, types.NamespacedName{Name: name}, app)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if app.Name == "" {
		app.Name = name
		labels := originApp.Labels
		if len(labels) == 0 {
			labels = make(map[string]string, 3)
		}
		labels[constants.ChartRepoIdLabelKey] = v1alpha1.AppStoreRepoId

		// assign a default category to app
		if labels[constants.CategoryIdLabelKey] == "" {
			labels[constants.CategoryIdLabelKey] = v1alpha1.UncategorizedId
		}
		// record the original workspace
		labels[v1alpha1.OriginWorkspaceLabelKey] = originApp.GetWorkspace()
		// apps in store are global resource.
		delete(labels, constants.WorkspaceLabelKey)

		app.Annotations = originApp.Annotations
		app.Labels = labels

		app.Spec = *originApp.Spec.DeepCopy()

		err = r.Create(context.TODO(), app)
		if err != nil {
			return err
		}
	}

	if app.Status.State == "" {
		// update status if needed
		return updateHelmApplicationStatus(r.Client, originApp.Name, true)
	}

	return nil
}

func (r *ReconcileHelmApplication) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmApplication{}).Complete(r)
}

func inAppStore(app *v1alpha1.HelmApplication) bool {
	return strings.HasSuffix(app.Name, v1alpha1.HelmApplicationAppStoreSuffix)
}
