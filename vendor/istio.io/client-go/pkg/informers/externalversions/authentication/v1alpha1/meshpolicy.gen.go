// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	versioned "istio.io/client-go/pkg/clientset/versioned"
	internalinterfaces "istio.io/client-go/pkg/informers/externalversions/internalinterfaces"
	v1alpha1 "istio.io/client-go/pkg/listers/authentication/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MeshPolicyInformer provides access to a shared informer and lister for
// MeshPolicies.
type MeshPolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MeshPolicyLister
}

type meshPolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewMeshPolicyInformer constructs a new informer for MeshPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMeshPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMeshPolicyInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredMeshPolicyInformer constructs a new informer for MeshPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMeshPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AuthenticationV1alpha1().MeshPolicies().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AuthenticationV1alpha1().MeshPolicies().Watch(options)
			},
		},
		&authenticationv1alpha1.MeshPolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *meshPolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMeshPolicyInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *meshPolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&authenticationv1alpha1.MeshPolicy{}, f.defaultInformer)
}

func (f *meshPolicyInformer) Lister() v1alpha1.MeshPolicyLister {
	return v1alpha1.NewMeshPolicyLister(f.Informer().GetIndexer())
}
