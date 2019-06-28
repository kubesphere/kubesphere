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
		Doc("Get the specified DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProject{}).
		Writes(devops.DevOpsProject{}))

	webservice.Route(webservice.PATCH("/devops/{devops}").
		To(devopsapi.UpdateProjectHandler).
		Doc("Update the specified DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Reads(devops.DevOpsProject{}).
		Returns(http.StatusOK, RespOK, devops.DevOpsProject{}).
		Writes(devops.DevOpsProject{}))

	webservice.Route(webservice.GET("/devops/{devops}/defaultroles").
		To(devopsapi.GetDevOpsProjectDefaultRoles).
		Doc("Get the build-in roles info of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, []devops.Role{}).
		Writes([]devops.Role{}))

	webservice.Route(webservice.GET("/devops/{devops}/members").
		To(devopsapi.GetDevOpsProjectMembersHandler).
		Doc("Get the members of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, RespOK, []devops.DevOpsProjectMembership{}).
		Writes([]devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.GET("/devops/{devops}/members/{member}").
		To(devopsapi.GetDevOpsProjectMemberHandler).
		Doc("Get the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/members").
		To(devopsapi.AddDevOpsProjectMemberHandler).
		Doc("Add a member to the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.PATCH("/devops/{devops}/members/{member}").
		To(devopsapi.UpdateDevOpsProjectMemberHandler).
		Doc("Update the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Reads(devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/members/{member}").
		To(devopsapi.DeleteDevOpsProjectMemberHandler).
		Doc("Delete the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/pipelines").
		To(devopsapi.CreateDevOpsProjectPipelineHandler).
		Doc("Create a DevOps project pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.PUT("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.UpdateDevOpsProjectPipelineHandler).
		Doc("Update the specified pipeline of the DevOps project").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.DeleteDevOpsProjectPipelineHandler).
		Doc("Delete the specified pipeline of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/config").
		To(devopsapi.GetDevOpsProjectPipelineHandler).
		Doc("Get the configuration information of the specified pipeline of the DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/sonarstatus").
		To(devopsapi.GetPipelineSonarStatusHandler).
		Doc("Get the sonar quality information for the specified pipeline of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipelines}/branches/{branch}/sonarstatus").
		To(devopsapi.GetMultiBranchesPipelineSonarStatusHandler).
		Doc("Get the sonar quality check information for the specified pipeline branch of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipelines", "the name of pipeline, e.g. sample-pipeline")).
		Param(webservice.PathParameter("branch", "branch name, e.g. master")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.POST("/devops/{devops}/credentials").
		To(devopsapi.CreateDevOpsProjectCredentialHandler).
		Doc("Create a Credential in the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.PUT("/devops/{devops}/credentials/{credential}").
		To(devopsapi.UpdateDevOpsProjectCredentialHandler).
		Doc("Update the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/credentials/{credential}").
		To(devopsapi.DeleteDevOpsProjectCredentialHandler).
		Doc("Delete the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")))

	webservice.Route(webservice.GET("/devops/{devops}/credentials/{credential}").
		To(devopsapi.GetDevOpsProjectCredentialHandler).
		Doc("Get the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps Project's Id, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "Credential's Id, e.g. dockerhub-id")).
		Param(webservice.QueryParameter("content", `
Get extra content, if not none will get credential's content.
There are three main types of fields in the credential.
All credential fields, such as name, id, etc., will be returned each time.
Some credential non-encrypted fields, such as the username of the username-password type credential, which returns when the "content" parameter is set to non-empty.
Some encrypted fields, such as the password of the username password type credential, this part of the field will never return.
`)).
		Returns(http.StatusOK, RespOK, devops.JenkinsCredential{}).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.GET("/devops/{devops}/credentials").
		To(devopsapi.GetDevOpsProjectCredentialsHandler).
		Doc("Get all credentials of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, []devops.JenkinsCredential{}).
		Reads([]devops.JenkinsCredential{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.GetPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get the specified pipeline of the DevOps project").
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, devops.Pipeline{}).
		Writes(devops.Pipeline{}))

	// match Jenkisn api: "jenkins_api/blue/rest/search"
	webservice.Route(webservice.GET("/search").
		To(devopsapi.SearchPipelines).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Search DevOps resource.").
		Param(webservice.QueryParameter("q", "query pipelines, condition for filtering.").
			Required(false).
			DataFormat("q=%s")).
		Param(webservice.QueryParameter("filter", "Filter some types of jobs. e.g. no-folder，will not get a job of type folder").
			Required(false).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.Pipeline{}).
		Writes([]devops.Pipeline{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs").
		To(devopsapi.SearchPipelineRuns).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get all runs of the specified pipeline").
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search").
			Required(false).
			DataFormat("limit=%d")).
		Param(webservice.QueryParameter("branch", "the name of branch, same as repository branch, will be filtered by branch.").
			Required(false).
			DataFormat("branch=%s")).
		Returns(http.StatusOK, RespOK, []devops.BranchPipelineRun{}).
		Writes([]devops.BranchPipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}").
		To(devopsapi.GetBranchPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get all runs in a branch").
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(false).
			DataFormat("start=%d")).
		Returns(http.StatusOK, RespOK, devops.BranchPipelineRun{}).
		Writes(devops.BranchPipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/nodes"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes").
		To(devopsapi.GetPipelineRunNodesbyBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get run nodes.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d").
			DefaultValue("limit=10000")).
		Returns(http.StatusOK, RespOK, []devops.BranchPipelineRunNodes{}).
		Writes([]devops.BranchPipelineRunNodes{}))

	// match "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}/log").
		To(devopsapi.GetBranchStepLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get pipelines step log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step id, the one step in pipeline.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log").
		To(devopsapi.GetStepLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get pipelines step log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step id, the one step in pipeline.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/scm/github/validate/"
	webservice.Route(webservice.POST("/scms/{scm}/verify").
		To(devopsapi.Validate).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Validate the access token of the specified source configuration management (SCM) such as Github").
		Param(webservice.PathParameter("scm", "the id of the source configuration management (SCM).")).
		Returns(http.StatusOK, RespOK, devops.Validates{}).
		Writes(devops.Validates{}))

	// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/?credentialId=github"
	webservice.Route(webservice.GET("/scms/{scm}/organizations").
		To(devopsapi.GetSCMOrg).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("List all organizations of the specified source configuration management (SCM) such as Github.").
		Param(webservice.PathParameter("scm", "the id of the source configuration management (SCM).")).
		Param(webservice.QueryParameter("credentialId", "credential id for source configuration management (SCM).").
			Required(true).
			DataFormat("credentialId=%s")).
		Returns(http.StatusOK, RespOK, []devops.SCMOrg{}).
		Writes([]devops.SCMOrg{}))

	// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/{organization}/repositories/?credentialId=&pageNumber&pageSize="
	webservice.Route(webservice.GET("/scms/{scm}/organizations/{organization}/repositories").
		To(devopsapi.GetOrgRepo).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("List all repositories in the specified organization.").
		Param(webservice.PathParameter("scm", "The id of the source configuration management (SCM).")).
		Param(webservice.PathParameter("organization", "organization Id, such as github username.")).
		Param(webservice.QueryParameter("credentialId", "credential id for SCM.").
			Required(true).
			DataFormat("credentialId=%s")).
		Param(webservice.QueryParameter("pageNumber", "page number.").
			Required(true).
			DataFormat("pageNumber=%d")).
		Param(webservice.QueryParameter("pageSize", "the item count of one page.").
			Required(true).
			DataFormat("pageSize=%d")).
		Returns(http.StatusOK, RespOK, []devops.OrgRepo{}).
		Writes([]devops.OrgRepo{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/stop/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/stop").
		To(devopsapi.StopBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Stop pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("blocking", "stop and between each retries will sleep.").
			Required(false).
			DataFormat("blocking=%t").
			DefaultValue("blocking=false")).
		Param(webservice.QueryParameter("timeOutInSecs", "the time of stop and between each retries sleep.").
			Required(false).
			DataFormat("timeOutInSecs=%d").
			DefaultValue("timeOutInSecs=10")).
		Returns(http.StatusOK, RespOK, devops.StopPipe{}).
		Writes(devops.StopPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/stop/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/stop").
		To(devopsapi.StopPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Stop pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("blocking", "stop and between each retries will sleep.").
			Required(false).
			DataFormat("blocking=%t").
			DefaultValue("blocking=false")).
		Param(webservice.QueryParameter("timeOutInSecs", "the time of stop and between each retries sleep.").
			Required(false).
			DataFormat("timeOutInSecs=%d").
			DefaultValue("timeOutInSecs=10")).
		Returns(http.StatusOK, RespOK, devops.StopPipe{}).
		Writes(devops.StopPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/Replay/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/replay").
		To(devopsapi.ReplayBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Replay pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/Replay/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/replay").
		To(devopsapi.ReplayPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Replay pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/log/?start=0
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/log").
		To(devopsapi.GetBranchRunLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get run logs of the specified pipeline activity.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/log/?start=0
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/log").
		To(devopsapi.GetRunLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get run logs of the specified pipeline activity.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(true).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/artifacts
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/artifacts").
		To(devopsapi.GetBranchArtifacts).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get all artifacts in the specified pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, "The filed of \"Url\" in response can download artifacts", []devops.Artifacts{}).
		Writes([]devops.Artifacts{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/artifacts
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/artifacts").
		To(devopsapi.GetArtifacts).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get all artifacts in the specified pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number of the search starts from.").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, "The filed of \"Url\" in response can download artifacts", []devops.Artifacts{}).
		Writes([]devops.Artifacts{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/?filter=&start&limit=
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches").
		To(devopsapi.GetPipeBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get all branches in the specified pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.QueryParameter("filter", "filter remote scm. e.g. origin").
			Required(true).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "the count of branches start.").
			Required(true).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the count of branches limit.").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.PipeBranch{}).
		Writes([]devops.PipeBranch{}))

	// /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}").
		To(devopsapi.CheckBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Proceed or Break the paused pipeline which waiting for user input.").
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step id, the one step in pipeline.")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}").
		To(devopsapi.CheckPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Proceed or Break the paused pipeline which waiting for user input.").
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step id")))

	// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/consolelog").
		To(devopsapi.GetConsoleLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get scan reponsitory logs in the specified pipeline.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")))

	// match /job/{devops}/job/{pipeline}/build?delay=0
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/scan").
		To(devopsapi.ScanBranch).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Scan remote Repositorie, Start a build if have new branch.").
		Produces("text/html; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.QueryParameter("delay", "will be delay time to scan.").
			Required(false).
			DataFormat("delay=%d")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{}/runs/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/run").
		To(devopsapi.RunBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Run pipeline.").
		Reads(devops.RunPayload{}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Returns(http.StatusOK, RespOK, devops.QueuedBlueRun{}).
		Writes(devops.QueuedBlueRun{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/run").
		To(devopsapi.RunPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Run pipeline.").
		Reads(devops.RunPayload{}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Returns(http.StatusOK, RespOK, devops.QueuedBlueRun{}).
		Writes(devops.QueuedBlueRun{}))

	// match /crumbIssuer/api/json/
	webservice.Route(webservice.GET("/crumbissuer").
		To(devopsapi.GetCrumb).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get crumb issuer. A CrumbIssuer represents an algorithm to generate a nonce value, known as a crumb, to counter cross site request forgery exploits. Crumbs are typically hashes incorporating information that uniquely identifies an agent that sends a request, along with a guarded secret so that the crumb value cannot be forged by a third party.").
		Returns(http.StatusOK, RespOK, devops.Crumb{}).
		Writes(devops.Crumb{}))

	// TODO are not used in this version. will be added in 2.1.0
	//// match /job/init-job/descriptorByName/org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition/checkScriptCompile
	//webservice.Route(webservice.POST("/devops/check/scriptcompile").
	//	To(devopsapi.CheckScriptCompile).
	//	Metadata(restfulspec.KeyOpenAPITags, tags).
	//	Consumes("application/x-www-form-urlencoded", "charset=utf-8").
	//	Produces("application/json", "charset=utf-8").
	//	Doc("Check pipeline script compile.").
	//	Reads(devops.ReqScript{}).
	//	Returns(http.StatusOK, RespOK, devops.CheckScript{}).
	//	Writes(devops.CheckScript{}))

	// match /job/init-job/descriptorByName/hudson.triggers.TimerTrigger/checkSpec
	//webservice.Route(webservice.GET("/devops/check/cron").
	//	To(devopsapi.CheckCron).
	//	Metadata(restfulspec.KeyOpenAPITags, tags).
	//	Produces("application/json", "charset=utf-8").
	//	Doc("Check cron script compile.").
	//	Param(webservice.QueryParameter("value", "string of cron script.").
	//		Required(true).
	//		DataFormat("value=%s")).
	//	Returns(http.StatusOK, RespOK, []devops.QueuedBlueRun{}).
	//	Returns(http.StatusOK, RespOK, devops.CheckCronRes{}).
	//	Writes(devops.CheckCronRes{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}").
		To(devopsapi.GetPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get all activities in the specified pipeline.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
		Writes(devops.PipelineRun{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}").
		To(devopsapi.GetBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get all activities in the specified pipeline.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach")).
		Returns(http.StatusOK, RespOK, devops.BranchPipeline{}).
		Writes(devops.BranchPipeline{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/?limit=10000
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes").
		To(devopsapi.GetPipelineRunNodes).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get all nodes in the specified activity. node is the stage in the pipeline task").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.PipelineRunNodes{}).
		Writes([]devops.PipelineRunNodes{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps").
		To(devopsapi.GetBranchNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Get all steps in the specified node.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps").
		To(devopsapi.GetNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get all steps in the specified node.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build")).
		Param(webservice.PathParameter("node", "pipeline node id, the one node in pipeline.")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /pipeline-model-converter/toJenkinsfile
	webservice.Route(webservice.POST("/tojenkinsfile").
		To(devopsapi.ToJenkinsfile).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Convert json to jenkinsfile format.").
		Reads(devops.ReqJson{}).
		Returns(http.StatusOK, RespOK, devops.ResJenkinsfile{}).
		Writes(devops.ResJenkinsfile{}))

	// match /pipeline-model-converter/toJson
	webservice.Route(webservice.POST("/tojson").
		To(devopsapi.ToJson).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Convert jenkinsfile to json format. Usually the frontend uses json to show or edit pipeline").
		Reads(devops.ReqJenkinsfile{}).
		Returns(http.StatusOK, RespOK, devops.ResJson{}).
		Writes(devops.ResJson{}))

	// match /git/notifyCommit/?url=
	webservice.Route(webservice.GET("/webhook/git").
		To(devopsapi.GetNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get commit notification by HTTP GET method. Git webhook will request here.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "url of git scm").
			Required(true).
			DataFormat("url=%s")))

	// Gitlab or some other scm managers can only use HTTP method. match /git/notifyCommit/?url=
	webservice.Route(webservice.POST("/webhook/git").
		To(devopsapi.PostNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get commit notification by HTTP POST method. Git webhook will request here.").
		Consumes("application/json").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "url of git scm").
			Required(true).
			DataFormat("url=%s")))

	webservice.Route(webservice.POST("/webhook/github").
		To(devopsapi.GithubWebhook).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get commit notification. Github webhook will request here."))

	// in scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodesdetail").
		To(devopsapi.GetBranchNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("(MultiBranchesPipeline) Gives steps details inside a activity node. For a node, the steps which defined inside the node.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	// out of scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodesdetail").
		To(devopsapi.GetNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Gives steps details inside a activity node. For a node, the steps which defined inside the node.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, which helps to deliver continuous integration continuous deployment.")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository brnach.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(true).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	c.Add(webservice)

	return nil
}
