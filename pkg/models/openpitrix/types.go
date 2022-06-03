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

package openpitrix

import (
	"github.com/go-openapi/strfmt"
)

type ModifyAppRequest struct {
	// content of attachment
	AttachmentContent []byte `json:"attachment_content,omitempty"`

	// only for screenshot, range: [0, 5]
	Sequence *int32 `json:"sequence,omitempty"`

	// optional: icon/screenshot
	Type *string `json:"type,omitempty"`

	// abstraction of app
	Abstraction *string `json:"abstraction,omitempty"`

	// category id of the app
	CategoryID *string `json:"category_id,omitempty"`

	// description of the app
	Description *string `json:"description,omitempty"`

	// home page of the app
	Home *string `json:"home,omitempty"`

	// key words of the app
	Keywords *string `json:"keywords,omitempty"`

	// maintainers who maintainer the app
	Maintainers *string `json:"maintainers,omitempty"`

	// name of the app
	Name *string `json:"name,omitempty"`

	// instructions of the app
	Readme *string `json:"readme,omitempty"`

	// sources of app
	Sources *string `json:"sources,omitempty"`

	// tos of app
	Tos *string `json:"tos,omitempty"`
}

type ModifyAppVersionRequest struct {
	// app description
	Description *string `json:"description,omitempty"`

	// app name
	Name *string `json:"name,omitempty"`

	// package of app to replace other
	Package []byte `json:"package,omitempty"`

	// filename map to file_content
	PackageFiles map[string][]byte `json:"package_files,omitempty"`

	// required, version id of app to modify
	VersionID *string `json:"version_id,omitempty"`
}
type AppVersionAudit struct {

	// id of specific version app
	AppId string `json:"app_id,omitempty"`

	// name of specific version app
	AppName string `json:"app_name,omitempty"`

	// audit message
	Message string `json:"message,omitempty"`

	// user of auditer
	Operator string `json:"operator,omitempty"`

	// operator of auditer eg.[global_admin|developer|business|technical|isv]
	OperatorType string `json:"operator_type,omitempty"`

	// review id
	ReviewId string `json:"review_id,omitempty"`

	// audit status, eg.[draft|submitted|passed|rejected|active|in-review|deleted|suspended]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`

	// id of version to audit
	VersionId string `json:"version_id,omitempty"`

	// version name
	VersionName string `json:"version_name,omitempty"`

	// version type
	VersionType string `json:"version_type,omitempty"`
}

type ValidatePackageRequest struct {

	// required, version package eg.[the wordpress-0.0.1.tgz will be encoded to bytes]
	VersionPackage strfmt.Base64 `json:"version_package,omitempty"`

	// optional, vmbased/helm
	VersionType string `json:"version_type,omitempty"`
}
type App struct {

	// abstraction of app
	Abstraction string `json:"abstraction,omitempty"`

	// whether there is a released version in the app
	Active bool `json:"active,omitempty"`

	// app id
	AppId string `json:"app_id,omitempty"`

	// app version types eg.[vmbased|helm]
	AppVersionTypes string `json:"app_version_types,omitempty"`

	// category set
	CategorySet AppCategorySet `json:"category_set"`

	// chart name of app
	ChartName string `json:"chart_name,omitempty"`

	// company join time
	CompanyJoinTime *strfmt.DateTime `json:"company_join_time,omitempty"`

	// company name
	CompanyName string `json:"company_name,omitempty"`

	// company profile
	CompanyProfile string `json:"company_profile,omitempty"`

	// company website
	CompanyWebsite string `json:"company_website,omitempty"`

	// the time when app create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// app description
	Description string `json:"description,omitempty"`

	// app home page
	Home string `json:"home,omitempty"`

	// app icon
	Icon string `json:"icon,omitempty"`

	// the isv user who create the app
	Isv string `json:"isv,omitempty"`

	// app key words
	Keywords string `json:"keywords,omitempty"`

	// latest version of app
	LatestAppVersion *AppVersion `json:"latest_app_version,omitempty"`

	// app maintainers
	Maintainers string `json:"maintainers,omitempty"`

	// app name
	Name string `json:"name,omitempty"`

	// owner of app
	Owner string `json:"owner,omitempty"`

	// app instructions
	Readme string `json:"readme,omitempty"`

	// repository(store app package) id
	RepoId string `json:"repo_id,omitempty"`

	// app screenshots
	Screenshots string `json:"screenshots,omitempty"`

	// sources of app
	Sources string `json:"sources,omitempty"`

	// status eg.[modify|submit|review|cancel|release|delete|pass|reject|suspend|recover]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`

	// tos of app
	Tos string `json:"tos,omitempty"`

	// the time when app update
	UpdateTime *strfmt.DateTime `json:"update_time,omitempty"`

	ClusterTotal *int `json:"cluster_total,omitempty"`
}
type AppVersionReviewPhase struct {

	// review message
	Message string `json:"message,omitempty"`

	// user of reviewer
	Operator string `json:"operator,omitempty"`

	// operator type of reviewer eg.[global_admin|developer|business|technical|isv]
	OperatorType string `json:"operator_type,omitempty"`

	// app version review time
	ReviewTime *strfmt.DateTime `json:"review_time,omitempty"`

	// review status of app version eg.[isv-in-review|isv-passed|isv-rejected|isv-draft|business-in-review|business-passed|business-rejected|develop-draft|develop-in-review|develop-passed|develop-rejected|develop-draft]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`
}

