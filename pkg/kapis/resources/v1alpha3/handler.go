package v1alpha3

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	resourcev1alpha2 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/server/params"
	"strings"
)

type Handler struct {
	resourceGetterV1alpha3  *resource.ResourceGetter
	resourcesGetterV1alpha2 *resourcev1alpha2.ResourceGetter
	componentsGetter        components.ComponentsGetter
}

func New(factory informers.InformerFactory) *Handler {
	return &Handler{
		resourceGetterV1alpha3:  resource.NewResourceGetter(factory),
		resourcesGetterV1alpha2: resourcev1alpha2.NewResourceGetter(factory),
		componentsGetter:        components.NewComponentsGetter(factory.KubernetesSharedInformerFactory()),
	}
}

// handleListResources retrieves resources
func (h *Handler) handleListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.resourceGetterV1alpha3.List(resourceType, namespace, query)

	if err == nil {
		response.WriteEntity(result)
		return
	}

	if err != resource.ErrResourceNotSupported {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	// fallback to v1alpha2
	result, err = h.fallback(resourceType, namespace, query)

	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

func (h *Handler) fallback(resourceType string, namespace string, q *query.Query) (*api.ListResult, error) {
	orderBy := string(q.SortBy)
	limit, offset := q.Pagination.Limit, q.Pagination.Offset
	reverse := !q.Ascending
	conditions := &params.Conditions{Match: make(map[string]string, 0), Fuzzy: make(map[string]string, 0)}
	for field, value := range q.Filters {
		switch field {
		case query.FieldName:
			conditions.Fuzzy[v1alpha2.Name] = string(value)
			break
		case query.FieldCreationTimeStamp:
			conditions.Match[v1alpha2.CreateTime] = string(value)
			break
		case query.FieldLastUpdateTimestamp:
			conditions.Match[v1alpha2.UpdateTime] = string(value)
			break
		case query.FieldLabel:
			values := strings.SplitN(string(value), ":", 2)
			if len(values) == 2 {
				conditions.Match[values[0]] = values[1]
			} else {
				conditions.Match[v1alpha2.Label] = values[0]
			}
			break
		case query.FieldAnnotation:
			values := strings.SplitN(string(value), ":", 2)
			if len(values) == 2 {
				conditions.Match[v1alpha2.Annotation] = values[1]
			} else {
				conditions.Match[v1alpha2.Annotation] = values[0]
			}
			break
		case query.FieldStatus:
			conditions.Match[v1alpha2.Status] = string(value)
			break
		case query.FieldOwnerReference:
			conditions.Match[v1alpha2.Owner] = string(value)
			break
		default:
			conditions.Match[string(field)] = string(value)
			break
		}
	}

	result, err := h.resourcesGetterV1alpha2.ListResources(namespace, resourceType, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &api.ListResult{
		Items:      result.Items,
		TotalItems: result.TotalCount,
	}, nil
}

func (h *Handler) handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)

	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

func (h *Handler) handleGetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()

	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

// get all componentsHandler
func (h *Handler) handleGetComponents(request *restful.Request, response *restful.Response) {

	result, err := h.componentsGetter.GetAllComponentsStatus()

	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}
