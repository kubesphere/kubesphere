package monitoring

type ClusterQuery struct {
}

type ClusterMetrics struct {
}

type WorkspaceQuery struct {
}

type WorkspaceMetrics struct {
}

type NamespaceQuery struct {
}

type NamespaceMetrics struct {
}

// Interface defines all the abstract behaviors of monitoring
type Interface interface {

	// Get
	GetClusterMetrics(query ClusterQuery) ClusterMetrics

	//
	GetWorkspaceMetrics(query WorkspaceQuery) WorkspaceMetrics

	//
	GetNamespaceMetrics(query NamespaceQuery) NamespaceMetrics
}
