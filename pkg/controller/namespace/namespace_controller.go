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
	"context"
	"fmt"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/apis/core"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	adminDescription    = "Allows admin access to perform any action on any resource, it gives full control over every resource in the namespace."
	operatorDescription = "The maintainer of the namespace who can manage resources other than users and roles in the namespace."
	viewerDescription   = "Allows viewer access to view all resources in the namespace."
)

var (
	log          = logf.Log.WithName("namespace-controller")
	defaultRoles = []rbac.Role{
		{ObjectMeta: metav1.ObjectMeta{Name: "admin", Annotations: map[string]string{constants.DescriptionAnnotationKey: adminDescription}}, Rules: []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "operator", Annotations: map[string]string{constants.DescriptionAnnotationKey: operatorDescription}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}},
			{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch", "logging.kubesphere.io", "monitoring.kubesphere.io", "iam.kubesphere.io", "resources.kubesphere.io", "autoscaling", "alerting.kubesphere.io"}, Resources: []string{"*"}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "viewer", Annotations: map[string]string{constants.DescriptionAnnotationKey: viewerDescription}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}},
	}
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
	return &ReconcileNamespace{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
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
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if err := r.deleteRuntime(instance); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	workspaceName := instance.Labels[constants.WorkspaceLabelKey]

	// delete default role bindings
	if workspaceName == "" {
		adminBinding := &rbac.RoleBinding{}
		adminBinding.Name = "admin"
		adminBinding.Namespace = instance.Name
		log.Info("Deleting default role binding", "namespace", instance.Name, "name", adminBinding.Name)
		err := r.Delete(context.TODO(), adminBinding)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		viewerBinding := &rbac.RoleBinding{}
		viewerBinding.Name = "viewer"
		viewerBinding.Namespace = instance.Name
		log.Info("Deleting default role binding", "namespace", instance.Name, "name", viewerBinding.Name)
		err = r.Delete(context.TODO(), viewerBinding)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if err = r.checkAndBindWorkspace(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.checkAndCreateRoles(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.checkAndCreateRoleBindings(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.checkAndCreateCephSecret(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.checkAndCreateRuntime(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Create default roles
func (r *ReconcileNamespace) checkAndCreateRoles(namespace *corev1.Namespace) error {
	for _, role := range defaultRoles {
		found := &rbac.Role{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: role.Name}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				role := role.DeepCopy()
				role.Namespace = namespace.Name
				log.Info("Creating default role", "namespace", namespace.Name, "role", role.Name)
				err = r.Create(context.TODO(), role)
				if err != nil {
					return err
				}
			}
			return err
		}
	}
	return nil
}

func (r *ReconcileNamespace) checkAndCreateRoleBindings(namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]
	creatorName := namespace.Annotations[constants.CreatorLabelAnnotationKey]

	creator := rbac.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: creatorName}

	workspaceAdminBinding := &rbac.ClusterRoleBinding{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("workspace:%s:admin", workspaceName)}, workspaceAdminBinding)

	if err != nil {
		return err
	}

	adminBinding := &rbac.RoleBinding{}
	adminBinding.Name = "admin"
	adminBinding.Namespace = namespace.Name
	adminBinding.RoleRef = rbac.RoleRef{Name: "admin", APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
	adminBinding.Subjects = workspaceAdminBinding.Subjects

	if creator.Name != "" {
		if adminBinding.Subjects == nil {
			adminBinding.Subjects = make([]rbac.Subject, 0)
		}
		if !k8sutil.ContainsUser(adminBinding.Subjects, creatorName) {
			adminBinding.Subjects = append(adminBinding.Subjects, creator)
		}
	}

	found := &rbac.RoleBinding{}

	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: adminBinding.Name}, found)

	if errors.IsNotFound(err) {
		log.Info("Creating default role binding", "namespace", namespace.Name, "name", adminBinding.Name)
		err = r.Create(context.TODO(), adminBinding)
		if err != nil {
			return err
		}
		found = adminBinding
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, adminBinding.RoleRef) {
		log.Info("Deleting conflict role binding", "namespace", namespace.Name, "name", adminBinding.Name)
		err = r.Delete(context.TODO(), found)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, adminBinding.Name)
	}

	if !reflect.DeepEqual(found.Subjects, adminBinding.Subjects) {
		found.Subjects = adminBinding.Subjects
		log.Info("Updating role binding", "namespace", namespace.Name, "name", adminBinding.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}

	workspaceViewerBinding := &rbac.ClusterRoleBinding{}

	err = r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("workspace:%s:viewer", workspaceName)}, workspaceViewerBinding)

	if err != nil {
		return err
	}

	viewerBinding := &rbac.RoleBinding{}
	viewerBinding.Name = "viewer"
	viewerBinding.Namespace = namespace.Name
	viewerBinding.RoleRef = rbac.RoleRef{Name: "viewer", APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
	viewerBinding.Subjects = workspaceViewerBinding.Subjects

	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: viewerBinding.Name}, found)

	if errors.IsNotFound(err) {
		log.Info("Creating default role binding", "namespace", namespace.Name, "name", viewerBinding.Name)
		err = r.Create(context.TODO(), viewerBinding)
		if err != nil {
			return err
		}
		found = viewerBinding
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, viewerBinding.RoleRef) {
		log.Info("Deleting conflict role binding", "namespace", namespace.Name, "name", viewerBinding.Name)
		err = r.Delete(context.TODO(), found)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, viewerBinding.Name)
	}

	if !reflect.DeepEqual(found.Subjects, viewerBinding.Subjects) {
		found.Subjects = viewerBinding.Subjects
		log.Info("Updating role binding", "namespace", namespace.Name, "name", viewerBinding.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}

	return nil
}

