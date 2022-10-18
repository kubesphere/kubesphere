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
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/apiserver/config"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
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
// Also check if all of the clusters are ready by the spec.connection.kubeconfig every resync period

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15

	kubefedNamespace  = "kube-federation-system"
	kubesphereManaged = "kubesphere.io/managed"

	// Actually host cluster name can be anything, there is only necessary when calling JoinFederation function
	hostClusterName = "kubesphere"

	// proxy format
	proxyFormat = "%s/api/v1/namespaces/kubesphere-system/services/:ks-apiserver:80/proxy/%s"

	// mulitcluster configuration name
	configzMultiCluster = "multicluster"

	NotificationCleanup   = "notification.kubesphere.io/cleanup"
	notificationAPIFormat = "%s/apis/notification.kubesphere.io/v2beta2/%s/%s"
	secretAPIFormat       = "%s/api/v1/namespaces/%s/secrets/%s"
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

type clusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	// build this only for host cluster
	k8sClient  kubernetes.Interface
	hostConfig *rest.Config

	ksClient kubesphere.Interface

	clusterLister    clusterlister.ClusterLister
	userLister       iamv1alpha2listers.UserLister
	clusterHasSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	resyncPeriod time.Duration

	hostClusterName string
}

func NewClusterController(
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface,
	config *rest.Config,
	clusterInformer clusterinformer.ClusterInformer,
	userLister iamv1alpha2listers.UserLister,
	resyncPeriod time.Duration,
	hostClusterName string,
) *clusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &clusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		k8sClient:        k8sClient,
		ksClient:         ksClient,
		hostConfig:       config,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
		resyncPeriod:     resyncPeriod,
		hostClusterName:  hostClusterName,
		userLister:       userLister,
	}
	c.clusterLister = clusterInformer.Lister()
	c.clusterHasSynced = clusterInformer.Informer().HasSynced

	clusterInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCluster := oldObj.(*clusterv1alpha1.Cluster)
			newCluster := newObj.(*clusterv1alpha1.Cluster)
			if !reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) || newCluster.DeletionTimestamp != nil {
				c.enqueueCluster(newObj)
			}
		},
		DeleteFunc: c.enqueueCluster,
	}, resyncPeriod)

	return c
}

func (c *clusterController) Start(ctx context.Context) error {
	return c.Run(3, ctx.Done())
}

func (c *clusterController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(0).Info("starting cluster controller")
	defer klog.Info("shutting down cluster controller")

	if !cache.WaitForCacheSync(stopCh, c.clusterHasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	// refresh cluster configz every resync period
	go wait.Until(func() {
		if err := c.reconcileHostCluster(); err != nil {
			klog.Errorf("Error create host cluster, error %v", err)
		}

		if err := c.resyncClusters(); err != nil {
			klog.Errorf("failed to reconcile cluster ready status, err: %v", err)
		}
	}, c.resyncPeriod, stopCh)

	<-stopCh
	return nil
}

func (c *clusterController) worker() {
	for c.processNextItem() {
	}
}

func (c *clusterController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncCluster(key.(string))
	c.handleErr(err, key)
	return true
}

// reconcileHostCluster will create a host cluster if there are no clusters labeled 'cluster-role.kubesphere.io/host'
func (c *clusterController) reconcileHostCluster() error {
	clusters, err := c.clusterLister.List(labels.SelectorFromSet(labels.Set{clusterv1alpha1.HostCluster: ""}))
	if err != nil {
		return err
	}

	hostKubeConfig, err := buildKubeconfigFromRestConfig(c.hostConfig)
	if err != nil {
		return err
	}

	// no host cluster, create one
	if len(clusters) == 0 {
		hostCluster.Spec.Connection.KubeConfig = hostKubeConfig
		hostCluster.Name = c.hostClusterName
		_, err = c.ksClient.ClusterV1alpha1().Clusters().Create(context.TODO(), hostCluster, metav1.CreateOptions{})
		return err
	} else if len(clusters) > 1 {
		return fmt.Errorf("there MUST not be more than one host clusters, while there are %d", len(clusters))
	}

	// only deal with cluster managed by kubesphere
	cluster := clusters[0].DeepCopy()
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
	_, err = c.ksClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
	return err
}

func (c *clusterController) resyncClusters() error {
	clusters, err := c.clusterLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		key, _ := cache.MetaNamespaceKeyFunc(cluster)
		c.queue.Add(key)
	}

	return nil
}

