// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Gojenkins is a Jenkins Client in Go, that exposes the jenkins REST api in a more developer friendly way.
package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Requester *Requester
}

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init Method. Should be called after creating a Jenkins Instance.
// e.g jenkins := CreateJenkins("url").Init()
// HTTP Client is set here, Connection to jenkins is tested here.
func (j *Jenkins) Init() (*Jenkins, error) {
	j.initLoggers()

	rsp, err := j.Requester.GetJSON("/", nil, nil)
	if err != nil {
		return nil, err
	}

	j.Version = rsp.Header.Get("X-Jenkins")
	//if j.Raw == nil {
	//	return nil, errors.New("Connection Failed, Please verify that the host and credentials are correct.")
	//}

	return j, nil
}

func (j *Jenkins) initLoggers() {
	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Create a new folder
// This folder can be nested in other parent folders
// Example: jenkins.CreateFolder("newFolder", "grandparentFolder", "parentFolder")
func (j *Jenkins) CreateFolder(name, description string, parents ...string) (*Folder, error) {
	folderObj := &Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, name), "/job/")}
	folder, err := folderObj.Create(name, description)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

// Create a new job in the folder
// Example: jenkins.CreateJobInFolder("<config></config>", "newJobName", "myFolder", "parentFolder")
func (j *Jenkins) CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, jobName), "/job/")}
	qr := map[string]string{
		"name": jobName,
	}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Create a new job from config File
// Method takes XML string as first Parameter, and if the name is not specified in the config file
// takes name as string as second Parameter
// e.g jenkins.CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	} else {
		return nil, errors.New("Error Creating Job, job name is missing")
	}
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + qr["name"]}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Rename a job.
// First Parameter job old name, Second Parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

// Create a copy of a job.
// First Parameter Name of the job to copy from, Second Parameter new job name.
func (j *Jenkins) CopyJob(copyFrom string, newName string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + copyFrom}
	_, err := job.Poll()
	if err != nil {
		return nil, err
	}
	return job.Copy(newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(name string, parentIDs ...string) (bool, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, name), "/job/")}
	return job.Delete()
}

// Invoke a job.
// First Parameter job name, second Parameter is optional Build parameters.
func (j *Jenkins) BuildJob(name string, options ...interface{}) (int64, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (j *Jenkins) GetBuild(jobName string, number int64) (*Build, error) {
	job, err := j.GetJob(jobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetBuild(number)

	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetJob(id string, parentIDs ...string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, id), "/job/")}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetFolder(id string, parents ...string) (*Folder, error) {
	folder := Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, id), "/job/")}
	status, err := folder.Poll()
	if err != nil {
		return nil, fmt.Errorf("trouble polling folder: %v", err)
	}
	if status == 200 {
		return &folder, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)

func (j *Jenkins) Poll() (int, error) {
	resp, err := j.Requester.GetJSON("/", nil, nil)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

func (j *Jenkins) GetGlobalRole(roleName string) (*GlobalRole, error) {
	roleResponse := &GlobalRoleResponse{
		RoleName: roleName,
	}
	stringResponse := ""
	response, err := j.Requester.Get("/role-strategy/strategy/getRole",
		&stringResponse,
		map[string]string{
			"roleName": roleName,
			"type":     GLOBAL_ROLE,
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	if stringResponse == "{}" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(stringResponse), roleResponse)
	if err != nil {
		return nil, err
	}
	return &GlobalRole{
		Jenkins: j,
		Raw:     *roleResponse,
	}, nil
}

func (j *Jenkins) GetProjectRole(roleName string) (*ProjectRole, error) {
	roleResponse := &ProjectRoleResponse{
		RoleName: roleName,
	}
	stringResponse := ""
	response, err := j.Requester.Get("/role-strategy/strategy/getRole",
		&stringResponse,
		map[string]string{
			"roleName": roleName,
			"type":     PROJECT_ROLE,
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	if stringResponse == "{}" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(stringResponse), roleResponse)
	if err != nil {
		return nil, err
	}
	return &ProjectRole{
		Jenkins: j,
		Raw:     *roleResponse,
	}, nil
}

func (j *Jenkins) AddGlobalRole(roleName string, ids GlobalPermissionIds, overwrite bool) (*GlobalRole, error) {
	responseRole := &GlobalRole{
		Jenkins: j,
		Raw: GlobalRoleResponse{
			RoleName:      roleName,
			PermissionIds: ids,
		}}
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      roleName,
		"type":          GLOBAL_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(overwrite),
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseRole, nil
}

func (j *Jenkins) DeleteProjectRoles(roleName ...string) error {
	responseString := ""

	response, err := j.Requester.Post("/role-strategy/strategy/removeRoles", nil, &responseString, map[string]string{
		"type":      PROJECT_ROLE,
		"roleNames": strings.Join(roleName, ","),
	})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println(responseString)
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *Jenkins) AddProjectRole(roleName string, pattern string, ids ProjectPermissionIds, overwrite bool) (*ProjectRole, error) {
	responseRole := &ProjectRole{
		Jenkins: j,
		Raw: ProjectRoleResponse{
			RoleName:      roleName,
			PermissionIds: ids,
			Pattern:       pattern,
		}}
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      roleName,
		"type":          PROJECT_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(overwrite),
		"pattern":       pattern,
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseRole, nil
}

func (j *Jenkins) DeleteUserInProject(username string) error {
	param := map[string]string{
		"type": PROJECT_ROLE,
		"sid":  username,
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/deleteSid", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *Jenkins) GetPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.Pipeline, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetPipelineUrl, projectName, pipelineName),
	}
	res, err := PipelineOjb.GetPipeline()
	return res, err
}

func (j *Jenkins) ListPipelines(httpParameters *devops.HttpParameters) (*devops.PipelineList, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           ListPipelinesUrl + httpParameters.Url.RawQuery,
	}
	res, err := PipelineOjb.ListPipelines()
	return res, err
}

func (j *Jenkins) GetPipelineRun(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetPipelineRunUrl, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.GetPipelineRun()
	return res, err
}

func (j *Jenkins) ListPipelineRuns(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.PipelineRunList, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           ListPipelineRunUrl + httpParameters.Url.RawQuery,
	}
	res, err := PipelineOjb.ListPipelineRuns()
	return res, err
}

func (j *Jenkins) StopPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(StopPipelineUrl, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.StopPipeline()
	return res, err
}

func (j *Jenkins) ReplayPipeline(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(ReplayPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.ReplayPipeline()
	return res, err
}

func (j *Jenkins) RunPipeline(projectName, pipelineName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(RunPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName),
	}
	res, err := PipelineOjb.RunPipeline()
	return res, err
}

func (j *Jenkins) GetArtifacts(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.Artifacts, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetArtifactsUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.GetArtifacts()
	return res, err
}

func (j *Jenkins) GetRunLog(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetRunLogUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.GetRunLog()
	return res, err
}

func (j *Jenkins) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetStepLogUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId, nodeId, stepId),
	}
	res, header, err := PipelineOjb.GetStepLog()
	return res, header, err
}

