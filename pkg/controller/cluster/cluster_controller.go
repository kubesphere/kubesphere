package cluster

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	clusterclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"time"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type ClusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	agentClient   clusterclient.AgentInterface
	clusterClient clusterclient.ClusterInterface

	agentLister    clusterlister.AgentLister
	agentHasSynced cache.InformerSynced

	clusterLister    clusterlister.ClusterLister
	clusterHasSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewClusterController(
	client kubernetes.Interface,
	clusterInformer clusterinformer.ClusterInformer,
	agentInformer clusterinformer.AgentInformer,
	agentClient clusterclient.AgentInterface,
	clusterClient clusterclient.ClusterInterface,
) *ClusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &ClusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		agentClient:      agentClient,
		clusterClient:    clusterClient,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
	}

	c.agentLister = agentInformer.Lister()
	c.agentHasSynced = agentInformer.Informer().HasSynced

	c.clusterLister = clusterInformer.Lister()
	c.clusterHasSynced = clusterInformer.Informer().HasSynced

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.addCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.addCluster(newObj)
		},
		DeleteFunc: c.addCluster,
	})

	agentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
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

	if !cache.WaitForCacheSync(stopCh, c.clusterHasSynced, c.agentHasSynced) {
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
			_, err = c.agentLister.Get(name)
			if err != nil && errors.IsNotFound(err) {
				return nil
			}

			if err != nil {
				klog.Errorf("Failed to get cluster agent %s, %#v", name, err)
				return err
			}

			// do the real cleanup work
			err = c.agentClient.Delete(name, &metav1.DeleteOptions{})
			return err
		}

		klog.Errorf("Failed to get cluster with name %s, %#v", name, err)
		return err
	}

	newAgent := &clusterv1alpha1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "tower",
				"cluster.kubesphere.io/name": name,
			},
		},
		Spec: clusterv1alpha1.AgentSpec{
			Token:                   "",
			KubeSphereAPIServerPort: 0,
			KubernetesAPIServerPort: 0,
			Proxy:                   "",
			Paused:                  !cluster.Spec.Active,
		},
	}

	agent, err := c.agentLister.Get(name)
	if err != nil && errors.IsNotFound(err) {
		agent, err = c.agentClient.Create(newAgent)
		if err != nil {
			klog.Errorf("Failed to create agent %s, %#v", name, err)
			return err
		}

		return nil
	}

	if err != nil {
		klog.Errorf("Failed to get agent %s, %#v", name, err)
		return err
	}

	if agent.Spec.Paused != newAgent.Spec.Paused {
		agent.Spec.Paused = newAgent.Spec.Paused
		return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			_, err = c.agentClient.Update(agent)
			return err
		})
	}

	// agent connection is ready, update cluster status
	// set
	if len(agent.Status.KubeConfig) != 0 && c.isAgentReady(agent) {
		clientConfig, err := clientcmd.NewClientConfigFromBytes(agent.Status.KubeConfig)
		if err != nil {
			klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
			return err
		}

		config, err := clientConfig.ClientConfig()
		if err != nil {
			klog.Errorf("Failed to get client config, %#v", err)
			return err
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Errorf("Failed to create ClientSet from config, %#v", err)
			return nil
		}

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
	}

	agentReadyCondition := clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterAgentAvailable,
		LastUpdateTime:     metav1.NewTime(time.Now()),
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             "",
		Message:            "Cluster agent is available now.",
	}

	if c.isAgentReady(agent) {
		agentReadyCondition.Status = v1.ConditionTrue
	} else {
		agentReadyCondition.Status = v1.ConditionFalse
	}

	c.addClusterCondition(cluster, agentReadyCondition)

	_, err = c.clusterClient.Update(cluster)
	if err != nil {
		klog.Errorf("Failed to update cluster status, %#v", err)
		return err
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
		klog.V(2).Infof("Error syncing virtualservice %s for service retrying, %#v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	klog.V(4).Infof("Dropping service %s out of the queue.", key)
	c.queue.Forget(key)
	utilruntime.HandleError(err)
}

func (c *ClusterController) addAgent(obj interface{}) {
	agent := obj.(*clusterv1alpha1.Agent)
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get agent key %s failed", agent.Name))
		return
	}

	c.queue.Add(key)
}

func (c *ClusterController) isAgentReady(agent *clusterv1alpha1.Agent) bool {
	for _, condition := range agent.Status.Conditions {
		if condition.Type == clusterv1alpha1.AgentConnected && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// addClusterCondition add condition
func (c *ClusterController) addClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
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
