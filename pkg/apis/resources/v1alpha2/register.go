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
package v1alpha2

import (
	"kubesphere.io/kubesphere/pkg/apibuilder"
	"kubesphere.io/kubesphere/pkg/apiserver/components"
	"kubesphere.io/kubesphere/pkg/apiserver/hpa"
	"kubesphere.io/kubesphere/pkg/apiserver/quotas"
	"kubesphere.io/kubesphere/pkg/apiserver/registries"
	"kubesphere.io/kubesphere/pkg/apiserver/resources"
	"kubesphere.io/kubesphere/pkg/apiserver/revisions"
	"kubesphere.io/kubesphere/pkg/apiserver/routers"
	"kubesphere.io/kubesphere/pkg/apiserver/workloadstatuses"
	"kubesphere.io/kubesphere/pkg/apiserver/workspaces"
)

var (
	GroupVersion = apibuilder.GroupVersion{Group: "resources.kubesphere.io", Version: "v1alpha2"}
	Routes       = []apibuilder.Route{components.V1Alpha2, hpa.V1Alpha2, quotas.V1Alpha2,
		registries.V1Alpha2, revisions.V1Alpha2, routers.V1Alpha2, resources.V1Alpha2,
		workloadstatuses.V1Alpha2, workspaces.V1Alpha2,
	}
	WebServiceBuilder = apibuilder.WebServiceBuilder{GroupVersion: GroupVersion, Routes: Routes}
	AddToContainer    = WebServiceBuilder.AddToContainer
)
