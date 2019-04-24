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
)

const GroupName = "devops.kubesphere.io"

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
		Param(webservice.PathParameter("devops", "projectId")).
		Writes(&devops.DevOpsProject{}))

	webservice.Route(webservice.PATCH("/devops/{devops}").
		To(devopsapi.UpdateProjectHandler).
		Doc("get devops project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("devops", "projectId")).
		Writes(&devops.DevOpsProject{}))

	webservice.Route(webservice.GET("/devops/{devops}/defaultroles").
		To(devopsapi.GetDevOpsProjectDefaultRoles).
		Doc("get devops project defaultroles").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&devops.DefaultRoles))

	webservice.Route(webservice.GET("/devops/{devops}/members").
		To(devopsapi.GetDevOpsProjectMembersHandler).
		Doc("get devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&[]*devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.GET("/devops/{devops}/members/{members}").
		To(devopsapi.GetDevOpsProjectMemberHandler).
		Doc("get devops project member").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.POST("/devops/{devops}/members").
		To(devopsapi.AddDevOpsProjectMemberHandler).
		Doc("add devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&devops.DevOpsProjectMembership{}))

	webservice.Route(webservice.PATCH("/devops/{devops}/members/{members}").
		To(devopsapi.UpdateDevOpsProjectMemberHandler).
		Doc("update devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&devops.DevOpsProjectMembership{}))
	webservice.Route(webservice.DELETE("/devops/{devops}/members/{members}").
		To(devopsapi.DeleteDevOpsProjectMemberHandler).
		Doc("delete devops project members").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(&devops.DevOpsProjectMembership{}))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}").
		To(devopsapi.GetPipeline).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")))

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
			DataFormat("limit=%d")))

	// match Jenkisn api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/runs/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/runs").
		To(devopsapi.SearchPipelineRuns).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Search DevOps Pipelines runs.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.QueryParameter("start", "start page").
			Required(false).
			DataFormat("start=%d")).
		Param(webservice.QueryParameter("limit", "limit count").
			Required(false).
			DataFormat("limit=%d")))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}").
		To(devopsapi.GetPipelineRun).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get DevOps Pipelines run.").
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")))

	// match Jenkins api "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes").
		To(devopsapi.GetPipelineRunNodes).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get node on DevOps Pipelines run.").
		Param(webservice.PathParameter("projectName", "devops project name")).
		Param(webservice.PathParameter("pipelineName", "pipeline name")).
		Param(webservice.PathParameter("branchName", "pipeline branch name")).
		Param(webservice.PathParameter("runId", "pipeline runs id")).
		Param(webservice.QueryParameter("limit", "limit").
			Required(false).
			DataFormat("limit=%d").
			DefaultValue("limit=10000")))

	// match "/blue/rest/organizations/jenkins/pipelines/{projectName}/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log/?start=0"
	webservice.Route(webservice.GET("/devops/{projectName}/pipelines/{pipelineName}/branches/{branchName}/runs/{runId}/nodes/{nodeId}/steps/{stepId}/log").
		To(devopsapi.GetStepLog).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get Pipelines step log.").
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
		To(devopsapi.Validate).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Validate Github personal access token.").
		Param(webservice.PathParameter("scmId", "SCM id")))

	// match "/blue/rest/organizations/jenkins/scm/{scmId}/organizations/?credentialId=github"
	webservice.Route(webservice.GET("/devops/scm/{scmId}/organizations").
		To(devopsapi.GetOrgSCM).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("List organizations of SCM").
		Param(webservice.PathParameter("scmId", "SCM id")).
		Param(webservice.QueryParameter("credentialId", "credentialId for SCM").
			Required(true).
			DataFormat("credentialId=%s")))

	c.Add(webservice)

	return nil
}
