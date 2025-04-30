/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"
	goruntime "runtime"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
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
