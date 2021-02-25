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
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	controllerName = "namespace-controller"
)

// Reconciler reconciles a Namespace object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger == nil {
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&corev1.Namespace{}).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=rolebases,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("namespace", req.NamespacedName)
	rootCtx := context.Background()
	namespace := &corev1.Namespace{}
	if err := r.Get(rootCtx, req.NamespacedName, namespace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// name of your custom finalizer
	finalizer := "finalizers.kubesphere.io/namespaces"

	if namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(namespace.ObjectMeta.Finalizers, finalizer) {
			// create only once, ignore already exists error
			if err := r.initCreatorRoleBinding(rootCtx, logger, namespace); err != nil {
				return ctrl.Result{}, err
			}
			namespace.ObjectMeta.Finalizers = append(namespace.ObjectMeta.Finalizers, finalizer)
			if namespace.Labels == nil {
				namespace.Labels = make(map[string]string)
			}
			// used for NetworkPolicyPeer.NamespaceSelector
			namespace.Labels[constants.NamespaceLabelKey] = namespace.Name
			if err := r.Update(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(namespace.ObjectMeta.Finalizers, finalizer) {
			if err := r.deleteRouter(rootCtx, logger, namespace.Name); err != nil {
				return ctrl.Result{}, err
			}
			// remove our finalizer from the list and update it.
			namespace.ObjectMeta.Finalizers = sliceutil.RemoveString(namespace.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			if err := r.Update(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}

	// Bind to workspace if the namespace created by kubesphere
	_, hasWorkspaceLabel := namespace.Labels[tenantv1alpha1.WorkspaceLabel]
	if hasWorkspaceLabel {
		if err := r.bindWorkspace(rootCtx, logger, namespace); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.unbindWorkspace(rootCtx, logger, namespace); err != nil {
			return ctrl.Result{}, err
		}
	}
	// Initialize roles for devops/project namespaces if created by kubesphere
	_, hasDevOpsProjectLabel := namespace.Labels[constants.DevOpsProjectLabelKey]
	if hasDevOpsProjectLabel || hasWorkspaceLabel {
		if err := r.initRoles(rootCtx, logger, namespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Recorder.Event(namespace, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	workspace := &tenantv1alpha1.Workspace{}
	if err := r.Get(ctx, types.NamespacedName{Name: namespace.Labels[constants.WorkspaceLabelKey]}, workspace); err != nil {
		// remove existed owner reference if workspace not found
		if errors.IsNotFound(err) && k8sutil.IsControlledBy(namespace.OwnerReferences, tenantv1alpha1.ResourceKindWorkspace, "") {
			return r.unbindWorkspace(ctx, logger, namespace)
		}
		// skip if workspace not found
		return client.IgnoreNotFound(err)
	}
	// owner reference not match workspace label
	if !metav1.IsControlledBy(namespace, workspace) {
		namespace := namespace.DeepCopy()
		namespace.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(namespace.OwnerReferences)
		if err := controllerutil.SetControllerReference(workspace, namespace, scheme.Scheme); err != nil {
			logger.Error(err, "set controller reference failed")
			return err
		}
		logger.V(4).Info("update namespace owner reference", "workspace", workspace.Name)
		if err := r.Update(ctx, namespace); err != nil {
			logger.Error(err, "update namespace failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) unbindWorkspace(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	if k8sutil.IsControlledBy(namespace.OwnerReferences, tenantv1alpha1.ResourceKindWorkspace, "") {
		namespace := namespace.DeepCopy()
		namespace.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(namespace.OwnerReferences)
		logger.V(4).Info("remove owner reference", "workspace", namespace.Labels[constants.WorkspaceLabelKey])
		if err := r.Update(ctx, namespace); err != nil {
			logger.Error(err, "update owner reference failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) deleteRouter(ctx context.Context, logger logr.Logger, namespace string) error {
	routerName := constants.IngressControllerPrefix + namespace

	// delete service first
	service := corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &service)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	logger.V(4).Info("delete router service", "namespace", service.Namespace, "service", service.Name)
	err = r.Delete(ctx, &service)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	// delete deployment
	deploy := appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &deploy)
	if err != nil {
		logger.Error(err, "delete router deployment failed")
		return err
	}

	logger.V(4).Info("delete router deployment", "namespace", deploy.Namespace, "deployment", deploy.Name)
	err = r.Delete(ctx, &deploy)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	return nil
}

func (r *Reconciler) initRoles(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	var templates iamv1alpha2.RoleBaseList
	var labelKey string
	// filtering initial roles by label
	if namespace.Labels[constants.DevOpsProjectLabelKey] != "" {
		// scope.kubesphere.io/devops: ""
		labelKey = fmt.Sprintf(iamv1alpha2.ScopeLabelFormat, iamv1alpha2.ScopeDevOps)
	} else {
		// scope.kubesphere.io/namespace: ""
		labelKey = fmt.Sprintf(iamv1alpha2.ScopeLabelFormat, iamv1alpha2.ScopeNamespace)
	}

	if err := r.List(ctx, &templates, client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(labels.Set{labelKey: ""})}); err != nil {
		logger.Error(err, "list role bases failed")
		return err
	}
	for _, template := range templates.Items {
		var role rbacv1.Role
		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(template.Role.Raw), 1024).Decode(&role); err == nil && role.Kind == iamv1alpha2.ResourceKindRole {
			var old rbacv1.Role
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace.Name, Name: role.Name}, &old); err != nil {
				if errors.IsNotFound(err) {
					role.Namespace = namespace.Name
					logger.V(4).Info("init builtin role", "role", role.Name)
					if err := r.Client.Create(ctx, &role); err != nil {
						logger.Error(err, "create role failed")
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

				logger.V(4).Info("update builtin role", "role", role.Name)
				if err := r.Update(ctx, &old); err != nil {
					logger.Error(err, "update role failed")
					return err
				}
			}
		} else if err != nil {
			logger.Error(fmt.Errorf("invalid role base found"), "init roles failed", "name", template.Name)
		}
	}
	return nil
}

func (r *Reconciler) initCreatorRoleBinding(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	creator := namespace.Annotations[constants.CreatorAnnotationKey]
	if creator == "" {
		return nil
	}
	var user iamv1alpha2.User
	if err := r.Get(ctx, types.NamespacedName{Name: creator}, &user); err != nil {
		return client.IgnoreNotFound(err)
	}
	creatorRoleBinding := newCreatorRoleBinding(creator, namespace.Name)
	logger.V(4).Info("init creator role binding", "creator", user.Name)
	if err := r.Client.Create(ctx, creatorRoleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		logger.Error(err, "create role binding failed")
		return err
	}
	return nil
}

func newCreatorRoleBinding(creator string, namespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", creator, iamv1alpha2.NamespaceAdmin),
			Labels:    map[string]string{iamv1alpha2.UserReferenceLabel: creator},
			Namespace: namespace,
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
}
