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

package helmrepo

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const (
	// min sync period in seconds
	MinSyncPeriod = 180

	MinRetryDuration     = 60
	MaxRetryDuration     = 600
	HelmRepoSyncStateLen = 10

	StateSuccess = "successful"
	StateFailed  = "failed"
	MessageLen   = 512
)

const (
	HelmRepoFinalizer = "helmrepo.application.kubesphere.io"
)

// Add creates a new Workspace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHelmRepo{Client: mgr.GetClient(), scheme: mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("workspace-controller"),
		config:   mgr.GetConfig(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("helm-repo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to HelmRelease
	err = c.Watch(&source.Kind{Type: &v1alpha1.HelmRepo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileHelmRepo{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileHelmRepo struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	config   *rest.Config
}

// Reconcile reads that state of the cluster for a helmrepoes object and makes changes based on the state read
// and what is in the helmreleases.Spec
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmrepos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmrepos/status,verbs=get;update;patch
func (r *ReconcileHelmRepo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.Infof("sync repo: %s", request.Name)
	defer func() {
		klog.Infof("sync repo end: %s, elapsed: %v", request.Name, time.Now().Sub(start))
	}()
	// Fetch the helmrepoes instance
	instance := &v1alpha1.HelmRepo{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if instance.Status.State == "" {
			instance.Status.State = v1alpha1.RepoStateSyncing
			return reconcile.Result{}, r.Status().Update(context.Background(), instance)
		}

		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmRepoFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmRepoFinalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmRepoFinalizer) {
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmRepoFinalizer {
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

	copyInstance := instance.DeepCopy()

	if copyInstance.Spec.SyncPeriod != 0 && copyInstance.Spec.SyncPeriod < MinSyncPeriod {
		copyInstance.Spec.SyncPeriod = MinSyncPeriod
	}

	retryAfter := 0
	if syncNow, after := needReSyncNow(copyInstance); syncNow {
		// sync repo
		syncErr := r.syncRepo(copyInstance)
		state := copyInstance.Status.SyncState
		now := metav1.Now()
		if syncErr != nil {
			// failed
			state = append([]v1alpha1.HelmRepoSyncState{{
				State:    v1alpha1.RepoStateFailed,
				Message:  strings.ShortenString(syncErr.Error(), MessageLen),
				SyncTime: &now,
			}}, state...)
			copyInstance.Status.State = v1alpha1.RepoStateFailed
		} else {
			state = append([]v1alpha1.HelmRepoSyncState{{
				State:    v1alpha1.RepoStateSuccessful,
				SyncTime: &now,
			}}, state...)

			copyInstance.Status.Version = instance.Spec.Version
			copyInstance.Status.State = v1alpha1.RepoStateSuccessful
		}

		copyInstance.Status.LastUpdateTime = &now
		if len(state) > HelmRepoSyncStateLen {
			state = state[0:HelmRepoSyncStateLen]
		}
		copyInstance.Status.SyncState = state

		err = r.Client.Status().Update(context.TODO(), copyInstance)
		if err != nil {
			klog.Errorf("update status failed, error: %s", err)
			return reconcile.Result{
				RequeueAfter: MinRetryDuration * time.Second,
			}, err
		} else {
			retryAfter = MinSyncPeriod
			if syncErr == nil {
				retryAfter = copyInstance.Spec.SyncPeriod
			}
		}
	} else {
		retryAfter = after
	}

	return reconcile.Result{
		RequeueAfter: time.Duration(retryAfter) * time.Second,
	}, nil
}

// needReSyncNow checks instance whether need resync now
// if resync is true, it should resync not
// if resync is false and after > 0, it should resync in after seconds
func needReSyncNow(instance *v1alpha1.HelmRepo) (syncNow bool, after int) {

	now := time.Now()
	if instance.Status.SyncState == nil || len(instance.Status.SyncState) == 0 {
		return true, 0
	}

	states := instance.Status.SyncState

	failedTimes := 0
	for i := range states {
		if states[i].State != StateSuccess {
			failedTimes += 1
		} else {
			break
		}
	}

	state := states[0]

	if instance.Spec.Version != instance.Status.Version && failedTimes == 0 {
		// repo has a successful synchronization
		diff := now.Sub(state.SyncTime.Time) / time.Second
		if diff > 0 && diff < MinRetryDuration {
			return false, int(math.Max(10, float64(MinRetryDuration-diff)))
		} else {
			return true, 0
		}
	}

	period := 0
	if state.State != StateSuccess {
		period = MinRetryDuration * failedTimes
		if period > MaxRetryDuration {
			period = MaxRetryDuration
		}
		if now.After(state.SyncTime.Add(time.Duration(period) * time.Second)) {
			return true, 0
		}
	} else {
		period = instance.Spec.SyncPeriod
		if period != 0 {
			if period < MinSyncPeriod {
				period = MinSyncPeriod
			}
			if now.After(state.SyncTime.Add(time.Duration(period) * time.Second)) {
				return true, 0
			}
		} else {
			// need not to sync
			return false, 0
		}
	}

	after = int(state.SyncTime.Time.Add(time.Duration(period) * time.Second).Sub(now).Seconds())

	// may be less than 10 second
	if after <= 10 {
		after = 10
	}
	return false, after
}

func (r *ReconcileHelmRepo) syncRepo(instance *v1alpha1.HelmRepo) error {
	// 1. load index from helm repo
	index, err := helmrepoindex.LoadRepoIndex(context.TODO(), instance.Spec.Url, &instance.Spec.Credential)

	if err != nil {
		klog.Errorf("load index failed, repo: %s, url: %s, err: %s", instance.GetTrueName(), instance.Spec.Url, err)
		return err
	}

	existsSavedIndex := &helmrepoindex.SavedIndex{}
	if len(instance.Status.Data) != 0 {
		existsSavedIndex, err = helmrepoindex.ByteArrayToSavedIndex([]byte(instance.Status.Data))
		if err != nil {
			klog.Errorf("json unmarshal failed, repo: %s,  error: %s", instance.GetTrueName(), err)
			return err
		}
	}

	// 2. merge new index with old index which is stored in crd
	savedIndex := helmrepoindex.MergeRepoIndex(index, existsSavedIndex)

	// 3. save index in crd
	data, err := savedIndex.Bytes()
	if err != nil {
		klog.Errorf("json marshal failed, error: %s", err)
		return err
	}

	instance.Status.Data = string(data)
	return nil
}
