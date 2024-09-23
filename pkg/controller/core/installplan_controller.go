/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/exp/slices"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	helmrelease "helm.sh/helm/v3/pkg/release"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"kubesphere.io/utils/helm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	clusterpredicate "kubesphere.io/kubesphere/pkg/controller/cluster/predicate"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/hashutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	installPlanController           = "installplan"
	installPlanProtection           = "kubesphere.io/installplan-protection"
	systemWorkspace                 = "system-workspace"
	agentReleaseFormat              = "%s-agent"
	defaultRoleFormat               = "kubesphere:%s:helm-executor"
	defaultRoleBindingFormat        = defaultRoleFormat
	defaultClusterRoleFormat        = "kubesphere:%s:helm-executor"
	permissionDefinitionFile        = "permissions.yaml"
	defaultClusterRoleBindingFormat = defaultClusterRoleFormat
	tagAgent                        = "agent"
	tagExtension                    = "extension"

	upgradeSuccessful       = "UpgradeSuccessful"
	upgradeFailed           = "UpgradeFailed"
	installSuccessful       = "InstallSuccessful"
	installFailed           = "InstallFailed"
	initialized             = "Initialized"
	uninstallSuccessful     = "UninstallSuccessful"
	uninstallFailed         = "UninstallFailed"
	relatedResourceNotReady = "RelatedResourceNotReady"
	relatedResourceReady    = "RelatedResourceReady"

	typeHelmRelease = "helm.sh/release.v1"

	globalExtensionIngressClassName    = "global.extension.ingress.ingressClassName"
	globalExtensionIngressDomainSuffix = "global.extension.ingress.domainSuffix"
	globalExtensionIngressHTTPPort     = "global.extension.ingress.httpPort"
	globalExtensionIngressHTTPSPort    = "global.extension.ingress.httpsPort"
	globalNodeSelector                 = "global.nodeSelector"
	globalImageRegistry                = "global.imageRegistry"
	globalClusterName                  = "global.clusterInfo.name"
	globalClusterRole                  = "global.clusterInfo.role"
	globalPortalURL                    = "global.portal.url"
)

var _ kscontroller.Controller = &InstallPlanReconciler{}
var _ reconcile.Reconciler = &InstallPlanReconciler{}

func (r *InstallPlanReconciler) Name() string {
	return installPlanController
}

func (r *InstallPlanReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

// InstallPlanReconciler reconciles a InstallPlan object.
type InstallPlanReconciler struct {
	client.Client
	recorder            record.EventRecorder
	logger              logr.Logger
	PortalURL           string
	HelmExecutorOptions *options.HelmExecutorOptions
	ExtensionOptions    *options.ExtensionOptions
	hostResetConfig     *rest.Config
	clusterClientSet    clusterclient.Interface
}

func (r *InstallPlanReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	if mgr.AuthenticationOptions != nil && mgr.Options.AuthenticationOptions.Issuer != nil {
		r.PortalURL = mgr.Options.AuthenticationOptions.Issuer.URL
	}
	r.HelmExecutorOptions = mgr.HelmExecutorOptions
	r.ExtensionOptions = mgr.ExtensionOptions
	r.hostResetConfig = mgr.K8sClient.Config()
	r.Client = mgr.GetClient()
	r.logger = mgr.GetLogger().WithName(installPlanController)
	r.recorder = mgr.GetEventRecorderFor(installPlanController)
	r.clusterClientSet = mgr.ClusterClient

	if r.HelmExecutorOptions == nil || r.HelmExecutorOptions.Image == "" {
		return fmt.Errorf("helm executor image is not specified")
	}

	labelSelector, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      corev1alpha1.ExtensionReferenceLabel,
			Operator: metav1.LabelSelectorOpExists,
		}}})
	if err != nil {
		return fmt.Errorf("failed to create label selector predicate: %s", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(installPlanController).
		For(&corev1alpha1.InstallPlan{}).
		Watches(
			&batchv1.Job{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, h client.Object) []reconcile.Request {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name: h.GetLabels()[corev1alpha1.ExtensionReferenceLabel],
						}}}
				}),
			builder.WithPredicates(predicate.And(labelSelector, predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					if e.ObjectNew.GetLabels()[corev1alpha1.ExtensionReferenceLabel] == "" {
						return false
					}
					oldJob := e.ObjectOld.(*batchv1.Job)
					newJob := e.ObjectNew.(*batchv1.Job)
					return !reflect.DeepEqual(oldJob.Status, newJob.Status)
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
			})),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, h client.Object) []reconcile.Request {
					releaseName := h.GetLabels()["name"]
					owner := h.GetLabels()["owner"]

					var result []reconcile.Request
					if releaseName != "" && owner == "helm" {
						result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{
							Name: releaseName,
						}})
					}
					return result
				}),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return e.ObjectNew.(*corev1.Secret).Type == typeHelmRelease
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return e.Object.(*corev1.Secret).Type == typeHelmRelease
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return e.Object.(*corev1.Secret).Type == typeHelmRelease
				},
			}),
		).
		Watches(
			&corev1alpha1.ExtensionVersion{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, h client.Object) []reconcile.Request {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name: h.GetLabels()[corev1alpha1.ExtensionReferenceLabel],
						}}}
				}),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return false
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
			}),
		).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(clusterpredicate.ClusterStatusChangedPredicate{}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}

func (r *InstallPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("installplan", req.Name)
	ctx = klog.NewContext(ctx, logger)

	plan := &corev1alpha1.InstallPlan{}
	if err := r.Get(ctx, req.NamespacedName, plan); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	executor, err := r.newExecutor(plan)
	if err != nil {
		logger.Error(err, "failed to create executor")
		return ctrl.Result{}, fmt.Errorf("failed to create executor: %v", err)
	}

	ctx = context.WithValue(ctx, contextKeyExecutor{}, executor)

	// fixed kubeconfig
	if kubeConfig, err := clusterutils.BuildKubeconfigFromRestConfig(r.hostResetConfig); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to build host cluster kubeconfig: %v", err)
	} else {
		ctx = context.WithValue(ctx, contextKeyHostKubeConfig{}, kubeConfig)
	}

	extensionVersion := &corev1alpha1.ExtensionVersion{}
	extensionVersionName := fmt.Sprintf("%s-%s", plan.Spec.Extension.Name, plan.Spec.Extension.Version)
	if err = r.Get(ctx, types.NamespacedName{Name: extensionVersionName}, extensionVersion); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("extension version not found", "name", extensionVersionName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get extension version: %v", err)
	}
	ctx = context.WithValue(ctx, contextKeyExtensionVersion{}, extensionVersion)

	if !plan.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, plan)
	}

	if !controllerutil.ContainsFinalizer(plan, installPlanProtection) {
		expected := plan.DeepCopy()
		controllerutil.AddFinalizer(expected, installPlanProtection)
		return ctrl.Result{Requeue: true}, r.Patch(ctx, expected, client.MergeFrom(plan))
	}

	targetNamespace := extensionVersion.Spec.Namespace
	if targetNamespace == "" {
		targetNamespace = fmt.Sprintf("extension-%s", plan.Spec.Extension.Name)
	}

	if plan.Status.TargetNamespace != targetNamespace {
		plan.Status.TargetNamespace = targetNamespace
		return ctrl.Result{Requeue: true}, r.updateInstallPlan(ctx, plan)
	}

	if err := r.syncInstallPlanStatus(ctx, plan); err != nil {
		logger.Error(err, "failed to sync installplan status")
		return ctrl.Result{}, fmt.Errorf("failed to sync installplan status: %v", err)
	}

	// Multi-cluster installation
	if plan.Spec.ClusterScheduling != nil {
		if err := r.syncClusterSchedulingStatus(ctx, plan); err != nil {
			logger.Error(err, "failed to sync scheduling status")
			return ctrl.Result{}, fmt.Errorf("failed to sync scheduling status: %v", err)
		}
	}

	logger.V(4).Info("Successfully synced")
	return ctrl.Result{}, nil
}

