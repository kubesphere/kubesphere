/*
Copyright 2018 The KubeSphere Authors.

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

package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"kubesphere.io/kubesphere/pkg/client"
)

var k8sClient *kubernetes.Clientset

func (ctl *JobCtl) generateObject(item v1.Job) *Job {
	var status, displayName string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	namespace := item.Namespace
	succeedPodNum := item.Status.Succeeded
	desirePodNum := *item.Spec.Completions
	createTime := item.CreationTimestamp.Time
	updteTime := createTime
	for _, condition := range item.Status.Conditions {
		if condition.Type == "Failed" && condition.Status == "True" {
			status = Failed
		}

		if condition.Type == "Complete" && condition.Status == "True" {
			status = Completed
		}

		if updteTime.Before(condition.LastProbeTime.Time) {
			updteTime = condition.LastProbeTime.Time
		}

		if updteTime.Before(condition.LastTransitionTime.Time) {
			updteTime = condition.LastTransitionTime.Time
		}
	}

	if desirePodNum > succeedPodNum && len(status) == 0 {
		status = Running
	}

	object := &Job{
		Namespace:   namespace,
		Name:        name,
		DisplayName: displayName,
		Desire:      desirePodNum,
		Completed:   succeedPodNum,
		UpdateTime:  updteTime,
		CreateTime:  createTime,
		Status:      status,
		Annotation:  MapString{item.Annotations},
		Labels:      MapString{item.Labels},
	}

	return object
}

func (ctl *JobCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *JobCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Job{}) {
		db.DropTable(&Job{})
	}

	db = db.CreateTable(&Job{})

	ctl.initListerAndInformer()
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)
	}

	ctl.informer.Run(stopChan)
}

func (ctl *JobCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *JobCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)
	ctl.lister = informerFactory.Batch().V1().Jobs().Lister()

	informer := informerFactory.Batch().V1().Jobs().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.Job)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Job)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item Job
			object := obj.(*v1.Job)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	ctl.informer = informer
}

func (ctl *JobCtl) CountWithConditions(conditions string) int {
	var object Job

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *JobCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Job
	var object Job
	var total int

	if len(order) == 0 {
		order = "updateTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *JobCtl) Lister() interface{} {

	return ctl.lister
}

func getRevisions(job v1.Job) (JobRevisions, error) {
	revisions := make(JobRevisions)

	if _, exist := job.Annotations["revisions"]; exist {
		revisionsStr := job.Annotations["revisions"]

		err := json.Unmarshal([]byte(revisionsStr), &revisions)
		if err != nil {
			glog.Errorf("failed to rerun job %s, reason: %s", err, err)
			return nil, fmt.Errorf("failed to rerun job %s", job.Name)
		}
	}

	return revisions, nil
}

func getStatus(item *v1.Job) JobStatus {
	var status JobStatus
	for _, condition := range item.Status.Conditions {
		if condition.Type == "Failed" && condition.Status == "True" {
			status.Status = Failed
			status.Reasons = append(status.Reasons, condition.Reason)
			status.Messages = append(status.Messages, condition.Message)
		}

		if condition.Type == "Complete" && condition.Status == "True" {
			status.Status = Completed
		}
	}

	if len(status.Status) == 0 {
		status.Status = Unfinished
	}

	status.DesirePodNum = *item.Spec.Completions
	status.Succeed = item.Status.Succeeded
	status.Failed = item.Status.Failed
	status.StartTime = item.Status.StartTime.Time
	if item.Status.CompletionTime != nil {
		status.CompletionTime = item.Status.CompletionTime.Time
	}

	return status
}

func deleteJob(namespace, job string) error {
	deletePolicy := metav1.DeletePropagationBackground
	err := k8sClient.BatchV1().Jobs(namespace).Delete(job, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	return err
}

func JobReRun(namespace, jobName string) (string, error) {
	k8sClient = client.NewK8sClient()
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	newJob := *job
	newJob.ResourceVersion = ""
	newJob.Status = v1.JobStatus{}
	newJob.ObjectMeta.UID = ""
	delete(newJob.Spec.Selector.MatchLabels, "controller-uid")
	delete(newJob.Spec.Template.ObjectMeta.Labels, "controller-uid")

	revisions, err := getRevisions(*job)

	if err != nil {
		return "", err
	}

	index := len(revisions) + 1
	value := getStatus(job)
	revisions[index] = value

	revisionsByte, err := json.Marshal(revisions)
	if err != nil {
		glog.Errorf("failed to rerun job %s, reason: %s", err, err)
		return "", fmt.Errorf("failed to rerun job %s", jobName)
	}

	newJob.Annotations["revisions"] = string(revisionsByte)

	err = deleteJob(job.Namespace, job.Name)
	if err != nil {
		glog.Errorf("failed to rerun job %s, reason: %s", err, err)
		return "", fmt.Errorf("failed to rerun job %s", jobName)
	}

	_, err = k8sClient.BatchV1().Jobs(namespace).Create(&newJob)
	if err != nil {
		glog.Errorf("failed to rerun job %s, reason: %s", err, err)
		return "", fmt.Errorf("failed to rerun job %s", jobName)
	}

	return "succeed", nil
}
