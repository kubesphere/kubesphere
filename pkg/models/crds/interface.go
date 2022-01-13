/*
Copyright 2022 KubeSphere Authors

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

package crds

import (
	"sort"
	"strings"

	"github.com/emicklei/go-restful"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

var (
	Filters      = map[schema.GroupVersionKind]FilterFunc{}
	Comparers    = map[schema.GroupVersionKind]CompareFunc{}
	Transformers = map[schema.GroupVersionKind][]TransformFunc{}
)

type Reader interface {
	// Get retrieves a single object by its namespace and name
	Get(key types.NamespacedName) (client.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.Query) (*api.ListResult, error)
}

type Handler interface {
	GetResources(request *restful.Request, response *restful.Response)
	ListResources(request *restful.Request, response *restful.Response)
}

type Client interface {
	Handler
	Reader
}

// CompareFunc return true is left great than right
type CompareFunc func(metav1.Object, metav1.Object, query.Field) bool

type FilterFunc func(metav1.Object, query.Filter) bool

type TransformFunc func(metav1.Object) runtime.Object

func DefaultList(objList runtime.Object, q *query.Query, compareFunc CompareFunc, filterFunc FilterFunc, transformFuncs ...TransformFunc) *api.ListResult {
	// selected matched ones
	var filtered []runtime.Object

	meta.EachListItem(objList, func(obj runtime.Object) error {
		selected := true
		o, err := meta.Accessor(obj)
		if err != nil {
			return err
		}
		for field, value := range q.Filters {

			if !filterFunc(o, query.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			for _, transform := range transformFuncs {
				obj = transform(o)
			}
			filtered = append(filtered, obj)
		}
		return nil
	})

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		l, err := meta.Accessor(filtered[i])
		if err != nil {
			return false
		}
		r, err := meta.Accessor(filtered[j])
		if err != nil {
			return false
		}
		if !q.Ascending {
			return compareFunc(l, r, q.SortBy)
		}
		return !compareFunc(l, r, q.SortBy)
	})

	total := len(filtered)

	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}

	start, end := q.Pagination.GetValidPagination(total)

	return &api.ListResult{
		TotalItems: len(filtered),
		Items:      objectsToInterfaces(filtered[start:end]),
	}
}

// DefaultObjectMetaCompare return true is left great than right
func DefaultObjectMetaCompare(left, right metav1.Object, sortBy query.Field) bool {
	switch sortBy {
	// ?sortBy=name
	case query.FieldName:
		return strings.Compare(left.GetName(), right.GetName()) > 0
	//	?sortBy=creationTimestamp
	default:
		fallthrough
	case query.FieldCreateTime:
		fallthrough
	case query.FieldCreationTimeStamp:
		// compare by name if creation timestamp is equal
		ltime := left.GetCreationTimestamp()
		rtime := right.GetCreationTimestamp()
		if ltime.Equal(&rtime) {
			return strings.Compare(left.GetName(), right.GetName()) > 0
		}
		return left.GetCreationTimestamp().After(right.GetCreationTimestamp().Time)
	}
}

//  Default metadata filter
func DefaultObjectMetaFilter(item metav1.Object, filter query.Filter) bool {
	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.GetName() == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		return strings.Contains(item.GetName(), string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return strings.Compare(string(item.GetUID()), string(filter.Value)) == 0
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return strings.Compare(item.GetNamespace(), string(filter.Value)) == 0
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.GetOwnerReferences() {
			if strings.Compare(string(ownerReference.UID), string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.GetOwnerReferences() {
			if strings.Compare(ownerReference.Kind, string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case query.FieldAnnotation:
		return labelMatch(item.GetAnnotations(), string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace:system-workspace
	case query.FieldLabel:
		return labelMatch(item.GetLabels(), string(filter.Value))
	default:
		return false
	}
}

func labelMatch(labels map[string]string, filter string) bool {
	fields := strings.SplitN(filter, "=", 2)
	var key, value string
	var opposite bool
	if len(fields) == 2 {
		key = fields[0]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			opposite = true
		}
		value = fields[1]
	} else {
		key = fields[0]
		value = "*"
	}
	for k, v := range labels {
		if opposite {
			if (k == key) && v != value {
				return true
			}
		} else {
			if (k == key) && (value == "*" || v == value) {
				return true
			}
		}
	}
	return false
}

func objectsToInterfaces(objs []runtime.Object) []interface{} {
	res := make([]interface{}, 0)
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}
