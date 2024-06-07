/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package loginrecord

import (
	"context"
	"sort"
	"strings"
	"time"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const controllerName = "loginrecord"

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	client.Client
	recorder                    record.EventRecorder
	loginHistoryRetentionPeriod time.Duration
	loginHistoryMaximumEntries  int
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.loginHistoryRetentionPeriod = mgr.AuthenticationOptions.LoginHistoryRetentionPeriod
	r.loginHistoryMaximumEntries = mgr.AuthenticationOptions.LoginHistoryMaximumEntries
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()

	return builder.
		ControllerManagedBy(mgr).
		For(
			&iamv1beta1.LoginRecord{},
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Named(controllerName).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	loginRecord := &iamv1beta1.LoginRecord{}
	if err := r.Get(ctx, req.NamespacedName, loginRecord); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !loginRecord.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}

	user, err := r.userForLoginRecord(ctx, loginRecord)
	if err != nil {
		// delete orphan object
		if errors.IsNotFound(err) {
			if err = r.Delete(ctx, loginRecord); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err = r.updateUserLastLoginTime(ctx, user, loginRecord); err != nil {
		return ctrl.Result{}, err
	}

	result := ctrl.Result{}
	now := time.Now()
	// login record beyonds retention period
	if loginRecord.CreationTimestamp.Add(r.loginHistoryRetentionPeriod).Before(now) {
		if err = r.Delete(ctx, loginRecord, client.GracePeriodSeconds(0)); err != nil {
			return ctrl.Result{}, err
		}
	} else { // put item back into the queue
		result = ctrl.Result{
			RequeueAfter: loginRecord.CreationTimestamp.Add(r.loginHistoryRetentionPeriod).Sub(now),
		}
	}

	if err = r.shrinkEntriesFor(ctx, user); err != nil {
		return ctrl.Result{}, err
	}

	r.recorder.Event(loginRecord, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return result, nil
}

// updateUserLastLoginTime accepts a login object and set user lastLoginTime field
func (r *Reconciler) updateUserLastLoginTime(ctx context.Context, user *iamv1beta1.User, loginRecord *iamv1beta1.LoginRecord) error {
	// update lastLoginTime
	if user.DeletionTimestamp.IsZero() &&
		(user.Status.LastLoginTime == nil || user.Status.LastLoginTime.Before(&loginRecord.CreationTimestamp)) {
		user.Status.LastLoginTime = &loginRecord.CreationTimestamp
		return r.Update(ctx, user)
	}
	return nil
}

// shrinkEntriesFor will delete old entries out of limit
func (r *Reconciler) shrinkEntriesFor(ctx context.Context, user *iamv1beta1.User) error {
	loginRecords := &iamv1beta1.LoginRecordList{}
	if err := r.List(ctx, loginRecords, client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(labels.Set{iamv1beta1.UserReferenceLabel: user.Name})}); err != nil {
		return err
	}

	if len(loginRecords.Items) <= r.loginHistoryMaximumEntries {
		return nil
	}

	sort.Slice(loginRecords.Items, func(i, j int) bool {
		return loginRecords.Items[j].CreationTimestamp.After(loginRecords.Items[i].CreationTimestamp.Time)
	})
	oldEntries := loginRecords.Items[:len(loginRecords.Items)-r.loginHistoryMaximumEntries]
	for i := range oldEntries {
		if err := r.Delete(ctx, &oldEntries[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) userForLoginRecord(ctx context.Context, loginRecord *iamv1beta1.LoginRecord) (*iamv1beta1.User, error) {
	username, ok := loginRecord.Labels[iamv1beta1.UserReferenceLabel]
	if !ok || len(username) == 0 {
		klog.V(4).Info("login doesn't belong to any user")
		return nil, errors.NewNotFound(iamv1beta1.Resource(iamv1beta1.ResourcesSingularUser), username)
	}
	user := &iamv1beta1.User{}
	if err := r.Get(ctx, client.ObjectKey{Name: username}, user); err != nil {
		return nil, err
	}
	return user, nil
}
