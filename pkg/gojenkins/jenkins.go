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
package gojenkins

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Raw       *ExecutorResponse
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

	// Check Connection
	j.Raw = new(ExecutorResponse)
	rsp, err := j.Requester.GetJSON("/", j.Raw, nil)
	if err != nil {
		return nil, err
	}

	j.Version = rsp.Header.Get("X-Jenkins")
	if j.Raw == nil {
		return nil, errors.New("Connection Failed, Please verify that the host and credentials are correct.")
	}

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

// Get Basic Information About Jenkins
func (j *Jenkins) Info() (*ExecutorResponse, error) {
	_, err := j.Requester.Get("/", j.Raw, nil)

	if err != nil {
		return nil, err
	}
	return j.Raw, nil
}

// Create a new Node
// Can be JNLPLauncher or SSHLauncher
// Example : jenkins.CreateNode("nodeName", 1, "Description", "/var/lib/jenkins", "jdk8 docker", map[string]string{"method": "JNLPLauncher"})
// By Default JNLPLauncher is created
// Multiple labels should be separated by blanks
func (j *Jenkins) CreateNode(name string, numExecutors int, description string, remoteFS string, label string, options ...interface{}) (*Node, error) {
	params := map[string]string{"method": "JNLPLauncher"}

	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}

	if _, ok := params["method"]; !ok {
		params["method"] = "JNLPLauncher"
	}

	method := params["method"]
	var launcher map[string]string
	switch method {
	case "":
		fallthrough
	case "JNLPLauncher":
		launcher = map[string]string{"stapler-class": "hudson.slaves.JNLPLauncher"}
	case "SSHLauncher":
		launcher = map[string]string{
			"stapler-class":        "hudson.plugins.sshslaves.SSHLauncher",
			"$class":               "hudson.plugins.sshslaves.SSHLauncher",
			"host":                 params["host"],
			"port":                 params["port"],
			"credentialsId":        params["credentialsId"],
			"jvmOptions":           params["jvmOptions"],
			"javaPath":             params["javaPath"],
			"prefixStartSlaveCmd":  params["prefixStartSlaveCmd"],
			"suffixStartSlaveCmd":  params["suffixStartSlaveCmd"],
			"maxNumRetries":        params["maxNumRetries"],
			"retryWaitTime":        params["retryWaitTime"],
			"lanuchTimeoutSeconds": params["lanuchTimeoutSeconds"],
			"type":                 "hudson.slaves.DumbSlave",
			"stapler-class-bag":    "true"}
	default:
		return nil, errors.New("launcher method not supported")
	}

	node := &Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
	NODE_TYPE := "hudson.slaves.DumbSlave$DescriptorImpl"
	MODE := "NORMAL"
	qr := map[string]string{
		"name": name,
		"type": NODE_TYPE,
		"json": makeJson(map[string]interface{}{
			"name":               name,
			"nodeDescription":    description,
			"remoteFS":           remoteFS,
			"numExecutors":       numExecutors,
			"mode":               MODE,
			"type":               NODE_TYPE,
			"labelString":        label,
			"retentionsStrategy": map[string]string{"stapler-class": "hudson.slaves.RetentionStrategy$Always"},
			"nodeProperties":     map[string]string{"stapler-class-bag": "true"},
			"launcher":           launcher,
		}),
	}

	resp, err := j.Requester.Post("/computer/doCreateItem", nil, nil, qr)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 400 {
		_, err := node.Poll()
		if err != nil {
			return nil, err
		}
		return node, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

// Delete a Jenkins slave node
func (j *Jenkins) DeleteNode(name string) (bool, error) {
	node := Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
	return node.Delete()
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
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
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
// First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
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
// First parameter job name, second parameter is optional Build parameters.
func (j *Jenkins) BuildJob(name string, options ...interface{}) (int64, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (j *Jenkins) GetNode(name string) (*Node, error) {
	node := Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
	status, err := node.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &node, nil
	}
	return nil, errors.New("No node found")
}

func (j *Jenkins) GetLabel(name string) (*Label, error) {
	label := Label{Jenkins: j, Raw: new(LabelResponse), Base: "/label/" + name}
	status, err := label.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &label, nil
	}
	return nil, errors.New("No label found")
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

func (j *Jenkins) GetSubJob(parentId string, childId string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + parentId + "/job/" + childId}
	status, err := job.Poll()
	if err != nil {
		return nil, fmt.Errorf("trouble polling job: %v", err)
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

func (j *Jenkins) GetAllNodes() ([]*Node, error) {
	computers := new(Computers)

	qr := map[string]string{
		"depth": "1",
	}

	_, err := j.Requester.GetJSON("/computer", computers, qr)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, len(computers.Computers))
	for i, node := range computers.Computers {
		nodes[i] = &Node{Jenkins: j, Raw: node, Base: "/computer/" + node.DisplayName}
	}

	return nodes, nil
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)
func (j *Jenkins) GetAllBuildIds(job string) ([]JobBuild, error) {
	jobObj, err := j.GetJob(job)
	if err != nil {
		return nil, err
	}
	return jobObj.GetAllBuildIds()
}

func (j *Jenkins) GetAllBuildStatus(jobId string) ([]JobBuildStatus, error) {
	job, err := j.GetJob(jobId)
	if err != nil {
		return nil, err
	}
	return job.GetAllBuildStatus()
}

// Get Only Array of Job Names, Color, URL
// Does not query each single Job.
func (j *Jenkins) GetAllJobNames() ([]InnerJob, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON("/", exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	return exec.Raw.Jobs, nil
}

// Get All Possible Job Objects.
// Each job will be queried.
func (j *Jenkins) GetAllJobs() ([]*Job, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON("/", exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		ji, err := j.GetJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

// Returns a Queue
func (j *Jenkins) GetQueue() (*Queue, error) {
	q := &Queue{Jenkins: j, Raw: new(queueResponse), Base: j.GetQueueUrl()}
	_, err := q.Poll()
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (j *Jenkins) GetQueueUrl() string {
	return "/queue"
}

// Get Artifact data by Hash
func (j *Jenkins) GetArtifactData(id string) (*FingerPrintResponse, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	return fp.GetInfo()
}

// Returns the list of all plugins installed on the Jenkins server.
// You can supply depth parameter, to limit how much data is returned.
func (j *Jenkins) GetPlugins(depth int) (*Plugins, error) {
	p := Plugins{Jenkins: j, Raw: new(PluginResponse), Base: "/pluginManager", Depth: depth}
	_, err := p.Poll()
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Check if the plugin is installed on the server.
// Depth level 1 is used. If you need to go deeper, you can use GetPlugins, and iterate through them.
func (j *Jenkins) HasPlugin(name string) (*Plugin, error) {
	p, err := j.GetPlugins(1)

	if err != nil {
		return nil, err
	}
	return p.Contains(name), nil
}

// Verify FingerPrint
func (j *Jenkins) ValidateFingerPrint(id string) (bool, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	valid, err := fp.Valid()
	if err != nil {
		return false, err
	}
	if valid {
		return true, nil
	}
	return false, nil
}

func (j *Jenkins) GetView(name string) (*View, error) {
	url := "/view/" + name
	view := View{Jenkins: j, Raw: new(ViewResponse), Base: url}
	_, err := view.Poll()
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (j *Jenkins) GetAllViews() ([]*View, error) {
	_, err := j.Poll()
	if err != nil {
		return nil, err
	}
	views := make([]*View, len(j.Raw.Views))
	for i, v := range j.Raw.Views {
		views[i], _ = j.GetView(v.Name)
	}
	return views, nil
}

// Create View
// First Parameter - name of the View
// Second parameter - Type
// Possible Types:
// 		gojenkins.LIST_VIEW
// 		gojenkins.NESTED_VIEW
// 		gojenkins.MY_VIEW
// 		gojenkins.DASHBOARD_VIEW
// 		gojenkins.PIPELINE_VIEW
// Example: jenkins.CreateView("newView",gojenkins.LIST_VIEW)
func (j *Jenkins) CreateView(name string, viewType string) (*View, error) {
	view := &View{Jenkins: j, Raw: new(ViewResponse), Base: "/view/" + name}
	endpoint := "/createView"
	data := map[string]string{
		"name":   name,
		"mode":   viewType,
		"Submit": "OK",
		"json": makeJson(map[string]string{
			"name": name,
			"mode": viewType,
		}),
	}
	r, err := j.Requester.Post(endpoint, nil, view.Raw, data)

	if err != nil {
		return nil, err
	}

	if r.StatusCode == 200 {
		return j.GetView(name)
	}
	return nil, errors.New(strconv.Itoa(r.StatusCode))
}

func (j *Jenkins) Poll() (int, error) {
	resp, err := j.Requester.GetJSON("/", j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

// Create a ssh credentials
// return credentials id
func (j *Jenkins) CreateSshCredential(id, username, passphrase, privateKey, description string) (*string, error) {
	requestStruct := NewCreateSshCredentialRequest(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	response, err := j.Requester.Post("/credentials/store/system/domain/_/createCredentials",
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateUsernamePasswordCredential(id, username, password, description string) (*string, error) {
	requestStruct := NewCreateUsernamePasswordRequest(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	response, err := j.Requester.Post("/credentials/store/system/domain/_/createCredentials",
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateSshCredentialRequest(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateUsernamePasswordRequest(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateSecretTextCredentialRequest(id, secret, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateKubeconfigCredentialRequest(id, content, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) UpdateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error) {
	requestStruct := NewSshCredential(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error) {
	requestStruct := NewUsernamePasswordCredential(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error) {
	requestStruct := NewSecretTextCredential(id, secret, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error) {
	requestStruct := NewKubeconfigCredential(id, content, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) GetCredentialInFolder(domain, id string, folders ...string) (*CredentialResponse, error) {
	responseStruct := &CredentialResponse{}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.GetJSON(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s", domain, id),
		responseStruct, map[string]string{
			"depth": "2",
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	responseStruct.Domain = domain
	return responseStruct, nil
}

func (j *Jenkins) GetCredentialContentInFolder(domain, id string, folders ...string) (string, error) {
	responseStruct := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return "", fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.GetHtml(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/update", domain, id),
		&responseStruct, nil)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseStruct, nil
}

func (j *Jenkins) GetCredentialsInFolder(domain string, folders ...string) ([]*CredentialResponse, error) {
	prePath := ""
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}

	if domain == "" {
		var responseStruct = &struct {
			Domains map[string]struct {
				Credentials []*CredentialResponse `json:"credentials"`
			} `json:"domains"`
		}{}
		response, err := j.Requester.GetJSON(prePath+
			"/credentials/store/folder/",
			responseStruct, map[string]string{
				"depth": "2",
			})
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			return nil, errors.New(strconv.Itoa(response.StatusCode))
		}
		responseArray := make([]*CredentialResponse, 0)
		for domainName, domain := range responseStruct.Domains {
			for _, credential := range domain.Credentials {
				credential.Domain = domainName
				responseArray = append(responseArray, credential)
			}
		}
		return responseArray, nil
	}

	var responseStruct = &struct {
		Credentials []*CredentialResponse `json:"credentials"`
	}{}
	response, err := j.Requester.GetJSON(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s", domain),
		responseStruct, map[string]string{
			"depth": "2",
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	for _, credential := range responseStruct.Credentials {
		credential.Domain = domain
	}
	return responseStruct.Credentials, nil

}

func (j *Jenkins) DeleteCredentialInFolder(domain, id string, folders ...string) (*string, error) {
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/doDelete", domain, id),
		nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
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

func (j *Jenkins) GetQueueItem(number int64) (*QueueItemResponse, error) {
	responseItem := &QueueItemResponse{}
	response, err := j.Requester.GetJSON(fmt.Sprintf("/queue/item/%s", strconv.FormatInt(number, 10)),
		responseItem, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseItem, nil
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
