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
	"time"
)

type Client interface {
	SearchAuditingEvent(filter *Filter, from, size int64, sort string) (*Events, error)
	CountOverTime(filter *Filter, interval string) (*Histogram, error)
	StatisticsOnResources(filter *Filter) (*Statistics, error)
}

type Filter struct {
	ObjectRefNamespaceMap   map[string]time.Time
	WorkspaceMap            map[string]time.Time
	ObjectRefNamespaces     []string
	ObjectRefNamespaceFuzzy []string
	Workspaces              []string
	WorkspaceFuzzy          []string
	ObjectRefNames          []string
	ObjectRefNameFuzzy      []string
	Levels                  []string
	Verbs                   []string
	Users                   []string
	UserFuzzy               []string
	GroupFuzzy              []string
	SourceIpFuzzy           []string
	ObjectRefResources      []string
	ObjectRefSubresources   []string
	ResponseCodes           []int32
	ResponseStatus          []string
	StartTime               *time.Time
	EndTime                 *time.Time
}

type Event map[string]interface{}

type Events struct {
	Total   int64    `json:"total" description:"total number of matched results"`
	Records []*Event `json:"records" description:"actual array of results"`
}

type Histogram struct {
	Total   int64    `json:"total" description:"total number of events"`
	Buckets []Bucket `json:"buckets" description:"actual array of histogram results"`
}
type Bucket struct {
	Time  int64 `json:"time" description:"timestamp"`
	Count int64 `json:"count" description:"total number of events at intervals"`
}

type Statistics struct {
	Resources int64 `json:"resources" description:"total number of resources"`
	Events    int64 `json:"events" description:"total number of events"`
}
