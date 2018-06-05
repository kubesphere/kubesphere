/*
Copyright 2016 The Kubernetes Authors.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

const (
	pvDeletionTimeout       = 3 * time.Minute
	statefulSetReadyTimeout = 3 * time.Minute
)

var _ = utils.SIGDescribe("Regional PD [Feature:RegionalPD]", func() {
	f := framework.NewDefaultFramework("regional-pd")

	// filled in BeforeEach
	var c clientset.Interface
	var ns string

	BeforeEach(func() {
		c = f.ClientSet
		ns = f.Namespace.Name

		framework.SkipUnlessProviderIs("gce", "gke")
		framework.SkipUnlessMultizone(c)
	})

	Describe("RegionalPD", func() {
		It("should provision storage [Slow]", func() {
			testVolumeProvisioning(c, ns)
		})

		It("should failover to a different zone when all nodes in one zone become unreachable [Slow] [Disruptive]", func() {
			testZonalFailover(c, ns)
		})

	})
})

func testVolumeProvisioning(c clientset.Interface, ns string) {
	cloudZones := getTwoRandomZones(c)

	// This test checks that dynamic provisioning can provision a volume
	// that can be used to persist data among pods.
	tests := []storageClassTest{
		{
			name:           "HDD Regional PD on GCE/GKE",
			cloudProviders: []string{"gce", "gke"},
			provisioner:    "kubernetes.io/gce-pd",
			parameters: map[string]string{
				"type":             "pd-standard",
				"zones":            strings.Join(cloudZones, ","),
				"replication-type": "regional-pd",
			},
			claimSize:    "1.5G",
			expectedSize: "2G",
			pvCheck: func(volume *v1.PersistentVolume) error {
				err := checkGCEPD(volume, "pd-standard")
				if err != nil {
					return err
				}
				return verifyZonesInPV(volume, sets.NewString(cloudZones...), true /* match */)
			},
		},
		{
			name:           "HDD Regional PD with auto zone selection on GCE/GKE",
			cloudProviders: []string{"gce", "gke"},
			provisioner:    "kubernetes.io/gce-pd",
			parameters: map[string]string{
				"type":             "pd-standard",
				"replication-type": "regional-pd",
			},
			claimSize:    "1.5G",
			expectedSize: "2G",
			pvCheck: func(volume *v1.PersistentVolume) error {
				err := checkGCEPD(volume, "pd-standard")
				if err != nil {
					return err
				}
				zones, err := framework.GetClusterZones(c)
				if err != nil {
					return err
				}
				return verifyZonesInPV(volume, zones, false /* match */)
			},
		},
	}

	for _, test := range tests {
		class := newStorageClass(test, ns, "" /* suffix */)
		claim := newClaim(test, ns, "" /* suffix */)
		claim.Spec.StorageClassName = &class.Name
		testDynamicProvisioning(test, c, claim, class)
	}
}

