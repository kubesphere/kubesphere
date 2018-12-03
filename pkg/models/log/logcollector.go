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
	//log.Printf("LogQuery level %v", level)
	var workspaces []string
	var projects []string
	var workloads []string
	var pods []string
	var containers []string

	switch level {
	case constants.Cluster:
		{}
	case constants.Workspace:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
		}
	case constants.Project:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
		}
	case constants.Workload:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
		}
	case constants.Pod:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
			pods = strings.Split(request.PathParameter("pod_name"), ",")
		}
	case constants.Container:
		{
			workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			projects = strings.Split(request.PathParameter("project_name"), ",")
			workloads = strings.Split(request.PathParameter("workload_name"), ",")
			pods = strings.Split(request.PathParameter("pod_name"), ",")
			containers = strings.Split(request.PathParameter("container_name"), ",")
		}
	}

	log.Printf("workspaces %v projects %v workloads %v pods %v containers %v", workspaces, projects, workloads, pods, containers)

	log_query := request.QueryParameter("log_query")
	start := request.QueryParameter("start")
	end := request.QueryParameter("end")

	glog.Infof("LogQuery with %s %s %s", log_query, start, end)

	return client.Query(log_query, start, end)
}
