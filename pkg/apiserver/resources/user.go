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
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"net/http"

	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func GetKubectl(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	kubectlPod, err := kubectl.GetKubectlPod(user)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(kubectlPod)
}

func GetKubeconfig(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")

	kubectlConfig, err := kubeconfig.GetKubeConfig(user)

	if err != nil {
		klog.Error(err)
		if k8serr.IsNotFound(err) {
			// recreate
			kubeconfig.CreateKubeConfig(user)
			resp.WriteHeaderAndJson(http.StatusNotFound, errors.Wrap(err), restful.MIME_JSON)
		} else {
			resp.WriteHeaderAndJson(http.StatusInternalServerError, errors.Wrap(err), restful.MIME_JSON)
		}
		return
	}

	resp.Write([]byte(kubectlConfig))
}
