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

package monitoring

import (
	"context"
	"time"

	"golang.org/x/oauth2/google"
	clientset "k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instrumentation "k8s.io/kubernetes/test/e2e/instrumentation/common"

	gcm "google.golang.org/api/monitoring/v3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/discovery"
	"k8s.io/kubernetes/test/e2e/framework"
	customclient "k8s.io/metrics/pkg/client/custom_metrics"
	externalclient "k8s.io/metrics/pkg/client/external_metrics"
)

const (
	stackdriverExporterPod1  = "stackdriver-exporter-1"
	stackdriverExporterPod2  = "stackdriver-exporter-2"
	stackdriverExporterLabel = "stackdriver-exporter"
)

var _ = instrumentation.SIGDescribe("Stackdriver Monitoring", func() {
	BeforeEach(func() {
		framework.SkipUnlessProviderIs("gce", "gke")
	})

	f := framework.NewDefaultFramework("stackdriver-monitoring")

	It("should run Custom Metrics - Stackdriver Adapter for old resource model [Feature:StackdriverCustomMetrics]", func() {
		kubeClient := f.ClientSet
		config, err := framework.LoadConfig()
		if err != nil {
			framework.Failf("Failed to load config: %s", err)
		}
		customMetricsClient := customclient.NewForConfigOrDie(config)
		discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)
		testCustomMetrics(f, kubeClient, customMetricsClient, discoveryClient, AdapterForOldResourceModel)
	})

	It("should run Custom Metrics - Stackdriver Adapter for new resource model [Feature:StackdriverCustomMetrics]", func() {
		kubeClient := f.ClientSet
		config, err := framework.LoadConfig()
		if err != nil {
			framework.Failf("Failed to load config: %s", err)
		}
		customMetricsClient := customclient.NewForConfigOrDie(config)
		discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)
		testCustomMetrics(f, kubeClient, customMetricsClient, discoveryClient, AdapterForNewResourceModel)
	})

	It("should run Custom Metrics - Stackdriver Adapter for external metrics [Feature:StackdriverExternalMetrics]", func() {
		kubeClient := f.ClientSet
		config, err := framework.LoadConfig()
		if err != nil {
			framework.Failf("Failed to load config: %s", err)
		}
		externalMetricsClient := externalclient.NewForConfigOrDie(config)
		testExternalMetrics(f, kubeClient, externalMetricsClient)
	})
})

func testCustomMetrics(f *framework.Framework, kubeClient clientset.Interface, customMetricsClient customclient.CustomMetricsClient, discoveryClient *discovery.DiscoveryClient, adapterDeployment string) {
	projectId := framework.TestContext.CloudConfig.ProjectID

	ctx := context.Background()
	client, err := google.DefaultClient(ctx, gcm.CloudPlatformScope)

	gcmService, err := gcm.New(client)
	if err != nil {
		framework.Failf("Failed to create gcm service, %v", err)
	}

	// Set up a cluster: create a custom metric and set up k8s-sd adapter
	err = CreateDescriptors(gcmService, projectId)
	if err != nil {
		framework.Failf("Failed to create metric descriptor: %s", err)
	}
	defer CleanupDescriptors(gcmService, projectId)

	err = CreateAdapter(adapterDeployment)
	if err != nil {
		framework.Failf("Failed to set up: %s", err)
	}
	defer CleanupAdapter(adapterDeployment)

	_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(HPAPermissions)
	defer kubeClient.RbacV1().ClusterRoleBindings().Delete("custom-metrics-reader", &metav1.DeleteOptions{})

	// Run application that exports the metric
	_, err = createSDExporterPods(f, kubeClient)
	if err != nil {
		framework.Failf("Failed to create stackdriver-exporter pod: %s", err)
	}
	defer cleanupSDExporterPod(f, kubeClient)

	// Wait a short amount of time to create a pod and export some metrics
	// TODO: add some events to wait for instead of fixed amount of time
	//       i.e. pod creation, first time series exported
	time.Sleep(60 * time.Second)

	verifyResponsesFromCustomMetricsAPI(f, customMetricsClient, discoveryClient)
}

// TODO(kawych): migrate this test to new resource model
func testExternalMetrics(f *framework.Framework, kubeClient clientset.Interface, externalMetricsClient externalclient.ExternalMetricsClient) {
	projectId := framework.TestContext.CloudConfig.ProjectID

	ctx := context.Background()
	client, err := google.DefaultClient(ctx, gcm.CloudPlatformScope)

	gcmService, err := gcm.New(client)
	if err != nil {
		framework.Failf("Failed to create gcm service, %v", err)
	}

	// Set up a cluster: create a custom metric and set up k8s-sd adapter
	err = CreateDescriptors(gcmService, projectId)
	if err != nil {
		framework.Failf("Failed to create metric descriptor: %s", err)
	}
	defer CleanupDescriptors(gcmService, projectId)

	// Both deployments - for old and new resource model - expose External Metrics API.
	err = CreateAdapter(AdapterForOldResourceModel)
	if err != nil {
		framework.Failf("Failed to set up: %s", err)
	}
	defer CleanupAdapter(AdapterForOldResourceModel)

	_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(HPAPermissions)
	defer kubeClient.RbacV1().ClusterRoleBindings().Delete("custom-metrics-reader", &metav1.DeleteOptions{})

	// Run application that exports the metric
	pod, err := createSDExporterPods(f, kubeClient)
	if err != nil {
		framework.Failf("Failed to create stackdriver-exporter pod: %s", err)
	}
	defer cleanupSDExporterPod(f, kubeClient)

	// Wait a short amount of time to create a pod and export some metrics
	// TODO: add some events to wait for instead of fixed amount of time
	//       i.e. pod creation, first time series exported
	time.Sleep(60 * time.Second)

	verifyResponseFromExternalMetricsAPI(f, externalMetricsClient, pod)
}

