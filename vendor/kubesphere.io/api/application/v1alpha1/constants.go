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
	BuiltinRepoPrefix              = "builtin-"
	HelmApplicationVersionIdPrefix = "appv-"
	HelmCategoryIdPrefix           = "ctg-"
	HelmAttachmentPrefix           = "att-"
	HelmReleasePrefix              = "rls-"
	UncategorizedName              = "uncategorized"
	UncategorizedId                = "ctg-uncategorized"
	AppStoreRepoId                 = "repo-helm"

	ApplicationInstance = "app.kubesphere.io/instance"

	RepoSyncPeriod          = "app.kubesphere.io/sync-period"
	OriginWorkspaceLabelKey = "kubesphere.io/workspace-origin"
)
