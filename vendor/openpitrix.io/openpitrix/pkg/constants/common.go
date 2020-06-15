// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package constants

import (
	"time"
)

const (
	prefix                     = "openpitrix-"
	ProviderPrefix             = "openpitrix-rp-"
	ApiGatewayHost             = prefix + "api-gateway"
	RepoManagerHost            = prefix + "repo-manager"
	AppManagerHost             = prefix + "app-manager"
	RuntimeManagerHost         = prefix + "runtime-manager"
	ClusterManagerHost         = prefix + "cluster-manager"
	JobManagerHost             = prefix + "job-manager"
	TaskManagerHost            = prefix + "task-manager"
	PilotServiceHost           = prefix + "pilot-service"
	AccountServiceHost         = prefix + "account-service"
	IMServiceHost              = prefix + "im-service"
	AMServiceHost              = prefix + "am-service"
	RepoIndexerHost            = prefix + "repo-indexer"
	CategoryManagerHost        = prefix + "category-manager"
	RuntimeProviderManagerHost = prefix + "rp-manager"
	NotificationHost           = prefix + "notification"
	MarketManagerHost          = prefix + "market-manager"
	AttachmentManagerHost      = prefix + "attachment-manager"
	IsvManagerHost             = prefix + "isv-manager"
	ReleaseManagerHost         = prefix + "release-manager"
)

const (
	ApiGatewayPort             = 9100 // 91 is similar as Pi, Open[Pi]trix
	RepoManagerPort            = 9101
	AppManagerPort             = 9102
	RuntimeManagerPort         = 9103
	ClusterManagerPort         = 9104
	JobManagerPort             = 9106
	TaskManagerPort            = 9107
	RepoIndexerPort            = 9108
	PilotServicePort           = 9110
	FrontgateServicePort       = 9111
	DroneServicePort           = 9112
	CategoryManagerPort        = 9113
	PilotTlsListenPort         = 9114 // public service for frontgate
	AccountServicePort         = 9115
	FrontgateFileServerPort    = 9116
	MarketManagerPort          = 9117
	IsvManagerPort             = 9118
	IMServicePort              = 9119
	AMServicePort              = 9120
	EtcdServicePort            = 2379
	AttachmentManagerPort      = 9122
	RuntimeProviderManagerPort = 9121
	KubernetesProviderPort     = 9123
	ReleaseManagerPort         = 9124
	NotificationPort           = 9201
	ServiceConfigPort          = 9202
	ServicePushPort            = 9203
)

const (
	StatusActive      = "active"
	StatusUsed        = "used"
	StatusEnabled     = "enabled"
	StatusDisabled    = "disabled"
	StatusCreating    = "creating"
	StatusDeleted     = "deleted"
	StatusDeleting    = "deleting"
	StatusUpgrading   = "upgrading"
	StatusUpdating    = "updating"
	StatusRollbacking = "rollbacking"
	StatusStopped     = "stopped"
	StatusStopping    = "stopping"
	StatusStarting    = "starting"
	StatusRecovering  = "recovering"
	StatusCeased      = "ceased"
	StatusCeasing     = "ceasing"
	StatusResizing    = "resizing"
	StatusScaling     = "scaling"
	StatusWorking     = "working"
	StatusPending     = "pending"
	StatusSuccessful  = "successful"
	StatusFailed      = "failed"

	StatusRunning    = "running"
	StatusTerminated = "terminated"

	StatusAvailable = "available"
	StatusInUse     = "in-use"

	StatusInUse2 = "in_use"

	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusPassed    = "passed"
	StatusRejected  = "rejected"
	StatusSuspended = "suspended"
	StatusInReview  = "in-review"
	StatusNew       = "new"
)

var DeletedStatuses = []string{
	StatusDeleted,
	StatusCeased,
}

const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

const (
	DefaultMaxWorkingJobs  = 20
	DefaultMaxWorkingTasks = 20
	DefaultMaxRepoEvents   = 20
)

const (
	MaxTaskTimeout               = 3600 * time.Second
	WaitHelmTaskTimeout          = 7200 * time.Second
	WaitTaskTimeout              = 600 * time.Second
	WaitFrontgateServiceTimeout  = 1800 * time.Second
	WaitDroneServiceTimeout      = 1800 * time.Second
	WaitTaskInterval             = 3 * time.Second
	WaitFrontgateServiceInterval = 10 * time.Second
	WaitDroneServiceInterval     = 10 * time.Second

	GrpcToPilotTimeout = 10 * time.Second

	TimeoutName           = "timeout"
	DefaultServiceTimeout = 600

	// Maybe metadata is upgrading
	PilotTasksRetry = 5
	PilotTasksSleep = 2 * time.Second
)

