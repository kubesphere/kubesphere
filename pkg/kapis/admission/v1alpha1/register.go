/*
Copyright 2021 KubeSphere Authors

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
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ksapi "kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/admission"
	"net/http"
)

const (
	GroupName = "admission.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(container *restful.Container, informers informers.InformerFactory, ksClient kubesphere.Interface, option *admission.Options) error {
	ws := runtime.NewWebService(GroupVersion)
	var handler admissionHandlerInterface = newAdmissionHandler(informers, ksClient, option)

	// List
	ws.Route(ws.GET("/policytemplates").
		To(handler.handleListPolicies).
		Doc("list the admission policy templates").
		Param(ws.QueryParameter("template_name", "policy generate from temple name")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.PolicyTemplateList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTemplateTag}))

	ws.Route(ws.GET("/policies").
		To(handler.handleListPolicies).
		Doc("list the admission policies").
		Param(ws.QueryParameter("template_name", "policy generate from temple name")).
		Param(ws.QueryParameter("state", "state of a policy, one of `active`, `pending`, `inactive`")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.PolicyList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTag}))

	ws.Route(ws.GET("/policies/{policy_name}/rules").
		To(handler.handleListRules).
		Doc("list the alerts of the cluster-level custom alerting rules").
		Param(ws.QueryParameter("state", "state, one of `firing`, `pending`, `inactive`")).
		Param(ws.QueryParameter("page", "page of the result set").DataType("integer").DefaultValue("1")).
		Param(ws.QueryParameter("limit", "limit size of the result set").DataType("integer").DefaultValue("10")).
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.RuleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionRuleTag}))

	// Get
	ws.Route(ws.GET("/policytemplates/{policy_template_name}").
		To(handler.handleGetPolicyTemplate).
		Doc("get the polcy template with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.PolicyTemplateDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTemplateTag}))

	ws.Route(ws.GET("/policies/{policy_name}").
		To(handler.handleGetPolicy).
		Doc("get the policy template with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.PolicyDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTag}))

	ws.Route(ws.GET("/policies/{policy_name}/rules/{rule_name}").
		To(handler.handleGetRule).
		Doc("get the policy template with the specified name").
		Returns(http.StatusOK, ksapi.StatusOK, v1alpha1.RuleDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionRuleTag}))

	// Create
	ws.Route(ws.POST("/policies/{policy_name}").
		To(handler.handleCreatePolicy).
		Doc("create the cluster-level policy").
		Reads(v1alpha1.PostPolicy{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTag}))

	ws.Route(ws.POST("/policies/{policy_name}/rules/{rule_name}").
		To(handler.handleCreateRule).
		Doc("create the cluster-level rule for the policy").
		Reads(v1alpha1.PostRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionRuleTag}))

	// Update
	ws.Route(ws.PUT("/policies/{policy_name}").
		To(handler.handleUpdatePolicy).
		Doc("update the cluster-level policy").
		Reads(v1alpha1.PostPolicy{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTag}))

	ws.Route(ws.PUT("/policies/{policy_name}/rules/{rule_name}").
		To(handler.handleUpdateRule).
		Doc("update the cluster-level rule for the policy").
		Reads(v1alpha1.PostRule{}).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionRuleTag}))

	// delete
	ws.Route(ws.DELETE("/policies").
		To(handler.handleDeletePolicy).
		Doc("delete the cluster-level policy").
		Param(ws.QueryParameter("name", "policy name").CollectionFormat(restful.CollectionFormatMulti).AllowMultiple(true)).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionPolicyTag}))

	ws.Route(ws.DELETE("/policies/{policy_name}/rules").
		To(handler.handleDeleteRule).
		Doc("delete the cluster-level rule for the policy").
		Param(ws.QueryParameter("name", "rule name").CollectionFormat(restful.CollectionFormatMulti).AllowMultiple(true)).
		Returns(http.StatusOK, ksapi.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AdmissionRuleTag}))

	container.Add(ws)
	return nil
}
