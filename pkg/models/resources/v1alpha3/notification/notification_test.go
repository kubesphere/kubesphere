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
	"kubesphere.io/kubesphere/pkg/apis/notification/v2alpha1"
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
			v2alpha1.ResourcesPluralDingTalkConfig,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralDingTalkReceiver,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralEmailConfig,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralEmailReceiver,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralSlackConfig,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralSlackReceiver,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralWebhookConfig,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralWebhookReceiver,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralWechatConfig,
		},
		{
			"test name filter",
			v2alpha1.ResourcesPluralWechatReceiver,
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
	case v2alpha1.ResourcesPluralDingTalkConfig:
		indexer = informer.Notification().V2alpha1().DingTalkConfigs().Informer().GetIndexer()
		getter = NewDingTalkConfigGetter
		obj = &v2alpha1.DingTalkConfig{}
	case v2alpha1.ResourcesPluralDingTalkReceiver:
		indexer = informer.Notification().V2alpha1().DingTalkReceivers().Informer().GetIndexer()
		getter = NewDingTalkReceiverGetter
		obj = &v2alpha1.DingTalkReceiver{}
	case v2alpha1.ResourcesPluralEmailConfig:
		indexer = informer.Notification().V2alpha1().EmailConfigs().Informer().GetIndexer()
		getter = NewEmailConfigGetter
		obj = &v2alpha1.EmailConfig{}
	case v2alpha1.ResourcesPluralEmailReceiver:
		indexer = informer.Notification().V2alpha1().EmailReceivers().Informer().GetIndexer()
		getter = NewEmailReceiverGetter
		obj = &v2alpha1.EmailReceiver{}
	case v2alpha1.ResourcesPluralSlackConfig:
		indexer = informer.Notification().V2alpha1().SlackConfigs().Informer().GetIndexer()
		getter = NewSlackConfigGetter
		obj = &v2alpha1.SlackConfig{}
	case v2alpha1.ResourcesPluralSlackReceiver:
		indexer = informer.Notification().V2alpha1().SlackReceivers().Informer().GetIndexer()
		getter = NewSlackReceiverGetter
		obj = &v2alpha1.SlackReceiver{}
	case v2alpha1.ResourcesPluralWebhookConfig:
		indexer = informer.Notification().V2alpha1().WebhookConfigs().Informer().GetIndexer()
		getter = NewWebhookConfigGetter
		obj = &v2alpha1.WebhookConfig{}
	case v2alpha1.ResourcesPluralWebhookReceiver:
		indexer = informer.Notification().V2alpha1().WebhookReceivers().Informer().GetIndexer()
		getter = NewWebhookReceiverGetter
		obj = &v2alpha1.WebhookReceiver{}
	case v2alpha1.ResourcesPluralWechatConfig:
		indexer = informer.Notification().V2alpha1().WechatConfigs().Informer().GetIndexer()
		getter = NewWechatConfigGetter
		obj = &v2alpha1.WechatConfig{}
	case v2alpha1.ResourcesPluralWechatReceiver:
		indexer = informer.Notification().V2alpha1().WechatReceivers().Informer().GetIndexer()
		getter = NewWechatReceiverGetter
		obj = &v2alpha1.WechatReceiver{}
	default:
		return nil, nil, errors.New("unowned type %s", key)
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
