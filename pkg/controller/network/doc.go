// +kubebuilder:rbac:groups=network.kubesphere.io,resources=namespacenetworkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups:core,resource=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups:core,resource=services,verbs=get;list;watch;create;update;patch
package network
