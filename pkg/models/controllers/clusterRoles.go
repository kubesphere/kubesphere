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

package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (ctl *ClusterRoleCtl) generateObjec(item v1.ClusterRole) *ClusterRole {
	name := item.Name
	if strings.HasPrefix(name, "system:") {
		return nil
	}

	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	annotation, _ := json.Marshal(item.Annotations)

	object := &ClusterRole{Name: name, CreateTime: createTime, AnnotationStr: string(annotation)}

	return object
}

func (ctl *ClusterRoleCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&ClusterRole{}) {
		db.DropTable(&ClusterRole{})

	}

	db = db.CreateTable(&ClusterRole{})

	k8sClient := ctl.K8sClient

	clusterRoleList, err := k8sClient.RbacV1().ClusterRoles().List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range clusterRoleList.Items {
		obj := ctl.generateObjec(item)
		if obj != nil {
			db.Create(obj)
		}
	}

	clusterRoleWatcher, err := k8sClient.RbacV1().ClusterRoles().Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-clusterRoleWatcher.ResultChan():
			var role ClusterRole
			if event.Object == nil {
				panic("watch timeout, restart clusterRole controller")
			}
			object := event.Object.(*v1.ClusterRole)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, "\"\"").Find(&role)
				db.Delete(role)
				break
			}
			obj := ctl.generateObjec(*object)
			if obj != nil {
				db.Save(obj)
			}
		}
	}
}

func (ctl *ClusterRoleCtl) CountWithConditions(conditions string) int {
	var object ClusterRole

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *ClusterRoleCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var object ClusterRole
	var list []ClusterRole
	var total int

	order := "createTime desc"
	db := ctl.DB

	listWithConditions(db, &total, &object, &list, conditions, paging, order)

	for index, item := range list {
		annotation := make(map[string]string)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
	return total, list, nil
}

func (ctl *ClusterRoleCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&ClusterRole{}).Count(&count)
	return count
}