// reconcileDelete delete the helm release involved and remove finalizer from installplan.
func (r *InstallPlanReconciler) reconcileDelete(ctx context.Context, plan *corev1alpha1.InstallPlan) (ctrl.Result, error) {
	// It has not been installed correctly.
	if plan.Status.ReleaseName == "" {
		if err := r.postRemove(ctx, plan); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %v", err)
		}
		return ctrl.Result{}, nil
	}

	executor := ctx.Value(contextKeyExecutor{}).(helm.Executor)

	if len(plan.Status.ClusterSchedulingStatuses) > 0 {
		for clusterName := range plan.Status.ClusterSchedulingStatuses {
			if err := r.uninstallClusterAgent(ctx, plan, clusterName); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	hostKubeConfig := ctx.Value(contextKeyHostKubeConfig{}).([]byte)

	opts := []helm.HelmOption{
		helm.SetKubeconfig(hostKubeConfig),
		helm.SetNamespace(plan.Status.TargetNamespace),
		helm.SetTimeout(r.HelmExecutorOptions.Timeout),
	}

	if _, ok := plan.Annotations[corev1alpha1.ForceDeleteAnnotation]; ok {
		if err := executor.ForceDelete(ctx, plan.Status.ReleaseName, opts...); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to force delete helm release: %v", err)
		}
		if err := r.postRemove(ctx, plan); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %v", err)
		}
	}

	// Check if the target helm release exists.
	// If it does, there is no need to execute the installation process again.
	_, err := executor.Get(ctx, plan.Status.ReleaseName, opts...)
	if err != nil {
		if isReleaseNotFoundError(err) {
			return ctrl.Result{}, r.postRemove(ctx, plan)
		}
		klog.FromContext(ctx).Error(err, "failed to get helm release status")
		return ctrl.Result{}, fmt.Errorf("failed to get helm release status: %v", err)
	}

	if err := r.syncInstallationStatus(ctx, hostKubeConfig, plan.Status.TargetNamespace, plan.Spec.Extension.Name, &plan.Status.InstallationStatus); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync installation status: %v", err)
	}

	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update install plan: %v", err)
	}

	installationStatus := plan.Status.InstallationStatus
	if installationStatus.State != corev1alpha1.StateUninstalling &&
		installationStatus.State != corev1alpha1.StateUninstallFailed &&
		installationStatus.State != corev1alpha1.StateUninstalled {

		extensionVersion, ok := ctx.Value(contextKeyExtensionVersion{}).(*corev1alpha1.ExtensionVersion)
		if !ok {
			return ctrl.Result{}, fmt.Errorf("failed to get extension version from context")
		}

		helmOptions := []helm.HelmOption{
			helm.SetKubeconfig(hostKubeConfig),
			helm.SetNamespace(plan.Status.TargetNamespace),
			helm.SetTimeout(r.HelmExecutorOptions.Timeout),
			helm.SetHookImage(r.getHookImageForUninstall(extensionVersion)),
		}

		jobName, err := executor.Uninstall(ctx, plan.Status.ReleaseName, helmOptions...)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to uninstall helm release: %v", err)
		}
		installationStatus.JobName = jobName
		updateStateAndConditions(&installationStatus, corev1alpha1.StateUninstalling, "", time.Now())
		plan.Status.InstallationStatus = installationStatus
		if err := r.updateInstallPlan(ctx, plan); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *InstallPlanReconciler) getHookImageForUninstall(extensionVersion *corev1alpha1.ExtensionVersion) string {
	hookImage := extensionVersion.Annotations[corev1alpha1.ExecutorHookImageAnnotation]
	if extensionVersion.Annotations[corev1alpha1.ExecutorUninstallHookImageAnnotation] != "" {
		hookImage = extensionVersion.Annotations[corev1alpha1.ExecutorUninstallHookImageAnnotation]
	}
	if r.ExtensionOptions.ImageRegistry != "" && hookImage != "" {
		hookImage = path.Join(r.ExtensionOptions.ImageRegistry, hookImage)
	}
	return hookImage
}

func latestJobCondition(job *batchv1.Job) batchv1.JobCondition {
	condition := batchv1.JobCondition{}
	if job == nil {
		return condition
	}

	jobConditions := job.Status.Conditions
	sort.Slice(jobConditions, func(i, j int) bool {
		return jobConditions[i].LastTransitionTime.After(jobConditions[j].LastTransitionTime.Time)
	})
	if len(job.Status.Conditions) > 0 {
		return jobConditions[0]
	}
	return condition
}

func (r *InstallPlanReconciler) loadChartData(ctx context.Context) ([]byte, string, error) {
	extensionVersion, ok := ctx.Value(contextKeyExtensionVersion{}).(*corev1alpha1.ExtensionVersion)
	if !ok {
		return nil, "", fmt.Errorf("failed to get extension version from context")
	}

	// load chart data from
	if extensionVersion.Spec.ChartDataRef != nil {
		configMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: extensionVersion.Spec.ChartDataRef.Namespace, Name: extensionVersion.Spec.ChartDataRef.Name}, configMap); err != nil {
			return nil, "", err
		}
		data := configMap.BinaryData[extensionVersion.Spec.ChartDataRef.Key]
		if data != nil {
			return data, "", nil
		}
		return nil, "", fmt.Errorf("binary data not found")
	}

	chartURL, err := url.Parse(extensionVersion.Spec.ChartURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse chart url: %v", err)
	}

	var caBundle string
	repo := &corev1alpha1.Repository{}
	if extensionVersion.Spec.Repository != "" {
		if err := r.Get(ctx, types.NamespacedName{Name: extensionVersion.Spec.Repository}, repo); err != nil {
			return nil, "", fmt.Errorf("failed to get repository: %v", err)
		}
		caBundle = repo.Spec.CABundle
	}

	var chartGetter getter.Getter
	switch chartURL.Scheme {
	case registry.OCIScheme:
		opts := make([]getter.Option, 0)
		if extensionVersion.Spec.Repository != "" {
			opts = append(opts, getter.WithInsecureSkipVerifyTLS(repo.Spec.Insecure))
		}
		if repo.Spec.BasicAuth != nil {
			opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewOCIGetter(opts...)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create chart getter: %v", err)
		}
	case "http", "https":
		opts := make([]getter.Option, 0)
		if chartURL.Scheme == "https" && extensionVersion.Spec.Repository != "" {
			opts = append(opts, getter.WithInsecureSkipVerifyTLS(repo.Spec.Insecure))
		}
		if repo.Spec.CABundle != "" {
			caFile, err := storeCAFile(repo.Spec.CABundle, repo.Name)
			if err != nil {
				return nil, "", fmt.Errorf("failed to store CABundle to local file: %s", err)
			}
			opts = append(opts, getter.WithTLSClientConfig("", "", caFile))
		}
		if chartURL.Scheme == "https" {
			opts = append(opts, getter.WithInsecureSkipVerifyTLS(repo.Spec.Insecure))
		}
		if repo.Spec.BasicAuth != nil {
			opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewHTTPGetter(opts...)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create chart getter: %v", err)
		}
	default:
		return nil, "", fmt.Errorf("unsupported scheme: %s", chartURL.Scheme)
	}

	buffer, err := chartGetter.Get(chartURL.String())
	if err != nil {
		return nil, "", fmt.Errorf("failed to get chart data: %v", err)
	}

	data, err := io.ReadAll(buffer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read chart data: %v", err)
	}

	return data, caBundle, nil
}

func updateState(status *corev1alpha1.InstallationStatus, state string, time time.Time) bool {
	var lastState corev1alpha1.InstallPlanState
	if len(status.StateHistory) > 0 {
		lastState = status.StateHistory[len(status.StateHistory)-1]
	}

	if lastState.State == state {
		return false
	}

	if time.Before(lastState.LastTransitionTime.Time) {
		return false
	}

	status.State = state

	newState := corev1alpha1.InstallPlanState{
		LastTransitionTime: metav1.NewTime(time),
		State:              state,
	}

	if status.StateHistory == nil {
		status.StateHistory = make([]corev1alpha1.InstallPlanState, 0)
	}

	status.StateHistory = append(status.StateHistory, newState)

	sort.Slice(status.StateHistory, func(i, j int) bool {
		return status.StateHistory[i].LastTransitionTime.Before(&status.StateHistory[j].LastTransitionTime)
	})

	if len(status.StateHistory) > corev1alpha1.MaxStateNum {
		status.StateHistory = status.StateHistory[len(status.StateHistory)-corev1alpha1.MaxStateNum:]
	}

	return true
}

