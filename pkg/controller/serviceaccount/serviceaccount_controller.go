/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package serviceaccount

import (
	"context"
	"fmt"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
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
	logger := r.logger.WithValues("serivceaccount", req.NamespacedName)
	// ctx := context.Background()
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

func (r *Reconciler) CreateOrUpdateRoleBinding(ctx context.Context, logger logr.Logger, sa *corev1.ServiceAccount) error {
	roleName := sa.Annotations[iamv1beta1.RoleAnnotation]
	if roleName == "" {
		return nil
	}
	var role rbacv1.Role
	if err := r.Get(ctx, types.NamespacedName{Name: roleName, Namespace: sa.Namespace}, &role); err != nil {
		return err
	}

	// Delete existing rolebindings.
	saRoleBinding := &rbacv1.RoleBinding{}
	_ = r.Client.DeleteAllOf(ctx, saRoleBinding, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1beta1.ServiceAccountReferenceLabel: sa.Name})

	saRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", sa.Name, roleName),
			Labels:       map[string]string{iamv1beta1.ServiceAccountReferenceLabel: sa.Name},
			Namespace:    sa.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     iamv1beta1.ResourceKindRole,
			Name:     roleName,
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
		logger.Error(err, "set controller reference failed")
		return err
	}

	logger.V(4).Info("create ServiceAccount rolebinding", "ServiceAccount", sa.Name)
	if err := r.Client.Create(ctx, saRoleBinding); err != nil {
		logger.Error(err, "create rolebinding failed")
		return err
	}
	return nil
}
