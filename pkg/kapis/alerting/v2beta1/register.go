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
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	alertingv2beta1 "kubesphere.io/api/alerting/v2beta1"

	kapi "kubesphere.io/kubesphere/pkg/api"
	kapialertingv2beta1 "kubesphere.io/kubesphere/pkg/api/alerting/v2beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

func AddToContainer(container *restful.Container, informers informers.InformerFactory, ruleClient alerting.RuleClient) error {

	ws := runtime.NewWebService(alertingv2beta1.SchemeGroupVersion)

	handler := newHandler(informers, ruleClient)

	ws.Route(ws.GET("/namespaces/{namespace}/rulegroups").
		To(handler.handleListRuleGroups).
		Doc("list the rulegroups in the specified namespace").
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state of a rulegroup, one of `firing`, `pending`, `inactive` depending on its rules")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/rulegroups/{name}").
		To(handler.handleGetRuleGroup).
		Doc("get the rulegroup with the specified name in the specified namespace").
		Returns(http.StatusOK, kapi.StatusOK, kapialertingv2beta1.RuleGroup{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/alerts").
		To(handler.handleListAlerts).
		Doc("list the alerts in the specified namespace").
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, one of `activeAt`. e.g. orderBy=activeAt")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state, one of `firing`, `pending`")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldAlertLabelFilters, "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/clusterrulegroups").
		To(handler.handleListClusterRuleGroups).
		Doc("list the clusterrulegroups").
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state of a rulegroup, one of `firing`, `pending`, `inactive` depending on its rules")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/clusterrulegroups/{name}").
		To(handler.handleGetClusterRuleGroup).
		Doc("get the clusterrulegroup with the specified name").
		Returns(http.StatusOK, kapi.StatusOK, kapialertingv2beta1.RuleGroup{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/clusteralerts").
		To(handler.handleListClusterAlerts).
		Doc("list the alerts of clusterrulegroups in the cluster").
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, one of `activeAt`. e.g. orderBy=activeAt")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state, one of `firing`, `pending`")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldAlertLabelFilters, "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/globalrulegroups").
		To(handler.handleListGlobalRuleGroups).
		Doc("list the globalrulegroups").
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state of a rulegroup, one of `firing`, `pending`, `inactive` depending on its rules")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldBuiltin, "filter rule groups, `true` for built-in rule groups and `false` for custom rule groups")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/globalrulegroups/{name}").
		To(handler.handleGetGlobalRuleGroup).
		Doc("get the globalrulegroup with the specified name").
		Returns(http.StatusOK, kapi.StatusOK, kapialertingv2beta1.RuleGroup{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	ws.Route(ws.GET("/globalalerts").
		To(handler.handleListGlobalAlerts).
		Doc("list the alerts of globalrulegroups in the cluster").
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, one of `activeAt`. e.g. orderBy=activeAt")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldState, "state, one of `firing`, `pending`")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldAlertLabelFilters, "label filters, concatenating multiple filters with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").DataFormat("key=%s,key~%s")).
		Param(ws.QueryParameter(kapialertingv2beta1.FieldBuiltin, "filter alerts, `true` for alerts from built-in rule groups and `false` for alerts from custom rule groups")).
		Returns(http.StatusOK, kapi.StatusOK, kapi.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AlertingTag}))

	container.Add(ws)

	return nil
}
