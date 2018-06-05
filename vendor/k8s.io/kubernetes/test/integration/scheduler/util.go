/*
Copyright 2017 The Kubernetes Authors.

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

package scheduler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/admission"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilfeaturetesting "k8s.io/apiserver/pkg/util/feature/testing"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	clientv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/scheduler"
	_ "k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/factory"
	"k8s.io/kubernetes/test/integration/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"
)

type TestContext struct {
	closeFn                framework.CloseFunc
	httpServer             *httptest.Server
	ns                     *v1.Namespace
	clientSet              *clientset.Clientset
	informerFactory        informers.SharedInformerFactory
	schedulerConfigFactory scheduler.Configurator
	schedulerConfig        *scheduler.Config
	scheduler              *scheduler.Scheduler
}

// createConfiguratorWithPodInformer creates a configurator for scheduler.
func createConfiguratorWithPodInformer(
	schedulerName string,
	clientSet clientset.Interface,
	podInformer coreinformers.PodInformer,
	informerFactory informers.SharedInformerFactory,
) scheduler.Configurator {
	return factory.NewConfigFactory(
		schedulerName,
		clientSet,
		informerFactory.Core().V1().Nodes(),
		podInformer,
		informerFactory.Core().V1().PersistentVolumes(),
		informerFactory.Core().V1().PersistentVolumeClaims(),
		informerFactory.Core().V1().ReplicationControllers(),
		informerFactory.Extensions().V1beta1().ReplicaSets(),
		informerFactory.Apps().V1beta1().StatefulSets(),
		informerFactory.Core().V1().Services(),
		informerFactory.Policy().V1beta1().PodDisruptionBudgets(),
		informerFactory.Storage().V1().StorageClasses(),
		v1.DefaultHardPodAffinitySymmetricWeight,
		utilfeature.DefaultFeatureGate.Enabled(features.EnableEquivalenceClassCache),
		false,
	)
}

// initTestMasterAndScheduler initializes a test environment and creates a master with default
// configuration.
func initTestMaster(t *testing.T, nsPrefix string, admission admission.Interface) *TestContext {
	var context TestContext

	// 1. Create master
	h := &framework.MasterHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	masterConfig := framework.NewIntegrationTestMasterConfig()

	if admission != nil {
		masterConfig.GenericConfig.AdmissionControl = admission
	}

	_, context.httpServer, context.closeFn = framework.RunAMasterUsingServer(masterConfig, s, h)

	if nsPrefix != "default" {
		context.ns = framework.CreateTestingNamespace(nsPrefix+string(uuid.NewUUID()), s, t)
	} else {
		context.ns = framework.CreateTestingNamespace("default", s, t)
	}

	// 2. Create kubeclient
	context.clientSet = clientset.NewForConfigOrDie(
		&restclient.Config{
			QPS: -1, Host: s.URL,
			ContentConfig: restclient.ContentConfig{
				GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"},
			},
		},
	)
	return &context
}

// initTestScheduler initializes a test environment and creates a scheduler with default
// configuration.
func initTestScheduler(
	t *testing.T,
	context *TestContext,
	controllerCh chan struct{},
	setPodInformer bool,
	policy *schedulerapi.Policy,
) *TestContext {
	// Pod preemption is enabled by default scheduler configuration, but preemption only happens when PodPriority
	// feature gate is enabled at the same time.
	return initTestSchedulerWithOptions(t, context, controllerCh, setPodInformer, policy, false)
}

// initTestSchedulerWithOptions initializes a test environment and creates a scheduler with default
// configuration and other options.
func initTestSchedulerWithOptions(
	t *testing.T,
	context *TestContext,
	controllerCh chan struct{},
	setPodInformer bool,
	policy *schedulerapi.Policy,
	disablePreemption bool,
) *TestContext {
	// Enable EnableEquivalenceClassCache for all integration tests.
	defer utilfeaturetesting.SetFeatureGateDuringTest(
		t,
		utilfeature.DefaultFeatureGate,
		features.EnableEquivalenceClassCache, true)()

	// 1. Create scheduler
	context.informerFactory = informers.NewSharedInformerFactory(context.clientSet, time.Second)

	var podInformer coreinformers.PodInformer

	// create independent pod informer if required
	if setPodInformer {
		podInformer = factory.NewPodInformer(context.clientSet, 12*time.Hour)
	} else {
		podInformer = context.informerFactory.Core().V1().Pods()
	}

	context.schedulerConfigFactory = createConfiguratorWithPodInformer(
		v1.DefaultSchedulerName, context.clientSet, podInformer, context.informerFactory)

	var err error

	if policy != nil {
		context.schedulerConfig, err = context.schedulerConfigFactory.CreateFromConfig(*policy)
	} else {
		context.schedulerConfig, err = context.schedulerConfigFactory.Create()
	}

	if err != nil {
		t.Fatalf("Couldn't create scheduler config: %v", err)
	}

	// set controllerCh if provided.
	if controllerCh != nil {
		context.schedulerConfig.StopEverything = controllerCh
	}

	// set DisablePreemption option
	context.schedulerConfig.DisablePreemption = disablePreemption

	// set setPodInformer if provided.
	if setPodInformer {
		go podInformer.Informer().Run(context.schedulerConfig.StopEverything)
	}

	eventBroadcaster := record.NewBroadcaster()
	context.schedulerConfig.Recorder = eventBroadcaster.NewRecorder(
		legacyscheme.Scheme,
		v1.EventSource{Component: v1.DefaultSchedulerName},
	)
	eventBroadcaster.StartRecordingToSink(&clientv1core.EventSinkImpl{
		Interface: context.clientSet.CoreV1().Events(""),
	})

	context.informerFactory.Start(context.schedulerConfig.StopEverything)
	context.informerFactory.WaitForCacheSync(context.schedulerConfig.StopEverything)

	context.scheduler, err = scheduler.NewFromConfigurator(&scheduler.FakeConfigurator{
		Config: context.schedulerConfig},
		nil...)
	if err != nil {
		t.Fatalf("Couldn't create scheduler: %v", err)
	}
	context.scheduler.Run()
	return context
}

// initTest initializes a test environment and creates master and scheduler with default
// configuration.
func initTest(t *testing.T, nsPrefix string) *TestContext {
	return initTestScheduler(t, initTestMaster(t, nsPrefix, nil), nil, true, nil)
}

// initTestDisablePreemption initializes a test environment and creates master and scheduler with default
// configuration but with pod preemption disabled.
func initTestDisablePreemption(t *testing.T, nsPrefix string) *TestContext {
	return initTestSchedulerWithOptions(
		t, initTestMaster(t, nsPrefix, nil), nil, true, nil, true)
}

// cleanupTest deletes the scheduler and the test namespace. It should be called
// at the end of a test.
func cleanupTest(t *testing.T, context *TestContext) {
	// Kill the scheduler.
	close(context.schedulerConfig.StopEverything)
	// Cleanup nodes.
	context.clientSet.CoreV1().Nodes().DeleteCollection(nil, metav1.ListOptions{})
	framework.DeleteTestingNamespace(context.ns, context.httpServer, t)
	context.closeFn()
}

// waitForReflection waits till the passFunc confirms that the object it expects
// to see is in the store. Used to observe reflected events.
func waitForReflection(t *testing.T, nodeLister corelisters.NodeLister, key string,
	passFunc func(n interface{}) bool) error {
	nodes := []*v1.Node{}
	err := wait.Poll(time.Millisecond*100, wait.ForeverTestTimeout, func() (bool, error) {
		n, err := nodeLister.Get(key)

		switch {
		case err == nil && passFunc(n):
			return true, nil
		case errors.IsNotFound(err):
			nodes = append(nodes, nil)
		case err != nil:
			t.Errorf("Unexpected error: %v", err)
		default:
			nodes = append(nodes, n)
		}

		return false, nil
	})
	if err != nil {
		t.Logf("Logging consecutive node versions received from store:")
		for i, n := range nodes {
			t.Logf("%d: %#v", i, n)
		}
	}
	return err
}

// nodeHasLabels returns a function that checks if a node has all the given labels.
func nodeHasLabels(cs clientset.Interface, nodeName string, labels map[string]string) wait.ConditionFunc {
	return func() (bool, error) {
		node, err := cs.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			// This could be a connection error so we want to retry.
			return false, nil
		}
		for k, v := range labels {
			if node.Labels == nil || node.Labels[k] != v {
				return false, nil
			}
		}
		return true, nil
	}
}

// waitForNodeLabels waits for the given node to have all the given labels.
func waitForNodeLabels(cs clientset.Interface, nodeName string, labels map[string]string) error {
	return wait.Poll(time.Millisecond*100, wait.ForeverTestTimeout, nodeHasLabels(cs, nodeName, labels))
}

// createNode creates a node with the given resource list and
// returns a pointer and error status. If 'res' is nil, a predefined amount of
// resource will be used.
func createNode(cs clientset.Interface, name string, res *v1.ResourceList) (*v1.Node, error) {
	// if resource is nil, we use a default amount of resources for the node.
	if res == nil {
		res = &v1.ResourceList{
			v1.ResourcePods: *resource.NewQuantity(32, resource.DecimalSI),
		}
	}
	n := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1.NodeSpec{Unschedulable: false},
		Status: v1.NodeStatus{
			Capacity: *res,
		},
	}
	return cs.CoreV1().Nodes().Create(n)
}

// updateNodeStatus updates the status of node.
func updateNodeStatus(cs clientset.Interface, node *v1.Node) error {
	_, err := cs.CoreV1().Nodes().UpdateStatus(node)
	return err
}

// createNodes creates `numNodes` nodes. The created node names will be in the
// form of "`prefix`-X" where X is an ordinal.
func createNodes(cs clientset.Interface, prefix string, res *v1.ResourceList, numNodes int) ([]*v1.Node, error) {
	nodes := make([]*v1.Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeName := fmt.Sprintf("%v-%d", prefix, i)
		node, err := createNode(cs, nodeName, res)
		if err != nil {
			return nodes[:], err
		}
		nodes[i] = node
	}
	return nodes[:], nil
}

type pausePodConfig struct {
	Name                              string
	Namespace                         string
	Affinity                          *v1.Affinity
	Annotations, Labels, NodeSelector map[string]string
	Resources                         *v1.ResourceRequirements
	Tolerations                       []v1.Toleration
	NodeName                          string
	SchedulerName                     string
	Priority                          *int32
}

// initPausePod initializes a pod API object from the given config. It is used
// mainly in pod creation process.
func initPausePod(cs clientset.Interface, conf *pausePodConfig) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        conf.Name,
			Namespace:   conf.Namespace,
			Labels:      conf.Labels,
			Annotations: conf.Annotations,
		},
		Spec: v1.PodSpec{
			NodeSelector: conf.NodeSelector,
			Affinity:     conf.Affinity,
			Containers: []v1.Container{
				{
					Name:  conf.Name,
					Image: imageutils.GetPauseImageName(),
				},
			},
			Tolerations:   conf.Tolerations,
			NodeName:      conf.NodeName,
			SchedulerName: conf.SchedulerName,
			Priority:      conf.Priority,
		},
	}
	if conf.Resources != nil {
		pod.Spec.Containers[0].Resources = *conf.Resources
	}
	return pod
}

// createPausePod creates a pod with "Pause" image and the given config and
// return its pointer and error status.
func createPausePod(cs clientset.Interface, p *v1.Pod) (*v1.Pod, error) {
	return cs.CoreV1().Pods(p.Namespace).Create(p)
}

// createPausePodWithResource creates a pod with "Pause" image and the given
// resources and returns its pointer and error status. The resource list can be
// nil.
func createPausePodWithResource(cs clientset.Interface, podName string,
	nsName string, res *v1.ResourceList) (*v1.Pod, error) {
	var conf pausePodConfig
	if res == nil {
		conf = pausePodConfig{
			Name:      podName,
			Namespace: nsName,
		}
	} else {
		conf = pausePodConfig{
			Name:      podName,
			Namespace: nsName,
			Resources: &v1.ResourceRequirements{
				Requests: *res,
			},
		}
	}
	return createPausePod(cs, initPausePod(cs, &conf))
}

// runPausePod creates a pod with "Pause" image and the given config and waits
// until it is scheduled. It returns its pointer and error status.
func runPausePod(cs clientset.Interface, pod *v1.Pod) (*v1.Pod, error) {
	pod, err := cs.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		return nil, fmt.Errorf("Error creating pause pod: %v", err)
	}
	if err = waitForPodToSchedule(cs, pod); err != nil {
		return pod, fmt.Errorf("Pod %v didn't schedule successfully. Error: %v", pod.Name, err)
	}
	if pod, err = cs.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
		return pod, fmt.Errorf("Error getting pod %v info: %v", pod.Name, err)
	}
	return pod, nil
}

// podDeleted returns true if a pod is not found in the given namespace.
func podDeleted(c clientset.Interface, podNamespace, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if pod.DeletionTimestamp != nil {
			return true, nil
		}
		return false, nil
	}
}

// podIsGettingEvicted returns true if the pod's deletion timestamp is set.
func podIsGettingEvicted(c clientset.Interface, podNamespace, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod.DeletionTimestamp != nil {
			return true, nil
		}
		return false, nil
	}
}

// podScheduled returns true if a node is assigned to the given pod.
func podScheduled(c clientset.Interface, podNamespace, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			// This could be a connection error so we want to retry.
			return false, nil
		}
		if pod.Spec.NodeName == "" {
			return false, nil
		}
		return true, nil
	}
}

// podUnschedulable returns a condition function that returns true if the given pod
// gets unschedulable status.
func podUnschedulable(c clientset.Interface, podNamespace, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			// This could be a connection error so we want to retry.
			return false, nil
		}
		_, cond := podutil.GetPodCondition(&pod.Status, v1.PodScheduled)
		return cond != nil && cond.Status == v1.ConditionFalse &&
			cond.Reason == v1.PodReasonUnschedulable, nil
	}
}

// waitForPodToScheduleWithTimeout waits for a pod to get scheduled and returns
// an error if it does not scheduled within the given timeout.
func waitForPodToScheduleWithTimeout(cs clientset.Interface, pod *v1.Pod, timeout time.Duration) error {
	return wait.Poll(100*time.Millisecond, timeout, podScheduled(cs, pod.Namespace, pod.Name))
}

// waitForPodToSchedule waits for a pod to get scheduled and returns an error if
// it does not get scheduled within the timeout duration (30 seconds).
func waitForPodToSchedule(cs clientset.Interface, pod *v1.Pod) error {
	return waitForPodToScheduleWithTimeout(cs, pod, 30*time.Second)
}

// waitForPodUnscheduleWithTimeout waits for a pod to fail scheduling and returns
// an error if it does not become unschedulable within the given timeout.
func waitForPodUnschedulableWithTimeout(cs clientset.Interface, pod *v1.Pod, timeout time.Duration) error {
	return wait.Poll(100*time.Millisecond, timeout, podUnschedulable(cs, pod.Namespace, pod.Name))
}

// waitForPodUnschedule waits for a pod to fail scheduling and returns
// an error if it does not become unschedulable within the timeout duration (30 seconds).
func waitForPodUnschedulable(cs clientset.Interface, pod *v1.Pod) error {
	return waitForPodUnschedulableWithTimeout(cs, pod, 30*time.Second)
}

// deletePod deletes the given pod in the given namespace.
func deletePod(cs clientset.Interface, podName string, nsName string) error {
	return cs.CoreV1().Pods(nsName).Delete(podName, metav1.NewDeleteOptions(0))
}

// cleanupPods deletes the given pods and waits for them to be actually deleted.
func cleanupPods(cs clientset.Interface, t *testing.T, pods []*v1.Pod) {
	for _, p := range pods {
		err := cs.CoreV1().Pods(p.Namespace).Delete(p.Name, metav1.NewDeleteOptions(0))
		if err != nil && !errors.IsNotFound(err) {
			t.Errorf("error while deleting pod %v/%v: %v", p.Namespace, p.Name, err)
		}
	}
	for _, p := range pods {
		if err := wait.Poll(time.Second, wait.ForeverTestTimeout,
			podDeleted(cs, p.Namespace, p.Name)); err != nil {
			t.Errorf("error while waiting for pod  %v/%v to get deleted: %v", p.Namespace, p.Name, err)
		}
	}
}
