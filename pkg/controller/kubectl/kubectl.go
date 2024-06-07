/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package kubectl

import (
	"context"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller"
)

const controllerName = "kubectl"

type Reconciler struct {
	client.Client

	resyncPeriod time.Duration
	renewPeriod  time.Duration
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) NeedLeaderElection() bool {
	return true
}

func (r *Reconciler) SetupWithManager(mgr *controller.Manager) error {
	r.Client = mgr.GetClient()
	r.resyncPeriod = time.Minute
	r.renewPeriod = time.Minute

	return mgr.Add(r)
}

func (r *Reconciler) Start(ctx context.Context) error {
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := r.reconcile(ctx); err != nil {
			klog.Errorf("%s controller reconcile error: %s\n", controllerName, err.Error())
		}
	}, r.resyncPeriod)
	return nil
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	leases := &coordinationv1.LeaseList{}
	if err := r.List(ctx, leases, client.MatchingLabels{constants.KubectlPodLabel: ""}); err != nil {
		return err
	}
	// The minimum required heartbeat time, the heartbeat time of all leases must be greater than this
	heartbeatTime := time.Now().Add(-r.renewPeriod)
	for i := range leases.Items {
		lease := &leases.Items[i]
		if lease.Spec.RenewTime.After(heartbeatTime) {
			continue
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: lease.Namespace,
				Name:      lease.Name,
			},
		}
		if err := r.Delete(ctx, pod, client.GracePeriodSeconds(0)); err != nil && !errors.IsNotFound(err) {
			klog.Errorf("deleting Pod %s/%s failed: %s, will retry", pod.Namespace, pod.Name, err.Error())
		}
	}
	return nil
}