func updateCondition(status *corev1alpha1.InstallationStatus, conditionType, reason, message string, conditionStatus metav1.ConditionStatus, time time.Time) {
	conditions := []metav1.Condition{
		{
			Type:               conditionType,
			Reason:             reason,
			Status:             conditionStatus,
			LastTransitionTime: metav1.NewTime(time),
			Message:            message,
		},
	}

	if len(status.Conditions) == 0 {
		status.Conditions = conditions
		return
	}

	for _, c := range status.Conditions {
		if c.Type != conditionType {
			conditions = append(conditions, c)
		}
	}

	status.Conditions = conditions
}

func (r *InstallPlanReconciler) postRemove(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	message := fmt.Sprintf("The extension %s has been successfully uninstalled.", plan.Spec.Extension.Name)
	updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StateUninstalled, message, time.Now())
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return err
	}
	// Remove the finalizer from the installplan and update it.
	if controllerutil.RemoveFinalizer(plan, installPlanProtection) {
		return r.Update(ctx, plan)
	}
	return nil
}

func (r *InstallPlanReconciler) syncExtendedAPIStatus(ctx context.Context, clusterClient client.Client, plan *corev1alpha1.InstallPlan) error {
	if err := syncJSBundleStatus(ctx, clusterClient, plan); err != nil {
		return err
	}

	if err := syncAPIServiceStatus(ctx, clusterClient, plan); err != nil {
		return err
	}

	if err := syncReverseProxyStatus(ctx, clusterClient, plan); err != nil {
		return err
	}

	if err := syncExtensionEntryStatus(ctx, clusterClient, plan); err != nil {
		return err
	}
	return nil
}

func (r *InstallPlanReconciler) updateInstallPlan(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		target := &corev1alpha1.InstallPlan{}
		if err := r.Get(ctx, client.ObjectKey{Name: plan.Name}, target); err != nil {
			return client.IgnoreNotFound(err)
		}
		if !reflect.DeepEqual(target.Labels, plan.Labels) ||
			!reflect.DeepEqual(target.Annotations, plan.Annotations) ||
			!reflect.DeepEqual(target.Status, plan.Status) {
			if r.logger.V(4).Enabled() {
				r.logger.Info("installplan status changed", "name", plan.Name, "status", plan.Status)
			}
			target.Labels = plan.Labels
			target.Annotations = plan.Annotations
			target.Status = plan.Status
			if err := r.Update(ctx, target); err != nil {
				return fmt.Errorf("failed to update install plan: %v", err)
			}
			target.DeepCopyInto(plan)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update install plan %s: %s", plan.Name, err)
	}
	if err := r.syncExtensionStatus(ctx, plan); err != nil {
		return fmt.Errorf("failed to sync extension %s status: %s", plan.Spec.Extension.Name, err)
	}
	return nil
}

func createNamespaceIfNotExists(ctx context.Context, client client.Client, namespace, extensionName string) error {
	var ns corev1.Namespace
	if err := client.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if errors.IsNotFound(err) {
			ns = corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
					Labels: map[string]string{
						tenantv1beta1.WorkspaceLabel:         systemWorkspace,
						corev1alpha1.ExtensionReferenceLabel: extensionName,
						constants.KubeSphereManagedLabel:     "true",
					},
				},
			}
			if err := client.Create(ctx, &ns); err != nil {
				return fmt.Errorf("failed to create namespace: %v", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get namespace: %v", err)
	}
	return nil
}

func createOrUpdateRole(ctx context.Context, client client.Client, namespace, extensionName string, rules []rbacv1.PolicyRule) error {
	roleName := fmt.Sprintf(defaultRoleFormat, extensionName)

	var defaultRules = []rbacv1.PolicyRule{
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"", "apps", "batch", "policy", "networking.k8s.io", "autoscaling", "metrics.k8s.io"},
			Resources: []string{"*"},
		},
	}

	role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: roleName}}
	op, err := controllerutil.CreateOrUpdate(ctx, client, role, func() error {
		role.Labels = map[string]string{corev1alpha1.ExtensionReferenceLabel: extensionName}
		role.Rules = rules
		_, uncoveredRules := rbachelper.Covers(role.Rules, defaultRules)
		if len(uncoveredRules) > 0 {
			role.Rules = append(role.Rules, uncoveredRules...)
		}
		return nil
	})

	if err != nil {
		return err
	}

	klog.V(4).Infof("role %s in namespace %s %s", role.Name, role.Namespace, op)

	return nil
}

func createOrUpdateRoleBinding(ctx context.Context, client client.Client, namespace, extensionName string, sa rbacv1.Subject) error {
	roleName := fmt.Sprintf(defaultRoleFormat, extensionName)
	roleBindingName := fmt.Sprintf(defaultRoleBindingFormat, extensionName)

	roleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: roleBindingName}}

	op, err := controllerutil.CreateOrUpdate(ctx, client, roleBinding, func() error {
		roleBinding.Labels = map[string]string{corev1alpha1.ExtensionReferenceLabel: extensionName}
		roleBinding.RoleRef = rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: roleName}
		roleBinding.Subjects = []rbacv1.Subject{sa}
		return nil
	})

	if err != nil {
		return err
	}

	klog.V(4).Infof("role binding %s in namespace %s %s", roleBinding.Name, roleBinding.Namespace, op)

	return nil
}

func initTargetNamespace(ctx context.Context, client client.Client, namespace, extensionName string, clusterRole rbacv1.ClusterRole, role rbacv1.Role) error {
	if err := createNamespaceIfNotExists(ctx, client, namespace, extensionName); err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}
	sa := rbacv1.Subject{
		Kind:      rbacv1.ServiceAccountKind,
		Name:      fmt.Sprintf("helm-executor.%s", extensionName),
		Namespace: namespace,
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := createOrUpdateRole(ctx, client, namespace, extensionName, role.Rules); err != nil {
			return err
		}
		if err := createOrUpdateRoleBinding(ctx, client, namespace, extensionName, sa); err != nil {
			return err
		}
		if err := createOrUpdateClusterRole(ctx, client, extensionName, clusterRole.Rules); err != nil {
			return err
		}
		if err := createOrUpdateClusterRoleBinding(ctx, client, extensionName, sa); err != nil {
			return err
		}
		return nil
	})
}

func createOrUpdateClusterRole(ctx context.Context, client client.Client, extensionName string, rules []rbacv1.PolicyRule) error {
	clusterRoleName := fmt.Sprintf(defaultClusterRoleFormat, extensionName)
	clusterRole := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleName}}

	op, err := controllerutil.CreateOrUpdate(ctx, client, clusterRole, func() error {
		clusterRole.Labels = map[string]string{corev1alpha1.ExtensionReferenceLabel: extensionName}
		clusterRole.Rules = rules
		return nil
	})

	if err != nil {
		return err
	}

	klog.V(4).Infof("cluster role %s %s", clusterRole.Name, op)

	return nil
}

func createOrUpdateClusterRoleBinding(ctx context.Context, client client.Client, extensionName string, sa rbacv1.Subject) error {
	clusterRoleName := fmt.Sprintf(defaultClusterRoleFormat, extensionName)
	clusterRoleBindingName := fmt.Sprintf(defaultClusterRoleBindingFormat, extensionName)
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleBindingName}}

	op, err := controllerutil.CreateOrUpdate(ctx, client, clusterRoleBinding, func() error {
		if clusterRoleBinding.RoleRef.Name != "" && clusterRoleBinding.RoleRef.Name != clusterRoleName {
			return fmt.Errorf("conflict cluster role binding found: %s", clusterRoleBindingName)
		}
		clusterRoleBinding.Labels = map[string]string{corev1alpha1.ExtensionReferenceLabel: extensionName}
		clusterRoleBinding.RoleRef = rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: clusterRoleName}
		clusterRoleBinding.Subjects = []rbacv1.Subject{sa}
		return nil
	})

	if err != nil {
		return err
	}

	klog.V(4).Infof("cluster role binding %s %s", clusterRoleBinding.Name, op)

	return nil
}