func testZonalFailover(c clientset.Interface, ns string) {
	nodes := framework.GetReadySchedulableNodesOrDie(c)
	nodeCount := len(nodes.Items)

	cloudZones := getTwoRandomZones(c)
	class := newRegionalStorageClass(ns, cloudZones)
	claimTemplate := newClaimTemplate(ns)
	claimTemplate.Spec.StorageClassName = &class.Name
	statefulSet, service, regionalPDLabels := newStatefulSet(claimTemplate, ns)

	By("creating a StorageClass " + class.Name)
	_, err := c.StorageV1().StorageClasses().Create(class)
	Expect(err).NotTo(HaveOccurred())
	defer func() {
		framework.Logf("deleting storage class %s", class.Name)
		framework.ExpectNoError(c.StorageV1().StorageClasses().Delete(class.Name, nil),
			"Error deleting StorageClass %s", class.Name)
	}()

	By("creating a StatefulSet")
	_, err = c.CoreV1().Services(ns).Create(service)
	Expect(err).NotTo(HaveOccurred())
	_, err = c.AppsV1().StatefulSets(ns).Create(statefulSet)
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		framework.Logf("deleting statefulset%q/%q", statefulSet.Namespace, statefulSet.Name)
		// typically this claim has already been deleted
		framework.ExpectNoError(c.AppsV1().StatefulSets(ns).Delete(statefulSet.Name, nil /* options */),
			"Error deleting StatefulSet %s", statefulSet.Name)

		framework.Logf("deleting claims in namespace %s", ns)
		pvc := getPVC(c, ns, regionalPDLabels)
		framework.ExpectNoError(c.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(pvc.Name, nil),
			"Error deleting claim %s.", pvc.Name)
		if pvc.Spec.VolumeName != "" {
			err = framework.WaitForPersistentVolumeDeleted(c, pvc.Spec.VolumeName, framework.Poll, pvDeletionTimeout)
			if err != nil {
				framework.Logf("WARNING: PV %s is not yet deleted, and subsequent tests may be affected.", pvc.Spec.VolumeName)
			}
		}
	}()

	err = framework.WaitForStatefulSetReplicasReady(statefulSet.Name, ns, c, framework.Poll, statefulSetReadyTimeout)
	if err != nil {
		pod := getPod(c, ns, regionalPDLabels)
		Expect(podutil.IsPodReadyConditionTrue(pod.Status)).To(BeTrue(),
			"The statefulset pod has the following conditions: %s", pod.Status.Conditions)
		Expect(err).NotTo(HaveOccurred())
	}

	pvc := getPVC(c, ns, regionalPDLabels)

	By("getting zone information from pod")
	pod := getPod(c, ns, regionalPDLabels)
	nodeName := pod.Spec.NodeName
	node, err := c.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	podZone := node.Labels[apis.LabelZoneFailureDomain]

	// TODO (verult) Consider using node taints to simulate zonal failure instead.
	By("deleting instance group belonging to pod's zone")

	// Asynchronously detect a pod reschedule is triggered during/after instance group deletion.
	waitStatus := make(chan error)
	go func() {
		waitStatus <- waitForStatefulSetReplicasNotReady(statefulSet.Name, ns, c)
	}()

	cloud, err := framework.GetGCECloud()
	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	instanceGroupName := framework.TestContext.CloudConfig.NodeInstanceGroup
	instanceGroup, err := cloud.GetInstanceGroup(instanceGroupName, podZone)
	Expect(err).NotTo(HaveOccurred(),
		"Error getting instance group %s in zone %s", instanceGroupName, podZone)
	err = framework.DeleteManagedInstanceGroup(podZone)
	Expect(err).NotTo(HaveOccurred(),
		"Error deleting instance group in zone %s", podZone)

	defer func() {
		framework.Logf("recreating instance group %s", instanceGroup.Name)

		// HACK improve this when Managed Instance Groups are available through the cloud provider API
		templateName := strings.Replace(instanceGroupName, "group", "template", 1 /* n */)

		framework.ExpectNoError(framework.CreateManagedInstanceGroup(instanceGroup.Size, podZone, templateName),
			"Error recreating instance group %s in zone %s", instanceGroup.Name, podZone)
		framework.ExpectNoError(framework.WaitForReadyNodes(c, nodeCount, framework.RestartNodeReadyAgainTimeout),
			"Error waiting for nodes from the new instance group to become ready.")
	}()

	err = <-waitStatus
	Expect(err).ToNot(HaveOccurred(), "Error waiting for replica to be deleted during failover: %v", err)

	err = framework.WaitForStatefulSetReplicasReady(statefulSet.Name, ns, c, 3*time.Second, framework.RestartPodReadyAgainTimeout)
	if err != nil {
		pod := getPod(c, ns, regionalPDLabels)
		Expect(podutil.IsPodReadyConditionTrue(pod.Status)).To(BeTrue(),
			"The statefulset pod has the following conditions: %s", pod.Status.Conditions)
		Expect(err).NotTo(HaveOccurred())
	}

	By("verifying the same PVC is used by the new pod")
	Expect(getPVC(c, ns, regionalPDLabels).Name).To(Equal(pvc.Name),
		"The same PVC should be used after failover.")

	By("verifying the container output has 2 lines, indicating the pod has been created twice using the same regional PD.")
	pod = getPod(c, ns, regionalPDLabels)
	logs, err := framework.GetPodLogs(c, ns, pod.Name, "")
	Expect(err).NotTo(HaveOccurred(),
		"Error getting logs from pod %s in namespace %s", pod.Name, ns)
	lineCount := len(strings.Split(strings.TrimSpace(logs), "\n"))
	expectedLineCount := 2
	Expect(lineCount).To(Equal(expectedLineCount),
		"Line count of the written file should be %d.", expectedLineCount)

	// Verify the pod is scheduled in the other zone.
	By("verifying the pod is scheduled in a different zone.")
	var otherZone string
	if cloudZones[0] == podZone {
		otherZone = cloudZones[1]
	} else {
		otherZone = cloudZones[0]
	}
	nodeName = pod.Spec.NodeName
	node, err = c.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	newPodZone := node.Labels[apis.LabelZoneFailureDomain]
	Expect(newPodZone).To(Equal(otherZone),
		"The pod should be scheduled in zone %s after all nodes in zone %s have been deleted", otherZone, podZone)

}

func getPVC(c clientset.Interface, ns string, pvcLabels map[string]string) *v1.PersistentVolumeClaim {
	selector := labels.Set(pvcLabels).AsSelector()
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pvcList, err := c.CoreV1().PersistentVolumeClaims(ns).List(options)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(pvcList.Items)).To(Equal(1), "There should be exactly 1 PVC matched.")

	return &pvcList.Items[0]
}

