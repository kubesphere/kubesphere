/*
Copyright 2020 KubeSphere Authors

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

package application

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/networking/v1beta1"
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
	"k8s.io/klog"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Mapper              meta.RESTMapper
	Scheme              *runtime.Scheme
	ApplicationSelector labels.Selector //
}

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var app appv1beta1.Application
	err := r.Get(context.Background(), req.NamespacedName, &app)
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

	resources, errs := r.updateComponents(context.Background(), &app)
	newApplicationStatus := r.getNewApplicationStatus(context.Background(), &app, resources, &errs)

	newApplicationStatus.ObservedGeneration = app.Generation
	if equality.Semantic.DeepEqual(newApplicationStatus, &app.Status) {
		return ctrl.Result{}, nil
	}

	err = r.updateApplicationStatus(context.Background(), req.NamespacedName, newApplicationStatus)
	return ctrl.Result{}, err
}

func (r *ApplicationReconciler) updateComponents(ctx context.Context, app *appv1beta1.Application) ([]*unstructured.Unstructured, []error) {
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

func (r *ApplicationReconciler) getNewApplicationStatus(ctx context.Context, app *appv1beta1.Application, resources []*unstructured.Unstructured, errList *[]error) *appv1beta1.ApplicationStatus {
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

func (r *ApplicationReconciler) fetchComponentListResources(ctx context.Context, groupKinds []metav1.GroupKind, selector *metav1.LabelSelector, namespace string, errs *[]error) []*unstructured.Unstructured {
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

func (r *ApplicationReconciler) setOwnerRefForResources(ctx context.Context, ownerRef metav1.OwnerReference, resources []*unstructured.Unstructured) error {
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

func (r *ApplicationReconciler) objectStatuses(ctx context.Context, resources []*unstructured.Unstructured, errs *[]error) []appv1beta1.ObjectStatus {
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

func (r *ApplicationReconciler) updateApplicationStatus(ctx context.Context, nn types.NamespacedName, status *appv1beta1.ApplicationStatus) error {
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

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		Named("application-controller").
		For(&appv1beta1.Application{}).Build(r)

	if err != nil {
		return err
	}

	sources := []runtime.Object{
		&v1.Deployment{},
		&corev1.Service{},
		&v1.StatefulSet{},
		&v1beta12.Ingress{},
		&servicemeshv1alpha2.ServicePolicy{},
		&servicemeshv1alpha2.Strategy{},
	}

	for _, s := range sources {
		// Watch for changes to Application
		err = c.Watch(
			&source.Kind{Type: s},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
				func(h handler.MapObject) []reconcile.Request {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{
						Name:      servicemesh.GetApplictionName(h.Meta.GetLabels()),
						Namespace: h.Meta.GetNamespace()}}}
				})},
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return isApp(e.MetaOld, e.MetaNew)
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return isApp(e.Meta)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return isApp(e.Meta)
				},
			})
		if err != nil {
			return err
		}
	}
	return nil
}

var _ reconcile.Reconciler = &ApplicationReconciler{}

func isApp(obs ...metav1.Object) bool {
	for _, o := range obs {
		if o.GetLabels() != nil && servicemesh.IsAppComponent(o.GetLabels()) {
			return true
		}
	}
	return false
}
