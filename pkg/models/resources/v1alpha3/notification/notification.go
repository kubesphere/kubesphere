/*
Copyright 2019 The KubeSphere Authors.

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

package notification

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type dingtalkConfigGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewDingTalkConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &dingtalkConfigGetter{ksInformer: informer}
}

func (g *dingtalkConfigGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().DingTalkConfigs().Lister().Get(name)
}

func (g *dingtalkConfigGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().DingTalkConfigs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type dingtalkReceiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewDingTalkReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &dingtalkReceiverGetter{ksInformer: informer}
}

func (g *dingtalkReceiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().DingTalkReceivers().Lister().Get(name)
}

func (g *dingtalkReceiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().DingTalkReceivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type emailConfigGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewEmailConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &emailConfigGetter{ksInformer: informer}
}

func (g *emailConfigGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().EmailConfigs().Lister().Get(name)
}

func (g *emailConfigGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().EmailConfigs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type emailReceiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewEmailReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &emailReceiverGetter{ksInformer: informer}
}

func (g *emailReceiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().EmailReceivers().Lister().Get(name)
}

func (g *emailReceiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().EmailReceivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type slackConfigGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewSlackConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &slackConfigGetter{ksInformer: informer}
}

func (g *slackConfigGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().SlackConfigs().Lister().Get(name)
}

func (g *slackConfigGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().SlackConfigs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type slackReceiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewSlackReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &slackReceiverGetter{ksInformer: informer}
}

func (g *slackReceiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().SlackReceivers().Lister().Get(name)
}

func (g *slackReceiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().SlackReceivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type webhookConfigGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewWebhookConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &webhookConfigGetter{ksInformer: informer}
}

func (g *webhookConfigGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().WebhookConfigs().Lister().Get(name)
}

func (g *webhookConfigGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().WebhookConfigs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type webhookReceiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewWebhookReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &webhookReceiverGetter{ksInformer: informer}
}

func (g *webhookReceiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().WebhookReceivers().Lister().Get(name)
}

func (g *webhookReceiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().WebhookReceivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type wechatConfigGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewWechatConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &wechatConfigGetter{ksInformer: informer}
}

func (g *wechatConfigGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().WechatConfigs().Lister().Get(name)
}

func (g *wechatConfigGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().WechatConfigs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type wechatReceiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewWechatReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &wechatReceiverGetter{ksInformer: informer}
}

func (g *wechatReceiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2alpha1().WechatReceivers().Lister().Get(name)
}

func (g *wechatReceiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2alpha1().WechatReceivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

func compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftObj, err := meta.Accessor(left)
	if err != nil {
		return false
	}

	rightObj, err := meta.Accessor(right)
	if err != nil {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(meta.AsPartialObjectMetadata(leftObj).ObjectMeta,
		meta.AsPartialObjectMetadata(rightObj).ObjectMeta, field)
}

func filter(object runtime.Object, filter query.Filter) bool {

	accessor, err := meta.Accessor(object)
	if err != nil {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(meta.AsPartialObjectMetadata(accessor).ObjectMeta, filter)
}
