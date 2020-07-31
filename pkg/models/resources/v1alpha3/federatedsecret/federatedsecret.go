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

package federatedsecret

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
)

type fedSecretGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedSecretGetter{sharedInformers: sharedInformers}
}

func (d *fedSecretGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Types().V1beta1().FederatedSecrets().Lister().FederatedSecrets(namespace).Get(name)
}

func (d *fedSecretGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	secrets, err := d.sharedInformers.Types().V1beta1().FederatedSecrets().Lister().FederatedSecrets(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, secret := range secrets {
		result = append(result, secret)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *fedSecretGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftSecret, ok := left.(*v1beta1.FederatedSecret)
	if !ok {
		return false
	}

	rightSecret, ok := right.(*v1beta1.FederatedSecret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftSecret.ObjectMeta, rightSecret.ObjectMeta, field)
}

func (d *fedSecretGetter) filter(object runtime.Object, filter query.Filter) bool {
	fedSecret, ok := object.(*v1beta1.FederatedSecret)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldType:
		return strings.Compare(string(fedSecret.Spec.Template.Type), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(fedSecret.ObjectMeta, filter)
	}
}
