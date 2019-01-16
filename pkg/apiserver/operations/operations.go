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
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/errors"
)

func V1Alpha2(ws *restful.WebService) {
	tags := []string{"Operations"}

	ws.Route(ws.POST("/nodes/{node}/drainage").
		To(drainNode).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Param(ws.PathParameter("node", "node name").
			DataType("string")).
		Writes(errors.Error{}))

	ws.Route(ws.POST("/namespaces/{namespace}/jobs/{job}").
		To(rerunJob).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Handle job operation").
		Param(ws.PathParameter("job", "job name").
			DataType("string")).
		Param(ws.PathParameter("namespace", "job's namespace").
			DataType("string")).
		Param(ws.QueryParameter("a", "action").
			DataType("string")).
		Writes(errors.Error{}))
}
