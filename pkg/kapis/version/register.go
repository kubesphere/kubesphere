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

package version

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/version"
)

// AddToContainer the api /kapis/version will be deprecated and instead with /version
func AddToContainer(container *restful.Container, k8sDiscovery discovery.DiscoveryInterface) error {
	webservice := runtime.NewWebService(schema.GroupVersion{})
	rootPathWebservice := &restful.WebService{}
	rootPathWebservice.Path("/").Produces(restful.MIME_JSON)

	versionFunc := func(request *restful.Request, response *restful.Response) {
		ksVersion := version.Get()

		if k8sDiscovery != nil {
			k8sVersion, err := k8sDiscovery.ServerVersion()
			if err == nil {
				ksVersion.Kubernetes = k8sVersion
			} else {
				klog.Errorf("Failed to get kubernetes version, error %v", err)
			}
		}

		response.WriteAsJson(ksVersion)
	}

	webservice.Route(webservice.GET("/version").
		Deprecate().
		To(versionFunc)).
		Doc("KubeSphere version. Deprecated: please use API `/version`")

	rootPathWebservice.Route(rootPathWebservice.GET("/version").
		To(versionFunc)).Doc("KubeSphere version")

	container.Add(webservice)
	container.Add(rootPathWebservice)

	return nil
}
