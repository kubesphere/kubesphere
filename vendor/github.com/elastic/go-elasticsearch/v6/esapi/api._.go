// Code generated from specification version 6.8.2-SNAPSHOT (8aa0ff06c9d): DO NOT EDIT

package esapi

// API contains the Elasticsearch APIs
//
type API struct {
	Cat        *Cat
	Cluster    *Cluster
	Indices    *Indices
	Ingest     *Ingest
	Nodes      *Nodes
	Remote     *Remote
	Snapshot   *Snapshot
	Tasks      *Tasks
	CCR        *CCR
	ILM        *ILM
	License    *License
	Migration  *Migration
	ML         *ML
	Monitoring *Monitoring
	Rollup     *Rollup
	Security   *Security
	SQL        *SQL
	SSL        *SSL
	Watcher    *Watcher
	XPack      *XPack

	Bulk                    Bulk
	ClearScroll             ClearScroll
	Count                   Count
	Create                  Create
	DeleteByQuery           DeleteByQuery
	DeleteByQueryRethrottle DeleteByQueryRethrottle
	Delete                  Delete
	DeleteScript            DeleteScript
	Exists                  Exists
	ExistsSource            ExistsSource
	Explain                 Explain
	FieldCaps               FieldCaps
	Get                     Get
	GetScript               GetScript
	GetSource               GetSource
	Index                   Index
	Info                    Info
	Mget                    Mget
	Msearch                 Msearch
	MsearchTemplate         MsearchTemplate
	Mtermvectors            Mtermvectors
	Ping                    Ping
	PutScript               PutScript
	RankEval                RankEval
	Reindex                 Reindex
	ReindexRethrottle       ReindexRethrottle
	RenderSearchTemplate    RenderSearchTemplate
	ScriptsPainlessExecute  ScriptsPainlessExecute
	Scroll                  Scroll
	Search                  Search
	SearchShards            SearchShards
	SearchTemplate          SearchTemplate
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
	Freeze             IndicesFreeze
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
	Unfreeze           IndicesUnfreeze
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

// CCR contains the CCR APIs
type CCR struct {
	DeleteAutoFollowPattern CCRDeleteAutoFollowPattern
	FollowInfo              CCRFollowInfo
	Follow                  CCRFollow
	FollowStats             CCRFollowStats
	ForgetFollower          CCRForgetFollower
	GetAutoFollowPattern    CCRGetAutoFollowPattern
	PauseFollow             CCRPauseFollow
	PutAutoFollowPattern    CCRPutAutoFollowPattern
	ResumeFollow            CCRResumeFollow
	Stats                   CCRStats
	Unfollow                CCRUnfollow
}

// ILM contains the ILM APIs
type ILM struct {
	DeleteLifecycle  ILMDeleteLifecycle
	ExplainLifecycle ILMExplainLifecycle
	GetLifecycle     ILMGetLifecycle
	GetStatus        ILMGetStatus
	MoveToStep       ILMMoveToStep
	PutLifecycle     ILMPutLifecycle
	RemovePolicy     ILMRemovePolicy
	Retry            ILMRetry
	Start            ILMStart
	Stop             ILMStop
}

// License contains the License APIs
type License struct {
}

// Migration contains the Migration APIs
type Migration struct {
}

// ML contains the ML APIs
type ML struct {
}

// Monitoring contains the Monitoring APIs
type Monitoring struct {
}

// Rollup contains the Rollup APIs
type Rollup struct {
}

// Security contains the Security APIs
type Security struct {
	CreateAPIKey     SecurityCreateAPIKey
	GetAPIKey        SecurityGetAPIKey
	InvalidateAPIKey SecurityInvalidateAPIKey
}

// SQL contains the SQL APIs
type SQL struct {
}

// SSL contains the SSL APIs
type SSL struct {
}

// Watcher contains the Watcher APIs
type Watcher struct {
}

// XPack contains the XPack APIs
type XPack struct {
	GraphExplore              XPackGraphExplore
	Info                      XPackInfo
	LicenseDelete             XPackLicenseDelete
	LicenseGetBasicStatus     XPackLicenseGetBasicStatus
	LicenseGet                XPackLicenseGet
	LicenseGetTrialStatus     XPackLicenseGetTrialStatus
	LicensePost               XPackLicensePost
	LicensePostStartBasic     XPackLicensePostStartBasic
	LicensePostStartTrial     XPackLicensePostStartTrial
	MLCloseJob                XPackMLCloseJob
	MLDeleteCalendarEvent     XPackMLDeleteCalendarEvent
	MLDeleteCalendarJob       XPackMLDeleteCalendarJob
	MLDeleteCalendar          XPackMLDeleteCalendar
	MLDeleteDatafeed          XPackMLDeleteDatafeed
	MLDeleteExpiredData       XPackMLDeleteExpiredData
	MLDeleteFilter            XPackMLDeleteFilter
	MLDeleteForecast          XPackMLDeleteForecast
	MLDeleteJob               XPackMLDeleteJob
	MLDeleteModelSnapshot     XPackMLDeleteModelSnapshot
	MLFindFileStructure       XPackMLFindFileStructure
	MLFlushJob                XPackMLFlushJob
	MLForecast                XPackMLForecast
	MLGetBuckets              XPackMLGetBuckets
	MLGetCalendarEvents       XPackMLGetCalendarEvents
	MLGetCalendars            XPackMLGetCalendars
	MLGetCategories           XPackMLGetCategories
	MLGetDatafeedStats        XPackMLGetDatafeedStats
	MLGetDatafeeds            XPackMLGetDatafeeds
	MLGetFilters              XPackMLGetFilters
	MLGetInfluencers          XPackMLGetInfluencers
	MLGetJobStats             XPackMLGetJobStats
	MLGetJobs                 XPackMLGetJobs
	MLGetModelSnapshots       XPackMLGetModelSnapshots
	MLGetOverallBuckets       XPackMLGetOverallBuckets
	MLGetRecords              XPackMLGetRecords
	MLInfo                    XPackMLInfo
	MLOpenJob                 XPackMLOpenJob
	MLPostCalendarEvents      XPackMLPostCalendarEvents
	MLPostData                XPackMLPostData
	MLPreviewDatafeed         XPackMLPreviewDatafeed
	MLPutCalendarJob          XPackMLPutCalendarJob
	MLPutCalendar             XPackMLPutCalendar
	MLPutDatafeed             XPackMLPutDatafeed
	MLPutFilter               XPackMLPutFilter
	MLPutJob                  XPackMLPutJob
	MLRevertModelSnapshot     XPackMLRevertModelSnapshot
	MLSetUpgradeMode          XPackMLSetUpgradeMode
	MLStartDatafeed           XPackMLStartDatafeed
	MLStopDatafeed            XPackMLStopDatafeed
	MLUpdateDatafeed          XPackMLUpdateDatafeed
	MLUpdateFilter            XPackMLUpdateFilter
	MLUpdateJob               XPackMLUpdateJob
	MLUpdateModelSnapshot     XPackMLUpdateModelSnapshot
	MLValidateDetector        XPackMLValidateDetector
	MLValidate                XPackMLValidate
	MigrationDeprecations     XPackMigrationDeprecations
	MigrationGetAssistance    XPackMigrationGetAssistance
	MigrationUpgrade          XPackMigrationUpgrade
	MonitoringBulk            XPackMonitoringBulk
	RollupDeleteJob           XPackRollupDeleteJob
	RollupGetJobs             XPackRollupGetJobs
	RollupGetRollupCaps       XPackRollupGetRollupCaps
	RollupGetRollupIndexCaps  XPackRollupGetRollupIndexCaps
	RollupPutJob              XPackRollupPutJob
	RollupRollupSearch        XPackRollupRollupSearch
	RollupStartJob            XPackRollupStartJob
	RollupStopJob             XPackRollupStopJob
	SQLClearCursor            XPackSQLClearCursor
	SQLQuery                  XPackSQLQuery
	SQLTranslate              XPackSQLTranslate
	SSLCertificates           XPackSSLCertificates
	SecurityAuthenticate      XPackSecurityAuthenticate
	SecurityChangePassword    XPackSecurityChangePassword
	SecurityClearCachedRealms XPackSecurityClearCachedRealms
	SecurityClearCachedRoles  XPackSecurityClearCachedRoles
	SecurityDeletePrivileges  XPackSecurityDeletePrivileges
	SecurityDeleteRoleMapping XPackSecurityDeleteRoleMapping
	SecurityDeleteRole        XPackSecurityDeleteRole
	SecurityDeleteUser        XPackSecurityDeleteUser
	SecurityDisableUser       XPackSecurityDisableUser
	SecurityEnableUser        XPackSecurityEnableUser
	SecurityGetPrivileges     XPackSecurityGetPrivileges
	SecurityGetRoleMapping    XPackSecurityGetRoleMapping
	SecurityGetRole           XPackSecurityGetRole
	SecurityGetToken          XPackSecurityGetToken
	SecurityGetUserPrivileges XPackSecurityGetUserPrivileges
	SecurityGetUser           XPackSecurityGetUser
	SecurityHasPrivileges     XPackSecurityHasPrivileges
	SecurityInvalidateToken   XPackSecurityInvalidateToken
	SecurityPutPrivileges     XPackSecurityPutPrivileges
	SecurityPutRoleMapping    XPackSecurityPutRoleMapping
	SecurityPutRole           XPackSecurityPutRole
	SecurityPutUser           XPackSecurityPutUser
	Usage                     XPackUsage
	WatcherAckWatch           XPackWatcherAckWatch
	WatcherActivateWatch      XPackWatcherActivateWatch
	WatcherDeactivateWatch    XPackWatcherDeactivateWatch
	WatcherDeleteWatch        XPackWatcherDeleteWatch
	WatcherExecuteWatch       XPackWatcherExecuteWatch
	WatcherGetWatch           XPackWatcherGetWatch
	WatcherPutWatch           XPackWatcherPutWatch
	WatcherRestart            XPackWatcherRestart
	WatcherStart              XPackWatcherStart
	WatcherStats              XPackWatcherStats
	WatcherStop               XPackWatcherStop
}

// New creates new API
//
func New(t Transport) *API {
	return &API{
		Bulk:                    newBulkFunc(t),
		ClearScroll:             newClearScrollFunc(t),
		Count:                   newCountFunc(t),
		Create:                  newCreateFunc(t),
		DeleteByQuery:           newDeleteByQueryFunc(t),
		DeleteByQueryRethrottle: newDeleteByQueryRethrottleFunc(t),
		Delete:                  newDeleteFunc(t),
		DeleteScript:            newDeleteScriptFunc(t),
		Exists:                  newExistsFunc(t),
		ExistsSource:            newExistsSourceFunc(t),
		Explain:                 newExplainFunc(t),
		FieldCaps:               newFieldCapsFunc(t),
		Get:                     newGetFunc(t),
		GetScript:               newGetScriptFunc(t),
		GetSource:               newGetSourceFunc(t),
		Index:                   newIndexFunc(t),
		Info:                    newInfoFunc(t),
		Mget:                    newMgetFunc(t),
		Msearch:                 newMsearchFunc(t),
		MsearchTemplate:         newMsearchTemplateFunc(t),
		Mtermvectors:            newMtermvectorsFunc(t),
		Ping:                    newPingFunc(t),
		PutScript:               newPutScriptFunc(t),
		RankEval:                newRankEvalFunc(t),
		Reindex:                 newReindexFunc(t),
		ReindexRethrottle:       newReindexRethrottleFunc(t),
		RenderSearchTemplate:    newRenderSearchTemplateFunc(t),
		ScriptsPainlessExecute:  newScriptsPainlessExecuteFunc(t),
		Scroll:                  newScrollFunc(t),
		Search:                  newSearchFunc(t),
		SearchShards:            newSearchShardsFunc(t),
		SearchTemplate:          newSearchTemplateFunc(t),
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
			Freeze:             newIndicesFreezeFunc(t),
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
			Unfreeze:           newIndicesUnfreezeFunc(t),
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
		CCR: &CCR{
			DeleteAutoFollowPattern: newCCRDeleteAutoFollowPatternFunc(t),
			FollowInfo:              newCCRFollowInfoFunc(t),
			Follow:                  newCCRFollowFunc(t),
			FollowStats:             newCCRFollowStatsFunc(t),
			ForgetFollower:          newCCRForgetFollowerFunc(t),
			GetAutoFollowPattern:    newCCRGetAutoFollowPatternFunc(t),
			PauseFollow:             newCCRPauseFollowFunc(t),
			PutAutoFollowPattern:    newCCRPutAutoFollowPatternFunc(t),
			ResumeFollow:            newCCRResumeFollowFunc(t),
			Stats:                   newCCRStatsFunc(t),
			Unfollow:                newCCRUnfollowFunc(t),
		},
		ILM: &ILM{
			DeleteLifecycle:  newILMDeleteLifecycleFunc(t),
			ExplainLifecycle: newILMExplainLifecycleFunc(t),
			GetLifecycle:     newILMGetLifecycleFunc(t),
			GetStatus:        newILMGetStatusFunc(t),
			MoveToStep:       newILMMoveToStepFunc(t),
			PutLifecycle:     newILMPutLifecycleFunc(t),
			RemovePolicy:     newILMRemovePolicyFunc(t),
			Retry:            newILMRetryFunc(t),
			Start:            newILMStartFunc(t),
			Stop:             newILMStopFunc(t),
		},
		License:    &License{},
		Migration:  &Migration{},
		ML:         &ML{},
		Monitoring: &Monitoring{},
		Rollup:     &Rollup{},
		Security: &Security{
			CreateAPIKey:     newSecurityCreateAPIKeyFunc(t),
			GetAPIKey:        newSecurityGetAPIKeyFunc(t),
			InvalidateAPIKey: newSecurityInvalidateAPIKeyFunc(t),
		},
		SQL:     &SQL{},
		SSL:     &SSL{},
		Watcher: &Watcher{},
		XPack: &XPack{
			GraphExplore:              newXPackGraphExploreFunc(t),
			Info:                      newXPackInfoFunc(t),
			LicenseDelete:             newXPackLicenseDeleteFunc(t),
			LicenseGetBasicStatus:     newXPackLicenseGetBasicStatusFunc(t),
			LicenseGet:                newXPackLicenseGetFunc(t),
			LicenseGetTrialStatus:     newXPackLicenseGetTrialStatusFunc(t),
			LicensePost:               newXPackLicensePostFunc(t),
			LicensePostStartBasic:     newXPackLicensePostStartBasicFunc(t),
			LicensePostStartTrial:     newXPackLicensePostStartTrialFunc(t),
			MLCloseJob:                newXPackMLCloseJobFunc(t),
			MLDeleteCalendarEvent:     newXPackMLDeleteCalendarEventFunc(t),
			MLDeleteCalendarJob:       newXPackMLDeleteCalendarJobFunc(t),
			MLDeleteCalendar:          newXPackMLDeleteCalendarFunc(t),
			MLDeleteDatafeed:          newXPackMLDeleteDatafeedFunc(t),
			MLDeleteExpiredData:       newXPackMLDeleteExpiredDataFunc(t),
			MLDeleteFilter:            newXPackMLDeleteFilterFunc(t),
			MLDeleteForecast:          newXPackMLDeleteForecastFunc(t),
			MLDeleteJob:               newXPackMLDeleteJobFunc(t),
			MLDeleteModelSnapshot:     newXPackMLDeleteModelSnapshotFunc(t),
			MLFindFileStructure:       newXPackMLFindFileStructureFunc(t),
			MLFlushJob:                newXPackMLFlushJobFunc(t),
			MLForecast:                newXPackMLForecastFunc(t),
			MLGetBuckets:              newXPackMLGetBucketsFunc(t),
			MLGetCalendarEvents:       newXPackMLGetCalendarEventsFunc(t),
			MLGetCalendars:            newXPackMLGetCalendarsFunc(t),
			MLGetCategories:           newXPackMLGetCategoriesFunc(t),
			MLGetDatafeedStats:        newXPackMLGetDatafeedStatsFunc(t),
			MLGetDatafeeds:            newXPackMLGetDatafeedsFunc(t),
			MLGetFilters:              newXPackMLGetFiltersFunc(t),
			MLGetInfluencers:          newXPackMLGetInfluencersFunc(t),
			MLGetJobStats:             newXPackMLGetJobStatsFunc(t),
			MLGetJobs:                 newXPackMLGetJobsFunc(t),
			MLGetModelSnapshots:       newXPackMLGetModelSnapshotsFunc(t),
			MLGetOverallBuckets:       newXPackMLGetOverallBucketsFunc(t),
			MLGetRecords:              newXPackMLGetRecordsFunc(t),
			MLInfo:                    newXPackMLInfoFunc(t),
			MLOpenJob:                 newXPackMLOpenJobFunc(t),
			MLPostCalendarEvents:      newXPackMLPostCalendarEventsFunc(t),
			MLPostData:                newXPackMLPostDataFunc(t),
			MLPreviewDatafeed:         newXPackMLPreviewDatafeedFunc(t),
			MLPutCalendarJob:          newXPackMLPutCalendarJobFunc(t),
			MLPutCalendar:             newXPackMLPutCalendarFunc(t),
			MLPutDatafeed:             newXPackMLPutDatafeedFunc(t),
			MLPutFilter:               newXPackMLPutFilterFunc(t),
			MLPutJob:                  newXPackMLPutJobFunc(t),
			MLRevertModelSnapshot:     newXPackMLRevertModelSnapshotFunc(t),
			MLSetUpgradeMode:          newXPackMLSetUpgradeModeFunc(t),
			MLStartDatafeed:           newXPackMLStartDatafeedFunc(t),
			MLStopDatafeed:            newXPackMLStopDatafeedFunc(t),
			MLUpdateDatafeed:          newXPackMLUpdateDatafeedFunc(t),
			MLUpdateFilter:            newXPackMLUpdateFilterFunc(t),
			MLUpdateJob:               newXPackMLUpdateJobFunc(t),
			MLUpdateModelSnapshot:     newXPackMLUpdateModelSnapshotFunc(t),
			MLValidateDetector:        newXPackMLValidateDetectorFunc(t),
			MLValidate:                newXPackMLValidateFunc(t),
			MigrationDeprecations:     newXPackMigrationDeprecationsFunc(t),
			MigrationGetAssistance:    newXPackMigrationGetAssistanceFunc(t),
			MigrationUpgrade:          newXPackMigrationUpgradeFunc(t),
			MonitoringBulk:            newXPackMonitoringBulkFunc(t),
			RollupDeleteJob:           newXPackRollupDeleteJobFunc(t),
			RollupGetJobs:             newXPackRollupGetJobsFunc(t),
			RollupGetRollupCaps:       newXPackRollupGetRollupCapsFunc(t),
			RollupGetRollupIndexCaps:  newXPackRollupGetRollupIndexCapsFunc(t),
			RollupPutJob:              newXPackRollupPutJobFunc(t),
			RollupRollupSearch:        newXPackRollupRollupSearchFunc(t),
			RollupStartJob:            newXPackRollupStartJobFunc(t),
			RollupStopJob:             newXPackRollupStopJobFunc(t),
			SQLClearCursor:            newXPackSQLClearCursorFunc(t),
			SQLQuery:                  newXPackSQLQueryFunc(t),
			SQLTranslate:              newXPackSQLTranslateFunc(t),
			SSLCertificates:           newXPackSSLCertificatesFunc(t),
			SecurityAuthenticate:      newXPackSecurityAuthenticateFunc(t),
			SecurityChangePassword:    newXPackSecurityChangePasswordFunc(t),
			SecurityClearCachedRealms: newXPackSecurityClearCachedRealmsFunc(t),
			SecurityClearCachedRoles:  newXPackSecurityClearCachedRolesFunc(t),
			SecurityDeletePrivileges:  newXPackSecurityDeletePrivilegesFunc(t),
			SecurityDeleteRoleMapping: newXPackSecurityDeleteRoleMappingFunc(t),
			SecurityDeleteRole:        newXPackSecurityDeleteRoleFunc(t),
			SecurityDeleteUser:        newXPackSecurityDeleteUserFunc(t),
			SecurityDisableUser:       newXPackSecurityDisableUserFunc(t),
			SecurityEnableUser:        newXPackSecurityEnableUserFunc(t),
			SecurityGetPrivileges:     newXPackSecurityGetPrivilegesFunc(t),
			SecurityGetRoleMapping:    newXPackSecurityGetRoleMappingFunc(t),
			SecurityGetRole:           newXPackSecurityGetRoleFunc(t),
			SecurityGetToken:          newXPackSecurityGetTokenFunc(t),
			SecurityGetUserPrivileges: newXPackSecurityGetUserPrivilegesFunc(t),
			SecurityGetUser:           newXPackSecurityGetUserFunc(t),
			SecurityHasPrivileges:     newXPackSecurityHasPrivilegesFunc(t),
			SecurityInvalidateToken:   newXPackSecurityInvalidateTokenFunc(t),
			SecurityPutPrivileges:     newXPackSecurityPutPrivilegesFunc(t),
			SecurityPutRoleMapping:    newXPackSecurityPutRoleMappingFunc(t),
			SecurityPutRole:           newXPackSecurityPutRoleFunc(t),
			SecurityPutUser:           newXPackSecurityPutUserFunc(t),
			Usage:                     newXPackUsageFunc(t),
			WatcherAckWatch:           newXPackWatcherAckWatchFunc(t),
			WatcherActivateWatch:      newXPackWatcherActivateWatchFunc(t),
			WatcherDeactivateWatch:    newXPackWatcherDeactivateWatchFunc(t),
			WatcherDeleteWatch:        newXPackWatcherDeleteWatchFunc(t),
			WatcherExecuteWatch:       newXPackWatcherExecuteWatchFunc(t),
			WatcherGetWatch:           newXPackWatcherGetWatchFunc(t),
			WatcherPutWatch:           newXPackWatcherPutWatchFunc(t),
			WatcherRestart:            newXPackWatcherRestartFunc(t),
			WatcherStart:              newXPackWatcherStartFunc(t),
			WatcherStats:              newXPackWatcherStatsFunc(t),
			WatcherStop:               newXPackWatcherStopFunc(t),
		},
	}
}
