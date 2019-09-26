// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package constants

const (
	ColumnAppId                    = "app_id"
	ColumnCategoryId               = "category_id"
	ColumnChartName                = "chart_name"
	ColumnClusterId                = "cluster_id"
	ColumnClusterType              = "cluster_type"
	ColumnCreateTime               = "create_time"
	ColumnCredential               = "credential"
	ColumnDescription              = "description"
	ColumnExecutor                 = "executor"
	ColumnFrontgateId              = "frontgate_id"
	ColumnHome                     = "home"
	ColumnIcon                     = "icon"
	ColumnInstanceId               = "instance_id"
	ColumnJobAction                = "job_action"
	ColumnJobId                    = "job_id"
	ColumnKeywords                 = "keywords"
	ColumnLabelKey                 = "label_key"
	ColumnLabelValue               = "label_value"
	ColumnLocale                   = "locale"
	ColumnMaintainers              = "maintainers"
	ColumnName                     = "name"
	ColumnNodeId                   = "node_id"
	ColumnKeyPairId                = "key_pair_id"
	ColumnOwner                    = "owner"
	ColumnOwnerPath                = "owner_path"
	ColumnPackageName              = "package_name"
	ColumnPrivateIp                = "private_ip"
	ColumnProvider                 = "provider"
	ColumnReadme                   = "readme"
	ColumnRepoEventId              = "repo_event_id"
	ColumnRepoId                   = "repo_id"
	ColumnRepoLabelId              = "repo_label_id"
	ColumnRepoSelectorId           = "repo_selector_id"
	ColumnResouceId                = "resource_id"
	ColumnRole                     = "role"
	ColumnRuntimeId                = "runtime_id"
	ColumnRuntimeCredentialId      = "runtime_credential_id"
	ColumnRuntimeUrl               = "runtime_url"
	ColumnDebug                    = "debug"
	ColumnRuntimeLabelId           = "runtime_label_id"
	ColumnScreenshots              = "screenshots"
	ColumnSelectorKey              = "selector_key"
	ColumnSelectorValue            = "selector_value"
	ColumnSequence                 = "sequence"
	ColumnSources                  = "sources"
	ColumnStatus                   = "status"
	ColumnStatusTime               = "status_time"
	ColumnTarget                   = "target"
	ColumnTaskAction               = "task_action"
	ColumnTaskId                   = "task_id"
	ColumnTransitionStatus         = "transition_status"
	ColumnType                     = "type"
	ColumnUpdateTime               = "update_time"
	ColumnUrl                      = "url"
	ColumnVersionId                = "version_id"
	ColumnVisibility               = "visibility"
	ColumnVolumeId                 = "volume_id"
	ColumnZone                     = "zone"
	ColumnVpcId                    = "vpc_id"
	ColumnEnv                      = "env"
	ColumnLoadbalancerListenerId   = "loadbalancer_listener_id"
	ColumnResult                   = "result"
	ColumnDirective                = "directive"
	ColumnRuntimeCredentialContent = "runtime_credential_content"
	ColumnUserId                   = "user_id"
	ColumnGroupId                  = "group_id"
	ColumnResetId                  = "reset_id"
	ColumnPassword                 = "password"
	ColumnEmail                    = "email"
	ColumnClientId                 = "client_id"
	ColumnClientSecret             = "client_secret"
	ColumnRefreshToken             = "refresh_token"
	ColumnAccessToken              = "access_token"
	ColumnTokenId                  = "token_id"
	ColumnScope                    = "scope"
	ColumnUsername                 = "username"
	ColumnAttachmentId             = "attachment_id"
	ColumnMessage                  = "message"
	ColumnAppDefaultStatus         = "app_default_status"
	ColumnMarketId                 = "market_id"
	ColumnController               = "controller"
	ColumnActive                   = "active"
	ColumnOperator                 = "operator"
	ColumnReviewId                 = "review_id"
	ColumnPhase                    = "phase"
	ColumnReviewer                 = "reviewer"
	ColumnCompanyName              = "company_name"
	ColumnCompanyWebsite           = "company_website"
	ColumnCompanyProfile           = "company_profile"
	ColumnAuthorizerName           = "authorizer_name"
	ColumnAuthorizerEmail          = "authorizer_email"
	ColumnAuthorizerPhone          = "authorizer_phone"
	ColumnBankName                 = "bank_name"
	ColumnBankAccountName          = "bank_account_name"
	ColumnBankAccountNumber        = "bank_account_number"
	ColumnRejectMessage            = "reject_message"
	ColumnSubmitTime               = "submit_time"
	ColumnApprover                 = "approver"
	ColumnIsv                      = "isv"
)

var PushEventTables = map[string][]string{
	TableRepoEvent: {
		ColumnRepoEventId, ColumnRepoId, ColumnStatus,
	},
	TableCluster: {
		ColumnClusterId, ColumnStatus, ColumnTransitionStatus,
	},
	TableClusterNode: {
		ColumnNodeId, ColumnStatus, ColumnTransitionStatus,
	},
	TableJob: {
		ColumnJobId, ColumnStatus, ColumnClusterId, ColumnAppId, ColumnAppId,
	},
}