type AppVersionReviewPhaseOAIGen map[string]AppVersionReviewPhase

type GetAppVersionPackageResponse struct {

	// app id of package
	AppId string `json:"app_id,omitempty"`

	// package of specific app version
	Package strfmt.Base64 `json:"package,omitempty"`

	// version id of package
	VersionId string `json:"version_id,omitempty"`
}

type GetAppVersionPackageFilesResponse struct {

	// filename map to content
	Files map[string]strfmt.Base64 `json:"files,omitempty"`

	// version id
	VersionId string `json:"version_id,omitempty"`
}
type AppVersionReview struct {

	// app id
	AppId string `json:"app_id,omitempty"`

	// app name
	AppName string `json:"app_name,omitempty"`

	// phase
	Phase AppVersionReviewPhaseOAIGen `json:"phase,omitempty"`

	// review id
	ReviewId string `json:"review_id,omitempty"`

	// user who review the app version
	Reviewer string `json:"reviewer,omitempty"`

	// review status eg.[isv-in-review|isv-passed|isv-rejected|isv-draft|business-in-review|business-passed|business-rejected|develop-draft|develop-in-review|develop-passed|develop-rejected|develop-draft]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"status_time,omitempty"`

	// id of app version
	VersionID string `json:"version_id,omitempty"`

	// version name of specific app version
	VersionName string `json:"version_name,omitempty"`

	// version type
	VersionType string `json:"version_type,omitempty"`

	// Workspace of the app version
	Workspace string `json:"workspace,omitempty"`
}

type CreateAppRequest struct {

	// app icon
	Icon strfmt.Base64 `json:"icon,omitempty"`

	// isv
	Isv string `json:"isv,omitempty"`

	// required, app name
	Name string `json:"name,omitempty"`

	// required, version name of the app
	VersionName string `json:"version_name,omitempty"`

	// required, version with specific app package
	VersionPackage strfmt.Base64 `json:"version_package,omitempty"`

	// optional, vmbased/helm
	VersionType string `json:"version_type,omitempty"`

	Username string `json:"-"`
}

type CreateAppResponse struct {

	// app id
	AppID string `json:"app_id,omitempty"`

	// version id of the app
	VersionID string `json:"version_id,omitempty"`
}

type AppVersion struct {

	// active or not
	Active bool `json:"active,omitempty"`

	// app id
	AppId string `json:"app_id,omitempty"`

	// the time when app version create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// description of app of specific version
	Description string `json:"description,omitempty"`

	// home of app of specific version
	Home string `json:"home,omitempty"`

	// icon of app of specific version
	Icon string `json:"icon,omitempty"`

	// keywords of app of specific version
	Keywords string `json:"keywords,omitempty"`

	// maintainers of app of specific version
	Maintainers string `json:"maintainers,omitempty"`

	// message path of app of specific version
	Message string `json:"message,omitempty"`

	// version name
	Name string `json:"name,omitempty"`

	// owner
	Owner string `json:"owner,omitempty"`

	// package name of app of specific version
	PackageName string `json:"package_name,omitempty"`

	// readme of app of specific version
	Readme string `json:"readme,omitempty"`

	// review id of app of specific version
	ReviewId string `json:"review_id,omitempty"`

	// screenshots of app of specific version
	Screenshots string `json:"screenshots,omitempty"`

	// sequence of app of specific version
	Sequence int64 `json:"sequence,omitempty"`

	// sources of app of specific version
	Sources string `json:"sources,omitempty"`

	// status of app of specific version eg.[draft|submitted|passed|rejected|active|in-review|deleted|suspended]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`

	// type of app of specific version
	Type string `json:"type,omitempty"`

	// the time when app version update
	UpdateTime *strfmt.DateTime `json:"update_time,omitempty"`

	// version id of app
	VersionId string `json:"version_id,omitempty"`

	ClusterTotal *int `json:"cluster_total,omitempty"`
}

