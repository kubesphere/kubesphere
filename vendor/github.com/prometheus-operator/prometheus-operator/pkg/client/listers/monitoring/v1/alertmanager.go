// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AlertmanagerLister helps list Alertmanagers.
type AlertmanagerLister interface {
	// List lists all Alertmanagers in the indexer.
	List(selector labels.Selector) (ret []*v1.Alertmanager, err error)
	// Alertmanagers returns an object that can list and get Alertmanagers.
	Alertmanagers(namespace string) AlertmanagerNamespaceLister
	AlertmanagerListerExpansion
}

// alertmanagerLister implements the AlertmanagerLister interface.
type alertmanagerLister struct {
	indexer cache.Indexer
}

// NewAlertmanagerLister returns a new AlertmanagerLister.
func NewAlertmanagerLister(indexer cache.Indexer) AlertmanagerLister {
	return &alertmanagerLister{indexer: indexer}
}

// List lists all Alertmanagers in the indexer.
func (s *alertmanagerLister) List(selector labels.Selector) (ret []*v1.Alertmanager, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Alertmanager))
	})
	return ret, err
}

// Alertmanagers returns an object that can list and get Alertmanagers.
func (s *alertmanagerLister) Alertmanagers(namespace string) AlertmanagerNamespaceLister {
	return alertmanagerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AlertmanagerNamespaceLister helps list and get Alertmanagers.
type AlertmanagerNamespaceLister interface {
	// List lists all Alertmanagers in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.Alertmanager, err error)
	// Get retrieves the Alertmanager from the indexer for a given namespace and name.
	Get(name string) (*v1.Alertmanager, error)
	AlertmanagerNamespaceListerExpansion
}

// alertmanagerNamespaceLister implements the AlertmanagerNamespaceLister
// interface.
type alertmanagerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Alertmanagers in the indexer for a given namespace.
func (s alertmanagerNamespaceLister) List(selector labels.Selector) (ret []*v1.Alertmanager, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Alertmanager))
	})
	return ret, err
}

// Get retrieves the Alertmanager from the indexer for a given namespace and name.
func (s alertmanagerNamespaceLister) Get(name string) (*v1.Alertmanager, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("alertmanager"), name)
	}
	return obj.(*v1.Alertmanager), nil
}
