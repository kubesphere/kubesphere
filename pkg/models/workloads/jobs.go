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
package workloads

import (
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"strings"
	"time"

	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const retryTimes = 3

func JobReRun(namespace, jobName string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newJob := *job
	newJob.ResourceVersion = ""
	newJob.Status = v1.JobStatus{}
	newJob.ObjectMeta.UID = ""
	newJob.Annotations["revisions"] = strings.Replace(job.Annotations["revisions"], "running", "unfinished", -1)

	delete(newJob.Spec.Selector.MatchLabels, "controller-uid")
	delete(newJob.Spec.Template.ObjectMeta.Labels, "controller-uid")

	err = deleteJob(namespace, jobName)

	if err != nil {
		klog.Errorf("failed to rerun job %s, reason: %s", jobName, err)
		return fmt.Errorf("failed to rerun job %s", jobName)
	}

	for i := 0; i < retryTimes; i++ {
		_, err = k8sClient.BatchV1().Jobs(namespace).Create(&newJob)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if err != nil {
		klog.Errorf("failed to rerun job %s, reason: %s", jobName, err)
		return fmt.Errorf("failed to rerun job %s", jobName)
	}

	return nil
}

func deleteJob(namespace, job string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	deletePolicy := metav1.DeletePropagationBackground
	err := k8sClient.BatchV1().Jobs(namespace).Delete(job, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	return err
}
