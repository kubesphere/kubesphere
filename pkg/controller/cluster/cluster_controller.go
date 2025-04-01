/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1beta1"
	"kubesphere.io/utils/helm"

	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/version"
)

// Cluster controller only runs under multicluster mode. Cluster controller is following below steps,
//   1. Wait for cluster agent is ready if the connection type is proxy
//   2. Join cluster into federation control plane if kubeconfig is ready.
//   3. Pull cluster version, set result to cluster status
// Also put all clusters back into queue every 5 * time.Minute to sync cluster status, this is needed
// in case there aren't any cluster changes made.
// Also check if all the clusters are ready by the spec.connection.kubeconfig every resync period

const (
	controllerName  = "cluster"
	installAction   = "install"
	upgradeAction   = "upgrade"
	reInstallAction = "reinstall"
)

const (
	initializedAnnotation  = "kubesphere.io/initialized"
	installJobAnnotation   = "kubesphere.io/install-core-jobname"
	ksCoreActionAnnotation = "kubesphere.io/ks-core-action"
)

// Cluster template for reconcile host cluster if there is none.
var hostClusterTemplate = &clusterv1alpha1.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: "host",
		Annotations: map[string]string{
			"kubesphere.io/description": "The description was created by KubeSphere automatically. " +
				"It is recommended that you use the Host Cluster to manage clusters only " +
				"and deploy workloads on Member Clusters.",
			constants.CreatorAnnotationKey: "admin",
		},
		Labels: map[string]string{
			clusterv1alpha1.HostCluster:      "",
			constants.KubeSphereManagedLabel: "true",
		},
	},
	Spec: clusterv1alpha1.ClusterSpec{
		Provider: "kubesphere",
		Connection: clusterv1alpha1.Connection{
			Type: clusterv1alpha1.ConnectionTypeDirect,
		},
	},
}

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type Reconciler struct {
	client.Client
	hostConfig          *rest.Config
	hostClusterName     string
	resyncPeriod        time.Duration
	installLock         *sync.Map
	clusterClient       clusterclient.Interface
	clusterUID          types.UID
	tls                 bool
	HelmExecutorOptions *options.HelmExecutorOptions
}

// SetupWithManager setups the Reconciler with manager.
func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	kubeSystem, err := mgr.K8sClient.CoreV1().Namespaces().Get(context.Background(), metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return err
	}
	r.hostConfig = mgr.K8sClient.Config()
	r.clusterClient = mgr.ClusterClient
	r.hostClusterName = mgr.MultiClusterOptions.HostClusterName
	r.resyncPeriod = mgr.MultiClusterOptions.ClusterControllerResyncPeriod
	r.clusterUID = kubeSystem.UID
	r.installLock = &sync.Map{}
	r.tls = mgr.Options.KubeSphereOptions.TLS
	r.HelmExecutorOptions = mgr.Options.HelmExecutorOptions
	r.Client = mgr.GetClient()
	if err := mgr.Add(r); err != nil {
		return fmt.Errorf("unable to add cluster-controller to manager: %v", err)
	}
	return builder.
		ControllerManagedBy(mgr).
		For(
			&clusterv1alpha1.Cluster{},
			builder.WithPredicates(
				clusterChangedPredicate{
					stateChangedAnnotations: []string{
						"kubesphere.io/syncAt",
						ksCoreActionAnnotation,
					},
				},
			),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Named(controllerName).
		Complete(r)
}

type clusterChangedPredicate struct {
	predicate.Funcs
	stateChangedAnnotations []string
}

func (c clusterChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldCluster := e.ObjectOld.(*clusterv1alpha1.Cluster)
	newCluster := e.ObjectNew.(*clusterv1alpha1.Cluster)
	if !reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) ||
		newCluster.DeletionTimestamp != nil {
		return true
	}
	for _, key := range c.stateChangedAnnotations {
		oldValue, oldExist := oldCluster.Annotations[key]
		newValue, newExist := newCluster.Annotations[key]
		if oldExist != newExist || (oldExist && newExist && oldValue != newValue) {
			return true
		}
	}
	return false
}

