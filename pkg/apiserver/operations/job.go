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

package operations

import (
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/workloads"

	"github.com/emicklei/go-restful"

	"fmt"
)

func rerunJob(req *restful.Request, resp *restful.Response) {
	var err error

	job := req.PathParameter("job")
	namespace := req.PathParameter("namespace")
	action := req.QueryParameter("a")

	switch action {
	case "rerun":
		err = workloads.JobReRun(namespace, job)
	default:
		errors.HandleError(errors.New(errors.InvalidArgument, fmt.Sprintf("invalid operation %s", action)), resp)
		return
	}
	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}