func syncJSBundleStatus(ctx context.Context, clusterClient client.Client, plan *corev1alpha1.InstallPlan) error {
	jsBundles := &extensionsv1alpha1.JSBundleList{}
	if err := clusterClient.List(ctx, jsBundles, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}); err != nil {
		return err
	}
	for _, jsBundle := range jsBundles.Items {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := clusterClient.Get(ctx, types.NamespacedName{Name: jsBundle.Name}, &jsBundle); err != nil {
				return err
			}
			// TODO unavailable state should be considered
			expected := jsBundle.DeepCopy()
			if plan.Spec.Enabled {
				expected.Status.State = extensionsv1alpha1.StateAvailable
			} else {
				expected.Status.State = extensionsv1alpha1.StateDisabled
			}
			if expected.Status.State != jsBundle.Status.State {
				if err := clusterClient.Update(ctx, expected); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update js bundle status: %v", err)
		}
	}

	return nil
}

func syncAPIServiceStatus(ctx context.Context, clusterClient client.Client, plan *corev1alpha1.InstallPlan) error {
	apiServices := &extensionsv1alpha1.APIServiceList{}
	if err := clusterClient.List(ctx, apiServices, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}); err != nil {
		return err
	}

	for _, apiService := range apiServices.Items {

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := clusterClient.Get(ctx, types.NamespacedName{Name: apiService.Name}, &apiService); err != nil {
				return err
			}
			// TODO unavailable state should be considered
			expected := apiService.DeepCopy()
			if plan.Spec.Enabled {
				expected.Status.State = extensionsv1alpha1.StateAvailable
			} else {
				expected.Status.State = extensionsv1alpha1.StateDisabled
			}
			if expected.Status.State != apiService.Status.State {
				if err := clusterClient.Update(ctx, expected); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to update api service status: %v", err)
		}
	}

	return nil
}

func syncReverseProxyStatus(ctx context.Context, clusterClient client.Client, plan *corev1alpha1.InstallPlan) error {
	reverseProxies := &extensionsv1alpha1.ReverseProxyList{}
	if err := clusterClient.List(ctx, reverseProxies, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}); err != nil {
		return err
	}

	for _, reverseProxy := range reverseProxies.Items {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := clusterClient.Get(ctx, types.NamespacedName{Name: reverseProxy.Name}, &reverseProxy); err != nil {
				return err
			}
			expected := reverseProxy.DeepCopy()
			if plan.Spec.Enabled {
				expected.Status.State = extensionsv1alpha1.StateAvailable
			} else {
				expected.Status.State = extensionsv1alpha1.StateDisabled
			}
			if expected.Status.State != reverseProxy.Status.State {
				if err := clusterClient.Update(ctx, expected); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to update reverse proxy status: %v", err)
		}
	}

	return nil
}

func syncExtensionEntryStatus(ctx context.Context, clusterClient client.Client, plan *corev1alpha1.InstallPlan) error {
	extensionEntries := &extensionsv1alpha1.ExtensionEntryList{}
	if err := clusterClient.List(ctx, extensionEntries, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}); err != nil {
		return err
	}

	for _, extensionEntry := range extensionEntries.Items {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := clusterClient.Get(ctx, types.NamespacedName{Name: extensionEntry.Name}, &extensionEntry); err != nil {
				return err
			}
			expected := extensionEntry.DeepCopy()
			if plan.Spec.Enabled {
				expected.Status.State = extensionsv1alpha1.StateAvailable
			} else {
				expected.Status.State = extensionsv1alpha1.StateDisabled
			}
			if expected.Status.State != extensionEntry.Status.State {
				if err := clusterClient.Update(ctx, expected); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to update extension entry status: %v", err)
		}
	}

	return nil
}

