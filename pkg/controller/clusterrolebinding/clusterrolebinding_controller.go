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

package clusterrolebinding

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Namespace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterRoleBinding{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterrolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Namespace
	err = c.Watch(&source.Kind{Type: &rbac.ClusterRoleBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterRoleBinding{}

// ReconcileClusterRoleBinding reconciles a Namespace object
type ReconcileClusterRoleBinding struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Namespace object and makes changes based on the state read
// and what is in the Namespace.Spec
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces/status,verbs=get;update;patch
func (r *ReconcileClusterRoleBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Namespace instance
	instance := &rbac.ClusterRoleBinding{}
	if err := r.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	workspaceName := instance.Labels[constants.WorkspaceLabelKey]

	if workspaceName != "" && k8sutil.IsControlledBy(instance.OwnerReferences, "Workspace", workspaceName) {
		if instance.Name == iam.GetWorkspaceAdminRoleBindingName(workspaceName) ||
			instance.Name == iam.GetWorkspaceViewerRoleBindingName(workspaceName) {
			nsList := &corev1.NamespaceList{}
			options := client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspaceName})}

			if err := r.List(context.TODO(), &options, nsList); err != nil {
				return reconcile.Result{}, err
			}
			for _, ns := range nsList.Items {
				if !ns.DeletionTimestamp.IsZero() {
					// skip if the namespace is being deleted
					continue
				}
				if err := r.updateRoleBindings(instance, &ns); err != nil {
					return reconcile.Result{}, err
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileClusterRoleBinding) createRoleBindingsIfNotExist(roleBinding *rbac.RoleBinding) error {
	found := &rbac.RoleBinding{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: roleBinding.Namespace, Name: roleBinding.Name}, found); err != nil {
		if errors.IsNotFound(err) {
			err := r.Create(context.TODO(), roleBinding)
			if err != nil {
				log.Errorln(err)
			}
			return err
		}
		log.Errorln(err)
		return err
	}

	if !reflect.DeepEqual(found.Subjects, roleBinding.Subjects) {
		found.Subjects = roleBinding.Subjects
		if err := r.Update(context.TODO(), found); err != nil {
			log.Errorln(err)
			return err
		}
	}

	return nil
}

func (r *ReconcileClusterRoleBinding) updateRoleBindings(clusterRoleBinding *rbac.ClusterRoleBinding, namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	if clusterRoleBinding.Name == iam.GetWorkspaceAdminRoleBindingName(workspaceName) {
		adminBinding := &rbac.RoleBinding{}
		adminBinding.Name = iam.NamespaceAdminRoleBindName
		adminBinding.Namespace = namespace.Name
		adminBinding.RoleRef = rbac.RoleRef{Name: iam.NamespaceAdminRoleName, APIGroup: "rbac.authorization.k8s.io", Kind: iam.RoleKind}
		adminBinding.Subjects = clusterRoleBinding.Subjects

		if err := r.createRoleBindingsIfNotExist(adminBinding); err != nil {
			log.Errorln(err)
			return err
		}
	}

	if clusterRoleBinding.Name == iam.GetWorkspaceViewerRoleBindingName(workspaceName) {
		viewerBinding := &rbac.RoleBinding{}
		viewerBinding.Name = iam.NamespaceViewerRoleBindName
		viewerBinding.Namespace = namespace.Name
		viewerBinding.RoleRef = rbac.RoleRef{Name: iam.NamespaceViewerRoleName, APIGroup: "rbac.authorization.k8s.io", Kind: iam.RoleKind}
		viewerBinding.Subjects = clusterRoleBinding.Subjects

		if err := r.createRoleBindingsIfNotExist(viewerBinding); err != nil {
			log.Errorln(err)
			return err
		}
	}

	return nil
}