// NeedLeaderElection implements the LeaderElectionRunnable interface,
// controllers need to be run in leader election mode.
func (r *Reconciler) NeedLeaderElection() bool {
	return true
}

func (r *Reconciler) Start(ctx context.Context) error {
	// refresh cluster configz every resync period
	go wait.Until(func() {
		if err := r.createHostClusterIfNotExists(); err != nil {
			klog.Errorf("failed to reconcile cluster ready status, err: %v", err)
		}
	}, r.resyncPeriod, ctx.Done())
	return nil
}

func (r *Reconciler) createHostClusterIfNotExists() error {
	hostKubeConfig, err := clusterutils.BuildKubeconfigFromRestConfig(r.hostConfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for host cluster: %v", err)
	}

	cluster := &clusterv1alpha1.Cluster{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: r.hostClusterName}, cluster); err != nil {
		if errors.IsNotFound(err) {
			cluster = hostClusterTemplate.DeepCopy()
			cluster.Spec.Connection.KubeConfig = hostKubeConfig
			cluster.Name = r.hostClusterName
			if err = r.Create(context.Background(), cluster); err != nil {
				return fmt.Errorf("failed to create host cluster: %v", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get host cluster: %v", err)
	}

	// update host cluster config
	if !bytes.Equal(cluster.Spec.Connection.KubeConfig, hostKubeConfig) ||
		cluster.Labels[clusterv1alpha1.HostCluster] != "" {
		cluster.Spec.Connection.KubeConfig = hostKubeConfig
		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string)
		}
		cluster.Labels[clusterv1alpha1.HostCluster] = ""
		if err = r.Update(context.Background(), cluster); err != nil {
			return fmt.Errorf("failed to update host cluster: %v", err)
		}
	}

	return nil
}

// Reconcile reconciles the Cluster object.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("Starting to sync cluster %s", req.Name)
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing cluster %s in %s", req.Name, time.Since(startTime))
	}()

	cluster := &clusterv1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// The object is being deleted
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sets.New(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			return ctrl.Result{}, nil
		}

		if err := r.unbindWorkspaceTemplate(ctx, cluster); err != nil {
			klog.Errorf("Failed to unbind workspace for %s: %v", req.Name, err)
			return ctrl.Result{}, err
		}

		// cleanup after cluster has been deleted
		if err := r.cleanup(ctx, cluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("cleanup for cluster %s failed: %s", cluster.Name, err.Error())
		}
		if err := r.syncClusterMembers(ctx, cluster); err != nil {
			klog.Errorf("Failed to sync cluster members for %s: %v", req.Name, err)
			return ctrl.Result{}, err
		}

		// remove our cluster finalizer
		finalizers := sets.New(cluster.ObjectMeta.Finalizers...)
		finalizers.Delete(clusterv1alpha1.Finalizer)
		cluster.ObjectMeta.Finalizers = finalizers.UnsortedList()
		return ctrl.Result{}, r.Update(ctx, cluster)
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then let's add the finalizer and update the object.
	// This is equivalent to registering our finalizer.
	if !sets.New(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
		cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
		if err := r.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer for cluster %s: %s", cluster.Name, err)
		}
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		klog.V(5).Infof("Skipping to join cluster %s cause the kubeconfig is empty", cluster.Name)
		return ctrl.Result{}, nil
	}

	clusterClient, err := r.clusterClient.GetClusterClient(cluster.Name)
	if err != nil {
		return ctrl.Result{}, r.updateClusterReadyCondition(
			ctx, cluster, fmt.Errorf("failed to get cluster client for %s: %s", cluster.Name, err),
		)
	}

	// Use kube-system namespace UID as cluster ID
	kubeSystem := &corev1.Namespace{}
	if err = clusterClient.Client.Get(ctx, client.ObjectKey{Name: metav1.NamespaceSystem}, kubeSystem); err != nil {
		return ctrl.Result{}, r.updateClusterReadyCondition(
			ctx, cluster, fmt.Errorf("failed to get kube-system namespace for %s: %s", cluster.Name, err),
		)
	}

	// cluster is ready, we can pull kubernetes cluster info through agent
	// since there is no agent necessary for host cluster, so updates for host cluster
	// are safe.
	if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 ||
		cluster.Status.UID != kubeSystem.UID {
		cluster.Spec.Connection.KubernetesAPIEndpoint = clusterClient.RestConfig.Host
		cluster.Status.UID = kubeSystem.UID
		// Prevent the situation where only the status update causes it to never enter the queue again

		if err = r.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if !r.checkIfClusterIsHostCluster(kubeSystem.UID) {
		if err = r.reconcileMemberCluster(ctx, cluster, clusterClient); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to reconcile cluster %s: %s", cluster.Name, err)
		}
	}

	if err := r.syncClusterLabel(ctx, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync cluster label for %s: %s", cluster.Name, err)
	}

	if err := r.syncKubeSphereVersion(ctx, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync kubesphere version for %s: %s", cluster.Name, err)
	}

	if err := r.syncKubernetesVersion(ctx, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync kubernetes version for %s: %s", cluster.Name, err)
	}

	if err := r.syncClusterName(ctx, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync cluster name for %s: %s", cluster.Name, err)
	}

	if err := r.syncClusterMembers(ctx, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync cluster membership for %s: %s", cluster.Name, err)
	}

	return ctrl.Result{RequeueAfter: r.resyncPeriod}, nil
}

