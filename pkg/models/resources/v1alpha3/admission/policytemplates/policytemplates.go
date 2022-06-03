/*
Copyright 2021 The KubeSphere Authors.

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

package policytemplates

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	admissionv1alpha1 "kubesphere.io/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type policyTemplateGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func New(f ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &policyTemplateGetter{f}
}

func (g *policyTemplateGetter) Get(_, name string) (runtime.Object, error) {
	template, err := g.ksInformer.Admission().V1alpha1().PolicyTemplates().Lister().Get(name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return template, nil
}

func (g *policyTemplateGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	templates, err := g.ksInformer.Admission().V1alpha1().PolicyTemplates().Lister().List(query.Selector())
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var result []runtime.Object
	for _, p := range templates {
		result = append(result, p)
	}
	return v1alpha3.DefaultList(result, query, g.compare, g.filter), nil
}

func (g *policyTemplateGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftTemplate, ok := left.(*admissionv1alpha1.PolicyTemplate)
	if !ok {
		return false
	}
	rightTemplate, ok := right.(*admissionv1alpha1.PolicyTemplate)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftTemplate.ObjectMeta, rightTemplate.ObjectMeta, field)
}

func (g *policyTemplateGetter) filter(object runtime.Object, filter query.Filter) bool {
	p, ok := object.(*admissionv1alpha1.PolicyTemplate)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaFilter(p.ObjectMeta, filter)
}
