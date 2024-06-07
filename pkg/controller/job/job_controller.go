/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	revisionsAnnotationKey = "revisions"
	controllerName         = "job-revision"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	client.Client
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return builder.
		ControllerManagedBy(mgr).
		For(
			&batchv1.Job{},
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	defer func() {
		klog.V(4).Info("Finished syncing job.", "key", req.String(), "duration", time.Since(startTime))
	}()

	job := &batchv1.Job{}
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.makeRevision(ctx, job); err != nil {
		klog.Error(err, "make job revision failed", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) makeRevision(ctx context.Context, job *batchv1.Job) error {
	revisionIndex := -1
	revisions, err := r.getRevisions(job)
	// failed get revisions
	if err != nil {
		return nil
	}

	uid := job.UID
	for index, revision := range revisions {
		if revision.Uid == string(uid) {
			currentRevision := r.getCurrentRevision(job)
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

	revisions[revisionIndex] = r.getCurrentRevision(job)

	revisionsByte, err := json.Marshal(revisions)
	if err != nil {
		klog.Error("generate reversion string failed", err)
		return nil
	}

	if job.Annotations == nil {
		job.Annotations = make(map[string]string)
	}
	job.Annotations[revisionsAnnotationKey] = string(revisionsByte)
	return r.Update(ctx, job)
}

func (r *Reconciler) getRevisions(job *batchv1.Job) (JobRevisions, error) {
	revisions := make(JobRevisions)

	if revisionsStr := job.Annotations[revisionsAnnotationKey]; revisionsStr != "" {
		err := json.Unmarshal([]byte(revisionsStr), &revisions)
		if err != nil {
			return nil, fmt.Errorf("failed to get job %s's revisions, reason: %s", job.Name, err)
		}
	}

	return revisions, nil
}

func (r *Reconciler) getCurrentRevision(item *batchv1.Job) JobRevision {
	var revision JobRevision
	for _, condition := range item.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			revision.Status = Failed
			revision.Reasons = append(revision.Reasons, condition.Reason)
			revision.Messages = append(revision.Messages, condition.Message)
		} else if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
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
