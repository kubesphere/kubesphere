/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

const (
	defaultRegoQuery    = "data.filter.match"
	defaultRegoFileName = "filter.rego"
	labelResourceServed = "kubesphere.io/resource-served"
	defaultConfigFile   = "configuration.yaml"
	defaultResyncPeriod = 10 * time.Hour
)

// Note:
// If the Custom Resource Definition (CRD) is deleted in the cluster while it is running,
// the `client.cache` may not return an error but instead provide an empty result.

func New(ctx context.Context, runtimeClient client.Client, runtimeCache cache.Cache) (ResourceManager, error) {
	secretInformer, err := runtimeCache.GetInformer(ctx, &corev1.Secret{})
	if err != nil {
		return nil, fmt.Errorf("failed to get informer for secret: %v", err)
	}

	resourceManager := &resourceManager{
		client: runtimeClient,
		ctx:    ctx,
	}

	_, err = secretInformer.AddEventHandlerWithResyncPeriod(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			secret := obj.(*corev1.Secret)

			customFilter := customResourceFilterFromSecret(secret)
			if customFilter != nil {
				resourceManager.onCustomResourceFilterChange(customFilter)
			}
		},
		UpdateFunc: func(_, obj interface{}) {
			secret := obj.(*corev1.Secret)
			customFilter := customResourceFilterFromSecret(secret)
			if customFilter != nil {
				resourceManager.onCustomResourceFilterChange(customFilter)
			}
		},
		DeleteFunc: func(obj interface{}) {
			secret := obj.(*corev1.Secret)
			customFilter := customResourceFilterFromSecret(secret)
			if customFilter != nil {
				resourceManager.onCustomResourceFilterDelete(customFilter)
			}
		},
	}, defaultResyncPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler for secret: %v", err)
	}

	return resourceManager, nil
}

type resourceManager struct {
	client                client.Client
	ctx                   context.Context
	customResourceFilters sync.Map
}

