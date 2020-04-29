package query

type Field string
type Value string

const (
	FieldName                = "name"
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