func (c *clusterController) syncCluster(key string) error {
	klog.V(5).Infof("starting to sync cluster %s", key)
	startTime := time.Now()

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("not a valid controller key %s, %#v", key, err)
		return err
	}

	defer func() {
		klog.V(4).Infof("Finished syncing cluster %s in %s", name, time.Since(startTime))
	}()

	cluster, err := c.clusterLister.Get(name)
	if err != nil {
		// cluster not found, possibly been deleted
		// need to do the cleanup
		if errors.IsNotFound(err) {
			return nil
		}

		klog.Errorf("Failed to get cluster with name %s, %#v", name, err)
		return err
	}

	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
			if cluster, err = c.ksClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			// need to unJoin federation first, before there are
			// some cleanup work to do in member cluster which depends
			// agent to proxy traffic
			if err = c.unJoinFederation(nil, name); err != nil {
				klog.Errorf("Failed to unjoin federation for cluster %s, error %v", name, err)
				return err
			}

			// cleanup after cluster has been deleted
			if err := c.syncClusterMembers(nil, cluster); err != nil {
				klog.Errorf("Failed to sync cluster members for %s: %v", name, err)
				return err
			}

			if cluster.Annotations[NotificationCleanup] == "true" {
				if err := c.cleanupNotification(cluster); err != nil {
					klog.Errorf("Failed to cleanup notification config in cluster %s: %v", name, err)
					return err
				}
			}

			// remove our cluster finalizer
			finalizers := sets.NewString(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.List()
			if _, err = c.ksClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		return nil
	}

	// save a old copy of cluster
	oldCluster := cluster.DeepCopy()

	// currently we didn't set cluster.Spec.Enable when creating cluster at client side, so only check
	// if we enable cluster.Spec.JoinFederation now
	if !cluster.Spec.JoinFederation {
		klog.V(5).Infof("Skipping to join cluster %s cause it is not expected to join", cluster.Name)
		return nil
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		klog.V(5).Infof("Skipping to join cluster %s cause the kubeconfig is empty", cluster.Name)
		return nil
	}

	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create cluster config for %s: %s", cluster.Name, err)
	}

	clusterClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to create cluster client for %s: %s", cluster.Name, err)
	}

	proxyTransport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to create proxy transport for %s: %s", cluster.Name, err)
	}

	if !cluster.Spec.JoinFederation { // trying to unJoin federation
		err = c.unJoinFederation(clusterConfig, cluster.Name)
		if err != nil {
			klog.Errorf("Failed to unJoin federation for cluster %s, error %v", cluster.Name, err)
			c.eventRecorder.Event(cluster, v1.EventTypeWarning, "UnJoinFederation", err.Error())
			return err
		}
	} else { // join federation
		_, err = c.joinFederation(clusterConfig, cluster.Name, cluster.Labels)
		if err != nil {
			klog.Errorf("Failed to join federation for cluster %s, error %v", cluster.Name, err)

			federationNotReadyCondition := clusterv1alpha1.ClusterCondition{
				Type:               clusterv1alpha1.ClusterFederated,
				Status:             v1.ConditionFalse,
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             err.Error(),
				Message:            "Cluster can not join federation control plane",
			}
			c.updateClusterCondition(cluster, federationNotReadyCondition)
			notReadyCondition := clusterv1alpha1.ClusterCondition{
				Type:               clusterv1alpha1.ClusterReady,
				Status:             v1.ConditionFalse,
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "Cluster join federation control plane failed",
				Message:            "Cluster is Not Ready now",
			}
			c.updateClusterCondition(cluster, notReadyCondition)

			_, err = c.ksClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Failed to update cluster status, %#v", err)
			}

			return err
		}

		klog.Infof("successfully joined federation for cluster %s", cluster.Name)

		federationReadyCondition := clusterv1alpha1.ClusterCondition{
			Type:               clusterv1alpha1.ClusterFederated,
			Status:             v1.ConditionTrue,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "",
			Message:            "Cluster has joined federation control plane successfully",
		}

		c.updateClusterCondition(cluster, federationReadyCondition)
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
		return err
	}
	cluster.Status.KubernetesVersion = serverVersion.GitVersion

	nodes, err := clusterClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get cluster nodes, %#v", err)
		return err
	}
	cluster.Status.NodeCount = len(nodes.Items)

	// TODO use rest.Interface instead
	configz, err := c.tryToFetchKubeSphereComponents(clusterConfig.Host, proxyTransport)
	if err != nil {
		klog.Warningf("failed to fetch kubesphere components status in cluster %s: %s", cluster.Name, err)
	} else {
		cluster.Status.Configz = configz
	}

	// TODO use rest.Interface instead
	v, err := c.tryFetchKubeSphereVersion(clusterConfig.Host, proxyTransport)
	if err != nil {
		klog.Errorf("failed to get KubeSphere version, err: %#v", err)
	} else {
		cluster.Status.KubeSphereVersion = v
	}

	// Use kube-system namespace UID as cluster ID
	kubeSystem, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cluster.Status.UID = kubeSystem.UID

	// label cluster host cluster if configz["multicluster"]==true
	if mc, ok := configz[configzMultiCluster]; ok && mc && c.checkIfClusterIsHostCluster(nodes) {
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
	c.updateClusterCondition(cluster, readyCondition)

	if err = c.updateKubeConfigExpirationDateCondition(cluster); err != nil {
		klog.Errorf("sync KubeConfig expiration date for cluster %s failed: %v", cluster.Name, err)
		return err
	}

	if !reflect.DeepEqual(oldCluster.Status, cluster.Status) {
		_, err = c.ksClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Failed to update cluster status, %#v", err)
			return err
		}
	}

	if err = c.setClusterNameInConfigMap(clusterClient, cluster.Name); err != nil {
		return err
	}

	if err = c.syncClusterMembers(clusterClient, cluster); err != nil {
		return fmt.Errorf("failed to sync cluster membership for %s: %s", cluster.Name, err)
	}

	return nil
}

