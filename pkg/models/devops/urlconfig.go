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

	GetPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/"
	SearchPipelineUrl      = "/blue/rest/search/?"
	GetPipelineRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/"
	SearchPipelineRunUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/?"
	StopPipelineUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/stop/?"
	ReplayPipelineUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/replay/"
	RunPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/"
	GetArtifactsUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/artifacts/?"
	GetRunLogUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/log/?"
	GetStepLogUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetNodeStepsUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/?"
	GetPipelineRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/?"
	CheckPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/"

	GetBranchPipeUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/"
	GetPipeBranchRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/"
	StopBranchPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/stop/?"
	ReplayBranchPipelineUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/replay/"
	RunBranchPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/"
	GetBranchArtifactsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/artifacts/?"
	GetBranchRunLogUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/log/?"
	GetBranchStepLogUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetBranchNodeStepsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/?"
	GetBranchPipeRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/?"
	CheckBranchPipelineUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/"










)
