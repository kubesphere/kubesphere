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
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"

	//"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"

	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

const (
	GroupName = "devops.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, devopsClient devops.Interface,
	dbClient *mysql.Database, sonarClient sonarqube.SonarInterface, ksClient versioned.Interface,
	ksInformer externalversions.SharedInformerFactory, s3Client s3.Interface) error {

	webservice := runtime.NewWebService(GroupVersion)

	sonarEnable := devopsClient != nil && dbClient != nil && sonarClient != nil
	projectPipleineEnable := devopsClient != nil && dbClient != nil
	s2iEnable := ksClient != nil && ksInformer != nil && s3Client != nil

	if sonarEnable {
		sonarHandler := NewPipelineSonarHandler(devopsClient, dbClient, sonarClient)
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/sonarstatus").
			To(sonarHandler.GetPipelineSonarStatusHandler).
			Doc("Get the sonar quality information for the specified pipeline of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
			Returns(http.StatusOK, RespOK, []sonarqube.SonarStatus{}).
			Writes([]sonarqube.SonarStatus{}))

		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/sonarstatus").
			To(sonarHandler.GetMultiBranchesPipelineSonarStatusHandler).
			Doc("Get the sonar quality check information for the specified pipeline branch of the DevOps project. More info: https://docs.sonarqube.org/7.4/user-guide/metric-definitions/").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of pipeline, e.g. sample-pipeline")).
			Param(webservice.PathParameter("branch", "branch name, e.g. master")).
			Returns(http.StatusOK, RespOK, []sonarqube.SonarStatus{}).
			Writes([]sonarqube.SonarStatus{}))
	}

	if projectPipleineEnable {
		projectPipelineHander := NewProjectPipelineHandler(devopsClient, dbClient)
		webservice.Route(webservice.GET("/devops/{devops}").
			To(projectPipelineHander.GetDevOpsProjectHandler).
			Doc("Get the specified DevOps Project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Returns(http.StatusOK, RespOK, v1alpha2.DevOpsProject{}).
			Writes(v1alpha2.DevOpsProject{}))

		webservice.Route(webservice.PATCH("/devops/{devops}").
			To(projectPipelineHander.UpdateProjectHandler).
			Doc("Update the specified DevOps Project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Reads(v1alpha2.DevOpsProject{}).
			Returns(http.StatusOK, RespOK, v1alpha2.DevOpsProject{}).
			Writes(v1alpha2.DevOpsProject{}))

		webservice.Route(webservice.GET("/devops/{devops}/defaultroles").
			To(GetDevOpsProjectDefaultRoles).
			Doc("Get the build-in roles info of the specified DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Returns(http.StatusOK, RespOK, []devops.Role{}).
			Writes([]devops.Role{}))

		webservice.Route(webservice.GET("/devops/{devops}/members").
			To(projectPipelineHander.GetDevOpsProjectMembersHandler).
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
			Returns(http.StatusOK, RespOK, []devops.ProjectMembership{}).
			Writes([]devops.ProjectMembership{}))

		webservice.Route(webservice.GET("/devops/{devops}/members/{member}").
			To(projectPipelineHander.GetDevOpsProjectMemberHandler).
			Doc("Get the specified member of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("member", "member's username, e.g. admin")).
			Returns(http.StatusOK, RespOK, devops.ProjectMembership{}).
			Writes(devops.ProjectMembership{}))

		webservice.Route(webservice.POST("/devops/{devops}/members").
			To(projectPipelineHander.AddDevOpsProjectMemberHandler).
			Doc("Add a member to the specified DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Returns(http.StatusOK, RespOK, devops.ProjectMembership{}).
			Writes(devops.ProjectMembership{}).
			Reads(devops.ProjectMembership{}))

		webservice.Route(webservice.PATCH("/devops/{devops}/members/{member}").
			To(projectPipelineHander.UpdateDevOpsProjectMemberHandler).
			Doc("Update the specified member of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("member", "member's username, e.g. admin")).
			Returns(http.StatusOK, RespOK, devops.ProjectMembership{}).
			Reads(devops.ProjectMembership{}).
			Writes(devops.ProjectMembership{}))

		webservice.Route(webservice.DELETE("/devops/{devops}/members/{member}").
			To(projectPipelineHander.DeleteDevOpsProjectMemberHandler).
			Doc("Delete the specified member of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("member", "member's username, e.g. admin")).
			Writes(devops.ProjectMembership{}))

		webservice.Route(webservice.POST("/devops/{devops}/credentials").
			To(projectPipelineHander.CreateDevOpsProjectCredentialHandler).
			Doc("Create a credential in the specified DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Reads(devops.Credential{}))

		webservice.Route(webservice.PUT("/devops/{devops}/credentials/{credential}").
			To(projectPipelineHander.UpdateDevOpsProjectCredentialHandler).
			Doc("Update the specified credential of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")).
			Reads(devops.Credential{}))

		webservice.Route(webservice.DELETE("/devops/{devops}/credentials/{credential}").
			To(projectPipelineHander.DeleteDevOpsProjectCredentialHandler).
			Doc("Delete the specified credential of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")))

		webservice.Route(webservice.GET("/devops/{devops}/credentials/{credential}").
			To(projectPipelineHander.GetDevOpsProjectCredentialHandler).
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
			Returns(http.StatusOK, RespOK, devops.Credential{}))

		webservice.Route(webservice.GET("/devops/{devops}/credentials").
			To(projectPipelineHander.GetDevOpsProjectCredentialsHandler).
			Doc("Get all credentials of the specified DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Returns(http.StatusOK, RespOK, []devops.Credential{}))

		// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}").
			To(projectPipelineHander.GetPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get the specified pipeline of the DevOps project").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Returns(http.StatusOK, RespOK, devops.Pipeline{}).
			Writes(devops.Pipeline{}))

		// match Jenkisn api: "jenkins_api/blue/rest/search"
		webservice.Route(webservice.GET("/search").
			To(projectPipelineHander.ListPipelines).
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
			Returns(http.StatusOK, RespOK, devops.PipelineList{}).
			Writes(devops.PipelineList{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}").
			To(projectPipelineHander.GetPipelineRun).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get details in the specified pipeline activity.").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
			Writes(devops.PipelineRun{}))

		// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs").
			To(projectPipelineHander.ListPipelineRuns).
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
			Returns(http.StatusOK, RespOK, devops.PipelineRunList{}).
			Writes(devops.PipelineRunList{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/stop/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/stop").
			To(projectPipelineHander.StopPipeline).
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
			Returns(http.StatusOK, RespOK, devops.StopPipeline{}).
			Writes(devops.StopPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/Replay/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/replay").
			To(projectPipelineHander.ReplayPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Replay pipeline").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.ReplayPipeline{}).
			Writes(devops.ReplayPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs").
			To(projectPipelineHander.RunPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Run pipeline.").
			Reads(devops.RunPayload{}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Returns(http.StatusOK, RespOK, devops.RunPipeline{}).
			Writes(devops.RunPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/artifacts
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/artifacts").
			To(projectPipelineHander.GetArtifacts).
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

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/log/?start=0
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/log").
			To(projectPipelineHander.GetRunLog).
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

		// match "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log/?start=0"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}/log").
			To(projectPipelineHander.GetStepLog).
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

		// match /blue/rest/organizations/jenkins/pipelines/%s/%s/runs/%s/nodes/%s/steps/?limit=
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps").
			To(projectPipelineHander.GetNodeSteps).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get all steps in the specified node.").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build")).
			Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
			Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
			Writes([]devops.NodeSteps{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/?limit=10000
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes").
			To(projectPipelineHander.GetPipelineRunNodes).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get all nodes in the specified activity. node is the stage in the pipeline task").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build")).
			Returns(http.StatusOK, RespOK, []devops.PipelineRunNodes{}).
			Writes([]devops.PipelineRunNodes{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}").
			To(projectPipelineHander.SubmitInputStep).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Proceed or Break the paused pipeline which is waiting for user input.").
			Reads(devops.CheckPlayload{}).
			Produces("text/plain; charset=utf-8").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
			Param(webservice.PathParameter("step", "pipeline step ID")))

		// out of scm get all steps in nodes.
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodesdetail").
			To(projectPipelineHander.GetNodesDetail).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get steps details inside a activity node. For a node, the steps which defined inside the node.").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
			Writes(devops.NodesDetail{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}").
			To(projectPipelineHander.GetBranchPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Get the specified branch pipeline of the DevOps project").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch")).
			Returns(http.StatusOK, RespOK, devops.BranchPipeline{}).
			Writes(devops.BranchPipeline{}))

		// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}").
			To(projectPipelineHander.GetBranchPipelineRun).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Get details in the specified pipeline activity.").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Param(webservice.PathParameter("run", "pipeline run id, the unique id for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
			Writes(devops.PipelineRun{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/stop/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/stop").
			To(projectPipelineHander.StopBranchPipeline).
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
			Returns(http.StatusOK, RespOK, devops.StopPipeline{}).
			Writes(devops.StopPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/Replay/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/replay").
			To(projectPipelineHander.ReplayBranchPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Replay the specified pipeline of the DevOps project").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.ReplayPipeline{}).
			Writes(devops.ReplayPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{}/runs/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs").
			To(projectPipelineHander.RunBranchPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Run the specified pipeline of the DevOps project.").
			Reads(devops.RunPayload{}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Returns(http.StatusOK, RespOK, devops.RunPipeline{}).
			Writes(devops.RunPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/artifacts
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/artifacts").
			To(projectPipelineHander.GetBranchArtifacts).
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

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/log/?start=0
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/log").
			To(projectPipelineHander.GetBranchRunLog).
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

		// match "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}/log/?start=0"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}/log").
			To(projectPipelineHander.GetBranchStepLog).
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

		// match /blue/rest/organizations/jenkins/pipelines/%s/%s/branches/%s/runs/%s/nodes/%s/steps/?limit=
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps").
			To(projectPipelineHander.GetBranchNodeSteps).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Get all steps in the specified node.").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Param(webservice.PathParameter("node", "pipeline node ID, the stage in pipeline.")).
			Returns(http.StatusOK, RespOK, []devops.NodeSteps{}).
			Writes([]devops.NodeSteps{}))

		// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/nodes"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes").
			To(projectPipelineHander.GetBranchPipelineRunNodes).
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

		// /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodes/{node}/steps/{step}").
			To(projectPipelineHander.SubmitBranchInputStep).
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

		// in scm get all steps in nodes.
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}/nodesdetail").
			To(projectPipelineHander.GetBranchNodesDetail).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Get steps details in an activity node. For a node, the steps which is defined inside the node.").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch.")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
			Writes(devops.NodesDetail{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/?filter=&start&limit=
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches").
			To(projectPipelineHander.GetPipelineBranch).
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
			Returns(http.StatusOK, RespOK, []devops.PipelineBranch{}).
			Writes([]devops.PipelineBranch{}))

		// match /job/{devops}/job/{pipeline}/build?delay=0
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/scan").
			To(projectPipelineHander.ScanBranch).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Scan remote Repository, Start a build if have new branch.").
			Produces("text/html; charset=utf-8").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.QueryParameter("delay", "the delay time to scan").
				Required(false).
				DataFormat("delay=%d")))

		// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/consolelog").
			To(projectPipelineHander.GetConsoleLog).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get scan reponsitory logs in the specified pipeline.").
			Produces("text/plain; charset=utf-8").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")))

		// match /crumbIssuer/api/json/
		webservice.Route(webservice.GET("/crumbissuer").
			To(projectPipelineHander.GetCrumb).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get crumb issuer. A CrumbIssuer represents an algorithm to generate a nonce value, known as a crumb, to counter cross site request forgery exploits. Crumbs are typically hashes incorporating information that uniquely identifies an agent that sends a request, along with a guarded secret so that the crumb value cannot be forged by a third party.").
			Returns(http.StatusOK, RespOK, devops.Crumb{}).
			Writes(devops.Crumb{}))

		// match "/blue/rest/organizations/jenkins/scm/%s/servers/"
		webservice.Route(webservice.GET("/scms/{scm}/servers").
			To(projectPipelineHander.GetSCMServers).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("List all servers in the jenkins.").
			Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
			Returns(http.StatusOK, RespOK, []devops.SCMServer{}).
			Writes([]devops.SCMServer{}))

		// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/?credentialId=github"
		webservice.Route(webservice.GET("/scms/{scm}/organizations").
			To(projectPipelineHander.GetSCMOrg).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("List all organizations of the specified source configuration management (SCM) such as Github.").
			Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
			Param(webservice.QueryParameter("credentialId", "credential ID for source configuration management (SCM).").
				Required(true).
				DataFormat("credentialId=%s")).
			Returns(http.StatusOK, RespOK, []devops.SCMOrg{}).
			Writes([]devops.SCMOrg{}))

		// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/{organization}/repositories/?credentialId=&pageNumber&pageSize="
		webservice.Route(webservice.GET("/scms/{scm}/organizations/{organization}/repositories").
			To(projectPipelineHander.GetOrgRepo).
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

		// match "/blue/rest/organizations/jenkins/scm/%s/servers/" create bitbucket server
		webservice.Route(webservice.POST("/scms/{scm}/servers").
			To(projectPipelineHander.CreateSCMServers).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("Create scm server in the jenkins.").
			Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
			Reads(devops.CreateScmServerReq{}).
			Returns(http.StatusOK, RespOK, devops.SCMServer{}).
			Writes(devops.SCMServer{}))

		// match "/blue/rest/organizations/jenkins/scm/github/validate/"
		webservice.Route(webservice.POST("/scms/{scm}/verify").
			To(projectPipelineHander.Validate).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("Validate the access token of the specified source configuration management (SCM) such as Github").
			Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
			Returns(http.StatusOK, RespOK, devops.Validates{}).
			Writes(devops.Validates{}))

		// match /git/notifyCommit/?url=
		webservice.Route(webservice.GET("/webhook/git").
			To(projectPipelineHander.GetNotifyCommit).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification by HTTP GET method. Git webhook will request here.").
			Produces("text/plain; charset=utf-8").
			Param(webservice.QueryParameter("url", "Git url").
				Required(true).
				DataFormat("url=%s")))

		// Gitlab or some other scm managers can only use HTTP method. match /git/notifyCommit/?url=
		webservice.Route(webservice.POST("/webhook/git").
			To(projectPipelineHander.PostNotifyCommit).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification by HTTP POST method. Git webhook will request here.").
			Consumes("application/json").
			Produces("text/plain; charset=utf-8").
			Param(webservice.QueryParameter("url", "Git url").
				Required(true).
				DataFormat("url=%s")))

		webservice.Route(webservice.POST("/webhook/github").
			To(projectPipelineHander.GithubWebhook).
			Consumes("application/x-www-form-urlencoded", "application/json").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification. Github webhook will request here."))

		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/checkScriptCompile").
			To(projectPipelineHander.CheckScriptCompile).
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
			To(projectPipelineHander.CheckCron).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Produces("application/json", "charset=utf-8").
			Doc("Check cron script compile.").
			Reads(devops.CronData{}).
			Returns(http.StatusOK, RespOK, devops.CheckCronRes{}).
			Writes(devops.CheckCronRes{}))

		// match /pipeline-model-converter/toJenkinsfile
		webservice.Route(webservice.POST("/tojenkinsfile").
			To(projectPipelineHander.ToJenkinsfile).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
			Consumes("application/x-www-form-urlencoded").
			Produces("application/json", "charset=utf-8").
			Doc("Convert json to jenkinsfile format.").
			Reads(devops.ReqJson{}).
			Returns(http.StatusOK, RespOK, devops.ResJenkinsfile{}).
			Writes(devops.ResJenkinsfile{}))

		// match /pipeline-model-converter/toJson
		webservice.Route(webservice.POST("/tojson").
			To(projectPipelineHander.ToJson).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
			Consumes("application/x-www-form-urlencoded").
			Produces("application/json", "charset=utf-8").
			Doc("Convert jenkinsfile to json format. Usually the frontend uses json to show or edit pipeline").
			Reads(devops.ReqJenkinsfile{}).
			Returns(http.StatusOK, RespOK, devops.ResJson{}).
			Writes(devops.ResJson{}))
	}

	if s2iEnable {
		s2iHandler := NewS2iBinaryHandler(ksClient, ksInformer, s3Client)
		webservice.Route(webservice.PUT("/namespaces/{namespace}/s2ibinaries/{s2ibinary}/file").
			To(s2iHandler.UploadS2iBinaryHandler).
			Consumes("multipart/form-data").
			Produces(restful.MIME_JSON).
			Doc("Upload S2iBinary file").
			Param(webservice.PathParameter("namespace", "the name of namespaces")).
			Param(webservice.PathParameter("s2ibinary", "the name of s2ibinary")).
			Param(webservice.FormParameter("s2ibinary", "file to upload")).
			Param(webservice.FormParameter("md5", "md5 of file")).
			Returns(http.StatusOK, RespOK, devopsv1alpha1.S2iBinary{}))

		webservice.Route(webservice.GET("/namespaces/{namespace}/s2ibinaries/{s2ibinary}/file/{file}").
			To(s2iHandler.DownloadS2iBinaryHandler).
			Produces(restful.MIME_OCTET).
			Doc("Download S2iBinary file").
			Param(webservice.PathParameter("namespace", "the name of namespaces")).
			Param(webservice.PathParameter("s2ibinary", "the name of s2ibinary")).
			Param(webservice.PathParameter("file", "the name of binary file")).
			Returns(http.StatusOK, RespOK, nil))
	}

	c.Add(webservice)

	return nil
}
