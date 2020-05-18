package cluster

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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
	"math/rand"
	"reflect"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	"time"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15

	kubefedNamespace = "kube-federation-system"

	hostClusterName = "kubesphere"

	// allocate kubernetesAPIServer port in range [portRangeMin, portRangeMax] for agents if port is not specified
	// kubesphereAPIServer port is defaulted to kubernetesAPIServerPort + 10000
	portRangeMin = 6000
	portRangeMax = 7000

	// Service port
	kubernetesPort = 6443
	kubespherePort = 80

	defaultAgentNamespace = "kubesphere-system"
)

type ClusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	client     kubernetes.Interface
	hostConfig *rest.Config

	clusterClient clusterclient.ClusterInterface

	clusterLister    clusterlister.ClusterLister
	clusterHasSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewClusterController(
	client kubernetes.Interface,
	config *rest.Config,
	clusterInformer clusterinformer.ClusterInformer,
	clusterClient clusterclient.ClusterInterface,
) *ClusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &ClusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		client:           client,
		hostConfig:       config,
		clusterClient:    clusterClient,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
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

func (c *ClusterController) Start(stopCh <-chan struct{}) error {
	return c.Run(5, stopCh)
}

func (c *ClusterController) Run(workers int, stopCh <-chan struct{}) error {
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

	<-stopCh
	return nil
}

func (c *ClusterController) worker() {
	for c.processNextItem() {
	}
}

func (c *ClusterController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncCluster(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *ClusterController) syncCluster(key string) error {
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

			finalizers := sets.NewString(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.List()
			if _, err = c.clusterClient.Update(cluster); err != nil {
				return err
			}
		}
		return nil
	}

	oldCluster := cluster.DeepCopy()

	// prepare for proxy to member cluster
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
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
		if err != nil {
			if errors.IsNotFound(err) {
				service, err = c.client.CoreV1().Services(defaultAgentNamespace).Create(&mcService)
				if err != nil {
					return err
				}
			}

			return err
		} else {
			if !reflect.DeepEqual(service.Spec, mcService.Spec) {
				mcService.ObjectMeta = service.ObjectMeta
				mcService.Spec.ClusterIP = service.Spec.ClusterIP

				service, err = c.client.CoreV1().Services(defaultAgentNamespace).Update(&mcService)
				if err != nil {
					return err
				}
			}
		}

		// populated the kubernetes apiEndpoint and kubesphere apiEndpoint
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

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		return nil
	}

	var clientSet kubernetes.Interface
	var clusterConfig *rest.Config

	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return err
	}

	clusterConfig, err = clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return err
	}

	clientSet, err = kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create ClientSet from config, %#v", err)
		return nil
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
		version, err := clientSet.Discovery().ServerVersion()
		if err != nil {
			klog.Errorf("Failed to get kubernetes version, %#v", err)
			return err
		}

		cluster.Status.KubernetesVersion = version.GitVersion

		nodes, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to get cluster nodes, %#v", err)
			return err
		}

		cluster.Status.NodeCount = len(nodes.Items)

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

	if !reflect.DeepEqual(oldCluster, cluster) {
		_, err = c.clusterClient.Update(cluster)
		if err != nil {
			klog.Errorf("Failed to update cluster status, %#v", err)
			return err
		}
	}

	return nil
}

func (c *ClusterController) addCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get cluster key %s failed", cluster.Name))
		return
	}

	c.queue.Add(key)
}

func (c *ClusterController) handleErr(err error, key interface{}) {
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
func (c *ClusterController) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
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

func isHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	for k, v := range cluster.Annotations {
		if k == clusterv1alpha1.IsHostCluster && v == "true" {
			return true
		}
	}

	return false
}

// joinFederation joins a cluster into federation clusters.
// return nil error if kubefed cluster already exists.
func (c *ClusterController) joinFederation(clusterConfig *rest.Config, joiningClusterName string, labels map[string]string) (*fedv1b1.KubeFedCluster, error) {

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
func (c *ClusterController) unJoinFederation(clusterConfig *rest.Config, unjoiningClusterName string) error {
	return unjoinCluster(c.hostConfig,
		clusterConfig,
		kubefedNamespace,
		hostClusterName,
		unjoiningClusterName,
		true,
		false)
}

// allocatePort find a available port between [portRangeMin, portRangeMax] in maximumRetries
// TODO: only works with handful clusters
func (c *ClusterController) allocatePort() (uint16, error) {
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
func (c *ClusterController) generateToken() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
