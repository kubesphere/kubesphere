/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2beta1

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"

	kapi "kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	alertingmodels "kubesphere.io/kubesphere/pkg/models/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

type handler struct {
	operator alertingmodels.RuleGroupOperator
}

func newHandler(informers informers.InformerFactory, ruleClient alerting.RuleClient) *handler {
	return &handler{
		operator: alertingmodels.NewRuleGroupOperator(informers, ruleClient),
	}
}

func (h *handler) handleListRuleGroups(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListRuleGroups(req.Request.Context(), namespace, query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleGetRuleGroup(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	result, err := h.operator.GetRuleGroup(req.Request.Context(), namespace, name)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleListAlerts(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListAlerts(req.Request.Context(), namespace, query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleListClusterRuleGroups(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListClusterRuleGroups(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleGetClusterRuleGroup(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	result, err := h.operator.GetClusterRuleGroup(req.Request.Context(), name)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleListClusterAlerts(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListClusterAlerts(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleListGlobalRuleGroups(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListGlobalRuleGroups(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleGetGlobalRuleGroup(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	result, err := h.operator.GetGlobalRuleGroup(req.Request.Context(), name)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *handler) handleListGlobalAlerts(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)

	result, err := h.operator.ListGlobalAlerts(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		kapi.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}