type CreateAppVersionResponse struct {

	// version id
	VersionId string `json:"version_id,omitempty"`
}

type ValidatePackageResponse struct {

	// app description
	Description string `json:"description,omitempty"`

	// error eg.[json error]
	Error string `json:"error,omitempty"`

	// filename map to detail
	ErrorDetails map[string]string `json:"error_details,omitempty"`

	// app name eg.[wordpress|mysql|...]
	Name string `json:"name,omitempty"`

	// app url
	URL string `json:"url,omitempty"`

	// app version name.eg.[0.1.0]
	VersionName string `json:"version_name,omitempty"`
}

type CreateAppVersionRequest struct {

	// required, id of app to create new version
	AppId string `json:"app_id,omitempty"`

	// description of app of specific version
	Description string `json:"description,omitempty"`

	// required, version name eg.[0.1.0|0.1.3|...]
	Name string `json:"name,omitempty"`

	// package of app of specific version
	Package strfmt.Base64 `json:"package,omitempty"`

	// optional: vmbased/helm
	Type string `json:"type,omitempty"`

	Username string `json:"-"`
}

type GetAppVersionFilesRequest struct {
	Files []string `json:"files,omitempty"`
}

type ActionRequest struct {
	Action   string `json:"action"`
	Message  string `json:"message,omitempty"`
	Username string `json:"-"`
}

type Attachment struct {

	// filename map to content
	AttachmentContent map[string]strfmt.Base64 `json:"attachment_content,omitempty"`

	// attachment id
	AttachmentID string `json:"attachment_id,omitempty"`

	// the time attachment create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`
}

type CreateCategoryRequest struct {

	// category description
	Description string `json:"description,omitempty"`

	// category icon
	Icon strfmt.Base64 `json:"icon,omitempty"`

	// the i18n of this category, json format, sample: {"zh_cn": "数据库", "en": "database"}
	Locale string `json:"locale,omitempty"`

	// required, category name
	Name string `json:"name,omitempty"`
}
type ModifyCategoryRequest struct {
	// category description
	Description *string `json:"description,omitempty"`

	// category icon
	Icon []byte `json:"icon,omitempty"`

	// the i18n of this category, json format, sample: {"zh_cn": "数据库", "en": "database"}
	Locale *string `json:"locale,omitempty"`

	// category name
	Name *string `json:"name,omitempty"`
}

type AppCategorySet []*ResourceCategory

type ResourceCategory struct {

	// category id
	CategoryId string `json:"category_id,omitempty"`

	// create time
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// locale
	Locale string `json:"locale,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// status
	Status string `json:"status,omitempty"`

	// status time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`
}

type Category struct {

	// category id
	CategoryID string `json:"category_id,omitempty"`

	// the time when category create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// category description
	Description string `json:"description,omitempty"`

	// category icon
	Icon string `json:"icon,omitempty"`

	// the i18n of this category, json format, sample: {"zh_cn": "数据库", "en": "database"}
	Locale string `json:"locale,omitempty"`

	// category name,app belong to a category,eg.[AI|Firewall|cache|...]
	Name string `json:"name,omitempty"`

	// owner
	Owner string `json:"owner,omitempty"`

	AppTotal *int `json:"app_total,omitempty"`

	// the time when category update
	UpdateTime *strfmt.DateTime `json:"update_time,omitempty"`
}

type CreateCategoryResponse struct {

	// id of category created
	CategoryId string `json:"category_id,omitempty"`
}

type RepoEvent struct {
	// repository event create time
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// owner
	Owner string `json:"owner,omitempty"`

	// repository event id
	RepoEventId string `json:"repo_event_id,omitempty"`

	// repository id
	RepoId string `json:"repo_id,omitempty"`

	// result
	Result string `json:"result,omitempty"`

	// repository event status eg.[failed|successful|working|pending]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`
}

