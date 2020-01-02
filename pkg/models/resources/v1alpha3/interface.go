package v1alpha3

import (
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

type Interface interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (interface{}, error)

	// List retrieves a collection of objects matches given query
	List(namespace string) ([]interface{}, error)

	//
	Filter(item interface{}, filter query.Filter) bool

	//
	Compare(left interface{}, right interface{}, field query.Field) bool
}
