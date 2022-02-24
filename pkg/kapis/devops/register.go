/*
Copyright 2022 The KubeSphere Authors.

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

package devops

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/kapis/generic"
)

var devopsGroupVersions = []schema.GroupVersion{
	{Group: "devops.kubesphere.io", Version: "v1alpha2"},
	{Group: "devops.kubesphere.io", Version: "v1alpha3"},
	// TODO Add other group versions here, like cd.devops.kubesphere.io
}

// AddToContainer registers DevOps proxies to the container.
func AddToContainer(container *restful.Container, endpoint string) error {
	for _, groupVersion := range devopsGroupVersions {
		proxy, err := generic.NewGenericProxy(endpoint, groupVersion.Group, groupVersion.Version)
		if err != nil {
			return err
		}
		if err = proxy.AddToContainer(container); err != nil {
			return err
		}
	}
	return nil
}
