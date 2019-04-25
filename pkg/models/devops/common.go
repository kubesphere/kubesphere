package devops

import (
	"github.com/fatih/structs"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

func GetColumnsFromStruct(s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = stringutils.CamelCaseToUnderscore(name)
	}
	return names
}

func GetColumnsFromStructWithPrefix(prefix string, s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = WithPrefix(prefix, stringutils.CamelCaseToUnderscore(name))
	}
	return names
}

func WithPrefix(prefix, str string) string {
	return prefix + "." + str
}

const (
	StatusActive     = "active"
	StatusDeleted    = "deleted"
	StatusDeleting   = "deleting"
	StatusFailed     = "failed"
	StatusPending    = "pending"
	StatusWorking    = "working"
	StatusSuccessful = "successful"
)

const (
	StatusColumn     = "status"
	StatusTimeColumn = "status_time"
)

const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
)

const (
	KS_ADMIN = "admin"
)

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
)
