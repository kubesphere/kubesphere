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

	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
)

func (ctl *RoleBindingCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *RoleBindingCtl) sync(stopChan chan struct{}) {
	ctl.initListerAndInformer()
	ctl.informer.Run(stopChan)
}

func (ctl *RoleBindingCtl) total() int {
	return 0
}

func (ctl *RoleBindingCtl) initListerAndInformer() {

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Rbac().V1().RoleBindings().Lister()
	ctl.informer = informerFactory.Rbac().V1().RoleBindings().Informer()
}

func (ctl *RoleBindingCtl) CountWithConditions(conditions string) int {
	return 0
}

func (ctl *RoleBindingCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	return 0, nil, errors.New("not implement")
}

func (ctl *RoleBindingCtl) Lister() interface{} {
	return ctl.lister
}
