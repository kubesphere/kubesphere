/*
Copyright 2015 The Kubernetes Authors.

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

package scalability

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/apis/batch"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/timer"
	testutils "k8s.io/kubernetes/test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

const (
	smallGroupSize  = 5
	mediumGroupSize = 30
	bigGroupSize    = 250
	smallGroupName  = "load-small"
	mediumGroupName = "load-medium"
	bigGroupName    = "load-big"
	// We start RCs/Services/pods/... in different namespace in this test.
	// nodeCountPerNamespace determines how many namespaces we will be using
	// depending on the number of nodes in the underlying cluster.
	nodeCountPerNamespace = 100
	// How many threads will be used to create/delete services during this test.
	serviceOperationsParallelism = 1
	svcLabelKey                  = "svc-label"
)

var randomKind = schema.GroupKind{Kind: "Random"}

var knownKinds = []schema.GroupKind{
	api.Kind("ReplicationController"),
	extensions.Kind("Deployment"),
	// TODO: uncomment when Jobs are fixed: #38497
	//batch.Kind("Job"),
	extensions.Kind("ReplicaSet"),
}

// This test suite can take a long time to run, so by default it is added to
// the ginkgo.skip list (see driver.go).
// To run this suite you must explicitly ask for it by setting the
// -t/--test flag or ginkgo.focus flag.
var _ = SIGDescribe("Load capacity", func() {
	var clientset clientset.Interface
	var nodeCount int
	var ns string
	var configs []testutils.RunObjectConfig
	var secretConfigs []*testutils.SecretConfig
	var configMapConfigs []*testutils.ConfigMapConfig

	testCaseBaseName := "load"
	var testPhaseDurations *timer.TestPhaseTimer
	var profileGathererStopCh chan struct{}

	// Gathers metrics before teardown
	// TODO add flag that allows to skip cleanup on failure
	AfterEach(func() {
		// Stop apiserver CPU profile gatherer and gather memory allocations profile.
		close(profileGathererStopCh)
		wg := sync.WaitGroup{}
		wg.Add(1)
		framework.GatherApiserverMemoryProfile(&wg, "load")
		wg.Wait()

		// Verify latency metrics
		highLatencyRequests, metrics, err := framework.HighLatencyRequests(clientset, nodeCount)
		framework.ExpectNoError(err)
		if err == nil {
			summaries := make([]framework.TestDataSummary, 0, 2)
			summaries = append(summaries, metrics)
			summaries = append(summaries, testPhaseDurations)
			framework.PrintSummaries(summaries, testCaseBaseName)
			Expect(highLatencyRequests).NotTo(BeNumerically(">", 0), "There should be no high-latency requests")
		}
	})

	// We assume a default throughput of 10 pods/second throughput.
	// We may want to revisit it in the future.
	// However, this can be overridden by LOAD_TEST_THROUGHPUT env var.
	throughput := 10
	if throughputEnv := os.Getenv("LOAD_TEST_THROUGHPUT"); throughputEnv != "" {
		if newThroughput, err := strconv.Atoi(throughputEnv); err == nil {
			throughput = newThroughput
		}
	}

	// Explicitly put here, to delete namespace at the end of the test
	// (after measuring latency metrics, etc.).
	options := framework.FrameworkOptions{
		ClientQPS:   float32(math.Max(50.0, float64(2*throughput))),
		ClientBurst: int(math.Max(100.0, float64(4*throughput))),
	}
	f := framework.NewFramework(testCaseBaseName, options, nil)
	f.NamespaceDeletionTimeout = time.Hour

	BeforeEach(func() {
		testPhaseDurations = timer.NewTestPhaseTimer()
		clientset = f.ClientSet

		ns = f.Namespace.Name
		nodes := framework.GetReadySchedulableNodesOrDie(clientset)
		nodeCount = len(nodes.Items)
		Expect(nodeCount).NotTo(BeZero())

		// Terminating a namespace (deleting the remaining objects from it - which
		// generally means events) can affect the current run. Thus we wait for all
		// terminating namespace to be finally deleted before starting this test.
		err := framework.CheckTestingNSDeletedExcept(clientset, ns)
		framework.ExpectNoError(err)

		framework.ExpectNoError(framework.ResetMetrics(clientset))

		// Start apiserver CPU profile gatherer with frequency based on cluster size.
		profileGatheringDelay := time.Duration(5+nodeCount/100) * time.Minute
		profileGathererStopCh = framework.StartApiserverCPUProfileGatherer(profileGatheringDelay)
	})

	type Load struct {
		podsPerNode int
		image       string
		command     []string
		// What kind of resource we want to create
		kind             schema.GroupKind
		services         bool
		secretsPerPod    int
		configMapsPerPod int
		daemonsPerNode   int
		quotas           bool
	}

	loadTests := []Load{
		// The container will consume 1 cpu and 512mb of memory.
		{podsPerNode: 3, image: "jess/stress", command: []string{"stress", "-c", "1", "-m", "2"}, kind: api.Kind("ReplicationController")},
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: api.Kind("ReplicationController")},
		// Tests for other resource types
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: extensions.Kind("Deployment")},
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: batch.Kind("Job")},
		// Test scheduling when daemons are preset
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: api.Kind("ReplicationController"), daemonsPerNode: 2},
		// Test with secrets
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: extensions.Kind("Deployment"), secretsPerPod: 2},
		// Test with configmaps
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: extensions.Kind("Deployment"), configMapsPerPod: 2},
		// Special test case which randomizes created resources
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: randomKind},
		// Test with quotas
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: api.Kind("ReplicationController"), quotas: true},
		{podsPerNode: 30, image: framework.ServeHostnameImage, kind: randomKind, quotas: true},
	}

	isCanonical := func(test *Load) bool {
		return test.podsPerNode == 30 && test.kind == api.Kind("ReplicationController") && test.daemonsPerNode == 0 && test.secretsPerPod == 0 && test.configMapsPerPod == 0 && !test.quotas
	}

	for _, testArg := range loadTests {
		feature := "ManualPerformance"
		if isCanonical(&testArg) {
			feature = "Performance"
		}
		name := fmt.Sprintf("[Feature:%s] should be able to handle %v pods per node %v with %v secrets, %v configmaps and %v daemons",
			feature,
			testArg.podsPerNode,
			testArg.kind,
			testArg.secretsPerPod,
			testArg.configMapsPerPod,
			testArg.daemonsPerNode,
		)
		if testArg.quotas {
			name += " with quotas"
		}
		itArg := testArg
		itArg.services = os.Getenv("CREATE_SERVICES") != "false"

		It(name, func() {
			// Create a number of namespaces.
			namespaceCount := (nodeCount + nodeCountPerNamespace - 1) / nodeCountPerNamespace
			namespaces, err := CreateNamespaces(f, namespaceCount, fmt.Sprintf("load-%v-nodepods", itArg.podsPerNode), testPhaseDurations.StartPhase(110, "namespace creation"))
			framework.ExpectNoError(err)

			totalPods := (itArg.podsPerNode - itArg.daemonsPerNode) * nodeCount
			configs, secretConfigs, configMapConfigs = generateConfigs(totalPods, itArg.image, itArg.command, namespaces, itArg.kind, itArg.secretsPerPod, itArg.configMapsPerPod)

			if itArg.quotas {
				framework.ExpectNoError(CreateQuotas(f, namespaces, 2*totalPods, testPhaseDurations.StartPhase(115, "quota creation")))
			}

			serviceCreationPhase := testPhaseDurations.StartPhase(120, "services creation")
			defer serviceCreationPhase.End()
			if itArg.services {
				framework.Logf("Creating services")
				services := generateServicesForConfigs(configs)
				createService := func(i int) {
					defer GinkgoRecover()
					framework.ExpectNoError(testutils.CreateServiceWithRetries(clientset, services[i].Namespace, services[i]))
				}
				workqueue.Parallelize(serviceOperationsParallelism, len(services), createService)
				framework.Logf("%v Services created.", len(services))
				defer func(services []*v1.Service) {
					serviceCleanupPhase := testPhaseDurations.StartPhase(800, "services deletion")
					defer serviceCleanupPhase.End()
					framework.Logf("Starting to delete services...")
					deleteService := func(i int) {
						defer GinkgoRecover()
						framework.ExpectNoError(testutils.DeleteResourceWithRetries(clientset, api.Kind("Service"), services[i].Namespace, services[i].Name, nil))
					}
					workqueue.Parallelize(serviceOperationsParallelism, len(services), deleteService)
					framework.Logf("Services deleted")
				}(services)
			} else {
				framework.Logf("Skipping service creation")
			}
			serviceCreationPhase.End()
			// Create all secrets.
			secretsCreationPhase := testPhaseDurations.StartPhase(130, "secrets creation")
			defer secretsCreationPhase.End()
			for i := range secretConfigs {
				secretConfigs[i].Run()
				defer secretConfigs[i].Stop()
			}
			secretsCreationPhase.End()
			// Create all configmaps.
			configMapsCreationPhase := testPhaseDurations.StartPhase(140, "configmaps creation")
			defer configMapsCreationPhase.End()
			for i := range configMapConfigs {
				configMapConfigs[i].Run()
				defer configMapConfigs[i].Stop()
			}
			configMapsCreationPhase.End()
			// StartDaemon if needed
			daemonSetCreationPhase := testPhaseDurations.StartPhase(150, "daemonsets creation")
			defer daemonSetCreationPhase.End()
			for i := 0; i < itArg.daemonsPerNode; i++ {
				daemonName := fmt.Sprintf("load-daemon-%v", i)
				daemonConfig := &testutils.DaemonConfig{
					Client:    f.ClientSet,
					Name:      daemonName,
					Namespace: f.Namespace.Name,
					LogFunc:   framework.Logf,
				}
				daemonConfig.Run()
				defer func(config *testutils.DaemonConfig) {
					framework.ExpectNoError(framework.DeleteResourceAndPods(
						f.ClientSet,
						f.InternalClientset,
						f.ScalesGetter,
						extensions.Kind("DaemonSet"),
						config.Namespace,
						config.Name,
					))
				}(daemonConfig)
			}
			daemonSetCreationPhase.End()

			// Simulate lifetime of RC:
			//  * create with initial size
			//  * scale RC to a random size and list all pods
			//  * scale RC to a random size and list all pods
			//  * delete it
			//
			// This will generate ~5 creations/deletions per second assuming:
			//  - X small RCs each 5 pods   [ 5 * X = totalPods / 2 ]
			//  - Y medium RCs each 30 pods [ 30 * Y = totalPods / 4 ]
			//  - Z big RCs each 250 pods   [ 250 * Z = totalPods / 4]

			// We would like to spread creating replication controllers over time
			// to make it possible to create/schedule them in the meantime.
			// Currently we assume <throughput> pods/second average throughput.
			// We may want to revisit it in the future.
			framework.Logf("Starting to create %v objects...", itArg.kind)
			creatingTime := time.Duration(totalPods/throughput) * time.Second

			createAllResources(configs, creatingTime, testPhaseDurations.StartPhase(200, "load pods creation"))
			By("============================================================================")

			// We would like to spread scaling replication controllers over time
			// to make it possible to create/schedule & delete them in the meantime.
			// Currently we assume that <throughput> pods/second average throughput.
			// The expected number of created/deleted pods is less than totalPods/3.
			scalingTime := time.Duration(totalPods/(3*throughput)) * time.Second
			framework.Logf("Starting to scale %v objects first time...", itArg.kind)
			scaleAllResources(configs, scalingTime, testPhaseDurations.StartPhase(300, "scaling first time"))
			By("============================================================================")

			framework.Logf("Starting to scale %v objects second time...", itArg.kind)
			scaleAllResources(configs, scalingTime, testPhaseDurations.StartPhase(400, "scaling second time"))
			By("============================================================================")

			// Cleanup all created replication controllers.
			// Currently we assume <throughput> pods/second average deletion throughput.
			// We may want to revisit it in the future.
			deletingTime := time.Duration(totalPods/throughput) * time.Second
			framework.Logf("Starting to delete %v objects...", itArg.kind)
			deleteAllResources(configs, deletingTime, testPhaseDurations.StartPhase(500, "load pods deletion"))
		})
	}
})

func createClients(numberOfClients int) ([]clientset.Interface, []internalclientset.Interface, []scaleclient.ScalesGetter, error) {
	clients := make([]clientset.Interface, numberOfClients)
	internalClients := make([]internalclientset.Interface, numberOfClients)
	scalesClients := make([]scaleclient.ScalesGetter, numberOfClients)

	for i := 0; i < numberOfClients; i++ {
		config, err := framework.LoadConfig()
		Expect(err).NotTo(HaveOccurred())
		config.QPS = 100
		config.Burst = 200
		if framework.TestContext.KubeAPIContentType != "" {
			config.ContentType = framework.TestContext.KubeAPIContentType
		}

		// For the purpose of this test, we want to force that clients
		// do not share underlying transport (which is a default behavior
		// in Kubernetes). Thus, we are explicitly creating transport for
		// each client here.
		transportConfig, err := config.TransportConfig()
		if err != nil {
			return nil, nil, nil, err
		}
		tlsConfig, err := transport.TLSConfigFor(transportConfig)
		if err != nil {
			return nil, nil, nil, err
		}
		config.Transport = utilnet.SetTransportDefaults(&http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
			MaxIdleConnsPerHost: 100,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		})
		// Overwrite TLS-related fields from config to avoid collision with
		// Transport field.
		config.TLSClientConfig = restclient.TLSClientConfig{}

		c, err := clientset.NewForConfig(config)
		if err != nil {
			return nil, nil, nil, err
		}
		clients[i] = c
		internalClient, err := internalclientset.NewForConfig(config)
		if err != nil {
			return nil, nil, nil, err
		}
		internalClients[i] = internalClient

		// create scale client, if GroupVersion or NegotiatedSerializer are not set
		// assign default values - these fields are mandatory (required by RESTClientFor).
		if config.GroupVersion == nil {
			config.GroupVersion = &schema.GroupVersion{}
		}
		if config.NegotiatedSerializer == nil {
			config.NegotiatedSerializer = legacyscheme.Codecs
		}
		restClient, err := restclient.RESTClientFor(config)
		if err != nil {
			return nil, nil, nil, err
		}
		discoClient, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			return nil, nil, nil, err
		}
		cachedDiscoClient := cacheddiscovery.NewMemCacheClient(discoClient)
		restMapper := discovery.NewDeferredDiscoveryRESTMapper(cachedDiscoClient)
		restMapper.Reset()
		resolver := scaleclient.NewDiscoveryScaleKindResolver(cachedDiscoClient)
		scalesClients[i] = scaleclient.New(restClient, restMapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	}
	return clients, internalClients, scalesClients, nil
}

func computePodCounts(total int) (int, int, int) {
	// Small RCs owns ~0.5 of total number of pods, medium and big RCs ~0.25 each.
	// For example for 3000 pods (100 nodes, 30 pods per node) there are:
	//  - 300 small RCs each 5 pods
	//  - 25 medium RCs each 30 pods
	//  - 3 big RCs each 250 pods
	bigGroupCount := total / 4 / bigGroupSize
	total -= bigGroupCount * bigGroupSize
	mediumGroupCount := total / 3 / mediumGroupSize
	total -= mediumGroupCount * mediumGroupSize
	smallGroupCount := total / smallGroupSize
	return smallGroupCount, mediumGroupCount, bigGroupCount
}

func generateConfigs(
	totalPods int,
	image string,
	command []string,
	nss []*v1.Namespace,
	kind schema.GroupKind,
	secretsPerPod int,
	configMapsPerPod int,
) ([]testutils.RunObjectConfig, []*testutils.SecretConfig, []*testutils.ConfigMapConfig) {
	configs := make([]testutils.RunObjectConfig, 0)
	secretConfigs := make([]*testutils.SecretConfig, 0)
	configMapConfigs := make([]*testutils.ConfigMapConfig, 0)

	smallGroupCount, mediumGroupCount, bigGroupCount := computePodCounts(totalPods)
	newConfigs, newSecretConfigs, newConfigMapConfigs := GenerateConfigsForGroup(nss, smallGroupName, smallGroupSize, smallGroupCount, image, command, kind, secretsPerPod, configMapsPerPod)
	configs = append(configs, newConfigs...)
	secretConfigs = append(secretConfigs, newSecretConfigs...)
	configMapConfigs = append(configMapConfigs, newConfigMapConfigs...)
	newConfigs, newSecretConfigs, newConfigMapConfigs = GenerateConfigsForGroup(nss, mediumGroupName, mediumGroupSize, mediumGroupCount, image, command, kind, secretsPerPod, configMapsPerPod)
	configs = append(configs, newConfigs...)
	secretConfigs = append(secretConfigs, newSecretConfigs...)
	configMapConfigs = append(configMapConfigs, newConfigMapConfigs...)
	newConfigs, newSecretConfigs, newConfigMapConfigs = GenerateConfigsForGroup(nss, bigGroupName, bigGroupSize, bigGroupCount, image, command, kind, secretsPerPod, configMapsPerPod)
	configs = append(configs, newConfigs...)
	secretConfigs = append(secretConfigs, newSecretConfigs...)
	configMapConfigs = append(configMapConfigs, newConfigMapConfigs...)

	// Create a number of clients to better simulate real usecase
	// where not everyone is using exactly the same client.
	rcsPerClient := 20
	clients, internalClients, scalesClients, err := createClients((len(configs) + rcsPerClient - 1) / rcsPerClient)
	framework.ExpectNoError(err)

	for i := 0; i < len(configs); i++ {
		configs[i].SetClient(clients[i%len(clients)])
		configs[i].SetInternalClient(internalClients[i%len(internalClients)])
		configs[i].SetScalesClient(scalesClients[i%len(clients)])
	}
	for i := 0; i < len(secretConfigs); i++ {
		secretConfigs[i].Client = clients[i%len(clients)]
	}
	for i := 0; i < len(configMapConfigs); i++ {
		configMapConfigs[i].Client = clients[i%len(clients)]
	}

	return configs, secretConfigs, configMapConfigs
}

func GenerateConfigsForGroup(
	nss []*v1.Namespace,
	groupName string,
	size, count int,
	image string,
	command []string,
	kind schema.GroupKind,
	secretsPerPod int,
	configMapsPerPod int,
) ([]testutils.RunObjectConfig, []*testutils.SecretConfig, []*testutils.ConfigMapConfig) {
	configs := make([]testutils.RunObjectConfig, 0, count)
	secretConfigs := make([]*testutils.SecretConfig, 0, count*secretsPerPod)
	configMapConfigs := make([]*testutils.ConfigMapConfig, 0, count*configMapsPerPod)
	savedKind := kind
	for i := 1; i <= count; i++ {
		kind = savedKind
		namespace := nss[i%len(nss)].Name
		secretNames := make([]string, 0, secretsPerPod)
		configMapNames := make([]string, 0, configMapsPerPod)

		for j := 0; j < secretsPerPod; j++ {
			secretName := fmt.Sprintf("%v-%v-secret-%v", groupName, i, j)
			secretConfigs = append(secretConfigs, &testutils.SecretConfig{
				Content:   map[string]string{"foo": "bar"},
				Client:    nil, // this will be overwritten later
				Name:      secretName,
				Namespace: namespace,
				LogFunc:   framework.Logf,
			})
			secretNames = append(secretNames, secretName)
		}

		for j := 0; j < configMapsPerPod; j++ {
			configMapName := fmt.Sprintf("%v-%v-configmap-%v", groupName, i, j)
			configMapConfigs = append(configMapConfigs, &testutils.ConfigMapConfig{
				Content:   map[string]string{"foo": "bar"},
				Client:    nil, // this will be overwritten later
				Name:      configMapName,
				Namespace: namespace,
				LogFunc:   framework.Logf,
			})
			configMapNames = append(configMapNames, configMapName)
		}

		baseConfig := &testutils.RCConfig{
			Client:         nil, // this will be overwritten later
			InternalClient: nil, // this will be overwritten later
			Name:           groupName + "-" + strconv.Itoa(i),
			Namespace:      namespace,
			Timeout:        10 * time.Minute,
			Image:          image,
			Command:        command,
			Replicas:       size,
			CpuRequest:     10,       // 0.01 core
			MemRequest:     26214400, // 25MB
			SecretNames:    secretNames,
			ConfigMapNames: configMapNames,
			// Define a label to group every 2 RCs into one service.
			Labels: map[string]string{svcLabelKey: groupName + "-" + strconv.Itoa((i+1)/2)},
		}

		if kind == randomKind {
			kind = knownKinds[rand.Int()%len(knownKinds)]
		}

		var config testutils.RunObjectConfig
		switch kind {
		case api.Kind("ReplicationController"):
			config = baseConfig
		case extensions.Kind("ReplicaSet"):
			config = &testutils.ReplicaSetConfig{RCConfig: *baseConfig}
		case extensions.Kind("Deployment"):
			config = &testutils.DeploymentConfig{RCConfig: *baseConfig}
		case batch.Kind("Job"):
			config = &testutils.JobConfig{RCConfig: *baseConfig}
		default:
			framework.Failf("Unsupported kind for config creation: %v", kind)
		}
		configs = append(configs, config)
	}
	return configs, secretConfigs, configMapConfigs
}

func generateServicesForConfigs(configs []testutils.RunObjectConfig) []*v1.Service {
	services := make([]*v1.Service, 0)
	currentSvcLabel := ""
	for _, config := range configs {
		svcLabel, found := config.GetLabelValue(svcLabelKey)
		if !found || svcLabel == currentSvcLabel {
			continue
		}
		currentSvcLabel = svcLabel
		serviceName := config.GetName() + "-svc"
		labels := map[string]string{
			"name":      config.GetName(),
			svcLabelKey: currentSvcLabel,
		}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: config.GetNamespace(),
			},
			Spec: v1.ServiceSpec{
				Selector: labels,
				Ports: []v1.ServicePort{{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				}},
			},
		}
		services = append(services, service)
	}
	return services
}

func sleepUpTo(d time.Duration) {
	if d.Nanoseconds() > 0 {
		time.Sleep(time.Duration(rand.Int63n(d.Nanoseconds())))
	}
}

func retryWithExponentialBackOff(initialDuration time.Duration, fn wait.ConditionFunc) error {
	backoff := wait.Backoff{
		Duration: initialDuration,
		Factor:   3,
		Jitter:   0,
		Steps:    6,
	}
	return wait.ExponentialBackoff(backoff, fn)
}

func createAllResources(configs []testutils.RunObjectConfig, creatingTime time.Duration, testPhase *timer.Phase) {
	defer testPhase.End()
	var wg sync.WaitGroup
	wg.Add(len(configs))
	for _, config := range configs {
		go createResource(&wg, config, creatingTime)
	}
	wg.Wait()
}

func createResource(wg *sync.WaitGroup, config testutils.RunObjectConfig, creatingTime time.Duration) {
	defer GinkgoRecover()
	defer wg.Done()

	sleepUpTo(creatingTime)
	framework.ExpectNoError(config.Run(), fmt.Sprintf("creating %v %s", config.GetKind(), config.GetName()))
}

func scaleAllResources(configs []testutils.RunObjectConfig, scalingTime time.Duration, testPhase *timer.Phase) {
	defer testPhase.End()
	var wg sync.WaitGroup
	wg.Add(len(configs))
	for _, config := range configs {
		go scaleResource(&wg, config, scalingTime)
	}
	wg.Wait()
}

// Scales RC to a random size within [0.5*size, 1.5*size] and lists all the pods afterwards.
// Scaling happens always based on original size, not the current size.
func scaleResource(wg *sync.WaitGroup, config testutils.RunObjectConfig, scalingTime time.Duration) {
	defer GinkgoRecover()
	defer wg.Done()

	sleepUpTo(scalingTime)
	newSize := uint(rand.Intn(config.GetReplicas()) + config.GetReplicas()/2)
	framework.ExpectNoError(framework.ScaleResource(
		config.GetClient(),
		config.GetScalesGetter(),
		config.GetNamespace(),
		config.GetName(),
		newSize,
		true,
		config.GetKind(),
		config.GetGroupResource(),
	),
		fmt.Sprintf("scaling %v %v", config.GetKind(), config.GetName()))

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"name": config.GetName()}))
	options := metav1.ListOptions{
		LabelSelector:   selector.String(),
		ResourceVersion: "0",
	}
	listResourcePodsFunc := func() (bool, error) {
		_, err := config.GetClient().CoreV1().Pods(config.GetNamespace()).List(options)
		if err == nil {
			return true, nil
		}
		framework.Logf("Failed to list pods from %v %v due to: %v", config.GetKind(), config.GetName(), err)
		if testutils.IsRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to list pods from %v %v with non-retriable error: %v", config.GetKind(), config.GetName(), err)
	}
	err := retryWithExponentialBackOff(100*time.Millisecond, listResourcePodsFunc)
	framework.ExpectNoError(err)
}

func deleteAllResources(configs []testutils.RunObjectConfig, deletingTime time.Duration, testPhase *timer.Phase) {
	defer testPhase.End()
	var wg sync.WaitGroup
	wg.Add(len(configs))
	for _, config := range configs {
		go deleteResource(&wg, config, deletingTime)
	}
	wg.Wait()
}

func deleteResource(wg *sync.WaitGroup, config testutils.RunObjectConfig, deletingTime time.Duration) {
	defer GinkgoRecover()
	defer wg.Done()

	sleepUpTo(deletingTime)
	if framework.TestContext.GarbageCollectorEnabled && config.GetKind() != extensions.Kind("Deployment") {
		framework.ExpectNoError(framework.DeleteResourceAndWaitForGC(
			config.GetClient(), config.GetKind(), config.GetNamespace(), config.GetName()),
			fmt.Sprintf("deleting %v %s", config.GetKind(), config.GetName()))
	} else {
		framework.ExpectNoError(framework.DeleteResourceAndPods(
			config.GetClient(), config.GetInternalClient(), config.GetScalesGetter(), config.GetKind(), config.GetNamespace(), config.GetName()),
			fmt.Sprintf("deleting %v %s", config.GetKind(), config.GetName()))
	}
}

func CreateNamespaces(f *framework.Framework, namespaceCount int, namePrefix string, testPhase *timer.Phase) ([]*v1.Namespace, error) {
	defer testPhase.End()
	namespaces := []*v1.Namespace{}
	for i := 1; i <= namespaceCount; i++ {
		namespace, err := f.CreateNamespace(fmt.Sprintf("%v-%d", namePrefix, i), nil)
		if err != nil {
			return []*v1.Namespace{}, err
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func CreateQuotas(f *framework.Framework, namespaces []*v1.Namespace, podCount int, testPhase *timer.Phase) error {
	defer testPhase.End()
	quotaTemplate := &v1.ResourceQuota{
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{"pods": *resource.NewQuantity(int64(podCount), resource.DecimalSI)},
		},
	}
	for _, ns := range namespaces {
		quotaTemplate.Name = ns.Name + "-quota"
		if err := testutils.CreateResourceQuotaWithRetries(f.ClientSet, ns.Name, quotaTemplate); err != nil {
			return fmt.Errorf("Error creating quota: %v", err)
		}
	}
	return nil
}
