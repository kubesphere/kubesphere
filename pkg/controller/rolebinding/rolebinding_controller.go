/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rolebinding

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rbacutils "kubesphere.io/kubesphere/pkg/utils/rbac"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerName = "rolebinding-controller"
	roleBindingRef = "iam.kubesphere.io/rolebinding-ref"
)

var _ kscontroller.Controller = &Reconciler{}

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
		For(&iamv1beta1.RoleBinding{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.WithValues("RoleBinding", req.String())
	ctx = logr.NewContext(ctx, log)
	roleBinding := &iamv1beta1.RoleBinding{}
	if err := r.Get(ctx, req.NamespacedName, roleBinding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.syncToKubernetes(ctx, roleBinding); err != nil {
		log.Error(err, "sync role binding failed")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) syncToKubernetes(ctx context.Context, roleBinding *iamv1beta1.RoleBinding) error {
	k8sRolBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: roleBinding.Namespace, Name: rbacutils.RelatedK8sResourceName(roleBinding.Name)},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, k8sRolBinding, func() error {
		if k8sRolBinding.Labels == nil {
			k8sRolBinding.Labels = make(map[string]string)
		}
		k8sRolBinding.Labels[roleBindingRef] = roleBinding.Name
		k8sRolBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     roleBinding.RoleRef.Kind,
			Name:     rbacutils.RelatedK8sResourceName(roleBinding.RoleRef.Name),
		}
		var subjects []rbacv1.Subject
		for _, subject := range roleBinding.Subjects {
			newSubject := rbacv1.Subject{
				Kind:      subject.Kind,
				APIGroup:  rbacv1.GroupName,
				Name:      subject.Name,
				Namespace: subject.Namespace,
			}
			subjects = append(subjects, newSubject)
		}
		k8sRolBinding.Subjects = subjects
		if err := controllerutil.SetOwnerReference(roleBinding, k8sRolBinding, r.Scheme()); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
		return nil
	})

	if err != nil {
		r.logger.Error(err, "sync role binding failed", "cluster role binding", roleBinding.Name)
	}

	r.logger.V(4).Info("sync role binding to K8s", "cluster role binding", roleBinding.Name, "op", op)
	return nil
}
