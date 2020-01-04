package monitoring

import (
    "net/http"
    "time"
)

// prometheus implements monitoring interface backed by Prometheus
type prometheus struct {
    options *Options
    client            *http.Client
}

func NewPrometheus(options *Options) Interface {
    return &prometheus{
        options:options,
        client: &http.Client{ Timeout: 10 * time.Second },
    }
}

func (p prometheus) GetClusterMetrics(query ClusterQuery) ClusterMetrics {
    panic("implement me")
}

func (p prometheus) GetWorkspaceMetrics(query WorkspaceQuery) WorkspaceMetrics {
    panic("implement me")
}

func (p prometheus) GetNamespaceMetrics(query NamespaceQuery) NamespaceMetrics {
    panic("implement me")
}
