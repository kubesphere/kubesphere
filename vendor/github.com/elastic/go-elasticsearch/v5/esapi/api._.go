// Code generated from specification version 5.6.15 (fe7575a32e2): DO NOT EDIT

package esapi

// API contains the Elasticsearch APIs
//
type API struct {
	Cat      *Cat
	Cluster  *Cluster
	Indices  *Indices
	Ingest   *Ingest
	Nodes    *Nodes
	Remote   *Remote
	Snapshot *Snapshot
	Tasks    *Tasks

	Bulk                    Bulk
	ClearScroll             ClearScroll
	CountPercolate          CountPercolate
	Count                   Count
	Create                  Create
	DeleteByQuery           DeleteByQuery
	DeleteByQueryRethrottle DeleteByQueryRethrottle
	Delete                  Delete
	DeleteScript            DeleteScript
	DeleteTemplate          DeleteTemplate
	Exists                  Exists
	ExistsSource            ExistsSource
	Explain                 Explain
	FieldCaps               FieldCaps
	FieldStats              FieldStats
	Get                     Get
	GetScript               GetScript
	GetSource               GetSource
	GetTemplate             GetTemplate
	Index                   Index
	Info                    Info
	Mget                    Mget
	Mpercolate              Mpercolate
	Msearch                 Msearch
	MsearchTemplate         MsearchTemplate
	Mtermvectors            Mtermvectors
	Percolate               Percolate
	Ping                    Ping
	PutScript               PutScript
	PutTemplate             PutTemplate
	RankEval                RankEval
	Reindex                 Reindex
	ReindexRethrottle       ReindexRethrottle
	RenderSearchTemplate    RenderSearchTemplate
	ScriptsPainlessExecute  ScriptsPainlessExecute
	Scroll                  Scroll
	Search                  Search
	SearchShards            SearchShards
	SearchTemplate          SearchTemplate
	Suggest                 Suggest
	Termvectors             Termvectors
	UpdateByQuery           UpdateByQuery
	UpdateByQueryRethrottle UpdateByQueryRethrottle
	Update                  Update
}

// Cat contains the Cat APIs
type Cat struct {
	Aliases      CatAliases
	Allocation   CatAllocation
	Count        CatCount
	Fielddata    CatFielddata
	Health       CatHealth
	Help         CatHelp
	Indices      CatIndices
	Master       CatMaster
	Nodeattrs    CatNodeattrs
	Nodes        CatNodes
	PendingTasks CatPendingTasks
	Plugins      CatPlugins
	Recovery     CatRecovery
	Repositories CatRepositories
	Segments     CatSegments
	Shards       CatShards
	Snapshots    CatSnapshots
	Tasks        CatTasks
	Templates    CatTemplates
	ThreadPool   CatThreadPool
}

// Cluster contains the Cluster APIs
type Cluster struct {
	AllocationExplain ClusterAllocationExplain
	GetSettings       ClusterGetSettings
	Health            ClusterHealth
	PendingTasks      ClusterPendingTasks
	PutSettings       ClusterPutSettings
	RemoteInfo        ClusterRemoteInfo
	Reroute           ClusterReroute
	State             ClusterState
	Stats             ClusterStats
}

// Indices contains the Indices APIs
type Indices struct {
	Analyze            IndicesAnalyze
	ClearCache         IndicesClearCache
	Close              IndicesClose
	Create             IndicesCreate
	DeleteAlias        IndicesDeleteAlias
	Delete             IndicesDelete
	DeleteTemplate     IndicesDeleteTemplate
	ExistsAlias        IndicesExistsAlias
	ExistsDocumentType IndicesExistsDocumentType
	Exists             IndicesExists
	ExistsTemplate     IndicesExistsTemplate
	Flush              IndicesFlush
	FlushSynced        IndicesFlushSynced
	Forcemerge         IndicesForcemerge
	GetAlias           IndicesGetAlias
	GetFieldMapping    IndicesGetFieldMapping
	GetMapping         IndicesGetMapping
	Get                IndicesGet
	GetSettings        IndicesGetSettings
	GetTemplate        IndicesGetTemplate
	GetUpgrade         IndicesGetUpgrade
	Open               IndicesOpen
	PutAlias           IndicesPutAlias
	PutMapping         IndicesPutMapping
	PutSettings        IndicesPutSettings
	PutTemplate        IndicesPutTemplate
	Recovery           IndicesRecovery
	Refresh            IndicesRefresh
	Rollover           IndicesRollover
	Segments           IndicesSegments
	ShardStores        IndicesShardStores
	Shrink             IndicesShrink
	Split              IndicesSplit
	Stats              IndicesStats
	UpdateAliases      IndicesUpdateAliases
	Upgrade            IndicesUpgrade
	ValidateQuery      IndicesValidateQuery
}

