/*
Copyright 2019 The KubeSphere Authors.
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
