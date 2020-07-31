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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	resourcesV1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"sync"
)

const (
	channelMaxCapacity = 100
)

type DevopsOperator interface {
	CreateDevOpsProject(workspace string, project *v1alpha3.DevOpsProject) (*v1alpha3.DevOpsProject, error)
	GetDevOpsProject(workspace string, projectName string) (*v1alpha3.DevOpsProject, error)
	DeleteDevOpsProject(workspace string, projectName string) error
	UpdateDevOpsProject(workspace string, project *v1alpha3.DevOpsProject) (*v1alpha3.DevOpsProject, error)
	ListDevOpsProject(workspace string, limit, offset int) (api.ListResult, error)

	CreatePipelineObj(projectName string, pipeline *v1alpha3.Pipeline) (*v1alpha3.Pipeline, error)
	GetPipelineObj(projectName string, pipelineName string) (*v1alpha3.Pipeline, error)
	DeletePipelineObj(projectName string, pipelineName string) error
	UpdatePipelineObj(projectName string, pipeline *v1alpha3.Pipeline) (*v1alpha3.Pipeline, error)
	ListPipelineObj(projectName string, limit, offset int) (api.ListResult, error)

	CreateCredentialObj(projectName string, s *v1.Secret) (*v1.Secret, error)
	GetCredentialObj(projectName string, secretName string) (*v1.Secret, error)
	DeleteCredentialObj(projectName string, secretName string) error
	UpdateCredentialObj(projectName string, secret *v1.Secret) (*v1.Secret, error)
	ListCredentialObj(projectName string, query *query.Query) (api.ListResult, error)

	GetPipeline(projectName, pipelineName string, req *http.Request) (*devops.Pipeline, error)
	ListPipelines(req *http.Request) (*devops.PipelineList, error)
	GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) (*devops.PipelineRun, error)
	ListPipelineRuns(projectName, pipelineName string, req *http.Request) (*devops.PipelineRunList, error)
	StopPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.StopPipeline, error)
	ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.ReplayPipeline, error)
	RunPipeline(projectName, pipelineName string, req *http.Request) (*devops.RunPipeline, error)
	GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]devops.Artifacts, error)
	GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error)
	GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error)
	GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error)
	GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]devops.PipelineRunNodes, error)
	SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error)
	GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]devops.NodesDetail, error)

	GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.BranchPipeline, error)
	GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.PipelineRun, error)
	StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.StopPipeline, error)
	ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.ReplayPipeline, error)
	RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.RunPipeline, error)
	GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.Artifacts, error)
	GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error)
	GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error)
	GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error)
	GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.BranchPipelineRunNodes, error)
	SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error)
	GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.NodesDetail, error)
	GetPipelineBranch(projectName, pipelineName string, req *http.Request) (*devops.PipelineBranch, error)
	ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error)

	GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error)
	GetCrumb(req *http.Request) (*devops.Crumb, error)

	GetSCMServers(scmId string, req *http.Request) ([]devops.SCMServer, error)
	GetSCMOrg(scmId string, req *http.Request) ([]devops.SCMOrg, error)
	GetOrgRepo(scmId, organizationId string, req *http.Request) (devops.OrgRepo, error)
	CreateSCMServers(scmId string, req *http.Request) (*devops.SCMServer, error)
	Validate(scmId string, req *http.Request) (*devops.Validates, error)

	GetNotifyCommit(req *http.Request) ([]byte, error)
	GithubWebhook(req *http.Request) ([]byte, error)

	CheckScriptCompile(projectName, pipelineName string, req *http.Request) (*devops.CheckScript, error)
	CheckCron(projectName string, req *http.Request) (*devops.CheckCronRes, error)

	ToJenkinsfile(req *http.Request) (*devops.ResJenkinsfile, error)
	ToJson(req *http.Request) (*devops.ResJson, error)
}

type devopsOperator struct {
	devopsClient devops.Interface
	k8sclient    kubernetes.Interface
	ksclient     kubesphere.Interface
	ksInformers  externalversions.SharedInformerFactory
	k8sInformers informers.SharedInformerFactory
}

