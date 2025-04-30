/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package serviceaccount

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	rbacutils "kubesphere.io/kubesphere/pkg/utils/rbac"
)

const (
	controllerName = "serviceaccount"
)

var _ kscontroller.Controller = &Reconciler{}

// Reconciler reconciles a ServiceAccount object
type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("serviceaccount", req.NamespacedName)
	sa := &corev1.ServiceAccount{}
	if err := r.Get(ctx, req.NamespacedName, sa); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, ok := sa.Annotations[iamv1beta1.RoleAnnotation]; ok && sa.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := r.CreateOrUpdateRoleBinding(ctx, logger, sa); err != nil {
			r.recorder.Event(sa, corev1.EventTypeWarning, kscontroller.Synced, err.Error())
			return ctrl.Result{}, err
		}
		r.recorder.Event(sa, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) getReferenceRole(ctx context.Context, roleName, namespace string) (*rbacv1.Role, error) {
	refRole := &rbacv1.Role{}
	refRoleName := rbacutils.RelatedK8sResourceName(roleName)
	if err := r.Get(ctx, types.NamespacedName{Name: refRoleName, Namespace: namespace}, refRole); err != nil {
		return nil, err
	}
	if refRole.Labels[iamv1beta1.RoleReferenceLabel] != roleName {
		return nil, apierrors.NewNotFound(rbacv1.Resource("roles"), refRoleName)
	}
	return refRole, nil
}

func (r *Reconciler) CreateOrUpdateRoleBinding(ctx context.Context, logger logr.Logger, sa *corev1.ServiceAccount) error {
	roleName := sa.Annotations[iamv1beta1.RoleAnnotation]
	if roleName == "" {
		return nil
	}

	role, err := r.getReferenceRole(ctx, roleName, sa.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(4).Info("related role not found", "namespace", sa.Namespace, "role", roleName)
			return nil
		}
		return errors.Wrapf(err, "cannot get reference role %s/%s", sa.Namespace, roleName)
	}

	// Delete existing rolebindings.
	saRoleBinding := &rbacv1.RoleBinding{}
	if err = r.DeleteAllOf(ctx, saRoleBinding, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1beta1.ServiceAccountReferenceLabel: sa.Name}); err != nil {
		return errors.Wrapf(err, "failed to delete RoleBindings for %s/%s", sa.Namespace, sa.Name)
	}

	saRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", sa.Name, roleName),
			Labels:       map[string]string{iamv1beta1.ServiceAccountReferenceLabel: sa.Name},
			Namespace:    sa.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      sa.Name,
				Kind:      rbacv1.ServiceAccountKind,
				Namespace: sa.Namespace,
			},
		},
	}

	if err := controllerutil.SetControllerReference(sa, saRoleBinding, r.Scheme()); err != nil {
		return errors.Wrapf(err, "failed to set controller reference for RoleBinding %s/%s", sa.Namespace, saRoleBinding.Name)
	}

	logger.V(4).Info("create ServiceAccount rolebinding", "ServiceAccount", sa.Name)
	if err := r.Create(ctx, saRoleBinding); err != nil {
		return errors.Wrapf(err, "failed to create RoleBinding %s/%s", sa.Namespace, saRoleBinding.Name)
	}
	return nil
}
