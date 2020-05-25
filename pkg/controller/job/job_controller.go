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

package job

import (
	"encoding/json"
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchv1informers "k8s.io/client-go/informers/batch/v1"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	log "k8s.io/klog"
	"time"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries             = 15
	revisionsAnnotationKey = "revisions"
)

type JobController struct {
	client           clientset.Interface
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	jobLister batchv1listers.JobLister
	jobSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewJobController(jobInformer batchv1informers.JobInformer, client clientset.Interface) *JobController {
	v := &JobController{
		client:           client,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "job"),
		workerLoopPeriod: time.Second,
	}

	v.jobLister = jobInformer.Lister()
	v.jobSynced = jobInformer.Informer().HasSynced

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			v.enqueueJob(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			v.enqueueJob(cur)
		},
	})

	return v

}

func (v *JobController) Start(stopCh <-chan struct{}) error {
	return v.Run(5, stopCh)
}

func (v *JobController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	log.Info("starting job controller")
	defer log.Info("shutting down job controller")

	if !cache.WaitForCacheSync(stopCh, v.jobSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(v.worker, v.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

func (v *JobController) enqueueJob(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	v.queue.Add(key)
}

func (v *JobController) worker() {
	for v.processNextWorkItem() {

	}
}

func (v *JobController) processNextWorkItem() bool {
	eKey, quit := v.queue.Get()
	if quit {
		return false
	}

	defer v.queue.Done(eKey)

	err := v.syncJob(eKey.(string))
	v.handleErr(err, eKey)

	return true
}

// main function of the reconcile for job
// job's name is same with the service that created it
func (v *JobController) syncJob(key string) error {
	startTime := time.Now()
	defer func() {
		log.V(4).Info("Finished syncing job.", "key", key, "duration", time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	job, err := v.jobLister.Jobs(namespace).Get(name)
	if err != nil {
		// has been deleted
		if errors.IsNotFound(err) {
			return nil
		}
		log.Error(err, "get job failed", "namespace", namespace, "name", name)
		return err
	}

	err = v.makeRevision(job)

	if err != nil {
		log.Error(err, "make job revision failed", "namespace", namespace, "name", name)
		return err
	}

	return nil
}

// When a job is added, figure out which service it will be used
// and enqueue it. obj must have *batchv1.Job type
func (v *JobController) addJob(obj interface{}) {
	deploy := obj.(*batchv1.Job)

	v.queue.Add(deploy.Name)

	return
}

func (v *JobController) handleErr(err error, key interface{}) {
	if err == nil {
		v.queue.Forget(key)
		return
	}

	if v.queue.NumRequeues(key) < maxRetries {
		log.V(2).Info("Error syncing job, retrying.", "key", key, "error", err)
		v.queue.AddRateLimited(key)
		return
	}

	log.V(4).Info("Dropping job out of the queue", "key", key, "error", err)
	v.queue.Forget(key)
	utilruntime.HandleError(err)
}

func (v *JobController) makeRevision(job *batchv1.Job) error {
	revisionIndex := -1
	revisions, err := v.getRevisions(job)

	// failed get revisions
	if err != nil {
		return nil
	}

	uid := job.UID
	for index, revision := range revisions {
		if revision.Uid == string(uid) {
			currentRevision := v.getCurrentRevision(job)
			if reflect.DeepEqual(currentRevision, revision) {
				return nil
			} else {
				revisionIndex = index
				break
			}
		}
	}

	if revisionIndex == -1 {
		revisionIndex = len(revisions) + 1
	}

	revisions[revisionIndex] = v.getCurrentRevision(job)

	revisionsByte, err := json.Marshal(revisions)
	if err != nil {
		log.Error("generate reversion string failed", err)
		return nil
	}

	if job.Annotations == nil {
		job.Annotations = make(map[string]string)
	}

	job.Annotations[revisionsAnnotationKey] = string(revisionsByte)
	_, err = v.client.BatchV1().Jobs(job.Namespace).Update(job)

	if err != nil {
		return err
	}
	return nil
}

func (v *JobController) getRevisions(job *batchv1.Job) (JobRevisions, error) {
	revisions := make(JobRevisions)

	if revisionsStr := job.Annotations[revisionsAnnotationKey]; revisionsStr != "" {
		err := json.Unmarshal([]byte(revisionsStr), &revisions)
		if err != nil {
			return nil, fmt.Errorf("failed to get job %s's revisions, reason: %s", job.Name, err)
		}
	}

	return revisions, nil
}

func (v *JobController) getCurrentRevision(item *batchv1.Job) JobRevision {
	var revision JobRevision
	for _, condition := range item.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == v1.ConditionTrue {
			revision.Status = Failed
			revision.Reasons = append(revision.Reasons, condition.Reason)
			revision.Messages = append(revision.Messages, condition.Message)
		} else if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
			revision.Status = Completed
		}
	}

	if len(revision.Status) == 0 {
		revision.Status = Running
	}

	if item.Spec.Completions != nil {
		revision.DesirePodNum = *item.Spec.Completions
	}

	revision.Succeed = item.Status.Succeeded
	revision.Failed = item.Status.Failed
	revision.StartTime = item.CreationTimestamp.Time
	revision.Uid = string(item.UID)
	if item.Status.CompletionTime != nil {
		revision.CompletionTime = item.Status.CompletionTime.Time
	}

	return revision
}
