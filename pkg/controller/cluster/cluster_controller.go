/*
Copyright 2020 KubeSphere Authors

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

package cluster

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/multicluster"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/version"
)

// Cluster controller only runs under multicluster mode. Cluster controller is following below steps,
//   1. Wait for cluster agent is ready if connection type is proxy
//   2. Join cluster into federation control plane if kubeconfig is ready.
//   3. Pull cluster version and configz, set result to cluster status
// Also put all clusters back into queue every 5 * time.Minute to sync cluster status, this is needed
// in case there aren't any cluster changes made.
// Also check if all the clusters are ready by the spec.connection.kubeconfig every resync period

const (
	kubesphereManaged = "kubesphere.io/managed"

	// proxy format
	proxyFormat = "%s/api/v1/namespaces/kubesphere-system/services/:ks-apiserver:80/proxy/%s"

	// mulitcluster configuration name
	configzMultiCluster = "multicluster"
)

// Cluster template for reconcile host cluster if there is none.
var hostCluster = &clusterv1alpha1.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: "host",
		Annotations: map[string]string{
			"kubesphere.io/description": "The description was created by KubeSphere automatically. " +
				"It is recommended that you use the Host Cluster to manage clusters only " +
				"and deploy workloads on Member Clusters.",
		},
		Labels: map[string]string{
			clusterv1alpha1.HostCluster: "",
			kubesphereManaged:           "true",
		},
	},
	Spec: clusterv1alpha1.ClusterSpec{
		JoinFederation: true,
		Enable:         true,
		Provider:       "kubesphere",
		Connection: clusterv1alpha1.Connection{
			Type: clusterv1alpha1.ConnectionTypeDirect,
		},
	},
}

type Reconciler struct {
	client.Client

	HostConfig      *rest.Config
	HostClusterName string
	ResyncPeriod    time.Duration
}

// InjectClient is used to inject the client into NodeReconciler.
func (r *Reconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

// SetupWithManager setups the NodeReconciler with manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return builder.
		ControllerManagedBy(mgr).
		For(
			&clusterv1alpha1.Cluster{},
			builder.WithPredicates(
				clusterChangedPredicate{},
			),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 3,
		}).
		Complete(r)
}

type clusterChangedPredicate struct {
	predicate.Funcs
}

func (clusterChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldCluster := e.ObjectOld.(*clusterv1alpha1.Cluster)
	newCluster := e.ObjectNew.(*clusterv1alpha1.Cluster)
	if !reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) || newCluster.DeletionTimestamp != nil {
		return true
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
		if err := r.reconcileHostCluster(); err != nil {
			klog.Errorf("Error create host cluster, error %v", err)
		}

		if err := r.resyncClusters(); err != nil {
			klog.Errorf("failed to reconcile cluster ready status, err: %v", err)
		}
	}, r.ResyncPeriod, ctx.Done())
	return nil
}

// Reconcile reconciles the Node object.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(5).Infof("starting to sync cluster %s", req.Name)
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing cluster %s in %s", req.Name, time.Since(startTime))
	}()

	cluster := &clusterv1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
			if err := r.Update(ctx, cluster); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			// cleanup after cluster has been deleted
			if err := r.syncClusterMembers(ctx, nil, cluster); err != nil {
				klog.Errorf("Failed to sync cluster members for %s: %v", req.Name, err)
				return ctrl.Result{}, err
			}

			// remove our cluster finalizer
			finalizers := sets.NewString(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.List()
			if err := r.Update(ctx, cluster); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		klog.V(5).Infof("Skipping to join cluster %s cause the kubeconfig is empty", cluster.Name)
		return ctrl.Result{}, nil
	}

	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create cluster config for %s: %s", cluster.Name, err)
	}

	clusterClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create cluster client for %s: %s", cluster.Name, err)
	}

	proxyTransport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create proxy transport for %s: %s", cluster.Name, err)
	}

	// cluster is ready, we can pull kubernetes cluster info through agent
	// since there is no agent necessary for host cluster, so updates for host cluster
	// is safe.
	if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 {
		cluster.Spec.Connection.KubernetesAPIEndpoint = clusterConfig.Host
	}

	serverVersion, err := clusterClient.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("Failed to get kubernetes version, %#v", err)
		return ctrl.Result{}, err
	}
	cluster.Status.KubernetesVersion = serverVersion.GitVersion

	nodes, err := clusterClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get cluster nodes, %#v", err)
		return ctrl.Result{}, err
	}
	cluster.Status.NodeCount = len(nodes.Items)

	// TODO use rest.Interface instead
	configz, err := r.tryToFetchKubeSphereComponents(clusterConfig.Host, proxyTransport)
	if err != nil {
		klog.Warningf("failed to fetch kubesphere components status in cluster %s: %s", cluster.Name, err)
	} else {
		cluster.Status.Configz = configz
	}

	// TODO use rest.Interface instead
	v, err := r.tryFetchKubeSphereVersion(clusterConfig.Host, proxyTransport)
	if err != nil {
		klog.Errorf("failed to get KubeSphere version, err: %#v", err)
	} else {
		cluster.Status.KubeSphereVersion = v
	}

	// Use kube-system namespace UID as cluster ID
	kubeSystem, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}
	cluster.Status.UID = kubeSystem.UID

	// label cluster host cluster if configz["multicluster"]==true
	if mc, ok := configz[configzMultiCluster]; ok && mc && r.checkIfClusterIsHostCluster(ctx, nodes) {
		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string)
		}
		cluster.Labels[clusterv1alpha1.HostCluster] = ""
	}

	readyCondition := clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterReady,
		Status:             v1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterReady),
		Message:            "Cluster is available now",
	}
	r.updateClusterCondition(cluster, readyCondition)

	if err = r.updateKubeConfigExpirationDateCondition(cluster); err != nil {
		klog.Errorf("sync KubeConfig expiration date for cluster %s failed: %v", cluster.Name, err)
		return ctrl.Result{}, err
	}

	if err = r.Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.setClusterNameInConfigMap(clusterClient, cluster.Name); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.syncClusterMembers(ctx, clusterClient, cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync cluster membership for %s: %s", cluster.Name, err)
	}

	return ctrl.Result{}, nil
}

// reconcileHostCluster will create a host cluster if there are no clusters labeled 'cluster-role.kubesphere.io/host'
func (r *Reconciler) reconcileHostCluster() error {
	clusters := &clusterv1alpha1.ClusterList{}
	if err := r.List(context.TODO(), clusters); err != nil {
		return err
	}

	hostKubeConfig, err := buildKubeconfigFromRestConfig(r.HostConfig)
	if err != nil {
		return err
	}

	// no host cluster, create one
	if len(clusters.Items) == 0 {
		hostCluster.Spec.Connection.KubeConfig = hostKubeConfig
		hostCluster.Name = r.HostClusterName
		return r.Create(context.TODO(), hostCluster)
	} else if len(clusters.Items) > 1 {
		return fmt.Errorf("there MUST not be more than one host clusters, while there are %d", len(clusters.Items))
	}

	// only deal with cluster managed by kubesphere
	cluster := clusters.Items[0].DeepCopy()
	managedByKubesphere, ok := cluster.Labels[kubesphereManaged]
	if !ok || managedByKubesphere != "true" {
		return nil
	}

	// no kubeconfig, not likely to happen
	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		cluster.Spec.Connection.KubeConfig = hostKubeConfig
	} else {
		// if kubeconfig are the same, then there is nothing to do
		if bytes.Equal(cluster.Spec.Connection.KubeConfig, hostKubeConfig) {
			return nil
		}
	}

	// update host cluster config
	return r.Update(context.TODO(), cluster)
}

func (r *Reconciler) resyncClusters() error {
	clusters := &clusterv1alpha1.ClusterList{}
	if err := r.List(context.TODO(), clusters); err != nil {
		return err
	}

	for _, cluster := range clusters.Items {
		if _, err := r.Reconcile(context.TODO(), ctrl.Request{NamespacedName: types.NamespacedName{Name: cluster.Name}}); err != nil {
			klog.Errorf("resync cluster %s failed: %v", cluster.Name, err)
		}
	}
	return nil
}

func (r *Reconciler) setClusterNameInConfigMap(client kubernetes.Interface, name string) error {
	cm, err := client.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Get(context.TODO(), constants.KubeSphereConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configData, err := config.GetFromConfigMap(cm)
	if err != nil {
		return err
	}
	if configData.MultiClusterOptions == nil {
		configData.MultiClusterOptions = &multicluster.Options{}
	}
	if configData.MultiClusterOptions.ClusterName == name {
		return nil
	}

	configData.MultiClusterOptions.ClusterName = name
	newConfigData, err := yaml.Marshal(configData)
	if err != nil {
		return err
	}
	cm.Data[constants.KubeSphereConfigMapDataKey] = string(newConfigData)
	if _, err = client.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) checkIfClusterIsHostCluster(ctx context.Context, memberClusterNodes *v1.NodeList) bool {
	hostNodes := &v1.NodeList{}
	if err := r.List(ctx, hostNodes); err != nil {
		return false
	}

	if hostNodes.Items == nil || memberClusterNodes == nil {
		return false
	}

	if len(hostNodes.Items) != len(memberClusterNodes.Items) {
		return false
	}

	if len(hostNodes.Items) > 0 && (hostNodes.Items[0].Status.NodeInfo.MachineID != memberClusterNodes.Items[0].Status.NodeInfo.MachineID) {
		return false
	}

	return true
}

// tryToFetchKubeSphereComponents will send requests to member cluster configz api using kube-apiserver proxy way
func (r *Reconciler) tryToFetchKubeSphereComponents(host string, transport http.RoundTripper) (map[string]bool, error) {
	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	response, err := client.Get(fmt.Sprintf(proxyFormat, host, "kapis/config.kubesphere.io/v1alpha2/configs/configz"))
	if err != nil {
		klog.V(4).Infof("Failed to get kubesphere components, error %v", err)
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		klog.V(4).Infof("Response status code isn't 200.")
		return nil, fmt.Errorf("response code %d", response.StatusCode)
	}

	configz := make(map[string]bool)
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&configz)
	if err != nil {
		klog.V(4).Infof("Decode error %v", err)
		return nil, err
	}
	return configz, nil
}

func (r *Reconciler) tryFetchKubeSphereVersion(host string, transport http.RoundTripper) (string, error) {
	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	response, err := client.Get(fmt.Sprintf(proxyFormat, host, "kapis/version"))
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		klog.V(4).Infof("Response status code isn't 200.")
		return "", fmt.Errorf("response code %d", response.StatusCode)
	}

	info := version.Info{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&info)
	if err != nil {
		return "", err
	}

	// currently, we kubesphere v2.1 can not be joined as a member cluster and it will never be reconciled,
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

func parseKubeConfigExpirationDate(kubeconfig []byte) (time.Time, error) {
	config, err := k8sutil.LoadKubeConfigFromBytes(kubeconfig)
	if err != nil {
		return time.Time{}, err
	}
	if config.CertData == nil {
		return time.Time{}, fmt.Errorf("empty CertData")
	}
	block, _ := pem.Decode(config.CertData)
	if block == nil {
		return time.Time{}, fmt.Errorf("pem.Decode failed, got empty block data")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func (r *Reconciler) updateKubeConfigExpirationDateCondition(cluster *clusterv1alpha1.Cluster) error {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return nil
	}
	// we don't need to check member clusters which using proxy mode, their certs are managed and will be renewed by tower.
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
		return nil
	}

	klog.V(5).Infof("sync KubeConfig expiration date for cluster %s", cluster.Name)
	notAfter, err := parseKubeConfigExpirationDate(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return fmt.Errorf("parseKubeConfigExpirationDate for cluster %s failed: %v", cluster.Name, err)
	}
	expiresInSevenDays := v1.ConditionFalse
	if time.Now().AddDate(0, 0, 7).Sub(notAfter) > 0 {
		expiresInSevenDays = v1.ConditionTrue
	}

	r.updateClusterCondition(cluster, clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays,
		Status:             expiresInSevenDays,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays),
		Message:            notAfter.String(),
	})
	return nil
}

// syncClusterMembers Sync granted clusters for users periodically
func (r *Reconciler) syncClusterMembers(ctx context.Context, clusterClient *kubernetes.Clientset, cluster *clusterv1alpha1.Cluster) error {
	users := &iamv1alpha2.UserList{}
	if err := r.List(ctx, users); err != nil {
		return err
	}

	grantedUsers := sets.NewString()
	clusterName := cluster.Name
	if cluster.DeletionTimestamp.IsZero() {
		list, err := clusterClient.RbacV1().ClusterRoleBindings().List(context.Background(),
			metav1.ListOptions{LabelSelector: iamv1alpha2.UserReferenceLabel})
		if err != nil {
			return fmt.Errorf("failed to list clusterrolebindings: %s", err)
		}
		for _, clusterRoleBinding := range list.Items {
			for _, sub := range clusterRoleBinding.Subjects {
				if sub.Kind == iamv1alpha2.ResourceKindUser {
					grantedUsers.Insert(sub.Name)
				}
			}
		}
	}

	for i := range users.Items {
		user := &users.Items[i]
		grantedClustersAnnotation := user.Annotations[iamv1alpha2.GrantedClustersAnnotation]
		var grantedClusters sets.String
		if len(grantedClustersAnnotation) > 0 {
			grantedClusters = sets.NewString(strings.Split(grantedClustersAnnotation, ",")...)
		} else {
			grantedClusters = sets.NewString()
		}
		if grantedUsers.Has(user.Name) && !grantedClusters.Has(clusterName) {
			grantedClusters.Insert(clusterName)
		} else if !grantedUsers.Has(user.Name) && grantedClusters.Has(clusterName) {
			grantedClusters.Delete(clusterName)
		}
		grantedClustersAnnotation = strings.Join(grantedClusters.List(), ",")
		if user.Annotations[iamv1alpha2.GrantedClustersAnnotation] != grantedClustersAnnotation {
			if user.Annotations == nil {
				user.Annotations = make(map[string]string, 0)
			}
			user.Annotations[iamv1alpha2.GrantedClustersAnnotation] = grantedClustersAnnotation
			if err := r.Update(ctx, user); err != nil {
				return err
			}
		}
	}
	return nil
}
