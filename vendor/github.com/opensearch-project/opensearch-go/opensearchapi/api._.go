// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package opensearchapi

// API contains the OpenSearch APIs
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

	Bulk                               Bulk
	ClearScroll                        ClearScroll
	Count                              Count
	Create                             Create
	DanglingIndicesDeleteDanglingIndex DanglingIndicesDeleteDanglingIndex
	DanglingIndicesImportDanglingIndex DanglingIndicesImportDanglingIndex
	DanglingIndicesListDanglingIndices DanglingIndicesListDanglingIndices
	DeleteByQuery                      DeleteByQuery
	DeleteByQueryRethrottle            DeleteByQueryRethrottle
	Delete                             Delete
	DeleteScript                       DeleteScript
	Exists                             Exists
	ExistsSource                       ExistsSource
	Explain                            Explain
	FieldCaps                          FieldCaps
	Get                                Get
	GetScriptContext                   GetScriptContext
	GetScriptLanguages                 GetScriptLanguages
	GetScript                          GetScript
	GetSource                          GetSource
	Index                              Index
	Info                               Info
	Mget                               Mget
	Msearch                            Msearch
	MsearchTemplate                    MsearchTemplate
	Mtermvectors                       Mtermvectors
	Ping                               Ping
	PutScript                          PutScript
	RankEval                           RankEval
	Reindex                            Reindex
	ReindexRethrottle                  ReindexRethrottle
	RenderSearchTemplate               RenderSearchTemplate
	ScriptsPainlessExecute             ScriptsPainlessExecute
	Scroll                             Scroll
	Search                             Search
	SearchShards                       SearchShards
	SearchTemplate                     SearchTemplate
	TermsEnum                          TermsEnum
	Termvectors                        Termvectors
	UpdateByQuery                      UpdateByQuery
	UpdateByQueryRethrottle            UpdateByQueryRethrottle
	Update                             Update
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
	AllocationExplain            ClusterAllocationExplain
	DeleteComponentTemplate      ClusterDeleteComponentTemplate
	DeleteVotingConfigExclusions ClusterDeleteVotingConfigExclusions
	ExistsComponentTemplate      ClusterExistsComponentTemplate
	GetComponentTemplate         ClusterGetComponentTemplate
	GetSettings                  ClusterGetSettings
	Health                       ClusterHealth
	PendingTasks                 ClusterPendingTasks
	PostVotingConfigExclusions   ClusterPostVotingConfigExclusions
	PutComponentTemplate         ClusterPutComponentTemplate
	PutSettings                  ClusterPutSettings
	RemoteInfo                   ClusterRemoteInfo
	Reroute                      ClusterReroute
	State                        ClusterState
	Stats                        ClusterStats
}

