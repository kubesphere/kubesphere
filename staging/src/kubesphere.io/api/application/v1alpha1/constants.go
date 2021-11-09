/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

const (
	MsgLen               = 512
	HelmRepoSyncStateLen = 10

	// app version state
	StateDraft     = "draft"
	StateSubmitted = "submitted"
	StatePassed    = "passed"
	StateRejected  = "rejected"
	StateSuspended = "suspended"
	StateActive    = "active"

	// repo state
	RepoStateSuccessful = "successful"
	RepoStateFailed     = "failed"
	RepoStateSyncing    = "syncing"

	// helm release state
	HelmStatusActive      = "active"
	HelmStatusCreating    = "creating"
	HelmStatusDeleting    = "deleting"
	HelmStatusUpgrading   = "upgrading"
	HelmStatusRollbacking = "rollbacking"
	HelmStatusFailed      = "failed"
	HelmStatusCreated     = "created"
	HelmStatusUpgraded    = "upgraded"

	AttachmentTypeScreenshot = "screenshot"
	AttachmentTypeIcon       = "icon"

	HelmApplicationAppStoreSuffix  = "-store"
	HelmApplicationIdPrefix        = "app-"
	HelmRepoIdPrefix               = "repo-"
	HelmApplicationVersionIdPrefix = "appv-"
	HelmCategoryIdPrefix           = "ctg-"
	HelmAttachmentPrefix           = "att-"
	HelmReleasePrefix              = "rls-"
	UncategorizedName              = "uncategorized"
	UncategorizedId                = "ctg-uncategorized"
	AppStoreRepoId                 = "repo-helm"

	ApplicationInstance = "app.kubesphere.io/instance"

	OriginWorkspaceLabelKey = "kubesphere.io/workspace-origin"

	// operator app
	OperatorApplicationSuffix = "-operator"
	OperatorAppLabelKey       = "application.kubesphere.io/operator-app"
	// custom resource status
	ManifestCreating = "ManifestCreating"
	ManifestCreated  = "ManifestCreated"
	Failed           = "Failed"
	Error            = "Error"

	// front state
	FrontCreating       string = "Creating"
	FrontUpdating       string = "InProgress"
	FrontCompleted      string = "Completed"
	FrontRunning        string = "Running"
	FrontClosed         string = "Closed"
	FrontCreateFailed   string = "CreateFailed"
	FrontUpdateFailed   string = "UpdateFailed"
	FrontTerminating    string = "Terminating"
	StatusBootstrapping string = "Bootstrapping"
	StatusBootstrapped  string = "Bootstrapped"
	StatusRestoring     string = "Restoring"

	// MySQL state
	// ClusterInitState  indicates whether the cluster is initializing.
	ClusterInitState string = "Initializing"
	// ClusterUpdateState indicates whether the cluster is being updated.
	ClusterUpdateState string = "Updating"
	// ClusterReadyState indicates whether all containers in the pod are ready.
	ClusterReadyState string = "Ready"
	// ClusterCloseState indicates whether the cluster is closed.
	ClusterCloseState string = "Closed"

	// ClickHouse state
	StatusCreating     = "Creating"
	StatusInProgress   = "InProgress"
	StatusCompleted    = "Completed"
	StatusRunning      = "Running"
	StatusCreateFailed = "CreateFailed"
	StatusUpdateFailed = "UpdateFailed"
	StatusTerminating  = "Terminating"

	// PostgreSQL state
	PgclusterStateCreated       string = "pgcluster Created"
	PgclusterStateProcessed     string = "pgcluster Processed"
	PgclusterStateInitialized   string = "pgcluster Initialized"
	PgclusterStateBootstrapping string = "pgcluster Bootstrapping"
	PgclusterStateBootstrapped  string = "pgcluster Bootstrapped"
	PgclusterStateRestore       string = "pgcluster Restoring"
	PgclusterStateShutdown      string = "pgcluster Shutdown"

	// kind of operator cr
	DBTypeMysql = "MySQL"

	// kind of cluster
	KindPostgreSQLCluster = "PostgreSQLCluster"
	KindMysqlCluster      = "MysqlCluster"
	KindClickHouseCluster = "ClickHouseInstallation"
	KindPgCluster         = "Pgcluster"
	KindPgClusterVersion  = "radondb.com/v1"

	// cluster status
	ClusterStatusUnknown = "unknown"
	// suffix of secret name
	SuffixSecretName = "-userpassword-secret"
)
