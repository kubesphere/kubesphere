/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
)

const webhookName = "cluster-webhook"

func (v *Webhook) Name() string {
	return webhookName
}

func (v *Webhook) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

var _ kscontroller.Controller = &Webhook{}
var _ admission.CustomValidator = &Webhook{}

type Webhook struct {
}

func (v *Webhook) SetupWithManager(mgr *kscontroller.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(&clusterv1alpha1.Cluster{}).
		WithValidator(v).
		Complete()
}

func (v *Webhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (v *Webhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldCluster, ok := oldObj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil, fmt.Errorf("expected a Cluster but got a %T", oldObj)
	}
	newCluster, ok := newObj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil, fmt.Errorf("expected a Cluster but got a %T", newObj)
	}

	// The cluster created for the first time has no status information
	if oldCluster.Status.UID == "" {
		return nil, nil
	}

	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig(newCluster.Spec.Connection.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load cluster config for %s: %s", newCluster.Name, err)
	}
	clusterClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}
	kubeSystem, err := clusterClient.CoreV1().Namespaces().Get(ctx, metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if oldCluster.Status.UID != kubeSystem.UID {
		return nil, errors.New("this kubeconfig corresponds to a different cluster than the previous one, you need to make sure that kubeconfig is not from another cluster")
	}
	return nil, nil
}

func (v *Webhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
