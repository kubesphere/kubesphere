/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	clusterpredicate "kubesphere.io/kubesphere/pkg/controller/cluster/predicate"

	"k8s.io/apimachinery/pkg/api/errors"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
)

const (
	controllerName = "user"
	finalizer      = "finalizers.kubesphere.io/users"
)

var _ kscontroller.Controller = &Reconciler{}
var _ kscontroller.ClusterSelector = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a User object
type Reconciler struct {
	client.Client
	authenticationOptions *authentication.Options
	logger                logr.Logger
	recorder              record.EventRecorder
	clusterClient         clusterclient.Interface
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.authenticationOptions = mgr.AuthenticationOptions
	r.Client = mgr.GetClient()
	r.logger = mgr.GetLogger().WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.clusterClient = mgr.ClusterClient
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&iamv1beta1.User{}).
		Watches(&iamv1beta1.GlobalRoleBinding{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
			var result []reconcile.Request
			if username := object.GetLabels()[iamv1beta1.UserReferenceLabel]; username != "" {
				result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: username}})
			}
			return result
		}), builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectOld.GetLabels()[iamv1beta1.UserReferenceLabel] != e.ObjectNew.GetLabels()[iamv1beta1.UserReferenceLabel]
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return e.Object.GetLabels()[iamv1beta1.UserReferenceLabel] != ""
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return e.Object.GetLabels()[iamv1beta1.UserReferenceLabel] != ""
			},
		})).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(clusterpredicate.ClusterStatusChangedPredicate{}),
		).
		Complete(r)
}

func (r *Reconciler) mapper(ctx context.Context, o client.Object) []reconcile.Request {
	cluster := o.(*clusterv1alpha1.Cluster)
	var requests []reconcile.Request
	if !clusterutils.IsClusterReady(cluster) {
		return requests
	}
	users := &iamv1beta1.UserList{}
	if err := r.List(ctx, users); err != nil {
		r.logger.Error(err, "failed to list users")
		return requests
	}
	for _, user := range users.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: user.Name}})
	}
	return requests
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.logger.WithValues("user", req.NamespacedName)
	ctx = klog.NewContext(ctx, logger)

	user := &iamv1beta1.User{}
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if user.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(user, finalizer) {
			expected := user.DeepCopy()
			controllerutil.AddFinalizer(expected, finalizer)
			return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(user))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(user, finalizer) {
			if err := r.deleteRelatedResources(ctx, user); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete related resources: %s", err)
			}
			if err := r.deleteRelatedResourcesInMemberCluster(ctx, user); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete related resources: %s", err)
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(user, finalizer)
			if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if err := r.updateGlobalRoleAnnotation(ctx, user); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.encryptPassword(ctx, user); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileUserStatus(ctx, user); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.multiClusterSync(ctx, user); err != nil {
		return ctrl.Result{}, err
	}

	r.recorder.Event(user, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	// block user for AuthenticateRateLimiterDuration duration, after that put it back to the queue to unblock
	if user.Status.State == iamv1beta1.UserAuthLimitExceeded {
		return ctrl.Result{Requeue: true, RequeueAfter: r.authenticationOptions.AuthenticateRateLimiterDuration}, nil
	}

	return ctrl.Result{}, nil
}

// encryptPassword Encrypt and update the user password
func (r *Reconciler) encryptPassword(ctx context.Context, user *iamv1beta1.User) error {
	// password must be encrypted if not empty
	if user.Spec.EncryptedPassword != "" && !isEncrypted(user.Spec.EncryptedPassword) {
		encryptedPassword, err := encrypt(user.Spec.EncryptedPassword)
		if err != nil {
			return err
		}
		user.Spec.EncryptedPassword = encryptedPassword
		if user.Annotations == nil {
			user.Annotations = make(map[string]string)
		}
		user.Annotations[iamv1beta1.LastPasswordChangeTimeAnnotation] = time.Now().UTC().Format(time.RFC3339)
		// ensure plain text password won't be kept anywhere
		delete(user.Annotations, corev1.LastAppliedConfigAnnotation)
		if err = r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) deleteRelatedResources(ctx context.Context, user *iamv1beta1.User) error {
	if err := r.DeleteAllOf(ctx, &iamv1beta1.LoginRecord{}, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
		return err
	}
	if err := r.DeleteAllOf(ctx, &iamv1beta1.GlobalRoleBinding{}, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
		return err
	}
	if err := r.DeleteAllOf(ctx, &iamv1beta1.WorkspaceRoleBinding{}, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) deleteRelatedResourcesInMemberCluster(ctx context.Context, user *iamv1beta1.User) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %s", err)
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		// skip if the cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %s", err)
		}
		if err = clusterClient.DeleteAllOf(ctx, &iamv1beta1.ClusterRoleBinding{}, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
			return err
		}
		roleBindings := &iamv1beta1.RoleBindingList{}
		if err = clusterClient.List(ctx, roleBindings, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
			return err
		}
		for _, roleBinding := range roleBindings.Items {
			if err = clusterClient.Delete(ctx, &roleBinding); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}
		}
		if err = clusterClient.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: constants.KubeSphereNamespace, Name: fmt.Sprintf(kubeconfig.UserKubeConfigSecretNameFormat, user.Name)}}); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
	}
	if len(notReadyClusters) > 0 {
		err = fmt.Errorf("cluster not ready: %s", strings.Join(notReadyClusters, ","))
		klog.FromContext(ctx).Error(err, "failed to delete related resources")
		r.recorder.Event(user, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
		return err
	}
	return nil
}

