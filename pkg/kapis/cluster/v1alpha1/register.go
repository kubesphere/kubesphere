/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	apiv1alpha1 "kubesphere.io/kubesphere/pkg/api/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "cluster.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func NewHandler(cacheClient runtimeclient.Client) rest.Handler {
	return &handler{
		client: cacheClient,
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(container *restful.Container) error {
	webservice := runtime.NewWebService(GroupVersion)
	// TODO use validating admission webhook instead
	webservice.Route(webservice.POST("/clusters/validation").
		To(h.validateCluster).
		Deprecate().
		Doc("Cluster validation").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}).
		Reads(clusterv1alpha1.Cluster{}).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.PUT("/clusters/{cluster}/kubeconfig").
		To(h.updateKubeConfig).
		Doc("Update kubeconfig").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}).
		Param(webservice.PathParameter("cluster", "The specified cluster.").Required(true)).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.POST("/labels").
		Doc("Create cluster labels.").
		Reads([]apiv1alpha1.CreateLabelRequest{}).
		To(h.createLabels).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}))

	webservice.Route(webservice.DELETE("/labels").
		Doc("Delete cluster labels.").
		Reads([]string{}).
		To(h.deleteLabels).
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}))

	webservice.Route(webservice.PUT("/labels/{label}").
		Doc("Update a label.").
		Param(webservice.PathParameter("label", "Name of the label.").Required(true)).
		Reads(apiv1alpha1.CreateLabelRequest{}).
		To(h.updateLabel).
		Returns(http.StatusOK, api.StatusOK, clusterv1alpha1.Label{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}))

	webservice.Route(webservice.POST("/labelbindings").
		Doc("Binding clusters.").
		Reads([]apiv1alpha1.BindingClustersRequest{}).
		To(h.bindingClusters).
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}))

	webservice.Route(webservice.GET("/labels").
		Doc("List labels.").
		To(h.listLabelGroups).
		Returns(http.StatusOK, api.StatusOK, map[string][]apiv1alpha1.LabelValue{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagMultiCluster}))

	webservice.Route(webservice.POST("/clusters/{cluster}/grantrequests").
		To(h.visibilityAuth).
		Doc("Patch workspace template's visibility in different clusters").
		Operation("patch-workspace-template-clusters-visibility").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}).
		Param(webservice.PathParameter("cluster", "The specified cluster.").Required(true)).
		Reads([]apiv1alpha1.UpdateVisibilityRequest{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1beta1.WorkspaceTemplate{}))

	container.Add(webservice)
	return nil
}
