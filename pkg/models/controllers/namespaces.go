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
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/options"
)

const (
	provider           = "kubernetes"
	admin              = "admin"
	normal             = "normal"
	view               = "view"
	kubectlNamespace   = "kubesphere"
	kubectlConfigKey   = "config"
	openpitrix_runtime = "openpitrix_runtime"
)

var adminRules = []rbac.PolicyRule{rbac.PolicyRule{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}
var normalRules = []rbac.PolicyRule{rbac.PolicyRule{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions"}, Resources: []string{"*"}}}
var viewRules = []rbac.PolicyRule{rbac.PolicyRule{Verbs: []string{"list", "get"}, APIGroups: []string{"", "apps", "extensions"}, Resources: []string{"*"}}}

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
	glog.Error(string(body))
	defer resp.Body.Close()
	return body, nil
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

func (ctl *NamespaceCtl) createOpRuntime(namespace, user string) ([]byte, error) {
	zone := namespace
	name := namespace
	kubeConfig, err := ctl.getKubeConfig(user)
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
	normalRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: normal, Namespace: ns}, Rules: normalRules}
	viewRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: view, Namespace: ns}, Rules: viewRules}

	role, _ := ctl.K8sClient.RbacV1().Roles(ns).Get(admin, metaV1.GetOptions{})

	if role.Name != admin {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(adminRole)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	role, _ = ctl.K8sClient.RbacV1().Roles(ns).Get(normal, metaV1.GetOptions{})

	if role.Name != normal {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(normalRole)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	role, _ = ctl.K8sClient.RbacV1().Roles(ns).Get(view, metaV1.GetOptions{})

	if role.Name != view {
		_, err := ctl.K8sClient.RbacV1().Roles(ns).Create(viewRole)
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

		err = ctl.createDefaultRoleBinding(ns, user)
		if err != nil {
			return
		}

		resp, err := ctl.createOpRuntime(ns, user)
		if err != nil {
			return
		}

		var runtime runTime
		err = json.Unmarshal(resp, &runtime)
		if err != nil {
			return
		}

		item.Annotations[openpitrix_runtime] = runtime.RuntimeId
		ctl.K8sClient.CoreV1().Namespaces().Update(&item)
	}
}

func (ctl *NamespaceCtl) generateObject(item v1.Namespace) *Namespace {

	name := item.Name
	createTime := item.CreationTimestamp.Time
	status := fmt.Sprintf("%v", item.Status.Phase)

	if createTime.IsZero() {
		createTime = time.Now()
	}

	annotation, _ := json.Marshal(item.Annotations)
	object := &Namespace{Name: name, CreateTime: createTime, Status: status, AnnotationStr: string(annotation)}

	return object
}

func (ctl *NamespaceCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
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

	k8sClient := client.NewK8sClient()
	list, err := k8sClient.CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)

		ctl.createRoleAndRuntime(item)
	}

	watcher, err := k8sClient.CoreV1().Namespaces().Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var ns Namespace
			if event.Object == nil {
				panic("watch timeout, restart namespace controller")
			}
			object := event.Object.(*v1.Namespace)
			if event.Type == watch.Deleted {
				db.Where("name=?", object.Name).Find(&ns)
				db.Delete(ns)

				ctl.deleteOpRuntime(*object)
				break
			}

			ctl.createRoleAndRuntime(*object)

			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
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

	for index, item := range list {
		annotation := make(map[string]string)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
	return total, list, nil
}

func (ctl *NamespaceCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&Namespace{}).Count(&count)
	return count
}
