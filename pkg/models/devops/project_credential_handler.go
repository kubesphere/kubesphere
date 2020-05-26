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

package devops

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

type ProjectCredentialGetter interface {
	GetProjectCredentialUsage(projectId, credentialId string) (*devops.Credential, error)
}

type projectCredentialGetter struct {
	devopsClient devops.Interface
}

// GetProjectCredentialUsage get the usage of Credential
func (o *projectCredentialGetter) GetProjectCredentialUsage(projectId, credentialId string) (*devops.Credential, error) {
	credential, err := o.devopsClient.GetCredentialInProject(projectId,
		credentialId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, err
	}
	return credential, nil
}

func NewProjectCredentialOperator(devopsClient devops.Interface) ProjectCredentialGetter {
	return &projectCredentialGetter{devopsClient: devopsClient}
}
