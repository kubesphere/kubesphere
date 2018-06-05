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

package storage

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/model"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	kubeletmetrics "k8s.io/kubernetes/pkg/kubelet/metrics"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/metrics"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

// This test needs to run in serial because other tests could interfere
// with metrics being tested here.
var _ = utils.SIGDescribe("[Serial] Volume metrics", func() {
	var (
		c              clientset.Interface
		ns             string
		pvc            *v1.PersistentVolumeClaim
		metricsGrabber *metrics.MetricsGrabber
	)
	f := framework.NewDefaultFramework("pv")

	BeforeEach(func() {
		c = f.ClientSet
		ns = f.Namespace.Name
		framework.SkipUnlessProviderIs("gce", "gke", "aws")
		defaultScName := getDefaultStorageClassName(c)
		verifyDefaultStorageClass(c, defaultScName, true)

		test := storageClassTest{
			name:      "default",
			claimSize: "2Gi",
		}

		pvc = newClaim(test, ns, "default")
		var err error
		metricsGrabber, err = metrics.NewMetricsGrabber(c, nil, true, false, true, false, false)

		if err != nil {
			framework.Failf("Error creating metrics grabber : %v", err)
		}
	})

	AfterEach(func() {
		framework.DeletePersistentVolumeClaim(c, pvc.Name, pvc.Namespace)
	})

	It("should create prometheus metrics for volume provisioning and attach/detach", func() {
		var err error

		if !metricsGrabber.HasRegisteredMaster() {
			framework.Skipf("Environment does not support getting controller-manager metrics - skipping")
		}

		controllerMetrics, err := metricsGrabber.GrabFromControllerManager()

		Expect(err).NotTo(HaveOccurred(), "Error getting c-m metrics : %v", err)

		storageOpMetrics := getControllerStorageMetrics(controllerMetrics)

		pvc, err = c.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(pvc)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvc).ToNot(Equal(nil))

		claims := []*v1.PersistentVolumeClaim{pvc}

		pod := framework.MakePod(ns, nil, claims, false, "")
		pod, err = c.CoreV1().Pods(ns).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		err = framework.WaitForPodRunningInNamespace(c, pod)
		framework.ExpectNoError(framework.WaitForPodRunningInNamespace(c, pod), "Error starting pod ", pod.Name)

		framework.Logf("Deleting pod %q/%q", pod.Namespace, pod.Name)
		framework.ExpectNoError(framework.DeletePodWithWait(f, c, pod))

		updatedStorageMetrics := waitForDetachAndGrabMetrics(storageOpMetrics, metricsGrabber)

		Expect(len(updatedStorageMetrics)).ToNot(Equal(0), "Error fetching c-m updated storage metrics")

		volumeOperations := []string{"volume_provision", "volume_detach", "volume_attach"}

		for _, volumeOp := range volumeOperations {
			verifyMetricCount(storageOpMetrics, updatedStorageMetrics, volumeOp)
		}
	})

	It("should create volume metrics with the correct PVC ref", func() {
		var err error
		pvc, err = c.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(pvc)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvc).ToNot(Equal(nil))

		claims := []*v1.PersistentVolumeClaim{pvc}
		pod := framework.MakePod(ns, nil, claims, false, "")
		pod, err = c.CoreV1().Pods(ns).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		err = framework.WaitForPodRunningInNamespace(c, pod)
		framework.ExpectNoError(framework.WaitForPodRunningInNamespace(c, pod), "Error starting pod ", pod.Name)

		pod, err = c.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// Verify volume stat metrics were collected for the referenced PVC
		volumeStatKeys := []string{
			kubeletmetrics.VolumeStatsUsedBytesKey,
			kubeletmetrics.VolumeStatsCapacityBytesKey,
			kubeletmetrics.VolumeStatsAvailableBytesKey,
			kubeletmetrics.VolumeStatsUsedBytesKey,
			kubeletmetrics.VolumeStatsInodesFreeKey,
			kubeletmetrics.VolumeStatsInodesUsedKey,
		}
		// Poll kubelet metrics waiting for the volume to be picked up
		// by the volume stats collector
		var kubeMetrics metrics.KubeletMetrics
		waitErr := wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
			framework.Logf("Grabbing Kubelet metrics")
			// Grab kubelet metrics from the node the pod was scheduled on
			var err error
			kubeMetrics, err = metricsGrabber.GrabFromKubelet(pod.Spec.NodeName)
			if err != nil {
				framework.Logf("Error fetching kubelet metrics")
				return false, err
			}
			key := volumeStatKeys[0]
			kubeletKeyName := fmt.Sprintf("%s_%s", kubeletmetrics.KubeletSubsystem, key)
			if !findVolumeStatMetric(kubeletKeyName, pvc.Namespace, pvc.Name, kubeMetrics) {
				return false, nil
			}
			return true, nil
		})
		Expect(waitErr).NotTo(HaveOccurred(), "Error finding volume metrics : %v", waitErr)

		for _, key := range volumeStatKeys {
			kubeletKeyName := fmt.Sprintf("%s_%s", kubeletmetrics.KubeletSubsystem, key)
			found := findVolumeStatMetric(kubeletKeyName, pvc.Namespace, pvc.Name, kubeMetrics)
			Expect(found).To(BeTrue(), "PVC %s, Namespace %s not found for %s", pvc.Name, pvc.Namespace, kubeletKeyName)
		}

		framework.Logf("Deleting pod %q/%q", pod.Namespace, pod.Name)
		framework.ExpectNoError(framework.DeletePodWithWait(f, c, pod))
	})

	// Test for pv controller metrics, concretely: bound/unbound pv/pvc count.
	Describe("PVController", func() {
		const (
			classKey     = "storage_class"
			namespaceKey = "namespace"

			boundPVKey    = "pv_collector_bound_pv_count"
			unboundPVKey  = "pv_collector_unbound_pv_count"
			boundPVCKey   = "pv_collector_bound_pvc_count"
			unboundPVCKey = "pv_collector_unbound_pvc_count"
		)

		var (
			pv  *v1.PersistentVolume
			pvc *v1.PersistentVolumeClaim

			className = "bound-unbound-count-test-sc"
			pvConfig  = framework.PersistentVolumeConfig{
				PVSource: v1.PersistentVolumeSource{
					HostPath: &v1.HostPathVolumeSource{Path: "/data"},
				},
				NamePrefix:       "pv-test-",
				StorageClassName: className,
			}
			pvcConfig = framework.PersistentVolumeClaimConfig{StorageClassName: &className}

			metrics = []struct {
				name      string
				dimension string
			}{
				{boundPVKey, classKey},
				{unboundPVKey, classKey},
				{boundPVCKey, namespaceKey},
				{unboundPVCKey, namespaceKey},
			}

			// Original metric values before we create any PV/PVCs. The length should be 4,
			// and the elements should be bound pv count, unbound pv count, bound pvc count,
			// unbound pvc count in turn.
			// We use these values to calculate relative increment of each test.
			originMetricValues []map[string]int64
		)

		// validator used to validate each metric's values, the length of metricValues
		// should be 4, and the elements should be bound pv count, unbound pv count, bound
		// pvc count, unbound pvc count in turn.
		validator := func(metricValues []map[string]int64) {
			Expect(len(metricValues)).To(Equal(4),
				"Wrong metric size: %d", len(metricValues))

			controllerMetrics, err := metricsGrabber.GrabFromControllerManager()
			Expect(err).NotTo(HaveOccurred(), "Error getting c-m metricValues: %v", err)

			for i, metric := range metrics {
				expectValues := metricValues[i]
				if expectValues == nil {
					expectValues = make(map[string]int64)
				}
				// We using relative increment value instead of absolute value to reduce unexpected flakes.
				// Concretely, we expect the difference of the updated values and original values for each
				// test suit are equal to expectValues.
				actualValues := calculateRelativeValues(originMetricValues[i],
					getPVControllerMetrics(controllerMetrics, metric.name, metric.dimension))
				Expect(actualValues).To(Equal(expectValues),
					"Wrong pv controller metric %s(%s): wanted %v, got %v",
					metric.name, metric.dimension, expectValues, actualValues)
			}
		}

		BeforeEach(func() {
			if !metricsGrabber.HasRegisteredMaster() {
				framework.Skipf("Environment does not support getting controller-manager metrics - skipping")
			}

			pv = framework.MakePersistentVolume(pvConfig)
			pvc = framework.MakePersistentVolumeClaim(pvcConfig, ns)

			// Initializes all original metric values.
			controllerMetrics, err := metricsGrabber.GrabFromControllerManager()
			Expect(err).NotTo(HaveOccurred(), "Error getting c-m metricValues: %v", err)
			for _, metric := range metrics {
				originMetricValues = append(originMetricValues,
					getPVControllerMetrics(controllerMetrics, metric.name, metric.dimension))
			}
		})

		AfterEach(func() {
			if err := framework.DeletePersistentVolume(c, pv.Name); err != nil {
				framework.Failf("Error deleting pv: %v", err)
			}
			if err := framework.DeletePersistentVolumeClaim(c, pvc.Name, pvc.Namespace); err != nil {
				framework.Failf("Error deleting pvc: %v", err)
			}

			// Clear original metric values.
			originMetricValues = nil
		})

		It("should create none metrics for pvc controller before creating any PV or PVC", func() {
			validator([]map[string]int64{nil, nil, nil, nil})
		})

		It("should create unbound pv count metrics for pvc controller after creating pv only",
			func() {
				var err error
				pv, err = framework.CreatePV(c, pv)
				Expect(err).NotTo(HaveOccurred(), "Error creating pv: %v", err)
				waitForPVControllerSync(metricsGrabber, unboundPVKey, classKey)
				validator([]map[string]int64{nil, {className: 1}, nil, nil})
			})

		It("should create unbound pvc count metrics for pvc controller after creating pvc only",
			func() {
				var err error
				pvc, err = framework.CreatePVC(c, ns, pvc)
				Expect(err).NotTo(HaveOccurred(), "Error creating pvc: %v", err)
				waitForPVControllerSync(metricsGrabber, unboundPVCKey, namespaceKey)
				validator([]map[string]int64{nil, nil, nil, {ns: 1}})
			})

		It("should create bound pv/pvc count metrics for pvc controller after creating both pv and pvc",
			func() {
				var err error
				pv, pvc, err = framework.CreatePVPVC(c, pvConfig, pvcConfig, ns, true)
				Expect(err).NotTo(HaveOccurred(), "Error creating pv pvc: %v", err)
				waitForPVControllerSync(metricsGrabber, boundPVKey, classKey)
				waitForPVControllerSync(metricsGrabber, boundPVCKey, namespaceKey)
				validator([]map[string]int64{{className: 1}, nil, {ns: 1}, nil})

			})
	})
})

