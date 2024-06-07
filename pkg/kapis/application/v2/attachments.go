/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"kubesphere.io/kubesphere/pkg/simple/client/application"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/strfmt"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
)

type Attachment struct {
	AttachmentContent map[string]strfmt.Base64 `json:"attachment_content,omitempty"`
	AttachmentID      string                   `json:"attachment_id,omitempty"`
}

func (h *appHandler) DescribeAttachment(req *restful.Request, resp *restful.Response) {
	attachmentId := req.PathParameter("attachment")
	data, err := application.FailOverGet(h.cmStore, h.ossStore, attachmentId, h.client, false)
	if requestDone(err, resp) {
		return
	}
	result := &Attachment{AttachmentID: attachmentId,
		AttachmentContent: map[string]strfmt.Base64{
			"raw": data,
		},
	}
	resp.WriteEntity(result)
}

func (h *appHandler) CreateAttachment(req *restful.Request, resp *restful.Response) {
	err := req.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	var att *Attachment
	// just save one attachment
	for fName := range req.Request.MultipartForm.File {
		f, _, err := req.Request.FormFile(fName)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
		data, _ := io.ReadAll(f)
		f.Close()

		id := application.GetUuid36(fmt.Sprintf("%s-att-", fName))

		err = application.FailOverUpload(h.cmStore, h.ossStore, id, bytes.NewBuffer(data), len(data))
		if err != nil {
			klog.Errorf("upload attachment failed, err: %s", err)
			api.HandleBadRequest(resp, nil, err)
			return
		}
		klog.V(4).Infof("upload attachment success")
		att = &Attachment{AttachmentID: id}
		break
	}

	resp.WriteEntity(att)
}

func (h *appHandler) DeleteAttachments(req *restful.Request, resp *restful.Response) {
	attachmentId := req.PathParameter("attachment")
	ids := strings.Split(attachmentId, ",")
	err := application.FailOverDelete(h.cmStore, h.ossStore, ids)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
}
