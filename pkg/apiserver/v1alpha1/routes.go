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
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/components"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/hpa"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/metrics"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/operations"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/quotas"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/registries"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/revisions"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/routers"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/workloadstatuses"
	"kubesphere.io/kubesphere/pkg/apiserver/v1alpha1/workspaces"
)

func init() {
	addToWebServiceFuncs = append(addToWebServiceFuncs,
		components.Route,
		hpa.Route,
		metrics.Route,
		operations.Route,
		quotas.Route,
		registries.Route,
		routers.Route,
		revisions.Route,
		workloadstatuses.Route,
		workspaces.Route)
}
