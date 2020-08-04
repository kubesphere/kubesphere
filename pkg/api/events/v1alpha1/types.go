/*
Copyright 2020 KubeSphere Authors

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
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"strconv"
	"time"
)

type APIResponse struct {
	Events     *events.Events     `json:"query,omitempty" description:"query results"`
	Statistics *events.Statistics `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *events.Histogram  `json:"histogram,omitempty" description:"histogram results"`
}

type Query struct {
	Operation                     string `json:"operation,omitempty"`
	WorkspaceFilter               string `json:"workspace_filter,omitempty"`
	WorkspaceSearch               string `json:"workspace_search,omitempty"`
	InvolvedObjectNamespaceFilter string `json:"involved_object_namespace_filter,omitempty"`
	InvolvedObjectNamespaceSearch string `json:"involved_object_namespace_search,omitempty"`
	InvolvedObjectNameFilter      string `json:"involved_object_name_filter,omitempty"`
	InvolvedObjectNameSearch      string `json:"involved_object_name_search,omitempty"`
	InvolvedObjectKindFilter      string `json:"involved_object_kind_filter,omitempty"`
	ReasonFilter                  string `json:"reason_filter,omitempty"`
	ReasonSearch                  string `json:"reason_search,omitempty"`
	MessageSearch                 string `json:"message_search,omitempty"`
	TypeFilter                    string `json:"type_filter,omitempty"`

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
	q.InvolvedObjectNamespaceFilter = req.QueryParameter("involved_object_namespace_filter")
	q.InvolvedObjectNamespaceSearch = req.QueryParameter("involved_object_namespace_search")
	q.InvolvedObjectNameFilter = req.QueryParameter("involved_object_name_filter")
	q.InvolvedObjectNameSearch = req.QueryParameter("involved_object_name_search")
	q.InvolvedObjectKindFilter = req.QueryParameter("involved_object_kind_filter")
	q.ReasonFilter = req.QueryParameter("reason_filter")
	q.ReasonSearch = req.QueryParameter("reason_search")
	q.MessageSearch = req.QueryParameter("message_search")
	q.TypeFilter = req.QueryParameter("type_filter")

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