// syncClusterLabel syncs label IDs from annotations to the individual Label CRs.
func (r *Reconciler) syncClusterLabel(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	annotations := cluster.Annotations
	if len(annotations) == 0 {
		return nil
	}
	labels := strings.Split(annotations[clusterv1alpha1.ClusterLabelIDsAnnotation], ",")
	if len(labels) == 0 {
		return nil
	}

	klog.V(4).Infof("sync cluster %s to labels: %v", cluster.Name, labels)
	for _, name := range labels {
		label := &clusterv1alpha1.Label{}
		if err := r.Get(ctx, client.ObjectKey{Name: strings.TrimSpace(name)}, label); err != nil {
			if errors.IsNotFound(err) {
				continue
			} else {
				return err
			}
		}
		clusters := sets.NewString(label.Spec.Clusters...)
		if clusters.Has(cluster.Name) {
			continue
		}
		clusters.Insert(cluster.Name)
		label.Spec.Clusters = clusters.List()
		if err := r.Update(ctx, label); err != nil {
			return err
		}
	}

	delete(annotations, clusterv1alpha1.ClusterLabelIDsAnnotation)
	// the cluster object will be updated at the end of the reconciling
	cluster.Annotations = annotations
	return nil
}

func (r *Reconciler) needInstall(ctx context.Context, member *clusterv1alpha1.Cluster) (bool, error) {
	conditions := member.Status.Conditions
	action := member.Annotations[ksCoreActionAnnotation]
	switch action {
	case "", installAction:
		for _, condition := range conditions {
			if condition.Type == clusterv1alpha1.ClusterKSCoreReady {
				return false, nil
			}
		}

	case upgradeAction:
		install := false
		for _, condition := range conditions {
			if condition.Type == clusterv1alpha1.ClusterKSCoreReady && condition.Status == corev1.ConditionTrue {
				install = true
			}
		}

		clusters := &clusterv1alpha1.ClusterList{}
		if err := r.List(ctx, clusters); err != nil {
			return false, err
		}

		host := &clusterv1alpha1.Cluster{}
		for _, c := range clusters.Items {
			if c.Status.UID == r.clusterUID {
				host = &c
				break
			}
		}

		if install && host.Status.KubeSphereVersion != "" &&
			host.Status.KubeSphereVersion != member.Status.KubeSphereVersion {
			klog.Infof("host cluster ks core version: %s, member cluster ks core version: %s",
				host.Status.KubeSphereVersion, member.Status.KubeSphereVersion)
			return true, nil
		}

	case reInstallAction:
		return true, nil

	default:
		klog.Warningf("unknown action %s", action)
	}

	return false, nil
}

