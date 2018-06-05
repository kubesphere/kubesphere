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

package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type dnsTestCommon struct {
	f      *framework.Framework
	c      clientset.Interface
	ns     string
	name   string
	labels []string

	dnsPod       *v1.Pod
	utilPod      *v1.Pod
	utilService  *v1.Service
	dnsServerPod *v1.Pod

	cm *v1.ConfigMap
}

func newDnsTestCommon() dnsTestCommon {
	return dnsTestCommon{
		f:    framework.NewDefaultFramework("dns-config-map"),
		ns:   "kube-system",
		name: "kube-dns",
	}
}

func (t *dnsTestCommon) init() {
	By("Finding a DNS pod")
	label := labels.SelectorFromSet(labels.Set(map[string]string{"k8s-app": "kube-dns"}))
	options := metav1.ListOptions{LabelSelector: label.String()}

	pods, err := t.f.ClientSet.CoreV1().Pods("kube-system").List(options)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(pods.Items)).Should(BeNumerically(">=", 1))

	t.dnsPod = &pods.Items[0]
	framework.Logf("Using DNS pod: %v", t.dnsPod.Name)
}

func (t *dnsTestCommon) checkDNSRecord(name string, predicate func([]string) bool, timeout time.Duration) {
	t.checkDNSRecordFrom(name, predicate, "kube-dns", timeout)
}

func (t *dnsTestCommon) checkDNSRecordFrom(name string, predicate func([]string) bool, target string, timeout time.Duration) {
	var actual []string

	err := wait.PollImmediate(
		time.Duration(1)*time.Second,
		timeout,
		func() (bool, error) {
			actual = t.runDig(name, target)
			if predicate(actual) {
				return true, nil
			}
			return false, nil
		})

	if err != nil {
		framework.Failf("dig result did not match: %#v after %v",
			actual, timeout)
	}
}

// runDig queries for `dnsName`. Returns a list of responses.
func (t *dnsTestCommon) runDig(dnsName, target string) []string {
	cmd := []string{"/usr/bin/dig", "+short"}
	switch target {
	case "kube-dns":
		cmd = append(cmd, "@"+t.dnsPod.Status.PodIP, "-p", "10053")
	case "dnsmasq":
		break
	default:
		panic(fmt.Errorf("invalid target: " + target))
	}
	if strings.HasSuffix(dnsName, "in-addr.arpa") || strings.HasSuffix(dnsName, "in-addr.arpa.") {
		cmd = append(cmd, []string{"-t", "ptr"}...)
	}
	cmd = append(cmd, dnsName)

	stdout, stderr, err := t.f.ExecWithOptions(framework.ExecOptions{
		Command:       cmd,
		Namespace:     t.f.Namespace.Name,
		PodName:       t.utilPod.Name,
		ContainerName: "util",
		CaptureStdout: true,
		CaptureStderr: true,
	})

	framework.Logf("Running dig: %v, stdout: %q, stderr: %q, err: %v",
		cmd, stdout, stderr, err)

	if stdout == "" {
		return []string{}
	} else {
		return strings.Split(stdout, "\n")
	}
}

func (t *dnsTestCommon) setConfigMap(cm *v1.ConfigMap) {
	if t.cm != nil {
		t.cm = cm
	}

	cm.ObjectMeta.Namespace = t.ns
	cm.ObjectMeta.Name = t.name

	options := metav1.ListOptions{
		FieldSelector: fields.Set{
			"metadata.namespace": t.ns,
			"metadata.name":      t.name,
		}.AsSelector().String(),
	}
	cmList, err := t.c.CoreV1().ConfigMaps(t.ns).List(options)
	Expect(err).NotTo(HaveOccurred())

	if len(cmList.Items) == 0 {
		By(fmt.Sprintf("Creating the ConfigMap (%s:%s) %+v", t.ns, t.name, *cm))
		_, err := t.c.CoreV1().ConfigMaps(t.ns).Create(cm)
		Expect(err).NotTo(HaveOccurred())
	} else {
		By(fmt.Sprintf("Updating the ConfigMap (%s:%s) to %+v", t.ns, t.name, *cm))
		_, err := t.c.CoreV1().ConfigMaps(t.ns).Update(cm)
		Expect(err).NotTo(HaveOccurred())
	}
}

func (t *dnsTestCommon) deleteConfigMap() {
	By(fmt.Sprintf("Deleting the ConfigMap (%s:%s)", t.ns, t.name))
	t.cm = nil
	err := t.c.CoreV1().ConfigMaps(t.ns).Delete(t.name, nil)
	Expect(err).NotTo(HaveOccurred())
}

