package v1alpha2

import (
	"kubesphere.io/kubesphere/pkg/server/params"
	"strings"
)

const (
	App              = "app"
	Chart            = "chart"
	Release          = "release"
	Name             = "name"
	Label            = "label"
	OwnerKind        = "ownerKind"
	OwnerName        = "ownerName"
	TargetKind       = "targetKind"
	TargetName       = "targetName"
	Role             = "role"
	CreateTime       = "createTime"
	UpdateTime       = "updateTime"
	StartTime        = "startTime"
	LastScheduleTime = "lastScheduleTime"
	Annotation       = "Annotation"
	Keyword          = "keyword"
	UserFacing       = "userfacing"
	Status           = "status"

	StatusRunning            = "running"
	StatusPaused             = "paused"
	StatusPending            = "pending"
	StatusUpdating           = "updating"
	StatusStopped            = "stopped"
	StatusFailed             = "failed"
	StatusBound              = "bound"
	StatusLost               = "lost"
	StatusComplete           = "complete"
	StatusWarning            = "warning"
	StatusUnschedulable      = "unschedulable"
	Deployments              = "deployments"
	DaemonSets               = "daemonsets"
	Roles                    = "roles"
	Workspaces               = "workspaces"
	WorkspaceRoles           = "workspaceroles"
	CronJobs                 = "cronjobs"
	ConfigMaps               = "configmaps"
	Ingresses                = "ingresses"
	Jobs                     = "jobs"
	PersistentVolumeClaims   = "persistentvolumeclaims"
	Pods                     = "pods"
	Secrets                  = "secrets"
	Services                 = "services"
	StatefulSets             = "statefulsets"
	HorizontalPodAutoscalers = "horizontalpodautoscalers"
	Applications             = "applications"
	Nodes                    = "nodes"
	Namespaces               = "namespaces"
	StorageClasses           = "storageclasses"
	ClusterRoles             = "clusterroles"
	S2iBuilderTemplates      = "s2ibuildertemplates"
	S2iBuilders              = "s2ibuilders"
	S2iRuns                  = "s2iruns"
)

type Interface interface {
	Get(namespace, name string) (interface{}, error)
	Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error)
}

func SearchFuzzy(m map[string]string, key, value string) bool {

	val, exist := m[key]

	if value == "" && (!exist || val == "") {
		return true
	} else if value != "" && strings.Contains(val, value) {
		return true
	}

	return false
}