func getPod(c clientset.Interface, ns string, podLabels map[string]string) *v1.Pod {
	selector := labels.Set(podLabels).AsSelector()
	options := metav1.ListOptions{LabelSelector: selector.String()}
	podList, err := c.CoreV1().Pods(ns).List(options)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(podList.Items)).To(Equal(1), "There should be exactly 1 pod matched.")

	return &podList.Items[0]
}

// Generates the spec of a StatefulSet with 1 replica that mounts a Regional PD.
func newStatefulSet(claimTemplate *v1.PersistentVolumeClaim, ns string) (sts *appsv1.StatefulSet, svc *v1.Service, labels map[string]string) {
	var replicas int32 = 1
	labels = map[string]string{"app": "regional-pd-workload"}

	svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regional-pd-service",
			Namespace: ns,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Port: 80,
				Name: "web",
			}},
			ClusterIP: v1.ClusterIPNone,
			Selector:  labels,
		},
	}

	sts = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regional-pd-sts",
			Namespace: ns,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName:          svc.Name,
			Replicas:             &replicas,
			Template:             *newPodTemplate(labels),
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{*claimTemplate},
		},
	}

	return
}

func newPodTemplate(labels map[string]string) *v1.PodTemplateSpec {
	return &v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				// This container writes its pod name to a file in the Regional PD
				// and prints the entire file to stdout.
				{
					Name:    "busybox",
					Image:   "k8s.gcr.io/busybox",
					Command: []string{"sh", "-c"},
					Args: []string{
						"echo ${POD_NAME} >> /mnt/data/regional-pd/pods.txt;" +
							"cat /mnt/data/regional-pd/pods.txt;" +
							"sleep 3600;",
					},
					Env: []v1.EnvVar{{
						Name: "POD_NAME",
						ValueFrom: &v1.EnvVarSource{
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					}},
					Ports: []v1.ContainerPort{{
						ContainerPort: 80,
						Name:          "web",
					}},
					VolumeMounts: []v1.VolumeMount{{
						Name:      "regional-pd-vol",
						MountPath: "/mnt/data/regional-pd",
					}},
				},
			},
		},
	}
}

func newClaimTemplate(ns string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regional-pd-vol",
			Namespace: ns,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func newRegionalStorageClass(namespace string, zones []string) *storage.StorageClass {
	return &storage.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind: "StorageClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace + "-sc",
		},
		Provisioner: "kubernetes.io/gce-pd",
		Parameters: map[string]string{
			"type":             "pd-standard",
			"zones":            strings.Join(zones, ","),
			"replication-type": "regional-pd",
		},
	}
}

func getTwoRandomZones(c clientset.Interface) []string {
	zones, err := framework.GetClusterZones(c)
	Expect(err).ToNot(HaveOccurred())
	Expect(zones.Len()).To(BeNumerically(">=", 2),
		"The test should only be run in multizone clusters.")

	zone1, _ := zones.PopAny()
	zone2, _ := zones.PopAny()
	return []string{zone1, zone2}
}

// Waits for at least 1 replica of a StatefulSet to become not ready or until timeout occurs, whichever comes first.
func waitForStatefulSetReplicasNotReady(statefulSetName, ns string, c clientset.Interface) error {
	const poll = 3 * time.Second
	const timeout = statefulSetReadyTimeout

	framework.Logf("Waiting up to %v for StatefulSet %s to have at least 1 replica to become not ready", timeout, statefulSetName)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		sts, err := c.AppsV1().StatefulSets(ns).Get(statefulSetName, metav1.GetOptions{})
		if err != nil {
			framework.Logf("Get StatefulSet %s failed, ignoring for %v: %v", statefulSetName, poll, err)
			continue
		} else {
			if sts.Status.ReadyReplicas < *sts.Spec.Replicas {
				framework.Logf("%d replicas are ready out of a total of %d replicas in StatefulSet %s. (%v)",
					sts.Status.ReadyReplicas, *sts.Spec.Replicas, statefulSetName, time.Since(start))
				return nil
			} else {
				framework.Logf("StatefulSet %s found but there are %d ready replicas and %d total replicas.", statefulSetName, sts.Status.ReadyReplicas, *sts.Spec.Replicas)
			}
		}
	}
	return fmt.Errorf("All replicas in StatefulSet %s are still ready within %v", statefulSetName, timeout)
}

// If match is true, check if zones in PV exactly match zones given.
// Otherwise, check whether zones in PV is superset of zones given.
func verifyZonesInPV(volume *v1.PersistentVolume, zones sets.String, match bool) error {
	pvZones, err := util.LabelZonesToSet(volume.Labels[apis.LabelZoneFailureDomain])
	if err != nil {
		return err
	}

	if match && zones.Equal(pvZones) || !match && zones.IsSuperset(pvZones) {
		return nil
	}

	return fmt.Errorf("Zones in StorageClass are %v, but zones in PV are %v", zones, pvZones)

}