func (t *dnsTestCommon) createUtilPod() {
	// Actual port # doesn't matter, just needs to exist.
	const servicePort = 10101

	t.utilPod = &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    t.f.Namespace.Name,
			Labels:       map[string]string{"app": "e2e-dns-configmap"},
			GenerateName: "e2e-dns-configmap-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "util",
					Image:   imageutils.GetE2EImage(imageutils.Dnsutils),
					Command: []string{"sleep", "10000"},
					Ports: []v1.ContainerPort{
						{ContainerPort: servicePort, Protocol: "TCP"},
					},
				},
			},
		},
	}

	var err error
	t.utilPod, err = t.c.CoreV1().Pods(t.f.Namespace.Name).Create(t.utilPod)
	Expect(err).NotTo(HaveOccurred())
	framework.Logf("Created pod %v", t.utilPod)
	Expect(t.f.WaitForPodRunning(t.utilPod.Name)).NotTo(HaveOccurred())

	t.utilService = &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.f.Namespace.Name,
			Name:      "e2e-dns-configmap",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{"app": "e2e-dns-configmap"},
			Ports: []v1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       servicePort,
					TargetPort: intstr.FromInt(servicePort),
				},
			},
		},
	}

	t.utilService, err = t.c.CoreV1().Services(t.f.Namespace.Name).Create(t.utilService)
	Expect(err).NotTo(HaveOccurred())
	framework.Logf("Created service %v", t.utilService)
}

func (t *dnsTestCommon) deleteUtilPod() {
	podClient := t.c.CoreV1().Pods(t.f.Namespace.Name)
	if err := podClient.Delete(t.utilPod.Name, metav1.NewDeleteOptions(0)); err != nil {
		framework.Logf("Delete of pod %v:%v failed: %v",
			t.utilPod.Namespace, t.utilPod.Name, err)
	}
}

func generateDNSServerPod(aRecords map[string]string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "e2e-dns-configmap-dns-server-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "dns",
					Image: imageutils.GetE2EImage(imageutils.DNSMasq),
					Command: []string{
						"/usr/sbin/dnsmasq",
						"-u", "root",
						"-k",
						"--log-facility", "-",
						"-q",
					},
				},
			},
			DNSPolicy: "Default",
		},
	}

	for name, ip := range aRecords {
		pod.Spec.Containers[0].Command = append(
			pod.Spec.Containers[0].Command,
			fmt.Sprintf("-A/%v/%v", name, ip))
	}
	return pod
}

func (t *dnsTestCommon) createDNSPodFromObj(pod *v1.Pod) {
	t.dnsServerPod = pod

	var err error
	t.dnsServerPod, err = t.c.CoreV1().Pods(t.f.Namespace.Name).Create(t.dnsServerPod)
	Expect(err).NotTo(HaveOccurred())
	framework.Logf("Created pod %v", t.dnsServerPod)
	Expect(t.f.WaitForPodRunning(t.dnsServerPod.Name)).NotTo(HaveOccurred())

	t.dnsServerPod, err = t.c.CoreV1().Pods(t.f.Namespace.Name).Get(
		t.dnsServerPod.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func (t *dnsTestCommon) createDNSServer(aRecords map[string]string) {
	t.createDNSPodFromObj(generateDNSServerPod(aRecords))
}

func (t *dnsTestCommon) createDNSServerWithPtrRecord() {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "e2e-dns-configmap-dns-server-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "dns",
					Image: imageutils.GetE2EImage(imageutils.DNSMasq),
					Command: []string{
						"/usr/sbin/dnsmasq",
						"-u", "root",
						"-k",
						"--log-facility", "-",
						"--host-record=my.test,192.0.2.123",
						"-q",
					},
				},
			},
			DNSPolicy: "Default",
		},
	}

	t.createDNSPodFromObj(pod)
}

func (t *dnsTestCommon) deleteDNSServerPod() {
	podClient := t.c.CoreV1().Pods(t.f.Namespace.Name)
	if err := podClient.Delete(t.dnsServerPod.Name, metav1.NewDeleteOptions(0)); err != nil {
		framework.Logf("Delete of pod %v:%v failed: %v",
			t.utilPod.Namespace, t.dnsServerPod.Name, err)
	}
}

