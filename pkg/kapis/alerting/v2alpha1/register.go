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

package v2alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ksapi "kubesphere.io/kubesphere/pkg/api"
	alertingv2alpha1 "kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

const (
	groupName = "alerting.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v2alpha1"}

func AddToContainer(container *restful.Container, informers informers.InformerFactory,
	promResourceClient promresourcesclient.Interface, ruleClient alerting.RuleClient,
	option *alerting.Options) error {

	ws := runtime.NewWebService(GroupVersion)

	if informers == nil || promResourceClient == nil || ruleClient == nil || option == nil {
		h := func(req *restful.Request, resp *restful.Response) {
			ksapi.HandleBadRequest(resp, nil, alertingv2alpha1.ErrAlertingAPIV2NotEnabled)
			return
		}
		ws.Route(ws.GET("/{path:*}").To(h).Returns(http.StatusOK, ksapi.StatusOK, nil))
		ws.Route(ws.PUT("/{path:*}").To(h).Returns(http.StatusOK, ksapi.StatusOK, nil))
		ws.Route(ws.POST("/{path:*}").To(h).Returns(http.StatusOK, ksapi.StatusOK, nil))
		ws.Route(ws.DELETE("/{path:*}").To(h).Returns(http.StatusOK, ksapi.StatusOK, nil))
		ws.Route(ws.PATCH("/{path:*}").To(h).Returns(http.StatusOK, ksapi.StatusOK, nil))
		container.Add(ws)
		return nil
	}

	handler := newHandler(informers, promResourceClient, ruleClient, option)

	ws.Route(ws.GET("/rules").
		To(handler.handleListCustomAlertingRules).
		Doc("list the cluster-level custom alerting rules").
		Param(ws.QueryParameter("name", "rule name")).
		Param(ws.QueryParameter("state", "state of a rule based on its alerts, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("health", "health state of a rule based on the last execution, one of `ok`, `err`, `unknown`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("sort_field", "sort field, one of `name`, `lastEvaluation`, `evaluationTime`")).
		Param(ws.QueryParameter("sort_type", "sort type, one of `asc`, `desc`")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRuleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/alerts").
		To(handler.handleListCustomRulesAlerts).
		Doc("list the alerts of the cluster-level custom alerting rules").
		Param(ws.QueryParameter("state", "state, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/rules/{rule_name}").
		To(handler.handleGetCustomAlertingRule).
		Doc("get the cluster-level custom alerting rule with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/rules/{rule_name}/alerts").
		To(handler.handleListCustomRuleAlerts).
		Doc("list the alerts of the cluster-level custom alerting rule with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.POST("/rules").
		To(handler.handleCreateCustomAlertingRule).
		Doc("create a cluster-level custom alerting rule").
		Reads(alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.DELETE("/rules").
		To(handler.handleDeleteCustomAlertingRules).
		Doc("delete multiple cluster-level custom alerting rules").
		Param(ws.QueryParameter("name", "rule name").CollectionFormat(restful.CollectionFormatMulti).AllowMultiple(true)).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.BulkResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.POST("/bulkrules").
		To(handler.handleCreateOrUpdateCustomAlertingRules).
		Doc("create or update cluster-level custom alerting rules in bulk").
		Reads([]alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.BulkResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.PUT("/rules/{rule_name}").
		To(handler.handleUpdateCustomAlertingRule).
		Doc("update the cluster-level custom alerting rule with the specified name").
		Reads(alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.DELETE("/rules/{rule_name}").
		To(handler.handleDeleteCustomAlertingRule).
		Doc("delete the cluster-level custom alerting rule with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/rules").
		To(handler.handleListCustomAlertingRules).
		Doc("list the custom alerting rules in the specified namespace").
		Param(ws.QueryParameter("name", "rule name")).
		Param(ws.QueryParameter("state", "state of a rule based on its alerts, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("health", "health state of a rule based on the last execution, one of `ok`, `err`, `unknown`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("sort_field", "sort field, one of `name`, `lastEvaluation`, `evaluationTime`")).
		Param(ws.QueryParameter("sort_type", "sort type, one of `asc`, `desc`")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRuleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/alerts").
		To(handler.handleListCustomRulesAlerts).
		Doc("list the alerts of the custom alerting rules in the specified namespace.").
		Param(ws.QueryParameter("state", "state, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/rules/{rule_name}").
		To(handler.handleGetCustomAlertingRule).
		Doc("get the custom alerting rule with the specified name in the specified namespace").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/rules/{rule_name}/alerts").
		To(handler.handleListCustomRuleAlerts).
		Doc("get the alerts of the custom alerting rule with the specified name in the specified namespace").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/rules").
		To(handler.handleDeleteCustomAlertingRules).
		Doc("delete multiple custom alerting rules in the specified namespace").
		Param(ws.QueryParameter("name", "rule name").CollectionFormat(restful.CollectionFormatMulti).AllowMultiple(true)).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.BulkResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.POST("/namespaces/{namespace}/rules").
		To(handler.handleCreateCustomAlertingRule).
		Doc("create a custom alerting rule in the specified namespace").
		Reads(alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, "").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.POST("/namespaces/{namespace}/bulkrules").
		To(handler.handleCreateOrUpdateCustomAlertingRules).
		Doc("create or update custom alerting rules in bulk in the specified namespace").
		Reads([]alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.BulkResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.PUT("/namespaces/{namespace}/rules/{rule_name}").
		To(handler.handleUpdateCustomAlertingRule).
		Doc("update the custom alerting rule with the specified name in the specified namespace").
		Reads(alertingv2alpha1.PostableAlertingRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, "").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/rules/{rule_name}").
		To(handler.handleDeleteCustomAlertingRule).
		Doc("delete the custom alerting rule with the specified rule name in the specified namespace").
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/builtin/rules").
		To(handler.handleListBuiltinAlertingRules).
		Doc("list the builtin(non-custom) alerting rules").
		Param(ws.QueryParameter("name", "rule name")).
		Param(ws.QueryParameter("state", "state of a rule based on its alerts, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("health", "health state of a rule based on the last execution, one of `ok`, `err`, `unknown`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("sort_field", "sort field, one of `name`, `lastEvaluation`, `evaluationTime`")).
		Param(ws.QueryParameter("sort_type", "sort type, one of `asc`, `desc`")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRuleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/builtin/alerts").
		To(handler.handleListBuiltinRulesAlerts).
		Doc("list the alerts of the builtin(non-custom) rules").
		Param(ws.QueryParameter("state", "state, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("label_filters", "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/builtin/rules/{rule_id}").
		To(handler.handleGetBuiltinAlertingRule).
		Doc("get the builtin(non-custom) alerting rule with specified id").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.GettableAlertingRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/builtin/rules/{rule_id}/alerts").
		To(handler.handleListBuiltinRuleAlerts).
		Doc("list the alerts of the builtin(non-custom) alerting rule with the specified id").
		Returns(http.StatusOK, ksapi.StatusOK, alertingv2alpha1.AlertList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	container.Add(ws)

	return nil
}
