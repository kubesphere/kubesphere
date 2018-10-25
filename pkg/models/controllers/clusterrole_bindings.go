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

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
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

func (ctl *ClusterRoleBindingCtl) initListerAndInformer() {
	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)
	ctl.lister = informerFactory.Rbac().V1().ClusterRoleBindings().Lister()
	ctl.informer = informerFactory.Rbac().V1().ClusterRoleBindings().Informer()
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
