package resource

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Interface interface {
	GetResource(gvr schema.GroupVersionResource, name, namespace string) (client.Object, error)
	ListResources(gvr schema.GroupVersionResource, query *query.Query) (client.ObjectList, error)
}
