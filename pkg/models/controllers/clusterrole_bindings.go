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
	"time"

	"fmt"
	"regexp"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *ClusterRoleBindingCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *ClusterRoleBindingCtl) sync(stopChan chan struct{}) {
	ctl.initListerAndInformer()
	ctl.informer.Run(stopChan)
}

func (ctl *ClusterRoleBindingCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *ClusterRoleBindingCtl) handleWorkspaceRoleChange(clusterRole *rbac.ClusterRoleBinding) {
	if groups := regexp.MustCompile(`^system:(\S+):(admin|operator|viewer)$`).FindStringSubmatch(clusterRole.Name); len(groups) == 3 {
		workspace := groups[1]
		go ctl.restNamespaceRoleBinding(workspace)
	}
}

func (ctl *ClusterRoleBindingCtl) restNamespaceRoleBinding(workspace string) {
	selector := labels.SelectorFromSet(labels.Set{"kubesphere.io/workspace": workspace})
	namespaces, err := ctl.K8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		glog.Warning("workspace roles sync failed", workspace, err)
		return
	}

	for _, namespace := range namespaces.Items {
		pathJson := fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%s"}}}`, initTimeAnnotateKey, "")
		_, err := ctl.K8sClient.CoreV1().Namespaces().Patch(namespace.Name, "application/strategic-merge-patch+json", []byte(pathJson))
		if err != nil {
			glog.Warning("workspace roles sync failed", workspace, err)
			return
		}
	}
}

func (ctl *ClusterRoleBindingCtl) initListerAndInformer() {
	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)
	ctl.lister = informerFactory.Rbac().V1().ClusterRoleBindings().Lister()
	ctl.informer = informerFactory.Rbac().V1().ClusterRoleBindings().Informer()
	ctl.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
		},
		UpdateFunc: func(old, new interface{}) {
			oldValue := old.(*rbac.ClusterRoleBinding)
			newValue := new.(*rbac.ClusterRoleBinding)
			if !subjectsCompile(oldValue.Subjects, newValue.Subjects) {
				ctl.handleWorkspaceRoleChange(newValue)
			}
		},
		DeleteFunc: func(obj interface{}) {

		},
	})
}

func subjectsCompile(s1 []rbac.Subject, s2 []rbac.Subject) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v := range s1 {
		if v.Name != s2[i].Name || v.Kind != s2[i].Kind {
			return false
		}
	}
	return true
}

func (ctl *ClusterRoleBindingCtl) CountWithConditions(conditions string) int {
	return 0
}

func (ctl *ClusterRoleBindingCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	return 0, nil, errors.New("not implement")
}

func (ctl *ClusterRoleBindingCtl) Lister() interface{} {
	return ctl.lister
}