func (r *Reconciler) reconcileMemberCluster(ctx context.Context, cluster *clusterv1alpha1.Cluster, clusterClient *clusterclient.ClusterClient) error {
	// Install KS Core in member cluster
	need, err := r.needInstall(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to check if need install ks core: %v", err)
	}

	if need || configChanged(cluster) {
		// get the lock, make sure only one thread is executing the helm task
		if _, ok := r.installLock.Load(cluster.Name); ok {
			return nil
		}
		r.installLock.Store(cluster.Name, "")
		defer r.installLock.Delete(cluster.Name)
		klog.Infof("Starting installing KS Core for the cluster %s", cluster.Name)
		defer klog.Infof("Finished installing KS Core for the cluster %s", cluster.Name)

		hostConfig, err := getKubeSphereConfig(ctx, r.Client)
		if err != nil {
			return fmt.Errorf("failed to get KubeSphere config: %v", err)
		}
		status := corev1.ConditionTrue
		message := "KS Core is available now"
		if err := r.installOrUpgradeKSCoreInMemberCluster(ctx, r.HelmExecutorOptions, cluster,
			hostConfig.AuthenticationOptions.Issuer.JWTSecret, hostConfig.MultiClusterOptions.ChartPath); err != nil {
			status = corev1.ConditionFalse
			message = "KS Core installation failed"
			klog.Errorf("failed to install KS Core in cluster %s: %v", cluster.Name, err)
		}

		r.updateClusterCondition(cluster, clusterv1alpha1.ClusterCondition{
			Type:               clusterv1alpha1.ClusterKSCoreReady,
			Status:             status,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             clusterv1alpha1.ClusterKSCoreReady,
			Message:            message,
		})
		setConfigHash(cluster)
		if err = r.Update(ctx, cluster); err != nil {
			return fmt.Errorf("failed to update cluster %s: %v", cluster.Name, err)
		}
		return nil
	}
	if err := r.updateKubeConfigExpirationDateCondition(ctx, cluster, clusterClient.Client, clusterClient.RestConfig); err != nil {
		// should not block the whole process
		klog.Warningf("sync KubeConfig expiration date for cluster %s failed: %v", cluster.Name, err)
	}
	return nil
}

func (r *Reconciler) syncClusterName(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %s", err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		kubeSphereNamespace := &corev1.Namespace{}
		if err = clusterClient.Get(ctx, client.ObjectKey{Name: constants.KubeSphereNamespace}, kubeSphereNamespace); err != nil {
			return err
		}
		annotations := kubeSphereNamespace.Annotations
		if annotations[clusterv1alpha1.AnnotationClusterName] == cluster.Name &&
			annotations[clusterv1alpha1.AnnotationHostClusterName] == r.hostClusterName {
			return nil
		}
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[clusterv1alpha1.AnnotationClusterName] = cluster.Name
		annotations[clusterv1alpha1.AnnotationHostClusterName] = r.hostClusterName
		kubeSphereNamespace.Annotations = annotations
		return clusterClient.Update(ctx, kubeSphereNamespace)
	})
}

func (r *Reconciler) checkIfClusterIsHostCluster(clusterKubeSystemUID types.UID) bool {
	return r.clusterUID == clusterKubeSystemUID
}

func (r *Reconciler) tryFetchKubeSphereVersion(ctx context.Context, cluster *clusterv1alpha1.Cluster) (string, error) {
	clusterClient, err := r.clusterClient.GetClusterClient(cluster.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster client: %s", err)
	}

	scheme := "http"
	port := "80"
	if r.tls {
		scheme = "https"
		port = "443"
	}
	response, err := clusterClient.KubernetesClient.CoreV1().Services(constants.KubeSphereNamespace).
		ProxyGet(scheme, constants.KubeSphereAPIServerName, port, "/version", nil).
		DoRaw(ctx)
	if err != nil {
		return "", err
	}

	info := version.Info{}
	if err = json.Unmarshal(response, &info); err != nil {
		return "", err
	}

	// currently, we kubesphere v2.1 cannot be joined as a member cluster, and it will never be reconciled,
	// so we don't consider that situation
	// for kubesphere v3.0.0, the gitVersion is always v0.0.0, so we return v3.0.0
	if info.GitVersion == "v0.0.0" {
		return "v3.0.0", nil
	}

	if len(info.GitVersion) == 0 {
		return "unknown", nil
	}

	return info.GitVersion, nil
}