func (r *InstallPlanReconciler) syncExtensionStatus(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		extension := &corev1alpha1.Extension{}
		if err := r.Get(ctx, types.NamespacedName{Name: plan.Spec.Extension.Name}, extension); err != nil {
			return client.IgnoreNotFound(err)
		}

		updated := extension.DeepCopy()
		updated.Status.State = plan.Status.State
		updated.Status.Enabled = plan.Status.Enabled
		updated.Status.Conditions = plan.Status.Conditions
		updated.Status.PlannedInstallVersion = plan.Spec.Extension.Version
		updated.Status.InstalledVersion = plan.Status.Version
		updated.Status.ClusterSchedulingStatuses = plan.Status.ClusterSchedulingStatuses

		if plan.Status.State == corev1alpha1.StateUninstalled {
			updated.Status.State = ""
			updated.Status.Enabled = false
			updated.Status.Conditions = nil
			updated.Status.PlannedInstallVersion = ""
			updated.Status.InstalledVersion = ""
			updated.Status.ClusterSchedulingStatuses = nil
		}

		if !reflect.DeepEqual(extension.Status, updated.Status) {
			if err := r.Update(ctx, updated); err != nil {
				return err
			}
			if r.logger.V(4).Enabled() {
				r.logger.Info("extension status changed", "name", extension.Name, "status", updated.Status)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update extension status: %s", err)
	}

	return nil
}

func (r *InstallPlanReconciler) syncClusterSchedulingStatus(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	logger := klog.FromContext(ctx)
	if plan.Status.State != corev1alpha1.StateDeployed {
		return nil
	}
	// extension is already installed
	var targetClusters []clusterv1alpha1.Cluster
	if len(plan.Spec.ClusterScheduling.Placement.Clusters) > 0 {
		for _, target := range plan.Spec.ClusterScheduling.Placement.Clusters {
			var cluster clusterv1alpha1.Cluster
			if err := r.Get(ctx, types.NamespacedName{Name: target}, &cluster); err != nil {
				if errors.IsNotFound(err) {
					logger.V(4).Info("cluster not found")
					continue
				}
				return err
			}
			targetClusters = append(targetClusters, cluster)
		}
	} else if plan.Spec.ClusterScheduling.Placement.ClusterSelector != nil {
		clusterList := &clusterv1alpha1.ClusterList{}
		selector, err := metav1.LabelSelectorAsSelector(plan.Spec.ClusterScheduling.Placement.ClusterSelector)
		if err != nil {
			return err
		}
		if err := r.List(ctx, clusterList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
			return err
		}
		targetClusters = clusterList.Items
	}

	for _, cluster := range targetClusters {
		if err := r.syncClusterAgentStatus(ctx, plan, &cluster); err != nil {
			return err
		}
	}

	for clusterName := range plan.Status.ClusterSchedulingStatuses {
		if !hasCluster(targetClusters, clusterName) {
			if err := r.uninstallClusterAgent(ctx, plan, clusterName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *InstallPlanReconciler) cleanupOutdatedJobsAndConfigMaps(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	jobNames := make([]string, 0)
	if plan.Status.JobName != "" {
		jobNames = append(jobNames, plan.Status.JobName)
	}
	for _, s := range plan.Status.ClusterSchedulingStatuses {
		if s.JobName != "" {
			jobNames = append(jobNames, s.JobName)
		}
	}
	if len(jobNames) == 0 {
		return nil
	}

	selector, _ := labels.Parse(fmt.Sprintf(
		"%s=%s,%s=%s,name notin (%s)",
		constants.KubeSphereManagedLabel, "true",
		corev1alpha1.ExtensionReferenceLabel, plan.Spec.Extension.Name,
		strings.Join(jobNames, ","),
	))

	deletePolicy := metav1.DeletePropagationBackground
	if err := r.DeleteAllOf(ctx, &batchv1.Job{}, client.InNamespace(plan.Status.TargetNamespace), &client.DeleteAllOfOptions{
		ListOptions:   client.ListOptions{LabelSelector: selector},
		DeleteOptions: client.DeleteOptions{PropagationPolicy: &deletePolicy},
	}); err != nil {
		return fmt.Errorf("failed to delete related helm executor jobs: %s", err)
	}
	if err := r.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(plan.Status.TargetNamespace), &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{LabelSelector: selector},
	}); err != nil {
		return fmt.Errorf("failed to delete related helm executor configmaps: %s", err)
	}
	return nil
}

// syncInstallPlanStatus syncs the installation status of an extension.
func (r *InstallPlanReconciler) syncInstallPlanStatus(ctx context.Context, plan *corev1alpha1.InstallPlan) error {
	releaseName := plan.Spec.Extension.Name
	targetNamespace := plan.Status.TargetNamespace

	hostKubeConfig := ctx.Value(contextKeyHostKubeConfig{}).([]byte)
	installationStatus := plan.Status.InstallationStatus

	if err := r.syncInstallationStatus(ctx, hostKubeConfig, targetNamespace, releaseName, &installationStatus); err != nil {
		return fmt.Errorf("failed to sync release status: %v", err)
	}

	if !reflect.DeepEqual(plan.Status.InstallationStatus, installationStatus) {
		plan.Status.InstallationStatus = installationStatus
		if err := r.updateInstallPlan(ctx, plan); err != nil {
			return fmt.Errorf("failed to sync extension status: %v", err)
		}
	}

	switch plan.Status.State {
	case "":
		return r.installOrUpgradeExtension(ctx, plan, false)
	case corev1alpha1.StateInstallFailed:
		// upgrade after configuration changes
		if configChanged(plan, "") || versionChanged(plan, "") {
			return r.installOrUpgradeExtension(ctx, plan, false)
		}
	case corev1alpha1.StatePreparing, corev1alpha1.StateInstalling, corev1alpha1.StateUpgrading:
		// waiting for the installation to complete
		return nil
	case corev1alpha1.StateDeployed, corev1alpha1.StateUpgradeFailed:
		// upgrade after configuration changes
		if configChanged(plan, "") || versionChanged(plan, "") {
			return r.installOrUpgradeExtension(ctx, plan, true)
		}
	}

	if plan.Status.State == corev1alpha1.StateDeployed {
		if err := r.syncExtendedAPIStatus(ctx, r.Client, plan); err != nil {
			return fmt.Errorf("failed to sync extended api status: %v", err)
		}

		if plan.Status.Enabled != plan.Spec.Enabled {
			plan.Status.Enabled = plan.Spec.Enabled
			if err := r.updateInstallPlan(ctx, plan); err != nil {
				return fmt.Errorf("failed to sync extension status: %v", err)
			}
		}
	}

	return nil
}

func (r *InstallPlanReconciler) syncClusterAgentStatus(ctx context.Context,
	plan *corev1alpha1.InstallPlan, cluster *clusterv1alpha1.Cluster) error {
	if !clusterutils.IsClusterSchedulable(cluster) {
		klog.V(4).Infof("cluster %s is not schedulable", cluster.Name)
		return nil
	}
	if plan.Status.ClusterSchedulingStatuses == nil {
		plan.Status.ClusterSchedulingStatuses = make(map[string]corev1alpha1.InstallationStatus)
	}

	releaseName := fmt.Sprintf(agentReleaseFormat, plan.Spec.Extension.Name)
	targetNamespace := plan.Status.TargetNamespace
	kubeConfig := cluster.Spec.Connection.KubeConfig
	installationStatus := plan.Status.ClusterSchedulingStatuses[cluster.Name]

	if err := r.syncInstallationStatus(ctx, kubeConfig, targetNamespace, releaseName, &installationStatus); err != nil {
		return fmt.Errorf("failed to sync cluster agent release status: %v", err)
	}

	plan.Status.ClusterSchedulingStatuses[cluster.Name] = installationStatus
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return fmt.Errorf("failed to sync cluster agent status: %v", err)
	}

	switch plan.Status.ClusterSchedulingStatuses[cluster.Name].State {
	case "":
		return r.installOrUpgradeClusterAgent(ctx, plan, cluster, false)
	case corev1alpha1.StateInstallFailed:
		// upgrade after configuration changes
		if configChanged(plan, cluster.Name) || versionChanged(plan, cluster.Name) {
			return r.installOrUpgradeClusterAgent(ctx, plan, cluster, false)
		}
	case corev1alpha1.StatePreparing, corev1alpha1.StateInstalling, corev1alpha1.StateUpgrading:
		// waiting for the installation to complete
		return nil
	case corev1alpha1.StateDeployed, corev1alpha1.StateUpgradeFailed:
		// upgrade after configuration changes
		if configChanged(plan, cluster.Name) || versionChanged(plan, cluster.Name) {
			return r.installOrUpgradeClusterAgent(ctx, plan, cluster, true)
		}
	}

	if plan.Status.ClusterSchedulingStatuses[cluster.Name].State == corev1alpha1.StateDeployed {
		clusterClient, err := r.clusterClientSet.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %v", err)
		}

		if err := r.syncExtendedAPIStatus(ctx, clusterClient, plan); err != nil {
			return err
		}
	}

	return nil
}

func (r *InstallPlanReconciler) installOrUpgradeExtension(ctx context.Context, plan *corev1alpha1.InstallPlan, upgrade bool) error {
	updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StatePreparing, "", time.Now())
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return err
	}
	var onFailed = func(err error) error {
		if upgrade {
			klog.FromContext(ctx).Error(err, "failed to upgrade extension")
			message := fmt.Sprintf("Failed to upgrade extension: %s", err)
			updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StateUpgradeFailed, message, time.Now())
		} else {
			klog.FromContext(ctx).Error(err, "failed to install extension")
			message := fmt.Sprintf("Failed to install extension: %s", err)
			updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StateInstallFailed, message, time.Now())
		}
		return r.updateInstallPlan(ctx, plan)
	}

	chartData, caBundle, err := r.loadChartData(ctx)
	if err != nil {
		return onFailed(err)
	}

	mainChart, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return onFailed(fmt.Errorf("failed to load chart data: %v", err))
	}

	releaseName := plan.Spec.Extension.Name
	clusterRole, role := usesPermissions(mainChart)
	if err = initTargetNamespace(ctx, r.Client, plan.Status.TargetNamespace, plan.Spec.Extension.Name, clusterRole, role); err != nil {
		return onFailed(fmt.Errorf("failed to init target namespace: %v", err))
	}

	if !upgrade {
		message := fmt.Sprintf("The extension %s has been successfully initialized.", plan.Spec.Extension.Name)
		updateCondition(&plan.Status.InstallationStatus, corev1alpha1.ConditionTypeInitialized, initialized, message, metav1.ConditionTrue, time.Now())
		if err := r.updateInstallPlan(ctx, plan); err != nil {
			return err
		}
	}

	hostKubeConfig := ctx.Value(contextKeyHostKubeConfig{}).([]byte)

	extensionVersion, ok := ctx.Value(contextKeyExtensionVersion{}).(*corev1alpha1.ExtensionVersion)
	if !ok {
		return fmt.Errorf("failed to get extension version from context")
	}

	helmOptions := []helm.HelmOption{
		helm.SetKubeconfig(hostKubeConfig),
		helm.SetNamespace(plan.Status.TargetNamespace),
		helm.SetInstall(!upgrade),
		helm.SetTimeout(r.HelmExecutorOptions.Timeout),
		helm.SetKubeAsUser(fmt.Sprintf("system:serviceaccount:%s:helm-executor.%s", plan.Status.TargetNamespace, plan.Spec.Extension.Name)),
		helm.SetLabels(map[string]string{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}),
		helm.SetOverrides(r.getOverrides(mainChart, tagExtension, nil)),
		helm.SetCABundle(caBundle),
		helm.SetHistoryMax(r.HelmExecutorOptions.HistoryMax),
		helm.SetHookImage(r.getHookImageForInstall(extensionVersion, upgrade)),
	}

	executor, ok := ctx.Value(contextKeyExecutor{}).(helm.Executor)
	if !ok {
		return fmt.Errorf("failed to get executor from context")
	}

	chartURL, helmOptions := fixedOptions(extensionVersion.Spec.ChartURL, chartData, helmOptions)
	values := clusterConfig(plan, "")
	jobName, err := executor.Upgrade(ctx, releaseName, chartURL, values, helmOptions...)
	if err != nil {
		return onFailed(fmt.Errorf("failed to create executor job: %v", err))
	}

	plan.Status.ConfigHash = hashutil.FNVString(values)
	plan.Status.Version = extensionVersion.Spec.Version
	plan.Status.ReleaseName = releaseName
	plan.Status.JobName = jobName
	if upgrade {
		updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StateUpgrading, "", time.Now())
	} else {
		updateStateAndConditions(&plan.Status.InstallationStatus, corev1alpha1.StateInstalling, "", time.Now())
	}

	if err = r.updateInstallPlan(ctx, plan); err != nil {
		return err
	}

	if err = r.cleanupOutdatedJobsAndConfigMaps(ctx, plan); err != nil {
		return err
	}

	return nil
}

