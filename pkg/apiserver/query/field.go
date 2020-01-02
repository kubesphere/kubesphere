package query

type Field string

const (
	FieldName                = "name"
	FieldCreationTimeStamp   = "creationTimestamp"
	FieldLastUpdateTimestamp = "lastUpdateTimestamp"
	FieldNamespace           = "namespace"
	FieldStatus              = "status"
	FieldApplication         = "application"
	FieldOwner               = "owner"
	FieldOwnerKind           = "ownerKind"
)

var SortableFields = []Field{
	FieldCreationTimeStamp,
	FieldLastUpdateTimestamp,
	FieldName,
}

// Field contains all the query field that can be compared
var ComparableFields = []Field{
	FieldName,
	FieldNamespace,
	FieldStatus,
	FieldApplication,
	FieldOwner,
	FieldOwnerKind,
}
