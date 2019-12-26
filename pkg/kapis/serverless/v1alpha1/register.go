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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "serverless.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	c.Add(ws)
	return nil
}
