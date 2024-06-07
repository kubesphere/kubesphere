/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package lease

import (
	"context"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"kubesphere.io/kubesphere/pkg/constants"
)

type Operator struct {
	client kubernetes.Interface
}

func NewOperator(client kubernetes.Interface) *Operator {
	return &Operator{
		client: client,
	}
}

func (o *Operator) Create(ctx context.Context, owner *corev1.Pod) error {
	now := metav1.NowMicro()
	lease := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: owner.Namespace,
			Name:      owner.Name,
			Labels: map[string]string{
				constants.KubectlPodLabel:        "",
				constants.KubeSphereManagedLabel: "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Pod",
					UID:        owner.GetUID(),
					Name:       owner.GetName(),
				},
			},
		},
		Spec: coordinationv1.LeaseSpec{
			AcquireTime: &now,
			RenewTime:   &now,
		},
	}
	if _, err := o.client.CoordinationV1().Leases(owner.Namespace).Create(ctx, lease, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			return o.Renew(ctx, lease.Namespace, lease.Name)
		}
		return err
	}
	return nil
}

func (o *Operator) Renew(ctx context.Context, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		lease, err := o.client.CoordinationV1().Leases(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		now := metav1.NowMicro()
		lease.Spec.RenewTime = &now
		if _, err = o.client.CoordinationV1().Leases(namespace).Update(ctx, lease, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
}
