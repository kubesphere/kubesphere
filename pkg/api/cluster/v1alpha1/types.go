/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

type UpdateClusterRequest struct {
	KubeConfig []byte `json:"kubeconfig"`
}

type CreateLabelRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UnbindClustersRequest struct {
	Clusters []string `json:"clusters"`
}

type BindingClustersRequest struct {
	Labels   []string `json:"labels"`
	Clusters []string `json:"clusters"`
}

type LabelValue struct {
	Value string `json:"value"`
	ID    string `json:"id"`
}

type UpdateVisibilityRequest struct {
	Op        string `json:"op"`
	Workspace string `json:"workspace"`
}
