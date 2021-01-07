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

package v1alpha1

import (
	"github.com/emicklei/go-restful"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/klog"
	ksapi "kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/customalerting/v1alpha1"
	"kubesphere.io/kubesphere/pkg/informers"
	customalertingmodels "kubesphere.io/kubesphere/pkg/models/customalerting"
	"kubesphere.io/kubesphere/pkg/simple/client/customalerting"
)

type handler struct {
	operator customalertingmodels.Operator
}

func newHandler(informers informers.InformerFactory,
	promResourceClient promresourcesclient.Interface, ruleClient customalerting.RuleClient,
	option *customalerting.Options) *handler {
	return &handler{
		operator: customalertingmodels.NewOperator(
			informers, promResourceClient, ruleClient, option),
	}
}

func (h *handler) handleListCustomAlertingRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	query, err := v1alpha1.ParseAlertingRuleQueryParams(req)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	rules, err := h.operator.ListCustomAlertingRules(req.Request.Context(), namespace, query)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
	resp.WriteEntity(rules)
}

func (h *handler) handleListCustomRulesAlerts(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	query, err := v1alpha1.ParseAlertQueryParams(req)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	alerts, err := h.operator.ListCustomRulesAlerts(req.Request.Context(), namespace, query)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
	resp.WriteEntity(alerts)
}

func (h *handler) handleGetCustomAlertingRule(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	ruleName := req.PathParameter("rule_name")

	rule, err := h.operator.GetCustomAlertingRule(req.Request.Context(), namespace, ruleName)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
	if rule == nil {
		ksapi.HandleNotFound(resp, nil, err)
		return
	}
	resp.WriteEntity(rule)
}

func (h *handler) handleListCustomRuleAlerts(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	ruleName := req.PathParameter("rule_name")

	alerts, err := h.operator.ListCustomRuleAlerts(req.Request.Context(), namespace, ruleName)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
	resp.WriteEntity(alerts)
}

func (h *handler) handleCreateCustomAlertingRule(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var rule v1alpha1.PostableAlertingRule
	if err := req.ReadEntity(&rule); err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}
	if err := rule.Validate(); err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	err := h.operator.CreateCustomAlertingRule(req.Request.Context(), namespace, &rule)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		case err == v1alpha1.ErrAlertingRuleAlreadyExists:
			ksapi.HandleConflict(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
}

func (h *handler) handleUpdateCustomAlertingRule(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	ruleName := req.PathParameter("rule_name")

	var rule v1alpha1.PostableAlertingRule
	if err := req.ReadEntity(&rule); err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}
	if err := rule.Validate(); err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	err := h.operator.UpdateCustomAlertingRule(req.Request.Context(), namespace, ruleName, &rule)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
}

func (h *handler) handleDeleteCustomAlertingRule(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("rule_name")

	err := h.operator.DeleteCustomAlertingRule(req.Request.Context(), namespace, name)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrThanosRulerNotEnabled:
			ksapi.HandleBadRequest(resp, nil, err)
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
}

func (h *handler) handleListBuiltinAlertingRules(req *restful.Request, resp *restful.Response) {
	query, err := v1alpha1.ParseAlertingRuleQueryParams(req)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	rules, err := h.operator.ListBuiltinAlertingRules(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(rules)
}

func (h *handler) handleListBuiltinRulesAlerts(req *restful.Request, resp *restful.Response) {
	query, err := v1alpha1.ParseAlertQueryParams(req)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, nil, err)
		return
	}

	alerts, err := h.operator.ListBuiltinRulesAlerts(req.Request.Context(), query)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(alerts)
}

func (h *handler) handleGetBuiltinAlertingRule(req *restful.Request, resp *restful.Response) {
	ruleId := req.PathParameter("rule_id")

	rule, err := h.operator.GetBuiltinAlertingRule(req.Request.Context(), ruleId)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}
	if rule == nil {
		ksapi.HandleNotFound(resp, nil, err)
		return
	}

	resp.WriteEntity(rule)
}

func (h *handler) handleListBuiltinRuleAlerts(req *restful.Request, resp *restful.Response) {
	ruleId := req.PathParameter("rule_id")

	alerts, err := h.operator.ListBuiltinRuleAlerts(req.Request.Context(), ruleId)
	if err != nil {
		klog.Error(err)
		switch {
		case err == v1alpha1.ErrAlertingRuleNotFound:
			ksapi.HandleNotFound(resp, nil, err)
		default:
			ksapi.HandleInternalError(resp, nil, err)
		}
		return
	}

	resp.WriteEntity(alerts)
}
