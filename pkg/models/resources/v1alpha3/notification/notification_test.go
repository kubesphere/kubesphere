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
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"kubesphere.io/kubesphere/pkg/api"
	v2 "kubesphere.io/kubesphere/pkg/apis/notification/v2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"math/rand"
	"sort"
	"testing"
)

const (
	Prefix    = "foo"
	LengthMin = 3
	LengthMax = 10
)

func TestListObjects(t *testing.T) {
	tests := []struct {
		description string
		key         string
	}{
		{
			"test name filter",
			v2.ResourcesPluralDingTalkConfig,
		},
		{
			"test name filter",
			v2.ResourcesPluralDingTalkReceiver,
		},
		{
			"test name filter",
			v2.ResourcesPluralEmailConfig,
		},
		{
			"test name filter",
			v2.ResourcesPluralEmailReceiver,
		},
		{
			"test name filter",
			v2.ResourcesPluralSlackConfig,
		},
		{
			"test name filter",
			v2.ResourcesPluralSlackReceiver,
		},
		{
			"test name filter",
			v2.ResourcesPluralWebhookConfig,
		},
		{
			"test name filter",
			v2.ResourcesPluralWebhookReceiver,
		},
		{
			"test name filter",
			v2.ResourcesPluralWechatConfig,
		},
		{
			"test name filter",
			v2.ResourcesPluralWechatReceiver,
		},
	}

	q := &query.Query{
		Pagination: &query.Pagination{
			Limit:  10,
			Offset: 0,
		},
		SortBy:    query.FieldName,
		Ascending: true,
		Filters:   map[query.Field]query.Value{query.FieldName: query.Value(Prefix)},
	}

	for _, test := range tests {

		getter, objects, err := prepare(test.key)
		if err != nil {
			t.Fatal(err)
		}

		got, err := getter.List("", q)
		if err != nil {
			t.Fatal(err)
		}

		expected := &api.ListResult{
			Items:      objects,
			TotalItems: len(objects),
		}

		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("[%s] %T differ (-got, +want): %s", test.description, expected, diff)
		}
	}
}

func prepare(key string) (v1alpha3.Interface, []interface{}, error) {
	client := fake.NewSimpleClientset()
	informer := ksinformers.NewSharedInformerFactory(client, 0)

	var obj runtime.Object
	var indexer cache.Indexer
	var getter func(informer ksinformers.SharedInformerFactory) v1alpha3.Interface
	switch key {
	case v2.ResourcesPluralDingTalkConfig:
		indexer = informer.Notification().V2().DingTalkConfigs().Informer().GetIndexer()
		getter = NewDingTalkConfigGetter
		obj = &v2.DingTalkConfig{}
	case v2.ResourcesPluralDingTalkReceiver:
		indexer = informer.Notification().V2().DingTalkReceivers().Informer().GetIndexer()
		getter = NewDingTalkReceiverGetter
		obj = &v2.DingTalkReceiver{}
	case v2.ResourcesPluralEmailConfig:
		indexer = informer.Notification().V2().EmailConfigs().Informer().GetIndexer()
		getter = NewEmailConfigGetter
		obj = &v2.EmailConfig{}
	case v2.ResourcesPluralEmailReceiver:
		indexer = informer.Notification().V2().EmailReceivers().Informer().GetIndexer()
		getter = NewEmailReceiverGetter
		obj = &v2.EmailReceiver{}
	case v2.ResourcesPluralSlackConfig:
		indexer = informer.Notification().V2().SlackConfigs().Informer().GetIndexer()
		getter = NewSlackConfigGetter
		obj = &v2.SlackConfig{}
	case v2.ResourcesPluralSlackReceiver:
		indexer = informer.Notification().V2().SlackReceivers().Informer().GetIndexer()
		getter = NewSlackReceiverGetter
		obj = &v2.SlackReceiver{}
	case v2.ResourcesPluralWebhookConfig:
		indexer = informer.Notification().V2().WebhookConfigs().Informer().GetIndexer()
		getter = NewWebhookConfigGetter
		obj = &v2.WebhookConfig{}
	case v2.ResourcesPluralWebhookReceiver:
		indexer = informer.Notification().V2().WebhookReceivers().Informer().GetIndexer()
		getter = NewWebhookReceiverGetter
		obj = &v2.WebhookReceiver{}
	case v2.ResourcesPluralWechatConfig:
		indexer = informer.Notification().V2().WechatConfigs().Informer().GetIndexer()
		getter = NewWechatConfigGetter
		obj = &v2.WechatConfig{}
	case v2.ResourcesPluralWechatReceiver:
		indexer = informer.Notification().V2().WechatReceivers().Informer().GetIndexer()
		getter = NewWechatReceiverGetter
		obj = &v2.WechatReceiver{}
	default:
		return nil, nil, errors.New("unkonwed type %s", key)
	}

	num := rand.Intn(LengthMax)
	if num < LengthMin {
		num = LengthMin
	}

	var suffix []string
	for i := 0; i < num; i++ {
		s := uuid.New().String()
		suffix = append(suffix, s)
	}
	sort.Strings(suffix)

	var objects []interface{}
	for i := 0; i < num; i++ {
		val := obj.DeepCopyObject()
		accessor, err := meta.Accessor(val)
		if err != nil {
			return nil, nil, err
		}

		accessor.SetName(Prefix + "-" + suffix[i])
		err = indexer.Add(accessor)
		if err != nil {
			return nil, nil, err
		}
		objects = append(objects, val)
	}

	return getter(informer), objects, nil
}