// Create openpitrix runtime
func (r *ReconcileNamespace) checkAndCreateRuntime(namespace *corev1.Namespace) error {

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		return nil
	}

	cm := &corev1.ConfigMap{}
	configName := fmt.Sprintf("kubeconfig-%s", constants.AdminUserName)
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: constants.KubeSphereControlNamespace, Name: configName}, cm)

	if err != nil {
		return err
	}

	runtime := &openpitrix.RunTime{Name: namespace.Name, Zone: namespace.Name, Provider: "kubernetes", RuntimeCredential: cm.Data["config"]}

	log.Info("Creating openpitrix runtime", "namespace", namespace.Name)
	if err := openpitrix.Client().CreateRuntime(runtime); err != nil {
		return err
	}

	return nil
}

// Delete openpitrix runtime
func (r *ReconcileNamespace) deleteRuntime(namespace *corev1.Namespace) error {

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		log.Info("Deleting openpitrix runtime", "namespace", namespace.Name, "runtime", runtimeId)
		if err := openpitrix.Client().DeleteRuntime(runtimeId); err != nil {
			return err
		}
	}

	return nil
}

// Create openpitrix runtime
func (r *ReconcileNamespace) checkAndBindWorkspace(namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	if workspaceName == "" {
		return nil
	}

	workspace := &v1alpha1.Workspace{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: workspaceName}, workspace)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("namespace %s bind workspace %s but not found", namespace.Name, workspaceName))
			return nil
		}
		return err
	}

	if !metav1.IsControlledBy(namespace, workspace) {
		if err := controllerutil.SetControllerReference(workspace, namespace, r.scheme); err != nil {
			return err
		}
		log.Info("Bind workspace", "namespace", namespace.Name, "workspace", workspaceName)
		err = r.Update(context.TODO(), namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

//Create Ceph secret in the new namespace
func (r *ReconcileNamespace) checkAndCreateCephSecret(namespace *corev1.Namespace) error {

	newNsName := namespace.Name
	scList := &v1.StorageClassList{}
	err := r.List(context.TODO(), &client.ListOptions{}, scList)
	if err != nil {
		return err
	}
	for _, sc := range scList.Items {
		if sc.Provisioner == "kubernetes.io/rbd" {
			log.Info("would create Ceph user secret in storage class %s at namespace %s", sc.GetName(), newNsName)
			if secretName, ok := sc.Parameters["userSecretName"]; ok {
				secret := &corev1.Secret{}
				r.Get(context.TODO(), types.NamespacedName{Namespace: core.NamespaceSystem, Name: secretName}, secret)
				if err != nil {
					if errors.IsNotFound(err) {
						log.Error(err, "cannot find secret in namespace %s, error: %s", core.NamespaceSystem, secretName)
						continue
					}
					log.Error(err, fmt.Sprintf("failed to find secret in namespace %s", core.NamespaceSystem))
					continue
				}
				glog.Infof("succeed to find secret %s in namespace %s", secret.GetName(), secret.GetNamespace())

				newSecret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       secret.Kind,
						APIVersion: secret.APIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:                       secret.GetName(),
						Namespace:                  newNsName,
						Labels:                     secret.GetLabels(),
						Annotations:                secret.GetAnnotations(),
						DeletionGracePeriodSeconds: secret.GetDeletionGracePeriodSeconds(),
						ClusterName:                secret.GetClusterName(),
					},
					Data:       secret.Data,
					StringData: secret.StringData,
					Type:       secret.Type,
				}
				log.Info(fmt.Sprintf("creating secret %s in namespace %s...", newSecret.GetName(), newSecret.GetNamespace()))

				err = r.Create(context.TODO(), newSecret)
				if err != nil {
					log.Error(err, fmt.Sprintf("failed to create secret in namespace %s", newSecret.GetNamespace()))
					continue
				}
			} else {
				log.Error(err, fmt.Sprintf("failed to find user secret name in storage class %s", sc.GetName()))
			}
		}
	}

	return nil
}
