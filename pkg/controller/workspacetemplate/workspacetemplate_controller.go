/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacetemplate

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/cluster/predicate"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate/utils"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

const (
	controllerName             = "workspacetemplate"
	workspaceTemplateFinalizer = "finalizers.workspacetemplate.kubesphere.io"
	orphanFinalizer            = "orphan.finalizers.kubesphere.io"
)

// Reconciler reconciles a WorkspaceRoleBinding object
type Reconciler struct {
	client.Client
	logger        logr.Logger
	recorder      record.EventRecorder
	clusterClient clusterclient.Interface
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.clusterClient = mgr.ClusterClient
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&tenantv1beta1.WorkspaceTemplate{}).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(predicate.ClusterStatusChangedPredicate{}),
		).
		Complete(r)
}

func (r *Reconciler) mapper(ctx context.Context, o client.Object) []reconcile.Request {
	cluster := o.(*clusterv1alpha1.Cluster)
	if !clusterutils.IsClusterReady(cluster) {
		return []reconcile.Request{}
	}
	workspaceTemplates := &tenantv1beta1.WorkspaceTemplateList{}
	if err := r.List(ctx, workspaceTemplates); err != nil {
		r.logger.Error(err, "failed to list workspace templates")
		return []reconcile.Request{}
	}
	var result []reconcile.Request
	for _, workspaceTemplate := range workspaceTemplates.Items {
		if utils.WorkspaceTemplateMatchTargetCluster(&workspaceTemplate, cluster) {
			result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: workspaceTemplate.Name}})
		}
	}
	return result
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("workspacetemplate", req.NamespacedName)
	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, req.NamespacedName, workspaceTemplate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ctx = klog.NewContext(ctx, logger)
	if workspaceTemplate.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(workspaceTemplate, workspaceTemplateFinalizer) {
			updated := workspaceTemplate.DeepCopy()
			controllerutil.AddFinalizer(updated, workspaceTemplateFinalizer)
			return ctrl.Result{}, r.Patch(ctx, updated, client.MergeFrom(workspaceTemplate))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(workspaceTemplate, workspaceTemplateFinalizer) ||
			controllerutil.ContainsFinalizer(workspaceTemplate, orphanFinalizer) {
			if err := r.reconcileDelete(ctx, workspaceTemplate); err != nil {
				return ctrl.Result{}, err
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(workspaceTemplate, workspaceTemplateFinalizer)
			controllerutil.RemoveFinalizer(workspaceTemplate, orphanFinalizer)
			if err := r.Update(ctx, workspaceTemplate); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}

	if err := r.initWorkspaceRoles(ctx, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.initManagerRoleBinding(ctx, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.multiClusterSync(ctx, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}

	r.recorder.Event(workspaceTemplate, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, workspaceTemplate *tenantv1beta1.WorkspaceTemplate) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %s", err)
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		// skip if cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		if err := r.syncWorkspaceTemplate(ctx, cluster, workspaceTemplate); err != nil {
			return fmt.Errorf("failed to sync workspace template %s to cluster %s: %s", workspaceTemplate.Name, cluster.Name, err)
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(workspaceTemplate, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
	}
	return nil
}

func (r *Reconciler) syncWorkspaceTemplate(ctx context.Context, cluster clusterv1alpha1.Cluster, workspaceTemplate *tenantv1beta1.WorkspaceTemplate) error {
	clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return err
	}
	if utils.WorkspaceTemplateMatchTargetCluster(workspaceTemplate, &cluster) {
		target := &tenantv1beta1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: workspaceTemplate.Name}}
		op, err := controllerutil.CreateOrUpdate(ctx, clusterClient, target, func() error {
			for k, v := range workspaceTemplate.Spec.Template.Labels {
				if target.Labels == nil {
					target.Labels = make(map[string]string)
				}
				target.Labels[k] = v
			}
			for k, v := range workspaceTemplate.Spec.Template.Annotations {
				if target.Annotations == nil {
					target.Annotations = make(map[string]string)
				}
				target.Annotations[k] = v
			}
			target.Spec = workspaceTemplate.Spec.Template.Spec
			return nil
		})
		if err != nil {
			return err
		}
		klog.FromContext(ctx).V(4).Info("workspace successfully synced", "operation", op)
	} else {
		orphan := metav1.DeletePropagationBackground
		err = clusterClient.Delete(ctx, &tenantv1beta1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: workspaceTemplate.Name}},
			&client.DeleteOptions{PropagationPolicy: &orphan})
		return client.IgnoreNotFound(err)
	}
	return nil
}

