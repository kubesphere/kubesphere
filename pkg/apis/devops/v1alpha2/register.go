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

package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	devopsapi "kubesphere.io/kubesphere/pkg/apiserver/devops"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/devops"

	"kubesphere.io/kubesphere/pkg/params"
	"net/http"
)

const (
	GroupName = "devops.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	webservice := runtime.NewWebService(GroupVersion)

	tags := []string{"DevOps"}

	webservice.Route(webservice.GET("/devops/{devops}").
		To(devopsapi.GetDevOpsProjectHandler).
		Doc("get devops project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProject{}).
		Writes(devops.DevOpsProject{}))

	webservice.Route(webservice.PATCH("/devops/{devops}").
		To(devopsapi.UpdateProjectHandler).
		Doc("get devops project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProject{}).
		Writes(devops.DevOpsProject{}))

	webservice.Route(webservice.GET("/devops/{devops}/defaultroles").
		To(devopsapi.GetDevOpsProjectDefaultRoles).
		Doc("get devops project defaultroles").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Returns(http.StatusOK, RespOK, []devops.Role{}).
		Writes([]devops.Role{}))

	webservice.Route(webservice.GET("/devops/{devops}/members").
		To(devopsapi.GetDevOpsProjectMembersHandler).
		Doc("get devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, RespOK, []devops.DevOpsProjectMembership{}).
		Writes([]devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.GET("/devops/{devops}/members/{members}").
		To(devopsapi.GetDevOpsProjectMemberHandler).
		Doc("get devops project member").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("members", "member's username")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/members").
		To(devopsapi.AddDevOpsProjectMemberHandler).
		Doc("add devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.PATCH("/devops/{devops}/members/{members}").
		To(devopsapi.UpdateDevOpsProjectMemberHandler).
		Doc("update devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("members", "member's username")).
		Reads(devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/members/{members}").
		To(devopsapi.DeleteDevOpsProjectMemberHandler).
		Doc("delete devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("members", "member's username")).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/pipelines").
		To(devopsapi.CreateDevOpsProjectPipelineHandler).
		Doc("add devops project pipeline").
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.PUT("/devops/{devops}/pipelines/{pipelines}").
		To(devopsapi.UpdateDevOpsProjectPipelineHandler).
		Doc("update devops project pipeline").
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("pipelines", "pipeline name")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipelines}/config").
		To(devopsapi.GetDevOpsProjectPipelineHandler).
		Doc("get devops project pipeline config").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("pipelines", "pipeline name")).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipelines}/sonarStatus").
		To(devopsapi.GetPipelineSonarStatusHandler).
		Doc("get devops project pipeline sonarStatus").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("pipelines", "pipeline name")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipelines}/branches/{branches}/sonarStatus").
		To(devopsapi.GetMultiBranchesPipelineSonarStatusHandler).
		Doc("get devops project pipeline sonarStatus").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("pipelines", "pipeline name")).
		Param(webservice.PathParameter("branches", "branch name")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/pipelines/{pipelines}").
		To(devopsapi.DeleteDevOpsProjectPipelineHandler).
		Doc("delete devops project pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("pipelines", "pipeline name")))

	webservice.Route(webservice.PUT("/devops/{devops}/pipelines").
		To(devopsapi.CreateDevOpsProjectPipelineHandler).
		Doc("update devops project pipeline").
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.POST("/devops/{devops}/credentials").
		To(devopsapi.CreateDevOpsProjectCredentialHandler).
		Doc("add project credential pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.PUT("/devops/{devops}/credentials/{credentials}").
		To(devopsapi.UpdateDevOpsProjectCredentialHandler).
		Doc("update project credential pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("credentials", "credential's Id")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/credentials/{credentials}").
		To(devopsapi.DeleteDevOpsProjectCredentialHandler).
		Doc("delete project credential pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("credentials", "credential's Id")))

	webservice.Route(webservice.GET("/devops/{devops}/credentials/{credentials}").
		To(devopsapi.GetDevOpsProjectCredentialHandler).
		Doc("get project credential pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("credentials", "credential's Id")).
		Param(webservice.QueryParameter("domain", "credential's domain")).
		Param(webservice.QueryParameter("content", "get additional content")).
		Returns(http.StatusOK, RespOK, devops.JenkinsCredential{}).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.GET("/devops/{devops}/credentials").
		To(devopsapi.GetDevOpsProjectCredentialsHandler).
		Doc("get project credential pipeline").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "devops project's Id")).
		Param(webservice.PathParameter("credentials", "credential's Id")).
		Param(webservice.QueryParameter("domain", "credential's domain")).
		Returns(http.StatusOK, RespOK, []devops.JenkinsCredential{}).
		Reads([]devops.JenkinsCredential{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}").
		To(devopsapi.GetPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Returns(http.StatusOK, RespOK, devops.Pipeline{}).
		Writes(devops.Pipeline{}))

	// match Jenkisn api: "jenkins_api/blue/rest/search"
	webservice.Route(webservice.GET("/devops/search").
		To(devopsapi.SearchPipelines).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Search DevOps resource.").
		Param(webservice.QueryParameter("q", "query pipelines").
			Required(false).
			DataFormat("q=%s")).
		Param(webservice.QueryParameter("filter", "filter resource").
			Required(false).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "start page").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.Pipeline{}).
		Writes([]devops.Pipeline{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs").
		To(devopsapi.SearchPipelineRuns).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Search DevOps Pipelines runs in branch.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.QueryParameter("start", "start page").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Param(webservice.QueryParameter("branch", "branch ").
			Required(false).
			DataFormat("branch=%s")).
		Returns(http.StatusOK, RespOK, []devops.BranchPipelineRun{}).
		Writes([]devops.BranchPipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}").
		To(devopsapi.GetBranchPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines run in branch.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start").
			Required(false).
			DataFormat("start=%d")).
		Returns(http.StatusOK, RespOK, devops.BranchPipelineRun{}).
		Writes(devops.BranchPipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes").
		To(devopsapi.GetPipelineRunNodesbyBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get node on DevOps Pipelines run.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("limit", "limit").
			Required(false).
			DataFormat("limit=%d").
			DefaultValue("limit=10000")).
		Returns(http.StatusOK, RespOK, []devops.BranchPipelineRunNodes{}).
		Writes([]devops.BranchPipelineRunNodes{}))

	// match "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log").
		To(devopsapi.GetBranchStepLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipelines step log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.PathParameter("nodeId", "pipeline runs node id")).
		Param(webservice.PathParameter("stepId", "pipeline runs step id")).
		Param(webservice.QueryParameter("start", "start").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log").
		To(devopsapi.GetStepLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipelines step log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.PathParameter("nodeId", "pipeline runs node id")).
		Param(webservice.PathParameter("stepId", "pipeline runs step id")).
		Param(webservice.QueryParameter("start", "start").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/scm/github/validate/"
	webservice.Route(webservice.PUT("/devops/scm/{scmId}/validate").
		To(devopsapi.Validate).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Validate Github personal access token.").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Returns(http.StatusOK, RespOK, devops.Validates{}).
		Writes(devops.Validates{}))

	// match "/blue/rest/organizations/jenkins/scm/{scmId}/organizations/?credentialId=github"
	webservice.Route(webservice.GET("/devops/scm/{scmId}/organizations").
		To(devopsapi.GetSCMOrg).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("List organizations of SCM").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Param(webservice.QueryParameter("credentialId", "credential id for SCM").
			Required(true).
			DataFormat("credentialId=%s")).
		Returns(http.StatusOK, RespOK, []devops.SCMOrg{}).
		Writes([]devops.SCMOrg{}))

	// match "/blue/rest/organizations/jenkins/scm/{scmId}/organizations/{organizationId}/repositories/?credentialId=&pageNumber&pageSize="
	webservice.Route(webservice.GET("/devops/scm/{scmId}/organizations/{organizationId}/repositories").
		To(devopsapi.GetOrgRepo).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get SCM repositories in an organization").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Param(webservice.PathParameter("organizationId", "organization Id, such as github username")).
		Param(webservice.QueryParameter("credentialId", "credential id for SCM").
			Required(true).
			DataFormat("credentialId=%s")).
		Param(webservice.QueryParameter("pageNumber", "page number").
			Required(true).
			DataFormat("pageNumber=%d")).
		Param(webservice.QueryParameter("pageSize", "page size").
			Required(true).
			DataFormat("pageSize=%d")).
		Returns(http.StatusOK, RespOK, []devops.OrgRepo{}).
		Writes([]devops.OrgRepo{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/stop/
	webservice.Route(webservice.PUT("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/stop").
		To(devopsapi.StopBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Stop pipeline in running").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("blocking", "stop and between each retries will sleep").
			Required(false).
			DataFormat("blocking=%t").
			DefaultValue("blocking=false")).
		Param(webservice.QueryParameter("timeOutInSecs", "the time of stop and between each retries sleep").
			Required(false).
			DataFormat("timeOutInSecs=%d").
			DefaultValue("timeOutInSecs=10")).
		Returns(http.StatusOK, RespOK, devops.StopPipe{}).
		Writes(devops.StopPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/runs/{runId}/stop/
	webservice.Route(webservice.PUT("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/stop").
		To(devopsapi.StopPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Stop pipeline in running").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("blocking", "stop and between each retries will sleep").
			Required(false).
			DataFormat("blocking=%t").
			DefaultValue("blocking=false")).
		Param(webservice.QueryParameter("timeOutInSecs", "the time of stop and between each retries sleep").
			Required(false).
			DataFormat("timeOutInSecs=%d").
			DefaultValue("timeOutInSecs=10")).
		Returns(http.StatusOK, RespOK, devops.StopPipe{}).
		Writes(devops.StopPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/Replay/
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/replay").
		To(devopsapi.ReplayBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Replay pipeline").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/runs/{runId}/Replay/
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/replay").
		To(devopsapi.ReplayPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Replay pipeline").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/log/?start=0
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/log").
		To(devopsapi.GetBranchRunLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Pipelines run log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/{runId}/log/?start=0
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/log").
		To(devopsapi.GetRunLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Pipelines run log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/artifacts
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/artifacts").
		To(devopsapi.GetBranchArtifacts).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline artifacts.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start page").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, "The filed of \"Url\" in response can download artifacts", []devops.Artifacts{}).
		Writes([]devops.Artifacts{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/{runId}/artifacts
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/artifacts").
		To(devopsapi.GetArtifacts).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline artifacts.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start page").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, "The filed of \"Url\" in response can download artifacts", []devops.Artifacts{}).
		Writes([]devops.Artifacts{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/?filter=&start&limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches").
		To(devopsapi.GetPipeBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline of branch.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.QueryParameter("filter", "filter remote").
			Required(true).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "start").
			Required(true).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.PipeBranch{}).
		Writes([]devops.PipeBranch{}))

	// /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}").
		To(devopsapi.CheckBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Pauses pipeline execution and allows the user to interact and control the flow of the build.").
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.PathParameter("stepId", "pipeline step id")))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}").
		To(devopsapi.CheckPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Pauses pipeline execution and allows the user to interact and control the flow of the build.").
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.PathParameter("stepId", "pipeline step id")))

	// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/console/log").
		To(devopsapi.GetConsoleLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get index console log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")))

	// match /job/{projectName}/job/{pipelineName}/build?delay=0
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/scan").
		To(devopsapi.ScanBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Start a build.").
		Produces("text/html; charset=utf-8").
		Param(webservice.PathParameter("projecFtName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.QueryParameter("delay", "delay time").
			Required(true).
			DataFormat("delay=%d")))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{}/runs/
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/run").
		To(devopsapi.RunPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline artifacts.").
		Reads(devops.RunPayload{}).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Returns(http.StatusOK, RespOK, devops.QueuedBlueRun{}).
		Writes(devops.QueuedBlueRun{}))

	// match /pipeline_status/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/?limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/status").
		To(devopsapi.GetBranchStepsStatus).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline steps status.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline run name")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.QueuedBlueRun{}).
		Writes([]devops.QueuedBlueRun{}))

	// match /pipeline_status/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/{runId}/nodes/?limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps/status").
		To(devopsapi.GetStepsStatus).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline steps status.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline run name")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.QueuedBlueRun{}).
		Writes([]devops.QueuedBlueRun{}))

	// match /crumbIssuer/api/json/
	webservice.Route(webservice.GET("/devops/crumbissuer").
		To(devopsapi.GetCrumb).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get crumb").
		Returns(http.StatusOK, RespOK, devops.Crumb{}).
		Writes(devops.Crumb{}))

	// match /job/init-job/descriptorByName/org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition/checkScriptCompile
	webservice.Route(webservice.POST("/devops/check/scriptcompile").
		To(devopsapi.CheckScriptCompile).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Consumes("application/x-www-form-urlencoded", "charset=utf-8").
		Produces("application/json", "charset=utf-8").
		Doc("Check pipeline script compile.").
		Reads("value=\"\"").
		Returns(http.StatusOK, RespOK, devops.CheckScript{}).
		Writes(devops.CheckScript{}))

	// match /job/init-job/descriptorByName/hudson.triggers.TimerTrigger/checkSpec
	webservice.Route(webservice.GET("/devops/check/cron").
		To(devopsapi.CheckCron).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces("application/json", "charset=utf-8").
		Doc("Check pipeline script compile.").
		Param(webservice.QueryParameter("value", "cpec value").
			Required(true).
			DataFormat("value=%s")).
		Returns(http.StatusOK, RespOK, []devops.QueuedBlueRun{}).
		Returns(http.StatusOK, RespOK, devops.CheckCronRes{}).
		Writes(devops.CheckCronRes{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/{runId}/
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}").
		To(devopsapi.GetPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get run pipeline in project.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline run id")).
		Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
		Writes(devops.PipelineRun{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}").
		To(devopsapi.GetBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Pipeline run in branch.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Returns(http.StatusOK, RespOK, devops.BranchPipeline{}).
		Writes(devops.BranchPipeline{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/?limit=10000
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes").
		To(devopsapi.GetPipelineRunNodes).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Pipeline run nodes.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline run id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.PipelineRunNodes{}).
		Writes([]devops.PipelineRunNodes{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps").
		To(devopsapi.GetBranchNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get steps in node.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline run id")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodes/{nodeId}/steps").
		To(devopsapi.GetNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get steps in node.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("runId", "pipeline run id")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK,[]devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /pipeline-model-converter/toJenkinsfile
	webservice.Route(webservice.POST("/devops/tojenkinsfile").
		To(devopsapi.ToJenkinsfile).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Json to Jenkinsfile.").
		Reads(devops.ReqJson{}).
		Returns(http.StatusOK, RespOK, devops.NodeSteps{}).
		Writes(devops.ResJenkinsfile{}))

	// match /pipeline-model-converter/toJson
	webservice.Route(webservice.POST("/devops/tojson").
		To(devopsapi.ToJson).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Jenkinsfile to Json.").
		Reads(devops.ReqJenkinsfile{}).
		Returns(http.StatusOK, RespOK, devops.ResJson{}).
		Writes(devops.ResJson{}))

	// match /git/notifyCommit/?url=
	webservice.Route(webservice.GET("/devops/notifycommit").
		To(devopsapi.GetNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Notify Commit by GET HTTP method.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "the url for webhook to push.").
			Required(true).
			DataFormat("url=%s")))

	// Gitlab or some other scm managers can only use HTTP method. match /git/notifyCommit/?url=
	webservice.Route(webservice.POST("/devops/notifycommit").
		To(devopsapi.GetNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Notify Commit by POST HTTP method.").
		Consumes("application/json").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "the url for webhook to push.").
			Required(true).
			DataFormat("url=%s")))

	webservice.Route(webservice.POST("/devops/github/webhook").
		To(devopsapi.GithubWebhook).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("receive webhook request."))

	// in scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodesdetail").
		To(devopsapi.GetBranchNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline nodes stages detail").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK,[]devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	// out of scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs/{runId}/nodesdetail").
		To(devopsapi.GetNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline nodes stages detail").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK,[]devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	c.Add(webservice)

	return nil
}
