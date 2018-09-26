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
	"fmt"
	"time"

	utilversion "k8s.io/kubernetes/pkg/util/version"

	"github.com/golang/glog"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/apis/core"
)

const (
	rbdPluginName        = "kubernetes.io/rbd"
	rbdUserSecretNameKey = "userSecretName"
)

func (ctl *StorageClassCtl) generateObject(item v1.StorageClass) *StorageClass {

	var displayName string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	createTime := item.CreationTimestamp.Time
	isDefault := false
	provisioner := item.Provisioner
	if item.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
		isDefault = true
	}

	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &StorageClass{
		Name:        name,
		DisplayName: displayName,
		CreateTime:  createTime,
		IsDefault:   isDefault,
		Annotation:  MapString{item.Annotations},
		Provisioner: provisioner,
	}

	return object
}

func (ctl *StorageClassCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *StorageClassCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&StorageClass{}) {
		db.DropTable(&StorageClass{})
	}

	db = db.CreateTable(&StorageClass{})

	ctl.initListerAndInformer()
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)
	}

	ctl.informer.Run(stopChan)
}

func (ctl *StorageClassCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *StorageClassCtl) createCephSecretAfterNewSc(item v1.StorageClass) {
	// Kubernetes version must < 1.11.0
	verInfo, err := ctl.K8sClient.ServerVersion()
	if err != nil {
		glog.Error("consult k8s server error: ", err)
		return
	}
	if !utilversion.MustParseSemantic(verInfo.String()).LessThan(utilversion.MustParseSemantic("v1.11.0")) {
		glog.Infof("disable Ceph secret controller due to k8s version %s >= v1.11.0", verInfo.String())
		return
	}

	// Find Ceph secret in the new storage class
	if item.Provisioner != rbdPluginName {
		return
	}
	var secret *coreV1.Secret
	if secretName, ok := item.Parameters[rbdUserSecretNameKey]; ok {
		secret, err = ctl.K8sClient.CoreV1().Secrets(core.NamespaceSystem).Get(secretName, metaV1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				glog.Errorf("cannot find secret %s in namespace %s", secretName, core.NamespaceSystem)
				return
			}
			glog.Error("failed to find secret, error: ", err)
			return
		}
		glog.Infof("succeed to find secret %s in namespace %s", secret.GetName(), secret.GetNamespace())
	} else {
		glog.Errorf("failed to find user secret name in storage class %s", item.GetName())
		return
	}

	// Create or update Ceph secret in each namespace
	nsList, err := ctl.K8sClient.CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return
	}
	for _, ns := range nsList.Items {
		if ns.GetName() == core.NamespaceSystem {
			glog.Infof("skip creating Ceph secret in namespace %s", core.NamespaceSystem)
			continue
		}
		newSecret := &coreV1.Secret{
			TypeMeta: metaV1.TypeMeta{
				Kind:       secret.Kind,
				APIVersion: secret.APIVersion,
			},
			ObjectMeta: metaV1.ObjectMeta{
				Name:                       secret.GetName(),
				Namespace:                  ns.GetName(),
				Labels:                     secret.GetLabels(),
				Annotations:                secret.GetAnnotations(),
				DeletionGracePeriodSeconds: secret.GetDeletionGracePeriodSeconds(),
				ClusterName:                secret.GetClusterName(),
			},
			Data:       secret.Data,
			StringData: secret.StringData,
			Type:       secret.Type,
		}

		_, err := ctl.K8sClient.CoreV1().Secrets(newSecret.GetNamespace()).Get(newSecret.GetName(), metaV1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				// Create secret
				_, err := ctl.K8sClient.CoreV1().Secrets(newSecret.GetNamespace()).Create(newSecret)
				if err != nil {
					glog.Errorf("failed to create secret in namespace %s, error: %v", newSecret.GetNamespace(), err)
				} else {
					glog.Infof("succeed to create secret %s in namespace %s", newSecret.GetName(),
						newSecret.GetNamespace())
				}
			} else {
				glog.Errorf("failed to find secret in namespace %s, error: %v", newSecret.GetNamespace(), err)
			}
		} else {
			// Update secret
			_, err = ctl.K8sClient.CoreV1().Secrets(newSecret.GetNamespace()).Update(newSecret)
			if err != nil {
				glog.Errorf("failed to update secret in namespace %s, error: %v", newSecret.GetNamespace(), err)
				continue
			} else {
				glog.Infof("succeed to update secret %s in namespace %s", newSecret.GetName(), newSecret.GetNamespace())
			}
		}
	}
}

func (ctl *StorageClassCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Storage().V1().StorageClasses().Lister()

	informer := informerFactory.Storage().V1().StorageClasses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			object := obj.(*v1.StorageClass)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
			ctl.createCephSecretAfterNewSc(*object)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.StorageClass)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item StorageClass
			object := obj.(*v1.StorageClass)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)

		},
	})

	ctl.informer = informer
}

func (ctl *StorageClassCtl) CountWithConditions(conditions string) int {
	var object StorageClass

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *StorageClassCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []StorageClass
	var object StorageClass
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index, storageClass := range list {
		name := storageClass.Name
		pvcCtl := ResourceControllers.Controllers[PersistentVolumeClaim]

		list[index].Count = pvcCtl.CountWithConditions(fmt.Sprintf("storage_class=\"%s\"", name))
	}

	return total, list, nil
}

func (ctl *StorageClassCtl) Lister() interface{} {

	return ctl.lister
}
