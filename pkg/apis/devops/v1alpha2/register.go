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
	"kubesphere.io/kubesphere/pkg/apiserver/devops"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	jktype "kubesphere.io/kubesphere/pkg/models/devops"
	"net/http"
)

const (
	GroupName   = "devops.kubesphere.io"
	RespMessage = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	webservice := runtime.NewWebService(GroupVersion)

	tags := []string{"devops"}

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}"
	// Check
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}").
		To(devops.GetPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Returns(http.StatusOK, RespMessage, jktype.Pipeline{}))

	// match Jenkisn api: "jenkins_api/blue/rest/search"
	webservice.Route(webservice.GET("/devops/search").
		To(devops.SearchPipelines).
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
		Returns(http.StatusOK, RespMessage, []jktype.Pipeline{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs").
		To(devops.SearchPipelineRuns).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Search DevOps Pipelines runs.").
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
		Returns(http.StatusOK, RespMessage, []jktype.PipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}").
		To(devops.GetPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines run.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("start", "start").
			Required(false).
			DataFormat("start=%d")).
		Returns(http.StatusOK, RespMessage, jktype.PipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes").
		To(devops.GetPipelineRunNodes).
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
		Returns(http.StatusOK, RespMessage, []jktype.Nodes{}))

	// match "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log").
		To(devops.GetStepLog).
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

	// match "/blue/rest/organizations/jenkins/scm/github/validate/"
	webservice.Route(webservice.PUT("/devops/scm/{scmId}/validate").
		To(devops.Validate).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Validate Github personal access token.").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Returns(http.StatusOK, RespMessage, jktype.Validates{}))

	// match "/blue/rest/organizations/jenkins/scm/{scmId}/organizations/?credentialId=github"
	webservice.Route(webservice.GET("/devops/scm/{scmId}/organizations").
		To(devops.GetSCMOrg).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("List organizations of SCM").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Param(webservice.QueryParameter("credentialId", "credential id for SCM").
			Required(true).
			DataFormat("credentialId=%s")).
		Returns(http.StatusOK, RespMessage, []jktype.SCMOrg{}))

	// match "/blue/rest/organizations/jenkins/scm/{scmId}/organizations/{organizationId}/repositories/?credentialId=&pageNumber&pageSize="
	webservice.Route(webservice.GET("/devops/scm/{scmId}/organizations/{organizationId}/repositories").
		To(devops.GetSCMOrgRepo).
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
		Returns(http.StatusOK, RespMessage, []jktype.OrgRepo{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/stop/
	webservice.Route(webservice.PUT("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/stop").
		To(devops.StopPipeline).
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
		Returns(http.StatusOK, RespMessage, jktype.StopPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/replay/
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/replay").
		To(devops.ReplayPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Replay pipeline").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Returns(http.StatusOK, RespMessage, jktype.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/log/?start=0
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/log").
		To(devops.GetRunLog).
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

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/artifacts
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/artifacts").
		To(devops.GetArtifacts).
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
		Returns(http.StatusOK, "The filed of \"Url\" in response can download artifacts", []jktype.Artifacts{}))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/?filter=&start&limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches").
		To(devops.GetPipeBranch).
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
		Returns(http.StatusOK, "", []jktype.PipeBranch{}))

	// /blue/rest/organizations/jenkins/pipelines/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}").
		To(devops.CheckPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Pauses pipeline execution and allows the user to interact and control the flow of the build.").
		Reads(jktype.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.PathParameter("nodeId", "pipeline node id")).
		Param(webservice.PathParameter("stepId", "pipeline step id")))

	// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/console/log").
		To(devops.GetConsoleLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get index console log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")))

	// match /job/{projectName}/job/{pipelineName}/build?delay=0
	webservice.Route(webservice.POST("/devops/{projectName}/pipelines/{pipelineName}/scan").
		To(devops.ScanBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Start a build.").
		Produces("text/html; charset=utf-8").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.QueryParameter("delay", "delay time").
			Required(true).
			DataFormat("delay=%d")))

	// match /blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{}/runs/
	webservice.Route(webservice.POST("/devops/{projectName}/pipeline/{pipelineName}/branches/{brancheName}/run").
		To(devops.RunPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipeline artifacts.").
		Reads(jktype.RunPayload{}).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Returns(http.StatusOK, "", jktype.QueuedBlueRun{}))

	// match /pipeline_status/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/?limit=
	webservice.Route(webservice.GET("/devops/{projectName}/pipeline/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/status").
		To(devops.GetStepsStatus).
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
		Returns(http.StatusOK, "", []jktype.QueuedBlueRun{}))

	// match /crumbIssuer/api/json/
	webservice.Route(webservice.GET("/devops/crumbIssuer").
		To(devops.GetCrumb).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get crumb").
		Returns(http.StatusOK, "", jktype.Crumb{}))

	c.Add(webservice)

	return nil
}
