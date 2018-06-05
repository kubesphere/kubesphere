/*
Copyright 2014 The Kubernetes Authors.

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

package internalversion

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	versionedfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/printers"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
)

type describeClient struct {
	T         *testing.T
	Namespace string
	Err       error
	internalclientset.Interface
}

func TestDescribePod(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := PodDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bar") || !strings.Contains(out, "Status:") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribePodNode(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Spec: api.PodSpec{
			NodeName: "all-in-one",
		},
		Status: api.PodStatus{
			HostIP:            "127.0.0.1",
			NominatedNodeName: "nodeA",
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := PodDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "all-in-one/127.0.0.1") {
		t.Errorf("unexpected out: %s", out)
	}
	if !strings.Contains(out, "nodeA") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribePodTolerations(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Spec: api.PodSpec{
			Tolerations: []api.Toleration{
				{Key: "key0", Operator: api.TolerationOpExists},
				{Key: "key1", Value: "value1"},
				{Key: "key2", Operator: api.TolerationOpEqual, Value: "value2", Effect: api.TaintEffectNoSchedule},
				{Key: "key3", Value: "value3", Effect: api.TaintEffectNoExecute, TolerationSeconds: &[]int64{300}[0]},
				{Key: "key4", Effect: api.TaintEffectNoExecute, TolerationSeconds: &[]int64{60}[0]},
			},
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := PodDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "key0\n") ||
		!strings.Contains(out, "key1=value1\n") ||
		!strings.Contains(out, "key2=value2:NoSchedule\n") ||
		!strings.Contains(out, "key3=value3:NoExecute for 300s\n") ||
		!strings.Contains(out, "key4:NoExecute for 60s\n") ||
		!strings.Contains(out, "Tolerations:") {
		t.Errorf("unexpected out:\n%s", out)
	}
}

func TestDescribeSecret(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Data: map[string][]byte{
			"username": []byte("YWRtaW4="),
			"password": []byte("MWYyZDFlMmU2N2Rm"),
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := SecretDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bar") || !strings.Contains(out, "foo") || !strings.Contains(out, "username") || !strings.Contains(out, "8 bytes") || !strings.Contains(out, "password") || !strings.Contains(out, "16 bytes") {
		t.Errorf("unexpected out: %s", out)
	}
	if strings.Contains(out, "YWRtaW4=") || strings.Contains(out, "MWYyZDFlMmU2N2Rm") {
		t.Errorf("sensitive data should not be shown, unexpected out: %s", out)
	}
}

func TestDescribeNamespace(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myns",
		},
	})
	c := &describeClient{T: t, Namespace: "", Interface: fake}
	d := NamespaceDescriber{c}
	out, err := d.Describe("", "myns", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "myns") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribePodPriority(t *testing.T) {
	priority := int32(1000)
	fake := fake.NewSimpleClientset(&api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
		Spec: api.PodSpec{
			PriorityClassName: "high-priority",
			Priority:          &priority,
		},
	})
	c := &describeClient{T: t, Namespace: "", Interface: fake}
	d := PodDescriber{c}
	out, err := d.Describe("", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "high-priority") || !strings.Contains(out, "1000") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribeConfigMap(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mycm",
			Namespace: "foo",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := ConfigMapDescriber{c}
	out, err := d.Describe("foo", "mycm", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") || !strings.Contains(out, "mycm") || !strings.Contains(out, "key1") || !strings.Contains(out, "value1") || !strings.Contains(out, "key2") || !strings.Contains(out, "value2") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribeLimitRange(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mylr",
			Namespace: "foo",
		},
		Spec: api.LimitRangeSpec{
			Limits: []api.LimitRangeItem{
				{
					Type:                 api.LimitTypePod,
					Max:                  getResourceList("100m", "10000Mi"),
					Min:                  getResourceList("5m", "100Mi"),
					MaxLimitRequestRatio: getResourceList("10", ""),
				},
				{
					Type:                 api.LimitTypeContainer,
					Max:                  getResourceList("100m", "10000Mi"),
					Min:                  getResourceList("5m", "100Mi"),
					Default:              getResourceList("50m", "500Mi"),
					DefaultRequest:       getResourceList("10m", "200Mi"),
					MaxLimitRequestRatio: getResourceList("10", ""),
				},
				{
					Type: api.LimitTypePersistentVolumeClaim,
					Max:  getStorageResourceList("10Gi"),
					Min:  getStorageResourceList("5Gi"),
				},
			},
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := LimitRangeDescriber{c}
	out, err := d.Describe("foo", "mylr", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	checks := []string{"foo", "mylr", "Pod", "cpu", "5m", "100m", "memory", "100Mi", "10000Mi", "10", "Container", "cpu", "10m", "50m", "200Mi", "500Mi", "PersistentVolumeClaim", "storage", "5Gi", "10Gi"}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("unexpected out: %s", out)
		}
	}
}

func getStorageResourceList(storage string) api.ResourceList {
	res := api.ResourceList{}
	if storage != "" {
		res[api.ResourceStorage] = resource.MustParse(storage)
	}
	return res
}

func getResourceList(cpu, memory string) api.ResourceList {
	res := api.ResourceList{}
	if cpu != "" {
		res[api.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		res[api.ResourceMemory] = resource.MustParse(memory)
	}
	return res
}

func TestDescribeService(t *testing.T) {
	testCases := []struct {
		service *api.Service
		expect  []string
	}{
		{
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
				Spec: api.ServiceSpec{
					Type: api.ServiceTypeLoadBalancer,
					Ports: []api.ServicePort{{
						Name:       "port-tcp",
						Port:       8080,
						Protocol:   api.ProtocolTCP,
						TargetPort: intstr.FromInt(9527),
						NodePort:   31111,
					}},
					Selector:              map[string]string{"blah": "heh"},
					ClusterIP:             "1.2.3.4",
					LoadBalancerIP:        "5.6.7.8",
					SessionAffinity:       "None",
					ExternalTrafficPolicy: "Local",
					HealthCheckNodePort:   32222,
				},
			},
			expect: []string{
				"Name", "bar",
				"Namespace", "foo",
				"Selector", "blah=heh",
				"Type", "LoadBalancer",
				"IP", "1.2.3.4",
				"Port", "port-tcp", "8080/TCP",
				"TargetPort", "9527/TCP",
				"NodePort", "port-tcp", "31111/TCP",
				"Session Affinity", "None",
				"External Traffic Policy", "Local",
				"HealthCheck NodePort", "32222",
			},
		},
		{
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
				Spec: api.ServiceSpec{
					Type: api.ServiceTypeLoadBalancer,
					Ports: []api.ServicePort{{
						Name:       "port-tcp",
						Port:       8080,
						Protocol:   api.ProtocolTCP,
						TargetPort: intstr.FromString("targetPort"),
						NodePort:   31111,
					}},
					Selector:              map[string]string{"blah": "heh"},
					ClusterIP:             "1.2.3.4",
					LoadBalancerIP:        "5.6.7.8",
					SessionAffinity:       "None",
					ExternalTrafficPolicy: "Local",
					HealthCheckNodePort:   32222,
				},
			},
			expect: []string{
				"Name", "bar",
				"Namespace", "foo",
				"Selector", "blah=heh",
				"Type", "LoadBalancer",
				"IP", "1.2.3.4",
				"Port", "port-tcp", "8080/TCP",
				"TargetPort", "targetPort/TCP",
				"NodePort", "port-tcp", "31111/TCP",
				"Session Affinity", "None",
				"External Traffic Policy", "Local",
				"HealthCheck NodePort", "32222",
			},
		},
	}
	for _, testCase := range testCases {
		fake := fake.NewSimpleClientset(testCase.service)
		c := &describeClient{T: t, Namespace: "foo", Interface: fake}
		d := ServiceDescriber{c}
		out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for _, expected := range testCase.expect {
			if !strings.Contains(out, expected) {
				t.Errorf("expected to find %q in output: %q", expected, out)
			}
		}
	}
}

func TestPodDescribeResultsSorted(t *testing.T) {
	// Arrange
	fake := fake.NewSimpleClientset(
		&api.EventList{
			Items: []api.Event{
				{
					ObjectMeta:     metav1.ObjectMeta{Name: "one"},
					Source:         api.EventSource{Component: "kubelet"},
					Message:        "Item 1",
					FirstTimestamp: metav1.NewTime(time.Date(2014, time.January, 15, 0, 0, 0, 0, time.UTC)),
					LastTimestamp:  metav1.NewTime(time.Date(2014, time.January, 15, 0, 0, 0, 0, time.UTC)),
					Count:          1,
					Type:           api.EventTypeNormal,
				},
				{
					ObjectMeta:     metav1.ObjectMeta{Name: "two"},
					Source:         api.EventSource{Component: "scheduler"},
					Message:        "Item 2",
					FirstTimestamp: metav1.NewTime(time.Date(1987, time.June, 17, 0, 0, 0, 0, time.UTC)),
					LastTimestamp:  metav1.NewTime(time.Date(1987, time.June, 17, 0, 0, 0, 0, time.UTC)),
					Count:          1,
					Type:           api.EventTypeNormal,
				},
				{
					ObjectMeta:     metav1.ObjectMeta{Name: "three"},
					Source:         api.EventSource{Component: "kubelet"},
					Message:        "Item 3",
					FirstTimestamp: metav1.NewTime(time.Date(2002, time.December, 25, 0, 0, 0, 0, time.UTC)),
					LastTimestamp:  metav1.NewTime(time.Date(2002, time.December, 25, 0, 0, 0, 0, time.UTC)),
					Count:          1,
					Type:           api.EventTypeNormal,
				},
			},
		},
		&api.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
	)
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := PodDescriber{c}

	// Act
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	VerifyDatesInOrder(out, "\n" /* rowDelimiter */, "\t" /* columnDelimiter */, t)
}