func NewDevopsOperator(client devops.Interface, k8sclient kubernetes.Interface, ksclient kubesphere.Interface,
	ksInformers externalversions.SharedInformerFactory, k8sInformers informers.SharedInformerFactory) DevopsOperator {
	return &devopsOperator{
		devopsClient: client,
		k8sclient:    k8sclient,
		ksclient:     ksclient,
		ksInformers:  ksInformers,
		k8sInformers: k8sInformers,
	}
}

func convertToHttpParameters(req *http.Request) *devops.HttpParameters {
	httpParameters := devops.HttpParameters{
		Method:   req.Method,
		Header:   req.Header,
		Body:     req.Body,
		Form:     req.Form,
		PostForm: req.PostForm,
		Url:      req.URL,
	}

	return &httpParameters
}

func (d devopsOperator) CreateDevOpsProject(workspace string, project *v1alpha3.DevOpsProject) (*v1alpha3.DevOpsProject, error) {
	// All resources of devops project belongs to the namespace of the same name
	// The devops project name is used as the name of the admin namespace, using generateName to avoid conflicts
	if project.GenerateName == "" {
		err := errors.NewInvalid(devopsv1alpha3.SchemeGroupVersion.WithKind(devopsv1alpha3.ResourceKindDevOpsProject).GroupKind(),
			"", []*field.Error{field.Required(field.NewPath("metadata.generateName"), "generateName is required")})
		klog.Error(err)
		return nil, err
	}
	// generateName is used as displayName
	// ensure generateName is unique in workspace scope
	if unique, err := d.isGenerateNameUnique(workspace, project.GenerateName); err != nil {
		return nil, err
	} else if !unique {
		err = errors.NewConflict(devopsv1alpha3.Resource(devopsv1alpha3.ResourceSingularDevOpsProject),
			project.GenerateName, fmt.Errorf(project.GenerateName, fmt.Errorf("a devops project named %s already exists in the workspace", project.GenerateName)))
		klog.Error(err)
		return nil, err
	}
	// metadata override
	if project.Labels == nil {
		project.Labels = make(map[string]string, 0)
	}
	project.Name = ""
	project.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	return d.ksclient.DevopsV1alpha3().DevOpsProjects().Create(project)
}

func (d devopsOperator) GetDevOpsProject(workspace string, projectName string) (*v1alpha3.DevOpsProject, error) {
	return d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
}

func (d devopsOperator) DeleteDevOpsProject(workspace string, projectName string) error {
	return d.ksclient.DevopsV1alpha3().DevOpsProjects().Delete(projectName, metav1.NewDeleteOptions(0))
}

func (d devopsOperator) UpdateDevOpsProject(workspace string, project *v1alpha3.DevOpsProject) (*v1alpha3.DevOpsProject, error) {
	if project.Labels == nil {
		project.Labels = make(map[string]string, 0)
	}
	project.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	return d.ksclient.DevopsV1alpha3().DevOpsProjects().Update(project)
}

func (d devopsOperator) ListDevOpsProject(workspace string, limit, offset int) (api.ListResult, error) {
	data, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().List(labels.SelectorFromValidatedSet(labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}))
	if err != nil {
		return api.ListResult{}, nil
	}
	items := make([]interface{}, 0)
	var result []interface{}
	for _, item := range data {
		result = append(result, *item)
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}
	items = result[offset : offset+limit]
	if items == nil {
		items = []interface{}{}
	}
	return api.ListResult{TotalItems: len(result), Items: items}, nil
}

// pipelineobj in crd
func (d devopsOperator) CreatePipelineObj(projectName string, pipeline *v1alpha3.Pipeline) (*v1alpha3.Pipeline, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.ksclient.DevopsV1alpha3().Pipelines(projectObj.Status.AdminNamespace).Create(pipeline)
}

func (d devopsOperator) GetPipelineObj(projectName string, pipelineName string) (*v1alpha3.Pipeline, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.ksInformers.Devops().V1alpha3().Pipelines().Lister().Pipelines(projectObj.Status.AdminNamespace).Get(pipelineName)
}