func (r *Reconciler) updateClusterReadyCondition(ctx context.Context, cluster *clusterv1alpha1.Cluster, err error) error {
	condition := clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterReady,
		Status:             corev1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterReady),
		Message:            "Cluster is available now",
	}

	if err != nil {
		condition.Status = corev1.ConditionFalse
		condition.Message = err.Error()
		r.updateClusterCondition(cluster, condition)
		if updateErr := r.Update(ctx, cluster); updateErr != nil {
			return updateErr
		}
		return err
	}

	r.updateClusterCondition(cluster, condition)
	return r.Update(ctx, cluster)
}

// updateClusterCondition updates condition in cluster conditions using giving condition
// adds condition if not existed
func (r *Reconciler) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = make([]clusterv1alpha1.ClusterCondition, 0)
	}

	newConditions := make([]clusterv1alpha1.ClusterCondition, 0)
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == condition.Type {
			continue
		}
		newConditions = append(newConditions, cond)
	}

	newConditions = append(newConditions, condition)
	cluster.Status.Conditions = newConditions
}

func (r *Reconciler) syncKubeSphereVersion(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	kubeSphereVersion, err := r.tryFetchKubeSphereVersion(ctx, cluster)
	if err != nil {
		// The KubeSphere service is unavailable
		klog.Errorf("failed to get KubeSphere version, err: %#v", err)
		return r.updateClusterReadyCondition(ctx, cluster, err)
	}

	cluster.Status.KubeSphereVersion = kubeSphereVersion
	return r.updateClusterReadyCondition(ctx, cluster, nil)
}

func (r *Reconciler) syncKubernetesVersion(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	clusterClient, err := r.clusterClient.GetClusterClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %s", err)
	}

	nodes := &corev1.NodeList{}
	if err = clusterClient.Client.List(ctx, nodes); err != nil {
		return fmt.Errorf("failed to list nodes: %s", err)
	}

	kubernetesVersion := clusterClient.KubernetesVersion
	nodeCount := len(nodes.Items)

	if cluster.Status.KubernetesVersion != kubernetesVersion ||
		cluster.Status.NodeCount != nodeCount {

		cluster = cluster.DeepCopy()
		cluster.Status.NodeCount = nodeCount
		cluster.Status.KubernetesVersion = kubernetesVersion

		if err = r.Update(ctx, cluster); err != nil {
			return fmt.Errorf("failed to update cluster: %s", err)
		}
	}

	return nil
}