func verifyResponsesFromCustomMetricsAPI(f *framework.Framework, customMetricsClient customclient.CustomMetricsClient, discoveryClient *discovery.DiscoveryClient) {
	resources, err := discoveryClient.ServerResourcesForGroupVersion("custom.metrics.k8s.io/v1beta1")
	if err != nil {
		framework.Failf("Failed to retrieve a list of supported metrics: %s", err)
	}
	gotCustomMetric, gotUnusedMetric := false, false
	for _, resource := range resources.APIResources {
		if resource.Name == "*/"+CustomMetricName {
			gotCustomMetric = true
		} else if resource.Name == "*/"+UnusedMetricName {
			gotUnusedMetric = true
		} else {
			framework.Failf("Unexpected metric %s. Only metric %s should be supported", resource.Name, CustomMetricName)
		}
	}
	if !gotCustomMetric {
		framework.Failf("Metric '%s' expected but not received", CustomMetricName)
	}
	if !gotUnusedMetric {
		framework.Failf("Metric '%s' expected but not received", UnusedMetricName)
	}
	value, err := customMetricsClient.NamespacedMetrics(f.Namespace.Name).GetForObject(schema.GroupKind{Group: "", Kind: "Pod"}, stackdriverExporterPod1, CustomMetricName)
	if err != nil {
		framework.Failf("Failed query: %s", err)
	}
	if value.Value.Value() != CustomMetricValue {
		framework.Failf("Unexpected metric value for metric %s: expected %v but received %v", CustomMetricName, CustomMetricValue, value.Value)
	}
	filter, err := labels.NewRequirement("name", selection.Equals, []string{stackdriverExporterLabel})
	if err != nil {
		framework.Failf("Couldn't create a label filter")
	}
	values, err := customMetricsClient.NamespacedMetrics(f.Namespace.Name).GetForObjects(schema.GroupKind{Group: "", Kind: "Pod"}, labels.NewSelector().Add(*filter), CustomMetricName)
	if err != nil {
		framework.Failf("Failed query: %s", err)
	}
	if len(values.Items) != 1 {
		framework.Failf("Expected results for exactly 1 pod, but %v results received", len(values.Items))
	}
	if values.Items[0].DescribedObject.Name != stackdriverExporterPod1 || values.Items[0].Value.Value() != CustomMetricValue {
		framework.Failf("Unexpected metric value for metric %s and pod %s: %v", CustomMetricName, values.Items[0].DescribedObject.Name, values.Items[0].Value.Value())
	}
}

func verifyResponseFromExternalMetricsAPI(f *framework.Framework, externalMetricsClient externalclient.ExternalMetricsClient, pod *v1.Pod) {
	req1, _ := labels.NewRequirement("resource.type", selection.Equals, []string{"gke_container"})
	// It's important to filter out only metrics from the right namespace, since multiple e2e tests
	// may run in the same project concurrently. "dummy" is added to test
	req2, _ := labels.NewRequirement("resource.labels.pod_id", selection.In, []string{string(pod.UID), "dummy"})
	req3, _ := labels.NewRequirement("resource.labels.namespace_id", selection.Exists, []string{})
	req4, _ := labels.NewRequirement("resource.labels.zone", selection.NotEquals, []string{"dummy"})
	req5, _ := labels.NewRequirement("resource.labels.cluster_name", selection.NotIn, []string{"foo", "bar"})
	values, err := externalMetricsClient.
		NamespacedMetrics("dummy").
		List("custom.googleapis.com|"+CustomMetricName, labels.NewSelector().Add(*req1, *req2, *req3, *req4, *req5))
	if err != nil {
		framework.Failf("Failed query: %s", err)
	}
	if len(values.Items) != 1 {
		framework.Failf("Expected exactly one external metric value, but % values received", len(values.Items))
	}
	if values.Items[0].MetricName != "custom.googleapis.com|"+CustomMetricName ||
		values.Items[0].Value.Value() != CustomMetricValue ||
		// Check one label just to make sure labels are included
		values.Items[0].MetricLabels["resource.labels.pod_id"] != string(pod.UID) {
		framework.Failf("Unexpected result for metric %s: %v", CustomMetricName, values.Items[0])
	}
}

func cleanupSDExporterPod(f *framework.Framework, cs clientset.Interface) {
	err := cs.CoreV1().Pods(f.Namespace.Name).Delete(stackdriverExporterPod1, &metav1.DeleteOptions{})
	if err != nil {
		framework.Logf("Failed to delete %s pod: %v", stackdriverExporterPod1, err)
	}
	err = cs.CoreV1().Pods(f.Namespace.Name).Delete(stackdriverExporterPod2, &metav1.DeleteOptions{})
	if err != nil {
		framework.Logf("Failed to delete %s pod: %v", stackdriverExporterPod2, err)
	}
}

func createSDExporterPods(f *framework.Framework, cs clientset.Interface) (*v1.Pod, error) {
	pod, err := cs.CoreV1().Pods(f.Namespace.Name).Create(StackdriverExporterPod(stackdriverExporterPod1, f.Namespace.Name, stackdriverExporterLabel, CustomMetricName, CustomMetricValue))
	if err != nil {
		return nil, err
	}
	_, err = cs.CoreV1().Pods(f.Namespace.Name).Create(StackdriverExporterPod(stackdriverExporterPod2, f.Namespace.Name, stackdriverExporterLabel, UnusedMetricName, UnusedMetricValue))
	return pod, err
}
