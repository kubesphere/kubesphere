package logging

import (
	"io"
	"time"
)

type Interface interface {
	// Current stats about log store, eg. total number of logs and containers
	GetCurrentStats(sf SearchFilter) (Statistics, error)

	CountLogsByInterval(sf SearchFilter, interval string) (Histogram, error)

	SearchLogs(sf SearchFilter, from, size int64, order string) (Logs, error)

	ExportLogs(sf SearchFilter, w io.Writer) error
}

// Log search result
type Logs struct {
	Total   int64    `json:"total" description:"total number of matched results"`
	Records []Record `json:"records,omitempty" description:"actual array of results"`
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
	Total   int64    `json:"total" description:"total number of logs"`
	Buckets []Bucket `json:"histograms" description:"actual array of histogram results"`
}

type Bucket struct {
	Time  int64 `json:"time" description:"timestamp"`
	Count int64 `json:"count" description:"total number of logs at intervals"`
}

// General query conditions
type SearchFilter struct {
	// xxxSearch for literal matching
	// xxxfilter for fuzzy matching

	// To prevent disclosing archived logs of a reopened namespace,
	// NamespaceFilter records the namespace creation time.
	// Any query to this namespace must begin after its creation.
	NamespaceFilter map[string]time.Time
	WorkloadSearch  []string
	WorkloadFilter  []string
	PodSearch       []string
	PodFilter       []string
	ContainerSearch []string
	ContainerFilter []string
	LogSearch       []string

	Starttime time.Time
	Endtime   time.Time
}
