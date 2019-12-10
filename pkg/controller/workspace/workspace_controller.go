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

package workspace

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
)

const (
	workspaceAdminDescription   = "Allows admin access to perform any action on any resource, it gives full control over every resource in the workspace."
	workspaceRegularDescription = "Normal user in the workspace, can create namespace and DevOps project."
	workspaceViewerDescription  = "Allows viewer access to view all resources in the workspace."
)

var log = logf.Log.WithName("workspace-controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Workspace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWorkspace{Client: mgr.GetClient(), scheme: mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("workspace-controller"), ksclient: cs.ClientSets().KubeSphere()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("workspace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Workspace
	err = c.Watch(&source.Kind{Type: &tenantv1alpha1.Workspace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileWorkspace{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileWorkspace struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	ksclient kubesphere.Interface
}

// Reconcile reads that state of the cluster for a Workspace object and makes changes based on the state read
// and what is in the Workspace.Spec
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces/status,verbs=get;update;patch
func (r *ReconcileWorkspace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Workspace instance
	instance := &tenantv1alpha1.Workspace{}
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

	// name of your custom finalizer
	finalizer := "finalizers.tenant.kubesphere.io"

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
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteDevOpsProjects(instance); err != nil {
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

	if err = r.createWorkspaceAdmin(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createWorkspaceRegular(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createWorkspaceViewer(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createWorkspaceRoleBindings(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.bindNamespaces(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileWorkspace) createWorkspaceAdmin(instance *tenantv1alpha1.Workspace) error {
	found := &rbac.ClusterRole{}

	admin := getWorkspaceAdmin(instance.Name)

	if err := controllerutil.SetControllerReference(instance, admin, r.scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: admin.Name}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating workspace role", "workspace", instance.Name, "name", admin.Name)
		err = r.Create(context.TODO(), admin)
		if err != nil {
			return err
		}
		found = admin
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(admin.Rules, found.Rules) || !reflect.DeepEqual(admin.Labels, found.Labels) || !reflect.DeepEqual(admin.Annotations, found.Annotations) {
		found.Rules = admin.Rules
		found.Labels = admin.Labels
		found.Annotations = admin.Annotations
		log.Info("Updating workspace role", "workspace", instance.Name, "name", admin.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileWorkspace) createWorkspaceRegular(instance *tenantv1alpha1.Workspace) error {
	found := &rbac.ClusterRole{}

	regular := getWorkspaceRegular(instance.Name)

	if err := controllerutil.SetControllerReference(instance, regular, r.scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: regular.Name}, found)

	if err != nil && errors.IsNotFound(err) {

		log.Info("Creating workspace role", "workspace", instance.Name, "name", regular.Name)
		err = r.Create(context.TODO(), regular)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		found = regular
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(regular.Rules, found.Rules) || !reflect.DeepEqual(regular.Labels, found.Labels) || !reflect.DeepEqual(regular.Annotations, found.Annotations) {
		found.Rules = regular.Rules
		found.Labels = regular.Labels
		found.Annotations = regular.Annotations
		log.Info("Updating workspace role", "workspace", instance.Name, "name", regular.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileWorkspace) createWorkspaceViewer(instance *tenantv1alpha1.Workspace) error {
	found := &rbac.ClusterRole{}

	viewer := getWorkspaceViewer(instance.Name)

	if err := controllerutil.SetControllerReference(instance, viewer, r.scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: viewer.Name}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating workspace role", "workspace", instance.Name, "name", viewer.Name)
		err = r.Create(context.TODO(), viewer)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		found = viewer
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(viewer.Rules, found.Rules) || !reflect.DeepEqual(viewer.Labels, found.Labels) || !reflect.DeepEqual(viewer.Annotations, found.Annotations) {
		found.Rules = viewer.Rules
		found.Labels = viewer.Labels
		found.Annotations = viewer.Annotations
		log.Info("Updating workspace role", "workspace", instance.Name, "name", viewer.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileWorkspace) createGroup(instance *tenantv1alpha1.Workspace) error {
	_, err := r.ksclient.DescribeGroup(instance.Name)

	group := &models.Group{
		Name: instance.Name,
	}

	if err != nil && kubesphere.IsNotFound(err) {
		log.Info("Creating group", "group name", instance.Name)
		_, err = r.ksclient.CreateGroup(group)
		if err != nil {
			if kubesphere.IsExist(err) {
				return nil
			}
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileWorkspace) deleteGroup(instance *tenantv1alpha1.Workspace) error {
	log.Info("Creating group", "group name", instance.Name)
	if err := r.ksclient.DeleteGroup(instance.Name); err != nil {
		if kubesphere.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *ReconcileWorkspace) deleteDevOpsProjects(instance *tenantv1alpha1.Workspace) error {
	if _, err := cs.ClientSets().Devops(); err != nil {
		// skip if devops is not enabled
		if _, notEnabled := err.(cs.ClientSetNotEnabledError); notEnabled {
			return nil
		} else {
			log.Error(err, "")
			return err
		}
	}
	var wg sync.WaitGroup

	log.Info("Delete DevOps Projects")
	for {
		errChan := make(chan error, 10)
		projects, err := r.ksclient.ListWorkspaceDevOpsProjects(instance.Name)
		if err != nil {
			log.Error(err, "Failed to Get Workspace's DevOps Projects", "ws", instance.Name)
			return err
		}
		if projects.TotalCount == 0 {
			return nil
		}
		for _, project := range projects.Items {
			wg.Add(1)
			go func(workspace, devops string) {
				err := r.ksclient.DeleteWorkspaceDevOpsProjects(workspace, devops)
				errChan <- err
				wg.Done()
			}(instance.Name, project.ProjectId)
		}
		wg.Wait()
		close(errChan)
		for err := range errChan {
			if err != nil {
				log.Error(err, "delete devops project error")
				return err
			}
		}

	}
}

func (r *ReconcileWorkspace) createWorkspaceRoleBindings(instance *tenantv1alpha1.Workspace) error {

	adminRoleBinding := &rbac.ClusterRoleBinding{}
	adminRoleBinding.Name = getWorkspaceAdminRoleBindingName(instance.Name)
	adminRoleBinding.Labels = map[string]string{constants.WorkspaceLabelKey: instance.Name}
	adminRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getWorkspaceAdminRoleName(instance.Name)}

	workspaceManager := rbac.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: instance.Spec.Manager}

	if workspaceManager.Name != "" {
		adminRoleBinding.Subjects = []rbac.Subject{workspaceManager}
	} else {
		adminRoleBinding.Subjects = []rbac.Subject{}
	}

	if err := controllerutil.SetControllerReference(instance, adminRoleBinding, r.scheme); err != nil {
		return err
	}

	foundRoleBinding := &rbac.ClusterRoleBinding{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: adminRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating workspace role binding", "workspace", instance.Name, "name", adminRoleBinding.Name)
		err = r.Create(context.TODO(), adminRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = adminRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(adminRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict workspace role binding", "workspace", instance.Name, "name", adminRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict workspace role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	if workspaceManager.Name != "" && !hasSubject(foundRoleBinding.Subjects, workspaceManager) {
		foundRoleBinding.Subjects = append(foundRoleBinding.Subjects, workspaceManager)
		log.Info("Updating workspace role binding", "workspace", instance.Name, "name", adminRoleBinding.Name)
		err = r.Update(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
	}

	regularRoleBinding := &rbac.ClusterRoleBinding{}
	regularRoleBinding.Name = getWorkspaceRegularRoleBindingName(instance.Name)
	regularRoleBinding.Labels = map[string]string{constants.WorkspaceLabelKey: instance.Name}
	regularRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getWorkspaceRegularRoleName(instance.Name)}
	regularRoleBinding.Subjects = []rbac.Subject{}

	if err = controllerutil.SetControllerReference(instance, regularRoleBinding, r.scheme); err != nil {
		return err
	}

	err = r.Get(context.TODO(), types.NamespacedName{Name: regularRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating workspace role binding", "workspace", instance.Name, "name", regularRoleBinding.Name)
		err = r.Create(context.TODO(), regularRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = regularRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(regularRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict workspace role binding", "workspace", instance.Name, "name", regularRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict workspace role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	viewerRoleBinding := &rbac.ClusterRoleBinding{}
	viewerRoleBinding.Name = getWorkspaceViewerRoleBindingName(instance.Name)
	viewerRoleBinding.Labels = map[string]string{constants.WorkspaceLabelKey: instance.Name}
	viewerRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getWorkspaceViewerRoleName(instance.Name)}
	viewerRoleBinding.Subjects = []rbac.Subject{}

	if err = controllerutil.SetControllerReference(instance, viewerRoleBinding, r.scheme); err != nil {
		return err
	}

	err = r.Get(context.TODO(), types.NamespacedName{Name: viewerRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating workspace role binding", "workspace", instance.Name, "name", viewerRoleBinding.Name)
		err = r.Create(context.TODO(), viewerRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = viewerRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(viewerRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict workspace role binding", "workspace", instance.Name, "name", viewerRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict workspace role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	return nil
}

func (r *ReconcileWorkspace) bindNamespaces(instance *tenantv1alpha1.Workspace) error {

	nsList := &corev1.NamespaceList{}
	options := client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: instance.Name})}
	err := r.List(context.TODO(), nsList, &options)

	if err != nil {
		log.Error(err, fmt.Sprintf("list workspace %s namespace failed", instance.Name))
		return err
	}

	for _, namespace := range nsList.Items {
		if !metav1.IsControlledBy(&namespace, instance) {
			if err := controllerutil.SetControllerReference(instance, &namespace, r.scheme); err != nil {
				return err
			}
			log.Info("Bind workspace", "namespace", namespace.Name, "workspace", instance.Name)
			err = r.Update(context.TODO(), &namespace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func hasSubject(subjects []rbac.Subject, user rbac.Subject) bool {
	for _, subject := range subjects {
		if reflect.DeepEqual(subject, user) {
			return true
		}
	}
	return false
}

func getWorkspaceAdminRoleName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:admin", workspaceName)
}
func getWorkspaceRegularRoleName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:regular", workspaceName)
}
func getWorkspaceViewerRoleName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:viewer", workspaceName)
}

func getWorkspaceAdminRoleBindingName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:admin", workspaceName)
}

func getWorkspaceRegularRoleBindingName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:regular", workspaceName)
}

func getWorkspaceViewerRoleBindingName(workspaceName string) string {
	return fmt.Sprintf("workspace:%s:viewer", workspaceName)
}

func getWorkspaceAdmin(workspaceName string) *rbac.ClusterRole {
	admin := &rbac.ClusterRole{}
	admin.Name = getWorkspaceAdminRoleName(workspaceName)
	admin.Labels = map[string]string{constants.WorkspaceLabelKey: workspaceName}
	admin.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.WorkspaceAdmin, constants.DescriptionAnnotationKey: workspaceAdminDescription, constants.CreatorAnnotationKey: constants.System}
	admin.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"*"},
			APIGroups:     []string{"*"},
			ResourceNames: []string{workspaceName},
			Resources:     []string{"workspaces", "workspaces/*"},
		},
		{
			Verbs:     []string{"watch"},
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
		},
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"iam.kubesphere.io"},
			Resources: []string{"users"},
		},
		{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"openpitrix.io"},
			Resources: []string{"categories"},
		},
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"openpitrix.io"},
			Resources: []string{"applications", "apps", "apps/versions", "apps/events", "apps/action", "apps/audits", "repos", "repos/action", "attachments"},
		},
	}

	return admin
}

func getWorkspaceRegular(workspaceName string) *rbac.ClusterRole {
	regular := &rbac.ClusterRole{}
	regular.Name = getWorkspaceRegularRoleName(workspaceName)
	regular.Labels = map[string]string{constants.WorkspaceLabelKey: workspaceName}
	regular.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.WorkspaceRegular, constants.DescriptionAnnotationKey: workspaceRegularDescription, constants.CreatorAnnotationKey: constants.System}
	regular.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"*"},
			Resources:     []string{"workspaces"},
			ResourceNames: []string{workspaceName},
		}, {
			Verbs:         []string{"create"},
			APIGroups:     []string{"tenant.kubesphere.io"},
			Resources:     []string{"workspaces/namespaces", "workspaces/devops"},
			ResourceNames: []string{workspaceName},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"iam.kubesphere.io"},
			ResourceNames: []string{workspaceName},
			Resources:     []string{"workspaces/members"},
		},
		{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"openpitrix.io"},
			Resources: []string{"apps/events", "apps/action", "apps/audits", "categories"},
		},

		{
			Verbs:     []string{"*"},
			APIGroups: []string{"openpitrix.io"},
			Resources: []string{"applications", "apps", "apps/versions", "repos", "repos/action", "attachments"},
		},
	}

	return regular
}

func getWorkspaceViewer(workspaceName string) *rbac.ClusterRole {
	viewer := &rbac.ClusterRole{}
	viewer.Name = getWorkspaceViewerRoleName(workspaceName)
	viewer.Labels = map[string]string{constants.WorkspaceLabelKey: workspaceName}
	viewer.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.WorkspaceViewer, constants.DescriptionAnnotationKey: workspaceViewerDescription, constants.CreatorAnnotationKey: constants.System}
	viewer.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"get", "list"},
			APIGroups:     []string{"*"},
			ResourceNames: []string{workspaceName},
			Resources:     []string{"workspaces", "workspaces/*"},
		},
		{
			Verbs:     []string{"watch"},
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
		},
		{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"openpitrix.io"},
			Resources: []string{"applications", "apps", "apps/events", "apps/action", "apps/audits", "apps/versions", "repos", "categories", "attachments"},
		},
	}
	return viewer
}
