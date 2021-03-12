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

package secret

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"

	"k8s.io/api/core/v1"
)

type secretSearcher struct {
	informers informers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory) v1alpha3.Interface {
	return &secretSearcher{informers: informers}
}

func (s *secretSearcher) Get(namespace, name string) (runtime.Object, error) {
	return s.informers.Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
}

func (s *secretSearcher) List(namespace string, query *query.Query) (*api.ListResult, error) {
	secrets, err := s.informers.Core().V1().Secrets().Lister().Secrets(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, secret := range secrets {
		result = append(result, secret)
	}

	return v1alpha3.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *secretSearcher) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftSecret, ok := left.(*v1.Secret)
	if !ok {
		return false
	}

	rightSecret, ok := right.(*v1.Secret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftSecret.ObjectMeta, rightSecret.ObjectMeta, field)
}

func (s *secretSearcher) filter(object runtime.Object, filter query.Filter) bool {
	secret, ok := object.(*v1.Secret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(secret.ObjectMeta, filter)
}