// Indices contains the Indices APIs
type Indices struct {
	AddBlock              IndicesAddBlock
	Analyze               IndicesAnalyze
	ClearCache            IndicesClearCache
	Clone                 IndicesClone
	Close                 IndicesClose
	Create                IndicesCreate
	DeleteAlias           IndicesDeleteAlias
	DeleteIndexTemplate   IndicesDeleteIndexTemplate
	Delete                IndicesDelete
	DeleteTemplate        IndicesDeleteTemplate
	DiskUsage             IndicesDiskUsage
	ExistsAlias           IndicesExistsAlias
	ExistsDocumentType    IndicesExistsDocumentType
	ExistsIndexTemplate   IndicesExistsIndexTemplate
	Exists                IndicesExists
	ExistsTemplate        IndicesExistsTemplate
	FieldUsageStats       IndicesFieldUsageStats
	Flush                 IndicesFlush
	FlushSynced           IndicesFlushSynced
	Forcemerge            IndicesForcemerge
	GetAlias              IndicesGetAlias
	GetFieldMapping       IndicesGetFieldMapping
	GetIndexTemplate      IndicesGetIndexTemplate
	GetMapping            IndicesGetMapping
	Get                   IndicesGet
	GetSettings           IndicesGetSettings
	GetTemplate           IndicesGetTemplate
	GetUpgrade            IndicesGetUpgrade
	Open                  IndicesOpen
	PutAlias              IndicesPutAlias
	PutIndexTemplate      IndicesPutIndexTemplate
	PutMapping            IndicesPutMapping
	PutSettings           IndicesPutSettings
	PutTemplate           IndicesPutTemplate
	Recovery              IndicesRecovery
	Refresh               IndicesRefresh
	ResolveIndex          IndicesResolveIndex
	Rollover              IndicesRollover
	Segments              IndicesSegments
	ShardStores           IndicesShardStores
	Shrink                IndicesShrink
	SimulateIndexTemplate IndicesSimulateIndexTemplate
	SimulateTemplate      IndicesSimulateTemplate
	Split                 IndicesSplit
	Stats                 IndicesStats
	UpdateAliases         IndicesUpdateAliases
	Upgrade               IndicesUpgrade
	ValidateQuery         IndicesValidateQuery
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
	CleanupRepository SnapshotCleanupRepository
	Clone             SnapshotClone
	CreateRepository  SnapshotCreateRepository
	Create            SnapshotCreate
	DeleteRepository  SnapshotDeleteRepository
	Delete            SnapshotDelete
	GetRepository     SnapshotGetRepository
	Get               SnapshotGet
	Restore           SnapshotRestore
	Status            SnapshotStatus
	VerifyRepository  SnapshotVerifyRepository
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
		Bulk:                               newBulkFunc(t),
		ClearScroll:                        newClearScrollFunc(t),
		Count:                              newCountFunc(t),
		Create:                             newCreateFunc(t),
		DanglingIndicesDeleteDanglingIndex: newDanglingIndicesDeleteDanglingIndexFunc(t),
		DanglingIndicesImportDanglingIndex: newDanglingIndicesImportDanglingIndexFunc(t),
		DanglingIndicesListDanglingIndices: newDanglingIndicesListDanglingIndicesFunc(t),
		DeleteByQuery:                      newDeleteByQueryFunc(t),
		DeleteByQueryRethrottle:            newDeleteByQueryRethrottleFunc(t),
		Delete:                             newDeleteFunc(t),
		DeleteScript:                       newDeleteScriptFunc(t),
		Exists:                             newExistsFunc(t),
		ExistsSource:                       newExistsSourceFunc(t),
		Explain:                            newExplainFunc(t),
		FieldCaps:                          newFieldCapsFunc(t),
		Get:                                newGetFunc(t),
		GetScriptContext:                   newGetScriptContextFunc(t),
		GetScriptLanguages:                 newGetScriptLanguagesFunc(t),
		GetScript:                          newGetScriptFunc(t),
		GetSource:                          newGetSourceFunc(t),
		Index:                              newIndexFunc(t),
		Info:                               newInfoFunc(t),
		Mget:                               newMgetFunc(t),
		Msearch:                            newMsearchFunc(t),
		MsearchTemplate:                    newMsearchTemplateFunc(t),
		Mtermvectors:                       newMtermvectorsFunc(t),
		Ping:                               newPingFunc(t),
		PutScript:                          newPutScriptFunc(t),
		RankEval:                           newRankEvalFunc(t),
		Reindex:                            newReindexFunc(t),
		ReindexRethrottle:                  newReindexRethrottleFunc(t),
		RenderSearchTemplate:               newRenderSearchTemplateFunc(t),
		ScriptsPainlessExecute:             newScriptsPainlessExecuteFunc(t),
		Scroll:                             newScrollFunc(t),
		Search:                             newSearchFunc(t),
		SearchShards:                       newSearchShardsFunc(t),
		SearchTemplate:                     newSearchTemplateFunc(t),
		TermsEnum:                          newTermsEnumFunc(t),
		Termvectors:                        newTermvectorsFunc(t),
		UpdateByQuery:                      newUpdateByQueryFunc(t),
		UpdateByQueryRethrottle:            newUpdateByQueryRethrottleFunc(t),
		Update:                             newUpdateFunc(t),
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
			AllocationExplain:            newClusterAllocationExplainFunc(t),
			DeleteComponentTemplate:      newClusterDeleteComponentTemplateFunc(t),
			DeleteVotingConfigExclusions: newClusterDeleteVotingConfigExclusionsFunc(t),
			ExistsComponentTemplate:      newClusterExistsComponentTemplateFunc(t),
			GetComponentTemplate:         newClusterGetComponentTemplateFunc(t),
			GetSettings:                  newClusterGetSettingsFunc(t),
			Health:                       newClusterHealthFunc(t),
			PendingTasks:                 newClusterPendingTasksFunc(t),
			PostVotingConfigExclusions:   newClusterPostVotingConfigExclusionsFunc(t),
			PutComponentTemplate:         newClusterPutComponentTemplateFunc(t),
			PutSettings:                  newClusterPutSettingsFunc(t),
			RemoteInfo:                   newClusterRemoteInfoFunc(t),
			Reroute:                      newClusterRerouteFunc(t),
			State:                        newClusterStateFunc(t),
			Stats:                        newClusterStatsFunc(t),
		},
		Indices: &Indices{
			AddBlock:              newIndicesAddBlockFunc(t),
			Analyze:               newIndicesAnalyzeFunc(t),
			ClearCache:            newIndicesClearCacheFunc(t),
			Clone:                 newIndicesCloneFunc(t),
			Close:                 newIndicesCloseFunc(t),
			Create:                newIndicesCreateFunc(t),
			DeleteAlias:           newIndicesDeleteAliasFunc(t),
			DeleteIndexTemplate:   newIndicesDeleteIndexTemplateFunc(t),
			Delete:                newIndicesDeleteFunc(t),
			DeleteTemplate:        newIndicesDeleteTemplateFunc(t),
			DiskUsage:             newIndicesDiskUsageFunc(t),
			ExistsAlias:           newIndicesExistsAliasFunc(t),
			ExistsDocumentType:    newIndicesExistsDocumentTypeFunc(t),
			ExistsIndexTemplate:   newIndicesExistsIndexTemplateFunc(t),
			Exists:                newIndicesExistsFunc(t),
			ExistsTemplate:        newIndicesExistsTemplateFunc(t),
			FieldUsageStats:       newIndicesFieldUsageStatsFunc(t),
			Flush:                 newIndicesFlushFunc(t),
			FlushSynced:           newIndicesFlushSyncedFunc(t),
			Forcemerge:            newIndicesForcemergeFunc(t),
			GetAlias:              newIndicesGetAliasFunc(t),
			GetFieldMapping:       newIndicesGetFieldMappingFunc(t),
			GetIndexTemplate:      newIndicesGetIndexTemplateFunc(t),
			GetMapping:            newIndicesGetMappingFunc(t),
			Get:                   newIndicesGetFunc(t),
			GetSettings:           newIndicesGetSettingsFunc(t),
			GetTemplate:           newIndicesGetTemplateFunc(t),
			GetUpgrade:            newIndicesGetUpgradeFunc(t),
			Open:                  newIndicesOpenFunc(t),
			PutAlias:              newIndicesPutAliasFunc(t),
			PutIndexTemplate:      newIndicesPutIndexTemplateFunc(t),
			PutMapping:            newIndicesPutMappingFunc(t),
			PutSettings:           newIndicesPutSettingsFunc(t),
			PutTemplate:           newIndicesPutTemplateFunc(t),
			Recovery:              newIndicesRecoveryFunc(t),
			Refresh:               newIndicesRefreshFunc(t),
			ResolveIndex:          newIndicesResolveIndexFunc(t),
			Rollover:              newIndicesRolloverFunc(t),
			Segments:              newIndicesSegmentsFunc(t),
			ShardStores:           newIndicesShardStoresFunc(t),
			Shrink:                newIndicesShrinkFunc(t),
			SimulateIndexTemplate: newIndicesSimulateIndexTemplateFunc(t),
			SimulateTemplate:      newIndicesSimulateTemplateFunc(t),
			Split:                 newIndicesSplitFunc(t),
			Stats:                 newIndicesStatsFunc(t),
			UpdateAliases:         newIndicesUpdateAliasesFunc(t),
			Upgrade:               newIndicesUpgradeFunc(t),
			ValidateQuery:         newIndicesValidateQueryFunc(t),
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
			CleanupRepository: newSnapshotCleanupRepositoryFunc(t),
			Clone:             newSnapshotCloneFunc(t),
			CreateRepository:  newSnapshotCreateRepositoryFunc(t),
			Create:            newSnapshotCreateFunc(t),
			DeleteRepository:  newSnapshotDeleteRepositoryFunc(t),
			Delete:            newSnapshotDeleteFunc(t),
			GetRepository:     newSnapshotGetRepositoryFunc(t),
			Get:               newSnapshotGetFunc(t),
			Restore:           newSnapshotRestoreFunc(t),
			Status:            newSnapshotStatusFunc(t),
			VerifyRepository:  newSnapshotVerifyRepositoryFunc(t),
		},
		Tasks: &Tasks{
			Cancel: newTasksCancelFunc(t),
			Get:    newTasksGetFunc(t),
			List:   newTasksListFunc(t),
		},
	}
}
