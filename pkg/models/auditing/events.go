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

package auditing

import (
	"kubesphere.io/kubesphere/pkg/api/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strconv"
)

type Interface interface {
	Events(queryParam *v1alpha1.Query, MutateFilterFunc func(*auditing.Filter)) (*v1alpha1.APIResponse, error)
}

type eventsOperator struct {
	client auditing.Client
}

func NewEventsOperator(client auditing.Client) Interface {
	return &eventsOperator{client}
}

func (eo *eventsOperator) Events(queryParam *v1alpha1.Query,
	MutateFilterFunc func(*auditing.Filter)) (*v1alpha1.APIResponse, error) {
	filter := &auditing.Filter{
		ObjectRefNamespaces:     stringutils.Split(queryParam.ObjectRefNamespaceFilter, ","),
		ObjectRefNamespaceFuzzy: stringutils.Split(queryParam.ObjectRefNamespaceSearch, ","),
		Workspaces:              stringutils.Split(queryParam.WorkspaceFilter, ","),
		WorkspaceFuzzy:          stringutils.Split(queryParam.WorkspaceSearch, ","),
		ObjectRefNames:          stringutils.Split(queryParam.ObjectRefNameFilter, ","),
		ObjectRefNameFuzzy:      stringutils.Split(queryParam.ObjectRefNameSearch, ","),
		Levels:                  stringutils.Split(queryParam.LevelFilter, ","),
		Verbs:                   stringutils.Split(queryParam.VerbFilter, ","),
		Users:                   stringutils.Split(queryParam.UserFilter, ","),
		UserFuzzy:               stringutils.Split(queryParam.UserSearch, ","),
		GroupFuzzy:              stringutils.Split(queryParam.GroupSearch, ","),
		SourceIpFuzzy:           stringutils.Split(queryParam.SourceIpSearch, ","),
		ObjectRefResources:      stringutils.Split(queryParam.ObjectRefResourceFilter, ","),
		ObjectRefSubresources:   stringutils.Split(queryParam.ObjectRefSubresourceFilter, ","),
		ResponseStatus:          stringutils.Split(queryParam.ResponseStatusFilter, ","),
		StartTime:               queryParam.StartTime,
		EndTime:                 queryParam.EndTime,
	}
	if MutateFilterFunc != nil {
		MutateFilterFunc(filter)
	}

	cs := stringutils.Split(queryParam.ResponseCodeFilter, ",")
	for _, c := range cs {
		code, err := strconv.ParseInt(c, 10, 64)
		if err != nil {
			continue
		}

		filter.ResponseCodes = append(filter.ResponseCodes, int32(code))
	}

	var ar v1alpha1.APIResponse
	var err error
	switch queryParam.Operation {
	case "histogram":
		if len(filter.ObjectRefNamespaceMap) == 0 && len(filter.WorkspaceMap) == 0 {
			ar.Histogram = &auditing.Histogram{}
		} else {
			ar.Histogram, err = eo.client.CountOverTime(filter, queryParam.Interval)
		}
	case "statistics":
		if len(filter.ObjectRefNamespaceMap) == 0 && len(filter.WorkspaceMap) == 0 {
			ar.Statistics = &auditing.Statistics{}
		} else {
			ar.Statistics, err = eo.client.StatisticsOnResources(filter)
		}
	default:
		if len(filter.ObjectRefNamespaceMap) == 0 && len(filter.WorkspaceMap) == 0 {
			ar.Events = &auditing.Events{}
		} else {
			ar.Events, err = eo.client.SearchAuditingEvent(filter, queryParam.From, queryParam.Size, queryParam.Sort)
		}
	}
	if err != nil {
		return nil, err
	}
	return &ar, nil
}
