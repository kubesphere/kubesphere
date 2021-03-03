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

package quota

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	evaluatorcore "kubesphere.io/kubesphere/kube/pkg/quota/v1/evaluator/core"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/generic"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/install"
	quotav1alpha2 "kubesphere.io/kubesphere/pkg/apis/quota/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	k8sinformers "k8s.io/client-go/informers"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"

	quotav1 "kubesphere.io/kubesphere/kube/pkg/quota/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	ControllerName                 = "resourcequota-controller"
	DefaultResyncPeriod            = 5 * time.Minute
	DefaultMaxConcurrentReconciles = 8
)

// Reconciler reconciles a Workspace object
type Reconciler struct {
	client.Client
	logger                  logr.Logger
	recorder                record.EventRecorder
	maxConcurrentReconciles int
	// Knows how to calculate usage
	registry quotav1.Registry
	// Controls full recalculation of quota usage
	resyncPeriod time.Duration
	scheme       *runtime.Scheme
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int, resyncPeriod time.Duration, informerFactory k8sinformers.SharedInformerFactory) error {
	r.logger = ctrl.Log.WithName("controllers").WithName(ControllerName)
	r.recorder = mgr.GetEventRecorderFor(ControllerName)
	r.scheme = mgr.GetScheme()
	r.registry = generic.NewRegistry(install.NewQuotaConfigurationForControllers(generic.ListerFuncForResourceFunc(informerFactory.ForResource)).Evaluators())
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if maxConcurrentReconciles > 0 {
		r.maxConcurrentReconciles = maxConcurrentReconciles
	} else {
		r.maxConcurrentReconciles = DefaultMaxConcurrentReconciles
	}
	r.resyncPeriod = time.Duration(math.Max(float64(resyncPeriod), float64(DefaultResyncPeriod)))
	c, err := ctrl.NewControllerManagedBy(mgr).
		Named(ControllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.maxConcurrentReconciles,
		}).
		For(&quotav1alpha2.ResourceQuota{}).
		WithEventFilter(predicate.GenerationChangedPredicate{
			Funcs: predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldQuota := e.ObjectOld.(*quotav1alpha2.ResourceQuota)
					newQuota := e.ObjectNew.(*quotav1alpha2.ResourceQuota)
					return !equality.Semantic.DeepEqual(oldQuota.Spec, newQuota.Spec)
				},
			},
		}).
		Build(r)
	if err != nil {
		return err
	}

	resources := []runtime.Object{
		&corev1.Pod{},
		&corev1.Service{},
		&corev1.PersistentVolumeClaim{},
	}
	realClock := clock.RealClock{}
	for _, resource := range resources {
		err := c.Watch(
			&source.Kind{Type: resource},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.mapper)},
			predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					notifyChange := false
					// we only want to queue the updates we care about though as too much noise will overwhelm queue.
					switch e.MetaOld.(type) {
					case *corev1.Pod:
						oldPod := e.ObjectOld.(*corev1.Pod)
						newPod := e.ObjectNew.(*corev1.Pod)
						notifyChange = evaluatorcore.QuotaV1Pod(oldPod, realClock) && !evaluatorcore.QuotaV1Pod(newPod, realClock)
					case *corev1.Service:
						oldService := e.ObjectOld.(*corev1.Service)
						newService := e.ObjectNew.(*corev1.Service)
						notifyChange = evaluatorcore.GetQuotaServiceType(oldService) != evaluatorcore.GetQuotaServiceType(newService)
					case *corev1.PersistentVolumeClaim:
						notifyChange = true
					}
					return notifyChange
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) mapper(h handler.MapObject) []reconcile.Request {
	// check if the quota controller can evaluate this kind, if not, ignore it altogether...
	var result []reconcile.Request
	evaluators := r.registry.List()
	ctx := context.TODO()
	resourceQuotaNames, err := resourceQuotaNamesFor(ctx, r.Client, h.Meta.GetNamespace())
	if err != nil {
		klog.Errorf("failed to get resource quota names for: %v %T %v, err: %v", h.Meta.GetNamespace(), h.Object, h.Meta.GetName(), err)
		return result
	}
	// only queue those quotas that are tracking a resource associated with this kind.
	for _, resourceQuotaName := range resourceQuotaNames {
		resourceQuota := &quotav1alpha2.ResourceQuota{}
		if err := r.Get(ctx, types.NamespacedName{Name: resourceQuotaName}, resourceQuota); err != nil {
			klog.Errorf("failed to get resource quota: %v, err: %v", resourceQuotaName, err)
			return result
		}
		resourceQuotaResources := quotav1.ResourceNames(resourceQuota.Status.Total.Hard)
		for _, evaluator := range evaluators {
			matchedResources := evaluator.MatchingResources(resourceQuotaResources)
			if len(matchedResources) > 0 {
				result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: resourceQuotaName}})
				break
			}
		}
	}
	klog.V(6).Infof("resource quota reconcile after resource change: %v %T %v, %+v", h.Meta.GetNamespace(), h.Object, h.Meta.GetName(), result)
	return result
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("resourcequota", req.NamespacedName)
	rootCtx := context.TODO()
	resourceQuota := &quotav1alpha2.ResourceQuota{}
	if err := r.Get(rootCtx, req.NamespacedName, resourceQuota); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.bindWorkspace(resourceQuota); err != nil {
		logger.Error(err, "failed to set owner reference")
		return ctrl.Result{}, err
	}

	if err := r.syncQuotaForNamespaces(resourceQuota); err != nil {
		logger.Error(err, "failed to sync quota")
		return ctrl.Result{}, err
	}

	r.recorder.Event(resourceQuota, corev1.EventTypeNormal, "Synced", "Synced successfully")
	return ctrl.Result{RequeueAfter: r.resyncPeriod}, nil
}

