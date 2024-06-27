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

package loginrecord

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const recordType = "type"

type loginRecordsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &loginRecordsGetter{cache: cache}
}

func (d *loginRecordsGetter) Get(_, name string) (runtime.Object, error) {
	loginRecord := &iamv1beta1.LoginRecord{}
	return loginRecord, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, loginRecord)
}

func (d *loginRecordsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	loginRecords := &iamv1beta1.LoginRecordList{}
	if err := d.cache.List(context.Background(), loginRecords,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range loginRecords.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *loginRecordsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftRecord, ok := left.(*iamv1beta1.LoginRecord)
	if !ok {
		return false
	}

	rightRecord, ok := right.(*iamv1beta1.LoginRecord)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRecord.ObjectMeta, rightRecord.ObjectMeta, field)
}

func (d *loginRecordsGetter) filter(object runtime.Object, filter query.Filter) bool {
	record, ok := object.(*iamv1beta1.LoginRecord)

	if !ok {
		return false
	}

	switch filter.Field {
	case recordType:
		return string(record.Spec.Type) == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(record.ObjectMeta, filter)
	}

}
