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
	GetPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/%s/"
	SearchPipelineUrl      = "/blue/rest/search/?"
	SearchPipelineRunUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/%s/runs/?"
	GetPipelineRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/"
	GetPipelineRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/?"
	GetStepLogUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/%s/log/?"
	ValidateUrl            = "/blue/rest/organizations/jenkins/scm/%s/validate"
	GetSCMOrgUrl           = "/blue/rest/organizations/jenkins/scm/%s/organizations/?"
	GetOrgRepoUrl          = "/blue/rest/organizations/jenkins/scm/%s/organizations/%s/repositories/?"
	StopPipelineUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/stop/?"
	ReplayPipelineUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/replay/"
	GetRunLogUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/log/?"
	GetArtifactsUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/artifacts/?"
	GetPipeBranchUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/?"
	GetConsoleLogUrl       = "/job/%s/job/%s/indexing/consoleText"
	ScanBranchUrl          = "/job/%s/job/%s/build?"
	RunPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/"
	GetStepsStatusUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/?"
	CheckPipelineUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/"
	GetCrumbUrl            = "/crumbIssuer/api/json/"
)
