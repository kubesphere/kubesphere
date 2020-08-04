/*
Copyright 2020 KubeSphere Authors

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

package fake

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"net/url"
	"strings"
)

type Devops struct {
	Data map[string]interface{}

	Projects map[string]interface{}

	Pipelines map[string]map[string]*devopsv1alpha3.Pipeline

	Credentials map[string]map[string]*v1.Secret
}

func New(projects ...string) *Devops {
	d := &Devops{
		Data:        nil,
		Projects:    map[string]interface{}{},
		Pipelines:   map[string]map[string]*devopsv1alpha3.Pipeline{},
		Credentials: map[string]map[string]*v1.Secret{},
	}
	for _, p := range projects {
		d.Projects[p] = true
	}
	return d
}
func NewWithPipelines(project string, pipelines ...*devopsv1alpha3.Pipeline) *Devops {
	d := &Devops{
		Data:        nil,
		Projects:    map[string]interface{}{},
		Pipelines:   map[string]map[string]*devopsv1alpha3.Pipeline{},
		Credentials: map[string]map[string]*v1.Secret{},
	}

	d.Projects[project] = true
	d.Pipelines[project] = map[string]*devopsv1alpha3.Pipeline{}
	for _, f := range pipelines {
		d.Pipelines[project][f.Name] = f
	}
	return d
}

func NewWithCredentials(project string, credentials ...*v1.Secret) *Devops {
	d := &Devops{
		Data:        nil,
		Projects:    map[string]interface{}{},
		Credentials: map[string]map[string]*v1.Secret{},
	}

	d.Projects[project] = true
	d.Credentials[project] = map[string]*v1.Secret{}
	for _, f := range credentials {
		d.Credentials[project][f.Name] = f
	}
	return d
}

func (d *Devops) CreateDevOpsProject(projectId string) (string, error) {
	if _, ok := d.Projects[projectId]; ok {
		return projectId, nil
	}
	d.Projects[projectId] = true
	d.Pipelines[projectId] = map[string]*devopsv1alpha3.Pipeline{}
	d.Credentials[projectId] = map[string]*v1.Secret{}
	return projectId, nil
}

func (d *Devops) DeleteDevOpsProject(projectId string) error {
	if _, ok := d.Projects[projectId]; ok {
		delete(d.Projects, projectId)
		delete(d.Pipelines, projectId)
		delete(d.Credentials, projectId)
		return nil
	} else {
		return &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
			},
			Message: "",
		}
	}
}

func (d *Devops) GetDevOpsProject(projectId string) (string, error) {
	if _, ok := d.Projects[projectId]; ok {
		return projectId, nil
	} else {
		return "", &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
	}
}

func NewFakeDevops(data map[string]interface{}) *Devops {
	var fakeData Devops
	fakeData.Data = data
	return &fakeData
}

// Pipelinne operator interface
func (d *Devops) GetPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.Pipeline, error) {
	return nil, nil
}

func (d *Devops) ListPipelines(httpParameters *devops.HttpParameters) (*devops.PipelineList, error) {
	return nil, nil
}
func (d *Devops) GetPipelineRun(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return nil, nil
}
func (d *Devops) ListPipelineRuns(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineRunList, error) {
	return nil, nil
}
func (d *Devops) StopPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return nil, nil
}
func (d *Devops) ReplayPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return nil, nil
}
func (d *Devops) RunPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return nil, nil
}
func (d *Devops) GetArtifacts(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return nil, nil
}
func (d *Devops) GetRunLog(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *Devops) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return nil, nil, nil
}
func (d *Devops) GetNodeSteps(projectName, pipelineName, runId, nodeId string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	s := []string{projectName, pipelineName, runId, nodeId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.NodeSteps)
	return res, nil
}
func (d *Devops) GetPipelineRunNodes(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.PipelineRunNodes, error) {
	s := []string{projectName, pipelineName, runId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.PipelineRunNodes)
	return res, nil
}
func (d *Devops) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

//BranchPipelinne operator interface
func (d *Devops) GetBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.BranchPipeline, error) {
	return nil, nil
}
func (d *Devops) GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	return nil, nil
}
func (d *Devops) StopBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	return nil, nil
}
func (d *Devops) ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	return nil, nil
}
func (d *Devops) RunBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	return nil, nil
}
func (d *Devops) GetBranchArtifacts(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	return nil, nil
}
func (d *Devops) GetBranchRunLog(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *Devops) GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	return nil, nil, nil
}
func (d *Devops) GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, httpParameters *devops.HttpParameters) ([]devops.NodeSteps, error) {
	s := []string{projectName, pipelineName, branchName, runId, nodeId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.NodeSteps)
	return res, nil
}
func (d *Devops) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]devops.BranchPipelineRunNodes, error) {
	s := []string{projectName, pipelineName, branchName, runId}
	key := strings.Join(s, "-")
	res := d.Data[key].([]devops.BranchPipelineRunNodes)
	return res, nil
}
func (d *Devops) SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *Devops) GetPipelineBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineBranch, error) {
	return nil, nil
}
func (d *Devops) ScanBranch(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

// Common pipeline operator interface
func (d *Devops) GetConsoleLog(projectName, pipelineName string, httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *Devops) GetCrumb(httpParameters *devops.HttpParameters) (*devops.Crumb, error) {
	return nil, nil
}

// SCM operator interface
func (d *Devops) GetSCMServers(scmId string, httpParameters *devops.HttpParameters) ([]devops.SCMServer, error) {
	return nil, nil
}
func (d *Devops) GetSCMOrg(scmId string, httpParameters *devops.HttpParameters) ([]devops.SCMOrg, error) {
	return nil, nil
}
func (d *Devops) GetOrgRepo(scmId, organizationId string, httpParameters *devops.HttpParameters) (devops.OrgRepo, error) {
	return devops.OrgRepo{}, nil
}
func (d *Devops) CreateSCMServers(scmId string, httpParameters *devops.HttpParameters) (*devops.SCMServer, error) {
	return nil, nil
}
func (d *Devops) Validate(scmId string, httpParameters *devops.HttpParameters) (*devops.Validates, error) {
	return nil, nil
}

//Webhook operator interface
func (d *Devops) GetNotifyCommit(httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}
func (d *Devops) GithubWebhook(httpParameters *devops.HttpParameters) ([]byte, error) {
	return nil, nil
}

func (d *Devops) CheckScriptCompile(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.CheckScript, error) {
	return nil, nil
}
func (d *Devops) CheckCron(projectName string, httpParameters *devops.HttpParameters) (*devops.CheckCronRes, error) {
	return nil, nil
}
func (d *Devops) ToJenkinsfile(httpParameters *devops.HttpParameters) (*devops.ResJenkinsfile, error) {
	return nil, nil
}
func (d *Devops) ToJson(httpParameters *devops.HttpParameters) (*devops.ResJson, error) {
	return nil, nil
}

// CredentialOperator
func (d *Devops) CreateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {
	if _, ok := d.Credentials[projectId][credential.Name]; ok {
		err := fmt.Errorf("credential name [%s] has been used", credential.Name)
		return "", restful.NewError(http.StatusConflict, err.Error())
	}
	d.Credentials[projectId][credential.Name] = credential
	return credential.Name, nil
}
func (d *Devops) UpdateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {
	if _, ok := d.Credentials[projectId][credential.Name]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return "", err
	}
	d.Credentials[projectId][credential.Name] = credential
	return credential.Name, nil
}

func (d *Devops) GetCredentialInProject(projectId, id string) (*devops.Credential, error) {
	if _, ok := d.Credentials[projectId][id]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return nil, err
	}
	return &devops.Credential{Id: id}, nil
}
func (d *Devops) GetCredentialsInProject(projectId string) ([]*devops.Credential, error) {
	return nil, nil
}
func (d *Devops) DeleteCredentialInProject(projectId, id string) (string, error) {
	if _, ok := d.Credentials[projectId][id]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return "", err
	}
	delete(d.Credentials[projectId], id)
	return "", nil
}

// BuildGetter
func (d *Devops) GetProjectPipelineBuildByType(projectId, pipelineId string, status string) (*devops.Build, error) {
	return nil, nil
}
func (d *Devops) GetMultiBranchPipelineBuildByType(projectId, pipelineId, branch string, status string) (*devops.Build, error) {
	return nil, nil
}

// ProjectPipelineOperator
func (d *Devops) CreateProjectPipeline(projectId string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	if _, ok := d.Pipelines[projectId][pipeline.Name]; ok {
		err := fmt.Errorf("pipeline name [%s] has been used", pipeline.Name)
		return "", restful.NewError(http.StatusConflict, err.Error())
	}
	d.Pipelines[projectId][pipeline.Name] = pipeline
	return "", nil
}

func (d *Devops) DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	if _, ok := d.Pipelines[projectId][pipelineId]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return "", err
	}
	delete(d.Pipelines[projectId], pipelineId)
	return "", nil
}

func (d *Devops) UpdateProjectPipeline(projectId string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	if _, ok := d.Pipelines[projectId][pipeline.Name]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return "", err
	}
	d.Pipelines[projectId][pipeline.Name] = pipeline
	return "", nil
}

func (d *Devops) GetProjectPipelineConfig(projectId, pipelineId string) (*devopsv1alpha3.Pipeline, error) {
	if _, ok := d.Pipelines[projectId][pipelineId]; !ok {
		err := &devops.ErrorResponse{
			Body: []byte{},
			Response: &http.Response{
				Status:        "404 Not Found",
				StatusCode:    404,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: 50,
				Header: http.Header{
					"Foo": []string{"Bar"},
				},
				Body: ioutil.NopCloser(strings.NewReader("foo")), // shouldn't be used
				Request: &http.Request{
					Method: "",
					URL: &url.URL{
						Scheme:     "",
						Opaque:     "",
						User:       nil,
						Host:       "",
						Path:       "",
						RawPath:    "",
						ForceQuery: false,
						RawQuery:   "",
						Fragment:   "",
					},
				},
			},
			Message: "",
		}
		return nil, err
	}

	return d.Pipelines[projectId][pipelineId], nil
}

func (d *Devops) AddGlobalRole(roleName string, ids devops.GlobalPermissionIds, overwrite bool) error {
	return nil
}

func (d *Devops) AddProjectRole(roleName string, pattern string, ids devops.ProjectPermissionIds, overwrite bool) error {
	return nil
}

func (d *Devops) DeleteProjectRoles(roleName ...string) error {
	return nil
}

func (d *Devops) AssignProjectRole(roleName string, sid string) error {
	return nil
}

func (d *Devops) UnAssignProjectRole(roleName string, sid string) error {
	return nil
}

func (d *Devops) AssignGlobalRole(roleName string, sid string) error {
	return nil
}

func (d *Devops) UnAssignGlobalRole(roleName string, sid string) error {
	return nil
}

func (d *Devops) DeleteUserInProject(sid string) error {
	return nil
}

func (d *Devops) GetGlobalRole(roleName string) (string, error) {
	return "", nil
}
