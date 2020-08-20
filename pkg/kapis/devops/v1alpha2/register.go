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

package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/klog"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"net/url"
	"strings"

	//"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

const (
	GroupName = "devops.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, ksInformers externalversions.SharedInformerFactory, devopsClient devops.Interface, sonarqubeClient sonarqube.SonarInterface, ksClient versioned.Interface, s3Client s3.Interface, endpoint string) error {
	ws := runtime.NewWebService(GroupVersion)

	err := AddPipelineToWebService(ws, devopsClient)
	if err != nil {
		return err
	}

	err = AddSonarToWebService(ws, devopsClient, sonarqubeClient)
	if err != nil {
		return err
	}

	err = AddS2IToWebService(ws, ksClient, ksInformers, s3Client)
	if err != nil {
		return err
	}

	err = AddJenkinsToContainer(ws, devopsClient, endpoint)
	if err != nil {
		return err
	}

	container.Add(ws)

	return nil
}

func AddPipelineToWebService(webservice *restful.WebService, devopsClient devops.Interface) error {

	projectPipelineEnable := devopsClient != nil

	if projectPipelineEnable {
		projectPipelineHandler := NewProjectPipelineHandler(devopsClient)

		webservice.Route(webservice.GET("/devops/{devops}/credentials/{credential}/usage").
			To(projectPipelineHandler.GetProjectCredentialUsage).
			Doc("Get the specified credential usage of the DevOps project").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectCredentialTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("credential", "credential's ID, e.g. dockerhub-id")).
			Returns(http.StatusOK, RespOK, devops.Credential{}))

		// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}").
			To(projectPipelineHandler.GetPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get the specified pipeline of the DevOps project").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Returns(http.StatusOK, RespOK, devops.Pipeline{}).
			Writes(devops.Pipeline{}))

		// match Jenkins api: "jenkins_api/blue/rest/search"
		webservice.Route(webservice.GET("/search").
			To(projectPipelineHandler.ListPipelines).
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
			To(projectPipelineHandler.GetPipelineRun).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get details in the specified pipeline activity.").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.PipelineRun{}).
			Writes(devops.PipelineRun{}))

		// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs").
			To(projectPipelineHandler.ListPipelineRuns).
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
			To(projectPipelineHandler.StopPipeline).
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
			To(projectPipelineHandler.ReplayPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Replay pipeline").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, devops.ReplayPipeline{}).
			Writes(devops.ReplayPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs").
			To(projectPipelineHandler.RunPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Run pipeline.").
			Reads(devops.RunPayload{}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Returns(http.StatusOK, RespOK, devops.RunPipeline{}).
			Writes(devops.RunPipeline{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/artifacts
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/runs/{run}/artifacts").
			To(projectPipelineHandler.GetArtifacts).
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
			To(projectPipelineHandler.GetRunLog).
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
			To(projectPipelineHandler.GetStepLog).
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
			To(projectPipelineHandler.GetNodeSteps).
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
			To(projectPipelineHandler.GetPipelineRunNodes).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get all nodes in the specified activity. node is the stage in the pipeline task").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build")).
			Returns(http.StatusOK, RespOK, []devops.PipelineRunNodes{}).
			Writes([]devops.PipelineRunNodes{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}
		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/runs/{run}/nodes/{node}/steps/{step}").
			To(projectPipelineHandler.SubmitInputStep).
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
			To(projectPipelineHandler.GetNodesDetail).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get steps details inside a activity node. For a node, the steps which defined inside the node.").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("run", "pipeline run ID, the unique ID for a pipeline once build.")).
			Returns(http.StatusOK, RespOK, []devops.NodesDetail{}).
			Writes(devops.NodesDetail{}))

		// match /blue/rest/organizations/jenkins/pipelines/{devops}/pipelines/{pipeline}/branches/{branch}
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}").
			To(projectPipelineHandler.GetBranchPipeline).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("(MultiBranchesPipeline) Get the specified branch pipeline of the DevOps project").
			Param(webservice.PathParameter("devops", "the name of devops project")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.PathParameter("branch", "the name of branch, same as repository branch")).
			Returns(http.StatusOK, RespOK, devops.BranchPipeline{}).
			Writes(devops.BranchPipeline{}))

		// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/branches/{branch}/runs/{run}/"
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/branches/{branch}/runs/{run}").
			To(projectPipelineHandler.GetBranchPipelineRun).
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
			To(projectPipelineHandler.StopBranchPipeline).
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
			To(projectPipelineHandler.ReplayBranchPipeline).
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
			To(projectPipelineHandler.RunBranchPipeline).
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
			To(projectPipelineHandler.GetBranchArtifacts).
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
			To(projectPipelineHandler.GetBranchRunLog).
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
			To(projectPipelineHandler.GetBranchStepLog).
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
			To(projectPipelineHandler.GetBranchNodeSteps).
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
			To(projectPipelineHandler.GetBranchPipelineRunNodes).
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
			To(projectPipelineHandler.SubmitBranchInputStep).
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
			To(projectPipelineHandler.GetBranchNodesDetail).
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
			To(projectPipelineHandler.GetPipelineBranch).
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
			To(projectPipelineHandler.ScanBranch).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Scan remote Repository, Start a build if have new branch.").
			Produces("text/html; charset=utf-8").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")).
			Param(webservice.QueryParameter("delay", "the delay time to scan").Required(false).DataFormat("delay=%d")))

		// match /job/project-8QnvykoJw4wZ/job/test-1/indexing/consoleText
		webservice.Route(webservice.GET("/devops/{devops}/pipelines/{pipeline}/consolelog").
			To(projectPipelineHandler.GetConsoleLog).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get scan reponsitory logs in the specified pipeline.").
			Produces("text/plain; charset=utf-8").
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline")))

		// match /crumbIssuer/api/json/
		webservice.Route(webservice.GET("/crumbissuer").
			To(projectPipelineHandler.GetCrumb).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Doc("Get crumb issuer. A CrumbIssuer represents an algorithm to generate a nonce value, known as a crumb, to counter cross site request forgery exploits. Crumbs are typically hashes incorporating information that uniquely identifies an agent that sends a request, along with a guarded secret so that the crumb value cannot be forged by a third party.").
			Returns(http.StatusOK, RespOK, devops.Crumb{}).
			Writes(devops.Crumb{}))

		// match "/blue/rest/organizations/jenkins/scm/%s/servers/"
		webservice.Route(webservice.GET("/scms/{scm}/servers").
			To(projectPipelineHandler.GetSCMServers).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("List all servers in the jenkins.").
			Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
			Returns(http.StatusOK, RespOK, []devops.SCMServer{}).
			Writes([]devops.SCMServer{}))

		// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/?credentialId=github"
		webservice.Route(webservice.GET("/scms/{scm}/organizations").
			To(projectPipelineHandler.GetSCMOrg).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("List all organizations of the specified source configuration management (SCM) such as Github.").
			Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
			Param(webservice.QueryParameter("credentialId", "credential ID for source configuration management (SCM).").Required(true).DataFormat("credentialId=%s")).
			Returns(http.StatusOK, RespOK, []devops.SCMOrg{}).
			Writes([]devops.SCMOrg{}))

		// match "/blue/rest/organizations/jenkins/scm/{scm}/organizations/{organization}/repositories/?credentialId=&pageNumber&pageSize="
		webservice.Route(webservice.GET("/scms/{scm}/organizations/{organization}/repositories").
			To(projectPipelineHandler.GetOrgRepo).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("List all repositories in the specified organization.").
			Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
			Param(webservice.PathParameter("organization", "organization ID, such as github username.")).
			Param(webservice.QueryParameter("credentialId", "credential ID for SCM.").Required(true).DataFormat("credentialId=%s")).
			Param(webservice.QueryParameter("pageNumber", "page number.").Required(true).DataFormat("pageNumber=%d")).
			Param(webservice.QueryParameter("pageSize", "the item count of one page.").Required(true).DataFormat("pageSize=%d")).
			Returns(http.StatusOK, RespOK, devops.OrgRepo{}).
			Writes(devops.OrgRepo{}))

		// match "/blue/rest/organizations/jenkins/scm/%s/servers/" create bitbucket server
		webservice.Route(webservice.POST("/scms/{scm}/servers").
			To(projectPipelineHandler.CreateSCMServers).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("Create scm server in the jenkins.").
			Param(webservice.PathParameter("scm", "The ID of the source configuration management (SCM).")).
			Reads(devops.CreateScmServerReq{}).
			Returns(http.StatusOK, RespOK, devops.SCMServer{}).
			Writes(devops.SCMServer{}))

		// match "/blue/rest/organizations/jenkins/scm/github/validate/"
		webservice.Route(webservice.POST("/scms/{scm}/verify").
			To(projectPipelineHandler.Validate).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsScmTag}).
			Doc("Validate the access token of the specified source configuration management (SCM) such as Github").
			Param(webservice.PathParameter("scm", "the ID of the source configuration management (SCM).")).
			Returns(http.StatusOK, RespOK, devops.Validates{}).
			Writes(devops.Validates{}))

		// match /git/notifyCommit/?url=
		webservice.Route(webservice.GET("/webhook/git").
			To(projectPipelineHandler.GetNotifyCommit).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification by HTTP GET method. Git webhook will request here.").
			Produces("text/plain; charset=utf-8").
			Param(webservice.QueryParameter("url", "Git url").Required(true).DataFormat("url=%s")))

		// Gitlab or some other scm managers can only use HTTP method. match /git/notifyCommit/?url=
		webservice.Route(webservice.POST("/webhook/git").
			To(projectPipelineHandler.PostNotifyCommit).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification by HTTP POST method. Git webhook will request here.").
			Consumes("application/json").
			Produces("text/plain; charset=utf-8").
			Param(webservice.QueryParameter("url", "Git url").Required(true).DataFormat("url=%s")))

		webservice.Route(webservice.POST("/webhook/github").
			To(projectPipelineHandler.GithubWebhook).
			Consumes("application/x-www-form-urlencoded", "application/json").
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsWebhookTag}).
			Doc("Get commit notification. Github webhook will request here."))

		webservice.Route(webservice.POST("/devops/{devops}/pipelines/{pipeline}/checkScriptCompile").
			To(projectPipelineHandler.CheckScriptCompile).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Param(webservice.PathParameter("pipeline", "the name of the CI/CD pipeline").DataFormat("pipeline=%s")).
			Consumes("application/x-www-form-urlencoded", "charset=utf-8").
			Produces("application/json", "charset=utf-8").
			Doc("Check pipeline script compile.").
			Reads(devops.ReqScript{}).
			Returns(http.StatusOK, RespOK, devops.CheckScript{}).
			Writes(devops.CheckScript{}))

		webservice.Route(webservice.POST("/devops/{devops}/checkCron").
			To(projectPipelineHandler.CheckCron).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsPipelineTag}).
			Param(webservice.PathParameter("devops", "DevOps project's ID, e.g. project-RRRRAzLBlLEm")).
			Produces("application/json", "charset=utf-8").
			Doc("Check cron script compile.").
			Reads(devops.CronData{}).
			Returns(http.StatusOK, RespOK, devops.CheckCronRes{}).
			Writes(devops.CheckCronRes{}))

		// match /pipeline-model-converter/toJenkinsfile
		webservice.Route(webservice.POST("/tojenkinsfile").
			To(projectPipelineHandler.ToJenkinsfile).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
			Consumes("application/x-www-form-urlencoded").
			Produces("application/json", "charset=utf-8").
			Doc("Convert json to jenkinsfile format.").
			Reads(devops.ReqJson{}).
			Returns(http.StatusOK, RespOK, devops.ResJenkinsfile{}).
			Writes(devops.ResJenkinsfile{}))

		// match /pipeline-model-converter/toJson
		webservice.Route(webservice.POST("/tojson").
			To(projectPipelineHandler.ToJson).
			Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsJenkinsfileTag}).
			Consumes("application/x-www-form-urlencoded").
			Produces("application/json", "charset=utf-8").
			Doc("Convert jenkinsfile to json format. Usually the frontend uses json to show or edit pipeline").
			Reads(devops.ReqJenkinsfile{}).
			Returns(http.StatusOK, RespOK, devops.ResJson{}).
			Writes(devops.ResJson{}))
	}
	return nil
}

