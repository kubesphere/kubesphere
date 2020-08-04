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
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	clusterclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"math/rand"
	"net/http"
	"reflect"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	"sync"
	"time"
)

// Cluster controller only runs under multicluster mode. Cluster controller is following below steps,
//   1. Populates proxy spec if cluster connection type is proxy
//     1.1 Wait for cluster agent is ready if connection type is proxy
//   2. Join cluster into federation control plane if kubeconfig is ready.
//   3. Pull cluster version and configz, set result to cluster status
// Also put all clusters back into queue every 5 * time.Minute to sync cluster status, this is needed
// in case there aren't any cluster changes made.

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15

	kubefedNamespace  = "kube-federation-system"
	openpitrixRuntime = "openpitrix.io/runtime"
	kubesphereManaged = "kubesphere.io/managed"

	// Actually host cluster name can be anything, there is only necessary when calling JoinFederation function
	hostClusterName = "kubesphere"

	// allocate kubernetesAPIServer port in range [portRangeMin, portRangeMax] for agents if port is not specified
	// kubesphereAPIServer port is defaulted to kubernetesAPIServerPort + 10000
	portRangeMin = 6000
	portRangeMax = 7000

	// Proxy service port
	kubernetesPort = 6443
	kubespherePort = 80

	defaultAgentNamespace = "kubesphere-system"

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
			"kubesphere.io/description": "Automatically created by kubesphere, " +
				"we encourage you to use host cluster for clusters management only, " +
				"deploy workloads to member clusters.",
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

// ClusterData stores cluster client
type clusterData struct {

	// cached rest.Config
	config *rest.Config

	// cached kubernetes client, rebuild once cluster changed
	client kubernetes.Interface

	// cached kubeconfig
	cachedKubeconfig []byte

	// cached transport, used to proxy kubesphere version request
	transport http.RoundTripper
}

type clusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	// build this only for host cluster
	client     kubernetes.Interface
	hostConfig *rest.Config

	clusterClient clusterclient.ClusterInterface

	clusterLister    clusterlister.ClusterLister
	clusterHasSynced cache.InformerSynced

	openpitrixClient openpitrix.Client

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	mu sync.RWMutex

	clusterMap map[string]*clusterData
}

func NewClusterController(
	client kubernetes.Interface,
	config *rest.Config,
	clusterInformer clusterinformer.ClusterInformer,
	clusterClient clusterclient.ClusterInterface,
	openpitrixClient openpitrix.Client,
) *clusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &clusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		client:           client,
		hostConfig:       config,
		clusterClient:    clusterClient,
		openpitrixClient: openpitrixClient,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
		clusterMap:       make(map[string]*clusterData),
	}

	c.clusterLister = clusterInformer.Lister()
	c.clusterHasSynced = clusterInformer.Informer().HasSynced

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.addCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			newCluster := newObj.(*clusterv1alpha1.Cluster)
			oldCluster := oldObj.(*clusterv1alpha1.Cluster)
			if newCluster.ResourceVersion == oldCluster.ResourceVersion {
				return
			}
			c.addCluster(newObj)
		},
		DeleteFunc: c.addCluster,
	})

	return c
}

func (c *clusterController) Start(stopCh <-chan struct{}) error {
	return c.Run(3, stopCh)
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

	// refresh cluster configz every 2 minutes
	go wait.Until(func() {
		if err := c.syncStatus(); err != nil {
			klog.Errorf("Error periodically sync cluster status, %v", err)
		}

		if err := c.reconcileHostCluster(); err != nil {
			klog.Errorf("Error create host cluster, error %v", err)
		}

	}, 2*time.Minute, stopCh)

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

func buildClusterData(kubeconfig []byte) (*clusterData, error) {
	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return nil, err
	}

	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return nil, err
	}

	transport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create transport, %#v", err)
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create ClientSet from config, %#v", err)
		return nil, err
	}

	return &clusterData{
		cachedKubeconfig: kubeconfig,
		config:           clusterConfig,
		client:           clientSet,
		transport:        transport,
	}, nil
}

