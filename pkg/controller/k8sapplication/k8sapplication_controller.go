/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sapplication

import (
	"context"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const controllerName = "k8sapplication"

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a Application object
type Reconciler struct {
	client.Client
	Mapper              meta.RESTMapper
	Scheme              *runtime.Scheme
	ApplicationSelector labels.Selector
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.Mapper = mgr.GetRESTMapper()
	r.Scheme = mgr.GetScheme()

	selector, err := labels.Parse(mgr.Options.ComposedAppOptions.AppSelector)
	if err != nil {
		return err
	}
	r.ApplicationSelector = selector

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.Application{}).
		Build(r)

	if err != nil {
		return err
	}

	sources := []client.Object{
		&v1.Deployment{},
		&corev1.Service{},
		&v1.StatefulSet{},
		&networkv1.Ingress{},
	}

	for _, s := range sources {
		// Watch for changes to Application
		err = c.Watch(
			source.Kind(mgr.GetCache(), s),
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{
						Name:      GetApplictionName(obj.GetLabels()),
						Namespace: obj.GetNamespace()}}}
				}),
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return isApp(e.ObjectOld, e.ObjectOld)
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return isApp(e.Object)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return isApp(e.Object)
				},
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var app appv1beta1.Application
	err := r.Get(ctx, req.NamespacedName, &app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// If label selector were given, only reconcile matched applications
	// match annotations and labels
	if !r.ApplicationSelector.Empty() {
		if !r.ApplicationSelector.Matches(labels.Set(app.Labels)) &&
			!r.ApplicationSelector.Matches(labels.Set(app.Annotations)) {
			return ctrl.Result{}, err
		}
	}

	// Application is in the process of being deleted, so no need to do anything.
	if app.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	resources, errs := r.updateComponents(ctx, &app)
	newApplicationStatus := r.getNewApplicationStatus(ctx, &app, resources, &errs)

	newApplicationStatus.ObservedGeneration = app.Generation
	if equality.Semantic.DeepEqual(newApplicationStatus, &app.Status) {
		return ctrl.Result{}, nil
	}

	err = r.updateApplicationStatus(ctx, req.NamespacedName, newApplicationStatus)
	return ctrl.Result{}, err
}

func (r *Reconciler) updateComponents(ctx context.Context, app *appv1beta1.Application) ([]*unstructured.Unstructured, []error) {
	var errs []error
	resources := r.fetchComponentListResources(ctx, app.Spec.ComponentGroupKinds, app.Spec.Selector, app.Namespace, &errs)

	if app.Spec.AddOwnerRef {
		ownerRef := metav1.NewControllerRef(app, appv1beta1.GroupVersion.WithKind("Application"))
		*ownerRef.Controller = false
		if err := r.setOwnerRefForResources(ctx, *ownerRef, resources); err != nil {
			errs = append(errs, err)
		}
	}
	return resources, errs
}

func (r *Reconciler) getNewApplicationStatus(ctx context.Context, app *appv1beta1.Application, resources []*unstructured.Unstructured, errList *[]error) *appv1beta1.ApplicationStatus {
	objectStatuses := r.objectStatuses(ctx, resources, errList)
	errs := utilerrors.NewAggregate(*errList)

	aggReady, countReady := aggregateReady(objectStatuses)

	newApplicationStatus := app.Status.DeepCopy()
	newApplicationStatus.ComponentList = appv1beta1.ComponentList{
		Objects: objectStatuses,
	}
	newApplicationStatus.ComponentsReady = fmt.Sprintf("%d/%d", countReady, len(objectStatuses))
	if errs != nil {
		setReadyUnknownCondition(newApplicationStatus, "ComponentsReadyUnknown", "failed to aggregate all components' statuses, check the Error condition for details")
	} else if aggReady {
		setReadyCondition(newApplicationStatus, "ComponentsReady", "all components ready")
	} else {
		setNotReadyCondition(newApplicationStatus, "ComponentsNotReady", fmt.Sprintf("%d components not ready", len(objectStatuses)-countReady))
	}

	if errs != nil {
		setErrorCondition(newApplicationStatus, "ErrorSeen", errs.Error())
	} else {
		clearErrorCondition(newApplicationStatus)
	}

	return newApplicationStatus
}

