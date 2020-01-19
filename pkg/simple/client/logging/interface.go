package logging

import (
	"bytes"
	"time"
)

const (
	Ascending = "asc"
	Descending = "desc"
)

// General query conditions
type SearchFilter struct {
	// xxxSearch for literally matching
	// xxxfilter for fuzzy matching

	// To prevent querying archived logs of a reopened namespace,
	// NamespaceFilter records the creation time of the namespace.
	// Query time span to this namespace must be limited to begin after its creation time.
	NamespaceFilter map[string]time.Time
	WorkloadSearch  []string
	WorkloadFilter  []string
	PodSearch       []string
	PodFilter       []string
	ContainerSearch []string
	ContainerFilter []string
	LogSearch       []string

	Starttime       string
	Endtime         string
}

type Interface interface {
	// returns current stat about the log store, eg. total number of logs, unique containers
	GetStatistics(sf SearchFilter) (Statistics, error)

	CountLogsByInterval(sf SearchFilter, interval string) (Histogram, error)

	SearchLogs(sf SearchFilter, from, size int64, order string) (Logs, error)

	ExportLogs(sf SearchFilter) (*bytes.Buffer, error)
}

// Log search result
type Logs struct {
	Total    int64       `json:"total" description:"total number of matched results"`
	Records  []Record `json:"records,omitempty" description:"actual array of results"`
}

type Record struct {
	Log       string `json:"log,omitempty" description:"log message"`
	Time      string `json:"time,omitempty" description:"log timestamp"`
	Namespace string `json:"namespace,omitempty" description:"namespace"`
	Pod       string `json:"pod,omitempty" description:"pod name"`
	Container string `json:"container,omitempty" description:"container name"`
}

// Log statistics result
type Statistics struct {
	Containers int64 `json:"containers" description:"total number of containers"`
	Logs       int64 `json:"logs" description:"total number of logs"`
}

// Log count result by interval
type Histogram struct {
	Total      int64             `json:"total" description:"total number of logs"`
	Buckets   []Bucket `json:"histograms" description:"actual array of histogram results"`
}

type Bucket struct {
	Time  int64 `json:"time" description:"timestamp"`
	Total int64 `json:"count" description:"total number of logs at intervals"`
}
