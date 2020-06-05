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

package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"strconv"
	"time"
)

type APIResponse struct {
	Events     *auditing.Events     `json:"query,omitempty" description:"query results"`
	Statistics *auditing.Statistics `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *auditing.Histogram  `json:"histogram,omitempty" description:"histogram results"`
}

type Query struct {
	Operation                  string `json:"operation,omitempty"`
	WorkspaceFilter            string `json:"workspace_filter,omitempty"`
	WorkspaceSearch            string `json:"workspace_search,omitempty"`
	ObjectRefNamespaceFilter   string `json:"objectref_namespace_filter,omitempty"`
	ObjectRefNamespaceSearch   string `json:"objectref_namespace_search,omitempty"`
	ObjectRefNameFilter        string `json:"objectref_name_filter,omitempty"`
	ObjectRefNameSearch        string `json:"objectref_name_search,omitempty"`
	LevelFilter                string `json:"level_filter,omitempty"`
	VerbFilter                 string `json:"verb_filter,omitempty"`
	UserFilter                 string `json:"user_filter,omitempty"`
	UserSearch                 string `json:"user_search,omitempty"`
	GroupSearch                string `json:"group_search,omitempty"`
	SourceIpSearch             string `json:"source_ip_search,omitempty"`
	ObjectRefResourceFilter    string `json:"objectref_resource_filter,omitempty"`
	ObjectRefSubresourceFilter string `json:"objectref_subresource_filter,omitempty"`
	ResponseCodeFilter         string `json:"response_code_filter,omitempty"`
	ResponseStatusFilter       string `json:"response_status_filter,omitempty"`

	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	Interval string `json:"interval,omitempty"`
	Sort     string `json:"sort,omitempty"`
	From     int64  `json:"from,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

func ParseQueryParameter(req *restful.Request) (*Query, error) {
	q := &Query{}

	q.Operation = req.QueryParameter("operation")
	q.WorkspaceFilter = req.QueryParameter("workspace_filter")
	q.WorkspaceSearch = req.QueryParameter("workspace_search")
	q.ObjectRefNamespaceFilter = req.QueryParameter("objectref_namespace_filter")
	q.ObjectRefNamespaceSearch = req.QueryParameter("objectref_namespace_search")
	q.ObjectRefNameFilter = req.QueryParameter("objectref_name_filter")
	q.ObjectRefNameSearch = req.QueryParameter("objectref_name_search")
	q.LevelFilter = req.QueryParameter("level_filter")
	q.VerbFilter = req.QueryParameter("verb_filter")
	q.SourceIpSearch = req.QueryParameter("source_ip_search")
	q.UserFilter = req.QueryParameter("user_filter")
	q.UserSearch = req.QueryParameter("user_search")
	q.GroupSearch = req.QueryParameter("group_search")
	q.ObjectRefResourceFilter = req.QueryParameter("objectref_resource_filter")
	q.ObjectRefSubresourceFilter = req.QueryParameter("objectref_subresource_filter")
	q.ResponseCodeFilter = req.QueryParameter("response_code_filter")
	q.ResponseStatusFilter = req.QueryParameter("response_status_filter")

	if tstr := req.QueryParameter("start_time"); tstr != "" {
		sec, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return nil, err
		}
		t := time.Unix(sec, 0)
		q.StartTime = &t
	}
	if tstr := req.QueryParameter("end_time"); tstr != "" {
		sec, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return nil, err
		}
		t := time.Unix(sec, 0)
		q.EndTime = &t
	}
	if q.Interval = req.QueryParameter("interval"); q.Interval == "" {
		q.Interval = "15m"
	}
	q.From, _ = strconv.ParseInt(req.QueryParameter("from"), 10, 64)
	size, err := strconv.ParseInt(req.QueryParameter("size"), 10, 64)
	if err != nil {
		size = 10
	}
	q.Size = size
	if q.Sort = req.QueryParameter("sort"); q.Sort != "asc" {
		q.Sort = "desc"
	}
	return q, nil
}