// syncClusterMembers Sync granted clusters for users periodically
func (r *Reconciler) syncClusterMembers(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	users := &iamv1beta1.UserList{}
	if err := r.List(ctx, users); err != nil {
		return err
	}

	grantedUsers := sets.New[string]()
	clusterName := cluster.Name
	if cluster.DeletionTimestamp.IsZero() {
		clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %s", err)
		}

		if err = r.createClusterAdmin(ctx, cluster); err != nil {
			return fmt.Errorf("failed to create cluster admin: %s", err)
		}

		clusterRoleBindings := &iamv1beta1.ClusterRoleBindingList{}
		if err := clusterClient.List(ctx, clusterRoleBindings, client.HasLabels{iamv1beta1.UserReferenceLabel}); err != nil {
			return fmt.Errorf("failed to list clusterrolebindings: %s", err)
		}
		for _, clusterRoleBinding := range clusterRoleBindings.Items {
			for _, sub := range clusterRoleBinding.Subjects {
				if sub.Kind == iamv1beta1.ResourceKindUser {
					grantedUsers.Insert(sub.Name)
				}
			}
		}
	}

	for i := range users.Items {
		user := &users.Items[i]
		grantedClustersAnnotation := user.Annotations[iamv1beta1.GrantedClustersAnnotation]
		var grantedClusters sets.Set[string]
		if len(grantedClustersAnnotation) > 0 {
			grantedClusters = sets.New(strings.Split(grantedClustersAnnotation, ",")...)
		} else {
			grantedClusters = sets.New[string]()
		}
		if grantedUsers.Has(user.Name) && !grantedClusters.Has(clusterName) {
			grantedClusters.Insert(clusterName)
		} else if !grantedUsers.Has(user.Name) && grantedClusters.Has(clusterName) {
			grantedClusters.Delete(clusterName)
		}
		grantedClustersAnnotation = strings.Join(grantedClusters.UnsortedList(), ",")
		if user.Annotations[iamv1beta1.GrantedClustersAnnotation] != grantedClustersAnnotation {
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				if err := r.Get(ctx, types.NamespacedName{Name: user.Name}, user); err != nil {
					return err
				}
				if user.Annotations == nil {
					user.Annotations = make(map[string]string)
				}
				user.Annotations[iamv1beta1.GrantedClustersAnnotation] = grantedClustersAnnotation
				return r.Update(ctx, user)
			})
			if err != nil {
				return fmt.Errorf("failed to update granted clusters annotation: %s", err)
			}
		}
	}
	return nil
}

func (r *Reconciler) cleanup(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	if !clusterutils.IsClusterReady(cluster) {
		return nil
	}

	clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		klog.Warningf("failed to get cluster client: %s, it seems the cluster is not ready, skipping cleanup", err)
		return nil
	}
	kubeSphereNamespace := &corev1.Namespace{}
	if err = clusterClient.Get(ctx, client.ObjectKey{Name: constants.KubeSphereNamespace}, kubeSphereNamespace); err != nil {
		klog.Warningf("failed to get %s namespace: %s, it seems the cluster is not ready, skipping cleanup", constants.KubeSphereNamespace, err)
		return nil
	}
	delete(kubeSphereNamespace.Annotations, clusterv1alpha1.AnnotationClusterName)
	delete(kubeSphereNamespace.Annotations, clusterv1alpha1.AnnotationHostClusterName)
	return clusterClient.Update(ctx, kubeSphereNamespace)
}

func (r *Reconciler) createClusterAdmin(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	annotations := cluster.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		cluster.Annotations = annotations
	}
	if _, ok := annotations[initializedAnnotation]; ok {
		return nil
	}
	if creatorName, ok := annotations[constants.CreatorAnnotationKey]; ok {
		creator := &iamv1beta1.User{}
		if err := r.Get(ctx, types.NamespacedName{Name: creatorName}, creator); err != nil {
			return err
		}

		clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %s", err)
		}

		clusterAdminRole := iamv1beta1.ClusterAdmin
		clusterRoleBindingName := fmt.Sprintf("%s-%s", creator.Name, clusterAdminRole)
		if err = clusterClient.Get(ctx, types.NamespacedName{Name: clusterRoleBindingName}, &iamv1beta1.ClusterRoleBinding{}); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			clusterRoleBinding := iamv1beta1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterRoleBindingName,
					Labels: map[string]string{iamv1beta1.UserReferenceLabel: creator.Name,
						iamv1beta1.RoleReferenceLabel: clusterAdminRole},
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     iamv1beta1.ResourceKindUser,
						APIGroup: iamv1beta1.SchemeGroupVersion.Group,
						Name:     creator.Name,
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindClusterRole,
					Name:     clusterAdminRole,
				},
			}
			if err = clusterClient.Create(ctx, &clusterRoleBinding); err != nil {
				return err
			}
			annotations[initializedAnnotation] = metav1.NewTime(time.Now().UTC()).Format(time.RFC3339)
			if err = r.Update(ctx, cluster); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Reconciler) unbindWorkspaceTemplate(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	workspaceTemplates := tenantv1alpha1.WorkspaceTemplateList{}
	if err := r.List(ctx, &workspaceTemplates); err != nil {
		return err
	}
	for _, workspaceTemplate := range workspaceTemplates.Items {
		if workspaceTemplate.Spec.Placement.Clusters == nil || len(workspaceTemplate.Spec.Placement.Clusters) == 0 {
			continue
		}
		newClusters := make([]tenantv1alpha1.GenericClusterReference, 0, len(workspaceTemplate.Spec.Placement.Clusters))
		needUpdate := false
		for _, clusterReference := range workspaceTemplate.Spec.Placement.Clusters {
			if clusterReference.Name == cluster.Name {
				needUpdate = true
			} else {
				newClusters = append(newClusters, clusterReference)
			}
		}
		if !needUpdate {
			continue
		}
		workspaceTemplate.Spec.Placement.Clusters = newClusters
		if err := r.Update(ctx, &workspaceTemplate); err != nil {
			return nil
		}
	}
	return nil
}

