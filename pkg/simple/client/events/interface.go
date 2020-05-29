package events

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type Client interface {
	SearchEvents(filter *Filter, from, size int64, sort string) (*Events, error)
	CountOverTime(filter *Filter, interval string) (*Histogram, error)
	StatisticsOnResources(filter *Filter) (*Statistics, error)
}

type Filter struct {
	InvolvedObjectNamespaceMap map[string]time.Time
	InvolvedObjectNames        []string
	InvolvedObjectNameFuzzy    []string
	InvolvedObjectkinds        []string
	Reasons                    []string
	ReasonFuzzy                []string
	MessageFuzzy               []string
	Type                       string
	StartTime                  *time.Time
	EndTime                    *time.Time
}

type Events struct {
	Total   int64       `json:"total" description:"total number of matched results"`
	Records []*v1.Event `json:"records" description:"actual array of results"`
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
