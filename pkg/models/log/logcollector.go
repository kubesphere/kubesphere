/*
Copyright 2018 The KubeSphere Authors.

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

package log

import (
	//"fmt"
	//"encoding/json"
	//"regexp"
	"strings"
	"log"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"

	//"time"

	//"k8s.io/api/core/v1"
	//metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
	//"kubesphere.io/kubesphere/pkg/models"
	"github.com/olivere/elastic"
)

func LogQuery(level constants.LogQueryLevel, request *restful.Request) *elastic.SearchResult {
	var param client.QueryParameters

	param.Level = level

	switch level {
	case constants.QueryLevelCluster:
		{
			param.Workspaces_query = strings.Split(request.QueryParameter("workspace_query"), ",")
			param.Projects_query = strings.Split(request.QueryParameter("project_query"), ",")
			param.Workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			param.Pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			param.Containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelWorkspace:
		{
			param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Projects_query = strings.Split(request.QueryParameter("project_query"), ",")
			param.Workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			param.Pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			param.Containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelProject:
		{
			param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Projects = strings.Split(request.PathParameter("project_name"), ",")
			param.Workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			param.Pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			param.Containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelWorkload:
		{
			param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Projects = strings.Split(request.PathParameter("project_name"), ",")
			param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			param.Containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelPod:
		{
			param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Projects = strings.Split(request.PathParameter("project_name"), ",")
			param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods = strings.Split(request.PathParameter("pod_name"), ",")
			param.Containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelContainer:
		{
			param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Projects = strings.Split(request.PathParameter("project_name"), ",")
			param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods = strings.Split(request.PathParameter("pod_name"), ",")
			param.Containers = strings.Split(request.PathParameter("container_name"), ",")
		}
	}

	param.Log_query = request.QueryParameter("log_query")
	param.Start_time = request.QueryParameter("start_time")
	param.End_time = request.QueryParameter("end_time")

	var err error
	param.From, err = strconv.Atoi(request.QueryParameter("from"))
	if err != nil {
		param.From = 0
	}
	param.Size, err = strconv.Atoi(request.QueryParameter("size"))
	if err != nil {
		param.Size = 10
	}

	log.Printf("LogQuery with %v", param)

	glog.Infof("LogQuery with %v", param)

	return client.Query(param)
}
