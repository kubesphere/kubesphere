/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"
	goruntime "runtime"

	"kubesphere.io/kubesphere/pkg/server/params"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resv1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

func (h *appHandler) conflictedDone(req *restful.Request, resp *restful.Response, pathParam string, obj client.Object) bool {
	existed, err := h.checkConflicted(req, pathParam, obj)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return true
	}
	if existed {
		kind := obj.GetObjectKind().GroupVersionKind().Kind
		api.HandleConflict(resp, req, fmt.Errorf("%s %s already exists", kind, obj.GetName()))
		return true
	}
	return false
}

func (h *appHandler) checkConflicted(req *restful.Request, pathParam string, obj client.Object) (bool, error) {

	key := runtimeclient.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	if req.PathParameter(pathParam) != "" {
		//if route like /repos/{repo} and request path has repo, then it's update, filling obj
		err := h.client.Get(req.Request.Context(), key, obj)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	//if route like /repos, then it's create
	err := h.client.Get(req.Request.Context(), key, obj)
	if err != nil && apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func requestDone(err error, resp *restful.Response) bool {

	_, file, line, _ := goruntime.Caller(1)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, nil, err)
			return true
		}
		klog.Errorf("%s:%d request done with error: %v", file, line, err)
		api.HandleInternalError(resp, nil, err)
		return true
	}
	return false
}
func removeQueryArg(req *restful.Request, args ...string) {
	//The default filter is a whitelist, so delete some of our custom logical parameters
	for _, i := range args {
		q := req.Request.URL.Query()
		q.Del(i)
		req.Request.URL.RawQuery = q.Encode()
	}
}

func convertToListResult(obj runtime.Object, req *restful.Request) (listResult api.ListResult) {
	removeQueryArg(req, params.ConditionsParam, "global", "create")
	_ = meta.EachListItem(obj, omitManagedFields)
	queryParams := query.ParseQueryParameter(req)
	list, _ := meta.ExtractList(obj)
	items, _, totalCount := resv1beta1.DefaultList(list, queryParams, resv1beta1.DefaultCompare, resv1beta1.DefaultFilter)

	listResult.Items = items
	listResult.TotalItems = totalCount

	return listResult
}
func omitManagedFields(o runtime.Object) error {
	a, err := meta.Accessor(o)
	if err != nil {
		return err
	}
	a.SetManagedFields(nil)
	return nil
}
