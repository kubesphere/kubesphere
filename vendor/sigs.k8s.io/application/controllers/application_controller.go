// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
)

const (
	loggerCtxKey = "logger"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Mapper meta.RESTMapper
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.k8s.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.k8s.io,resources=applications/status,verbs=get;update;patch

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	rootCtx := context.Background()
	logger := r.Log.WithValues("application", req.NamespacedName)
	ctx := context.WithValue(rootCtx, loggerCtxKey, logger)

	var app appv1beta1.Application
	err := r.Get(ctx, req.NamespacedName, &app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Application is in the process of being deleted, so no need to do anything.
	if app.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	resources, err := r.updateComponents(ctx, &app)
	newApplicationStatus := r.getNewApplicationStatus(ctx, &app, resources, err)

	newApplicationStatus.ObservedGeneration = app.Generation
	if equality.Semantic.DeepEqual(newApplicationStatus, &app.Status) {
		return ctrl.Result{}, nil
	}

	err = r.updateApplicationStatus(ctx, req.NamespacedName, newApplicationStatus)
	return ctrl.Result{}, err
}

func (r *ApplicationReconciler) updateComponents(ctx context.Context, app *appv1beta1.Application) ([]*unstructured.Unstructured, error) {
	resources, err := r.fetchComponentListResources(ctx, app.Spec.ComponentGroupKinds, app.Spec.Selector, app.Namespace)
	if err != nil {
		return resources, err
	}

	if app.Spec.AddOwnerRef {
		ownerRef := metav1.NewControllerRef(app, appv1beta1.SchemeGroupVersion.WithKind("Application"))
		*ownerRef.Controller = false
		if err := r.setOwnerRefForResources(ctx, *ownerRef, resources); err != nil {
			return resources, err
		}
	}
	return resources, nil
}

func (r *ApplicationReconciler) getNewApplicationStatus(ctx context.Context, app *appv1beta1.Application, resources []*unstructured.Unstructured, err error) *appv1beta1.ApplicationStatus {
	objectStatuses := r.objectStatuses(ctx, resources)
	aggReady, countReady := aggregateReady(objectStatuses)

	newApplicationStatus := app.Status.DeepCopy()
	newApplicationStatus.ComponentList = appv1beta1.ComponentList{
		Objects: objectStatuses,
	}
	newApplicationStatus.ComponentsReady = fmt.Sprintf("%d/%d", countReady, len(objectStatuses))
	if aggReady {
		setReadyCondition(newApplicationStatus, "ComponentsReady", "all components ready")
	} else {
		setNotReadyCondition(newApplicationStatus, "ComponentsNotReady", "some components not ready")
	}

	if err != nil {
		setErrorCondition(newApplicationStatus, "ErrorSeen", err.Error())
	} else {
		clearErrorCondition(newApplicationStatus)
	}

	return newApplicationStatus
}

func (r *ApplicationReconciler) fetchComponentListResources(ctx context.Context, groupKinds []metav1.GroupKind, selector *metav1.LabelSelector, namespace string) ([]*unstructured.Unstructured, error) {
	logger := getLoggerOrDie(ctx)
	var resources []*unstructured.Unstructured
	for _, gk := range groupKinds {
		mapping, err := r.Mapper.RESTMapping(schema.GroupKind{
			Group: appv1beta1.StripVersion(gk.Group),
			Kind:  gk.Kind,
		})
		if err != nil {
			logger.Info("NoMappingForGK", "gk", gk.String())
			continue
		}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(mapping.GroupVersionKind)
		options := client.ListOptions{
			LabelSelector: labels.SelectorFromSet(selector.MatchLabels),
			Namespace:     namespace,
		}
		if err = r.Client.List(ctx, &options, list); err != nil {
			return resources, err
		}

		for _, u := range list.Items {
			resource := u
			resources = append(resources, &resource)
		}
	}
	return resources, nil
}

func (r *ApplicationReconciler) setOwnerRefForResources(ctx context.Context, ownerRef metav1.OwnerReference, resources []*unstructured.Unstructured) error {
	logger := getLoggerOrDie(ctx)
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
			logger.Error(err, "ErrorSettingOwnerRef", "gvk", resource.GroupVersionKind().String(),
				"namespace", resource.GetNamespace(), "name", resource.GetName())
		}
	}
	return nil
}

func (r *ApplicationReconciler) objectStatuses(ctx context.Context, resources []*unstructured.Unstructured) []appv1beta1.ObjectStatus {
	logger := getLoggerOrDie(ctx)
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
			// Just logging the error for now. Not sure if this is the right way to handle it.
			logger.Error(err, "unable to compute status for resource", "gvk", resource.GroupVersionKind().String(),
				"namespace", resource.GetNamespace(), "name", resource.GetName())
			continue
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.Application{}).
		Complete(r)
}

func getLoggerOrDie(ctx context.Context) logr.Logger {
	logger, ok := ctx.Value(loggerCtxKey).(logr.Logger)
	if !ok {
		panic("context didn't contain logger")
	}
	return logger
}

func Add(manager manager.Manager) error {
	c := ApplicationReconciler{
		Client: manager.GetClient(),
		Mapper: manager.GetRESTMapper(),
		Log:    ctrl.Log.WithName("controllers").WithName("Application"),
		Scheme: manager.GetScheme(),
	}

	return c.SetupWithManager(manager)
}