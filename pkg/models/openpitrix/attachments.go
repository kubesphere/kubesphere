/*
Copyright 2020 The KubeSphere Authors.
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

package openpitrix

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
)

type AttachmentInterface interface {
	DescribeAttachment(id string) (*Attachment, error)
}

type attachmentOperator struct {
	opClient openpitrix.Client
}

func newAttachmentOperator(opClient openpitrix.Client) AttachmentInterface {
	return &attachmentOperator{
		opClient: opClient,
	}
}

func (c *attachmentOperator) DescribeAttachment(id string) (*Attachment, error) {
	resp, err := c.opClient.GetAttachments(openpitrix.SystemContext(), &pb.GetAttachmentsRequest{
		AttachmentId: []string{id},
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(resp.Attachments) > 0 {
		return convertAttachment(resp.Attachments[id]), nil
	} else {
		err := status.New(codes.NotFound, "resource not found").Err()
		klog.Error(err)
		return nil, err
	}
}
