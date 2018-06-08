/*
Copyright 2018 The KubeSphere Authors.

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

package resources

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type resource interface {
	equal(object resource) bool
}

type resourcetList interface {
	update(key string, value resource)
	del(key string, value resource)
}

type ResourceStatus struct {
	ResourceType    string        `json:"type"`
	ResourceList    resourcetList `json:"lists"`
	UpdateTimeStamp int64         `json:"updateTimestamp"`
}

type ResourceChan struct {
	Type       string
	StatusChan chan *ResourceStatus
	StopChan   chan struct{}
}

type Resource interface {
	list() (interface{}, error)
	getWatcher() (watch.Interface, error)
	updateWithObjects(workload *ResourceStatus, objects interface{})
	updateWithEvent(workload *ResourceStatus, event watch.Event)
}

type WorkLoadObject struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	App          string            `json:"app"`
	Available    int32             `json:"available"`
	Desire       int32             `json:"desire"`
	Ready        bool              `json:"ready"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	UpdateTime   meta_v1.Time      `json:"updateTime,omitempty"`
	CreateTime   meta_v1.Time      `json:"createTime,omitempty"`
}

type OtherResourceObject struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Resources map[string][]resource

func (resources Resources) update(namespace string, object resource) {
	for index, tmpObject := range resources[namespace] {
		if tmpObject.equal(object) {
			resources[namespace][index] = object
			return
		}
	}

	resources[namespace] = append(resources[namespace], object)
}

func (resources Resources) del(namespace string, object resource) {
	for index, tmpObject := range resources[namespace] {
		if tmpObject.equal(object) {
			resources[namespace] = append(resources[namespace][:index], resources[namespace][index+1:]...)
			return
		}
	}
}

func (workLoadObject WorkLoadObject) equal(object resource) bool {
	tmp := object.(WorkLoadObject)
	if workLoadObject.Name == tmp.Name && workLoadObject.Namespace == tmp.Namespace {
		return true
	}
	return false
}

func (otherResourceObject OtherResourceObject) equal(object resource) bool {
	tmp := object.(OtherResourceObject)
	if otherResourceObject.Name == tmp.Name && otherResourceObject.Namespace == tmp.Namespace {
		return true
	}
	return false
}
