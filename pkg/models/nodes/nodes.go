/*

 Copyright 2019 The KubeSphere Authors.

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
package nodes

import (
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"math"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func DrainNode(nodeName string) (err error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()

	node, err := k8sClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if node.Spec.Unschedulable {
		return fmt.Errorf("node %s have been drained", nodeName)
	}

	data := []byte(" {\"spec\":{\"unschedulable\":true}}")
	_, err = k8sClient.CoreV1().Nodes().Patch(nodeName, types.StrategicMergePatchType, data)
	if err != nil {
		return err
	}

	doneCh := make(chan bool)
	errCh := make(chan error)
	go drainEviction(nodeName, doneCh, errCh)
	for {
		select {
		case err := <-errCh:
			return err
		case <-doneCh:
			return nil
		}
	}

}

func getNodePods(client kubernetes.Interface, node v1.Node) (*v1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name +
		",status.phase!=" + string(v1.PodSucceeded) +
		",status.phase!=" + string(v1.PodFailed))

	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
}

func drainEviction(nodeName string, doneCh chan bool, errCh chan error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()

	node, err := k8sClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
		errCh <- err
	}
	podList, err := getNodePods(k8sClient, *node)
	if err != nil {
		klog.Fatal(err)
		errCh <- err
	}

	daemonSetList, err := k8sClient.AppsV1().DaemonSets("").List(metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
		errCh <- err

	}

	pods := make([]v1.Pod, 0)
	// remove mirror pod static pod
	if len(podList.Items) > 0 {
		for _, pod := range podList.Items {
			if !containDaemonset(pod, *daemonSetList) {
				//static or mirror pod
				if isStaticPod(&pod) || isMirrorPod(&pod) {
					continue
				} else {
					pods = append(pods, pod)
				}
			}
		}
	}
	if len(pods) == 0 {
		doneCh <- true
	} else {
		//create eviction
		getPodFn := func(namespace, name string) (*v1.Pod, error) {
			return k8sClient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		}
		evicerr := evictPods(pods, 0, getPodFn)
		if evicerr == nil {
			doneCh <- true
		} else {
			klog.Fatal(evicerr)
			errCh <- err
		}
	}
}

func getPodSource(pod *v1.Pod) (string, error) {
	if pod.Annotations != nil {
		if source, ok := pod.Annotations["kubernetes.io/config.source"]; ok {
			return source, nil
		}
	}
	return "", fmt.Errorf("cannot get source of pod %q", pod.UID)
}

func isStaticPod(pod *v1.Pod) bool {
	source, err := getPodSource(pod)
	return err == nil && source != "api"
}

func isMirrorPod(pod *v1.Pod) bool {
	_, ok := pod.Annotations[v1.MirrorPodAnnotationKey]
	return ok
}

func containDaemonset(pod v1.Pod, daemonsetList appsv1.DaemonSetList) bool {
	flag := false
	for _, daemonset := range daemonsetList.Items {
		if strings.Contains(pod.Name, daemonset.Name) {
			flag = true
		}
	}
	return flag
}

func evictPod(pod v1.Pod, GracePeriodSeconds int) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	deleteOptions := &metav1.DeleteOptions{}
	if GracePeriodSeconds >= 0 {
		gracePeriodSeconds := int64(GracePeriodSeconds)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	var eviction policy.Eviction
	eviction.Kind = "Eviction"
	eviction.APIVersion = "policy/v1beta1"
	eviction.Namespace = pod.Namespace
	eviction.Name = pod.Name
	eviction.DeleteOptions = deleteOptions
	err := k8sClient.CoreV1().Pods(pod.Namespace).Evict(&eviction)
	if err != nil {
		return err
	}

	return nil
}

func evictPods(pods []v1.Pod, GracePeriodSeconds int, getPodFn func(namespace, name string) (*v1.Pod, error)) error {
	doneCh := make(chan bool, len(pods))
	errCh := make(chan error, 1)

	for _, pod := range pods {
		go func(pod v1.Pod, doneCh chan bool, errCh chan error) {
			var err error
			var count int
			for {
				err = evictPod(pod, GracePeriodSeconds)
				if err == nil {
					count++
					if count > 2 {
						break
					} else {
						continue
					}
				} else if apierrors.IsNotFound(err) {
					count = 0
					doneCh <- true
					klog.Info(fmt.Sprintf("pod %s evict", pod.Name))
					return
				} else if apierrors.IsTooManyRequests(err) {
					count = 0
					time.Sleep(5 * time.Second)
				} else {
					count = 0
					errCh <- fmt.Errorf("error when evicting pod %q: %v", pod.Name, err)
					return
				}
			}

			podArray := []v1.Pod{pod}
			_, err = waitForDelete(podArray, time.Second, time.Duration(math.MaxInt64), getPodFn)
			if err == nil {
				doneCh <- true
				klog.Info(fmt.Sprintf("pod %s delete", pod.Name))
			} else {
				errCh <- fmt.Errorf("error when waiting for pod %q terminating: %v", pod.Name, err)
			}
		}(pod, doneCh, errCh)
	}

	Timeout := 300 * power(10, 9)
	doneCount := 0
	// 0 timeout means infinite, we use MaxInt64 to represent it.
	var globalTimeout time.Duration
	if Timeout == 0 {
		globalTimeout = time.Duration(math.MaxInt64)
	} else {
		globalTimeout = time.Duration(Timeout)
	}
	for {
		select {
		case err := <-errCh:
			return err
		case <-doneCh:
			doneCount++
			if doneCount == len(pods) {
				return nil
			}
		case <-time.After(globalTimeout):
			return fmt.Errorf("drain did not complete within %v", globalTimeout)
		}
	}
}

func waitForDelete(pods []v1.Pod, interval, timeout time.Duration, getPodFn func(string, string) (*v1.Pod, error)) ([]v1.Pod, error) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		var pendingPods []v1.Pod
		for i, pod := range pods {
			p, err := getPodFn(pod.Namespace, pod.Name)
			if apierrors.IsNotFound(err) || (p != nil && p.ObjectMeta.UID != pod.ObjectMeta.UID) {
				continue
			} else if err != nil {
				return false, err
			} else {
				pendingPods = append(pendingPods, pods[i])
			}
		}
		pods = pendingPods
		if len(pendingPods) > 0 {
			return false, nil
		}
		return true, nil
	})
	return pods, err
}

func power(x int64, n int) int64 {
	return int64(math.Pow(float64(x), float64(n)))
}
