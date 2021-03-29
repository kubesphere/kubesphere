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

package application

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type applicationsGetter struct {
	c cache.Cache
}

func New(c cache.Cache) v1alpha3.Interface {
	return &applicationsGetter{c}
}

func (d *applicationsGetter) Get(namespace, name string) (runtime.Object, error) {
	app := appv1beta1.Application{}
	err := d.c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &app)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &app, nil
}

func (d *applicationsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	applications := appv1beta1.ApplicationList{}
	err := d.c.List(context.Background(), &applications, &client.ListOptions{Namespace: namespace, LabelSelector: query.Selector()})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var result []runtime.Object
	for i := range applications.Items {
		result = append(result, &applications.Items[i])
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *applicationsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftApplication, ok := left.(*appv1beta1.Application)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*appv1beta1.Application)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftApplication).After(lastUpdateTime(rightApplication))
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftApplication.ObjectMeta, rightApplication.ObjectMeta, field)
	}
}

func (d *applicationsGetter) filter(object runtime.Object, filter query.Filter) bool {
	application, ok := object.(*appv1beta1.Application)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(application.ObjectMeta, filter)
}

func lastUpdateTime(application *appv1beta1.Application) time.Time {
	lut := application.CreationTimestamp.Time
	for _, condition := range application.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
