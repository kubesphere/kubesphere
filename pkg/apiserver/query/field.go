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

package query

type Field string
type Value string

const (
	FieldName                = "name"
	FieldNames               = "names"
	FieldUID                 = "uid"
	FieldCreationTimeStamp   = "creationTimestamp"
	FieldCreateTime          = "createTime"
	FieldLastUpdateTimestamp = "lastUpdateTimestamp"
	FieldUpdateTime          = "updateTime"
	FieldLabel               = "label"
	FieldAnnotation          = "annotation"
	FieldNamespace           = "namespace"
	FieldStatus              = "status"
	FieldOwnerReference      = "ownerReference"
	FieldOwnerKind           = "ownerKind"

	FieldType = "type"
)

var SortableFields = []Field{
	FieldCreationTimeStamp,
	FieldCreateTime,
	FieldUpdateTime,
	FieldLastUpdateTimestamp,
	FieldName,
}

// Field contains all the query field that can be compared
var ComparableFields = []Field{
	FieldName,
	FieldUID,
	FieldLabel,
	FieldAnnotation,
	FieldNamespace,
	FieldStatus,
	FieldOwnerReference,
	FieldOwnerKind,
}