func waitForDetachAndGrabMetrics(oldMetrics map[string]int64, metricsGrabber *metrics.MetricsGrabber) map[string]int64 {
	backoff := wait.Backoff{
		Duration: 10 * time.Second,
		Factor:   1.2,
		Steps:    21,
	}

	updatedStorageMetrics := make(map[string]int64)
	oldDetachCount, ok := oldMetrics["volume_detach"]
	if !ok {
		oldDetachCount = 0
	}

	verifyMetricFunc := func() (bool, error) {
		updatedMetrics, err := metricsGrabber.GrabFromControllerManager()

		if err != nil {
			framework.Logf("Error fetching controller-manager metrics")
			return false, err
		}

		updatedStorageMetrics = getControllerStorageMetrics(updatedMetrics)
		newDetachCount, ok := updatedStorageMetrics["volume_detach"]

		// if detach metrics are not yet there, we need to retry
		if !ok {
			return false, nil
		}

		// if old Detach count is more or equal to new detach count, that means detach
		// event has not been observed yet.
		if oldDetachCount >= newDetachCount {
			return false, nil
		}
		return true, nil
	}

	waitErr := wait.ExponentialBackoff(backoff, verifyMetricFunc)
	Expect(waitErr).NotTo(HaveOccurred(), "Timeout error fetching storage c-m metrics : %v", waitErr)
	return updatedStorageMetrics
}

