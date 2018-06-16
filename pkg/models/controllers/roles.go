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
	"kubesphere.io/kubesphere/pkg/client"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/api/rbac/v1"
	"time"
	"strings"
	"encoding/json"
)

func(ctl *RoleCtl) generateObject(item v1.Role) *Role{
	name := item.Name
	if strings.HasPrefix(name, "system:"){
		return nil
	}
	namespace := item.Namespace
	createTime := item.CreationTimestamp.Time
	if createTime.IsZero(){
		createTime = time.Now()
	}

	annotation, _ := json.Marshal(item.Annotations)

	object := &Role{Namespace: namespace, Name: name, CreateTime:createTime, AnnotationStr:string(annotation)}

	return object
}

func(ctl *RoleCtl) listAndWatch()  {
	defer func(){
		close(ctl.aliveChan)
		if err := recover(); err != nil{
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Role{}){
		db.DropTable(&Role{})

	}

	db = db.CreateTable(&Role{})

	k8sClient := client.NewK8sClient()
	roleList, err := k8sClient.RbacV1().Roles("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range roleList.Items {
		obj := ctl.generateObject(item)
		if obj != nil{
			db.Create(obj)
		}

	}

	roleWatcher, err := k8sClient.RbacV1().Roles("").Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}


	for{
		select{
		case <- ctl.stopChan:
			return
		case event := <- roleWatcher.ResultChan():
			var role Role
			if event.Object == nil{
				panic("watch timeout, restart role controller")
			}
			object := event.Object.(*v1.Role)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&role)
				db.Delete(role)
				break
			}
			obj := ctl.generateObject(*object)
			if obj != nil{
				db.Save(obj)
			}
			break
		}
	}
}


func(ctl *RoleCtl) CountWithConditions(conditions string) int {
	var object Role

	return countWithConditions(ctl.DB, conditions, &object)
}


func(ctl *RoleCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Role
	var object Role
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index, item := range list{
		annotation := make(map[string]string)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
	return total, list, nil
}

func(ctl *RoleCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&Role{}).Where("namespace = ?", namespace).Count(&count)
	return count
}