type CreateRepoRequest struct {
	// required app default status.eg:[draft|active]
	AppDefaultStatus string `json:"app_default_status,omitempty"`

	// category id
	CategoryId string `json:"category_id,omitempty"`

	// required, credential of visiting the repository
	Credential string `json:"credential,omitempty"`

	// repository description
	Description string `json:"description,omitempty"`

	// If workspace is empty, then it's a global repo
	Workspace *string `json:"workspace,omitempty"`

	// required, repository name
	Name string `json:"name,omitempty"`

	// required, runtime provider eg.[qingcloud|aliyun|aws|kubernetes]
	Providers []string `json:"providers"`

	// min sync period to sync helm repo, a duration string is a sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "180s", "2h" or "45m".
	SyncPeriod string `json:"sync_period"`

	// repository type
	Type string `json:"type,omitempty"`

	// required, url of visiting the repository
	URL string `json:"url,omitempty"`

	// required, visibility eg:[public|private]
	Visibility string `json:"visibility,omitempty"`
}

type RepoCategorySet []*ResourceCategory

type ModifyRepoRequest struct {
	// app default status eg:[draft|active]
	AppDefaultStatus *string `json:"app_default_status,omitempty"`

	// category id
	CategoryID *string `json:"category_id,omitempty"`

	// credential of visiting the repository
	Credential *string `json:"credential,omitempty"`

	// repository description
	Description *string `json:"description,omitempty"`

	Workspace *string `json:"workspace,omitempty"`

	// min sync period to sync helm repo
	SyncPeriod *string `json:"sync_period"`

	// repository name
	Name *string `json:"name,omitempty"`

	// runtime provider eg.[qingcloud|aliyun|aws|kubernetes]
	Providers []string `json:"providers"`

	// repository type
	Type *string `json:"type,omitempty"`

	// url of visiting the repository
	URL *string `json:"url,omitempty"`

	// visibility eg:[public|private]
	Visibility *string `json:"visibility,omitempty"`
}

type RepoActionRequest struct {
	Action    string `json:"action"`
	Workspace string `json:"workspace"`
}

type ValidateRepoRequest struct {
	Type       string `json:"type"`
	Credential string `json:"credential"`
	Url        string `json:"url"`
}

type RepoSelector struct {
	// the time when repository selector create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// selector key
	SelectorKey string `json:"selector_key,omitempty"`

	// selector value
	SelectorValue string `json:"selector_value,omitempty"`
}
type RepoLabel struct {
	// the time when repository label create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// label key
	LabelKey string `json:"label_key,omitempty"`

	// label value
	LabelValue string `json:"label_value,omitempty"`
}

type RepoLabels []*RepoLabel
type RepoSelectors []*RepoSelector

type Repo struct {
	ChartCount int `json:"chart_count,omitempty"`

	// app default status eg[active|draft]
	AppDefaultStatus string `json:"app_default_status,omitempty"`

	// category set
	CategorySet RepoCategorySet `json:"category_set"`

	// controller, value 0 for self resource, value 1 for openpitrix resource
	Controller int32 `json:"controller,omitempty"`

	// the time when repository create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// credential of visiting the repository
	Credential string `json:"credential,omitempty"`

	// repository description
	Description string `json:"description,omitempty"`

	// labels
	Labels RepoLabels `json:"labels"`

	// repository name
	Name string `json:"name,omitempty"`

	// creator
	Creator string `json:"creator,omitempty"`

	// runtime provider eg.[qingcloud|aliyun|aws|kubernetes]
	Providers []string `json:"providers"`

	// repository id
	RepoId string `json:"repo_id,omitempty"`

	// selectors
	Selectors RepoSelectors `json:"selectors"`

	// status eg.[active|deleted]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"status_time,omitempty"`

	// type of repository eg.[http|https|s3]
	Type string `json:"type,omitempty"`

	// url of visiting the repository
	URL string `json:"url,omitempty"`

	// visibility.eg:[public|private]
	Visibility string `json:"visibility,omitempty"`

	SyncPeriod string `json:"sync_period,omitempty"`
}

type CreateRepoResponse struct {

	// id of repository created
	RepoID string `json:"repo_id,omitempty"`
}

type ValidateRepoResponse struct {

	// if validate error,return error code
	ErrorCode int64 `json:"errorCode,omitempty"`

	// validate repository ok or not
	Ok bool `json:"ok,omitempty"`
}

type CreateClusterRequest struct {

	// release name
	Name string `json:"name"`

	// release install description
	Description string `json:"description"`

	// advanced param
	AdvancedParam []string `json:"advanced_param"`

	// required, id of app to run in cluster
	AppId string `json:"app_id,omitempty"`

	// required, conf a json string, include cpu, memory info of cluster
	Conf string `json:"conf,omitempty"`

	// required, id of runtime
	RuntimeId string `json:"runtime_id,omitempty"`

	// required, id of app version
	VersionId string `json:"version_id,omitempty"`

	Username string `json:"-"`

	// current workspace
	Workspace string `json:"workspace,omitempty"`
}