func verifyMetricCount(oldMetrics map[string]int64, newMetrics map[string]int64, metricName string) {
	oldCount, ok := oldMetrics[metricName]
	// if metric does not exist in oldMap, it probably hasn't been emitted yet.
	if !ok {
		oldCount = 0
	}

	newCount, ok := newMetrics[metricName]
	Expect(ok).To(BeTrue(), "Error getting updated metrics for %s", metricName)
	// It appears that in a busy cluster some spurious detaches are unavoidable
	// even if the test is run serially.  We really just verify if new count
	// is greater than old count
	Expect(newCount).To(BeNumerically(">", oldCount), "New count %d should be more than old count %d for action %s", newCount, oldCount, metricName)
}

func getControllerStorageMetrics(ms metrics.ControllerManagerMetrics) map[string]int64 {
	result := make(map[string]int64)

	for method, samples := range ms {
		if method != "storage_operation_duration_seconds_count" {
			continue
		}

		for _, sample := range samples {
			count := int64(sample.Value)
			operation := string(sample.Metric["operation_name"])
			result[operation] = count
		}
	}
	return result
}

// Finds the sample in the specified metric from `KubeletMetrics` tagged with
// the specified namespace and pvc name
func findVolumeStatMetric(metricKeyName string, namespace string, pvcName string, kubeletMetrics metrics.KubeletMetrics) bool {
	found := false
	errCount := 0
	framework.Logf("Looking for sample in metric `%s` tagged with namespace `%s`, PVC `%s`", metricKeyName, namespace, pvcName)
	if samples, ok := kubeletMetrics[metricKeyName]; ok {
		for _, sample := range samples {
			framework.Logf("Found sample %s", sample.String())
			samplePVC, ok := sample.Metric["persistentvolumeclaim"]
			if !ok {
				framework.Logf("Error getting pvc for metric %s, sample %s", metricKeyName, sample.String())
				errCount++
			}
			sampleNS, ok := sample.Metric["namespace"]
			if !ok {
				framework.Logf("Error getting namespace for metric %s, sample %s", metricKeyName, sample.String())
				errCount++
			}

			if string(samplePVC) == pvcName && string(sampleNS) == namespace {
				found = true
				break
			}
		}
	}
	Expect(errCount).To(Equal(0), "Found invalid samples")
	return found
}

