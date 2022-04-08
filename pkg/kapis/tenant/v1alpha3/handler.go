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

package v1alpha3

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	monitoringclient "kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type tenantHandler struct {
	tenant          tenant.Interface
	meteringOptions *meteringclient.Options
}

func newTenantHandler(factory informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface,
	evtsClient events.Client, loggingClient logging.Client, auditingclient auditing.Client,
	am am.AccessManagementInterface, authorizer authorizer.Authorizer,
	monitoringclient monitoringclient.Interface, resourceGetter *resourcev1alpha3.ResourceGetter,
	meteringOptions *meteringclient.Options, stopCh <-chan struct{}) *tenantHandler {

	if meteringOptions == nil || meteringOptions.RetentionDay == "" {
		meteringOptions = &meteringclient.DefaultMeteringOption
	}

	return &tenantHandler{
		tenant:          tenant.New(factory, k8sclient, ksclient, evtsClient, loggingClient, auditingclient, am, authorizer, monitoringclient, resourceGetter, stopCh),
		meteringOptions: meteringOptions,
	}
}

func (h *tenantHandler) ListWorkspaces(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	user, ok := request.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	result, err := h.tenant.ListWorkspaces(user, queryParam)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}
