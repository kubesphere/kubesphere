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

	"k8s.io/client-go/informers"
)

func (ctl *ReplicaSetCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *ReplicaSetCtl) sync(stopChan chan struct{}) {

	ctl.initListerAndInformer()
	ctl.informer.Run(stopChan)
}

func (ctl *ReplicaSetCtl) total() int {

	return 0
}

func (ctl *ReplicaSetCtl) initListerAndInformer() {

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Apps().V1().ReplicaSets().Lister()

	informer := informerFactory.Apps().V1().ReplicaSets().Informer()

	ctl.informer = informer
}

func (ctl *ReplicaSetCtl) CountWithConditions(conditions string) int {

	return 0
}

func (ctl *ReplicaSetCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {

	return 0, nil, nil
}

func (ctl *ReplicaSetCtl) Lister() interface{} {

	return ctl.lister
}
