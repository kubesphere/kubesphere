/*
Copyright 2023.

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

package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"kubesphere.io/api/types/v1beta1"
)

// log is for logging in this package.
var workspacetemplatelog = logf.Log.WithName("workspacetemplate-resource")

func (r *WorkspaceTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &WorkspaceTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *WorkspaceTemplate) Default() {
	workspacetemplatelog.Info("default", "name", r.Name)
	r.removeDuplicatedClusters()
}

func (r *WorkspaceTemplate) removeDuplicatedClusters() {
	clusters := r.Spec.Placement.Clusters
	clusterSet := make(map[v1beta1.GenericClusterReference]bool)

	for _, v := range clusters {
		clusterSet[v] = true
	}

	if len(clusterSet) != len(clusters) {
		newClusters := make([]v1beta1.GenericClusterReference, 0)
		for k := range clusterSet {
			newClusters = append(newClusters, k)
		}
		r.Spec.Placement.Clusters = newClusters
	}
}
