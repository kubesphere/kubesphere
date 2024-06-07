/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"

	"kubesphere.io/kubesphere/pkg/server/errors"
)

func (h *appHandler) CreateOrUpdateCategory(req *restful.Request, resp *restful.Response) {
	createCategoryRequest := &appv2.Category{}
	err := req.ReadEntity(createCategoryRequest)
	if requestDone(err, resp) {
		return
	}

	category := &appv2.Category{}
	category.Name = createCategoryRequest.Name

	if h.conflictedDone(req, resp, "category", category) {
		return
	}

	MutateFn := func() error {
		if category.GetAnnotations() == nil {
			category.SetAnnotations(make(map[string]string))
		}
		annotations := createCategoryRequest.GetAnnotations()
		category.Annotations[constants.DisplayNameAnnotationKey] = annotations[constants.DisplayNameAnnotationKey]
		category.Annotations[constants.DescriptionAnnotationKey] = annotations[constants.DescriptionAnnotationKey]
		category.Spec.Icon = createCategoryRequest.Spec.Icon
		return nil
	}
	_, err = controllerutil.CreateOrUpdate(req.Request.Context(), h.client, category, MutateFn)
	if requestDone(err, resp) {
		return
	}

	resp.WriteAsJson(category)
}

func (h *appHandler) DeleteCategory(req *restful.Request, resp *restful.Response) {
	categoryId := req.PathParameter("category")
	if categoryId == appv2.UncategorizedCategoryID {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s is default Category can't be delete", appv2.UncategorizedCategoryID))
		return
	}

	category := &appv2.Category{}
	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: categoryId}, category)
	if requestDone(err, resp) {
		return
	}

	if category.Status.Total > 0 {
		msg := fmt.Sprintf("can not delete helm category: %s which owns applications", categoryId)
		klog.Warningf(msg)
		api.HandleInternalError(resp, nil, errors.New(msg))
		return
	}
	err = h.client.Delete(req.Request.Context(), &appv2.Category{ObjectMeta: metav1.ObjectMeta{Name: categoryId}})
	if requestDone(err, resp) {
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *appHandler) DescribeCategory(req *restful.Request, resp *restful.Response) {
	categoryId := req.PathParameter("category")

	result := &appv2.Category{}
	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: categoryId}, result)
	if requestDone(err, resp) {
		return
	}
	result.SetManagedFields(nil)

	resp.WriteEntity(result)
}

func (h *appHandler) ListCategories(req *restful.Request, resp *restful.Response) {
	cList := &appv2.CategoryList{}
	err := h.client.List(req.Request.Context(), cList)
	if requestDone(err, resp) {
		return
	}
	resp.WriteEntity(convertToListResult(cList, req))
}