func createDNSPod(namespace, wheezyProbeCmd, jessieProbeCmd, podHostName, serviceName string) *v1.Pod {
	dnsPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dns-test-" + string(uuid.NewUUID()),
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "results",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
			Containers: []v1.Container{
				// TODO: Consider scraping logs instead of running a webserver.
				{
					Name:  "webserver",
					Image: imageutils.GetE2EImage(imageutils.TestWebserver),
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "results",
							MountPath: "/results",
						},
					},
				},
				{
					Name:    "querier",
					Image:   imageutils.GetE2EImage(imageutils.Dnsutils),
					Command: []string{"sh", "-c", wheezyProbeCmd},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "results",
							MountPath: "/results",
						},
					},
				},
				{
					Name:    "jessie-querier",
					Image:   imageutils.GetE2EImage(imageutils.JessieDnsutils),
					Command: []string{"sh", "-c", jessieProbeCmd},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "results",
							MountPath: "/results",
						},
					},
				},
			},
		},
	}

	dnsPod.Spec.Hostname = podHostName
	dnsPod.Spec.Subdomain = serviceName

	return dnsPod
}

func createProbeCommand(namesToResolve []string, hostEntries []string, ptrLookupIP string, fileNamePrefix, namespace string) (string, []string) {
	fileNames := make([]string, 0, len(namesToResolve)*2)
	probeCmd := "for i in `seq 1 600`; do "
	for _, name := range namesToResolve {
		// Resolve by TCP and UDP DNS.  Use $$(...) because $(...) is
		// expanded by kubernetes (though this won't expand so should
		// remain a literal, safe > sorry).
		lookup := "A"
		if strings.HasPrefix(name, "_") {
			lookup = "SRV"
		}
		fileName := fmt.Sprintf("%s_udp@%s", fileNamePrefix, name)
		fileNames = append(fileNames, fileName)
		probeCmd += fmt.Sprintf(`test -n "$$(dig +notcp +noall +answer +search %s %s)" && echo OK > /results/%s;`, name, lookup, fileName)
		fileName = fmt.Sprintf("%s_tcp@%s", fileNamePrefix, name)
		fileNames = append(fileNames, fileName)
		probeCmd += fmt.Sprintf(`test -n "$$(dig +tcp +noall +answer +search %s %s)" && echo OK > /results/%s;`, name, lookup, fileName)
	}

	for _, name := range hostEntries {
		fileName := fmt.Sprintf("%s_hosts@%s", fileNamePrefix, name)
		fileNames = append(fileNames, fileName)
		probeCmd += fmt.Sprintf(`test -n "$$(getent hosts %s)" && echo OK > /results/%s;`, name, fileName)
	}

	podARecByUDPFileName := fmt.Sprintf("%s_udp@PodARecord", fileNamePrefix)
	podARecByTCPFileName := fmt.Sprintf("%s_tcp@PodARecord", fileNamePrefix)
	probeCmd += fmt.Sprintf(`podARec=$$(hostname -i| awk -F. '{print $$1"-"$$2"-"$$3"-"$$4".%s.pod.cluster.local"}');`, namespace)
	probeCmd += fmt.Sprintf(`test -n "$$(dig +notcp +noall +answer +search $${podARec} A)" && echo OK > /results/%s;`, podARecByUDPFileName)
	probeCmd += fmt.Sprintf(`test -n "$$(dig +tcp +noall +answer +search $${podARec} A)" && echo OK > /results/%s;`, podARecByTCPFileName)
	fileNames = append(fileNames, podARecByUDPFileName)
	fileNames = append(fileNames, podARecByTCPFileName)

	if len(ptrLookupIP) > 0 {
		ptrLookup := fmt.Sprintf("%s.in-addr.arpa.", strings.Join(reverseArray(strings.Split(ptrLookupIP, ".")), "."))
		ptrRecByUDPFileName := fmt.Sprintf("%s_udp@PTR", ptrLookupIP)
		ptrRecByTCPFileName := fmt.Sprintf("%s_tcp@PTR", ptrLookupIP)
		probeCmd += fmt.Sprintf(`test -n "$$(dig +notcp +noall +answer +search %s PTR)" && echo OK > /results/%s;`, ptrLookup, ptrRecByUDPFileName)
		probeCmd += fmt.Sprintf(`test -n "$$(dig +tcp +noall +answer +search %s PTR)" && echo OK > /results/%s;`, ptrLookup, ptrRecByTCPFileName)
		fileNames = append(fileNames, ptrRecByUDPFileName)
		fileNames = append(fileNames, ptrRecByTCPFileName)
	}

	probeCmd += "sleep 1; done"
	return probeCmd, fileNames
}