func (d devopsOperator) DeletePipelineObj(projectName string, pipelineName string) error {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return err
	}
	return d.ksclient.DevopsV1alpha3().Pipelines(projectObj.Status.AdminNamespace).Delete(pipelineName, metav1.NewDeleteOptions(0))
}

func (d devopsOperator) UpdatePipelineObj(projectName string, pipeline *v1alpha3.Pipeline) (*v1alpha3.Pipeline, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.ksclient.DevopsV1alpha3().Pipelines(projectObj.Status.AdminNamespace).Update(pipeline)
}

func (d devopsOperator) ListPipelineObj(projectName string, limit, offset int) (api.ListResult, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return api.ListResult{}, err
	}
	data, err := d.ksInformers.Devops().V1alpha3().Pipelines().Lister().Pipelines(projectObj.Status.AdminNamespace).List(labels.Everything())
	if err != nil {
		return api.ListResult{}, err
	}
	items := make([]interface{}, 0)
	var result []interface{}
	for _, item := range data {
		result = append(result, *item)
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}
	items = result[offset : offset+limit]
	if items == nil {
		items = []interface{}{}
	}
	return api.ListResult{TotalItems: len(result), Items: items}, nil
}

//credentialobj in crd
func (d devopsOperator) CreateCredentialObj(projectName string, secret *v1.Secret) (*v1.Secret, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.k8sclient.CoreV1().Secrets(projectObj.Status.AdminNamespace).Create(secret)
}

func (d devopsOperator) GetCredentialObj(projectName string, secretName string) (*v1.Secret, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.k8sInformers.Core().V1().Secrets().Lister().Secrets(projectObj.Status.AdminNamespace).Get(secretName)
}

func (d devopsOperator) DeleteCredentialObj(projectName string, secret string) error {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return err
	}
	return d.k8sclient.CoreV1().Secrets(projectObj.Status.AdminNamespace).Delete(secret, metav1.NewDeleteOptions(0))
}

func (d devopsOperator) UpdateCredentialObj(projectName string, secret *v1.Secret) (*v1.Secret, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return nil, err
	}
	return d.k8sclient.CoreV1().Secrets(projectObj.Status.AdminNamespace).Update(secret)
}

func (d devopsOperator) ListCredentialObj(projectName string, query *query.Query) (api.ListResult, error) {
	projectObj, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().Get(projectName)
	if err != nil {
		return api.ListResult{}, err
	}
	credentialObjList, err := d.k8sInformers.Core().V1().Secrets().Lister().Secrets(projectObj.Status.AdminNamespace).List(query.Selector())
	if err != nil {
		return api.ListResult{}, err
	}
	var result []runtime.Object

	credentialTypeList := []v1.SecretType{
		v1alpha3.SecretTypeBasicAuth,
		v1alpha3.SecretTypeSSHAuth,
		v1alpha3.SecretTypeSecretText,
		v1alpha3.SecretTypeKubeConfig,
	}
	for _, credential := range credentialObjList {
		for _, credentialType := range credentialTypeList {
			if credential.Type == credentialType {
				result = append(result, credential)
			}
		}
	}

	return *resourcesV1alpha3.DefaultList(result, query, d.compareCredentialObj, d.filterCredentialObj), nil
}

func (d devopsOperator) compareCredentialObj(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftObj, ok := left.(*v1.Secret)
	if !ok {
		return false
	}

	rightObj, ok := right.(*v1.Secret)
	if !ok {
		return false
	}

	return resourcesV1alpha3.DefaultObjectMetaCompare(leftObj.ObjectMeta, rightObj.ObjectMeta, field)
}

func (d devopsOperator) filterCredentialObj(object runtime.Object, filter query.Filter) bool {

	secret, ok := object.(*v1.Secret)

	if !ok {
		return false
	}

	return resourcesV1alpha3.DefaultObjectMetaFilter(secret.ObjectMeta, filter)
}

