/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package openpitrix

import (
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
)

func DescribeAttachment(req *restful.Request, resp *restful.Response) {
	attachmentId := req.PathParameter("attachment")
	fileName := req.QueryParameter("filename")
	result, err := openpitrix.DescribeAttachment(attachmentId)
	// file raw
	if fileName != "" {
		data := result.AttachmentContent[fileName]
		resp.Write(data)
		resp.Header().Set("Content-Type", "text/plain")
		return
	}

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}
