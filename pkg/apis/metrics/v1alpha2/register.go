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
	"kubesphere.io/kubesphere/pkg/apiserver/metrics"
)

var (
	GroupVersion      = apibuilder.GroupVersion{Group: "metrics.kubesphere.io", Version: "v1alpha2"}
	WebServiceBuilder = apibuilder.WebServiceBuilder{GroupVersion: GroupVersion, Routes: []apibuilder.Route{metrics.V1Alpha2}}
	AddToContainer    = WebServiceBuilder.AddToContainer
)
