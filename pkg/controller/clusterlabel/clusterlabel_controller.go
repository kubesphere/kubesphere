/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package clusterlabel

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

// Reconciler is a reconciler for the Label object.
type Reconciler struct {
	client.Client
}

func (r *Reconciler) Name() string {
	return "clusterlabel"
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

// Reconcile reconciles the Label object, sync label to the individual Cluster CRs.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	label := &clusterv1alpha1.Label{}
	if err := r.Get(ctx, req.NamespacedName, label); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if label.DeletionTimestamp != nil {
		return ctrl.Result{}, r.deleteLabel(ctx, label)
	}

	if len(label.Finalizers) == 0 {
		label.Finalizers = []string{clusterv1alpha1.LabelFinalizer}
		if err := r.Update(ctx, label); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, r.syncLabelToClusters(ctx, label)
}

func (r *Reconciler) syncLabelToClusters(ctx context.Context, label *clusterv1alpha1.Label) error {
	klog.V(4).Infof("sync label %s[%s/%v] to clusters: %v", label.Name, label.Spec.Key, label.Spec.Value, label.Spec.Clusters)
	clusterSets := sets.NewString(label.Spec.Clusters...)
	for name := range clusterSets {
		cluster := &clusterv1alpha1.Cluster{}
		if err := r.Get(ctx, client.ObjectKey{Name: name}, cluster); err != nil {
			if errors.IsNotFound(err) {
				clusterSets.Delete(name)
				continue
			} else {
				return err
			}
		}

		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string)
		}
		if _, ok := cluster.Labels[fmt.Sprintf(clusterv1alpha1.ClusterLabelFormat, label.Name)]; ok {
			continue
		}
		cluster.Labels[fmt.Sprintf(clusterv1alpha1.ClusterLabelFormat, label.Name)] = ""
		if err := r.Update(ctx, cluster); err != nil {
			return err
		}
	}
	clusters := clusterSets.List()
	// some clusters have been deleted and this list needs to be updated
	if len(clusters) != len(label.Spec.Clusters) {
		label.Spec.Clusters = clusters
		return r.Update(ctx, label)
	}
	return nil
}

func (r *Reconciler) deleteLabel(ctx context.Context, label *clusterv1alpha1.Label) error {
	klog.V(4).Infof("deleting label %s, removing cluster %v related label", label.Name, label.Spec.Clusters)
	for _, name := range label.Spec.Clusters {
		cluster := &clusterv1alpha1.Cluster{}
		if err := r.Get(ctx, client.ObjectKey{Name: name}, cluster); err != nil {
			if errors.IsNotFound(err) {
				continue
			} else {
				return err
			}
		}
		delete(cluster.Labels, fmt.Sprintf(clusterv1alpha1.ClusterLabelFormat, label.Name))
		if err := r.Update(ctx, cluster); err != nil {
			return err
		}
	}
	label.Finalizers = nil
	return r.Update(ctx, label)
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()

	return builder.
		ControllerManagedBy(mgr).
		For(
			&clusterv1alpha1.Label{},
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			),
		).
		Complete(r)
}
