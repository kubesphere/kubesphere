/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package clusterrolebinding

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	rbacutils "kubesphere.io/kubesphere/pkg/utils/rbac"
)

const (
	controllerName = "clusterrolebinding"
	roleBindingRef = "iam.kubesphere.io/clusterrolebinding-ref"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
	helper   *rbachelper.Helper
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()
	r.helper = rbachelper.NewHelper(mgr.GetClient())

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.ClusterRoleBinding{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.WithValues("ClusterRoleBinding", req.String())
	ctx = logr.NewContext(ctx, log)
	clusterRole := &iamv1beta1.ClusterRoleBinding{}
	if err := r.Get(ctx, req.NamespacedName, clusterRole); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.syncToKubernetes(ctx, clusterRole); err != nil {
		log.Error(err, "sync cluster role binding failed")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) syncToKubernetes(ctx context.Context, clusterRoleBinding *iamv1beta1.ClusterRoleBinding) error {
	k8sClusterRolBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: rbacutils.RelatedK8sResourceName(clusterRoleBinding.Name)},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, k8sClusterRolBinding, func() error {
		if k8sClusterRolBinding.Labels == nil {
			k8sClusterRolBinding.Labels = make(map[string]string)
		}
		k8sClusterRolBinding.Labels[roleBindingRef] = clusterRoleBinding.Name
		k8sClusterRolBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     clusterRoleBinding.RoleRef.Kind,
			Name:     rbacutils.RelatedK8sResourceName(clusterRoleBinding.RoleRef.Name),
		}
		var subjects []rbacv1.Subject
		for _, subject := range clusterRoleBinding.Subjects {
			newSubject := rbacv1.Subject{
				Kind:      subject.Kind,
				Name:      subject.Name,
				Namespace: subject.Namespace,
			}
			if subject.Kind != rbacv1.ServiceAccountKind {
				newSubject.APIGroup = rbacv1.GroupName
			}
			subjects = append(subjects, newSubject)
		}
		k8sClusterRolBinding.Subjects = subjects
		if err := controllerutil.SetOwnerReference(clusterRoleBinding, k8sClusterRolBinding, r.Scheme()); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
		return nil
	})

	if err != nil {
		r.logger.Error(err, "sync cluster role binding failed", "cluster role binding", clusterRoleBinding.Name)
	}

	r.logger.V(4).Info("sync cluster role binding to K8s", "cluster role binding", clusterRoleBinding.Name, "op", op)
	return nil
}