// VerifyDatesInOrder checks the start of each line for a RFC1123Z date
// and posts error if all subsequent dates are not equal or increasing
func VerifyDatesInOrder(
	resultToTest, rowDelimiter, columnDelimiter string, t *testing.T) {
	lines := strings.Split(resultToTest, rowDelimiter)
	var previousTime time.Time
	for _, str := range lines {
		columns := strings.Split(str, columnDelimiter)
		if len(columns) > 0 {
			currentTime, err := time.Parse(time.RFC1123Z, columns[0])
			if err == nil {
				if previousTime.After(currentTime) {
					t.Errorf(
						"Output is not sorted by time. %s should be listed after %s. Complete output: %s",
						previousTime.Format(time.RFC1123Z),
						currentTime.Format(time.RFC1123Z),
						resultToTest)
				}
				previousTime = currentTime
			}
		}
	}
}

func TestDescribeContainers(t *testing.T) {
	trueVal := true
	testCases := []struct {
		container        api.Container
		status           api.ContainerStatus
		expectedElements []string
	}{
		// Running state.
		{
			container: api.Container{Name: "test", Image: "image"},
			status: api.ContainerStatus{
				Name: "test",
				State: api.ContainerState{
					Running: &api.ContainerStateRunning{
						StartedAt: metav1.NewTime(time.Now()),
					},
				},
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Running", "Ready", "True", "Restart Count", "7", "Image", "image", "Started"},
		},
		// Waiting state.
		{
			container: api.Container{Name: "test", Image: "image"},
			status: api.ContainerStatus{
				Name: "test",
				State: api.ContainerState{
					Waiting: &api.ContainerStateWaiting{
						Reason: "potato",
					},
				},
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "Reason", "potato"},
		},
		// Terminated state.
		{
			container: api.Container{Name: "test", Image: "image"},
			status: api.ContainerStatus{
				Name: "test",
				State: api.ContainerState{
					Terminated: &api.ContainerStateTerminated{
						StartedAt:  metav1.NewTime(time.Now()),
						FinishedAt: metav1.NewTime(time.Now()),
						Reason:     "potato",
						ExitCode:   2,
					},
				},
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Terminated", "Ready", "True", "Restart Count", "7", "Image", "image", "Reason", "potato", "Started", "Finished", "Exit Code", "2"},
		},
		// Last Terminated
		{
			container: api.Container{Name: "test", Image: "image"},
			status: api.ContainerStatus{
				Name: "test",
				State: api.ContainerState{
					Running: &api.ContainerStateRunning{
						StartedAt: metav1.NewTime(time.Now()),
					},
				},
				LastTerminationState: api.ContainerState{
					Terminated: &api.ContainerStateTerminated{
						StartedAt:  metav1.NewTime(time.Now().Add(time.Second * 3)),
						FinishedAt: metav1.NewTime(time.Now()),
						Reason:     "crashing",
						ExitCode:   3,
					},
				},
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Terminated", "Ready", "True", "Restart Count", "7", "Image", "image", "Started", "Finished", "Exit Code", "2", "crashing", "3"},
		},
		// No state defaults to waiting.
		{
			container: api.Container{Name: "test", Image: "image"},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image"},
		},
		// Env
		{
			container: api.Container{Name: "test", Image: "image", Env: []api.EnvVar{{Name: "envname", Value: "xyz"}}, EnvFrom: []api.EnvFromSource{{ConfigMapRef: &api.ConfigMapEnvSource{LocalObjectReference: api.LocalObjectReference{Name: "a123"}}}}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "envname", "xyz", "a123\tConfigMap\tOptional: false"},
		},
		{
			container: api.Container{Name: "test", Image: "image", Env: []api.EnvVar{{Name: "envname", Value: "xyz"}}, EnvFrom: []api.EnvFromSource{{Prefix: "p_", ConfigMapRef: &api.ConfigMapEnvSource{LocalObjectReference: api.LocalObjectReference{Name: "a123"}}}}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "envname", "xyz", "a123\tConfigMap with prefix 'p_'\tOptional: false"},
		},
		{
			container: api.Container{Name: "test", Image: "image", Env: []api.EnvVar{{Name: "envname", Value: "xyz"}}, EnvFrom: []api.EnvFromSource{{ConfigMapRef: &api.ConfigMapEnvSource{Optional: &trueVal, LocalObjectReference: api.LocalObjectReference{Name: "a123"}}}}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "envname", "xyz", "a123\tConfigMap\tOptional: true"},
		},
		{
			container: api.Container{Name: "test", Image: "image", Env: []api.EnvVar{{Name: "envname", Value: "xyz"}}, EnvFrom: []api.EnvFromSource{{SecretRef: &api.SecretEnvSource{LocalObjectReference: api.LocalObjectReference{Name: "a123"}, Optional: &trueVal}}}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "envname", "xyz", "a123\tSecret\tOptional: true"},
		},
		{
			container: api.Container{Name: "test", Image: "image", Env: []api.EnvVar{{Name: "envname", Value: "xyz"}}, EnvFrom: []api.EnvFromSource{{Prefix: "p_", SecretRef: &api.SecretEnvSource{LocalObjectReference: api.LocalObjectReference{Name: "a123"}}}}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "envname", "xyz", "a123\tSecret with prefix 'p_'\tOptional: false"},
		},
		// Command
		{
			container: api.Container{Name: "test", Image: "image", Command: []string{"sleep", "1000"}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "sleep", "1000"},
		},
		// Args
		{
			container: api.Container{Name: "test", Image: "image", Args: []string{"time", "1000"}},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"test", "State", "Waiting", "Ready", "True", "Restart Count", "7", "Image", "image", "time", "1000"},
		},
		// Using limits.
		{
			container: api.Container{
				Name:  "test",
				Image: "image",
				Resources: api.ResourceRequirements{
					Limits: api.ResourceList{
						api.ResourceName(api.ResourceCPU):     resource.MustParse("1000"),
						api.ResourceName(api.ResourceMemory):  resource.MustParse("4G"),
						api.ResourceName(api.ResourceStorage): resource.MustParse("20G"),
					},
				},
			},
			status: api.ContainerStatus{
				Name:         "test",
				Ready:        true,
				RestartCount: 7,
			},
			expectedElements: []string{"cpu", "1k", "memory", "4G", "storage", "20G"},
		},
		// Using requests.
		{
			container: api.Container{
				Name:  "test",
				Image: "image",
				Resources: api.ResourceRequirements{
					Requests: api.ResourceList{
						api.ResourceName(api.ResourceCPU):     resource.MustParse("1000"),
						api.ResourceName(api.ResourceMemory):  resource.MustParse("4G"),
						api.ResourceName(api.ResourceStorage): resource.MustParse("20G"),
					},
				},
			},
			expectedElements: []string{"cpu", "1k", "memory", "4G", "storage", "20G"},
		},
		// volumeMounts read/write
		{
			container: api.Container{
				Name:  "test",
				Image: "image",
				VolumeMounts: []api.VolumeMount{
					{
						Name:      "mounted-volume",
						MountPath: "/opt/",
					},
				},
			},
			expectedElements: []string{"mounted-volume", "/opt/", "(rw)"},
		},
		// volumeMounts readonly
		{
			container: api.Container{
				Name:  "test",
				Image: "image",
				VolumeMounts: []api.VolumeMount{
					{
						Name:      "mounted-volume",
						MountPath: "/opt/",
						ReadOnly:  true,
					},
				},
			},
			expectedElements: []string{"Mounts", "mounted-volume", "/opt/", "(ro)"},
		},

		// volumeDevices
		{
			container: api.Container{
				Name:  "test",
				Image: "image",
				VolumeDevices: []api.VolumeDevice{
					{
						Name:       "volume-device",
						DevicePath: "/dev/xvda",
					},
				},
			},
			expectedElements: []string{"Devices", "volume-device", "/dev/xvda"},
		},
	}

	for i, testCase := range testCases {
		out := new(bytes.Buffer)
		pod := api.Pod{
			Spec: api.PodSpec{
				Containers: []api.Container{testCase.container},
			},
			Status: api.PodStatus{
				ContainerStatuses: []api.ContainerStatus{testCase.status},
			},
		}
		writer := NewPrefixWriter(out)
		describeContainers("Containers", pod.Spec.Containers, pod.Status.ContainerStatuses, EnvValueRetriever(&pod), writer, "")
		output := out.String()
		for _, expected := range testCase.expectedElements {
			if !strings.Contains(output, expected) {
				t.Errorf("Test case %d: expected to find %q in output: %q", i, expected, output)
			}
		}
	}
}

func TestDescribers(t *testing.T) {
	first := &api.Event{}
	second := &api.Pod{}
	var third *api.Pod
	testErr := fmt.Errorf("test")
	d := Describers{}
	d.Add(
		func(e *api.Event, p *api.Pod) (string, error) {
			if e != first {
				t.Errorf("first argument not equal: %#v", e)
			}
			if p != second {
				t.Errorf("second argument not equal: %#v", p)
			}
			return "test", testErr
		},
	)
	if out, err := d.DescribeObject(first, second); out != "test" || err != testErr {
		t.Errorf("unexpected result: %s %v", out, err)
	}

	if out, err := d.DescribeObject(first, second, third); out != "" || err == nil {
		t.Errorf("unexpected result: %s %v", out, err)
	} else {
		if noDescriber, ok := err.(printers.ErrNoDescriber); ok {
			if !reflect.DeepEqual(noDescriber.Types, []string{"*core.Event", "*core.Pod", "*core.Pod"}) {
				t.Errorf("unexpected describer: %v", err)
			}
		} else {
			t.Errorf("unexpected error type: %v", err)
		}
	}

	d.Add(
		func(e *api.Event) (string, error) {
			if e != first {
				t.Errorf("first argument not equal: %#v", e)
			}
			return "simpler", testErr
		},
	)
	if out, err := d.DescribeObject(first); out != "simpler" || err != testErr {
		t.Errorf("unexpected result: %s %v", out, err)
	}
}

func TestDefaultDescribers(t *testing.T) {
	out, err := DefaultObjectDescriber.DescribeObject(&api.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("unexpected output: %s", out)
	}

	out, err = DefaultObjectDescriber.DescribeObject(&api.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("unexpected output: %s", out)
	}

	out, err = DefaultObjectDescriber.DescribeObject(&api.ReplicationController{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("unexpected output: %s", out)
	}

	out, err = DefaultObjectDescriber.DescribeObject(&api.Node{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestGetPodsTotalRequests(t *testing.T) {
	testCases := []struct {
		pods         *api.PodList
		expectedReqs map[api.ResourceName]resource.Quantity
	}{
		{
			pods: &api.PodList{
				Items: []api.Pod{
					{
						Spec: api.PodSpec{
							Containers: []api.Container{
								{
									Resources: api.ResourceRequirements{
										Requests: api.ResourceList{
											api.ResourceName(api.ResourceCPU):     resource.MustParse("1"),
											api.ResourceName(api.ResourceMemory):  resource.MustParse("300Mi"),
											api.ResourceName(api.ResourceStorage): resource.MustParse("1G"),
										},
									},
								},
								{
									Resources: api.ResourceRequirements{
										Requests: api.ResourceList{
											api.ResourceName(api.ResourceCPU):     resource.MustParse("90m"),
											api.ResourceName(api.ResourceMemory):  resource.MustParse("120Mi"),
											api.ResourceName(api.ResourceStorage): resource.MustParse("200M"),
										},
									},
								},
							},
						},
					},
					{
						Spec: api.PodSpec{
							Containers: []api.Container{
								{
									Resources: api.ResourceRequirements{
										Requests: api.ResourceList{
											api.ResourceName(api.ResourceCPU):     resource.MustParse("60m"),
											api.ResourceName(api.ResourceMemory):  resource.MustParse("43Mi"),
											api.ResourceName(api.ResourceStorage): resource.MustParse("500M"),
										},
									},
								},
								{
									Resources: api.ResourceRequirements{
										Requests: api.ResourceList{
											api.ResourceName(api.ResourceCPU):     resource.MustParse("34m"),
											api.ResourceName(api.ResourceMemory):  resource.MustParse("83Mi"),
											api.ResourceName(api.ResourceStorage): resource.MustParse("700M"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedReqs: map[api.ResourceName]resource.Quantity{
				api.ResourceName(api.ResourceCPU):     resource.MustParse("1.184"),
				api.ResourceName(api.ResourceMemory):  resource.MustParse("546Mi"),
				api.ResourceName(api.ResourceStorage): resource.MustParse("2.4G"),
			},
		},
	}

	for _, testCase := range testCases {
		reqs, _ := getPodsTotalRequestsAndLimits(testCase.pods)
		if !apiequality.Semantic.DeepEqual(reqs, testCase.expectedReqs) {
			t.Errorf("Expected %v, got %v", testCase.expectedReqs, reqs)
		}
	}
}

func TestPersistentVolumeDescriber(t *testing.T) {
	block := api.PersistentVolumeBlock
	file := api.PersistentVolumeFilesystem
	testCases := []struct {
		plugin             string
		pv                 *api.PersistentVolume
		expectedElements   []string
		unexpectedElements []string
	}{
		{
			plugin: "hostpath",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						HostPath: &api.HostPathVolumeSource{Type: new(api.HostPathType)},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "gce",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{},
					},
					VolumeMode: &file,
				},
			},
			expectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "ebs",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						AWSElasticBlockStore: &api.AWSElasticBlockStoreVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "nfs",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						NFS: &api.NFSVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "iscsi",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						ISCSI: &api.ISCSIPersistentVolumeSource{},
					},
					VolumeMode: &block,
				},
			},
			expectedElements: []string{"VolumeMode", "Block"},
		},
		{
			plugin: "gluster",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Glusterfs: &api.GlusterfsVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "rbd",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						RBD: &api.RBDPersistentVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "quobyte",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Quobyte: &api.QuobyteVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "cinder",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Cinder: &api.CinderVolumeSource{},
					},
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			plugin: "fc",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						FC: &api.FCVolumeSource{},
					},
					VolumeMode: &block,
				},
			},
			expectedElements: []string{"VolumeMode", "Block"},
		},
		{
			plugin: "local",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Local: &api.LocalVolumeSource{},
					},
				},
			},
			expectedElements:   []string{"Node Affinity:   <none>"},
			unexpectedElements: []string{"Required Terms", "Term "},
		},
		{
			plugin: "local",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Local: &api.LocalVolumeSource{},
					},
					NodeAffinity: &api.VolumeNodeAffinity{},
				},
			},
			expectedElements:   []string{"Node Affinity:   <none>"},
			unexpectedElements: []string{"Required Terms", "Term "},
		},
		{
			plugin: "local",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Local: &api.LocalVolumeSource{},
					},
					NodeAffinity: &api.VolumeNodeAffinity{
						Required: &api.NodeSelector{},
					},
				},
			},
			expectedElements:   []string{"Node Affinity", "Required Terms:  <none>"},
			unexpectedElements: []string{"Term "},
		},
		{
			plugin: "local",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Local: &api.LocalVolumeSource{},
					},
					NodeAffinity: &api.VolumeNodeAffinity{
						Required: &api.NodeSelector{
							NodeSelectorTerms: []api.NodeSelectorTerm{
								{
									MatchExpressions: []api.NodeSelectorRequirement{},
								},
								{
									MatchExpressions: []api.NodeSelectorRequirement{},
								},
							},
						},
					},
				},
			},
			expectedElements: []string{"Node Affinity", "Required Terms", "Term 0", "Term 1"},
		},
		{
			plugin: "local",
			pv: &api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
				Spec: api.PersistentVolumeSpec{
					PersistentVolumeSource: api.PersistentVolumeSource{
						Local: &api.LocalVolumeSource{},
					},
					NodeAffinity: &api.VolumeNodeAffinity{
						Required: &api.NodeSelector{
							NodeSelectorTerms: []api.NodeSelectorTerm{
								{
									MatchExpressions: []api.NodeSelectorRequirement{
										{
											Key:      "foo",
											Operator: "In",
											Values:   []string{"val1", "val2"},
										},
										{
											Key:      "foo",
											Operator: "Exists",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedElements: []string{"Node Affinity", "Required Terms", "Term 0",
				"foo in [val1, val2]",
				"foo exists"},
		},
	}

	for _, test := range testCases {
		fake := fake.NewSimpleClientset(test.pv)
		c := PersistentVolumeDescriber{fake}
		str, err := c.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
		if err != nil {
			t.Errorf("Unexpected error for test %s: %v", test.plugin, err)
		}
		if str == "" {
			t.Errorf("Unexpected empty string for test %s.  Expected PV Describer output", test.plugin)
		}
		for _, expected := range test.expectedElements {
			if !strings.Contains(str, expected) {
				t.Errorf("expected to find %q in output: %q", expected, str)
			}
		}
		for _, unexpected := range test.unexpectedElements {
			if strings.Contains(str, unexpected) {
				t.Errorf("unexpected to find %q in output: %q", unexpected, str)
			}
		}
	}
}

func TestPersistentVolumeClaimDescriber(t *testing.T) {
	block := api.PersistentVolumeBlock
	file := api.PersistentVolumeFilesystem
	goldClassName := "gold"
	now := time.Now()
	testCases := []struct {
		name               string
		pvc                *api.PersistentVolumeClaim
		expectedElements   []string
		unexpectedElements []string
	}{
		{
			name: "default",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume1",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Phase: api.ClaimBound,
				},
			},
			unexpectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			name: "filesystem",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume2",
					StorageClassName: &goldClassName,
					VolumeMode:       &file,
				},
				Status: api.PersistentVolumeClaimStatus{
					Phase: api.ClaimBound,
				},
			},
			expectedElements: []string{"VolumeMode", "Filesystem"},
		},
		{
			name: "block",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume3",
					StorageClassName: &goldClassName,
					VolumeMode:       &block,
				},
				Status: api.PersistentVolumeClaimStatus{
					Phase: api.ClaimBound,
				},
			},
			expectedElements: []string{"VolumeMode", "Block"},
		},
		// Tests for Status.Condition.
		{
			name: "condition-type",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume4",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{Type: api.PersistentVolumeClaimResizing},
					},
				},
			},
			expectedElements: []string{"Conditions", "Type", "Resizing"},
		},
		{
			name: "condition-status",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume5",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{Status: api.ConditionTrue},
					},
				},
			},
			expectedElements: []string{"Conditions", "Status", "True"},
		},
		{
			name: "condition-last-probe-time",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume6",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{LastProbeTime: metav1.Time{Time: now}},
					},
				},
			},
			expectedElements: []string{"Conditions", "LastProbeTime", now.Format(time.RFC1123Z)},
		},
		{
			name: "condition-last-transition-time",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume7",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{LastTransitionTime: metav1.Time{Time: now}},
					},
				},
			},
			expectedElements: []string{"Conditions", "LastTransitionTime", now.Format(time.RFC1123Z)},
		},
		{
			name: "condition-reason",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume8",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{Reason: "OfflineResize"},
					},
				},
			},
			expectedElements: []string{"Conditions", "Reason", "OfflineResize"},
		},
		{
			name: "condition-message",
			pvc: &api.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: api.PersistentVolumeClaimSpec{
					VolumeName:       "volume9",
					StorageClassName: &goldClassName,
				},
				Status: api.PersistentVolumeClaimStatus{
					Conditions: []api.PersistentVolumeClaimCondition{
						{Message: "User request resize"},
					},
				},
			},
			expectedElements: []string{"Conditions", "Message", "User request resize"},
		},
	}

	for _, test := range testCases {
		fake := fake.NewSimpleClientset(test.pvc)
		c := PersistentVolumeClaimDescriber{fake}
		str, err := c.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
		if err != nil {
			t.Errorf("Unexpected error for test %s: %v", test.name, err)
		}
		if str == "" {
			t.Errorf("Unexpected empty string for test %s.  Expected PVC Describer output", test.name)
		}
		for _, expected := range test.expectedElements {
			if !strings.Contains(str, expected) {
				t.Errorf("expected to find %q in output: %q", expected, str)
			}
		}
		for _, unexpected := range test.unexpectedElements {
			if strings.Contains(str, unexpected) {
				t.Errorf("unexpected to find %q in output: %q", unexpected, str)
			}
		}
	}
}