func (r *Reconciler) installOrUpgradeKSCoreInMemberCluster(ctx context.Context,
	opt *options.HelmExecutorOptions, cluster *clusterv1alpha1.Cluster, jwtSecret, chartPath string) error {
	chartBytes, err := getChartBytes(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read chart files: %v", err)
	}

	valuesBytes, err := generateChartValueBytes(cluster.Spec.Config, jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to generate chart values: %v", err)
	}

	executorOptions := []helm.ExecutorOption{
		helm.SetExecutorLabels(map[string]string{
			constants.KubeSphereManagedLabel: "true",
		}),
		helm.SetExecutorOwner(&metav1.OwnerReference{
			APIVersion: clusterv1alpha1.SchemeGroupVersion.String(),
			Kind:       clusterv1alpha1.ResourceKindCluster,
			Name:       cluster.Name,
			UID:        cluster.UID,
		}),
		helm.SetExecutorImage(opt.Image),
		helm.SetExecutorNamespace(constants.KubeSphereNamespace),
		helm.SetExecutorBackoffLimit(0),
		helm.SetTTLSecondsAfterFinished(opt.JobTTLAfterFinished),
		helm.SetExecutorAffinity(opt.Affinity),
	}
	if opt.Resources != nil {
		executorOptions = append(executorOptions, helm.SetExecutorResources(corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(opt.Resources.Limits[corev1.ResourceCPU]),
				corev1.ResourceMemory: resource.MustParse(opt.Resources.Limits[corev1.ResourceMemory]),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(opt.Resources.Requests[corev1.ResourceCPU]),
				corev1.ResourceMemory: resource.MustParse(opt.Resources.Requests[corev1.ResourceMemory]),
			},
		}))
	}

	executor, err := helm.NewExecutor(executorOptions...)
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	jobName, err := executor.Upgrade(ctx, releaseName, releaseName, valuesBytes,
		helm.SetKubeconfig(cluster.Spec.Connection.KubeConfig),
		helm.SetNamespace(constants.KubeSphereNamespace),
		helm.SetChartData(chartBytes),
		helm.SetTimeout(5*time.Minute),
		helm.SetInstall(true),
		helm.SetCreateNamespace(true))
	if err != nil {
		return fmt.Errorf("failed to create executor job: %v", err)
	}
	klog.Infof("Install/Upgrade job %s created", jobName)

	if cluster.Annotations == nil {
		cluster.Annotations = make(map[string]string)
	}
	cluster.Annotations[installJobAnnotation] = jobName
	delete(cluster.Annotations, ksCoreActionAnnotation)
	cluster.Status.Conditions = []clusterv1alpha1.ClusterCondition{}
	if err := r.Update(ctx, cluster); err != nil {
		return fmt.Errorf("failed to update cluster %s: %v", cluster.Name, err)
	}

	return wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		job := &batchv1.Job{}
		if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: constants.KubeSphereNamespace}, job); err != nil {
			return false, err
		}
		if job.Status.Succeeded == 1 {
			return true, nil
		}
		if job.Status.Failed > 0 {
			return false, fmt.Errorf("job %s failed", jobName)
		}

		return false, nil
	})
}
