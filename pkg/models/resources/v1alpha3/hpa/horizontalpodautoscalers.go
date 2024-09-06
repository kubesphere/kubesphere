/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package hpa

import (
	"context"

	"github.com/Masterminds/semver/v3"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

type hpaGetter struct {
	cache   runtimeclient.Reader
	gvk     schema.GroupVersionKind
	listGVK schema.GroupVersionKind
}

func New(cache runtimeclient.Reader, k8sVersion *semver.Version) v1alpha3.Interface {
	gvk := autoscalingv2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler")
	listGVK := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	if k8sutil.ServeAutoscalingV2beta2(k8sVersion) {
		gvk.Version = "v2beta2"
		listGVK.Version = "v2beta2"
	}
	return &hpaGetter{
		cache:   cache,
		gvk:     gvk,
		listGVK: listGVK,
	}
}

func (s *hpaGetter) Get(namespace, name string) (runtime.Object, error) {
	obj, err := scheme.Scheme.New(s.gvk)
	if err != nil {
		return nil, err
	}
	hpa := obj.(client.Object)
	return hpa, s.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, hpa)
}

func (s *hpaGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	listObj := &unstructured.UnstructuredList{}
	listObj.SetGroupVersionKind(s.listGVK)

	if err := s.cache.List(context.Background(), listObj, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	if err := listObj.EachListItem(func(object runtime.Object) error {
		result = append(result, object)
		return nil
	}); err != nil {
		return nil, err
	}
	return v1alpha3.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *hpaGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftHPA, ok := left.(*unstructured.Unstructured)
	if !ok {
		return false
	}

	rightHPA, ok := right.(*unstructured.Unstructured)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(k8sutil.GetObjectMeta(leftHPA), k8sutil.GetObjectMeta(rightHPA), field)
}

func (s *hpaGetter) filter(object runtime.Object, filter query.Filter) bool {
	hpa, ok := object.(*unstructured.Unstructured)
	if !ok {
		return false
	}

	targetKind, _, _ := unstructured.NestedString(hpa.UnstructuredContent(), "spec", "scaleTargetRef", "kind")
	targetName, _, _ := unstructured.NestedString(hpa.UnstructuredContent(), "spec", "scaleTargetRef", "name")

	switch filter.Field {
	case "targetKind":
		return targetKind == string(filter.Value)
	case "targetName":
		return targetName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(k8sutil.GetObjectMeta(hpa), filter)
	}
}