func TestDescribeDeployment(t *testing.T) {
	fake := fake.NewSimpleClientset()
	versionedFake := versionedfake.NewSimpleClientset(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: utilpointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Image: "mytest-image:latest"},
					},
				},
			},
		},
	})
	d := DeploymentDescriber{fake, versionedFake}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bar") || !strings.Contains(out, "foo") || !strings.Contains(out, "Containers:") || !strings.Contains(out, "mytest-image:latest") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribeStorageClass(t *testing.T) {
	reclaimPolicy := api.PersistentVolumeReclaimRetain
	bindingMode := storage.VolumeBindingMode("bindingmode")
	f := fake.NewSimpleClientset(&storage.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "foo",
			ResourceVersion: "4",
			Annotations: map[string]string{
				"name": "foo",
			},
		},
		Provisioner: "my-provisioner",
		Parameters: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		ReclaimPolicy:     &reclaimPolicy,
		VolumeBindingMode: &bindingMode,
	})
	s := StorageClassDescriber{f}
	out, err := s.Describe("", "foo", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo") ||
		!strings.Contains(out, "my-provisioner") ||
		!strings.Contains(out, "param1") ||
		!strings.Contains(out, "param2") ||
		!strings.Contains(out, "value1") ||
		!strings.Contains(out, "value2") ||
		!strings.Contains(out, "Retain") ||
		!strings.Contains(out, "bindingmode") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribePodDisruptionBudget(t *testing.T) {
	minAvailable := intstr.FromInt(22)
	f := fake.NewSimpleClientset(&policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "ns1",
			Name:              "pdb1",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(1.9e9)},
		},
		Spec: policy.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
		},
		Status: policy.PodDisruptionBudgetStatus{
			PodDisruptionsAllowed: 5,
		},
	})
	s := PodDisruptionBudgetDescriber{f}
	out, err := s.Describe("ns1", "pdb1", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "pdb1") ||
		!strings.Contains(out, "ns1") ||
		!strings.Contains(out, "22") ||
		!strings.Contains(out, "5") {
		t.Errorf("unexpected out: %s", out)
	}
}

