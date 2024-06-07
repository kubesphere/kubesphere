/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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
)