const (
	ActionCreateCluster      = "CreateCluster"
	ActionUpgradeCluster     = "UpgradeCluster"
	ActionRollbackCluster    = "RollbackCluster"
	ActionResizeCluster      = "ResizeCluster"
	ActionAddClusterNodes    = "AddClusterNodes"
	ActionDeleteClusterNodes = "DeleteClusterNodes"
	ActionStopClusters       = "StopClusters"
	ActionStartClusters      = "StartClusters"
	ActionDeleteClusters     = "DeleteClusters"
	ActionRecoverClusters    = "RecoverClusters"
	ActionCeaseClusters      = "CeaseClusters"
	ActionUpdateClusterEnv   = "UpdateClusterEnv"
	ActionAttachKeyPairs     = "AttachKeyPairs"
	ActionDetachKeyPairs     = "DetachKeyPairs"
)

const (
	ProviderQingCloud   = "qingcloud"
	ProviderKubernetes  = "kubernetes"
	ProviderAWS         = "aws"
	ProviderAliyun      = "aliyun"
	ProviderTypeVmbased = "vmbased"
	TargetPilot         = "pilot"
)

const (
	PlaceHolder       = "*"
	ReplicaRoleSuffix = "-replica"
)

const (
	NodesToExecuteOnName    = "nodes_to_execute_on"
	PostStartServiceName    = "post_start_service"
	PostStopServiceName     = "post_stop_service"
	AgentInstalledName      = "agent_installed"
	ServiceOrderName        = "order"
	ServiceTimeoutName      = "timeout"
	ServiceCmdName          = "cmd"
	ServicePreCheckName     = "pre_check"
	ScalingPolicyParallel   = "parallel"
	ScalingPolicySequential = "sequential"

	NormalClusterType    = 0
	FrontgateClusterType = 1

	ServiceInit           = "init"
	ServiceStart          = "start"
	ServiceStop           = "stop"
	ServiceScaleIn        = "scale_in"
	ServiceScaleOut       = "scale_out"
	ServiceCustom         = "custom_service"
	ServiceRestart        = "restart"
	ServiceDestroy        = "destroy"
	ServiceBackup         = "backup"
	ServiceRestore        = "restore"
	ServiceDeleteSnapshot = "delete_snapshot"
	ServiceUpgrade        = "upgrade"
)

const (
	NfContentTypeInvite = "invite"
	NfContentTypeVerify = "verify"

	NfTypeEmail = "email"
)

var ServiceNames = []string{
	ServiceInit, ServiceStart, ServiceStop, ServiceScaleIn, ServiceScaleOut, ServiceRestart,
	ServiceDestroy, ServiceBackup, ServiceRestore, ServiceDeleteSnapshot, ServiceUpgrade,
}

const (
	TypeS3    = "s3"
	TypeHttp  = "http"
	TypeHttps = "https"
)

const (
	RetryInterval = 3 * time.Second
)

const (
	RoleUser        = "user"
	RoleDeveloper   = "developer"
	RoleIsv         = "isv"
	RoleGlobalAdmin = "global_admin"

	PortalGlobalAdmin = "global_admin"

	GrantTypeClientCredentials = "client_credentials"
	GrantTypePassword          = "password"
	GrantTypeRefreshToken      = "refresh_token"

	OperatorTypeGlobalAdmin = "global_admin"
	OperatorTypeDeveloper   = "developer"
	OperatorTypeBusiness    = "business"
	OperatorTypeTechnical   = "technical"
	OperatorTypeIsv         = "isv"
	OperatorTypeAdmin       = "admin"

	ActionBundleBusinessReview  = "business_review"
	ActionBundleTechnicalReview = "technical_review"
	ActionBundleIsvReview       = "isv_review"
	ActionBundleIsvApply        = "isv_apply"
	ActionBundleIsvAuth         = "isv_auth"
)

var GrantTypeTokens = []string{
	GrantTypeClientCredentials,
	GrantTypePassword,
	GrantTypeRefreshToken,
}

var InternalRepos = []string{
	"repo-vmbased", "repo-helm",
}

var AllowedAppDefaultStatus = []string{
	"",
	StatusDraft,
	StatusActive,
}

const (
	ServiceTypeNotification = "notification"
	ServiceTypeRuntime      = "runtime"
	ServiceTypeBasicConfig  = "basic_config"
)

var ServiceTypes = []string{
	ServiceTypeNotification,
	ServiceTypeRuntime,
	ServiceTypeBasicConfig,
}
