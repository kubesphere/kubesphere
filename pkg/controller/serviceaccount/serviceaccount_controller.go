/*
Copyright 2020 The KubeSphere Authors.

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

package serviceaccount

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	controllerName = "serviceaccount-controller"
)

// Reconciler reconciles a ServiceAccount object
type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
	scheme   *runtime.Scheme
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.logger == nil {
		r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.scheme == nil {
		r.scheme = mgr.GetScheme()
	}
	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor(controllerName)
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("serivceaccount", req.NamespacedName)
	ctx := context.Background()
	sa := &corev1.ServiceAccount{}
	if err := r.Get(ctx, req.NamespacedName, sa); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, ok := sa.Annotations[iamv1alpha2.RoleAnnotation]; ok && sa.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := r.CreateOrUpdateRoleBinding(ctx, logger, sa); err != nil {
			r.recorder.Event(sa, corev1.EventTypeWarning, controllerutils.FailedSynced, err.Error())
			return ctrl.Result{}, err
		}
		r.recorder.Event(sa, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) CreateOrUpdateRoleBinding(ctx context.Context, logger logr.Logger, sa *corev1.ServiceAccount) error {
	roleName := sa.Annotations[iamv1alpha2.RoleAnnotation]
	if roleName == "" {
		return nil
	}
	var role rbacv1.Role
	if err := r.Get(ctx, types.NamespacedName{Name: roleName, Namespace: sa.Namespace}, &role); err != nil {
		return err
	}

	// Delete existing rolebindings.
	saRoleBinding := &rbacv1.RoleBinding{}
	_ = r.Client.DeleteAllOf(ctx, saRoleBinding, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1alpha2.ServiceAccountReferenceLabel: sa.Name})

	saRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", sa.Name, roleName),
			Labels:       map[string]string{iamv1alpha2.ServiceAccountReferenceLabel: sa.Name},
			Namespace:    sa.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     iamv1alpha2.ResourceKindRole,
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

	if err := controllerutil.SetControllerReference(sa, saRoleBinding, r.scheme); err != nil {
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
