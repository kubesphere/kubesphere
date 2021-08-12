/*
Copyright 2020 KubeSphere Authors

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

package client

import (
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ClientCache interface {
	GetObjMeta(obj runtime.Object) (*ObjMeta, error)
	GetResource(obj runtime.Object) (*ResourceMeta, error)
}

func NewClientCache(config *rest.Config, options Options) ClientCache {

	cc := &clientCache{
		config:         config,
		scheme:         options.Scheme,
		mapper:         options.Mapper,
		codecs:         serializer.NewCodecFactory(options.Scheme),
		resourceByType: make(map[schema.GroupVersionKind]*ResourceMeta),
	}

	return cc
}

// clientCache creates and caches rest clients and metadata for Kubernetes types
type clientCache struct {
	// config is the rest.Config to talk to an apiserver
	config *rest.Config

	// scheme maps go structs to GroupVersionKinds
	scheme *runtime.Scheme

	// mapper maps GroupVersionKinds to Resources
	mapper meta.RESTMapper

	// codecs are used to create a REST client for a gvk
	codecs serializer.CodecFactory

	// resourceByType caches type metadata
	resourceByType map[schema.GroupVersionKind]*ResourceMeta
	mu             sync.RWMutex
}

// newResource maps obj to a Kubernetes Resource and constructs a client for that Resource.
// If the object is a list, the resource represents the item's type instead.
func (c *clientCache) newResource(gvk schema.GroupVersionKind, isList bool) (*ResourceMeta, error) {
	if strings.HasSuffix(gvk.Kind, "List") && isList {
		// if this was a list, treat it as a request for the item's resource
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	client, err := apiutil.RESTClientForGVK(gvk, false, c.config, c.codecs)
	if err != nil {
		return nil, err
	}
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &ResourceMeta{Interface: client, mapping: mapping, gvk: gvk}, nil
}

// GetResource returns the resource meta information for the given type of object.
// If the object is a list, the resource represents the item's type instead.
func (c *clientCache) GetResource(obj runtime.Object) (*ResourceMeta, error) {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return nil, err
	}

	// It's better to do creation work twice than to not let multiple
	// people make requests at once
	c.mu.RLock()
	r, known := c.resourceByType[gvk]
	c.mu.RUnlock()

	if known {
		return r, nil
	}

	// Initialize a new Client
	c.mu.Lock()
	defer c.mu.Unlock()
	r, err = c.newResource(gvk, meta.IsListType(obj))
	if err != nil {
		return nil, err
	}
	c.resourceByType[gvk] = r
	return r, err
}

// getObjMeta returns ObjMeta containing both type and object metadata and state
func (c *clientCache) GetObjMeta(obj runtime.Object) (*ObjMeta, error) {
	r, err := c.GetResource(obj)
	if err != nil {
		return nil, err
	}
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return &ObjMeta{ResourceMeta: r, Object: m}, err
}

// ResourceMeta caches state for a Kubernetes type.
type ResourceMeta struct {
	// client is the rest client used to talk to the apiserver
	rest.Interface
	// gvk is the GroupVersionKind of the ResourceMeta
	gvk schema.GroupVersionKind
	// mapping is the rest mapping
	mapping *meta.RESTMapping
}

// IsNamespaced returns true if the type is namespaced
func (r *ResourceMeta) IsNamespaced() bool {
	return r.mapping.Scope.Name() != meta.RESTScopeNameRoot

}

// resource returns the resource name of the type
func (r *ResourceMeta) Resource() string {
	return r.mapping.Resource.Resource
}

// ObjMeta stores type and object information about a Kubernetes type
type ObjMeta struct {
	// ResourceMeta contains type information for the object
	*ResourceMeta

	// Object contains meta data for the object instance
	metav1.Object
}
