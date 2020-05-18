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
