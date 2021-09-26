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
	"github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
	"github.com/open-policy-agent/gatekeeper/pkg/target"
	"k8s.io/klog"
	ksapi "kubesphere.io/kubesphere/pkg/api"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	admissionmodel "kubesphere.io/kubesphere/pkg/models/admission"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
	"kubesphere.io/kubesphere/pkg/simple/client/admission"
)

type admissionHandlerInterface interface {
	// List
	handleListPolicyTemplates(req *restful.Request, resp *restful.Response)
	handleListPolicies(req *restful.Request, resp *restful.Response)
	handleListRules(req *restful.Request, resp *restful.Response)
	// Get
	handleGetPolicyTemplate(req *restful.Request, resp *restful.Response)
	handleGetPolicy(req *restful.Request, resp *restful.Response)
	handleGetRule(req *restful.Request, resp *restful.Response)
	// Create
	handleCreatePolicy(req *restful.Request, resp *restful.Response)
	handleCreateRule(req *restful.Request, resp *restful.Response)
	// Update
	handleUpdatePolicy(req *restful.Request, resp *restful.Response)
	handleUpdateRule(req *restful.Request, resp *restful.Response)
	// Delete
	handleDeletePolicy(req *restful.Request, resp *restful.Response)
	handleDeleteRule(req *restful.Request, resp *restful.Response)
}

type admissionHandler struct {
	Operator admissionmodel.Operator
}

func newAdmissionHandler(informers informers.InformerFactory, ksClient kubesphere.Interface, option *admission.Options) *admissionHandler {
	providers := map[string]provider.Provider{}

	driver := local.New(local.Tracing(false))
	backend, err := client.NewBackend(client.Driver(driver))
	if err != nil {
		klog.V(4).Infoln(err)
		return nil
	}

	c, err := backend.NewClient(client.Targets(&target.K8sValidationTarget{}))
	if option.Enable {
		if option.EnableGatekeeperProvider {
			providers[provider.GateKeeperProviderName] = provider.NewGateKeeperProvider(c)
		}
	}

	return &admissionHandler{
		Operator: admissionmodel.NewOperator(
			ksClient,
			informers.KubeSphereSharedInformerFactory(),
			providers,
		),
	}
}

// List

func (h admissionHandler) handleListPolicyTemplates(req *restful.Request, resp *restful.Response) {
	templateList, err := h.Operator.ListPolicyTemplates(req.Request.Context())
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, nil, err)
		return
	}
	_ = resp.WriteEntity(templateList)
}

func (h admissionHandler) handleListPolicies(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	policyList, err := h.Operator.ListPolicies(req.Request.Context(), namespace)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, nil, err)
		return
	}
	_ = resp.WriteEntity(policyList)
}

func (h admissionHandler) handleListRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	policyName := req.PathParameter("policy_name")
	ruleList, err := h.Operator.ListRules(req.Request.Context(), namespace, policyName)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, nil, err)
		return
	}
	_ = resp.WriteEntity(ruleList)
}

// Get

func (h admissionHandler) handleGetPolicyTemplate(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h admissionHandler) handleGetPolicy(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h admissionHandler) handleGetRule(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

// Create

func (h admissionHandler) handleCreatePolicy(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h admissionHandler) handleCreateRule(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

// Update

func (h admissionHandler) handleUpdatePolicy(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h admissionHandler) handleUpdateRule(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

// Delete

func (h admissionHandler) handleDeletePolicy(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h admissionHandler) handleDeleteRule(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}
