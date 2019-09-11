/*

  Copyright 2019 The KubeSphere Authors.

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

// Some apis for Jenkins.
const (
	GetPipeBranchUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/?"
	GetBranchPipeUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/"
	GetPipelineUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/"
	SearchPipelineUrl        = "/blue/rest/search/?"
	RunBranchPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/"
	RunPipelineUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/"
	GetPipelineRunUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/"
	GetPipeBranchRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/"
	SearchPipelineRunUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/?"
	GetBranchPipeRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/?"
	GetPipeRunNodesUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/?"
	GetBranchRunLogUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/log/?"
	GetRunLogUrl             = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/log/?"
	GetBranchStepLogUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetStepLogUrl            = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/log/?"
	StopBranchPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/stop/?"
	StopPipelineUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/stop/?"
	ReplayBranchPipelineUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/replay/"
	ReplayPipelineUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/replay/"
	GetBranchArtifactsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/artifacts/?"
	GetArtifactsUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/artifacts/?"
	CheckBranchPipelineUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/"
	CheckPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/"
	GetBranchNodeStepsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/?"
	GetNodeStepsUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/?"
	GetSCMServersUrl         = "/blue/rest/organizations/jenkins/scm/%s/servers/"
	CreateSCMServersUrl      = "/blue/rest/organizations/jenkins/scm/%s/servers/"
	ValidateUrl              = "/blue/rest/organizations/jenkins/scm/%s/validate"
	GetSCMOrgUrl             = "/blue/rest/organizations/jenkins/scm/%s/organizations/?"
	GetOrgRepoUrl            = "/blue/rest/organizations/jenkins/scm/%s/organizations/%s/repositories/?"
	GetConsoleLogUrl         = "/job/%s/job/%s/indexing/consoleText"
	ScanBranchUrl            = "/job/%s/job/%s/build?"
	GetCrumbUrl              = "/crumbIssuer/api/json/"
	ToJenkinsfileUrl         = "/pipeline-model-converter/toJenkinsfile"
	ToJsonUrl                = "/pipeline-model-converter/toJson"
	GetNotifyCommitUrl       = "/git/notifyCommit/?"
	GithubWebhookUrl         = "/github-webhook/"
	CheckScriptCompileUrl    = "/job/%s/job/%s/descriptorByName/org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition/checkScriptCompile"
	CheckPipelienCronUrl     = "/job/%s/job/%s/descriptorByName/hudson.triggers.TimerTrigger/checkSpec?value=%s"
	CheckCronUrl             = "/job/%s/descriptorByName/hudson.triggers.TimerTrigger/checkSpec?value=%s"
)