func TestDescribeHorizontalPodAutoscaler(t *testing.T) {
	minReplicasVal := int32(2)
	targetUtilizationVal := int32(80)
	currentUtilizationVal := int32(50)
	tests := []struct {
		name string
		hpa  autoscaling.HorizontalPodAutoscaler
	}{
		{
			"minReplicas unset",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MaxReplicas: 10,
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"external source type, target average value (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricSource{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:         "some-external-metric",
								TargetAverageValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"external source type, target average value (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricSource{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:         "some-external-metric",
								TargetAverageValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricStatus{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:          "some-external-metric",
								CurrentAverageValue: resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"external source type, target value (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricSource{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:  "some-external-metric",
								TargetValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"external source type, target value (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricSource{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:  "some-external-metric",
								TargetValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.ExternalMetricSourceType,
							External: &autoscaling.ExternalMetricStatus{
								MetricSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"label": "value",
									},
								},
								MetricName:   "some-external-metric",
								CurrentValue: *resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"pods source type (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricSource{
								MetricName:         "some-pods-metric",
								TargetAverageValue: *resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"pods source type (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricSource{
								MetricName:         "some-pods-metric",
								TargetAverageValue: *resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricStatus{
								MetricName:          "some-pods-metric",
								CurrentAverageValue: *resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"object source type (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ObjectMetricSourceType,
							Object: &autoscaling.ObjectMetricSource{
								Target: autoscaling.CrossVersionObjectReference{
									Name: "some-service",
									Kind: "Service",
								},
								MetricName:  "some-service-metric",
								TargetValue: *resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"object source type (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ObjectMetricSourceType,
							Object: &autoscaling.ObjectMetricSource{
								Target: autoscaling.CrossVersionObjectReference{
									Name: "some-service",
									Kind: "Service",
								},
								MetricName:  "some-service-metric",
								TargetValue: *resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.ObjectMetricSourceType,
							Object: &autoscaling.ObjectMetricStatus{
								Target: autoscaling.CrossVersionObjectReference{
									Name: "some-service",
									Kind: "Service",
								},
								MetricName:   "some-service-metric",
								CurrentValue: *resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"resource source type, target average value (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricSource{
								Name:               api.ResourceCPU,
								TargetAverageValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"resource source type, target average value (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricSource{
								Name:               api.ResourceCPU,
								TargetAverageValue: resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricStatus{
								Name:                api.ResourceCPU,
								CurrentAverageValue: *resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"resource source type, target utilization (no current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricSource{
								Name: api.ResourceCPU,
								TargetAverageUtilization: &targetUtilizationVal,
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
				},
			},
		},
		{
			"resource source type, target utilization (with current)",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricSource{
								Name: api.ResourceCPU,
								TargetAverageUtilization: &targetUtilizationVal,
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricStatus{
								Name: api.ResourceCPU,
								CurrentAverageUtilization: &currentUtilizationVal,
								CurrentAverageValue:       *resource.NewMilliQuantity(40, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
		{
			"multiple metrics",
			autoscaling.HorizontalPodAutoscaler{
				Spec: autoscaling.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscaling.CrossVersionObjectReference{
						Name: "some-rc",
						Kind: "ReplicationController",
					},
					MinReplicas: &minReplicasVal,
					MaxReplicas: 10,
					Metrics: []autoscaling.MetricSpec{
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricSource{
								MetricName:         "some-pods-metric",
								TargetAverageValue: *resource.NewMilliQuantity(100, resource.DecimalSI),
							},
						},
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricSource{
								Name: api.ResourceCPU,
								TargetAverageUtilization: &targetUtilizationVal,
							},
						},
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricSource{
								MetricName:         "other-pods-metric",
								TargetAverageValue: *resource.NewMilliQuantity(400, resource.DecimalSI),
							},
						},
					},
				},
				Status: autoscaling.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 4,
					DesiredReplicas: 5,
					CurrentMetrics: []autoscaling.MetricStatus{
						{
							Type: autoscaling.PodsMetricSourceType,
							Pods: &autoscaling.PodsMetricStatus{
								MetricName:          "some-pods-metric",
								CurrentAverageValue: *resource.NewMilliQuantity(50, resource.DecimalSI),
							},
						},
						{
							Type: autoscaling.ResourceMetricSourceType,
							Resource: &autoscaling.ResourceMetricStatus{
								Name: api.ResourceCPU,
								CurrentAverageUtilization: &currentUtilizationVal,
								CurrentAverageValue:       *resource.NewMilliQuantity(40, resource.DecimalSI),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.hpa.ObjectMeta = metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		}
		fake := fake.NewSimpleClientset(&test.hpa)
		desc := HorizontalPodAutoscalerDescriber{fake}
		str, err := desc.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
		if err != nil {
			t.Errorf("Unexpected error for test %s: %v", test.name, err)
		}
		if str == "" {
			t.Errorf("Unexpected empty string for test %s.  Expected HPA Describer output", test.name)
		}
		t.Logf("Description for %q:\n%s", test.name, str)
	}
}

func TestDescribeEvents(t *testing.T) {

	events := &api.EventList{
		Items: []api.Event{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
				},
				Source:         api.EventSource{Component: "kubelet"},
				Message:        "Item 1",
				FirstTimestamp: metav1.NewTime(time.Date(2014, time.January, 15, 0, 0, 0, 0, time.UTC)),
				LastTimestamp:  metav1.NewTime(time.Date(2014, time.January, 15, 0, 0, 0, 0, time.UTC)),
				Count:          1,
				Type:           api.EventTypeNormal,
			},
		},
	}

	m := map[string]printers.Describer{
		"DaemonSetDescriber": &DaemonSetDescriber{
			fake.NewSimpleClientset(&extensions.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		"DeploymentDescriber": &DeploymentDescriber{
			fake.NewSimpleClientset(events),
			versionedfake.NewSimpleClientset(&v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
				Spec: v1beta1.DeploymentSpec{
					Replicas: utilpointer.Int32Ptr(1),
					Selector: &metav1.LabelSelector{},
				},
			}),
		},
		"EndpointsDescriber": &EndpointsDescriber{
			fake.NewSimpleClientset(&api.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		// TODO(jchaloup): add tests for:
		// - IngressDescriber
		// - JobDescriber
		"NodeDescriber": &NodeDescriber{
			fake.NewSimpleClientset(&api.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:     "bar",
					SelfLink: "url/url/url",
				},
			}, events),
		},
		"PersistentVolumeDescriber": &PersistentVolumeDescriber{
			fake.NewSimpleClientset(&api.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:     "bar",
					SelfLink: "url/url/url",
				},
			}, events),
		},
		"PodDescriber": &PodDescriber{
			fake.NewSimpleClientset(&api.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
					SelfLink:  "url/url/url",
				},
			}, events),
		},
		"ReplicaSetDescriber": &ReplicaSetDescriber{
			fake.NewSimpleClientset(&extensions.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		"ReplicationControllerDescriber": &ReplicationControllerDescriber{
			fake.NewSimpleClientset(&api.ReplicationController{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		"Service": &ServiceDescriber{
			fake.NewSimpleClientset(&api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		"StorageClass": &StorageClassDescriber{
			fake.NewSimpleClientset(&storage.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
			}, events),
		},
		"HorizontalPodAutoscaler": &HorizontalPodAutoscalerDescriber{
			fake.NewSimpleClientset(&autoscaling.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
		"ConfigMap": &ConfigMapDescriber{
			fake.NewSimpleClientset(&api.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			}, events),
		},
	}

	for name, d := range m {
		out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
		if err != nil {
			t.Errorf("unexpected error for %q: %v", name, err)
		}
		if !strings.Contains(out, "bar") {
			t.Errorf("unexpected out for %q: %s", name, out)
		}
		if !strings.Contains(out, "Events:") {
			t.Errorf("events not found for %q when ShowEvents=true: %s", name, out)
		}

		out, err = d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: false})
		if err != nil {
			t.Errorf("unexpected error for %q: %s", name, err)
		}
		if !strings.Contains(out, "bar") {
			t.Errorf("unexpected out for %q: %s", name, out)
		}
		if strings.Contains(out, "Events:") {
			t.Errorf("events found for %q when ShowEvents=false: %s", name, out)
		}
	}
}

func TestPrintLabelsMultiline(t *testing.T) {
	var maxLenAnnotationStr string = "MaxLenAnnotation=Multicast addressing can be used in the link layer (Layer 2 in the OSI model), such as Ethernet multicast, and at the internet layer (Layer 3 for OSI) for Internet Protocol Version 4 "
	testCases := []struct {
		annotations map[string]string
		expectPrint string
	}{
		{
			annotations: map[string]string{"col1": "asd", "COL2": "zxc"},
			expectPrint: "Annotations:\tCOL2=zxc\n\tcol1=asd\n",
		},
		{
			annotations: map[string]string{"MaxLenAnnotation": maxLenAnnotationStr[17:]},
			expectPrint: "Annotations:\t" + maxLenAnnotationStr + "\n",
		},
		{
			annotations: map[string]string{"MaxLenAnnotation": maxLenAnnotationStr[17:] + "1"},
			expectPrint: "Annotations:\t" + maxLenAnnotationStr + "...\n",
		},
		{
			annotations: map[string]string{},
			expectPrint: "Annotations:\t<none>\n",
		},
	}
	for i, testCase := range testCases {
		out := new(bytes.Buffer)
		writer := NewPrefixWriter(out)
		printAnnotationsMultiline(writer, "Annotations", testCase.annotations)
		output := out.String()
		if output != testCase.expectPrint {
			t.Errorf("Test case %d: expected to find %q in output: %q", i, testCase.expectPrint, output)
		}
	}
}

func TestDescribeUnstructuredContent(t *testing.T) {
	testCases := []struct {
		expected   string
		unexpected string
	}{
		{
			expected: `API Version:	v1
Dummy 2:	present
Items:
  Item Bool:	true
  Item Int:	42
Kind:	Test
Metadata:
  Creation Timestamp:	2017-04-01T00:00:00Z
  Name:	MyName
  Namespace:	MyNamespace
  Resource Version:	123
  UID:	00000000-0000-0000-0000-000000000001
Status:	ok
URL:	http://localhost
`,
		},
		{
			unexpected: "\nDummy 1:\tpresent\n",
		},
		{
			unexpected: "Dummy 1",
		},
		{
			unexpected: "Dummy 3",
		},
		{
			unexpected: "Dummy3",
		},
		{
			unexpected: "dummy3",
		},
		{
			unexpected: "dummy 3",
		},
	}
	out := new(bytes.Buffer)
	w := NewPrefixWriter(out)
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Test",
			"dummy1":     "present",
			"dummy2":     "present",
			"metadata": map[string]interface{}{
				"name":              "MyName",
				"namespace":         "MyNamespace",
				"creationTimestamp": "2017-04-01T00:00:00Z",
				"resourceVersion":   123,
				"uid":               "00000000-0000-0000-0000-000000000001",
				"dummy3":            "present",
			},
			"items": []interface{}{
				map[string]interface{}{
					"itemBool": true,
					"itemInt":  42,
				},
			},
			"url":    "http://localhost",
			"status": "ok",
		},
	}
	printUnstructuredContent(w, LEVEL_0, obj.UnstructuredContent(), "", ".dummy1", ".metadata.dummy3")
	output := out.String()

	for _, test := range testCases {
		if len(test.expected) > 0 {
			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected to find %q in: %q", test.expected, output)
			}
		}
		if len(test.unexpected) > 0 {
			if strings.Contains(output, test.unexpected) {
				t.Errorf("Didn't expect to find %q in: %q", test.unexpected, output)
			}
		}
	}
}

func TestDescribePodSecurityPolicy(t *testing.T) {
	expected := []string{
		"Name:\\s*mypsp",
		"Allow Privileged:\\s*false",
		"Default Add Capabilities:\\s*<none>",
		"Required Drop Capabilities:\\s*<none>",
		"Allowed Capabilities:\\s*<none>",
		"Allowed Volume Types:\\s*<none>",
		"Allow Host Network:\\s*false",
		"Allow Host Ports:\\s*<none>",
		"Allow Host PID:\\s*false",
		"Allow Host IPC:\\s*false",
		"Read Only Root Filesystem:\\s*false",
		"SELinux Context Strategy: RunAsAny",
		"User:\\s*<none>",
		"Role:\\s*<none>",
		"Type:\\s*<none>",
		"Level:\\s*<none>",
		"Run As User Strategy: RunAsAny",
		"FSGroup Strategy: RunAsAny",
		"Supplemental Groups Strategy: RunAsAny",
	}

	fake := fake.NewSimpleClientset(&policy.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mypsp",
		},
		Spec: policy.PodSecurityPolicySpec{
			SELinux: policy.SELinuxStrategyOptions{
				Rule: policy.SELinuxStrategyRunAsAny,
			},
			RunAsUser: policy.RunAsUserStrategyOptions{
				Rule: policy.RunAsUserStrategyRunAsAny,
			},
			FSGroup: policy.FSGroupStrategyOptions{
				Rule: policy.FSGroupStrategyRunAsAny,
			},
			SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
				Rule: policy.SupplementalGroupsStrategyRunAsAny,
			},
		},
	})

	c := &describeClient{T: t, Namespace: "", Interface: fake}
	d := PodSecurityPolicyDescriber{c}
	out, err := d.Describe("", "mypsp", printers.DescriberSettings{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, item := range expected {
		if matched, _ := regexp.MatchString(item, out); !matched {
			t.Errorf("Expected to find %q in: %q", item, out)
		}
	}
}

func TestDescribeResourceQuota(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Status: api.ResourceQuotaStatus{
			Hard: api.ResourceList{
				api.ResourceName(api.ResourceCPU):            resource.MustParse("1"),
				api.ResourceName(api.ResourceLimitsCPU):      resource.MustParse("2"),
				api.ResourceName(api.ResourceLimitsMemory):   resource.MustParse("2G"),
				api.ResourceName(api.ResourceMemory):         resource.MustParse("1G"),
				api.ResourceName(api.ResourceRequestsCPU):    resource.MustParse("1"),
				api.ResourceName(api.ResourceRequestsMemory): resource.MustParse("1G"),
			},
			Used: api.ResourceList{
				api.ResourceName(api.ResourceCPU):            resource.MustParse("0"),
				api.ResourceName(api.ResourceLimitsCPU):      resource.MustParse("0"),
				api.ResourceName(api.ResourceLimitsMemory):   resource.MustParse("0G"),
				api.ResourceName(api.ResourceMemory):         resource.MustParse("0G"),
				api.ResourceName(api.ResourceRequestsCPU):    resource.MustParse("0"),
				api.ResourceName(api.ResourceRequestsMemory): resource.MustParse("0G"),
			},
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := ResourceQuotaDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOut := []string{"bar", "foo", "limits.cpu", "2", "limits.memory", "2G", "requests.cpu", "1", "requests.memory", "1G"}
	for _, expected := range expectedOut {
		if !strings.Contains(out, expected) {
			t.Errorf("expected to find %q in output: %q", expected, out)
		}
	}
}

func TestDescribeNetworkPolicies(t *testing.T) {
	expectedTime, err := time.Parse("2006-01-02 15:04:05 Z0700 MST", "2017-06-04 21:45:56 -0700 PDT")
	if err != nil {
		t.Errorf("unable to parse time %q error: %s", "2017-06-04 21:45:56 -0700 PDT", err)
	}
	expectedOut := `Name:         network-policy-1
Namespace:    default
Created on:   2017-06-04 21:45:56 -0700 PDT
Labels:       <none>
Annotations:  <none>
Spec:
  PodSelector:     foo in (bar1,bar2),foo2 notin (bar1,bar2),id1=app1,id2=app2
  Allowing ingress traffic:
    To Port: 80/TCP
    To Port: 82/TCP
    From:
      NamespaceSelector: id=ns1,id2=ns2
      PodSelector: id=pod1,id2=pod2
    From:
      PodSelector: id=app2,id2=app3
    From:
      NamespaceSelector: id=app2,id2=app3
    From:
      NamespaceSelector: foo in (bar1,bar2),id=app2,id2=app3
    From:
      IPBlock:
        CIDR: 192.168.0.0/16
        Except: 192.168.3.0/24, 192.168.4.0/24
    ----------
    To Port: <any> (traffic allowed to all ports)
    From: <any> (traffic not restricted by source)
  Allowing egress traffic:
    To Port: 80/TCP
    To Port: 82/TCP
    To:
      NamespaceSelector: id=ns1,id2=ns2
      PodSelector: id=pod1,id2=pod2
    To:
      PodSelector: id=app2,id2=app3
    To:
      NamespaceSelector: id=app2,id2=app3
    To:
      NamespaceSelector: foo in (bar1,bar2),id=app2,id2=app3
    To:
      IPBlock:
        CIDR: 192.168.0.0/16
        Except: 192.168.3.0/24, 192.168.4.0/24
    ----------
    To Port: <any> (traffic allowed to all ports)
    To: <any> (traffic not restricted by source)
  Policy Types: Ingress, Egress
`

	port80 := intstr.FromInt(80)
	port82 := intstr.FromInt(82)
	protoTCP := api.ProtocolTCP

	versionedFake := fake.NewSimpleClientset(&networking.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "network-policy-1",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(expectedTime),
		},
		Spec: networking.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"id1": "app1",
					"id2": "app2",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "foo", Operator: "In", Values: []string{"bar1", "bar2"}},
					{Key: "foo2", Operator: "NotIn", Values: []string{"bar1", "bar2"}},
				},
			},
			Ingress: []networking.NetworkPolicyIngressRule{
				{
					Ports: []networking.NetworkPolicyPort{
						{Port: &port80},
						{Port: &port82, Protocol: &protoTCP},
					},
					From: []networking.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "pod1",
									"id2": "pod2",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "ns1",
									"id2": "ns2",
								},
							},
						},
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{Key: "foo", Operator: "In", Values: []string{"bar1", "bar2"}},
								},
							},
						},
						{
							IPBlock: &networking.IPBlock{
								CIDR:   "192.168.0.0/16",
								Except: []string{"192.168.3.0/24", "192.168.4.0/24"},
							},
						},
					},
				},
				{},
			},
			Egress: []networking.NetworkPolicyEgressRule{
				{
					Ports: []networking.NetworkPolicyPort{
						{Port: &port80},
						{Port: &port82, Protocol: &protoTCP},
					},
					To: []networking.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "pod1",
									"id2": "pod2",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "ns1",
									"id2": "ns2",
								},
							},
						},
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"id":  "app2",
									"id2": "app3",
								},
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{Key: "foo", Operator: "In", Values: []string{"bar1", "bar2"}},
								},
							},
						},
						{
							IPBlock: &networking.IPBlock{
								CIDR:   "192.168.0.0/16",
								Except: []string{"192.168.3.0/24", "192.168.4.0/24"},
							},
						},
					},
				},
				{},
			},
			PolicyTypes: []networking.PolicyType{networking.PolicyTypeIngress, networking.PolicyTypeEgress},
		},
	})
	d := NetworkPolicyDescriber{versionedFake}
	out, err := d.Describe("", "network-policy-1", printers.DescriberSettings{})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if out != expectedOut {
		t.Errorf("want:\n%s\ngot:\n%s", expectedOut, out)
	}
}

