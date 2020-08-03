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

package namespace

import (
	"bytes"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Namespace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespace{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNamespace{}

// ReconcileNamespace reconciles a Namespace object
type ReconcileNamespace struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Namespace object and makes changes based on the state read
// and what is in the Namespace.Spec
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces/status,verbs=get;update;patch
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Namespace instance
	instance := &corev1.Namespace{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// The object is being deleted
			// our finalizer is present, so lets handle our external dependency
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// name of your custom finalizer
	finalizer := "finalizers.kubesphere.io/namespaces"

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizer)
			if instance.Labels == nil {
				instance.Labels = make(map[string]string)
			}
			instance.Labels[constants.NamespaceLabelKey] = instance.Name
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			if err = r.deleteRouter(instance.Name); err != nil {
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	if err = r.bindWorkspace(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.initRoles(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.initCreatorRoleBinding(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileNamespace) isControlledByWorkspace(namespace *corev1.Namespace) (bool, error) {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	// without workspace label
	if workspaceName == "" {
		return false, nil
	}

	return true, nil
}

func (r *ReconcileNamespace) bindWorkspace(namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	if workspaceName == "" {
		return nil
	}

	workspace := &tenantv1alpha1.Workspace{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: workspaceName}, workspace)

	if err != nil {
		// skip if workspace not found
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	// federated namespace not controlled by workspace
	if namespace.Labels[constants.KubefedManagedLabel] != "true" && !metav1.IsControlledBy(namespace, workspace) {
		namespace.OwnerReferences = nil
		if err := controllerutil.SetControllerReference(workspace, namespace, r.scheme); err != nil {
			klog.Error(err)
			return err
		}
		err = r.Update(context.TODO(), namespace)
		if err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (r *ReconcileNamespace) deleteRouter(namespace string) error {
	routerName := constants.IngressControllerPrefix + namespace
	// delete service first
	found := corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
	}

	err = r.Delete(context.TODO(), &found)
	if err != nil {
		klog.Error(err)
		return err
	}

	// delete deployment
	deploy := appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	err = r.Delete(context.TODO(), &deploy)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func (r *ReconcileNamespace) initRoles(namespace *corev1.Namespace) error {
	var roleBases iamv1alpha2.RoleBaseList

	var labelKey string
	// filtering initial roles by label
	if namespace.Labels[constants.DevOpsProjectLabelKey] != "" {
		// scope.kubesphere.io/devops: ""
		labelKey = fmt.Sprintf(iamv1alpha2.ScopeLabelFormat, iamv1alpha2.ScopeDevOps)
	} else {
		// scope.kubesphere.io/namespace: ""
		labelKey = fmt.Sprintf(iamv1alpha2.ScopeLabelFormat, iamv1alpha2.ScopeNamespace)
	}
	err := r.List(context.Background(), &roleBases, client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(labels.Set{labelKey: ""})})
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBase := range roleBases.Items {
		var role rbacv1.Role
		if err = yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(roleBase.Role.Raw), 1024).Decode(&role); err == nil && role.Kind == iamv1alpha2.ResourceKindRole {
			var old rbacv1.Role
			err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace.Name, Name: role.Name}, &old)
			if err != nil {
				if errors.IsNotFound(err) {
					role.Namespace = namespace.Name
					err = r.Client.Create(context.Background(), &role)
					if err != nil {
						klog.Error(err)
						return err
					}
					continue
				}
			}

			if !reflect.DeepEqual(role.Labels, old.Labels) ||
				!reflect.DeepEqual(role.Annotations, old.Annotations) ||
				!reflect.DeepEqual(role.Rules, old.Rules) {

				old.Labels = role.Labels
				old.Annotations = role.Annotations
				old.Rules = role.Rules

				return r.Update(context.Background(), &old)
			}
		}
	}
	return nil
}

func (r *ReconcileNamespace) initCreatorRoleBinding(namespace *corev1.Namespace) error {
	creator := namespace.Annotations[constants.CreatorAnnotationKey]
	if creator == "" {
		return nil
	}

	var user iamv1alpha2.User
	if err := r.Get(context.Background(), types.NamespacedName{Name: creator}, &user); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	creatorRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", creator, iamv1alpha2.NamespaceAdmin),
			Labels:    map[string]string{iamv1alpha2.UserReferenceLabel: creator},
			Namespace: namespace.Name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     iamv1alpha2.ResourceKindRole,
			Name:     iamv1alpha2.NamespaceAdmin,
		},
		Subjects: []rbacv1.Subject{
			{
				Name:     creator,
				Kind:     iamv1alpha2.ResourceKindUser,
				APIGroup: rbacv1.GroupName,
			},
		},
	}

	if err := r.Client.Create(context.Background(), creatorRoleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	return nil
}
