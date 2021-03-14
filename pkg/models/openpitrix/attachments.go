///*
//Copyright 2020 The KubeSphere Authors.
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package openpitrix

import (
	"bytes"
	"github.com/go-openapi/strfmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
)

type AttachmentInterface interface {
	DescribeAttachment(id string) (*Attachment, error)
	CreateAttachment(data []byte) (*Attachment, error)
	DeleteAttachments(ids []string) error
}

type attachmentOperator struct {
	backingStoreClient s3.Interface
}

func newAttachmentOperator(storeClient s3.Interface) AttachmentInterface {
	return &attachmentOperator{
		backingStoreClient: storeClient,
	}
}

func (c *attachmentOperator) DescribeAttachment(id string) (*Attachment, error) {
	if c.backingStoreClient == nil {
		return nil, invalidS3Config
	}
	data, err := c.backingStoreClient.Read(id)

	if err != nil {
		klog.Errorf("read attachment %s failed, error: %s", id, err)
		return nil, downloadFileFailed
	}

	att := &Attachment{AttachmentID: id,
		AttachmentContent: map[string]strfmt.Base64{
			"raw": data,
		},
	}

	return att, nil
}
func (c *attachmentOperator) CreateAttachment(data []byte) (*Attachment, error) {
	if c.backingStoreClient == nil {
		return nil, invalidS3Config
	}
	id := idutils.GetUuid36(v1alpha1.HelmAttachmentPrefix)

	err := c.backingStoreClient.Upload(id, id, bytes.NewBuffer(data))
	if err != nil {
		klog.Errorf("upload attachment failed, err: %s", err)
		return nil, err
	}
	klog.V(4).Infof("upload attachment success")

	att := &Attachment{AttachmentID: id}
	return att, nil
}

func (c *attachmentOperator) DeleteAttachments(ids []string) error {
	if c.backingStoreClient == nil {
		return invalidS3Config
	}
	for _, id := range ids {
		err := c.backingStoreClient.Delete(id)
		if err != nil {
			return err
		}
	}
	return nil
}