// reconcileUserStatus updates the user status based on various conditions.
func (r *Reconciler) reconcileUserStatus(ctx context.Context, user *iamv1beta1.User) error {
	// skip status sync if the user is disabled
	if user.Status.State == iamv1beta1.UserDisabled {
		return nil
	}

	if user.Spec.EncryptedPassword == "" {
		if user.Labels[iamv1beta1.IdentifyProviderLabel] != "" {
			// mapped user from another identity provider always active until disabled
			if user.Status.State != iamv1beta1.UserActive {
				user.Status = iamv1beta1.UserStatus{
					State:              iamv1beta1.UserActive,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
					return err
				}
			}
		} else {
			// empty password is not allowed for normal user
			if user.Status.State != iamv1beta1.UserDisabled {
				user.Status = iamv1beta1.UserStatus{
					State:              iamv1beta1.UserDisabled,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
		// skip auth limit check
		return nil
	}

	// becomes active after password encrypted
	if user.Status.State == "" && isEncrypted(user.Spec.EncryptedPassword) {
		user.Status = iamv1beta1.UserStatus{
			State:              iamv1beta1.UserActive,
			LastTransitionTime: &metav1.Time{Time: time.Now()},
		}
		if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	// determine whether there is a requirement to unblock the user who has been blocked.
	if user.Status.State == iamv1beta1.UserAuthLimitExceeded {
		if user.Status.LastTransitionTime != nil &&
			user.Status.LastTransitionTime.Add(r.authenticationOptions.AuthenticateRateLimiterDuration).Before(time.Now()) {
			// unblock user
			user.Status = iamv1beta1.UserStatus{
				State:              iamv1beta1.UserActive,
				LastTransitionTime: &metav1.Time{Time: time.Now()},
			}
			if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
				return err
			}
			return nil
		}
	}

	records := &iamv1beta1.LoginRecordList{}
	if err := r.List(ctx, records, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
		return err
	}

	// count failed login attempts during last AuthenticateRateLimiterDuration
	now := time.Now()
	failedLoginAttempts := 0
	for _, loginRecord := range records.Items {
		afterStateTransition := user.Status.LastTransitionTime == nil || loginRecord.CreationTimestamp.After(user.Status.LastTransitionTime.Time)
		if !loginRecord.Spec.Success &&
			afterStateTransition &&
			loginRecord.CreationTimestamp.Add(r.authenticationOptions.AuthenticateRateLimiterDuration).After(now) {
			failedLoginAttempts++
		}
	}

	// block user if failed login attempts exceeds maximum tries setting
	if failedLoginAttempts >= r.authenticationOptions.AuthenticateRateLimiterMaxTries {
		user.Status = iamv1beta1.UserStatus{
			State:              iamv1beta1.UserAuthLimitExceeded,
			Reason:             fmt.Sprintf("Failed login attempts exceed %d in last %s", failedLoginAttempts, r.authenticationOptions.AuthenticateRateLimiterDuration),
			LastTransitionTime: &metav1.Time{Time: time.Now()},
		}
		if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) updateGlobalRoleAnnotation(ctx context.Context, user *iamv1beta1.User) error {
	globalRoles := &iamv1beta1.GlobalRoleBindingList{}
	if err := r.List(ctx, globalRoles, client.MatchingLabels{iamv1beta1.UserReferenceLabel: user.Name}); err != nil {
		return err
	}

	var globalRole string
	if len(globalRoles.Items) == 1 {
		globalRole = globalRoles.Items[0].RoleRef.Name
	} else if len(globalRoles.Items) == 0 {
		globalRole = ""
	} else {
		klog.Warningf("User %s has more than one global role bindings", user.Name)
		globalRole = user.Annotations[iamv1beta1.GlobalRoleAnnotation]
	}

	if globalRole != user.Annotations[iamv1beta1.GlobalRoleAnnotation] {
		if user.Annotations == nil {
			user.Annotations = make(map[string]string)
		}
		user.Annotations[iamv1beta1.GlobalRoleAnnotation] = globalRole
		if err := r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, user *iamv1beta1.User) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %s", err)
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		// skip if the cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		if err := r.syncKubeConfigSecret(ctx, cluster, user); err != nil {
			return fmt.Errorf("failed to sync user %s to cluster %s: %s", user.Name, cluster.Name, err)
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(user, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
	}
	return nil
}

func (r *Reconciler) syncKubeConfigSecret(ctx context.Context, cluster clusterv1alpha1.Cluster, user *iamv1beta1.User) error {
	clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %s", err)
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: constants.KubeSphereNamespace, Name: fmt.Sprintf(kubeconfig.UserKubeConfigSecretNameFormat, user.Name)}}
	op, err := controllerutil.CreateOrUpdate(ctx, clusterClient, secret, func() error {
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		secret.Labels[constants.UsernameLabelKey] = user.Name
		if secret.Type == "" {
			secret.Type = kubeconfig.SecretTypeKubeConfig
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update kubeconfig secret: %s", err)
	}

	r.logger.V(4).Info("kubeconfig secret successfully synced", "cluster", cluster.Name, "operation", op, "name", secret.Name)
	return nil
}

func encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// isEncrypted returns whether the given password is encrypted
func isEncrypted(password string) bool {
	// bcrypt.Cost returns the hashing cost used to create the given hashed
	cost, _ := bcrypt.Cost([]byte(password))
	// cost > 0 means the password has been encrypted
	return cost > 0
}