// Ingest contains the Ingest APIs
type Ingest struct {
	DeletePipeline IngestDeletePipeline
	GetPipeline    IngestGetPipeline
	ProcessorGrok  IngestProcessorGrok
	PutPipeline    IngestPutPipeline
	Simulate       IngestSimulate
}

// Nodes contains the Nodes APIs
type Nodes struct {
	HotThreads           NodesHotThreads
	Info                 NodesInfo
	ReloadSecureSettings NodesReloadSecureSettings
	Stats                NodesStats
	Usage                NodesUsage
}

// Remote contains the Remote APIs
type Remote struct {
}

// Snapshot contains the Snapshot APIs
type Snapshot struct {
	CreateRepository SnapshotCreateRepository
	Create           SnapshotCreate
	DeleteRepository SnapshotDeleteRepository
	Delete           SnapshotDelete
	GetRepository    SnapshotGetRepository
	Get              SnapshotGet
	Restore          SnapshotRestore
	Status           SnapshotStatus
	VerifyRepository SnapshotVerifyRepository
}

// Tasks contains the Tasks APIs
type Tasks struct {
	Cancel TasksCancel
	Get    TasksGet
	List   TasksList
}

// New creates new API
//
func New(t Transport) *API {
	return &API{
		Bulk:                    newBulkFunc(t),
		ClearScroll:             newClearScrollFunc(t),
		CountPercolate:          newCountPercolateFunc(t),
		Count:                   newCountFunc(t),
		Create:                  newCreateFunc(t),
		DeleteByQuery:           newDeleteByQueryFunc(t),
		DeleteByQueryRethrottle: newDeleteByQueryRethrottleFunc(t),
		Delete:                  newDeleteFunc(t),
		DeleteScript:            newDeleteScriptFunc(t),
		DeleteTemplate:          newDeleteTemplateFunc(t),
		Exists:                  newExistsFunc(t),
		ExistsSource:            newExistsSourceFunc(t),
		Explain:                 newExplainFunc(t),
		FieldCaps:               newFieldCapsFunc(t),
		FieldStats:              newFieldStatsFunc(t),
		Get:                     newGetFunc(t),
		GetScript:               newGetScriptFunc(t),
		GetSource:               newGetSourceFunc(t),
		GetTemplate:             newGetTemplateFunc(t),
		Index:                   newIndexFunc(t),
		Info:                    newInfoFunc(t),
		Mget:                    newMgetFunc(t),
		Mpercolate:              newMpercolateFunc(t),
		Msearch:                 newMsearchFunc(t),
		MsearchTemplate:         newMsearchTemplateFunc(t),
		Mtermvectors:            newMtermvectorsFunc(t),
		Percolate:               newPercolateFunc(t),
		Ping:                    newPingFunc(t),
		PutScript:               newPutScriptFunc(t),
		PutTemplate:             newPutTemplateFunc(t),
		RankEval:                newRankEvalFunc(t),
		Reindex:                 newReindexFunc(t),
		ReindexRethrottle:       newReindexRethrottleFunc(t),
		RenderSearchTemplate:    newRenderSearchTemplateFunc(t),
		ScriptsPainlessExecute:  newScriptsPainlessExecuteFunc(t),
		Scroll:                  newScrollFunc(t),
		Search:                  newSearchFunc(t),
		SearchShards:            newSearchShardsFunc(t),
		SearchTemplate:          newSearchTemplateFunc(t),
		Suggest:                 newSuggestFunc(t),
		Termvectors:             newTermvectorsFunc(t),
		UpdateByQuery:           newUpdateByQueryFunc(t),
		UpdateByQueryRethrottle: newUpdateByQueryRethrottleFunc(t),
		Update:                  newUpdateFunc(t),
		Cat: &Cat{
			Aliases:      newCatAliasesFunc(t),
			Allocation:   newCatAllocationFunc(t),
			Count:        newCatCountFunc(t),
			Fielddata:    newCatFielddataFunc(t),
			Health:       newCatHealthFunc(t),
			Help:         newCatHelpFunc(t),
			Indices:      newCatIndicesFunc(t),
			Master:       newCatMasterFunc(t),
			Nodeattrs:    newCatNodeattrsFunc(t),
			Nodes:        newCatNodesFunc(t),
			PendingTasks: newCatPendingTasksFunc(t),
			Plugins:      newCatPluginsFunc(t),
			Recovery:     newCatRecoveryFunc(t),
			Repositories: newCatRepositoriesFunc(t),
			Segments:     newCatSegmentsFunc(t),
			Shards:       newCatShardsFunc(t),
			Snapshots:    newCatSnapshotsFunc(t),
			Tasks:        newCatTasksFunc(t),
			Templates:    newCatTemplatesFunc(t),
			ThreadPool:   newCatThreadPoolFunc(t),
		},
		Cluster: &Cluster{
			AllocationExplain: newClusterAllocationExplainFunc(t),
			GetSettings:       newClusterGetSettingsFunc(t),
			Health:            newClusterHealthFunc(t),
			PendingTasks:      newClusterPendingTasksFunc(t),
			PutSettings:       newClusterPutSettingsFunc(t),
			RemoteInfo:        newClusterRemoteInfoFunc(t),
			Reroute:           newClusterRerouteFunc(t),
			State:             newClusterStateFunc(t),
			Stats:             newClusterStatsFunc(t),
		},
		Indices: &Indices{
			Analyze:            newIndicesAnalyzeFunc(t),
			ClearCache:         newIndicesClearCacheFunc(t),
			Close:              newIndicesCloseFunc(t),
			Create:             newIndicesCreateFunc(t),
			DeleteAlias:        newIndicesDeleteAliasFunc(t),
			Delete:             newIndicesDeleteFunc(t),
			DeleteTemplate:     newIndicesDeleteTemplateFunc(t),
			ExistsAlias:        newIndicesExistsAliasFunc(t),
			ExistsDocumentType: newIndicesExistsDocumentTypeFunc(t),
			Exists:             newIndicesExistsFunc(t),
			ExistsTemplate:     newIndicesExistsTemplateFunc(t),
			Flush:              newIndicesFlushFunc(t),
			FlushSynced:        newIndicesFlushSyncedFunc(t),
			Forcemerge:         newIndicesForcemergeFunc(t),
			GetAlias:           newIndicesGetAliasFunc(t),
			GetFieldMapping:    newIndicesGetFieldMappingFunc(t),
			GetMapping:         newIndicesGetMappingFunc(t),
			Get:                newIndicesGetFunc(t),
			GetSettings:        newIndicesGetSettingsFunc(t),
			GetTemplate:        newIndicesGetTemplateFunc(t),
			GetUpgrade:         newIndicesGetUpgradeFunc(t),
			Open:               newIndicesOpenFunc(t),
			PutAlias:           newIndicesPutAliasFunc(t),
			PutMapping:         newIndicesPutMappingFunc(t),
			PutSettings:        newIndicesPutSettingsFunc(t),
			PutTemplate:        newIndicesPutTemplateFunc(t),
			Recovery:           newIndicesRecoveryFunc(t),
			Refresh:            newIndicesRefreshFunc(t),
			Rollover:           newIndicesRolloverFunc(t),
			Segments:           newIndicesSegmentsFunc(t),
			ShardStores:        newIndicesShardStoresFunc(t),
			Shrink:             newIndicesShrinkFunc(t),
			Split:              newIndicesSplitFunc(t),
			Stats:              newIndicesStatsFunc(t),
			UpdateAliases:      newIndicesUpdateAliasesFunc(t),
			Upgrade:            newIndicesUpgradeFunc(t),
			ValidateQuery:      newIndicesValidateQueryFunc(t),
		},
		Ingest: &Ingest{
			DeletePipeline: newIngestDeletePipelineFunc(t),
			GetPipeline:    newIngestGetPipelineFunc(t),
			ProcessorGrok:  newIngestProcessorGrokFunc(t),
			PutPipeline:    newIngestPutPipelineFunc(t),
			Simulate:       newIngestSimulateFunc(t),
		},
		Nodes: &Nodes{
			HotThreads:           newNodesHotThreadsFunc(t),
			Info:                 newNodesInfoFunc(t),
			ReloadSecureSettings: newNodesReloadSecureSettingsFunc(t),
			Stats:                newNodesStatsFunc(t),
			Usage:                newNodesUsageFunc(t),
		},
		Remote: &Remote{},
		Snapshot: &Snapshot{
			CreateRepository: newSnapshotCreateRepositoryFunc(t),
			Create:           newSnapshotCreateFunc(t),
			DeleteRepository: newSnapshotDeleteRepositoryFunc(t),
			Delete:           newSnapshotDeleteFunc(t),
			GetRepository:    newSnapshotGetRepositoryFunc(t),
			Get:              newSnapshotGetFunc(t),
			Restore:          newSnapshotRestoreFunc(t),
			Status:           newSnapshotStatusFunc(t),
			VerifyRepository: newSnapshotVerifyRepositoryFunc(t),
		},
		Tasks: &Tasks{
			Cancel: newTasksCancelFunc(t),
			Get:    newTasksGetFunc(t),
			List:   newTasksListFunc(t),
		},
	}
}