func (r *InstallPlanReconciler) getHookImageForInstall(extensionVersion *corev1alpha1.ExtensionVersion, upgrade bool) string {
	hookImage := extensionVersion.Annotations[corev1alpha1.ExecutorHookImageAnnotation]
	if !upgrade && extensionVersion.Annotations[corev1alpha1.ExecutorInstallHookImageAnnotation] != "" {
		hookImage = extensionVersion.Annotations[corev1alpha1.ExecutorInstallHookImageAnnotation]
	}
	if upgrade && extensionVersion.Annotations[corev1alpha1.ExecutorUpgradeHookImageAnnotation] != "" {
		hookImage = extensionVersion.Annotations[corev1alpha1.ExecutorUpgradeHookImageAnnotation]
	}
	if r.ExtensionOptions.ImageRegistry != "" && hookImage != "" {
		hookImage = path.Join(r.ExtensionOptions.ImageRegistry, hookImage)
	}
	return hookImage
}

func (r *InstallPlanReconciler) getOverrides(mainChart *chart.Chart, tag string, cluster *clusterv1alpha1.Cluster) (overrides []string) {
	disableSubchartsOverrides := func(mainChart *chart.Chart, tag string) (overrides []string) {
		for _, condition := range conditions(mainChart, tag) {
			overrides = append(overrides, fmt.Sprintf("%s=%s", condition, "false"))
		}
		return overrides
	}

	switch tag {
	case tagExtension:
		overrides = append(overrides,
			fmt.Sprintf("tags.%s=%s", tagExtension, "true"),
			fmt.Sprintf("tags.%s=%s", tagAgent, "false"),
		)

		if r.PortalURL != "" {
			overrides = append(overrides, fmt.Sprintf("%s=%s", globalPortalURL, r.PortalURL))
		}

		overrides = append(overrides, disableSubchartsOverrides(mainChart, tagAgent)...)
	case tagAgent:
		overrides = append(overrides,
			fmt.Sprintf("tags.%s=%s", tagAgent, "true"),
			fmt.Sprintf("tags.%s=%s", tagExtension, "false"),
		)
		overrides = append(overrides, disableSubchartsOverrides(mainChart, tagExtension)...)
	}

	if cluster != nil {
		clusterRole := clusterv1alpha1.ClusterRoleMember
		if clusterutils.IsHostCluster(cluster) {
			clusterRole = clusterv1alpha1.ClusterRoleHost
		}
		overrides = append(overrides,
			fmt.Sprintf("%s=%s", globalClusterName, cluster.Name),
			fmt.Sprintf("%s=%s", globalClusterRole, clusterRole),
		)
	}

	if r.ExtensionOptions != nil {
		if r.ExtensionOptions.ImageRegistry != "" {
			overrides = append(overrides, fmt.Sprintf("%s=%s", globalImageRegistry, r.ExtensionOptions.ImageRegistry))
		}
		if r.ExtensionOptions.NodeSelector != nil {
			for k, v := range r.ExtensionOptions.NodeSelector {
				k = strings.ReplaceAll(k, ".", "\\.")
				overrides = append(overrides, fmt.Sprintf("%s.%s=%s", globalNodeSelector, k, v))
			}
		}
		if r.ExtensionOptions.Ingress != nil {
			overrides = append(overrides, fmt.Sprintf("%s=%s", globalExtensionIngressClassName, r.ExtensionOptions.Ingress.IngressClassName))
			overrides = append(overrides, fmt.Sprintf("%s=%s", globalExtensionIngressDomainSuffix, r.ExtensionOptions.Ingress.DomainSuffix))
			overrides = append(overrides, fmt.Sprintf("%s=%d", globalExtensionIngressHTTPPort, r.ExtensionOptions.Ingress.HTTPPort))
			overrides = append(overrides, fmt.Sprintf("%s=%d", globalExtensionIngressHTTPSPort, r.ExtensionOptions.Ingress.HTTPSPort))
		}
	}

	return overrides
}

func conditions(mainChart *chart.Chart, tag string) []string {
	var conditions []string
	for _, dependency := range mainChart.Metadata.Dependencies {
		if dependency.Condition != "" && sliceutil.HasString(dependency.Tags, tag) {
			conditions = append(conditions, strings.Split(dependency.Condition, ",")...)
		}
	}
	return conditions
}