func (r *Reconciler) fetchComponentListResources(ctx context.Context, groupKinds []metav1.GroupKind, selector *metav1.LabelSelector, namespace string, errs *[]error) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured

	if selector == nil {
		klog.V(2).Info("No selector is specified")
		return resources
	}

	for _, gk := range groupKinds {
		mapping, err := r.Mapper.RESTMapping(schema.GroupKind{
			Group: appv1beta1.StripVersion(gk.Group),
			Kind:  gk.Kind,
		})
		if err != nil {
			klog.V(2).Info("NoMappingForGK", "gk", gk.String())
			continue
		}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(mapping.GroupVersionKind)
		if err = r.Client.List(ctx, list, client.InNamespace(namespace), client.MatchingLabels(selector.MatchLabels)); err != nil {
			klog.Error(err, "unable to list resources for GVK", "gvk", mapping.GroupVersionKind)
			*errs = append(*errs, err)
			continue
		}

		for _, u := range list.Items {
			resource := u
			resources = append(resources, &resource)
		}
	}
	return resources
}

func (r *Reconciler) setOwnerRefForResources(ctx context.Context, ownerRef metav1.OwnerReference, resources []*unstructured.Unstructured) error {
	for _, resource := range resources {
		ownerRefs := resource.GetOwnerReferences()
		ownerRefFound := false
		for i, refs := range ownerRefs {
			if ownerRef.Kind == refs.Kind &&
				ownerRef.APIVersion == refs.APIVersion &&
				ownerRef.Name == refs.Name {
				ownerRefFound = true
				if ownerRef.UID != refs.UID {
					ownerRefs[i] = ownerRef
				}
			}
		}

		if !ownerRefFound {
			ownerRefs = append(ownerRefs, ownerRef)
		}
		resource.SetOwnerReferences(ownerRefs)
		err := r.Client.Update(ctx, resource)
		if err != nil {
			// We log this error, but we continue and try to set the ownerRefs on the other resources.
			klog.Error(err, "ErrorSettingOwnerRef", "gvk", resource.GroupVersionKind().String(),
				"namespace", resource.GetNamespace(), "name", resource.GetName())
		}
	}
	return nil
}

func (r *Reconciler) objectStatuses(ctx context.Context, resources []*unstructured.Unstructured, errs *[]error) []appv1beta1.ObjectStatus {
	var objectStatuses []appv1beta1.ObjectStatus
	for _, resource := range resources {
		os := appv1beta1.ObjectStatus{
			Group: resource.GroupVersionKind().Group,
			Kind:  resource.GetKind(),
			Name:  resource.GetName(),
			Link:  resource.GetSelfLink(),
		}
		s, err := status(resource)
		if err != nil {
			klog.Error(err, "unable to compute status for resource", "gvk", resource.GroupVersionKind().String(),
				"namespace", resource.GetNamespace(), "name", resource.GetName())
			*errs = append(*errs, err)
		}
		os.Status = s
		objectStatuses = append(objectStatuses, os)
	}
	return objectStatuses
}

func aggregateReady(objectStatuses []appv1beta1.ObjectStatus) (bool, int) {
	countReady := 0
	for _, os := range objectStatuses {
		if os.Status == StatusReady {
			countReady++
		}
	}
	if countReady == len(objectStatuses) {
		return true, countReady
	}
	return false, countReady
}

func (r *Reconciler) updateApplicationStatus(ctx context.Context, nn types.NamespacedName, status *appv1beta1.ApplicationStatus) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		original := &appv1beta1.Application{}
		if err := r.Get(ctx, nn, original); err != nil {
			return err
		}
		original.Status = *status
		if err := r.Client.Status().Update(ctx, original); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update status of Application %s/%s: %v", nn.Namespace, nn.Name, err)
	}
	return nil
}

func isApp(obs ...metav1.Object) bool {
	for _, o := range obs {
		if o.GetLabels() != nil && IsAppComponent(o.GetLabels()) {
			return true
		}
	}
	return false
}
