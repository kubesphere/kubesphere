/*
Copyright 2020 The KubeSphere Authors.

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

// Code generated by informer-gen. DO NOT EDIT.

package v2

import (
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	notificationv2 "kubesphere.io/kubesphere/pkg/apis/notification/v2"
	versioned "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	internalinterfaces "kubesphere.io/kubesphere/pkg/client/informers/externalversions/internalinterfaces"
	v2 "kubesphere.io/kubesphere/pkg/client/listers/notification/v2"
)

// WechatReceiverInformer provides access to a shared informer and lister for
// WechatReceivers.
type WechatReceiverInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v2.WechatReceiverLister
}

type wechatReceiverInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewWechatReceiverInformer constructs a new informer for WechatReceiver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWechatReceiverInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWechatReceiverInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredWechatReceiverInformer constructs a new informer for WechatReceiver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWechatReceiverInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NotificationV2().WechatReceivers().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NotificationV2().WechatReceivers().Watch(context.TODO(), options)
			},
		},
		&notificationv2.WechatReceiver{},
		resyncPeriod,
		indexers,
	)
}

func (f *wechatReceiverInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWechatReceiverInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *wechatReceiverInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&notificationv2.WechatReceiver{}, f.defaultInformer)
}

func (f *wechatReceiverInformer) Lister() v2.WechatReceiverLister {
	return v2.NewWechatReceiverLister(f.Informer().GetIndexer())
}