func (r *InstallPlanReconciler) installOrUpgradeClusterAgent(ctx context.Context, plan *corev1alpha1.InstallPlan, cluster *clusterv1alpha1.Cluster, upgrade bool) error {
	clusterName := cluster.Name
	targetNamespace := plan.Status.TargetNamespace
	releaseName := fmt.Sprintf(agentReleaseFormat, plan.Spec.Extension.Name)
	kubeConfig := cluster.Spec.Connection.KubeConfig
	clusterClient, err := r.clusterClientSet.GetRuntimeClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %v", err)
	}

	installationStatus := plan.Status.ClusterSchedulingStatuses[cluster.Name]
	updateStateAndConditions(&installationStatus, corev1alpha1.StatePreparing, "", time.Now())
	plan.Status.ClusterSchedulingStatuses[cluster.Name] = installationStatus
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return fmt.Errorf("failed to sync cluster agent status: %v", err)
	}

	var onFailed = func(err error) error {
		if upgrade {
			klog.FromContext(ctx).Error(err, "failed to upgrade extension")
			message := fmt.Sprintf("failed to upgrade extension: %s", err)
			updateStateAndConditions(&installationStatus, corev1alpha1.StateUpgradeFailed, message, time.Now())
		} else {
			klog.FromContext(ctx).Error(err, "failed to install extension")
			message := fmt.Sprintf("Failed to install extension: %s", err)
			updateStateAndConditions(&installationStatus, corev1alpha1.StateInstallFailed, message, time.Now())
		}
		plan.Status.ClusterSchedulingStatuses[clusterName] = installationStatus
		return r.updateInstallPlan(ctx, plan)
	}

	chartData, caBundle, err := r.loadChartData(ctx)
	if err != nil {
		return onFailed(fmt.Errorf("failed to load chart data: %v", err))
	}

	mainChart, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return onFailed(fmt.Errorf("failed to load chart data: %v", err))
	}

	clusterRole, role := usesPermissions(mainChart)
	if err = initTargetNamespace(ctx, clusterClient, targetNamespace, plan.Spec.Extension.Name, clusterRole, role); err != nil {
		return onFailed(fmt.Errorf("failed to init target namespace: %v", err))
	}

	if !upgrade {
		updateCondition(&installationStatus, corev1alpha1.ConditionTypeInitialized,
			initialized, "The extension agent has been successfully initialized.", metav1.ConditionTrue, time.Now())
		plan.Status.ClusterSchedulingStatuses[clusterName] = installationStatus
		if err := r.updateInstallPlan(ctx, plan); err != nil {
			return err
		}
	}

	extensionVersion, ok := ctx.Value(contextKeyExtensionVersion{}).(*corev1alpha1.ExtensionVersion)
	if !ok {
		return fmt.Errorf("failed to get extension version from context")
	}

	executor, ok := ctx.Value(contextKeyExecutor{}).(helm.Executor)
	if !ok {
		return fmt.Errorf("failed to get executor from context")
	}

	clusterRoleName := clusterv1alpha1.ClusterRoleMember
	if clusterutils.IsHostCluster(cluster) {
		clusterRoleName = clusterv1alpha1.ClusterRoleHost
	}
	helmOptions := []helm.HelmOption{
		helm.SetKubeconfig(kubeConfig),
		helm.SetNamespace(targetNamespace),
		helm.SetInstall(!upgrade),
		helm.SetTimeout(r.HelmExecutorOptions.Timeout),
		helm.SetKubeAsUser(fmt.Sprintf("system:serviceaccount:%s:helm-executor.%s", targetNamespace, plan.Spec.Extension.Name)),
		helm.SetOverrides(r.getOverrides(mainChart, tagAgent, cluster)),
		helm.SetLabels(map[string]string{corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name}),
		helm.SetCABundle(caBundle),
		helm.SetHistoryMax(r.HelmExecutorOptions.HistoryMax),
		helm.SetHookImage(r.getHookImageForInstall(extensionVersion, upgrade)),
		helm.SetClusterRole(string(clusterRoleName)),
		helm.SetClusterName(clusterName),
	}

	chartURL, helmOptions := fixedOptions(extensionVersion.Spec.ChartURL, chartData, helmOptions)
	values := clusterConfig(plan, cluster.Name)
	jobName, err := executor.Upgrade(ctx, releaseName, chartURL, values, helmOptions...)
	if err != nil {
		return onFailed(fmt.Errorf("failed to create executor job: %v", err))
	}

	installationStatus.ConfigHash = hashutil.FNVString(values)
	installationStatus.ReleaseName = releaseName
	installationStatus.Version = extensionVersion.Spec.Version
	installationStatus.TargetNamespace = targetNamespace
	installationStatus.JobName = jobName
	if upgrade {
		updateStateAndConditions(&installationStatus, corev1alpha1.StateUpgrading, "", time.Now())
	} else {
		updateStateAndConditions(&installationStatus, corev1alpha1.StateInstalling, "", time.Now())
	}

	plan.Status.ClusterSchedulingStatuses[clusterName] = installationStatus
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return err
	}

	if err = r.cleanupOutdatedJobsAndConfigMaps(ctx, plan); err != nil {
		return err
	}
	return nil
}

func fixedOptions(chartURL string, chartData []byte, helmOptions []helm.HelmOption) (string, []helm.HelmOption) {
	if chartURL == "" {
		helmOptions = append(helmOptions, helm.SetChartData(chartData))
	} else {
		if strings.HasPrefix(chartURL, "oci://") {
			parts := strings.Split(chartURL, ":")
			if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
				tag := parts[len(parts)-1]
				if tag != "" {
					chartURL = strings.TrimSuffix(chartURL, ":"+tag)
					helmOptions = append(helmOptions, helm.SetVersion(tag))
				}
			}
		}
	}
	return chartURL, helmOptions
}

