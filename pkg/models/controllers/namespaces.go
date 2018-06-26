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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/options"
)

const (
	provider           = "kubernetes"
	admin              = "admin"
	editor             = "editor"
	viewer             = "viewer"
	kubectlNamespace   = "kubesphere"
	kubectlConfigKey   = "config"
	openpitrix_runtime = "openpitrix_runtime"
)

var adminRules = []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}
var editorRules = []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch"}, Resources: []string{"*"}}}
var viewerRules = []rbac.PolicyRule{{Verbs: []string{"list", "get", "watch"}, APIGroups: []string{"", "apps", "extensions", "batch"}, Resources: []string{"*"}}}

type runTime struct {
	RuntimeId         string `json:"runtime_id"`
	RuntimeUrl        string `json:"runtime_url"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Zone              string `json:"zone"`
	RuntimeCredential string `json:"runtime_credential"`
}

type DeleteRunTime struct {
	RuntimeId []string `json:"runtime_id"`
}

func makeHttpRequest(method, url, data string) ([]byte, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return body, err
}

func (ctl *NamespaceCtl) getKubeConfig(user string) (string, error) {
	k8sClient := client.NewK8sClient()
	configmap, err := k8sClient.CoreV1().ConfigMaps(kubectlNamespace).Get(user, metaV1.GetOptions{})
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	return configmap.Data[kubectlConfigKey], nil
}

func (ctl *NamespaceCtl) deleteOpRuntime(item v1.Namespace) {

	runtimeId := item.Annotations["openpitrix_runtime"]
	if len(runtimeId) == 0 {
		return
	}

	url := options.ServerOptions.GetOpAddress() + "/v1/runtimes"

	var deleteRuntime = DeleteRunTime{RuntimeId: []string{runtimeId}}

	body, err := json.Marshal(deleteRuntime)
	if err != nil {
		glog.Error(err)
		return
	}

	// todo: if delete failed, what's to be done?
	makeHttpRequest("DELETE", url, string(body))
}

func (ctl *NamespaceCtl) createOpRuntime(namespace string) ([]byte, error) {
	zone := namespace
	name := namespace
	kubeConfig, err := ctl.getKubeConfig("admin")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	url := options.ServerOptions.GetOpAddress() + "/v1/runtimes"

	option := runTime{Name: name, Provider: provider, RuntimeCredential: kubeConfig, Zone: zone}
	body, err := json.Marshal(option)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return makeHttpRequest("POST", url, string(body))
}

func (ctl *NamespaceCtl) createDefaultRoleBinding(ns, user string) error {
	rolebinding, _ := ctl.K8sClient.RbacV1().RoleBindings(ns).Get(admin, metaV1.GetOptions{})

	if rolebinding.Name != admin {

		roleBinding := &rbac.RoleBinding{ObjectMeta: metaV1.ObjectMeta{Name: admin, Namespace: ns},
			Subjects: []rbac.Subject{{Name: user, Kind: rbac.UserKind}}, RoleRef: rbac.RoleRef{Kind: "Role", Name: admin}}

		_, err := ctl.K8sClient.RbacV1().RoleBindings(ns).Create(roleBinding)

		if err != nil {
			glog.Error(err)
			return err
		}
	}

	return nil
}

func (ctl *NamespaceCtl) createDefaultRole(ns string) error {
	adminRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: admin, Namespace: ns}, Rules: adminRules}
	editorRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: editor, Namespace: ns}, Rules: editorRules}
	viewerRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: viewer, Namespace: ns}, Rules: viewerRules}

	role, _ := ctl.K8sClient.RbacV1().Roles(ns).Get(admin, metaV1.GetOptions{})

	if role.Name != admin {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(adminRole)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	role, _ = ctl.K8sClient.RbacV1().Roles(ns).Get(editor, metaV1.GetOptions{})

	if role.Name != editor {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(editorRole)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	role, _ = ctl.K8sClient.RbacV1().Roles(ns).Get(viewer, metaV1.GetOptions{})

	if role.Name != viewer {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(viewerRole)
		if err != nil {
			glog.Error(err)
			return err
		}
	}
	return nil
}

func (ctl *NamespaceCtl) createRoleAndRuntime(item v1.Namespace) {
	user := item.Annotations["creator"]
	ns := item.Name
	if len(user) > 0 && len(item.Annotations[openpitrix_runtime]) == 0 {
		err := ctl.createDefaultRole(ns)
		if err != nil {
			return
		}

		resp, err := ctl.createOpRuntime(ns)
		if err != nil {
			glog.Error(err)
			return
		}

		err = ctl.createDefaultRoleBinding(ns, user)
		if err != nil {
			glog.Error(err)
			return
		}

		var runtime runTime
		err = json.Unmarshal(resp, &runtime)
		if err != nil {
			glog.Error(err)
			return
		}

		item.Annotations[openpitrix_runtime] = runtime.RuntimeId
		_, err = ctl.K8sClient.CoreV1().Namespaces().Update(&item)
		if err != nil {
			glog.Error(err)
		}
	}
}

func (ctl *NamespaceCtl) generateObject(item v1.Namespace) *Namespace {

	name := item.Name
	createTime := item.CreationTimestamp.Time
	status := fmt.Sprintf("%v", item.Status.Phase)

	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &Namespace{Name: name, CreateTime: createTime, Status: status, Annotation: Annotation{item.Annotations}}

	return object
}

func (ctl *NamespaceCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Namespace{}) {
		db.DropTable(&Namespace{})
	}

	db = db.CreateTable(&Namespace{})

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Core().V1().Namespaces().Informer()
	lister := kubeInformerFactory.Core().V1().Namespaces().Lister()

	list, err := lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)
		ctl.createRoleAndRuntime(*item)

	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.Namespace)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
			ctl.createRoleAndRuntime(*object)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Namespace)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
			ctl.createRoleAndRuntime(*object)
		},
		DeleteFunc: func(obj interface{}) {
			var item Namespace
			object := obj.(*v1.Namespace)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)
			ctl.deleteOpRuntime(*object)

		},
	})

	informer.Run(ctl.stopChan)
}

func (ctl *NamespaceCtl) CountWithConditions(conditions string) int {
	var object Namespace

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *NamespaceCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Namespace
	var object Namespace
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index := range list {
		usage, err := ctl.GetNamespaceQuota(list[index].Name)
		if err == nil {
			list[index].Usaeg = usage
		}

	}
	return total, list, nil
}

func (ctl *NamespaceCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&Namespace{}).Count(&count)
	return count
}

func getUsage(namespace, resource string) int {
	ctl := rec.controllers[resource]
	return ctl.Count(namespace)
}

func (ctl *NamespaceCtl) GetNamespaceQuota(namespace string) (v1.ResourceList, error) {

	usage := make(v1.ResourceList)

	resourceList := []string{Daemonsets, Deployments, Ingresses, Roles, Services, Statefulsets, PersistentVolumeClaim, Pods}
	for _, resourceName := range resourceList {
		used := getUsage(namespace, resourceName)
		var quantity resource.Quantity
		quantity.Set(int64(used))
		usage[v1.ResourceName(resourceName)] = quantity
	}

	podCtl := rec.controllers[Pods]
	var quantity resource.Quantity
	used := podCtl.CountWithConditions(fmt.Sprintf("status=\"%s\" And namespace=\"%s\"", "Running", namespace))
	quantity.Set(int64(used))
	usage["runningPods"] = quantity
	return usage, nil
}
