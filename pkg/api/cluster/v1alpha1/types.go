package v1alpha1

type UpdateClusterRequest struct {
	KubeConfig []byte `json:"kubeconfig"`
}