// others
func (d devopsOperator) GetPipeline(projectName, pipelineName string, req *http.Request) (*devops.Pipeline, error) {

	res, err := d.devopsClient.GetPipeline(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelines(req *http.Request) (*devops.PipelineList, error) {

	res, err := d.devopsClient.ListPipelines(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) (*devops.PipelineRun, error) {

	res, err := d.devopsClient.GetPipelineRun(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelineRuns(projectName, pipelineName string, req *http.Request) (*devops.PipelineRunList, error) {

	res, err := d.devopsClient.ListPipelineRuns(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) StopPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.StopPipeline, error) {

	req.Method = http.MethodPut
	res, err := d.devopsClient.StopPipeline(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.ReplayPipeline, error) {

	res, err := d.devopsClient.ReplayPipeline(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) RunPipeline(projectName, pipelineName string, req *http.Request) (*devops.RunPipeline, error) {

	res, err := d.devopsClient.RunPipeline(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]devops.Artifacts, error) {

	res, err := d.devopsClient.GetArtifacts(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GetRunLog(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {

	resBody, header, err := d.devopsClient.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func (d devopsOperator) GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error) {
	res, err := d.devopsClient.GetNodeSteps(projectName, pipelineName, runId, nodeId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]devops.PipelineRunNodes, error) {

	res, err := d.devopsClient.GetPipelineRunNodes(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	fmt.Println()

	return res, err
}

func (d devopsOperator) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody

	resBody, err := d.devopsClient.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]devops.NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []devops.NodesDetail
	stepChan := make(chan *devops.NodesStepsIndex, channelMaxCapacity)

	respNodes, err := d.GetPipelineRunNodes(projectName, pipelineName, runId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	Nodes, err := json.Marshal(respNodes)
	err = json.Unmarshal(Nodes, &nodesDetails)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range respNodes {
		wg.Add(1)
		go func(nodeId string, index int) {
			Steps, err := d.GetNodeSteps(projectName, pipelineName, runId, nodeId, req)
			if err != nil {
				klog.Error(err)
				return
			}

			stepChan <- &devops.NodesStepsIndex{Id: index, Steps: Steps}
			wg.Done()
		}(v.ID, i)
	}

	wg.Wait()
	close(stepChan)

	for oneNodeSteps := range stepChan {
		if oneNodeSteps != nil {
			nodesDetails[oneNodeSteps.Id].Steps = append(nodesDetails[oneNodeSteps.Id].Steps, oneNodeSteps.Steps...)
		}
	}

	return nodesDetails, err
}

func (d devopsOperator) GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.BranchPipeline, error) {

	res, err := d.devopsClient.GetBranchPipeline(projectName, pipelineName, branchName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.PipelineRun, error) {

	res, err := d.devopsClient.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.StopPipeline, error) {

	req.Method = http.MethodPut
	res, err := d.devopsClient.StopBranchPipeline(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.ReplayPipeline, error) {

	res, err := d.devopsClient.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.RunPipeline, error) {

	res, err := d.devopsClient.RunBranchPipeline(projectName, pipelineName, branchName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.Artifacts, error) {

	res, err := d.devopsClient.GetBranchArtifacts(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GetBranchRunLog(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {

	resBody, header, err := d.devopsClient.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func (d devopsOperator) GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error) {

	res, err := d.devopsClient.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.BranchPipelineRunNodes, error) {

	res, err := d.devopsClient.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody
	resBody, err := d.devopsClient.SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []devops.NodesDetail
	stepChan := make(chan *devops.NodesStepsIndex, channelMaxCapacity)

	respNodes, err := d.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	Nodes, err := json.Marshal(respNodes)
	err = json.Unmarshal(Nodes, &nodesDetails)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range nodesDetails {
		wg.Add(1)
		go func(nodeId string, index int) {
			Steps, err := d.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req)
			if err != nil {
				klog.Error(err)
				return
			}

			stepChan <- &devops.NodesStepsIndex{Id: index, Steps: Steps}
			wg.Done()
		}(v.ID, i)
	}

	wg.Wait()
	close(stepChan)

	for oneNodeSteps := range stepChan {
		if oneNodeSteps != nil {
			nodesDetails[oneNodeSteps.Id].Steps = append(nodesDetails[oneNodeSteps.Id].Steps, oneNodeSteps.Steps...)
		}
	}

	return nodesDetails, err
}

func (d devopsOperator) GetPipelineBranch(projectName, pipelineName string, req *http.Request) (*devops.PipelineBranch, error) {

	res, err := d.devopsClient.GetPipelineBranch(projectName, pipelineName, convertToHttpParameters(req))
	//baseUrl+req.URL.RawQuery, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {

	resBody, err := d.devopsClient.ScanBranch(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {

	resBody, err := d.devopsClient.GetConsoleLog(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetCrumb(req *http.Request) (*devops.Crumb, error) {

	res, err := d.devopsClient.GetCrumb(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetSCMServers(scmId string, req *http.Request) ([]devops.SCMServer, error) {

	req.Method = http.MethodGet
	resBody, err := d.devopsClient.GetSCMServers(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return resBody, err
}

func (d devopsOperator) GetSCMOrg(scmId string, req *http.Request) ([]devops.SCMOrg, error) {

	res, err := d.devopsClient.GetSCMOrg(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetOrgRepo(scmId, organizationId string, req *http.Request) (devops.OrgRepo, error) {

	res, err := d.devopsClient.GetOrgRepo(scmId, organizationId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return devops.OrgRepo{}, err
	}

	return res, err
}

func (d devopsOperator) CreateSCMServers(scmId string, req *http.Request) (*devops.SCMServer, error) {

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	createReq := &devops.CreateScmServerReq{}
	err = json.Unmarshal(requestBody, createReq)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = nil
	servers, err := d.GetSCMServers(scmId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, server := range servers {
		if server.ApiURL == createReq.ApiURL {
			return &server, nil
		}
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(requestBody))

	req.Method = http.MethodPost
	resBody, err := d.devopsClient.CreateSCMServers(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return resBody, err
}

func (d devopsOperator) Validate(scmId string, req *http.Request) (*devops.Validates, error) {

	req.Method = http.MethodPut
	resBody, err := d.devopsClient.Validate(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetNotifyCommit(req *http.Request) ([]byte, error) {

	req.Method = http.MethodGet

	res, err := d.devopsClient.GetNotifyCommit(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GithubWebhook(req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GithubWebhook(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) CheckScriptCompile(projectName, pipelineName string, req *http.Request) (*devops.CheckScript, error) {

	resBody, err := d.devopsClient.CheckScriptCompile(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) CheckCron(projectName string, req *http.Request) (*devops.CheckCronRes, error) {

	res, err := d.devopsClient.CheckCron(projectName, convertToHttpParameters(req))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ToJenkinsfile(req *http.Request) (*devops.ResJenkinsfile, error) {

	res, err := d.devopsClient.ToJenkinsfile(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ToJson(req *http.Request) (*devops.ResJson, error) {

	res, err := d.devopsClient.ToJson(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) isGenerateNameUnique(workspace, generateName string) (bool, error) {
	selector := labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}
	projects, err := d.ksInformers.Devops().V1alpha3().DevOpsProjects().Lister().List(labels.SelectorFromSet(selector))
	if err != nil {
		klog.Error(err)
		return false, err
	}
	for _, p := range projects {
		if p.GenerateName == generateName {
			return false, err
		}
	}
	return true, nil
}

func getInputReqBody(reqBody io.ReadCloser) (newReqBody io.ReadCloser, err error) {
	var checkBody devops.CheckPlayload
	var jsonBody []byte
	var workRound struct {
		ID         string                           `json:"id,omitempty" description:"id"`
		Parameters []devops.CheckPlayloadParameters `json:"parameters"`
		Abort      bool                             `json:"abort,omitempty" description:"abort or not"`
	}

	Body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	err = json.Unmarshal(Body, &checkBody)

	if checkBody.Abort != true && checkBody.Parameters == nil {
		workRound.Parameters = []devops.CheckPlayloadParameters{}
		workRound.ID = checkBody.ID
		jsonBody, _ = json.Marshal(workRound)
	} else {
		jsonBody, _ = json.Marshal(checkBody)
	}

	newReqBody = parseBody(bytes.NewBuffer(jsonBody))

	return newReqBody, nil

}

func parseBody(body io.Reader) (newReqBody io.ReadCloser) {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	return rc
}