// createTargetedProbeCommand returns a command line that performs a DNS lookup for a specific record type
func createTargetedProbeCommand(nameToResolve string, lookup string, fileNamePrefix string) (string, string) {
	fileName := fmt.Sprintf("%s_udp@%s", fileNamePrefix, nameToResolve)
	probeCmd := fmt.Sprintf("dig +short +tries=12 +norecurse %s %s > /results/%s", nameToResolve, lookup, fileName)
	return probeCmd, fileName
}

func assertFilesExist(fileNames []string, fileDir string, pod *v1.Pod, client clientset.Interface) {
	assertFilesContain(fileNames, fileDir, pod, client, false, "")
}

func assertFilesContain(fileNames []string, fileDir string, pod *v1.Pod, client clientset.Interface, check bool, expected string) {
	var failed []string

	framework.ExpectNoError(wait.Poll(time.Second*10, time.Second*600, func() (bool, error) {
		failed = []string{}

		ctx, cancel := context.WithTimeout(context.Background(), framework.SingleCallTimeout)
		defer cancel()

		for _, fileName := range fileNames {
			contents, err := client.CoreV1().RESTClient().Get().
				Context(ctx).
				Namespace(pod.Namespace).
				Resource("pods").
				SubResource("proxy").
				Name(pod.Name).
				Suffix(fileDir, fileName).
				Do().Raw()

			if err != nil {
				if ctx.Err() != nil {
					framework.Failf("Unable to read %s from pod %s: %v", fileName, pod.Name, err)
				} else {
					framework.Logf("Unable to read %s from pod %s: %v", fileName, pod.Name, err)
				}
				failed = append(failed, fileName)
			} else if check && strings.TrimSpace(string(contents)) != expected {
				framework.Logf("File %s from pod %s contains '%s' instead of '%s'", fileName, pod.Name, string(contents), expected)
				failed = append(failed, fileName)
			}
		}
		if len(failed) == 0 {
			return true, nil
		}
		framework.Logf("Lookups using %s failed for: %v\n", pod.Name, failed)
		return false, nil
	}))
	Expect(len(failed)).To(Equal(0))
}

func validateDNSResults(f *framework.Framework, pod *v1.Pod, fileNames []string) {
	By("submitting the pod to kubernetes")
	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
	defer func() {
		By("deleting the pod")
		defer GinkgoRecover()
		podClient.Delete(pod.Name, metav1.NewDeleteOptions(0))
	}()
	if _, err := podClient.Create(pod); err != nil {
		framework.Failf("Failed to create %s pod: %v", pod.Name, err)
	}

	framework.ExpectNoError(f.WaitForPodRunning(pod.Name))

	By("retrieving the pod")
	pod, err := podClient.Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		framework.Failf("Failed to get pod %s: %v", pod.Name, err)
	}
	// Try to find results for each expected name.
	By("looking for the results for each expected name from probers")
	assertFilesExist(fileNames, "results", pod, f.ClientSet)

	// TODO: probe from the host, too.

	framework.Logf("DNS probes using %s succeeded\n", pod.Name)
}

func validateTargetedProbeOutput(f *framework.Framework, pod *v1.Pod, fileNames []string, value string) {
	By("submitting the pod to kubernetes")
	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
	defer func() {
		By("deleting the pod")
		defer GinkgoRecover()
		podClient.Delete(pod.Name, metav1.NewDeleteOptions(0))
	}()
	if _, err := podClient.Create(pod); err != nil {
		framework.Failf("Failed to create %s pod: %v", pod.Name, err)
	}

	framework.ExpectNoError(f.WaitForPodRunning(pod.Name))

	By("retrieving the pod")
	pod, err := podClient.Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		framework.Failf("Failed to get pod %s: %v", pod.Name, err)
	}
	// Try to find the expected value for each expected name.
	By("looking for the results for each expected name from probers")
	assertFilesContain(fileNames, "results", pod, f.ClientSet, true, value)

	framework.Logf("DNS probes using %s succeeded\n", pod.Name)
}

func reverseArray(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func generateDNSUtilsPod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "e2e-dns-utils-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "util",
					Image:   imageutils.GetE2EImage(imageutils.Dnsutils),
					Command: []string{"sleep", "10000"},
				},
			},
		},
	}
}