func TestDescribeServiceAccount(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Secrets: []api.ObjectReference{
			{
				Name: "test-objectref",
			},
		},
		ImagePullSecrets: []api.LocalObjectReference{
			{
				Name: "test-local-ref",
			},
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := ServiceAccountDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOut := `Name:                bar
Namespace:           foo
Labels:              <none>
Annotations:         <none>
Image pull secrets:  test-local-ref (not found)
Mountable secrets:   test-objectref (not found)
Tokens:              <none>
Events:              <none>` + "\n"
	if out != expectedOut {
		t.Errorf("expected : %q\n but got output:\n %q", expectedOut, out)
	}

}

func TestDescribeNode(t *testing.T) {
	fake := fake.NewSimpleClientset(&api.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
		Spec: api.NodeSpec{
			Unschedulable: true,
		},
	})
	c := &describeClient{T: t, Namespace: "foo", Interface: fake}
	d := NodeDescriber{c}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedOut := []string{"Unschedulable", "true"}
	for _, expected := range expectedOut {
		if !strings.Contains(out, expected) {
			t.Errorf("expected to find %q in output: %q", expected, out)
		}
	}

}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	o := b
	return &o
}

func TestControllerRef(t *testing.T) {
	f := fake.NewSimpleClientset(
		&api.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "foo",
				UID:       "123456",
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "ReplicationController",
			},
			Spec: api.ReplicationControllerSpec{
				Replicas: 1,
				Selector: map[string]string{"abc": "xyz"},
				Template: &api.PodTemplateSpec{
					Spec: api.PodSpec{
						Containers: []api.Container{
							{Image: "mytest-image:latest"},
						},
					},
				},
			},
		},
		&api.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "barpod",
				Namespace:       "foo",
				Labels:          map[string]string{"abc": "xyz"},
				OwnerReferences: []metav1.OwnerReference{{Name: "bar", UID: "123456", Controller: boolPtr(true)}},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Image: "mytest-image:latest"},
				},
			},
			Status: api.PodStatus{
				Phase: api.PodRunning,
			},
		},
		&api.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "orphan",
				Namespace: "foo",
				Labels:    map[string]string{"abc": "xyz"},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Image: "mytest-image:latest"},
				},
			},
			Status: api.PodStatus{
				Phase: api.PodRunning,
			},
		},
		&api.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "buzpod",
				Namespace:       "foo",
				Labels:          map[string]string{"abc": "xyz"},
				OwnerReferences: []metav1.OwnerReference{{Name: "buz", UID: "654321", Controller: boolPtr(true)}},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Image: "mytest-image:latest"},
				},
			},
			Status: api.PodStatus{
				Phase: api.PodRunning,
			},
		})
	d := ReplicationControllerDescriber{f}
	out, err := d.Describe("foo", "bar", printers.DescriberSettings{ShowEvents: false})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1 Running") {
		t.Errorf("unexpected out: %s", out)
	}
}