func (r *Reconciler) initWorkspaceRoles(ctx context.Context, workspaceTemplate *tenantv1beta1.WorkspaceTemplate) error {
	logger := klog.FromContext(ctx)
	var templates iamv1beta1.BuiltinRoleList
	// scope.iam.kubesphere.io/workspace: ""
	if err := r.List(ctx, &templates, client.MatchingLabels{iamv1beta1.ScopeLabel: iamv1beta1.ScopeWorkspace}); err != nil {
		return err
	}
	for _, template := range templates.Items {
		selector, err := metav1.LabelSelectorAsSelector(&template.TargetSelector)
		if err != nil {
			logger.V(4).Error(err, "failed to pares target selector", "template", template.Name)
			continue
		}
		if !selector.Matches(labels.Set(workspaceTemplate.Labels)) {
			continue
		}
		var builtinWorkspaceRole iamv1beta1.WorkspaceRole
		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(template.Role.Raw), 1024).Decode(&builtinWorkspaceRole); err == nil &&
			builtinWorkspaceRole.Kind == iamv1beta1.ResourceKindWorkspaceRole {
			target := &iamv1beta1.WorkspaceRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", workspaceTemplate.Name, builtinWorkspaceRole.Name),
				},
			}
			op, err := controllerutil.CreateOrUpdate(ctx, r.Client, target, func() error {
				target.Labels = builtinWorkspaceRole.Labels
				if target.Labels == nil {
					target.Labels = make(map[string]string)
				}
				target.Labels[tenantv1beta1.WorkspaceLabel] = workspaceTemplate.Name
				target.Annotations = builtinWorkspaceRole.Annotations
				target.AggregationRoleTemplates = builtinWorkspaceRole.AggregationRoleTemplates
				target.Rules = builtinWorkspaceRole.Rules
				return nil
			})
			if err != nil {
				return err
			}
			logger.V(4).Info("builtin workspace role successfully updated", "operation", op, "name", target.Name)
		} else if err != nil {
			logger.Error(err, "invalid builtin workspace role found", "name", template.Name)
		}
	}
	return nil
}

func (r *Reconciler) initManagerRoleBinding(ctx context.Context, workspaceTemplate *tenantv1beta1.WorkspaceTemplate) error {
	manager := workspaceTemplate.Spec.Template.Spec.Manager
	if manager == "" {
		return nil
	}
	workspaceAdminRoleName := fmt.Sprintf("%s-admin", workspaceTemplate.Name)
	existWorkspaceRoleBinding := &iamv1beta1.WorkspaceRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: workspaceAdminRoleName}}
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, existWorkspaceRoleBinding, func() error {
		existWorkspaceRoleBinding.Labels = map[string]string{
			tenantv1beta1.WorkspaceLabel:  workspaceTemplate.Name,
			iamv1beta1.UserReferenceLabel: manager,
			iamv1beta1.RoleReferenceLabel: workspaceAdminRoleName,
		}

		existWorkspaceRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: iamv1beta1.SchemeGroupVersion.Group,
			Kind:     iamv1beta1.ResourceKindWorkspaceRole,
			Name:     workspaceAdminRoleName,
		}
		existWorkspaceRoleBinding.Subjects = []rbacv1.Subject{
			{
				Name:     manager,
				Kind:     iamv1beta1.ResourceKindUser,
				APIGroup: iamv1beta1.SchemeGroupVersion.Group,
			},
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileDelete(ctx context.Context, workspaceTemplate *tenantv1beta1.WorkspaceTemplate) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
	if err != nil {
		return err
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		// skip if cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}

		if controllerutil.ContainsFinalizer(workspaceTemplate, orphanFinalizer) {
			orphan := metav1.DeletePropagationOrphan
			err = clusterClient.Delete(ctx, &tenantv1beta1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: workspaceTemplate.Name}}, &client.DeleteOptions{PropagationPolicy: &orphan})
		} else {
			err = clusterClient.Delete(ctx, &tenantv1beta1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: workspaceTemplate.Name}}, &client.DeleteOptions{})
		}

		if !errors.IsNotFound(err) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(workspaceTemplate, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
		return err
	}
	return nil
}
