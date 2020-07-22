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
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const recordType = "type"

type loginrecordsGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &loginrecordsGetter{ksInformer: ksinformer}
}

func (d *loginrecordsGetter) Get(_, name string) (runtime.Object, error) {
	return d.ksInformer.Iam().V1alpha2().Users().Lister().Get(name)
}

func (d *loginrecordsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	records, err := d.ksInformer.Iam().V1alpha2().LoginRecords().Lister().List(query.Selector())

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, user := range records {
		result = append(result, user)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *loginrecordsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftUser, ok := left.(*iamv1alpha2.LoginRecord)
	if !ok {
		return false
	}

	rightUser, ok := right.(*iamv1alpha2.LoginRecord)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftUser.ObjectMeta, rightUser.ObjectMeta, field)
}

func (d *loginrecordsGetter) filter(object runtime.Object, filter query.Filter) bool {
	record, ok := object.(*iamv1alpha2.LoginRecord)

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