// Wait for the count of a pv controller's metric specified by metricName and dimension bigger than zero.
func waitForPVControllerSync(metricsGrabber *metrics.MetricsGrabber, metricName, dimension string) {
	backoff := wait.Backoff{
		Duration: 10 * time.Second,
		Factor:   1.2,
		Steps:    21,
	}
	verifyMetricFunc := func() (bool, error) {
		updatedMetrics, err := metricsGrabber.GrabFromControllerManager()
		if err != nil {
			framework.Logf("Error fetching controller-manager metrics")
			return false, err
		}
		return len(getPVControllerMetrics(updatedMetrics, metricName, dimension)) > 0, nil
	}
	waitErr := wait.ExponentialBackoff(backoff, verifyMetricFunc)
	Expect(waitErr).NotTo(HaveOccurred(),
		"Timeout error fetching pv controller metrics : %v", waitErr)
}

func getPVControllerMetrics(ms metrics.ControllerManagerMetrics, metricName, dimension string) map[string]int64 {
	result := make(map[string]int64)
	for method, samples := range ms {
		if method != metricName {
			continue
		}
		for _, sample := range samples {
			count := int64(sample.Value)
			dimensionName := string(sample.Metric[model.LabelName(dimension)])
			result[dimensionName] = count
		}
	}
	return result
}

func calculateRelativeValues(originValues, updatedValues map[string]int64) map[string]int64 {
	relativeValues := make(map[string]int64)
	for key, value := range updatedValues {
		relativeValue := value - originValues[key]
		if relativeValue != 0 {
			relativeValues[key] = relativeValue
		}
	}
	for key, value := range originValues {
		if _, exist := updatedValues[key]; !exist && value > 0 {
			relativeValues[key] = -value
		}
	}
	return relativeValues
}
