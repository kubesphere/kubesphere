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
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const retryTimes = 3

type JobRunner interface {
	JobReRun(namespace, name, resourceVersion string) error
}

type jobRunner struct {
	client runtimeclient.Client
}

func NewJobRunner(client runtimeclient.Client) JobRunner {
	return &jobRunner{client: client}
}

func (r *jobRunner) JobReRun(namespace, jobName, resourceVersion string) error {
	job := &batchv1.Job{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: jobName}, job); err != nil {
		return err
	}
	// do not rerun job if resourceVersion not match
	if job.GetObjectMeta().GetResourceVersion() != resourceVersion {
		err := k8serr.NewConflict(schema.GroupResource{
			Group: job.GetObjectKind().GroupVersionKind().Group, Resource: "job",
		}, jobName, fmt.Errorf("please apply your changes to the latest version and try again"))
		klog.Warning(err)
		return err
	}

	newJob := *job
	newJob.ResourceVersion = ""
	newJob.Status = batchv1.JobStatus{}
	newJob.ObjectMeta.UID = ""
	newJob.Annotations["revisions"] = strings.Replace(job.Annotations["revisions"], "running", "unfinished", -1)

	delete(newJob.Spec.Selector.MatchLabels, "controller-uid")
	delete(newJob.Spec.Template.ObjectMeta.Labels, "controller-uid")

	if err := r.deleteJob(namespace, jobName); err != nil {
		klog.Errorf("failed to rerun job %s, reason: %s", jobName, err)
		return fmt.Errorf("failed to rerun job %s", jobName)
	}

	var err error
	for i := 0; i < retryTimes; i++ {
		if err = r.client.Create(context.Background(), &newJob); err != nil {
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

func (r *jobRunner) deleteJob(namespace, job string) error {
	return r.client.Delete(context.Background(),
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: job}})
}