func AddSonarToWebService(webservice *restful.WebService, devopsClient devops.Interface, sonarClient sonarqube.SonarInterface) error {
	sonarEnable := devopsClient != nil && sonarClient != nil
	if sonarEnable {
		sonarHandler := NewPipelineSonarHandler(devopsClient, sonarClient)
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
	return nil
}

func AddS2IToWebService(webservice *restful.WebService, ksClient versioned.Interface,
	ksInformer externalversions.SharedInformerFactory, s3Client s3.Interface) error {
	s2iEnable := ksClient != nil && ksInformer != nil && s3Client != nil

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
	return nil
}

func AddJenkinsToContainer(webservice *restful.WebService, devopsClient devops.Interface, endpoint string) error {
	if devopsClient == nil {
		return nil
	}
	parse, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	parse.Path = strings.Trim(parse.Path, "/")
	webservice.Route(webservice.GET("/jenkins/{path:*}").
		Param(webservice.PathParameter("path", "Path stands for any suffix path.")).
		To(func(request *restful.Request, response *restful.Response) {
			u := request.Request.URL
			u.Host = parse.Host
			u.Scheme = parse.Scheme
			jenkins.SetBasicBearTokenHeader(&request.Request.Header)
			u.Path = strings.Replace(request.Request.URL.Path, fmt.Sprintf("/kapis/%s/%s/jenkins", GroupVersion.Group, GroupVersion.Version), "", 1)
			httpProxy := proxy.NewUpgradeAwareHandler(u, http.DefaultTransport, false, false, &errorResponder{})
			httpProxy.ServeHTTP(response, request.Request)
		}).Returns(http.StatusOK, RespOK, nil))
	return nil
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
}
