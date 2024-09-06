/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	apiv1alpha1 "kubesphere.io/kubesphere/pkg/api/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/config"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/version"
)

const defaultTimeout = 10 * time.Second

type handler struct {
	client runtimeclient.Client
}

// updateKubeConfig updates the kubeconfig of the specific cluster, this API is used to update expired kubeconfig.
func (h *handler) updateKubeConfig(request *restful.Request, response *restful.Response) {
	var req apiv1alpha1.UpdateClusterRequest
	if err := request.ReadEntity(&req); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	ctx := request.Request.Context()

	clusterName := request.PathParameter("cluster")

	cluster := &clusterv1alpha1.Cluster{}
	if err := h.client.Get(ctx, types.NamespacedName{Name: clusterName}, cluster); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		api.HandleBadRequest(response, request, fmt.Errorf("update kubeconfig of the host cluster is not allowed"))
		return
	}
	// For member clusters that use proxy mode, we don't need to update the kubeconfig,
	// if the certs expired, just restart the tower component in the host cluster, it will renew the cert.
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
		api.HandleBadRequest(response, request, fmt.Errorf(
			"update kubeconfig of member clusters which using proxy mode is not allowed, their certs are managed and will be renewed by tower",
		))
		return
	}

	if len(req.KubeConfig) == 0 {
		api.HandleBadRequest(response, request, fmt.Errorf("cluster kubeconfig MUST NOT be empty"))
		return
	}
	config, err := k8sutil.LoadKubeConfigFromBytes(req.KubeConfig)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	config.Timeout = defaultTimeout
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	if _, err = clientSet.Discovery().ServerVersion(); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if _, err = validateKubeSphereAPIServer(ctx, clientSet); err != nil {
		api.HandleBadRequest(response, request, fmt.Errorf("unable validate kubesphere endpoint, %v", err))
		return
	}

	if err = h.validateMemberClusterConfiguration(ctx, clientSet); err != nil {
		api.HandleBadRequest(response, request, fmt.Errorf("failed to validate member cluster configuration, err: %v", err))
		return
	}

	// Check if the cluster is the same
	kubeSystem, err := clientSet.CoreV1().Namespaces().Get(ctx, metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	if kubeSystem.UID != cluster.Status.UID {
		api.HandleBadRequest(
			response, request, fmt.Errorf(
				"this kubeconfig corresponds to a different cluster than the previous one, you need to make sure that kubeconfig is not from another cluster",
			))
		return
	}

	cluster = cluster.DeepCopy()
	cluster.Spec.Connection.KubeConfig = req.KubeConfig
	if err = h.client.Update(ctx, cluster); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

// ValidateCluster validate cluster kubeconfig and kubesphere apiserver address, check their accessibility
func (h *handler) validateCluster(request *restful.Request, response *restful.Response) {
	var cluster clusterv1alpha1.Cluster
	if err := request.ReadEntity(&cluster); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	ctx := request.Request.Context()

	if cluster.Spec.Connection.Type != clusterv1alpha1.ConnectionTypeDirect {
		api.HandleBadRequest(response, request, fmt.Errorf("cluster connection type MUST be direct"))
		return
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		api.HandleBadRequest(response, request, fmt.Errorf("cluster kubeconfig MUST NOT be empty"))
		return
	}

	config, err := k8sutil.LoadKubeConfigFromBytes(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	config.Timeout = defaultTimeout
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if err = h.validateKubeConfig(ctx, cluster.Name, clientSet); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	// Check if the cluster is managed by other host cluster
	if err = clusterIsManaged(ctx, clientSet); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func clusterIsManaged(ctx context.Context, client kubernetes.Interface) error {
	kubeSphereNamespace, err := client.CoreV1().Namespaces().Get(ctx, constants.KubeSphereNamespace, metav1.GetOptions{})
	if err != nil {
		return runtimeclient.IgnoreNotFound(err)
	}
	hostClusterName := kubeSphereNamespace.Annotations[clusterv1alpha1.AnnotationHostClusterName]
	if hostClusterName != "" {
		return fmt.Errorf("current cluster is managed by another host cluster '%s'", hostClusterName)
	}
	return nil
}

// validateKubeConfig takes base64 encoded kubeconfig and check its validity
func (h *handler) validateKubeConfig(ctx context.Context, clusterName string, clientSet kubernetes.Interface) error {
	kubeSystem, err := clientSet.CoreV1().Namespaces().Get(ctx, metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return err
	}

	clusterList := &clusterv1alpha1.ClusterList{}
	if err := h.client.List(ctx, clusterList); err != nil {
		return err
	}

	// clusters with the exactly same kube-system namespace UID considered to be one
	// MUST not import the same cluster twice
	for _, existedCluster := range clusterList.Items {
		if existedCluster.Status.UID == kubeSystem.UID {
			return fmt.Errorf("cluster %s already exists (%s), MUST not import the same cluster twice", clusterName, existedCluster.Name)
		}
	}

	_, err = clientSet.Discovery().ServerVersion()
	return err
}

// validateKubeSphereAPIServer uses version api to check the accessibility
func validateKubeSphereAPIServer(ctx context.Context, clusterClient kubernetes.Interface) (*version.Info, error) {
	response, err := clusterClient.CoreV1().Services(constants.KubeSphereNamespace).
		ProxyGet("http", constants.KubeSphereAPIServerName, "80", "/version", nil).
		DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid response: %s, please make sure %s.%s.svc of member cluster is up and running", response, constants.KubeSphereAPIServerName, constants.KubeSphereNamespace)
	}

	ver := version.Info{}
	if err = json.Unmarshal(response, &ver); err != nil {
		return nil, fmt.Errorf("invalid response: %s, please make sure %s.%s.svc of member cluster is up and running", response, constants.KubeSphereAPIServerName, constants.KubeSphereNamespace)
	}
	return &ver, nil
}

// validateMemberClusterConfiguration compares host and member cluster jwt, if they are not same, it changes member
// cluster jwt to host's, then restart member cluster ks-apiserver.
func (h *handler) validateMemberClusterConfiguration(ctx context.Context, clientSet kubernetes.Interface) error {
	hConfig, err := h.getHostClusterConfig(ctx)
	if err != nil {
		return err
	}

	mConfig, err := h.getMemberClusterConfig(ctx, clientSet)
	if err != nil {
		return err
	}

	if hConfig.AuthenticationOptions.Issuer.JWTSecret != mConfig.AuthenticationOptions.Issuer.JWTSecret {
		return fmt.Errorf("hostcluster Jwt is not equal to member cluster jwt, please edit the member cluster cluster config")
	}

	return nil
}

// getMemberClusterConfig returns KubeSphere running config by the given member cluster kubeconfig
func (h *handler) getMemberClusterConfig(ctx context.Context, clientSet kubernetes.Interface) (*config.Config, error) {
	memberCm, err := clientSet.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Get(ctx, constants.KubeSphereConfigName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return config.FromConfigMap(memberCm)
}

// getHostClusterConfig returns KubeSphere running config from host cluster ConfigMap
func (h *handler) getHostClusterConfig(ctx context.Context) (*config.Config, error) {
	hostCm := &corev1.ConfigMap{}
	key := types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: constants.KubeSphereConfigName}
	if err := h.client.Get(ctx, key, hostCm); err != nil {
		return nil, fmt.Errorf("failed to get host cluster %s/configmap/%s, err: %s",
			constants.KubeSphereNamespace, constants.KubeSphereConfigName, err)
	}

	return config.FromConfigMap(hostCm)
}

func (h *handler) visibilityAuth(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	var visibilityRequests []apiv1alpha1.UpdateVisibilityRequest
	if err := req.ReadEntity(&visibilityRequests); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	patchData := make([]struct {
		workspace tenantv1beta1.WorkspaceTemplate
		patch     runtimeclient.Patch
	}, 0, 4)

	for _, visibilityRequest := range visibilityRequests {
		workspaceTemplate := tenantv1beta1.WorkspaceTemplate{}
		if err := h.client.Get(context.Background(), types.NamespacedName{Name: visibilityRequest.Workspace}, &workspaceTemplate); err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
		clusterSets := sets.New[string]()
		for _, clusterRef := range workspaceTemplate.Spec.Placement.Clusters {
			if clusterRef.Name != "" {
				clusterSets.Insert(clusterRef.Name)
			}
		}

		switch visibilityRequest.Op {
		case "add":
			clusterSets.Insert(clusterName)
		case "remove":
			if clusterSets.Has(clusterName) {
				clusterSets.Delete(clusterName)
			}
		default:
			api.HandleBadRequest(resp, req, errors.NewBadRequest("not support operation type"))
			return
		}
		newClusters := make([]tenantv1beta1.GenericClusterReference, 0, clusterSets.Len())
		for _, cluster := range clusterSets.UnsortedList() {
			newClusters = append(newClusters, tenantv1beta1.GenericClusterReference{Name: cluster})
		}
		workspaceTemplateCopy := workspaceTemplate.DeepCopy()
		workspaceTemplateCopy.Spec.Placement.Clusters = newClusters

		patchData = append(patchData, struct {
			workspace tenantv1beta1.WorkspaceTemplate
			patch     runtimeclient.Patch
		}{workspace: *workspaceTemplateCopy, patch: runtimeclient.MergeFrom(&workspaceTemplate)})
	}

	for _, pd := range patchData {
		if err := h.client.Patch(context.Background(), &pd.workspace, pd.patch); err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}
