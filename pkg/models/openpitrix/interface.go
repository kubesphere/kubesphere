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
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
)

type Interface interface {
	ApplicationInterface
	AppTemplateInterface
	AttachmentInterface
	CategoryInterface
	RepoInterface
}
type openpitrixOperator struct {
	ApplicationInterface
	AppTemplateInterface
	AttachmentInterface
	CategoryInterface
	RepoInterface
}

func NewOpenpitrixOperator(informers informers.SharedInformerFactory, opClient openpitrix.Client) Interface {
	if opClient == nil {
		return nil
	}

	return &openpitrixOperator{
		ApplicationInterface: newApplicationOperator(informers, opClient),
		AppTemplateInterface: newAppTemplateOperator(opClient),
		AttachmentInterface:  newAttachmentOperator(opClient),
		CategoryInterface:    newCategoryOperator(opClient),
		RepoInterface:        newRepoOperator(opClient),
	}
}