func (r *InstallPlanReconciler) uninstallClusterAgent(ctx context.Context, plan *corev1alpha1.InstallPlan, clusterName string) error {
	logger := klog.FromContext(ctx).WithValues("cluster", clusterName)
	installationStatus := plan.Status.ClusterSchedulingStatuses[clusterName]
	releaseName := installationStatus.ReleaseName
	targetNamespace := installationStatus.TargetNamespace

	if releaseName == "" {
		delete(plan.Status.ClusterSchedulingStatuses, clusterName)
		return r.updateInstallPlan(ctx, plan)
	}

	var cluster clusterv1alpha1.Cluster
	if err := r.Get(ctx, types.NamespacedName{Name: clusterName}, &cluster); err != nil {
		if errors.IsNotFound(err) {
			delete(plan.Status.ClusterSchedulingStatuses, clusterName)
			return r.updateInstallPlan(ctx, plan)
		}
		return err
	}

	kubeConfig := cluster.Spec.Connection.KubeConfig

	executor, ok := ctx.Value(contextKeyExecutor{}).(helm.Executor)
	if !ok {
		return fmt.Errorf("failed to get executor from context")
	}

	// Check if the target helm release exists.
	// If it does, there is no need to execute the installation process again.
	_, err := executor.Get(ctx, releaseName, helm.SetKubeconfig(kubeConfig), helm.SetNamespace(targetNamespace))
	if err != nil {
		if isReleaseNotFoundError(err) {
			delete(plan.Status.ClusterSchedulingStatuses, clusterName)
			return r.updateInstallPlan(ctx, plan)
		}
		klog.FromContext(ctx).Error(err, "failed to get helm release status")
		return fmt.Errorf("failed to get helm release status: %v", err)
	}

	if err := r.syncInstallationStatus(ctx, kubeConfig, targetNamespace, releaseName, &installationStatus); err != nil {
		return fmt.Errorf("failed to sync cluster agent release status: %v", err)
	}

	plan.Status.ClusterSchedulingStatuses[cluster.Name] = installationStatus
	if err := r.updateInstallPlan(ctx, plan); err != nil {
		return fmt.Errorf("failed to sync cluster agent status: %v", err)
	}

	if installationStatus.State != corev1alpha1.StateUninstalling &&
		installationStatus.State != corev1alpha1.StateUninstallFailed &&
		installationStatus.State != corev1alpha1.StateUninstalled {

		extensionVersion, ok := ctx.Value(contextKeyExtensionVersion{}).(*corev1alpha1.ExtensionVersion)
		if !ok {
			return fmt.Errorf("failed to get extension version from context")
		}

		clusterRoleName := clusterv1alpha1.ClusterRoleMember
		if clusterutils.IsHostCluster(&cluster) {
			clusterRoleName = clusterv1alpha1.ClusterRoleHost
		}
		helmOptions := []helm.HelmOption{
			helm.SetKubeconfig(kubeConfig),
			helm.SetNamespace(targetNamespace),
			helm.SetTimeout(r.HelmExecutorOptions.Timeout),
			helm.SetHookImage(r.getHookImageForUninstall(extensionVersion)),
			helm.SetClusterRole(string(clusterRoleName)),
			helm.SetClusterName(clusterName),
		}

		jobName, err := executor.Uninstall(ctx, releaseName, helmOptions...)
		if err != nil {
			logger.Error(err, "failed to uninstall helm release")
			return err
		}
		installationStatus.JobName = jobName
		updateStateAndConditions(&installationStatus, corev1alpha1.StateUninstalling, "", time.Now())
		plan.Status.ClusterSchedulingStatuses[clusterName] = installationStatus
		if err := r.updateInstallPlan(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

func (r *InstallPlanReconciler) newExecutor(plan *corev1alpha1.InstallPlan) (helm.Executor, error) {
	executorOptions := []helm.ExecutorOption{
		helm.SetExecutorLabels(map[string]string{
			constants.KubeSphereManagedLabel:     "true",
			corev1alpha1.ExtensionReferenceLabel: plan.Spec.Extension.Name,
		}),
		helm.SetExecutorOwner(&metav1.OwnerReference{
			APIVersion: corev1alpha1.SchemeGroupVersion.String(),
			Kind:       corev1alpha1.ResourceKindInstallPlan,
			Name:       plan.Name,
			UID:        plan.UID,
		}),
		helm.SetExecutorImage(r.HelmExecutorOptions.Image),
		helm.SetExecutorNamespace(plan.Status.TargetNamespace),
		helm.SetExecutorBackoffLimit(0),
		helm.SetTTLSecondsAfterFinished(r.HelmExecutorOptions.JobTTLAfterFinished),
	}
	if r.HelmExecutorOptions.Resources != nil {
		executorOptions = append(executorOptions, helm.SetExecutorResources(corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(r.HelmExecutorOptions.Resources.Limits[corev1.ResourceCPU]),
				corev1.ResourceMemory: resource.MustParse(r.HelmExecutorOptions.Resources.Limits[corev1.ResourceMemory]),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(r.HelmExecutorOptions.Resources.Requests[corev1.ResourceCPU]),
				corev1.ResourceMemory: resource.MustParse(r.HelmExecutorOptions.Resources.Requests[corev1.ResourceMemory]),
			},
		}))
	}
	return helm.NewExecutor(executorOptions...)
}

func updateStateAndConditions(installationStatus *corev1alpha1.InstallationStatus, state, message string, lastTransitionTime time.Time) {
	lastTransitionTime = lastTransitionTime.Round(time.Second)
	fixedState := state
	if state == corev1alpha1.StateInstalled || state == corev1alpha1.StateUpgraded {
		fixedState = corev1alpha1.StateDeployed
	}

	if updateState(installationStatus, fixedState, lastTransitionTime) {
		switch state {
		case corev1alpha1.StateInstalled:
			updateCondition(installationStatus, corev1alpha1.ConditionTypeInstalled, installSuccessful, message, metav1.ConditionTrue, lastTransitionTime)
		case corev1alpha1.StateInstallFailed:
			updateCondition(installationStatus, corev1alpha1.ConditionTypeInstalled, installFailed, message, metav1.ConditionFalse, lastTransitionTime)
		case corev1alpha1.StateUpgraded:
			updateCondition(installationStatus, corev1alpha1.ConditionTypeUpgraded, upgradeSuccessful, message, metav1.ConditionTrue, lastTransitionTime)
		case corev1alpha1.StateUpgradeFailed:
			updateCondition(installationStatus, corev1alpha1.ConditionTypeUpgraded, upgradeFailed, message, metav1.ConditionFalse, lastTransitionTime)
		case corev1alpha1.StateUninstallFailed:
			updateCondition(installationStatus, corev1alpha1.ConditionTypeUninstalled, uninstallFailed, message, metav1.ConditionFalse, lastTransitionTime)
		}
	}
}

func (r *InstallPlanReconciler) syncInstallationStatus(ctx context.Context, kubeConfig []byte, namespace string, releaseName string, installationStatus *corev1alpha1.InstallationStatus) error {
	var job *batchv1.Job
	if installationStatus.JobName != "" {
		job = &batchv1.Job{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: installationStatus.JobName}, job); err != nil {
			if errors.IsNotFound(err) {
				job = nil
				klog.FromContext(ctx).Info("related job not found", "namespace", installationStatus.TargetNamespace, "job", installationStatus.JobName)
			} else {
				return fmt.Errorf("failed to get job: %v", err)
			}
		}
	}

	executor, ok := ctx.Value(contextKeyExecutor{}).(helm.Executor)
	if !ok {
		return fmt.Errorf("failed to get executor from context")
	}

	// Check if the target helm release exists.
	// If it does, there is no need to execute the installation process again.
	release, err := executor.Get(ctx, releaseName, helm.SetKubeconfig(kubeConfig), helm.SetNamespace(namespace))
	if err != nil && !isReleaseNotFoundError(err) {
		klog.FromContext(ctx).Error(err, "failed to get helm release status")
		return fmt.Errorf("failed to get helm release status: %v", err)
	}

	if job != nil {
		action := job.Annotations[helm.ExecutorJobActionAnnotation]
		active, completed, failed := jobStatus(job)
		condition := latestJobCondition(job)
		lastTransitionTime := condition.LastTransitionTime.Time

		if failed {
			switch action {
			case helm.ActionInstall:
				updateStateAndConditions(installationStatus, corev1alpha1.StateInstallFailed, condition.Message, lastTransitionTime)
			case helm.ActionUpgrade:
				updateStateAndConditions(installationStatus, corev1alpha1.StateUpgradeFailed, condition.Message, lastTransitionTime)
			case helm.ActionUninstall:
				updateStateAndConditions(installationStatus, corev1alpha1.StateUninstallFailed, condition.Message, lastTransitionTime)
			}
		}

		if completed && action == helm.ActionUninstall && release == nil {
			updateStateAndConditions(installationStatus, corev1alpha1.StateUninstalled, condition.Message, lastTransitionTime)
		}

		if active {
			lastTransitionTime = job.CreationTimestamp.Time
			switch action {
			case helm.ActionInstall:
				updateStateAndConditions(installationStatus, corev1alpha1.StateInstalling, "", lastTransitionTime)
			case helm.ActionUpgrade:
				updateStateAndConditions(installationStatus, corev1alpha1.StateUpgrading, "", lastTransitionTime)
			case helm.ActionUninstall:
				updateStateAndConditions(installationStatus, corev1alpha1.StateUninstalling, "", lastTransitionTime)
			}
		}
	}

	if release != nil {
		switch release.Info.Status {
		case helmrelease.StatusFailed:
			if release.Version > 1 {
				updateStateAndConditions(installationStatus, corev1alpha1.StateUpgradeFailed, release.Info.Description, release.Info.LastDeployed.Time)
			} else {
				updateStateAndConditions(installationStatus, corev1alpha1.StateInstallFailed, release.Info.Description, release.Info.LastDeployed.Time)
			}
		case helmrelease.StatusDeployed:
			installationStatus.Version = release.Chart.Metadata.Version
			installationStatus.ReleaseName = release.Name
			if release.Version > 1 {
				updateStateAndConditions(installationStatus, corev1alpha1.StateUpgraded, release.Info.Description, release.Info.LastDeployed.Time)
			} else {
				updateStateAndConditions(installationStatus, corev1alpha1.StateInstalled, release.Info.Description, release.Info.LastDeployed.Time)
			}
		case helmrelease.StatusPendingInstall:
			updateStateAndConditions(installationStatus, corev1alpha1.StateInstalling, release.Info.Description, release.Info.LastDeployed.Time)
		case helmrelease.StatusPendingRollback, helmrelease.StatusPendingUpgrade:
			updateStateAndConditions(installationStatus, corev1alpha1.StateUpgrading, release.Info.Description, release.Info.LastDeployed.Time)
		case helmrelease.StatusUninstalling:
			updateStateAndConditions(installationStatus, corev1alpha1.StateUninstalling, release.Info.Description, release.Info.LastDeployed.Time)
		case helmrelease.StatusUninstalled:
			updateStateAndConditions(installationStatus, corev1alpha1.StateUninstalled, release.Info.Description, release.Info.LastDeployed.Time)
		}
	}
	return nil
}

func (r *InstallPlanReconciler) mapper(ctx context.Context, object client.Object) []reconcile.Request {
	var requests []reconcile.Request
	if cluster, ok := object.(*clusterv1alpha1.Cluster); ok {
		installPlans := &corev1alpha1.InstallPlanList{}
		if err := r.List(ctx, installPlans); err != nil {
			klog.Warningf("failed to list install plans: %v", err)
			return requests
		}
		for _, plan := range installPlans.Items {
			if plan.Spec.ClusterScheduling == nil {
				continue
			}
			if slices.Contains(plan.Spec.ClusterScheduling.Placement.Clusters, cluster.Name) {
				requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: plan.Name}})
			} else if plan.Spec.ClusterScheduling.Placement.ClusterSelector != nil {
				selector, err := metav1.LabelSelectorAsSelector(plan.Spec.ClusterScheduling.Placement.ClusterSelector)
				if err != nil {
					klog.Warningf("failed to parse cluster selector: %v", err)
					continue
				}
				if selector.Matches(labels.Set(cluster.Labels)) {
					requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: plan.Name}})
				}
			}
		}

	}
	return requests
}

func jobStatus(job *batchv1.Job) (active, completed, failed bool) {
	if job == nil {
		return
	}
	completed = job.Spec.Completions != nil && job.Status.Succeeded >= *job.Spec.Completions
	failed = job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit
	active = !completed && !failed
	return
}

type contextKeyExtensionVersion struct{}
type contextKeyExecutor struct{}
type contextKeyHostKubeConfig struct{}
