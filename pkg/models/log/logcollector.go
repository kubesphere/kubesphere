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
	var workspaces []string
	var projects []string
	var workloads []string
	var pods []string
	var containers []string

	var workspaces_query []string
	var projects_query []string
	var workloads_query []string
	var pods_query []string
	var containers_query []string

	switch level {
	case constants.QueryLevelCluster:
		{
			workspaces_query = strings.Split(request.QueryParameter("workspace_query"), ",")
			projects_query = strings.Split(request.QueryParameter("project_query"), ",")
			workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelWorkspace:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects_query = strings.Split(request.QueryParameter("project_query"), ",")
			workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelProject:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads_query = strings.Split(request.QueryParameter("workload_query"), ",")
			pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelWorkload:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
			pods_query = strings.Split(request.QueryParameter("pod_query"), ",")
			containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelPod:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
			pods = strings.Split(request.PathParameter("pod_name"), ",")
			containers_query = strings.Split(request.QueryParameter("container_query"), ",")
		}
	case constants.QueryLevelContainer:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
			pods = strings.Split(request.PathParameter("pod_name"), ",")
			containers = strings.Split(request.PathParameter("container_name"), ",")
		}
	}

	log.Printf("Level %v Spec workspaces %v projects %v workloads %v pods %v containers %v", level, workspaces, projects, workloads, pods, containers)
	log.Printf("Query workspaces %v projects %v workloads %v pods %v containers %v", workspaces_query, projects_query, workloads_query, pods_query, containers_query)

	log_query := request.QueryParameter("log_query")
	start_time := request.QueryParameter("start_time")
	end_time := request.QueryParameter("end_time")
	from := request.QueryParameter("from")
	size := request.QueryParameter("size")

	log.Printf("LogQuery with %s %s-%s %v-%v", log_query, start_time, end_time, from, size)

	glog.Infof("LogQuery with %s %s-%s %v-%v", log_query, start_time, end_time, from, size)

	return nil
	//return client.Query(log_query, start, end)
}
