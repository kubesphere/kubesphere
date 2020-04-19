package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type iamHandler struct {
	am am.AccessManagementInterface
	im im.IdentityManagementInterface
}

func newIAMHandler(im im.IdentityManagementInterface, am am.AccessManagementInterface, options *authoptions.AuthenticationOptions) *iamHandler {
	return &iamHandler{
		am: am,
		im: im,
	}
}

func (h *iamHandler) DescribeUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	user, err := h.im.DescribeUser(username)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	globalRole, err := h.am.GetGlobalRoleOfUser(username)

	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	result := iamv1alpha2.UserDetail{User: user, GlobalRole: globalRole}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListUsers(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.im.ListUsers(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	for i, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		user = user.DeepCopy()
		role, err := h.am.GetGlobalRoleOfUser(user.Name)
		if err != nil && !errors.IsNotFound(err) {
			klog.Error(err)
			api.HandleInternalError(resp, req, err)
			return
		}

		if user.Annotations == nil {
			user.Annotations = make(map[string]string, 0)
		}

		if role != nil {
			user.Annotations["iam.kubesphere.io/global-role"] = role.Name
		} else {
			user.Annotations["iam.kubesphere.io/global-role"] = ""
		}

		result.Items[i] = user
	}

	resp.WriteEntity(result)

}

func (h *iamHandler) ListRoles(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListRoles(namespace, queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) ListClusterRoles(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListClusterRoles(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) ListGlobalRoles(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListGlobalRoles(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) ListRoleUsers(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListNamespaceUsers(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	namespace := req.PathParameter("namespace")

	roleBindings, err := h.am.ListRoleBindings("", namespace)

	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	users := make([]runtime.Object, 0)

	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				user, err := h.im.DescribeUser(subject.Name)

				if errors.IsNotFound(err) {
					klog.Errorf("orphan subject: %+v", subject)
					continue
				}

				if err != nil {
					api.HandleInternalError(resp, req, err)
					return
				}

				user = user.DeepCopy()

				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}

				user.Annotations["iam.kubesphere.io/role"] = roleBinding.RoleRef.Name

				users = append(users, user)
			}
		}
	}

	result := resources.DefaultList(users, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*corev1.Namespace).ObjectMeta, right.(*corev1.Namespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		user := object.(*iamv1alpha2.User).ObjectMeta
		return resources.DefaultObjectMetaFilter(user, filter)
	})

	resp.WriteEntity(result)
}

func (h *iamHandler) ListWorkspaceRoles(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	workspace := request.PathParameter("workspace")
	queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s:%s", tenantv1alpha1.WorkspaceLabel, workspace))

	result, err := h.am.ListWorkspaceRoles(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListWorkspaceUsers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	workspace := request.PathParameter("workspace")

	roleBindings, err := h.am.ListWorkspaceRoleBindings("", workspace)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	users := make([]runtime.Object, 0)

	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				user, err := h.im.DescribeUser(subject.Name)

				if errors.IsNotFound(err) {
					klog.Errorf("orphan subject: %+v", subject)
					continue
				}

				if err != nil {
					api.HandleInternalError(response, request, err)
					return
				}

				user = user.DeepCopy()

				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}

				user.Annotations["iam.kubesphere.io/workspace-role"] = roleBinding.RoleRef.Name

				users = append(users, user)
			}
		}
	}

	result := resources.DefaultList(users, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*corev1.Namespace).ObjectMeta, right.(*corev1.Namespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		user := object.(*iamv1alpha2.User).ObjectMeta
		return resources.DefaultObjectMetaFilter(user, filter)
	})

	response.WriteEntity(result)
}