func (j *Jenkins) GetNodeSteps(projectName, pipelineName, runId, nodeId string, httpParameters *devops.HttpParameters) (*devops.NodeSteps, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetRunLogUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.GetNodeSteps()
	return res, err
}

func (j *Jenkins) GetPipelineRunNodes(projectName, pipelineName, runId string, httpParameters *devops.HttpParameters) ([]devops.PipelineRunNodes, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetPipelineRunNodesUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId),
	}
	res, err := PipelineOjb.GetPipelineRunNodes()
	return res, err
}

func (j *Jenkins) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(SubmitInputStepUrl+httpParameters.Url.RawQuery, projectName, pipelineName, runId, nodeId, stepId),
	}
	res, err := PipelineOjb.SubmitInputStep()
	return res, err
}

func (j *Jenkins) GetBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.BranchPipeline, error) {
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchPipelineUrl, projectName, pipelineName, branchName),
	}
	res, err := PipelineOjb.GetBranchPipeline()
	return res, err
}

func (j *Jenkins) GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.PipelineRun, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchPipelineRunUrl, projectName, pipelineName, branchName, runId),
	}
	res, err := PipelineOjb.GetBranchPipelineRun()
	return res, err
}

func (j *Jenkins) StopBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.StopPipeline, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(StopBranchPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId),
	}
	res, err := PipelineOjb.StopBranchPipeline()
	return res, err
}

func (j *Jenkins) ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.ReplayPipeline, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(ReplayBranchPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId),
	}
	res, err := PipelineOjb.ReplayBranchPipeline()
	return res, err
}

func (j *Jenkins) RunBranchPipeline(projectName, pipelineName, branchName string, httpParameters *devops.HttpParameters) (*devops.RunPipeline, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(RunBranchPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName),
	}
	res, err := PipelineOjb.RunBranchPipeline()
	return res, err
}

func (j *Jenkins) GetBranchRunLog(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) ([]byte, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchRunLogUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId),
	}
	res, err := PipelineOjb.GetBranchRunLog()
	return res, err
}

func (j *Jenkins) GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchStepLogUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId),
	}
	res,header, err := PipelineOjb.GetBranchStepLog()
	return res, header,err
}

func (j *Jenkins) GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, httpParameters *devops.HttpParameters) (*devops.NodeSteps, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchNodeStepsUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId, nodeId),
	}
	res, err := PipelineOjb.GetBranchNodeSteps()
	return res, err
}

func (j *Jenkins) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, httpParameters *devops.HttpParameters) (*devops.BranchPipelineRunNodes, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(GetBranchPipeRunNodesUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId),
	}
	res, err := PipelineOjb.GetBranchPipelineRunNodes()
	return res, err
}

func (j *Jenkins) SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *devops.HttpParameters) ([]byte, error){
	PipelineOjb := &Pipeline{
		HttpParameters: httpParameters,
		Jenkins:        j,
		Path:           fmt.Sprintf(CheckBranchPipelineUrl+httpParameters.Url.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId),
	}
	res, err := PipelineOjb.SubmitBranchInputStep()
	return res, err
}

// Creates a new Jenkins Instance
// Optional parameters are: client, username, password
// After creating an instance call init method.
func CreateJenkins(client *http.Client, base string, maxConnection int, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: true, Client: client, connControl: make(chan struct{}, maxConnection)}
	if j.Requester.Client == nil {
		j.Requester.Client = http.DefaultClient
	}
	if len(auth) == 2 {
		j.Requester.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}
