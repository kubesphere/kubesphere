package fake

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"strings"
)

type FakeDevops struct {
	Data map[string]interface{}
}

func NewFakeDevops(data map[string]interface{}) *FakeDevops {
	var fakeData FakeDevops
	fakeData.Data = data
	return &fakeData
}

// Pipelinne operator interface
func (d *FakeDevops) GetPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.Pipeline, error) {
	return nil, nil
}

func (d *FakeDevops) ListPipelines(httpParameters *devops.HttpParameters) (*devops.PipelineList, error) {
	return nil, nil
}
func (d *FakeDevops) GetPipelineRun(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return nil, nil
}
func (d *FakeDevops) ListPipelineRuns(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineRunList, error) {
	return nil, nil
}
func (d *FakeDevops) StopPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) ReplayPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) RunPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) GetArtifacts(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return nil, nil
}
func (d *FakeDevops) GetRunLog(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *FakeDevops) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return nil, nil, nil
}
func (d *FakeDevops) GetNodeSteps(projectName, pipelineName, runId, nodeId string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	s := []string{projectName, pipelineName, runId, nodeId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.NodeSteps)
	return res, nil
}
func (d *FakeDevops) GetPipelineRunNodes(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.PipelineRunNodes, error) {
	s := []string{projectName, pipelineName, runId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.PipelineRunNodes)
	return res, nil
}
func (d *FakeDevops) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

//BranchPipelinne operator interface
func (d *FakeDevops) GetBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.BranchPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return nil, nil
}
func (d *FakeDevops) StopBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) RunBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return nil, nil
}
func (d *FakeDevops) GetBranchArtifacts(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return nil, nil
}
func (d *FakeDevops) GetBranchRunLog(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *FakeDevops) GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return nil, nil, nil
}
func (d *FakeDevops) GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	s := []string{projectName, pipelineName, branchName, runId, nodeId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.NodeSteps)
	return res, nil
}
func (d *FakeDevops) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]devops.BranchPipelineRunNodes, error) {
	s := []string{projectName, pipelineName, branchName, runId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.BranchPipelineRunNodes)
	return res, nil
}
func (d *FakeDevops) SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *FakeDevops) GetPipelineBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineBranch, error) {
	return nil, nil
}
func (d *FakeDevops) ScanBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

// Common pipeline operator interface
func (d *FakeDevops) GetConsoleLog(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *FakeDevops) GetCrumb(httpParameters *devops.HttpParameters) (*devops.Crumb, error) {
	return nil, nil
}

// SCM operator interface
func (d *FakeDevops) GetSCMServers(scmId string, httpParameters *devops.HttpParameters) ([]devops.SCMServer, error) {
	return nil, nil
}
func (d *FakeDevops) GetSCMOrg(scmId string, httpParameters *devops.HttpParameters) ([]devops.SCMOrg, error) {
	return nil, nil
}
func (d *FakeDevops) GetOrgRepo(scmId, organizationId string, httpParameters *devops.HttpParameters) ([]devops.OrgRepo, error) {
	return nil, nil
}
func (d *FakeDevops) CreateSCMServers(scmId string, httpParameters *devops.HttpParameters) (*devops.SCMServer, error) {
	return nil, nil
}
func (d *FakeDevops) Validate(scmId string, httpParameters *devops.HttpParameters) (*devops.Validates, error) {
	return nil, nil
}

//Webhook operator interface
func (d *FakeDevops) GetNotifyCommit(httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *FakeDevops) GithubWebhook(httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

func (d *FakeDevops) CheckScriptCompile(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.CheckScript, error) {
	return nil, nil
}
func (d *FakeDevops) CheckCron(projectName string, httpParameters *devops.HttpParameters) (*devops.CheckCronRes, error) {
	return nil, nil
}
func (d *FakeDevops) ToJenkinsfile(httpParameters *devops.HttpParameters) (*devops.ResJenkinsfile, error) {
	return nil, nil
}
func (d *FakeDevops) ToJson(httpParameters *devops.HttpParameters) (*devops.ResJson, error) {
	return nil, nil
}

// CredentialOperator
func (d *FakeDevops) CreateCredentialInProject(projectId string, credential *devops.Credential) (*string, error) {
	return nil, nil
}
func (d *FakeDevops) UpdateCredentialInProject(projectId string, credential *devops.Credential) (*string, error) {
	return nil, nil
}
func (d *FakeDevops) GetCredentialInProject(projectId, id string, content bool) (*devops.Credential, error) {
	return nil, nil
}
func (d *FakeDevops) GetCredentialsInProject(projectId string) ([]*devops.Credential, error) {
	return nil, nil
}
func (d *FakeDevops) DeleteCredentialInProject(projectId, id string) (*string, error) {
	return nil, nil
}

// BuildGetter
func (d *FakeDevops) GetProjectPipelineBuildByType(projectId, pipelineId string, status string) (*devops.Build, error) {
	return nil, nil
}
func (d *FakeDevops) GetMultiBranchPipelineBuildByType(projectId, pipelineId, branch string, status string) (*devops.Build, error) {
	return nil, nil
}

// ProjectMemberOperator
func (d *FakeDevops) AddProjectMember(membership *devops.ProjectMembership) (*devops.ProjectMembership, error) {
	return nil, nil
}
func (d *FakeDevops) UpdateProjectMember(oldMembership, newMembership *devops.ProjectMembership) (*devops.ProjectMembership, error) {
	return nil, nil
}
func (d *FakeDevops) DeleteProjectMember(membership *devops.ProjectMembership) (*devops.ProjectMembership, error) {
	return nil, nil
}

// ProjectPipelineOperator
func (d *FakeDevops) CreateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error) {
	return "", nil
}
func (d *FakeDevops) DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	return "", nil
}
func (d *FakeDevops) UpdateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error) {
	return "", nil
}
func (d *FakeDevops) GetProjectPipelineConfig(projectId, pipelineId string) (*devops.ProjectPipeline, error) {
	return nil, nil
}