func (h *resourceManager) GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (client.Object, error) {
	var obj client.Object
	gvk, err := h.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	if h.client.Scheme().Recognizes(gvk) {
		gvkObject, err := h.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	} else {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := h.Get(ctx, namespace, name, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (h *resourceManager) CreateObjectFromRawData(gvr schema.GroupVersionResource, rawData []byte) (client.Object, error) {
	var obj client.Object
	gvk, err := h.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	if h.client.Scheme().Recognizes(gvk) {
		gvkObject, err := h.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	} else {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return nil, err
	}

	// The object's GroupVersionKind could be overridden if apiVersion and kind of rawData are different
	// with GroupVersionKind from url, so that we should check GroupVersionKind after Unmarshal rawDate.
	if obj.GetObjectKind().GroupVersionKind().String() != gvk.String() {
		return nil, errors.NewBadRequest("wrong resource GroupVersionKind")
	}

	return obj, nil
}

func (h *resourceManager) ListResources(ctx context.Context, gvr schema.GroupVersionResource, namespace string, query *query.Query) (client.ObjectList, error) {
	var obj client.ObjectList

	gvk, err := h.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	gvk = convertGVKToList(gvk)

	if h.client.Scheme().Recognizes(gvk) {
		gvkObject, err := h.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.ObjectList)
	} else {
		u := &unstructured.UnstructuredList{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := h.List(ctx, namespace, query, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (h *resourceManager) DeleteResource(ctx context.Context, object client.Object) error {
	return h.Delete(ctx, object)
}

func (h *resourceManager) UpdateResource(ctx context.Context, object client.Object) error {
	return h.Update(ctx, object)
}

func (h *resourceManager) PatchResource(ctx context.Context, object client.Object) error {
	old := object.DeepCopyObject().(client.Object)
	err := h.Get(ctx, object.GetNamespace(), object.GetName(), old)
	if err != nil {
		return err
	}
	return h.Patch(ctx, old, object)
}

func (h *resourceManager) CreateResource(ctx context.Context, object client.Object) error {
	return h.Create(ctx, object)
}

func convertGVKToList(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	if strings.HasSuffix(gvk.Kind, "List") {
		return gvk
	}
	gvk.Kind = gvk.Kind + "List"
	return gvk
}

func (h *resourceManager) getGVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	var (
		gvk schema.GroupVersionKind
		err error
	)
	gvk, err = h.client.RESTMapper().KindFor(gvr)
	if err != nil {
		return gvk, err
	}
	return gvk, nil
}

func (h *resourceManager) IsServed(gvr schema.GroupVersionResource) (bool, error) {
	// well-known group version is already registered
	if h.client.Scheme().IsVersionRegistered(gvr.GroupVersion()) {
		return true, nil
	}

	crd := &extv1.CustomResourceDefinition{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Name: gvr.GroupResource().String()}, crd); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if crd.Labels[labelResourceServed] == "true" {
		return true, nil
	}

	return false, nil
}

func (h *resourceManager) Get(ctx context.Context, namespace, name string, object client.Object) error {
	return h.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, object)
}

func (h *resourceManager) List(ctx context.Context, namespace string, query *query.Query, list client.ObjectList) error {
	listOpt := &client.ListOptions{
		LabelSelector: query.Selector(),
		Namespace:     namespace,
	}

	if err := h.client.List(ctx, list, listOpt); err != nil {
		return err
	}

	extractList, err := meta.ExtractList(list)
	if err != nil {
		return err
	}

	filtered, remainingItemCount, total := DefaultList(extractList, query, DefaultCompare, h.CustomResourceFilter)
	remaining := int64(remainingItemCount)
	list.SetRemainingItemCount(&remaining)
	list.SetContinue(strconv.Itoa(total))
	if err := meta.SetList(list, filtered); err != nil {
		return err
	}
	return nil
}

func (h *resourceManager) Create(ctx context.Context, object client.Object) error {
	return h.client.Create(ctx, object)
}

func (h *resourceManager) Delete(ctx context.Context, object client.Object) error {
	return h.client.Delete(ctx, object)
}

func (h *resourceManager) Update(ctx context.Context, new client.Object) error {
	return h.client.Update(ctx, new)
}

func (h *resourceManager) Patch(ctx context.Context, old, new client.Object) error {
	new.SetResourceVersion(old.GetResourceVersion())
	return h.client.Patch(ctx, new, client.MergeFrom(old))
}

func DefaultCompare(left, right runtime.Object, field query.Field) bool {
	l, err := meta.Accessor(left)
	if err != nil {
		return false
	}
	r, err := meta.Accessor(right)
	if err != nil {
		return false
	}
	return DefaultObjectMetaCompare(l, r, field)
}

type Input struct {
	Filter query.Filter   `json:"filter"`
	Object runtime.Object `json:"object"`
}

func (h *resourceManager) RegoFilter(object runtime.Object, filter query.Filter, preparedEvalQuery *rego.PreparedEvalQuery) bool {
	results, err := preparedEvalQuery.Eval(h.ctx, rego.EvalInput(Input{
		Filter: filter,
		Object: object,
	}))
	if err != nil {
		klog.Warningf("failed to eval reogo query: %s", err)
		return false
	}

	if len(results) > 0 {
		if match, ok := results[0].Expressions[0].Value.(bool); ok {
			return match
		} else {
			return false
		}
	}
	return false
}

func (h *resourceManager) CustomResourceFilter(object runtime.Object, filter query.Filter) bool {
	typed, err := meta.TypeAccessor(object)
	if err != nil {
		return false
	}
	if !DefaultFilter(object, filter) {
		return false
	}
	if value, ok := h.customResourceFilters.Load(schema.FromAPIVersionAndKind(typed.GetAPIVersion(), typed.GetKind())); ok {
		preparedEvalQuery := value.(*rego.PreparedEvalQuery)
		if !h.RegoFilter(object, filter, preparedEvalQuery) {
			return false
		}
	}
	return true
}

type Resource struct {
	Group   string `json:"group" yaml:"group"`
	Version string `json:"version" yaml:"version"`
	Kind    string `json:"kind" yaml:"kind"`
}

type CustomResourceFilter struct {
	Resource   Resource `json:"resource" yaml:"resource"`
	RegoPolicy string   `json:"regoPolicy" yaml:"regoPolicy"`
}

func customResourceFilterFromSecret(secret *corev1.Secret) *CustomResourceFilter {
	if secret == nil {
		return nil
	}
	if secret.Type != "config.kubesphere.io/custom-resource-filter" {
		return nil
	}
	filter := &CustomResourceFilter{}
	if err := yaml.Unmarshal(secret.Data[defaultConfigFile], filter); err != nil {
		klog.Warningf("failed to unmarshal custom resource filter: %s", err)
		return nil
	}
	return filter
}

func (h *resourceManager) onCustomResourceFilterChange(customResourceFilter *CustomResourceFilter) {
	if customResourceFilter == nil {
		return
	}
	regoInstance := rego.New(rego.Query(defaultRegoQuery), rego.Module(defaultRegoFileName, customResourceFilter.RegoPolicy))
	if preparedEvalQuery, err := regoInstance.PrepareForEval(h.ctx); err != nil {
		klog.Warningf("failed to prepare reogo query: %s", err)
	} else {
		gvk := schema.GroupVersionKind{
			Group:   customResourceFilter.Resource.Group,
			Version: customResourceFilter.Resource.Version,
			Kind:    customResourceFilter.Resource.Kind,
		}
		h.customResourceFilters.Swap(gvk, &preparedEvalQuery)
		klog.V(4).Infof("custom resource filter for %s is updated", gvk.String())
	}
}

func (h *resourceManager) onCustomResourceFilterDelete(filter *CustomResourceFilter) {
	if filter == nil {
		return
	}
	gvk := schema.GroupVersionKind{
		Group:   filter.Resource.Group,
		Version: filter.Resource.Version,
		Kind:    filter.Resource.Kind,
	}
	h.customResourceFilters.Delete(gvk)
}

func DefaultFilter(object runtime.Object, filter query.Filter) bool {
	o, err := meta.Accessor(object)
	if err != nil {
		return false
	}
	return DefaultObjectMetaFilter(o, filter)
}
