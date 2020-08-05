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

package e2e_test

import (
	"context"
	"time"

	"k8s.io/klog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var simpleDeployYaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  name:  nginx
  namespace: production
  labels:
    name:  nginx
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
       app: nginx
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        name:  nginx
        app: nginx
        color : red
    spec:
      containers:
      - image:  nginx:alpine
        name:  nginx
        imagePullPolicy: IfNotPresent
        resources:
          requests:
            cpu: "20m"
            memory: "55M"
        env:
        - name:  ENVVARNAME
          value:  ENVVARVALUE       
        ports:
        - containerPort:  80
          name:  http
      restartPolicy: Always`

var simpleNPYaml = `apiVersion: network.kubesphere.io/v1alpha1
kind: NamespaceNetworkPolicy
metadata:
  name: allow-icmp-only
  namespace: production
spec:
  selector: color == 'red'
  ingress:
  - action: Allow
    protocol: ICMP
    source:
      selector: color == 'blue'
      namespaceSelector: all()`

var simpleJobYaml = `apiVersion: batch/v1
kind: Job
metadata:
  name: test-connect
  namespace: production
spec:
  template:
    metadata:
      labels:
        color : blue
    spec:
      containers:
      - name: test-connect
        image: alpine
        command: ["ping", "1.1.1.1"]
      restartPolicy: Never
  backoffLimit: 1`

var testNs = "production"
var _ = Describe("E2e for network policy", func() {
	BeforeEach(func() {
		Expect(test.EnsureNamespace(ctx.Client, testNs)).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(test.DeleteNamespace(ctx.Client, testNs)).ShouldNot(HaveOccurred())
		ns := &corev1.Namespace{}
		ns.Name = testNs
		Expect(test.WaitForDeletion(ctx.Client, ns, time.Second*5, time.Minute)).ShouldNot(HaveOccurred())
	})

	It("Should work well in simple namespaceNetworkPolicy", func() {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(simpleDeployYaml), nil, nil)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to parse yaml")
		deploy := obj.(*appsv1.Deployment)
		Expect(ctx.Client.Create(context.TODO(), obj)).ShouldNot(HaveOccurred())
		Expect(test.WaitForController(ctx.Client, deploy.Namespace, deploy.Name, *deploy.Spec.Replicas, time.Second*2, time.Minute)).ShouldNot(HaveOccurred())
		defer func() {
			Expect(ctx.Client.Delete(context.TODO(), deploy)).ShouldNot(HaveOccurred())
		}()
		obj, _, err = decode([]byte(simpleNPYaml), nil, nil)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to parse networkpolicy yaml")
		np := obj.(*v1alpha1.NamespaceNetworkPolicy)
		Expect(ctx.Client.Create(context.TODO(), np)).ShouldNot(HaveOccurred())
		defer func() {
			Expect(ctx.Client.Delete(context.TODO(), np)).ShouldNot(HaveOccurred())
			Expect(test.WaitForDeletion(ctx.Client, np, time.Second*2, time.Minute)).ShouldNot(HaveOccurred())
		}()
		obj, _, err = decode([]byte(simpleJobYaml), nil, nil)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to parse job yaml")

		//create a job to test
		job := obj.(*batchv1.Job)
		selector, _ := labels.Parse("app=nginx")
		podlist := &corev1.PodList{}
		Expect(ctx.Client.List(context.TODO(), &client.ListOptions{
			Namespace:     deploy.Namespace,
			LabelSelector: selector,
		}, podlist)).ShouldNot(HaveOccurred())
		Expect(podlist.Items).To(HaveLen(int(*deploy.Spec.Replicas)))
		podip := podlist.Items[0].Status.PodIP
		job.Spec.Template.Spec.Containers[0].Command = []string{"ping", "-c", "4", podip}
		job.Spec.Template.Labels["color"] = "yellow"
		orginalJob := job.DeepCopy()
		Expect(ctx.Client.Create(context.TODO(), job)).ShouldNot(HaveOccurred())
		defer func() {
			Expect(ctx.Client.Delete(context.TODO(), job)).ShouldNot(HaveOccurred())
		}()
		klog.Infoln("sleep 10s to wait for controller creating np")
		time.Sleep(time.Second * 10)
		Expect(test.WaitForJobFail(ctx.Client, job.Namespace, job.Name, time.Second*3, time.Minute)).ShouldNot(HaveOccurred(), "Failed to block connection")

		//change job color
		job = orginalJob.DeepCopy()
		Expect(ctx.Client.Delete(context.TODO(), job)).ShouldNot(HaveOccurred())
		Expect(test.WaitForDeletion(ctx.Client, job, time.Second*2, time.Minute)).ShouldNot(HaveOccurred())
		job.Spec.Template.Labels["color"] = "blue"
		Expect(ctx.Client.Create(context.TODO(), job)).ShouldNot(HaveOccurred())
		Expect(test.WaitForJobSucceed(ctx.Client, job.Namespace, job.Name, time.Second*3, time.Minute)).ShouldNot(HaveOccurred(), "Connection failed")
	})
})