func (r *Reconciler) bindWorkspace(resourceQuota *quotav1alpha2.ResourceQuota) error {
	workspaceName := resourceQuota.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}

	workspace := &tenantv1alpha1.Workspace{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: workspaceName}, workspace)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	if !metav1.IsControlledBy(resourceQuota, workspace) {
		resourceQuota.OwnerReferences = nil
		if err := controllerutil.SetControllerReference(workspace, resourceQuota, r.scheme); err != nil {
			return err
		}
		err = r.Update(context.TODO(), resourceQuota)
		if err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (r *Reconciler) syncQuotaForNamespaces(originalQuota *quotav1alpha2.ResourceQuota) error {
	quota := originalQuota.DeepCopy()
	ctx := context.TODO()
	// get the list of namespaces that match this cluster quota
	matchingNamespaceList := corev1.NamespaceList{}
	if err := r.List(ctx, &matchingNamespaceList, &client.ListOptions{LabelSelector: labels.SelectorFromSet(quota.Spec.LabelSelector)}); err != nil {
		return err
	}

	matchingNamespaceNames := make([]string, 0)
	for _, namespace := range matchingNamespaceList.Items {
		matchingNamespaceNames = append(matchingNamespaceNames, namespace.Name)
	}

	for _, namespace := range matchingNamespaceList.Items {
		namespaceName := namespace.Name
		namespaceTotals, _ := getResourceQuotasStatusByNamespace(quota.Status.Namespaces, namespaceName)

		actualUsage, err := quotaUsageCalculationFunc(namespaceName, quota.Spec.Quota.Scopes, quota.Spec.Quota.Hard, r.registry, quota.Spec.Quota.ScopeSelector)
		if err != nil {
			return err
		}
		recalculatedStatus := corev1.ResourceQuotaStatus{
			Used: actualUsage,
			Hard: quota.Spec.Quota.Hard,
		}

		// subtract old usage, add new usage
		quota.Status.Total.Used = quotav1.Subtract(quota.Status.Total.Used, namespaceTotals.Used)
		quota.Status.Total.Used = quotav1.Add(quota.Status.Total.Used, recalculatedStatus.Used)
		insertResourceQuotasStatus(&quota.Status.Namespaces, quotav1alpha2.ResourceQuotaStatusByNamespace{
			Namespace:           namespaceName,
			ResourceQuotaStatus: recalculatedStatus,
		})
	}

	// Remove any namespaces from quota.status that no longer match.
	statusCopy := quota.Status.Namespaces.DeepCopy()
	for _, namespaceTotals := range statusCopy {
		namespaceName := namespaceTotals.Namespace
		if !sliceutil.HasString(matchingNamespaceNames, namespaceName) {
			quota.Status.Total.Used = quotav1.Subtract(quota.Status.Total.Used, namespaceTotals.Used)
			removeResourceQuotasStatusByNamespace(&quota.Status.Namespaces, namespaceName)
		}
	}

	quota.Status.Total.Hard = quota.Spec.Quota.Hard

	// if there's no change, no update, return early.  NewAggregate returns nil on empty input
	if equality.Semantic.DeepEqual(quota, originalQuota) {
		return nil
	}

	klog.V(6).Infof("update resource quota: %+v", quota)
	if err := r.Status().Update(ctx, quota, &client.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

// quotaUsageCalculationFunc is a function to calculate quota usage.  It is only configurable for easy unit testing
// NEVER CHANGE THIS OUTSIDE A TEST
var quotaUsageCalculationFunc = quotav1.CalculateUsage