func (c *clusterController) syncStatus() error {
	clusters, err := c.clusterLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		key, err := cache.MetaNamespaceKeyFunc(cluster)
		if err != nil {
			return err
		}

		c.queue.AddRateLimited(key)
	}

	return nil
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
		_, err = c.clusterClient.Create(hostCluster)
		return err
	} else if len(clusters) > 1 {
		return fmt.Errorf("there MUST not be more than one host clusters, while there are %d", len(clusters))
	}

	// only deal with cluster managed by kubesphere
	cluster := clusters[0]
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
	_, err = c.clusterClient.Update(cluster)
	return err
}

func (c *clusterController) syncCluster(key string) error {
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

	// proxy service name if needed
	serviceName := fmt.Sprintf("mc-%s", cluster.Name)

	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
			if cluster, err = c.clusterClient.Update(cluster); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			// need to unJoin federation first, before there are
			// some cleanup work to do in member cluster which depends
			// agent to proxy traffic
			err = c.unJoinFederation(nil, name)
			if err != nil {
				klog.Errorf("Failed to unjoin federation for cluster %s, error %v", name, err)
				return err
			}

			_, err = c.client.CoreV1().Services(defaultAgentNamespace).Get(serviceName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// nothing to do
				} else {
					klog.Errorf("Failed to get proxy service %s, error %v", serviceName, err)
					return err
				}
			} else {
				err = c.client.CoreV1().Services(defaultAgentNamespace).Delete(serviceName, metav1.NewDeleteOptions(0))
				if err != nil {
					klog.Errorf("Unable to delete service %s, error %v", serviceName, err)
					return err
				}
			}

			// clean up openpitrix runtime of the cluster
			if _, ok := cluster.Annotations[openpitrixRuntime]; ok {
				if c.openpitrixClient != nil {
					err = c.openpitrixClient.CleanupRuntime(cluster.Name)
					if err != nil {
						klog.Errorf("Unable to delete openpitrix for cluster %s, error %v", cluster.Name, err)
						return err
					}
				}
				delete(cluster.Annotations, openpitrixRuntime)
			}

			// remove our cluster finalizer
			finalizers := sets.NewString(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.List()
			if _, err = c.clusterClient.Update(cluster); err != nil {
				return err
			}
		}
		return nil
	}

	// save a old copy of cluster
	oldCluster := cluster.DeepCopy()

	// prepare for proxy to member cluster
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {

		// allocate ports for kubernetes and kubesphere endpoint
		if cluster.Spec.Connection.KubeSphereAPIServerPort == 0 ||
			cluster.Spec.Connection.KubernetesAPIServerPort == 0 {
			port, err := c.allocatePort()
			if err != nil {
				klog.Error(err)
				return err
			}

			cluster.Spec.Connection.KubernetesAPIServerPort = port
			cluster.Spec.Connection.KubeSphereAPIServerPort = port + 10000
		}

		// token uninitialized, generate a new token
		if len(cluster.Spec.Connection.Token) == 0 {
			cluster.Spec.Connection.Token = c.generateToken()
		}

		// create a proxy service spec
		mcService := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: cluster.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name": serviceName,
					"app":                    serviceName,
				},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"app.kubernetes.io/name": "tower",
					"app":                    "tower",
				},
				Ports: []v1.ServicePort{
					{
						Name:       "kubernetes",
						Protocol:   v1.ProtocolTCP,
						Port:       kubernetesPort,
						TargetPort: intstr.FromInt(int(cluster.Spec.Connection.KubernetesAPIServerPort)),
					},
					{
						Name:       "kubesphere",
						Protocol:   v1.ProtocolTCP,
						Port:       kubespherePort,
						TargetPort: intstr.FromInt(int(cluster.Spec.Connection.KubeSphereAPIServerPort)),
					},
				},
			},
		}

		service, err := c.client.CoreV1().Services(defaultAgentNamespace).Get(serviceName, metav1.GetOptions{})
		if err != nil { // proxy service not found
			if errors.IsNotFound(err) {
				service, err = c.client.CoreV1().Services(defaultAgentNamespace).Create(&mcService)
				if err != nil {
					return err
				}
			}

			return err
		} else { // update existed proxy service
			if !reflect.DeepEqual(service.Spec, mcService.Spec) {
				mcService.ObjectMeta = service.ObjectMeta
				mcService.Spec.ClusterIP = service.Spec.ClusterIP

				service, err = c.client.CoreV1().Services(defaultAgentNamespace).Update(&mcService)
				if err != nil {
					return err
				}
			}
		}

		// populates the kubernetes apiEndpoint and kubesphere apiEndpoint
		cluster.Spec.Connection.KubernetesAPIEndpoint = fmt.Sprintf("https://%s:%d", service.Spec.ClusterIP, kubernetesPort)
		cluster.Spec.Connection.KubeSphereAPIEndpoint = fmt.Sprintf("http://%s:%d", service.Spec.ClusterIP, kubespherePort)

		initializedCondition := clusterv1alpha1.ClusterCondition{
			Type:               clusterv1alpha1.ClusterInitialized,
			Status:             v1.ConditionTrue,
			Reason:             string(clusterv1alpha1.ClusterInitialized),
			Message:            "Cluster has been initialized",
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}
		c.updateClusterCondition(cluster, initializedCondition)

		if !reflect.DeepEqual(oldCluster, cluster) {
			cluster, err = c.clusterClient.Update(cluster)
			if err != nil {
				klog.Errorf("Error updating cluster %s, error %s", cluster.Name, err)
				return err
			}
			return nil
		}
	}

	// kubeconfig not ready, nothing to do
	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		return nil
	}

	// build up cached cluster data if there isn't any
	c.mu.Lock()
	clusterDt, ok := c.clusterMap[cluster.Name]
	if !ok || clusterDt == nil || !equality.Semantic.DeepEqual(clusterDt.cachedKubeconfig, cluster.Spec.Connection.KubeConfig) {
		clusterDt, err = buildClusterData(cluster.Spec.Connection.KubeConfig)
		if err != nil {
			c.mu.Unlock()
			return err
		}
		c.clusterMap[cluster.Name] = clusterDt
	}
	c.mu.Unlock()

	if !cluster.Spec.JoinFederation { // trying to unJoin federation
		err = c.unJoinFederation(clusterDt.config, cluster.Name)
		if err != nil {
			klog.Errorf("Failed to unJoin federation for cluster %s, error %v", cluster.Name, err)
			c.eventRecorder.Event(cluster, v1.EventTypeWarning, "UnJoinFederation", err.Error())
			return err
		}
	} else { // join federation
		_, err = c.joinFederation(clusterDt.config, cluster.Name, cluster.Labels)
		if err != nil {
			klog.Errorf("Failed to join federation for cluster %s, error %v", cluster.Name, err)
			c.eventRecorder.Event(cluster, v1.EventTypeWarning, "JoinFederation", err.Error())
			return err
		}
		c.eventRecorder.Event(cluster, v1.EventTypeNormal, "JoinFederation", "Cluster has joined federation.")

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

	// cluster agent is ready, we can pull kubernetes cluster info through agent
	// since there is no agent necessary for host cluster, so updates for host cluster
	// is safe.
	if isConditionTrue(cluster, clusterv1alpha1.ClusterAgentAvailable) ||
		cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeDirect {

		if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 {
			cluster.Spec.Connection.KubernetesAPIEndpoint = clusterDt.config.Host
		}

		version, err := clusterDt.client.Discovery().ServerVersion()
		if err != nil {
			klog.Errorf("Failed to get kubernetes version, %#v", err)
			return err
		}

		cluster.Status.KubernetesVersion = version.GitVersion

		nodes, err := clusterDt.client.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to get cluster nodes, %#v", err)
			return err
		}

		cluster.Status.NodeCount = len(nodes.Items)

		configz, err := c.tryToFetchKubeSphereComponents(clusterDt.config.Host, clusterDt.transport)
		if err == nil {
			cluster.Status.Configz = configz
		}

		// label cluster host cluster if configz["multicluster"]==true, this is
		if mc, ok := configz[configzMultiCluster]; ok && mc && c.checkIfClusterIsHostCluster(nodes) {
			if cluster.Labels == nil {
				cluster.Labels = make(map[string]string)
			}
			cluster.Labels[clusterv1alpha1.HostCluster] = ""
		}

		clusterReadyCondition := clusterv1alpha1.ClusterCondition{
			Type:               clusterv1alpha1.ClusterReady,
			Status:             v1.ConditionTrue,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(clusterv1alpha1.ClusterReady),
			Message:            "Cluster is available now",
		}

		c.updateClusterCondition(cluster, clusterReadyCondition)
	}

	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy &&
		!isConditionTrue(cluster, clusterv1alpha1.ClusterAgentAvailable) {
		clusterNotReadyCondition := clusterv1alpha1.ClusterCondition{
			Type:               clusterv1alpha1.ClusterReady,
			Status:             v1.ConditionFalse,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "Unable to establish connection with cluster",
			Message:            "Cluster is not available now",
		}

		c.updateClusterCondition(cluster, clusterNotReadyCondition)
	}

	if c.openpitrixClient != nil { // OpenPitrix is enabled, create runtime
		if cluster.GetAnnotations() == nil {
			cluster.Annotations = make(map[string]string)
		}

		if _, ok = cluster.Annotations[openpitrixRuntime]; !ok {
			err = c.openpitrixClient.UpsertRuntime(cluster.Name, string(cluster.Spec.Connection.KubeConfig))
			if err != nil {
				klog.Errorf("Failed to create runtime for cluster %s, error %v", cluster.Name, err)
				return err
			} else {
				cluster.Annotations[openpitrixRuntime] = cluster.Name
			}
		}
	}

	if !reflect.DeepEqual(oldCluster, cluster) {
		_, err = c.clusterClient.Update(cluster)
		if err != nil {
			klog.Errorf("Failed to update cluster status, %#v", err)
			return err
		}
	}

	return nil
}

