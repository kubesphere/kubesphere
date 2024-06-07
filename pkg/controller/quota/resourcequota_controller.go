/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package quota

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	quotav1 "kubesphere.io/kubesphere/kube/pkg/quota/v1"
	evaluatorcore "kubesphere.io/kubesphere/kube/pkg/quota/v1/evaluator/core"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/generic"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/install"
	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	controllerName                 = "resourcequota"
	DefaultResyncPeriod            = 5 * time.Minute
	DefaultMaxConcurrentReconciles = 8
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a Workspace object
type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
	// Knows how to calculate usage
	registry quotav1.Registry

	MaxConcurrentReconciles int
	// Controls full recalculation of quota usage
	ResyncPeriod time.Duration

	scheme *runtime.Scheme
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.scheme = mgr.GetScheme()
	r.registry = generic.NewRegistry(install.NewQuotaConfigurationForControllers(mgr.GetClient()).Evaluators())
	r.Client = mgr.GetClient()
	r.MaxConcurrentReconciles = DefaultMaxConcurrentReconciles
	r.ResyncPeriod = DefaultResyncPeriod
	c, err := ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
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

	resources := []client.Object{
		&corev1.Pod{},
		&corev1.Service{},
		&corev1.PersistentVolumeClaim{},
	}
	realClock := clock.RealClock{}
	for _, resource := range resources {
		if err = c.Watch(
			source.Kind(mgr.GetCache(), resource),
			handler.EnqueueRequestsFromMapFunc(r.mapper),
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
					switch e.ObjectOld.(type) {
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
			}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) mapper(ctx context.Context, h client.Object) []reconcile.Request {
	// check if the quota controller can evaluate this kind, if not, ignore it altogether...
	var result []reconcile.Request
	evaluators := r.registry.List()
	resourceQuotaNames, err := resourceQuotaNamesFor(ctx, r.Client, h.GetNamespace())
	if err != nil {
		klog.Errorf("failed to get resource quota names for: %v %T %v, err: %v", h.GetNamespace(), h, h.GetName(), err)
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
	klog.V(6).Infof("resource quota reconcile after resource change: %v %T %v, %+v", h.GetNamespace(), h, h.GetName(), result)
	return result
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	r.recorder.Event(resourceQuota, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return ctrl.Result{RequeueAfter: r.ResyncPeriod}, nil
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

	if quota.Status.Namespaces == nil {
		quota.Status.Namespaces = make([]quotav1alpha2.ResourceQuotaStatusByNamespace, 0)
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
	if err := r.Status().Update(ctx, quota); err != nil {
		return err
	}

	return nil
}

// quotaUsageCalculationFunc is a function to calculate quota usage.  It is only configurable for easy unit testing
// NEVER CHANGE THIS OUTSIDE A TEST
var quotaUsageCalculationFunc = quotav1.CalculateUsage
