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
package resources

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
)

func getKubectl(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("username")

	kubectlPod, err := kubectl.GetKubectlPod(user)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(kubectlPod)
}

func getKubeconfig(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("username")

	kubectlConfig, err := kubeconfig.GetKubeConfig(user)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(kubectlConfig)
}
