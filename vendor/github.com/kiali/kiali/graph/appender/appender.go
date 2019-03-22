package appender

import (
	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
)

// GlobalInfo caches information relevant to a single graph. It allows
// an appender to populate the cache and then it, or another appender
// can re-use the information.  A new instance is generated for graph and
// is initially empty.
type GlobalInfo struct {
	Business       *business.Layer
	PromClient     *prometheus.Client
	ServiceEntries map[string]string
}

func NewGlobalInfo() *GlobalInfo {
	return &GlobalInfo{}
}

// NamespaceInfo caches information relevant to a single namespace. It allows
// one appender to populate the cache and another to then re-use the information.
// A new instance is generated for each namespace of a single graph and is initially
// seeded with only Namespace.
type NamespaceInfo struct {
	Namespace    string // always provided
	WorkloadList *models.WorkloadList
}

func NewNamespaceInfo(namespace string) *NamespaceInfo {
	return &NamespaceInfo{Namespace: namespace}
}

func getWorkload(workloadName string, workloadList *models.WorkloadList) (*models.WorkloadListItem, bool) {
	if workloadName == "" || workloadName == graph.Unknown {
		return nil, false
	}

	for _, workload := range workloadList.Workloads {
		if workload.Name == workloadName {
			return &workload, true
		}
	}
	return nil, false
}

func getAppWorkloads(app, version string, workloadList *models.WorkloadList) []models.WorkloadListItem {
	cfg := config.Get()
	appLabel := cfg.IstioLabels.AppLabelName
	versionLabel := cfg.IstioLabels.VersionLabelName

	result := []models.WorkloadListItem{}
	versionOk := graph.IsOK(version)
	for _, workload := range workloadList.Workloads {
		if appVal, ok := workload.Labels[appLabel]; ok && app == appVal {
			if !versionOk {
				result = append(result, workload)
			} else if versionVal, ok := workload.Labels[versionLabel]; ok && version == versionVal {
				result = append(result, workload)
			}
		}
	}
	return result
}

// Appender is implemented by any code offering to append a service graph with
// supplemental information.  On error the appender should panic and it will be
// handled as an error response.
type Appender interface {
	// AppendGraph performs the appender work on the provided traffic map. The map
	// may be initially empty. An appender is allowed to add or remove map entries.
	AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo)

	// Name returns a unique appender name and which is the name used to identify the appender (e.g in 'appenders' query param)
	Name() string
}
