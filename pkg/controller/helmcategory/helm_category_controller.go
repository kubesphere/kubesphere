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

package helmcategory

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const (
	HelmCategoryFinalizer = "helmcategories.application.kubesphere.io"
)

func Add(mgr manager.Manager, multiClusterEnabled bool) error {
	if !multiClusterEnabled {
		_, err := mgr.GetRESTMapper().ResourceSingularizer(v1beta1.FederatedHelmApplicationVersionKind)
		if err != nil {
			klog.Info("federated helm application not exists, exit the controller")
			return nil
		}
	}
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHelmCategory{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("helm-category-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to HelmCategory
	err = c.Watch(&source.Kind{Type: &v1alpha1.HelmCategory{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	reconcileObj := r.(*ReconcileHelmCategory)
	// Watch for changes to FederatedHelmApplication
	err = c.Watch(&source.Kind{Type: &v1beta1.FederatedHelmApplication{}}, &handler.Funcs{
		CreateFunc: func(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
			app := event.Object.(*v1beta1.FederatedHelmApplication)
			err := reconcileObj.updateUncategorizedApplicationLabels(app)
			if err != nil {
				limitingInterface.AddAfter(event, 20*time.Second)
				return
			}

			repoId := app.GetHelmRepoId()
			if repoId == constants.AppStoreRepoId {
				ctgId := app.GetHelmCategoryId()
				if ctgId == "" {
					ctgId = constants.UncategorizedId
				}
				err := reconcileObj.updateCategoryCount(ctgId)
				if err != nil {
					klog.Errorf("reconcile category %s failed, error: %s", ctgId, err)
				}
			}
		},
		UpdateFunc: func(updateEvent event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
			oldApp := updateEvent.ObjectOld.(*v1beta1.FederatedHelmApplication)
			newApp := updateEvent.ObjectNew.(*v1beta1.FederatedHelmApplication)
			err := reconcileObj.updateUncategorizedApplicationLabels(newApp)
			if err != nil {
				limitingInterface.AddAfter(updateEvent, 20*time.Second)
				return
			}
			var oldId string
			repoId := newApp.GetHelmRepoId()
			if repoId == constants.AppStoreRepoId {
				oldId = oldApp.GetHelmCategoryId()
				if oldId == "" {
					oldId = constants.UncategorizedId
				}
				err := reconcileObj.updateCategoryCount(oldId)
				if err != nil {
					klog.Errorf("reconcile category %s failed, error: %s", oldId, err)
				}
			}

			//new labels and new repo id
			repoId = newApp.GetHelmRepoId()
			if repoId == constants.AppStoreRepoId {
				// new category id
				newId := newApp.GetHelmCategoryId()
				if newId == "" {
					newId = constants.UncategorizedId
				}
				if oldId != newId {
					err := reconcileObj.updateCategoryCount(newId)
					if err != nil {
						klog.Errorf("reconcile category %s failed, error: %s", newId, err)
					}
				}
			}
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
			app := deleteEvent.Object.(*v1beta1.FederatedHelmApplication)
			repoId := app.GetHelmRepoId()
			if repoId == constants.AppStoreRepoId {
				id := app.GetHelmCategoryId()
				if id == "" {
					id = constants.UncategorizedId
				}
				err := reconcileObj.updateCategoryCount(id)
				if err != nil {
					klog.Errorf("reconcile category %s failed, error: %s", id, err)
				}
			}
		},
	})
	if err != nil {
		return err
	}

	go func() {
		//create Uncategorized object
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			ctg := &v1alpha1.HelmCategory{}
			err := reconcileObj.Get(context.TODO(), types.NamespacedName{Name: constants.UncategorizedName}, ctg)
			if err != nil && !errors.IsNotFound(err) {
				klog.Errorf("get helm category: %s failed, error: %s", constants.UncategorizedName, err)
			}
			if ctg.Name != "" {
				return
			}

			ctg = &v1alpha1.HelmCategory{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.UncategorizedId,
				},
				Spec: v1alpha1.HelmCategorySpec{
					Description: constants.UncategorizedName,
					Name:        constants.UncategorizedName,
				},
			}
			err = reconcileObj.Create(context.TODO(), ctg)
			if err != nil {
				klog.Errorf("create helm category: %s failed, error: %s", constants.UncategorizedName, err)
			}
		}
	}()
	return nil
}

var _ reconcile.Reconciler = &ReconcileHelmCategory{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileHelmCategory struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	config   *rest.Config
}

// Reconcile reads that state of the cluster for a helmcategories object and makes changes based on the state read
// and what is in the helmreleases.Spec
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmcategories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmcategories/status,verbs=get;update;patch
func (r *ReconcileHelmCategory) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.Infof("sync helm category: %s", request.String())
	defer func() {
		klog.Infof("sync helm category end: %s, elapsed: %v", request.String(), time.Now().Sub(start))
	}()

	instance := &v1alpha1.HelmCategory{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			if request.Name == constants.UncategorizedId {
				err = r.ensureUncategorizedCategory()
				//If create uncategorized category failed, we need create it again
				return reconcile.Result{}, err
			}
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmCategoryFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmCategoryFinalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmCategoryFinalizer) {
			// our finalizer is present, so lets handle our external dependency
			//Delete release
			// remove our finalizer from the list and update it.

			if instance.Status.Total > 0 {
				klog.Errorf("can not delete helm category: %s, which owns applications", request.String())
				return reconcile.Result{}, nil
			}

			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmCategoryFinalizer {
					return true
				}
				return false
			})
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	err = r.updateCategoryCount(instance.Name)
	if err != nil {
		klog.Errorf("update helm category: %s status failed, error: %s", instance.Name, err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHelmCategory) ensureUncategorizedCategory() error {
	ctg := &v1alpha1.HelmCategory{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: constants.UncategorizedId}, ctg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	ctg.Name = constants.UncategorizedId
	ctg.Spec.Name = constants.UncategorizedName
	ctg.Spec.Description = constants.UncategorizedName
	err = r.Create(context.TODO(), ctg)

	return err
}

func (r *ReconcileHelmCategory) updateCategoryCount(id string) error {
	ctg := &v1alpha1.HelmCategory{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: id}, ctg)
	if err != nil {
		return err
	}

	count, err := r.countApplications(id)
	if err != nil {
		return err
	}

	if ctg.Status.Total == count {
		return nil
	}

	ctg.Status.Total = count

	err = r.Status().Update(context.TODO(), ctg)
	return err
}

func (r *ReconcileHelmCategory) countApplications(id string) (int, error) {
	list := v1beta1.FederatedHelmApplicationList{}
	var err error
	err = r.List(context.TODO(), &list, client.MatchingLabels{
		constants.CategoryIdLabelKey:  id,
		constants.ChartRepoIdLabelKey: constants.AppStoreRepoId,
	})

	if err != nil {
		return 0, err
	}

	count := 0
	//just count active helm application
	for _, app := range list.Items {
		if app.Spec.Template.Spec.Status == constants.StateActive {
			count += 1
		}
	}

	return count, nil
}

//add category id to helm application
func (r *ReconcileHelmCategory) updateUncategorizedApplicationLabels(app *v1beta1.FederatedHelmApplication) error {
	if app == nil {
		return nil
	}
	if app.GetHelmRepoId() == constants.AppStoreRepoId {
		ctgId := app.GetHelmCategoryId()
		if ctgId == "" {
			appCopy := app.DeepCopy()
			appCopy.Labels[constants.CategoryIdLabelKey] = constants.UncategorizedId
			patch := client.MergeFrom(app)
			err := r.Client.Patch(context.TODO(), appCopy, patch)
			if err != nil {
				klog.Errorf("patch application: %s failed, error: %s", app.Name, err)
				return err
			}
		}
	}
	return nil
}
