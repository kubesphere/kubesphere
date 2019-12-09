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
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsapi "kubesphere.io/kubesphere/pkg/apiserver/devops"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"

	"kubesphere.io/kubesphere/pkg/server/params"
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

	webservice.Route(webservice.GET("/devops/{devops}").
		To(devopsapi.GetDevOpsProjectHandler).
		Doc("Get the specified DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, v1alpha2.DevOpsProject{}).
		Writes(v1alpha2.DevOpsProject{}))

	webservice.Route(webservice.PATCH("/devops/{devops}").
		To(devopsapi.UpdateProjectHandler).
		Doc("Update the specified DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Reads(v1alpha2.DevOpsProject{}).
		Returns(http.StatusOK, RespOK, v1alpha2.DevOpsProject{}).
		Writes(v1alpha2.DevOpsProject{}))

	webservice.Route(webservice.GET("/devops/{devops}/defaultroles").
		To(devopsapi.GetDevOpsProjectDefaultRoles).
		Doc("Get the build-in roles info of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, []devops.Role{}).
		Writes([]devops.Role{}))

	webservice.Route(webservice.GET("/devops/{devops}/members").
		To(devopsapi.GetDevOpsProjectMembersHandler).
		Doc("Get the members of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, support using key-value pairs separated by comma to search, like 'conditions:somekey=somevalue,anotherkey=anothervalue'").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, RespOK, []devops.DevOpsProjectMembership{}).
		Writes([]devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.GET("/devops/{devops}/members/{member}").
		To(devopsapi.GetDevOpsProjectMemberHandler).
		Doc("Get the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/members").
		To(devopsapi.AddDevOpsProjectMemberHandler).
		Doc("Add a member to the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}).
		Reads(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.PATCH("/devops/{devops}/members/{member}").
		To(devopsapi.UpdateDevOpsProjectMemberHandler).
		Doc("Update the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Returns(http.StatusOK, RespOK, devops.DevOpsProjectMembership{}).
		Reads(devops.DevOpsProjectMembership{}).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/members/{member}").
		To(devopsapi.DeleteDevOpsProjectMemberHandler).
		Doc("Delete the specified member of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("member", "member's username, e.g. admin")).
		Writes(devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/pipelines").
		To(devopsapi.CreateDevOpsProjectPipelineHandler).
		Doc("Create a DevOps project pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.PUT("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.UpdateDevOpsProjectPipelineHandler).
		Doc("Update the specified pipeline of the DevOps project").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Writes(devops.ProjectPipeline{}).
		Reads(devops.ProjectPipeline{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.DeleteDevOpsProjectPipelineHandler).
		Doc("Delete the specified pipeline of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/config").
		To(devopsapi.GetDevOpsProjectPipelineHandler).
		Doc("Get the configuration information of the specified pipeline of the DevOps Project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Returns(http.StatusOK, RespOK, devops.ProjectPipeline{}).
		Writes(devops.ProjectPipeline{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/sonarstatus").
		To(devopsapi.GetPipelineSonarStatusHandler).
		Doc("Get the sonar quality information for the specified pipeline of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/sonarstatus").
		To(devopsapi.GetMultiBranchesPipelineSonarStatusHandler).
		Doc("Get the sonar quality check information for the specified pipeline branch of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
		Param(webservice.PathParameter("branch", "branch name, e.g. master")).
		Returns(http.StatusOK, RespOK, []devops.SonarStatus{}).
		Writes([]devops.SonarStatus{}))

	webservice.Route(webservice.POST("/devops/{devops}/credentials").
		To(devopsapi.CreateDevOpsProjectCredentialHandler).
		Doc("Create a credential in the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.PUT("/devops/{devops}/credentials/{credential}").
		To(devopsapi.UpdateDevOpsProjectCredentialHandler).
		Doc("Update the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")).
		Reads(devops.JenkinsCredential{}))

	webservice.Route(webservice.DELETE("/devops/{devops}/credentials/{credential}").
		To(devopsapi.DeleteDevOpsProjectCredentialHandler).
		Doc("Delete the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")))

	webservice.Route(webservice.GET("/devops/{devops}/credentials/{credential}").
		To(devopsapi.GetDevOpsProjectCredentialHandler).
		Doc("Get the specified credential of the DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")).
		Param(webservice.QueryParameter("content", `
Get extra credential content if this query parameter is set. 
Specifically, there are three types of info in a credential. One is the basic info that must be returned for each query such as name, id, etc.
The second one is non-encrypted info such as the username of the username-password type of credential, which returns when the "content" parameter is set to non-empty.
The last one is encrypted info, such as the password of the username-password type of credential, which never returns.
`)).
		Returns(http.StatusOK, RespOK, devops.JenkinsCredential{}))

	webservice.Route(webservice.GET("/devops/{devops}/credentials").
		To(devopsapi.GetDevOpsProjectCredentialsHandler).
		Doc("Get all credentials of the specified DevOps project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Returns(http.StatusOK, RespOK, []devops.JenkinsCredential{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}").
		To(devopsapi.GetPipeline).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get the specified pipeline of the DevOps project").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Returns(http.StatusOK, RespOK, devops.Pipeline{}).
		Writes(devops.Pipeline{}))

	// match Jenkisn api: "jenkins_api/blue/rest/search"
	webservice.Route(webservice.GET("/search").
		To(devopsapi.SearchPipelines).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Search DevOps resource. More info: https://github.com/jenkinsci/blueocean-plugin/tree/master/blueocean-rest#get-pipelines-across-organization").
		Param(webservice.QueryParameter("q", "query pipelines, condition for filtering.").
			Required(true).
			DataFormat("q=%s")).
		Param(webservice.QueryParameter("filter", "Filter some types of jobs. e.g. no-folder，will not get a job of type folder").
			Required(false).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search.").
			Required(false).
			DataFormat("limit=%d")).
		Returns(http.StatusOK, RespOK, struct {
			Items []devops.Pipeline `json:"items"`
			Total int               `json:"total_count"`
		}{}).
		Writes(struct {
			Items []devops.Pipeline `json:"items"`
			Total int               `json:"total_count"`
		}{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs").
		To(devopsapi.SearchPipelineRuns).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get all runs of the specified pipeline").
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "the limit item count of the search").
			Required(false).
			DataFormat("limit=%d")).
		Param(webservice.QueryParameter("branch", "the name of branch, same as repository branch, will be filtered by branch.").
			Required(false).
			DataFormat("branch=%s")).
		Returns(http.StatusOK, RespOK, struct {
			Items []devops.BranchPipelineRun `json:"items"`
			Total int                        `json:"total_count"`
		}{}).
		Writes(struct {
			Items []devops.BranchPipelineRun `json:"items"`
			Total int                        `json:"total_count"`
		}{}).
		Writes(struct {
			Items []devops.BranchPipelineRun `json:"items"`
			Total int                        `json:"total_count"`
		}{}).
		Writes(struct {
			Items []devops.BranchPipelineRun `json:"items"`
			Total int                        `json:"total_count"`
		}{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}").
		To(devopsapi.GetBranchPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get all runs in the specified branch").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.BranchPipelineRun{}).
		Writes(devops.BranchPipelineRun{}))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/nodes"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes").
		To(devopsapi.GetPipelineRunNodesbyBranch).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get run nodes.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get the step logs in the specified pipeline activity.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node id, the stage in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step id, the step in pipeline.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
			Required(false).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log").
		To(devopsapi.GetStepLog).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get pipelines step log.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step ID, the step in pipeline.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
			Required(false).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match "/blue/rest/organizations/jenkins/scm/github/validate/"
	webservice.Route(webservice.POST("/scms/{scm}/verify").
		To(devopsapi.Validate).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
		Doc("Validate the access token of the specified source configuration management (SCM) such as Github").
		Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
		Returns(http.StatusOK, RespOK, devops.Validates{}).
		Writes(devops.Validates{}))

	// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/?credentialId=github"
	webservice.Route(webservice.GET("/scms/{scm}/organizations").
		To(devopsapi.GetSCMOrg).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
		Doc("List all organizations of the specified source configuration management (SCM) such as Github.").
		Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
		Param(webservice.QueryParameter("credentialId", "credential ID for source configuration management (SCM).").
			Required(true).
			DataFormat("credentialId=%s")).
		Returns(http.StatusOK, RespOK, []devops.SCMOrg{}).
		Writes([]devops.SCMOrg{}))

	// match "/blue/rest/organizations/jenkins/scm/%s/servers/"
	webservice.Route(webservice.GET("/scms/{scm}/servers").
		To(devopsapi.GetSCMServers).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
		Doc("List all servers in the jenkins.").
		Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
		Returns(http.StatusOK, RespOK, []devops.SCMServer{}).
		Writes([]devops.SCMServer{}))

	// match "/blue/rest/organizations/jenkins/scm/%s/servers/"
	webservice.Route(webservice.POST("/scms/{scm}/servers").
		To(devopsapi.CreateSCMServers).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
		Doc("Create scm server in the jenkins.").
		Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
		Reads(devops.CreateScmServerReq{}).
		Returns(http.StatusOK, RespOK, devops.SCMServer{}).
		Writes(devops.SCMServer{}))

	// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/{organization}/repositories/?credentialId=&pageNumber&pageSize="
	webservice.Route(webservice.GET("/scms/{scm}/organizations/{organization}/repositories").
		To(devopsapi.GetOrgRepo).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
		Doc("List all repositories in the specified organization.").
		Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
		Param(webservice.PathParameter("organization", "organization ID, such as github username.")).
		Param(webservice.QueryParameter("credentialId", "credential ID for SCM.").
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Stop the specified pipeline of the DevOps project.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Stop pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Replay the specified pipeline of the DevOps project").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/Replay/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/replay").
		To(devopsapi.ReplayPipeline).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Replay pipeline").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.ReplayPipe{}).
		Writes(devops.ReplayPipe{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/log/?start=0
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/log").
		To(devopsapi.GetBranchRunLog).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get run logs of the specified pipeline activity.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
			Required(false).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/log/?start=0
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/log").
		To(devopsapi.GetRunLog).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get run logs of the specified pipeline activity.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
			Required(false).
			DataFormat("start=%d").
			DefaultValue("start=0")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/artifacts
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/artifacts").
		To(devopsapi.GetBranchArtifacts).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get all artifacts generated from the specified run of the pipeline branch.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get all artifacts in the specified pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.QueryParameter("start", "the item number that the search starts from.").
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
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get all branches in the specified pipeline.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.QueryParameter("filter", "filter remote scm. e.g. origin").
			Required(false).
			DataFormat("filter=%s")).
		Param(webservice.QueryParameter("start", "the count of branches start.").
			Required(false).
			DataFormat("start=%d").DefaultValue("start=0")).
		Param(webservice.QueryParameter("limit", "the count of branches limit.").
			Required(false).
			DataFormat("limit=%d").DefaultValue("limit=100")).
		Returns(http.StatusOK, RespOK, []devops.PipeBranch{}).
		Writes([]devops.PipeBranch{}))

	// /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}").
		To(devopsapi.SubmitBranchInputStep).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Proceed or Break the paused pipeline which waiting for user input.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step ID, the step in pipeline.")).
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8"))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}").
		To(devopsapi.SubmitInputStep).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Proceed or Break the paused pipeline which is waiting for user input.").
		Reads(devops.CheckPlayload{}).
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
		Param(webservice.PathParameter("step", "pipeline step ID")))

	// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/consolelog").
		To(devopsapi.GetConsoleLog).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get scan reponsitory logs in the specified pipeline.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")))

	// match /job/{devops}/job/{pipeline}/build?delay=0
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/scan").
		To(devopsapi.ScanBranch).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Scan remote Repository, Start a build if have new branch.").
		Produces("text/html; charset=utf-8").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.QueryParameter("delay", "the delay time to scan").
			Required(false).
			DataFormat("delay=%d")))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{}/runs/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs").
		To(devopsapi.RunBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Run the specified pipeline of the DevOps project.").
		Reads(devops.RunPayload{}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Returns(http.StatusOK, RespOK, devops.QueuedBlueRun{}).
		Writes(devops.QueuedBlueRun{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/
	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs").
		To(devopsapi.RunPipeline).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Run pipeline.").
		Reads(devops.RunPayload{}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Returns(http.StatusOK, RespOK, devops.QueuedBlueRun{}).
		Writes(devops.QueuedBlueRun{}))

	// match /crumbIssuer/api/json/
	webservice.Route(webservice.GET("/crumbissuer").
		To(devopsapi.GetCrumb).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get crumb issuer. A CrumbIssuer represents an algorithm to generate a nonce value, known as a crumb, to counter cross site request forgery exploits. Crumbs are typically hashes incorporating information that uniquely identifies an agent that sends a request, along with a guarded secret so that the crumb value cannot be forged by a third party.").
		Returns(http.StatusOK, RespOK, devops.Crumb{}).
		Writes(devops.Crumb{}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/s2ibinaries/{s2ibinary}/file").
		To(devopsapi.UploadS2iBinary).
		Consumes("multipart/form-data").
		Produces(restful.MIME_JSON).
		Doc("Upload S2iBinary file").
		Param(webservice.PathParameter("namespace", "the name of namespaces")).
		Param(webservice.PathParameter("s2ibinary", "the name of s2ibinary")).
		Param(webservice.FormParameter("s2ibinary", "file to upload")).
		Param(webservice.FormParameter("md5", "md5 of file")).
		Returns(http.StatusOK, RespOK, devopsv1alpha1.S2iBinary{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/s2ibinaries/{s2ibinary}/file/{file}").
		To(devopsapi.DownloadS2iBinary).
		Produces(restful.MIME_OCTET).
		Doc("Download S2iBinary file").
		Param(webservice.PathParameter("namespace", "the name of namespaces")).
		Param(webservice.PathParameter("s2ibinary", "the name of s2ibinary")).
		Param(webservice.PathParameter("file", "the name of binary file")).
		Returns(http.StatusOK, RespOK, nil))

	webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/checkScriptCompile").
		To(devopsapi.CheckScriptCompile).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.QueryParameter("pipeline", "the name of the CI/CD pipeline").
			Required(false).
			DataFormat("pipeline=%s")).
		Consumes("application/x-www-form-urlencoded", "charset=utf-8").
		Produces("application/json", "charset=utf-8").
		Doc("Check pipeline script compile.").
		Reads(devops.ReqScript{}).
		Returns(http.StatusOK, RespOK, devops.CheckScript{}).
		Writes(devops.CheckScript{}))

	webservice.Route(webservice.POST("/devops/{devops}/checkCron").
		To(devopsapi.CheckCron).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Produces("application/json", "charset=utf-8").
		Doc("Check cron script compile.").
		Reads(devops.CronData{}).
		Returns(http.StatusOK, RespOK, devops.CheckCronRes{}).
		Writes(devops.CheckCronRes{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}").
		To(devopsapi.GetPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get all activities in the specified pipeline.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
		Writes(devops.PipelineRun{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}").
		To(devopsapi.GetBranchPipeline).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get all activities in the specified pipeline.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch")).
		Returns(http.StatusOK, RespOK, devops.BranchPipeline{}).
		Writes(devops.BranchPipeline{}))

	// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/?limit=10000
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes").
		To(devopsapi.GetPipelineRunNodes).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get all nodes in the specified activity. node is the stage in the pipeline task").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build")).
		Returns(http.StatusOK, RespOK, []devops.PipelineRunNodes{}).
		Writes([]devops.PipelineRunNodes{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps").
		To(devopsapi.GetBranchNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get all steps in the specified node.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
		Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /blue/rest/organizations/jenkins/pipelines/%s/%s/runs/%s/nodes/%s/steps/?limit=
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps").
		To(devopsapi.GetNodeSteps).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get all steps in the specified node.").
		Param(webservice.PathParameter("devops", "the name of devops project")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build")).
		Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
		Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
		Writes([]devops.NodeSteps{}))

	// match /pipeline-model-converter/toJenkinsfile
	webservice.Route(webservice.POST("/tojenkinsfile").
		To(devopsapi.ToJenkinsfile).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Convert json to jenkinsfile format.").
		Reads(devops.ReqJson{}).
		Returns(http.StatusOK, RespOK, devops.ResJenkinsfile{}).
		Writes(devops.ResJenkinsfile{}))

	// match /pipeline-model-converter/toJson
	webservice.Route(webservice.POST("/tojson").
		To(devopsapi.ToJson).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
		Consumes("application/x-www-form-urlencoded").
		Produces("application/json", "charset=utf-8").
		Doc("Convert jenkinsfile to json format. Usually the frontend uses json to show or edit pipeline").
		Reads(devops.ReqJenkinsfile{}).
		Returns(http.StatusOK, RespOK, devops.ResJson{}).
		Writes(devops.ResJson{}))

	// match /git/notifyCommit/?url=
	webservice.Route(webservice.GET("/webhook/git").
		To(devopsapi.GetNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
		Doc("Get commit notification by HTTP GET method. Git webhook will request here.").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "Git url").
			Required(true).
			DataFormat("url=%s")))

	// Gitlab or some other scm managers can only use HTTP method. match /git/notifyCommit/?url=
	webservice.Route(webservice.POST("/webhook/git").
		To(devopsapi.PostNotifyCommit).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
		Doc("Get commit notification by HTTP POST method. Git webhook will request here.").
		Consumes("application/json").
		Produces("text/plain; charset=utf-8").
		Param(webservice.QueryParameter("url", "Git url").
			Required(true).
			DataFormat("url=%s")))

	webservice.Route(webservice.POST("/webhook/github").
		To(devopsapi.GithubWebhook).
		Consumes("application/x-www-form-urlencoded", "application/json").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
		Doc("Get commit notification. Github webhook will request here."))

	// in scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodesdetail").
		To(devopsapi.GetBranchNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("(MultiBranchesPipeline) Get steps details in an activity node. For a node, the steps which is defined inside the node.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	// out of scm get all steps in nodes.
	webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodesdetail").
		To(devopsapi.GetNodesDetail).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
		Doc("Get steps details inside a activity node. For a node, the steps which defined inside the node.").
		Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
		Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
		Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
		Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
		Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
		Writes(devops.NodesDetail{}))

	c.Add(webservice)

	return nil
}
