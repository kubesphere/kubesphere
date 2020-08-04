/*
Copyright 2020 KubeSphere Authors

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
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
	Owner            = "owner"

	StatusRunning            = "running"
	StatusPaused             = "paused"
	StatusPending            = "pending"
	StatusUpdating           = "updating"
	StatusStopped            = "stopped"
	StatusFailed             = "failed"
	StatusBound              = "bound"
	StatusLost               = "lost"
	StatusComplete           = "completed"
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

func ObjectMetaExactlyMath(key, value string, item metav1.ObjectMeta) bool {
	switch key {
	case Name:
		names := strings.Split(value, ",")
		if !sliceutil.HasString(names, item.Name) {
			return false
		}
	case Keyword:
		if !strings.Contains(item.Name, value) && !FuzzyMatch(item.Labels, "", value) && !FuzzyMatch(item.Annotations, "", value) {
			return false
		}
	case Owner:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(string(ownerReference.UID), value) == 0 {
				return true
			}
		}
		return false
	default:
		// label not exist or value not equal
		if val, ok := item.Labels[key]; !ok || val != value {
			return false
		}
	}
	return true
}

func ObjectMetaFuzzyMath(key, value string, item metav1.ObjectMeta) bool {
	switch key {
	case Name:
		if !strings.Contains(item.Name, value) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], value) {
			return false
		}
	case Label:
		if !FuzzyMatch(item.Labels, "", value) {
			return false
		}
	case Annotation:
		if !FuzzyMatch(item.Annotations, "", value) {
			return false
		}
		return false
	case App:
		if !strings.Contains(item.Labels[Chart], value) && !strings.Contains(item.Labels[Release], value) {
			return false
		}
	default:
		if !FuzzyMatch(item.Labels, key, value) {
			return false
		}
	}
	return true
}

func FuzzyMatch(m map[string]string, key, value string) bool {

	val, exist := m[key]

	if value == "" && (!exist || val == "") {
		return true
	} else if value != "" && strings.Contains(val, value) {
		return true
	}

	return false
}

func ObjectMetaCompare(left, right metav1.ObjectMeta, compareField string) bool {
	switch compareField {
	case CreateTime:
		if left.CreationTimestamp.Time.Equal(right.CreationTimestamp.Time) {
			if left.Namespace == right.Namespace {
				return strings.Compare(left.Name, right.Name) < 0
			}
			return strings.Compare(left.Namespace, right.Namespace) < 0
		}
		return left.CreationTimestamp.Time.Before(right.CreationTimestamp.Time)
	case Name:
		fallthrough
	default:
		return strings.Compare(left.Name, right.Name) <= 0
	}
}