func (c *clusterController) checkIfClusterIsHostCluster(memberClusterNodes *v1.NodeList) bool {
	hostNodes, err := c.client.CoreV1().Nodes().List(metav1.ListOptions{})
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

func (c *clusterController) addCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get cluster key %s failed", cluster.Name))
		return
	}

	c.queue.Add(key)
}

func hasHostClusterLabel(cluster *clusterv1alpha1.Cluster) bool {
	if cluster.Labels == nil || len(cluster.Labels) == 0 {
		return false
	}

	_, ok := cluster.Labels[clusterv1alpha1.HostCluster]

	return ok
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

// isConditionTrue checks cluster specific condition value is True, return false if condition not exists
func isConditionTrue(cluster *clusterv1alpha1.Cluster, conditionType clusterv1alpha1.ClusterConditionType) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == conditionType && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// updateClusterCondition updates condition in cluster conditions using giving condition
// adds condition if not existed
func (c *clusterController) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = make([]clusterv1alpha1.ClusterCondition, 0)
	}

	newConditions := make([]clusterv1alpha1.ClusterCondition, 0)
	needToUpdate := true
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == condition.Type {
			if cond.Status == condition.Status {
				needToUpdate = false
				continue
			} else {
				newConditions = append(newConditions, cond)
			}
		}
		newConditions = append(newConditions, cond)
	}

	if needToUpdate {
		newConditions = append(newConditions, condition)
		cluster.Status.Conditions = newConditions
	}
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
		apiextv1b1.ClusterScoped,
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

// allocatePort find a available port between [portRangeMin, portRangeMax] in maximumRetries
// TODO: only works with handful clusters
func (c *clusterController) allocatePort() (uint16, error) {
	rand.Seed(time.Now().UnixNano())

	clusters, err := c.clusterLister.List(labels.Everything())
	if err != nil {
		return 0, err
	}

	const maximumRetries = 10
	for i := 0; i < maximumRetries; i++ {
		collision := false
		port := uint16(portRangeMin + rand.Intn(portRangeMax-portRangeMin+1))

		for _, item := range clusters {
			if item.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy &&
				item.Spec.Connection.KubernetesAPIServerPort != 0 &&
				item.Spec.Connection.KubeSphereAPIServerPort == port {
				collision = true
				break
			}
		}

		if !collision {
			return port, nil
		}
	}

	return 0, fmt.Errorf("unable to allocate port after %d retries", maximumRetries)
}

// generateToken returns a random 32-byte string as token
func (c *clusterController) generateToken() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