type UpgradeClusterRequest struct {
	// release namespace
	Namespace string `json:"namespace,omitempty"`

	// cluster id
	ClusterId string `json:"cluster_id"`

	// helm app id
	AppId string `json:"app_id"`

	// advanced param
	AdvancedParam []string `json:"advanced_param"`

	// required, conf a json string, include cpu, memory info of cluster
	Conf string `json:"conf,omitempty"`

	// Deprecated: required, id of runtime
	RuntimeId string `json:"runtime_id,omitempty"`

	// required, id of app version
	VersionId string `json:"version_id,omitempty"`

	Username string `json:"-"`
}

type Cluster struct {

	// additional info
	AdditionalInfo string `json:"additional_info,omitempty"`

	// id of app run in cluster
	AppId string `json:"app_id,omitempty"`

	// cluster id
	ClusterId string `json:"cluster_id,omitempty"`

	// cluster type, frontgate or normal cluster
	ClusterType int64 `json:"cluster_type,omitempty"`

	// the time when cluster create
	CreateTime *strfmt.DateTime `json:"create_time,omitempty"`

	// cluster used to debug or not
	Debug bool `json:"debug,omitempty"`

	// cluster description
	Description string `json:"description,omitempty"`

	// endpoint of cluster
	Endpoints string `json:"endpoints,omitempty"`

	// cluster env
	Env string `json:"env,omitempty"`

	// frontgate id, a proxy for vpc to communicate
	FrontgateId string `json:"frontgate_id,omitempty"`

	// global uuid
	GlobalUUID string `json:"global_uuid,omitempty"`

	// metadata root access
	MetadataRootAccess bool `json:"metadata_root_access,omitempty"`

	// cluster name
	Name string `json:"name,omitempty"`

	// owner
	Owner string `json:"owner,omitempty"`

	// cluster runtime id
	RuntimeId string `json:"runtime_id,omitempty"`

	// cluster status eg.[active|used|enabled|disabled|deleted|stopped|ceased]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime *strfmt.DateTime `json:"status_time,omitempty"`

	// subnet id, cluster run in a subnet
	SubnetId string `json:"subnet_id,omitempty"`

	// cluster transition status eg.[creating|deleting|upgrading|updating|rollbacking|stopping|starting|recovering|ceasing|resizing|scaling]
	TransitionStatus string `json:"transition_status,omitempty"`

	// upgrade status, unused
	UpgradeStatus string `json:"upgrade_status,omitempty"`

	// cluster upgraded time
	UpgradeTime *strfmt.DateTime `json:"upgrade_time,omitempty"`

	// id of version of app run in cluster
	VersionId string `json:"version_id,omitempty"`

	// vpc id, a vpc contain one more subnet
	VpcId string `json:"vpc_id,omitempty"`

	// zone of cluster eg.[pek3a|pek3b]
	Zone string `json:"zone,omitempty"`
}

type Runtime struct {
	// runtime id
	RuntimeId string `protobuf:"bytes,1,opt,name=runtime_id,json=runtimeId,proto3" json:"runtime_id,omitempty"`
	// runtime name,create by owner.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// runtime description
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

type ModifyClusterAttributesRequest struct {
	ClusterName string `json:"clusterName,omitempty"`
	Namespace   string `json:"namespace,omitempty"`

	// required, id of cluster to modify
	ClusterID string `json:"cluster_id"`

	// cluster description
	Description *string `json:"description,omitempty"`

	// cluster name
	Name *string `json:"name,omitempty"`
}

const (
	CreateTime = "create_time"
	StatusTime = "status_time"

	VersionId       = "version_id"
	RepoId          = "repo_id"
	CategoryId      = "category_id"
	Status          = "status"
	Type            = "type"
	Visibility      = "visibility"
	AppId           = "app_id"
	Keyword         = "keyword"
	ISV             = "isv"
	WorkspaceLabel  = "workspace"
	BuiltinRepoId   = "repo-helm"
	StatusActive    = "active"
	StatusSuspended = "suspended"
	ActionRecover   = "recover"
	ActionSuspend   = "suspend"
	ActionCancel    = "cancel"
	ActionPass      = "pass"
	ActionReject    = "reject"
	ActionSubmit    = "submit"
	ActionRelease   = "release"
	Ascending       = "ascending"
	ActionIndex     = "index"
)
