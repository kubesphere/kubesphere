/*

 Copyright 2019 The KubeSphere Authors.

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
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/informers"
	monitorhle "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type meterHandler interface {
	HandleClusterMetersQuery(req *restful.Request, resp *restful.Response)
	HandleNodeMetersQuery(req *restful.Request, resp *restful.Response)
	HandleWorkspaceMetersQuery(req *restful.Request, resp *restful.Response)
	HandleNamespaceMetersQuery(re *restful.Request, resp *restful.Response)
	HandleWorkloadMetersQuery(req *restful.Request, resp *restful.Response)
	HandleApplicationMetersQuery(req *restful.Request, resp *restful.Response)
	HandlePodMetersQuery(req *restful.Request, resp *restful.Response)
	HandleServiceMetersQuery(req *restful.Request, resp *restful.Response)
	HandlePVCMetersQuery(req *restful.Request, resp *restful.Response)
}

func newHandler(k kubernetes.Interface, m monitoring.Interface, f informers.InformerFactory, resourceGetter *resourcev1alpha3.ResourceGetter) meterHandler {
	return monitorhle.NewHandler(k, m, nil, f, resourceGetter)
}
