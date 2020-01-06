package v1

import (
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/openpitrix/application"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/server/params"
	"openpitrix.io/openpitrix/pkg/pb"
)

type openpitrixHandler struct {
	namespacesGetter    *resource.NamespacedResourceGetter
	applicationOperator application.Interface
}

func newOpenpitrixHandler(factory informers.InformerFactory, client pb.ClusterManagerClient) *openpitrixHandler {
	return &openpitrixHandler{
		namespacesGetter:    resource.New(factory),
		applicationOperator: application.NewApplicaitonOperator(factory.KubernetesSharedInformerFactory(), client),
	}
}

func (h *openpitrixHandler) handleListApplications(request *restful.Request, response *restful.Response) {
	limit, offset := params.ParsePaging(request.QueryParameter(params.PagingParam))
	namespaceName := request.PathParameter("namespace")
	conditions, err := params.ParseConditions(request.QueryParameter(params.ConditionsParam))
	orderBy := request.QueryParameter(params.OrderByParam)
	reverse := params.ParseReverse(request)

	if orderBy == "" {
		orderBy = "create_time"
		reverse = true
	}

	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	if namespaceName != "" {
		namespace, err := h.namespacesGetter.Get(api.ResourceKindNamespace, "", namespaceName)

		if err != nil {
			api.HandleInternalError(response, err)
			return
		}
		var runtimeId string

		if ns, ok := namespace.(*v1.Namespace); ok {
			runtimeId = ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]
		}

		if runtimeId == "" {
			response.WriteAsJson(models.PageableResponse{Items: []interface{}{}, TotalCount: 0})
			return
		} else {
			conditions.Match["runtime_id"] = runtimeId
		}
	}

	result, err := h.applicationOperator.List(conditions, limit, offset, orderBy, reverse)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(response, err)
		return
	}

	resp.WriteAsJson(result)
}