func (c *clusterController) setClusterNameInConfigMap(client kubernetes.Interface, name string) error {
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

func (c *clusterController) checkIfClusterIsHostCluster(memberClusterNodes *v1.NodeList) bool {
	hostNodes, err := c.k8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false
	}

	if hostNodes == nil || memberClusterNodes == nil {
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
func (c *clusterController) tryToFetchKubeSphereComponents(host string, transport http.RoundTripper) (map[string]bool, error) {
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

func (c *clusterController) tryFetchKubeSphereVersion(host string, transport http.RoundTripper) (string, error) {
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

func (c *clusterController) enqueueCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get cluster key %s failed", cluster.Name))
		return
	}

	c.queue.Add(key)
}

func (c *clusterController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		klog.V(2).Infof("Error syncing cluster %s, retrying, %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	klog.V(4).Infof("Dropping cluster %s out of the queue.", key)
	c.queue.Forget(key)
	utilruntime.HandleError(err)
}

// updateClusterCondition updates condition in cluster conditions using giving condition
// adds condition if not existed
func (c *clusterController) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
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

// joinFederation joins a cluster into federation clusters.
// return nil error if kubefed cluster already exists.
func (c *clusterController) joinFederation(clusterConfig *rest.Config, joiningClusterName string, labels map[string]string) (*fedv1b1.KubeFedCluster, error) {

	return joinClusterForNamespace(c.hostConfig,
		clusterConfig,
		kubefedNamespace,
		kubefedNamespace,
		hostClusterName,
		joiningClusterName,
		fmt.Sprintf("%s-secret", joiningClusterName),
		labels,
		apiextv1.ClusterScoped,
		false,
		false)
}

// unJoinFederation unjoins a cluster from federation control plane.
// It will first do normal unjoin process, if maximum retries reached, it will skip
// member cluster resource deletion, only delete resources in host cluster.
func (c *clusterController) unJoinFederation(clusterConfig *rest.Config, unjoiningClusterName string) error {
	localMaxRetries := 5
	retries := 0

	for {
		err := unjoinCluster(c.hostConfig,
			clusterConfig,
			kubefedNamespace,
			hostClusterName,
			unjoiningClusterName,
			true,
			false,
			false)
		if err != nil {
			klog.Errorf("Failed to unJoin federation for cluster %s, error %v", unjoiningClusterName, err)
		} else {
			return nil
		}

		retries += 1
		if retries >= localMaxRetries {
			err = unjoinCluster(c.hostConfig,
				clusterConfig,
				kubefedNamespace,
				hostClusterName,
				unjoiningClusterName,
				true,
				false,
				true)
			return err
		}
	}
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

func (c *clusterController) updateKubeConfigExpirationDateCondition(cluster *clusterv1alpha1.Cluster) error {
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

	c.updateClusterCondition(cluster, clusterv1alpha1.ClusterCondition{
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
func (c *clusterController) syncClusterMembers(clusterClient *kubernetes.Clientset, cluster *clusterv1alpha1.Cluster) error {
	users, err := c.userLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list users: %s", err)
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

	for _, user := range users {
		user = user.DeepCopy()
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
			if _, err := c.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("failed to update user %s: %s", user.Name, err)
			}
		}
	}
	return nil
}

func (c *clusterController) cleanupNotification(cluster *clusterv1alpha1.Cluster) error {

	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create cluster config for %s: %s", cluster.Name, err)
	}

	proxyTransport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to create proxy transport for %s: %s", cluster.Name, err)
	}

	client := http.Client{
		Transport: proxyTransport,
		Timeout:   5 * time.Second,
	}

	doDelete := func(kind, name string) error {
		url := fmt.Sprintf(notificationAPIFormat, clusterConfig.Host, kind, name)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("failed to delete notification %s %s in cluster %s, %s", kind, name, cluster.Name, resp.Status)
		}

		return nil
	}

	if fedConfigs, err := c.ksClient.TypesV1beta2().FederatedNotificationConfigs().List(context.Background(), metav1.ListOptions{}); err != nil {
		return err
	} else {
		for _, fedConfig := range fedConfigs.Items {
			if err := doDelete("configs", fedConfig.Name); err != nil {
				return err
			}
		}
	}

	if fedReceivers, err := c.ksClient.TypesV1beta2().FederatedNotificationReceivers().List(context.Background(), metav1.ListOptions{}); err != nil {
		return err
	} else {
		for _, fedReceiver := range fedReceivers.Items {
			if err := doDelete("receivers", fedReceiver.Name); err != nil {
				return err
			}
		}
	}

	if fedRouters, err := c.ksClient.TypesV1beta2().FederatedNotificationRouters().List(context.Background(), metav1.ListOptions{}); err != nil {
		return err
	} else {
		for _, fedRouter := range fedRouters.Items {
			if err := doDelete("routers", fedRouter.Name); err != nil {
				return err
			}
		}
	}

	if fedSilences, err := c.ksClient.TypesV1beta2().FederatedNotificationSilences().List(context.Background(), metav1.ListOptions{}); err != nil {
		return err
	} else {
		for _, fedSilence := range fedSilences.Items {
			if err := doDelete("silences", fedSilence.Name); err != nil {
				return err
			}
		}
	}

	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			constants.NotificationManagedLabel: "true",
		},
	}
	if secrets, err := c.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&selector)}); err != nil {
		return err
	} else {
		for _, secret := range secrets.Items {
			url := fmt.Sprintf(secretAPIFormat, clusterConfig.Host, constants.NotificationSecretNamespace, secret.Name)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
				return fmt.Errorf("failed to delete notification secret %s in cluster %s, %s", secret.Name, cluster.Name, resp.Status)
			}
		}
	}

	return nil
}
