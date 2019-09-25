// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package gerr

import "fmt"

type ErrorMessage struct {
	Name string
	en   string
	zhCN string
}

func (em ErrorMessage) Message(locale string, err error, a ...interface{}) string {
	format := ""
	switch locale {
	case En:
		format = em.en
	case ZhCN:
		if len(em.zhCN) > 0 {
			format = em.zhCN
		} else {
			format = em.en
		}
	}
	if err != nil {
		return fmt.Sprintf("%s: %s", fmt.Sprintf(format, a...), err.Error())
	} else {
		return fmt.Sprintf(format, a...)
	}
}

var (
	ErrorPermissionDenied = ErrorMessage{
		Name: "permission_denied",
		en:   "permission denied",
		zhCN: "没有权限",
	}
	ErrorAuthFailure = ErrorMessage{
		Name: "auth_failure",
		en:   "auth failure",
		zhCN: "认证失败",
	}
	ErrorAccessTokenExpired = ErrorMessage{
		Name: "access_token_expired",
		en:   "access token expired",
		zhCN: "访问令牌已过期",
	}
	ErrorRefreshTokenExpired = ErrorMessage{
		Name: "refresh_token_expired",
		en:   "refresh token expired",
		zhCN: "刷新令牌已过期",
	}
	ErrorEmailPasswordNotMatched = ErrorMessage{
		Name: "email_password_not_matched",
		en:   "email and password does not match",
		zhCN: "邮箱和密码不匹配",
	}
	ErrorPasswordIncorrect = ErrorMessage{
		Name: "password_incorrect",
		en:   "password incorrect",
		zhCN: "密码不正确",
	}
	ErrorRuntimeCredentialExists = ErrorMessage{
		Name: "runtime_credential_exists",
		en:   "runtime credential exists",
		zhCN: "环境授权信息已存在",
	}
	ErrorUnsupportedRuntimeProvider = ErrorMessage{
		Name: "unsupported_runtime_provider",
		en:   "unsupported runtime provider [%s]",
		zhCN: "不支持云环境服务商[%s]",
	}
	ErrorRuntimeExists = ErrorMessage{
		Name: "runtime_exists",
		en:   "runtime exists",
		zhCN: "环境已存在",
	}
	ErrorEmailExists = ErrorMessage{
		Name: "email_exists",
		en:   "email [%s] exists",
		zhCN: "邮箱[%s]已存在",
	}
	ErrorEmailNotExists = ErrorMessage{
		Name: "email_not_exists",
		en:   "email [%s] not exists",
		zhCN: "邮箱[%s]不存在",
	}
	ErrorCreateResourcesFailed = ErrorMessage{
		Name: "create_resources_failed",
		en:   "create resources failed",
		zhCN: "创建资源失败",
	}
	ErrorCreateResourceFailed = ErrorMessage{
		Name: "create_resource_failed",
		en:   "create resource [%s] failed",
		zhCN: "创建资源[%s]失败",
	}
	ErrorDeleteResourcesFailed = ErrorMessage{
		Name: "delete_resources_failed",
		en:   "delete resources failed",
		zhCN: "删除资源失败",
	}
	ErrorDeleteResourceFailed = ErrorMessage{
		Name: "delete_resource_failed",
		en:   "delete resource [%s] failed",
		zhCN: "删除资源[%s]失败",
	}
	ErrorDeleteFrontgateWithClustersFailed = ErrorMessage{
		Name: "delete_frontgate_with_clusters_failed",
		en:   "delete frontgate [%s] with clusters [%s] failed",
		zhCN: "删除代理[%s]失败，仍有[%s]依赖",
	}
	ErrorUpgradeResourceFailed = ErrorMessage{
		Name: "upgrade_resource_failed",
		en:   "upgrade resource [%s] failed",
		zhCN: "升级资源[%s]失败",
	}
	ErrorRollbackResourceFailed = ErrorMessage{
		Name: "rollback_resource_failed",
		en:   "rollback resource [%s] failed",
		zhCN: "回滚资源[%s]失败",
	}
	ErrorResizeResourceFailed = ErrorMessage{
		Name: "resize_resource_failed",
		en:   "resize resource [%s] failed",
		zhCN: "调整资源[%s]失败",
	}
	ErrorAddResourceNodeFailed = ErrorMessage{
		Name: "add_resource_node_failed",
		en:   "add resource [%s] node failed",
		zhCN: "为资源[%s]增加节点失败",
	}
	ErrorDeleteResourceNodeFailed = ErrorMessage{
		Name: "delete_resource_node_failed",
		en:   "delete resource [%s] node failed",
		zhCN: "删除资源[%s]的节点失败",
	}
	ErrorUpdateResourceEnvFailed = ErrorMessage{
		Name: "update_resource_env_failed",
		en:   "update resource [%s] env failed",
		zhCN: "更新资源[%s]环境变量失败",
	}
	ErrorUpdateResourceFailed = ErrorMessage{
		Name: "update_resource_failed",
		en:   "update resource [%s] failed",
		zhCN: "更新资源[%s]失败",
	}
	ErrorStopResourceFailed = ErrorMessage{
		Name: "stop_resource_failed",
		en:   "stop resource [%s] failed",
		zhCN: "暂停资源[%s]失败",
	}
	ErrorStartResourceFailed = ErrorMessage{
		Name: "start_resource_failed",
		en:   "start resource [%s] failed",
		zhCN: "启动资源[%s]失败",
	}
	ErrorRecoverResourceFailed = ErrorMessage{
		Name: "recover_resource_failed",
		en:   "recover resource [%s] failed",
		zhCN: "回复资源[%s]失败",
	}
	ErrorCeaseResourceFailed = ErrorMessage{
		Name: "cease_resource_failed",
		en:   "cease resource [%s] failed",
		zhCN: "释放资源[%s]失败",
	}
	ErrorRetryTaskFailed = ErrorMessage{
		Name: "retry_task_failed",
		en:   "retry task [%s] failed",
		zhCN: "重试任务[%s]失败",
	}
	ErrorDescribeResourcesFailed = ErrorMessage{
		Name: "describe_resources_failed",
		en:   "describe resources failed",
		zhCN: "获取资源失败",
	}
	ErrorDescribeResourceFailed = ErrorMessage{
		Name: "describe_resource_failed",
		en:   "describe resource [%s] failed",
		zhCN: "获取资源[%s]失败",
	}
	ErrorModifyResourcesFailed = ErrorMessage{
		Name: "modify_resources_failed",
		en:   "modify resources failed",
		zhCN: "修改资源失败",
	}
	ErrorModifyResourceFailed = ErrorMessage{
		Name: "modify_resource_failed",
		en:   "modify resource [%s] failed",
		zhCN: "修改资源[%s]失败",
	}
	ErrorResourceNotFound = ErrorMessage{
		Name: "resource_not_found",
		en:   "resource [%s] not found",
		zhCN: "没有找到资源[%s]",
	}
	ErrorResourceRoleNotFound = ErrorMessage{
		Name: "resource_role_not_found",
		en:   "resource [%s] role [%s] not found",
		zhCN: "没有找到资源[%s]对应的角色[%s]",
	}
	ErrorSubnetNotFound = ErrorMessage{
		Name: "subnet_not_found",
		en:   "subnet [%s] not found or vpc not bind eip",
		zhCN: "没有找到子网[%s]或者VPC没有绑定公网IP",
	}
	ErrorThereAreNoAvailableSubnet = ErrorMessage{
		Name: "there_are_no_available_subnet",
		en:   "there are no available subnet",
		zhCN: "没有可用的子网",
	}
	ErrorProviderNotFound = ErrorMessage{
		Name: "provider_not_found",
		en:   "provider [%s] not found",
		zhCN: "云服务商[%s]不存在",
	}
	ErrorInternalError = ErrorMessage{
		Name: "internal_error",
		en:   "internal error",
		zhCN: "内部错误",
	}
	ErrorMissingParameter = ErrorMessage{
		Name: "missing_parameter",
		en:   "missing parameter [%s]",
		zhCN: "缺少参数[%s]",
	}
	ErrorValidateFailed = ErrorMessage{
		Name: "validate_failed",
		en:   "validate failed",
		zhCN: "校验失败",
	}
	ErrorParameterParseFailed = ErrorMessage{
		Name: "parameter_parse_failed",
		en:   "parameter [%s] parse failed",
		zhCN: "参数[%s]解析失败",
	}
	ErrorResourceAlreadyDeleted = ErrorMessage{
		Name: "resource_already_deleted",
		en:   "resource [%s] has already been deleted",
		zhCN: "资源[%s]已被删除",
	}
	ErrorResourceNotInStatus = ErrorMessage{
		Name: "resource_not_in_status",
		en:   "resource [%s] is not in status [%s]",
		zhCN: "资源[%s]不处于[%s]状态",
	}
	ErrorResourceTransitionStatus = ErrorMessage{
		Name: "resource_transition_status",
		en:   "resource [%s] is [%s]",
		zhCN: "资源[%s]处于[%s]状态",
	}
	ErrorIllegalParameterLength = ErrorMessage{
		Name: "illegal_parameter_length",
		en:   "illegal parameter [%s] length",
		zhCN: "参数[%s]的长度非法",
	}
	ErrorParameterShouldNotBeEmpty = ErrorMessage{
		Name: "parameter_should_not_be_empty",
		en:   "parameter [%s] should not be empty",
		zhCN: "参数[%s]不应该为空",
	}
	ErrorUnsupportedParameterValue = ErrorMessage{
		Name: "unsupported_parameter_value",
		en:   "unsupported parameter [%s] value [%s]",
		zhCN: "参数[%s]不支持值[%s]",
	}
	ErrorIllegalUrlFormat = ErrorMessage{
		Name: "illegal_url_format",
		en:   "illegal URL format [%s]",
		zhCN: "非法的URL格式[%s]",
	}
	ErrorConflictRepoName = ErrorMessage{
		Name: "conflict_repo_name",
		en:   "conflict repo name [%s]",
		zhCN: "仓库名称[%s]已存在",
	}
	ErrorResourceQuotaNotEnough = ErrorMessage{
		Name: "resource_quota_not_enough",
		en:   "resource quota not enough: %s",
		zhCN: "资源配额不足: %s",
	}
	ErrorHelmReleaseExists = ErrorMessage{
		Name: "helm_release_exists",
		en:   "helm release [%s] already exists",
		zhCN: "",
	}
	ErrorUnsupportedApiVersion = ErrorMessage{
		Name: "unsupported_api_version",
		en:   "unsupported api version [%s]",
		zhCN: "不支持的API版本 [%s]",
	}
	ErrorCannotDeleteDefaultCategory = ErrorMessage{
		Name: "cannot_delete_default_category",
		en:   "cannot delete default category",
		zhCN: "无法删除默认的分类",
	}
	ErrorAttachKeyPairsFailed = ErrorMessage{
		Name: "attach_key_pairs_failed",
		en:   "attach key pairs failed",
		zhCN: "绑定key pair失败",
	}
	ErrorDetachKeyPairsFailed = ErrorMessage{
		Name: "detach_key_pairs_failed",
		en:   "detach key pairs failed",
		zhCN: "解除key pair失败",
	}
	ErrorAppVersionIncorrectStatus = ErrorMessage{
		Name: "app_version_incorrect_status",
		en:   "app version [%s] has incorrect status [%s], cannot execute the current action",
		zhCN: "应用版本[%s]状态为[%s], 无法执行此操作",
	}
	ErrorAppVersionInReview = ErrorMessage{
		Name: "app_version_in_review",
		en:   "app version is under review, app cannot be modified",
		zhCN: "应用版本审核中, 应用无法修改",
	}
	ErrorLoadPackageFailed = ErrorMessage{
		Name: "load_package_failed",
		en:   "load package failed, reason: [%s]",
		zhCN: "载入配置包失败, 原因: [%s]",
	}
	ErrorCannotChangeAppName = ErrorMessage{
		Name: "cannot_change_app_name",
		en:   "cannot change app name",
		zhCN: "无法修改应用名称",
	}
	ErrorAppNameExists = ErrorMessage{
		Name: "app_name_exists",
		en:   "app name [%s] exists",
		zhCN: "应用名称[%s]已存在",
	}
	ErrorAppVersionExists = ErrorMessage{
		Name: "app_version_exists",
		en:   "app version [%s:%s] exists",
		zhCN: "应用版本[%s:%s]已存在",
	}
	ErrorCompanyNameExists = ErrorMessage{
		Name: "company_name_exists",
		en:   "company name [%s] exists",
		zhCN: "公司名称[%s]已存在",
	}
	ErrorCannotAccessRepo = ErrorMessage{
		Name: "cannot_access_repo",
		en:   "cannot access repo",
		zhCN: "仓库无法访问",
	}
	ErrorCannotWriteRepo = ErrorMessage{
		Name: "cannot_write_repo",
		en:   "cannot write repo [%s]",
		zhCN: "仓库[%s]无法写入",
	}
	ErrorCannotDeleteInternalRepo = ErrorMessage{
		Name: "cannot_delete_internal_repo",
		en:   "cannot delete internal repo [%s]",
		zhCN: "无法删除内置仓库[%s]",
	}
	ErrorResourceAccessDenied = ErrorMessage{
		Name: "error_resource_access_denied",
		en:   "access denied for resource [%s]",
		zhCN: "拒绝访问资源[%s]",
	}
	ErrorExistsNoDeleteVersions = ErrorMessage{
		Name: "exists_no_delete_versions",
		en:   "app [%s] had some versions not deleted",
		zhCN: "应用[%s]还有未删除的版本",
	}
	ErrorTillerNotServe = ErrorMessage{
		Name: "tiller_not_serve",
		en:   "tiller not serve in namespace [%s]",
		zhCN: "tiller 在命名空间[%s]下未正常服务",
	}
	ErrorNamespaceUnavailable = ErrorMessage{
		Name: "namespace_unavailable",
		en:   "namespace [%s] unavailable",
		zhCN: "命名空间[%s]不可用",
	}
	ErrorNamespaceNotMatchWithRegex = ErrorMessage{
		Name: "namespace_not_match_with_regex",
		en:   "namespace [%s] not match with regex [%s]",
		zhCN: "命名空间[%s]命名不合法, 需要满足[%s]",
	}
	ErrorCredentialIllegal = ErrorMessage{
		Name: "credential_illegal",
		en:   "credential [%s] illegal",
		zhCN: "credential [%s]不合法",
	}
	ErrorNamespaceExists = ErrorMessage{
		Name: "namespace exists",
		en:   "namespace [%s] exists",
		zhCN: "命名空间[%s]已存在",
	}
	ErrorPackageParseFailed = ErrorMessage{
		Name: "package_parse_failed",
		en:   "package parse failed",
		zhCN: "配置包解析失败",
	}
	ErrorAppNameConflictWithPackage = ErrorMessage{
		Name: "app_name_conflict_with_package",
		en:   "app name conflict with package",
		zhCN: "应用名称与配置包内信息冲突",
	}
	ErrorImageDecodeFailed = ErrorMessage{
		Name: "image_decode_failed",
		en:   "image decode failed",
		zhCN: "图片解码失败",
	}
	ErrorIllegalEmailFormat = ErrorMessage{
		Name: "illegal_email_format",
		en:   "illegal Email format [%s]",
		zhCN: "非法的Email格式[%s]",
	}
	ErrorIllegalPhoneNumFormat = ErrorMessage{
		Name: "illegal_phone_num_format",
		en:   "illegal phone number format [%s]",
		zhCN: "非法的电话号码格式[%s]",
	}
	ErrorIllegalBankAccountNumberFormat = ErrorMessage{
		Name: "illegal_bankAccountNumber_format",
		en:   "illegal BankAccountNumber format [%s]",
		zhCN: "非法的银行账号格式[%s]",
	}
	ErrorGroupHadMembers = ErrorMessage{
		Name: "group_had_members",
		en:   "group had members",
		zhCN: "组内还有成员",
	}
	ErrorSetNotificationConfig = ErrorMessage{
		Name: "error_set_notification_config",
		en:   "set notification config failed",
		zhCN: "设置通知服务配置失败",
	}
	ErrorSetServiceConfig = ErrorMessage{
		Name: "error_set_service_config",
		en:   "set service config failed",
		zhCN: "设置服务配置失败",
	}
	ErrorGetNotificationConfig = ErrorMessage{
		Name: "error_get_notification_config",
		en:   "get notification config failed",
		zhCN: "查看通知服务配置失败",
	}
	ErrorCannotDeleteUsers = ErrorMessage{
		Name: "error_cannot_delete_users",
		en:   "cannot delete users",
		zhCN: "无法删除用户",
	}
	ErrorCannotDeleteGroups = ErrorMessage{
		Name: "error_cannot_delete_groups",
		en:   "cannot delete groups",
		zhCN: "无法删除用户组",
	}
	ErrorGroupNotFound = ErrorMessage{
		Name: "error_group_not_found",
		en:   "group [%s] not found",
		zhCN: "没有找到用户组[%s]",
	}
	ErrorGroupAccessDenied = ErrorMessage{
		Name: "error_group_access_denied",
		en:   "access denied for group [%s]",
		zhCN: "拒绝访问用户组[%s]",
	}
	ErrorUserNotFound = ErrorMessage{
		Name: "error_user_not_found",
		en:   "user [%s] not found",
		zhCN: "没有找到用户[%s]",
	}
	ErrorUserAccessDenied = ErrorMessage{
		Name: "error_user_access_denied",
		en:   "access denied for user [%s]",
		zhCN: "拒绝访问用户[%s]",
	}
	ErrorCannotJoinGroup = ErrorMessage{
		Name: "error_cannot_join_group",
		en:   "cannot join group",
		zhCN: "无法加入用户组",
	}
	ErrorCannotLeaveGroup = ErrorMessage{
		Name: "error_cannot_leave_group",
		en:   "cannot leave group",
		zhCN: "无法离开用户组",
	}
	ErrorCannotCreateUserWithRole = ErrorMessage{
		Name: "error_cannot_create_user_with_role",
		en:   "cannot create user with role [%s]",
		zhCN: "无法创建[%s]角色的用户",
	}
	ErrorValidateEmailService = ErrorMessage{
		Name: "error_validate_email_service",
		en:   "validate email service failed",
		zhCN: "验证邮件服务配置失败",
	}
)