// columns that can be search through sql '=' operator
var IndexedColumns = map[string][]string{
	TableApp: {
		ColumnAppId, ColumnName, ColumnRepoId, ColumnDescription, ColumnStatus,
		ColumnHome, ColumnIcon, ColumnScreenshots, ColumnMaintainers, ColumnSources,
		ColumnReadme, ColumnOwner, ColumnChartName, ColumnIsv,
	},
	TableAppVersion: {
		ColumnVersionId, ColumnAppId, ColumnName, ColumnOwner, ColumnDescription,
		ColumnPackageName, ColumnStatus, ColumnType,
	},
	TableJob: {
		ColumnJobId, ColumnClusterId, ColumnAppId, ColumnVersionId,
		ColumnExecutor, ColumnProvider, ColumnStatus, ColumnOwner,
	},
	TableTask: {
		ColumnJobId, ColumnTaskId, ColumnExecutor, ColumnStatus, ColumnOwner,
	},
	TableRepo: {
		ColumnRepoId, ColumnName, ColumnType, ColumnVisibility, ColumnStatus,
		ColumnAppDefaultStatus, ColumnOwner, ColumnController,
	},
	TableRuntime: {
		ColumnRuntimeId, ColumnProvider, ColumnZone, ColumnStatus, ColumnOwner,
	},
	TableRuntimeCredential: {
		ColumnRuntimeCredentialId, ColumnStatus, ColumnProvider, ColumnOwner,
	},
	TableRepoLabel: {
		ColumnRepoId, ColumnRepoLabelId, ColumnStatus,
	},
	TableRepoSelector: {
		ColumnRepoId, ColumnRepoSelectorId, ColumnStatus,
	},
	TableRepoEvent: {
		ColumnRepoEventId, ColumnRepoId, ColumnStatus, ColumnOwner,
	},
	TableCluster: {
		ColumnClusterId, ColumnAppId, ColumnVersionId, ColumnStatus,
		ColumnRuntimeId, ColumnFrontgateId, ColumnOwner, ColumnClusterType,
	},
	TableKeyPair: {
		ColumnKeyPairId, ColumnName, ColumnOwner,
	},
	TableClusterNode: {
		ColumnClusterId, ColumnNodeId, ColumnStatus, ColumnOwner,
	},
	TableCategory: {
		ColumnCategoryId, ColumnStatus, ColumnLocale, ColumnOwner, ColumnName,
	},
	TableMarket: {
		ColumnMarketId, ColumnName, ColumnVisibility, ColumnStatus, ColumnOwner,
	},
	TableMarketUser: {
		ColumnMarketId, ColumnUserId,
	},
	TableAppVersionAudit: {
		ColumnVersionId, ColumnAppId, ColumnStatus, ColumnOperator, ColumnRole,
	},
	TableAppVersionReview: {
		ColumnReviewId, ColumnVersionId, ColumnAppId, ColumnStatus, ColumnReviewer,
	},
	TableVendorVerifyInfo: {
		ColumnUserId, ColumnStatus,
	},
}

var SearchWordColumnTable = []string{
	TableRuntime,
	TableRuntimeCredential,
	TableApp,
	TableAppVersion,
	TableRepo,
	TableJob,
	TableTask,
	TableCluster,
	TableClusterNode,
	TableCategory,
	TableVendorVerifyInfo,
}

// columns that can be search through sql 'like' operator
var SearchColumns = map[string][]string{
	TableApp: {
		ColumnAppId, ColumnName, ColumnRepoId, ColumnOwner, ColumnChartName, ColumnKeywords,
	},
	TableAppVersion: {
		ColumnVersionId, ColumnAppId, ColumnName, ColumnDescription, ColumnOwner, ColumnPackageName,
	},
	TableJob: {
		ColumnJobId, ColumnClusterId, ColumnOwner, ColumnJobAction, ColumnExecutor, ColumnProvider, ColumnExecutor, ColumnProvider,
	},
	TableTask: {
		ColumnJobId, ColumnTaskId, ColumnTaskAction, ColumnOwner, ColumnNodeId, ColumnTarget,
	},
	TableRuntime: {
		ColumnRuntimeId, ColumnName, ColumnOwner, ColumnProvider, ColumnZone,
	},
	TableRuntimeCredential: {
		ColumnRuntimeCredentialId, ColumnName, ColumnOwner, ColumnProvider,
	},
	TableCluster: {
		ColumnClusterId, ColumnName, ColumnOwner, ColumnAppId, ColumnVersionId, ColumnRuntimeId,
	},
	TableClusterNode: {
		ColumnNodeId, ColumnClusterId, ColumnName, ColumnInstanceId, ColumnVolumeId, ColumnPrivateIp, ColumnRole, ColumnOwner,
	},
	TableRepo: {
		ColumnName, ColumnDescription,
	},
	TableCategory: {
		ColumnCategoryId, ColumnLocale, ColumnOwner, ColumnName,
	},
	TableVendorVerifyInfo: {
		ColumnUserId, ColumnCompanyName, ColumnCompanyWebsite, ColumnAuthorizerName, ColumnAuthorizerEmail,
	},
}
