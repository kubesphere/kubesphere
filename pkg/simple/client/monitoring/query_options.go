package monitoring

type QueryOption interface {
	Apply(*QueryOptions)
}

type QueryOptions struct {
	Level        MonitoringLevel
	NamedMetrics []string

	MetricFilter              string
	ResourceFilter            string
	NodeName                  string
	WorkspaceName             string
	NamespaceName             string
	WorkloadKind              string
	WorkloadName              string
	PodName                   string
	ContainerName             string
	StorageClassName          string
	PersistentVolumeClaimName string
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

type ClusterOption struct {
	MetricFilter string
}

func (co ClusterOption) Apply(o *QueryOptions) {
	o.Level = LevelCluster
	o.NamedMetrics = ClusterMetrics
}

type NodeOption struct {
	MetricFilter   string
	ResourceFilter string
	NodeName       string
}

func (no NodeOption) Apply(o *QueryOptions) {
	o.Level = LevelNode
	o.NamedMetrics = NodeMetrics
	o.ResourceFilter = no.ResourceFilter
	o.NodeName = no.NodeName
}

type WorkspaceOption struct {
	MetricFilter   string
	ResourceFilter string
	WorkspaceName  string
}

func (wo WorkspaceOption) Apply(o *QueryOptions) {
	o.Level = LevelWorkspace
	o.NamedMetrics = WorkspaceMetrics
	o.MetricFilter = wo.MetricFilter
	o.ResourceFilter = wo.ResourceFilter
	o.WorkspaceName = wo.WorkspaceName
}

type NamespaceOption struct {
	MetricFilter   string
	ResourceFilter string
	WorkspaceName  string
	NamespaceName  string
}

func (no NamespaceOption) Apply(o *QueryOptions) {
	o.Level = LevelNamespace
	o.NamedMetrics = NamespaceMetrics
	o.MetricFilter = no.MetricFilter
	o.ResourceFilter = no.ResourceFilter
	o.WorkspaceName = no.WorkspaceName
	o.NamespaceName = no.NamespaceName
}

type WorkloadOption struct {
	MetricFilter   string
	ResourceFilter string
	NamespaceName  string
	WorkloadKind   string
	WorkloadName   string
}

func (wo WorkloadOption) Apply(o *QueryOptions) {
	o.Level = LevelWorkload
	o.NamedMetrics = WorkspaceMetrics
	o.MetricFilter = wo.MetricFilter
	o.ResourceFilter = wo.ResourceFilter
	o.NamespaceName = wo.NamespaceName
	o.WorkloadKind = wo.WorkloadKind
	o.WorkloadName = wo.WorkloadName
}

type PodOption struct {
	MetricFilter   string
	ResourceFilter string
	NodeName       string
	NamespaceName  string
	WorkloadKind   string
	WorkloadName   string
	PodName        string
}

func (po PodOption) Apply(o *QueryOptions) {
	o.Level = LevelPod
	o.NamedMetrics = PodMetrics
	o.MetricFilter = po.MetricFilter
	o.ResourceFilter = po.ResourceFilter
	o.NamespaceName = po.NamespaceName
	o.WorkloadKind = po.WorkloadKind
	o.WorkloadName = po.WorkloadName
}

type ContainerOption struct {
	MetricFilter   string
	ResourceFilter string
	NamespaceName  string
	PodName        string
	ContainerName  string
}

func (co ContainerOption) Apply(o *QueryOptions) {
	o.Level = LevelContainer
	o.NamedMetrics = ContainerMetrics
	o.MetricFilter = co.MetricFilter
	o.ResourceFilter = co.ResourceFilter
	o.NamespaceName = co.NamespaceName
	o.PodName = co.PodName
	o.ContainerName = co.ContainerName
}

type PVCOption struct {
	MetricFilter              string
	ResourceFilter            string
	NamespaceName             string
	StorageClassName          string
	PersistentVolumeClaimName string
}

func (po PVCOption) Apply(o *QueryOptions) {
	o.Level = LevelPVC
	o.NamedMetrics = PVCMetrics
	o.MetricFilter = po.MetricFilter
	o.ResourceFilter = po.ResourceFilter
	o.NamespaceName = po.NamespaceName
	o.StorageClassName = po.StorageClassName
	o.PersistentVolumeClaimName = po.PersistentVolumeClaimName
}

type ComponentOption struct {
	MetricFilter string
}

func (co ComponentOption) Apply(o *QueryOptions) {
	o.Level = LevelComponent
	o.NamedMetrics = ComponentMetrics
	o.MetricFilter = co.MetricFilter
}
