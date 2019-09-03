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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/core"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"math"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const (
	adminDescription    = "Allows admin access to perform any action on any resource, it gives full control over every resource in the namespace."
	operatorDescription = "The maintainer of the namespace who can manage resources other than users and roles in the namespace."
	viewerDescription   = "Allows viewer access to view all resources in the namespace."
)

var (
	admin    = rbac.Role{ObjectMeta: metav1.ObjectMeta{Name: "admin", Annotations: map[string]string{constants.DescriptionAnnotationKey: adminDescription, constants.CreatorAnnotationKey: constants.System}}, Rules: []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}}
	operator = rbac.Role{ObjectMeta: metav1.ObjectMeta{Name: "operator", Annotations: map[string]string{constants.DescriptionAnnotationKey: operatorDescription, constants.CreatorAnnotationKey: constants.System}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}},
		{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch", "logging.kubesphere.io", "monitoring.kubesphere.io", "iam.kubesphere.io", "resources.kubesphere.io", "autoscaling", "alerting.kubesphere.io", "app.k8s.io", "servicemesh.kubesphere.io", "operations.kubesphere.io"}, Resources: []string{"*"}}}}
	viewer       = rbac.Role{ObjectMeta: metav1.ObjectMeta{Name: "viewer", Annotations: map[string]string{constants.DescriptionAnnotationKey: viewerDescription, constants.CreatorAnnotationKey: constants.System}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}}
	defaultRoles = []rbac.Role{admin, operator, viewer}
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
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			if err := r.deleteRouter(instance.Name); err != nil {
				return reconcile.Result{}, err
			}

			// delete runtime in the background, retry 3 times
			go r.deleteRuntime(instance)

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

	controlledByWorkspace, err := r.isControlledByWorkspace(instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	if !controlledByWorkspace {

		err = r.deleteRoleBindings(instance)

		return reconcile.Result{}, err
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

func (r *ReconcileNamespace) isControlledByWorkspace(namespace *corev1.Namespace) (bool, error) {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	// without workspace label
	if workspaceName == "" {
		return false, nil
	}

	return true, nil
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
				err = r.Create(context.TODO(), role)
				if err != nil {
					klog.Errorf("creating role namespace: %s,role: %s,error: %s", namespace.Name, role.Name, err)
					return err
				}
			} else {
				klog.Errorf("get role namespace: %s,role: %s,error: %s", namespace.Name, role.Name, err)
				return err
			}
		}
		if !reflect.DeepEqual(found.Rules, role.Rules) {
			found.Rules = role.Rules
			err := r.Update(context.TODO(), found)
			if err != nil {
				klog.Errorf("updating default role namespace: %s, role: %s,error: %s", namespace.Name, role.Name, err)
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileNamespace) checkAndCreateRoleBindings(namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]
	creatorName := namespace.Annotations[constants.CreatorAnnotationKey]

	creator := rbac.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: creatorName}

	workspaceAdminBinding := &rbac.ClusterRoleBinding{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("workspace:%s:admin", workspaceName)}, workspaceAdminBinding)

	if err != nil {
		return err
	}

	adminBinding := &rbac.RoleBinding{}
	adminBinding.Name = admin.Name
	adminBinding.Namespace = namespace.Name
	adminBinding.RoleRef = rbac.RoleRef{Name: admin.Name, APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
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
		err = r.Create(context.TODO(), adminBinding)
		if err != nil {
			klog.Errorf("creating role binding namespace: %s,role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
		found = adminBinding
	} else if err != nil {
		klog.Errorf("get role binding namespace: %s,role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, adminBinding.RoleRef) {
		err = r.Delete(context.TODO(), found)
		if err != nil {
			klog.Errorf("deleting role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
		err = fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, adminBinding.Name)
		klog.Errorf("conflict role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.Subjects, adminBinding.Subjects) {
		found.Subjects = adminBinding.Subjects
		err = r.Update(context.TODO(), found)
		if err != nil {
			klog.Errorf("updating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
	}

	workspaceViewerBinding := &rbac.ClusterRoleBinding{}

	err = r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("workspace:%s:viewer", workspaceName)}, workspaceViewerBinding)

	if err != nil {
		return err
	}

	viewerBinding := &rbac.RoleBinding{}
	viewerBinding.Name = viewer.Name
	viewerBinding.Namespace = namespace.Name
	viewerBinding.RoleRef = rbac.RoleRef{Name: viewer.Name, APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
	viewerBinding.Subjects = workspaceViewerBinding.Subjects

	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: viewerBinding.Name}, found)

	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), viewerBinding)
		if err != nil {
			klog.Errorf("creating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
		found = viewerBinding
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, viewerBinding.RoleRef) {
		err = r.Delete(context.TODO(), found)
		if err != nil {
			klog.Errorf("deleting conflict role binding namespace: %s, role binding: %s, %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
		err = fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, viewerBinding.Name)
		klog.Errorf("conflict role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.Subjects, viewerBinding.Subjects) {
		found.Subjects = viewerBinding.Subjects
		err = r.Update(context.TODO(), found)
		if err != nil {
			klog.Errorf("updating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
	}

	return nil
}

// Create openpitrix runtime
func (r *ReconcileNamespace) checkAndCreateRuntime(namespace *corev1.Namespace) error {
	openPitrixClient, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		return err
	}

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		return nil
	}

	cm := &corev1.ConfigMap{}
	configName := fmt.Sprintf("kubeconfig-%s", constants.AdminUserName)
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: constants.KubeSphereControlNamespace, Name: configName}, cm)

	if err != nil {
		return err
	}

	runtime := &openpitrix.RunTime{Name: namespace.Name, Zone: namespace.Name, Provider: "kubernetes", RuntimeCredential: cm.Data["config"]}

	if err := openPitrixClient.CreateRuntime(runtime); err != nil {
		klog.Errorf("creating openpitrix runtime namespace: %s, error: %s", namespace.Name, err)
		return err
	}

	return nil
}

// Delete openpitrix runtime
func (r *ReconcileNamespace) deleteRuntime(namespace *corev1.Namespace) error {

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		maxRetries := float64(3)
		for i := float64(0); i < maxRetries; i++ {
			time.Sleep(time.Duration(i*math.Pow(2, i)) * time.Second)

			openPitrixClient, err := cs.ClientSets().OpenPitrix()
			if err != nil {
				return err
			}

			err = openPitrixClient.DeleteRuntime(runtimeId)

			if err == nil || openpitrix.IsNotFound(err) || openpitrix.IsDeleted(err) {
				return nil
			}

			klog.Errorf("delete openpitrix runtime: %v times left, error: %s", maxRetries-i-1, err)
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
		// skip if workspace not found
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
		return err
	}

	if !metav1.IsControlledBy(namespace, workspace) {
		if err := controllerutil.SetControllerReference(workspace, namespace, r.scheme); err != nil {
			klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
			return err
		}
		err = r.Update(context.TODO(), namespace)
		if err != nil {
			klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
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
		klog.Errorln(err)
		return err
	}
	for _, sc := range scList.Items {
		if sc.Provisioner == "kubernetes.io/rbd" {
			if secretName, ok := sc.Parameters["userSecretName"]; ok {
				secret := &corev1.Secret{}
				err = r.Get(context.TODO(), types.NamespacedName{Namespace: core.NamespaceSystem, Name: secretName}, secret)
				if err != nil {
					if errors.IsNotFound(err) {
						klog.Errorf("cannot find secret %s in namespace %s, error: %s", secretName, core.NamespaceSystem, err)
						continue
					}
					klog.Errorf("failed to find secret in namespace %s, error: %s", core.NamespaceSystem, err)
					continue
				}

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

				err = r.Create(context.TODO(), newSecret)
				if err != nil {
					klog.Errorf("failed to create secret in namespace %s,error: %s", newSecret.GetNamespace(), err)
					continue
				}
			} else {
				klog.Errorf("failed to find user secret name in storage class %s,error: %s", sc.GetName(), err)
			}
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
		klog.V(6).Info("get router service failed", err)
	}

	err = r.Delete(context.TODO(), &found)
	if err != nil {
		klog.Error(err, "delete router failed")
		return err
	}

	// delete deployment
	deploy := appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.V(6).Info("get router deployment failed", err)
		return err
	}

	err = r.Delete(context.TODO(), &deploy)
	if err != nil {
		klog.Error(err, "delete router deployment failed")
		return err
	}

	return nil

}

func (r *ReconcileNamespace) deleteRoleBindings(namespace *corev1.Namespace) error {
	klog.V(6).Info("deleting role bindings namespace: ", namespace.Name)
	adminBinding := &rbac.RoleBinding{}
	adminBinding.Name = admin.Name
	adminBinding.Namespace = namespace.Name
	err := r.Delete(context.TODO(), adminBinding)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("deleting role binding namespace: %s, role binding: %s,error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}
	viewerBinding := &rbac.RoleBinding{}
	viewerBinding.Name = viewer.Name
	viewerBinding.Namespace = namespace.Name
	err = r.Delete(context.TODO(), viewerBinding)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("deleting role binding namespace: %s,role binding: %s,error: %s", namespace.Name, viewerBinding.Name, err)
		return err
	}
	return nil
}
